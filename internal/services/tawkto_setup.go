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
	"strconv"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// TawkToService maneja la integración con Tawk.to
type TawkToService struct {
	config     *config.TawkToConfig
	repo       domain.ChannelIntegrationRepository
	logger     logger.Logger
	httpClient *http.Client
}

// TawkToConfig representa la configuración de Tawk.to para un tenant
type TawkToConfig struct {
	WidgetID   string    `json:"widget_id"`
	PropertyID string    `json:"property_id"`
	APIKey     string    `json:"api_key"`
	BaseURL    string    `json:"base_url"`
	WebhookURL string    `json:"webhook_url"`
	CustomCSS  string    `json:"custom_css,omitempty"`
	CustomJS   string    `json:"custom_js,omitempty"`
	Greeting   string    `json:"greeting,omitempty"`
	OfflineMsg string    `json:"offline_msg,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TawkToWebhookPayload representa el payload de webhook de Tawk.to
type TawkToWebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Visitor   TawkToVisitor          `json:"visitor"`
	Chat      TawkToChat             `json:"chat"`
}

// TawkToVisitor representa un visitante de Tawk.to
type TawkToVisitor struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
}

// TawkToChat representa un chat de Tawk.to
type TawkToChat struct {
	ID       string          `json:"id"`
	Session  string          `json:"session"`
	Status   string          `json:"status"`
	Messages []TawkToMessage `json:"messages"`
}

// TawkToMessage representa un mensaje de Tawk.to
type TawkToMessage struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
}

// NewTawkToService crea una nueva instancia del servicio Tawk.to
func NewTawkToService(cfg *config.TawkToConfig, repo domain.ChannelIntegrationRepository, logger logger.Logger) *TawkToService {
	return &TawkToService{
		config: cfg,
		repo:   repo,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetupTawkToIntegration configura la integración de Tawk.to para un tenant
func (s *TawkToService) SetupTawkToIntegration(tenantID string, config *TawkToConfig) (*domain.ChannelIntegration, error) {
	s.logger.Info("Configurando integración Tawk.to", "tenant_id", tenantID)

	// Validar configuración
	if err := s.validateTawkToConfig(config); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	// Verificar credenciales con Tawk.to
	if err := s.verifyTawkToCredentials(config); err != nil {
		return nil, fmt.Errorf("credenciales inválidas: %w", err)
	}

	// Crear configuración en formato JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error serializando configuración: %w", err)
	}

	// Crear integración en la base de datos
	integration := &domain.ChannelIntegration{
		TenantID:  tenantID,
		Platform:  domain.PlatformWebchat,
		Provider:  domain.ProviderCustom,
		Config:    configJSON,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar en la base de datos
	if err := s.repo.Create(context.Background(), integration); err != nil {
		return nil, fmt.Errorf("error guardando integración: %w", err)
	}

	// Configurar webhook en Tawk.to
	if err := s.setupTawkToWebhook(config, integration.ID); err != nil {
		s.logger.Warn("Error configurando webhook de Tawk.to", "error", err)
		// No fallamos la integración por esto, solo loggeamos
	}

	s.logger.Info("Integración Tawk.to configurada exitosamente", "tenant_id", tenantID, "integration_id", integration.ID)
	return integration, nil
}

// GetTawkToConfig obtiene la configuración de Tawk.to para un tenant
func (s *TawkToService) GetTawkToConfig(tenantID string) (*TawkToConfig, error) {
	integration, err := s.repo.GetByPlatformAndTenant(context.Background(), domain.PlatformWebchat, tenantID)
	if err != nil {
		return nil, fmt.Errorf("integración no encontrada: %w", err)
	}

	var config TawkToConfig
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return nil, fmt.Errorf("error deserializando configuración: %w", err)
	}

	return &config, nil
}

// UpdateTawkToConfig actualiza la configuración de Tawk.to
func (s *TawkToService) UpdateTawkToConfig(tenantID string, config *TawkToConfig) error {
	s.logger.Info("Actualizando configuración Tawk.to", "tenant_id", tenantID)

	// Validar configuración
	if err := s.validateTawkToConfig(config); err != nil {
		return fmt.Errorf("configuración inválida: %w", err)
	}

	// Obtener integración existente
	integration, err := s.repo.GetByPlatformAndTenant(context.Background(), domain.PlatformWebchat, tenantID)
	if err != nil {
		return fmt.Errorf("integración no encontrada: %w", err)
	}

	// Actualizar configuración
	config.UpdatedAt = time.Now()
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error serializando configuración: %w", err)
	}

	integration.Config = configJSON
	integration.UpdatedAt = time.Now()

	// Guardar cambios
	if err := s.repo.Update(context.Background(), integration); err != nil {
		return fmt.Errorf("error actualizando integración: %w", err)
	}

	s.logger.Info("Configuración Tawk.to actualizada exitosamente", "tenant_id", tenantID)
	return nil
}

// ProcessTawkToWebhook procesa los webhooks de Tawk.to
func (s *TawkToService) ProcessTawkToWebhook(payload []byte, signature string) (*NormalizedMessage, error) {
	s.logger.Info("Procesando webhook de Tawk.to")

	// Validar firma del webhook
	if err := s.validateWebhookSignature(payload, signature); err != nil {
		return nil, fmt.Errorf("firma inválida: %w", err)
	}

	// Parsear payload
	var webhookPayload TawkToWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("error parseando payload: %w", err)
	}

	// Normalizar mensaje
	message := s.normalizeTawkToMessage(&webhookPayload)

	s.logger.Info("Webhook de Tawk.to procesado exitosamente", "event", webhookPayload.Event, "chat_id", webhookPayload.Chat.ID)
	return message, nil
}

// GetTawkToAnalytics obtiene analytics de Tawk.to
func (s *TawkToService) GetTawkToAnalytics(tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	config, err := s.GetTawkToConfig(tenantID)
	if err != nil {
		return nil, err
	}

	// Construir URL para analytics
	url := fmt.Sprintf("%s/analytics/chat", config.BaseURL)

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Agregar headers
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Agregar query parameters
	q := req.URL.Query()
	q.Add("startDate", startDate.Format("2006-01-02"))
	q.Add("endDate", endDate.Format("2006-01-02"))
	q.Add("propertyId", config.PropertyID)
	req.URL.RawQuery = q.Encode()

	// Hacer request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error de API: %d - %s", resp.StatusCode, string(body))
	}

	// Parsear respuesta
	var analytics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&analytics); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	return analytics, nil
}

// GetTawkToSessions obtiene sesiones de chat de Tawk.to
func (s *TawkToService) GetTawkToSessions(tenantID string, limit int) ([]map[string]interface{}, error) {
	config, err := s.GetTawkToConfig(tenantID)
	if err != nil {
		return nil, err
	}

	// Construir URL para sesiones
	url := fmt.Sprintf("%s/chat/sessions", config.BaseURL)

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Agregar headers
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Agregar query parameters
	q := req.URL.Query()
	q.Add("propertyId", config.PropertyID)
	q.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	// Hacer request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error de API: %d - %s", resp.StatusCode, string(body))
	}

	// Parsear respuesta
	var sessions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	return sessions, nil
}

// validateTawkToConfig valida la configuración de Tawk.to
func (s *TawkToService) validateTawkToConfig(config *TawkToConfig) error {
	if config.WidgetID == "" {
		return fmt.Errorf("widget_id es requerido")
	}
	if config.PropertyID == "" {
		return fmt.Errorf("property_id es requerido")
	}
	if config.APIKey == "" {
		return fmt.Errorf("api_key es requerido")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base_url es requerido")
	}
	return nil
}

// verifyTawkToCredentials verifica las credenciales con Tawk.to
func (s *TawkToService) verifyTawkToCredentials(config *TawkToConfig) error {
	url := fmt.Sprintf("%s/properties/%s", config.BaseURL, config.PropertyID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error verificando credenciales: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("credenciales inválidas: %d", resp.StatusCode)
	}

	return nil
}

// setupTawkToWebhook configura el webhook en Tawk.to
func (s *TawkToService) setupTawkToWebhook(config *TawkToConfig, integrationID string) error {
	webhookURL := fmt.Sprintf("%s/webhook/tawkto/%s", s.config.BaseURL, integrationID)

	webhookData := map[string]interface{}{
		"url": webhookURL,
		"events": []string{
			"chat_message",
			"chat_start",
			"chat_end",
			"visitor_join",
			"visitor_leave",
		},
	}

	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("error serializando webhook: %w", err)
	}

	url := fmt.Sprintf("%s/webhooks", config.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(webhookJSON))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error configurando webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error configurando webhook: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// validateWebhookSignature valida la firma del webhook
func (s *TawkToService) validateWebhookSignature(payload []byte, signature string) error {
	if s.config.WebhookSecret == "" {
		s.logger.Warn("Webhook secret no configurado, saltando validación de firma")
		return nil
	}

	// Calcular HMAC SHA256
	h := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	h.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("firma inválida")
	}

	return nil
}

// normalizeTawkToMessage normaliza un mensaje de Tawk.to a nuestro formato
func (s *TawkToService) normalizeTawkToMessage(webhook *TawkToWebhookPayload) *NormalizedMessage {
	// Extraer el último mensaje si existe
	var lastMessage TawkToMessage
	if len(webhook.Chat.Messages) > 0 {
		lastMessage = webhook.Chat.Messages[len(webhook.Chat.Messages)-1]
	}

	// Determinar el tipo de contenido
	contentType := "text"
	if lastMessage.Type != "" {
		contentType = lastMessage.Type
	}

	// Determinar el remitente
	sender := "visitor"
	if lastMessage.Sender == "agent" {
		sender = "agent"
	}

	// Crear contenido del mensaje
	messageContent := &domain.MessageContent{
		Type: contentType,
		Text: lastMessage.Content,
	}

	// Convertir webhook a JSON para RawPayload
	rawPayload, _ := json.Marshal(webhook)

	return &NormalizedMessage{
		Platform:   domain.PlatformWebchat,
		Sender:     sender,
		Recipient:  webhook.Visitor.ID,
		Content:    messageContent,
		Timestamp:  lastMessage.Timestamp.Unix(),
		MessageID:  lastMessage.ID,
		ChannelID:  webhook.Chat.ID,
		RawPayload: rawPayload,
	}
}
