package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// MailchimpSetupService maneja la configuración de integraciones con Mailchimp
type MailchimpSetupService struct {
	config     *config.MailchimpConfig
	repo       domain.ChannelIntegrationRepository
	logger     logger.Logger
	httpClient *http.Client
}

// MailchimpConfig representa la configuración de Mailchimp para un tenant
type MailchimpConfig struct {
	APIKey       string    `json:"api_key"`
	ServerPrefix string    `json:"server_prefix"`
	BaseURL      string    `json:"base_url"`
	AudienceID   string    `json:"audience_id"`
	DataCenter   string    `json:"data_center"`
	WebhookURL   string    `json:"webhook_url"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MailchimpAccountInfo representa la información de la cuenta de Mailchimp
type MailchimpAccountInfo struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	Contact     struct {
		Company  string `json:"company"`
		Address1 string `json:"address1"`
		Address2 string `json:"address2"`
		City     string `json:"city"`
		State    string `json:"state"`
		Zip      string `json:"zip"`
		Country  string `json:"country"`
		Phone    string `json:"phone"`
	} `json:"contact"`
	Enabled bool `json:"enabled"`
}

// MailchimpAudienceInfo representa la información de una audiencia de Mailchimp
type MailchimpAudienceInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmailType    string `json:"email_type"`
	Status       string `json:"status"`
	SubscriberCount int `json:"stats.subscriber_count"`
	UnsubscribeCount int `json:"stats.unsubscribe_count"`
	CleanCount   int `json:"stats.clean_count"`
	MemberCount  int `json:"stats.member_count"`
	CreatedAt    string `json:"date_created"`
	UpdatedAt    string `json:"date_updated"`
}

// MailchimpWebhookPayload representa el payload de webhook de Mailchimp
type MailchimpWebhookPayload struct {
	Type    string                 `json:"type"`
	FiredAt string                 `json:"fired_at"`
	Data    map[string]interface{} `json:"data"`
	ListID  string                 `json:"list_id"`
}

// NewMailchimpSetupService crea una nueva instancia del servicio de configuración de Mailchimp
func NewMailchimpSetupService(cfg *config.MailchimpConfig, repo domain.ChannelIntegrationRepository, logger logger.Logger) *MailchimpSetupService {
	return &MailchimpSetupService{
		config: cfg,
		repo:   repo,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetupMailchimpIntegration configura la integración de Mailchimp para un tenant
func (s *MailchimpSetupService) SetupMailchimpIntegration(tenantID string, config *MailchimpConfig) (*domain.ChannelIntegration, error) {
	s.logger.Info("Configurando integración Mailchimp", "tenant_id", tenantID)

	// Validar configuración
	if err := s.validateMailchimpConfig(config); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	// Verificar credenciales con Mailchimp
	if err := s.verifyMailchimpCredentials(config); err != nil {
		return nil, fmt.Errorf("credenciales inválidas: %w", err)
	}

	// Obtener información de la cuenta
	accountInfo, err := s.GetAccountInfo(config)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información de cuenta: %w", err)
	}

	// Obtener información de la audiencia
	audienceInfo, err := s.GetAudienceInfo(config)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información de audiencia: %w", err)
	}

	// Crear configuración en formato JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error serializando configuración: %w", err)
	}

	// Crear integración en la base de datos
	integration := &domain.ChannelIntegration{
		TenantID:  tenantID,
		Platform:  domain.PlatformMailchimp,
		Provider:  domain.ProviderMailchimp,
		Config:    configJSON,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar en la base de datos
	if err := s.repo.Create(context.Background(), integration); err != nil {
		return nil, fmt.Errorf("error guardando integración: %w", err)
	}

	// Configurar webhook en Mailchimp
	if err := s.setupMailchimpWebhook(config, integration.ID); err != nil {
		s.logger.Warn("Error configurando webhook, continuando sin él", "error", err.Error())
	}

	s.logger.Info("Integración Mailchimp configurada exitosamente", map[string]interface{}{
		"tenant_id":     tenantID,
		"integration_id": integration.ID,
		"account_name":  accountInfo.AccountName,
		"audience_name": audienceInfo.Name,
		"subscriber_count": audienceInfo.SubscriberCount,
	})

	return integration, nil
}

// GetMailchimpConfig obtiene la configuración de Mailchimp para un tenant
func (s *MailchimpSetupService) GetMailchimpConfig(tenantID string) (*MailchimpConfig, error) {
	integrations, err := s.repo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo integraciones: %w", err)
	}

	for _, integration := range integrations {
		if integration.Platform == domain.PlatformMailchimp {
			var config MailchimpConfig
			if err := json.Unmarshal(integration.Config, &config); err != nil {
				return nil, fmt.Errorf("error deserializando configuración: %w", err)
			}
			return &config, nil
		}
	}

	return nil, fmt.Errorf("no se encontró configuración de Mailchimp para el tenant")
}

// UpdateMailchimpConfig actualiza la configuración de Mailchimp para un tenant
func (s *MailchimpSetupService) UpdateMailchimpConfig(tenantID string, config *MailchimpConfig) error {
	integrations, err := s.repo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return fmt.Errorf("error obteniendo integraciones: %w", err)
	}

	for _, integration := range integrations {
		if integration.Platform == domain.PlatformMailchimp {
			// Validar nueva configuración
			if err := s.validateMailchimpConfig(config); err != nil {
				return fmt.Errorf("configuración inválida: %w", err)
			}

			// Verificar credenciales
			if err := s.verifyMailchimpCredentials(config); err != nil {
				return fmt.Errorf("credenciales inválidas: %w", err)
			}

			// Actualizar configuración
			configJSON, err := json.Marshal(config)
			if err != nil {
				return fmt.Errorf("error serializando configuración: %w", err)
			}

			integration.Config = configJSON
			integration.UpdatedAt = time.Now()

			if err := s.repo.Update(context.Background(), integration); err != nil {
				return fmt.Errorf("error actualizando integración: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("no se encontró integración de Mailchimp para actualizar")
}

// ProcessMailchimpWebhook procesa los webhooks de Mailchimp
func (s *MailchimpSetupService) ProcessMailchimpWebhook(payload []byte, signature string) (*NormalizedMessage, error) {
	// Validar firma del webhook
	if err := s.validateWebhookSignature(payload, signature); err != nil {
		return nil, fmt.Errorf("firma de webhook inválida: %w", err)
	}

	// Parsear payload
	var webhookPayload MailchimpWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("error parseando payload: %w", err)
	}

	// Normalizar mensaje
	normalizedMessage := s.normalizeMailchimpMessage(&webhookPayload)

	return normalizedMessage, nil
}

// GetMailchimpAnalytics obtiene analytics de Mailchimp para un tenant
func (s *MailchimpSetupService) GetMailchimpAnalytics(tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	config, err := s.GetMailchimpConfig(tenantID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	// Construir URL para analytics
	url := fmt.Sprintf("%s/3.0/reports", s.buildAPIURL(config))
	
	// Agregar parámetros de fecha
	url += fmt.Sprintf("?since_send_time=%s&before_send_time=%s", 
		startDate.Format("2006-01-02"), 
		endDate.Format("2006-01-02"))

	// Realizar request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error realizando request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error en API de Mailchimp: %d - %s", resp.StatusCode, string(body))
	}

	var analytics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&analytics); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return analytics, nil
}

// GetAccountInfo obtiene información de la cuenta de Mailchimp
func (s *MailchimpSetupService) GetAccountInfo(config *MailchimpConfig) (*MailchimpAccountInfo, error) {
	url := s.buildAPIURL(config) + "/3.0/account"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error realizando request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error en API de Mailchimp: %d - %s", resp.StatusCode, string(body))
	}

	var accountInfo MailchimpAccountInfo
	if err := json.NewDecoder(resp.Body).Decode(&accountInfo); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return &accountInfo, nil
}

// GetAudienceInfo obtiene información de la audiencia de Mailchimp
func (s *MailchimpSetupService) GetAudienceInfo(config *MailchimpConfig) (*MailchimpAudienceInfo, error) {
	url := s.buildAPIURL(config) + "/3.0/lists/" + config.AudienceID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error realizando request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error en API de Mailchimp: %d - %s", resp.StatusCode, string(body))
	}

	var audienceInfo MailchimpAudienceInfo
	if err := json.NewDecoder(resp.Body).Decode(&audienceInfo); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return &audienceInfo, nil
}

// validateMailchimpConfig valida la configuración de Mailchimp
func (s *MailchimpSetupService) validateMailchimpConfig(config *MailchimpConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key es requerida")
	}
	if config.ServerPrefix == "" {
		return fmt.Errorf("server prefix es requerido")
	}
	if config.AudienceID == "" {
		return fmt.Errorf("audience ID es requerido")
	}
	return nil
}

// verifyMailchimpCredentials verifica las credenciales de Mailchimp
func (s *MailchimpSetupService) verifyMailchimpCredentials(config *MailchimpConfig) error {
	// Intentar obtener información de la cuenta para verificar credenciales
	_, err := s.GetAccountInfo(config)
	return err
}

// setupMailchimpWebhook configura el webhook en Mailchimp
func (s *MailchimpSetupService) setupMailchimpWebhook(config *MailchimpConfig, integrationID string) error {
	webhookURL := fmt.Sprintf("%s/api/v1/integrations/webhooks/mailchimp", config.WebhookURL)
	
	webhookData := map[string]interface{}{
		"url":    webhookURL,
		"events": map[string]bool{
			"subscribe":   true,
			"unsubscribe": true,
			"profile":     true,
			"cleaned":     true,
			"upemail":     true,
			"campaign":    true,
		},
		"sources": map[string]bool{
			"user":  true,
			"admin": true,
			"api":   true,
		},
	}

	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("error serializando datos del webhook: %w", err)
	}

	url := s.buildAPIURL(config) + "/3.0/lists/" + config.AudienceID + "/webhooks"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error realizando request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error configurando webhook: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// validateWebhookSignature valida la firma del webhook
func (s *MailchimpSetupService) validateWebhookSignature(payload []byte, signature string) error {
	if s.config.WebhookSecret == "" {
		s.logger.Warn("Webhook secret no configurado, saltando validación de firma")
		return nil
	}

	// Mailchimp usa HMAC-SHA256 para firmar webhooks
	h := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("firma de webhook inválida")
	}

	return nil
}

// normalizeMailchimpMessage normaliza un mensaje de Mailchimp
func (s *MailchimpSetupService) normalizeMailchimpMessage(webhook *MailchimpWebhookPayload) *NormalizedMessage {
	// Extraer información del payload
	var sender, recipient, content string
	var messageType string

	switch webhook.Type {
	case "subscribe":
		messageType = "subscription"
		if data, ok := webhook.Data["email"].(string); ok {
			recipient = data
		}
		content = "Usuario suscrito a la lista"
	case "unsubscribe":
		messageType = "unsubscription"
		if data, ok := webhook.Data["email"].(string); ok {
			recipient = data
		}
		content = "Usuario desuscrito de la lista"
	case "profile":
		messageType = "profile_update"
		if data, ok := webhook.Data["email"].(string); ok {
			recipient = data
		}
		content = "Perfil de usuario actualizado"
	case "cleaned":
		messageType = "email_cleaned"
		if data, ok := webhook.Data["email"].(string); ok {
			recipient = data
		}
		content = "Email limpiado de la lista"
	case "upemail":
		messageType = "email_changed"
		if data, ok := webhook.Data["new_email"].(string); ok {
			recipient = data
		}
		content = "Email de usuario cambiado"
	case "campaign":
		messageType = "campaign_event"
		if data, ok := webhook.Data["campaign_id"].(string); ok {
			content = fmt.Sprintf("Evento de campaña: %s", data)
		}
	default:
		messageType = "unknown"
		content = fmt.Sprintf("Evento desconocido: %s", webhook.Type)
	}

	// Parsear timestamp
	timestamp := time.Now()
	if webhook.FiredAt != "" {
		if ts, err := time.Parse(time.RFC3339, webhook.FiredAt); err == nil {
			timestamp = ts
		}
	}

	// Convertir webhook.Data a json.RawMessage
	rawPayload, _ := json.Marshal(webhook.Data)
	
	// Crear MessageContent
	messageContent := &domain.MessageContent{
		Type: messageType,
		Text: content,
	}

	return &NormalizedMessage{
		Platform:   domain.PlatformMailchimp,
		MessageID:  fmt.Sprintf("mailchimp_%s_%d", webhook.Type, timestamp.Unix()),
		Sender:     sender,
		Recipient:  recipient,
		Content:    messageContent,
		Timestamp:  timestamp.Unix(),
		RawPayload: rawPayload,
	}
}

// buildAPIURL construye la URL base de la API de Mailchimp
func (s *MailchimpSetupService) buildAPIURL(config *MailchimpConfig) string {
	if config.BaseURL != "" {
		return config.BaseURL
	}
	return fmt.Sprintf("https://%s.api.mailchimp.com", config.ServerPrefix)
}
