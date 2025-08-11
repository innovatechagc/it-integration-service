package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// WhatsAppSetupService maneja la configuración específica de WhatsApp
type WhatsAppSetupService struct {
	logger logger.Logger
}

// NewWhatsAppSetupService crea una nueva instancia del servicio de configuración de WhatsApp
func NewWhatsAppSetupService(logger logger.Logger) *WhatsAppSetupService {
	return &WhatsAppSetupService{
		logger: logger,
	}
}

// WhatsAppBusinessInfo representa la información de la cuenta de WhatsApp Business
type WhatsAppBusinessInfo struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Category          string `json:"category"`
	Description       string `json:"description"`
	Website           string `json:"website"`
	Email             string `json:"email"`
	PhoneNumber       string `json:"phone_number"`
	ProfilePictureURL string `json:"profile_picture_url"`
	Status            string `json:"status"`
}

// WhatsAppPhoneNumberInfo representa la información del número de teléfono
type WhatsAppPhoneNumberInfo struct {
	ID                   string `json:"id"`
	DisplayPhoneNumber   string `json:"display_phone_number"`
	VerifiedName         string `json:"verified_name"`
	CodeVerificationStatus string `json:"code_verification_status"`
	QualityRating        string `json:"quality_rating"`
	PlatformType         string `json:"platform_type"`
	ThroughputLevel      string `json:"throughput_level"`
}

// WhatsAppWebhookSubscription representa una suscripción de webhook
type WhatsAppWebhookSubscription struct {
	Object string   `json:"object"`
	Fields []string `json:"fields"`
}



// GetBusinessInfo obtiene información de la cuenta de WhatsApp Business
func (s *WhatsAppSetupService) GetBusinessInfo(ctx context.Context, accessToken, businessAccountID string) (*WhatsAppBusinessInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", businessAccountID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get business info: %w", err)
	}
	defer resp.Body.Close()

	var apiResp MetaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("meta API error: %s", apiResp.Error.Message)
	}

	var businessInfo WhatsAppBusinessInfo
	if err := json.Unmarshal(apiResp.Data, &businessInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal business info: %w", err)
	}

	return &businessInfo, nil
}

// GetPhoneNumberInfo obtiene información del número de teléfono
func (s *WhatsAppSetupService) GetPhoneNumberInfo(ctx context.Context, accessToken, phoneNumberID string) (*WhatsAppPhoneNumberInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", phoneNumberID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get phone number info: %w", err)
	}
	defer resp.Body.Close()

	var phoneInfo WhatsAppPhoneNumberInfo
	if err := json.NewDecoder(resp.Body).Decode(&phoneInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &phoneInfo, nil
}

// SubscribeToWebhooks suscribe la aplicación a webhooks de WhatsApp
func (s *WhatsAppSetupService) SubscribeToWebhooks(ctx context.Context, accessToken, appID string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/subscriptions", appID)
	
	payload := map[string]interface{}{
		"object":       "whatsapp_business_account",
		"callback_url": "https://tu-dominio.com/api/v1/integrations/webhooks/whatsapp", // Se actualizará dinámicamente
		"fields":       []string{"messages", "message_deliveries", "message_reads", "message_reactions"},
		"verify_token": "wpp-it-app-webhook-verify-token",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to webhooks: %w", err)
	}
	defer resp.Body.Close()

	var apiResp MetaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("meta API error: %s", apiResp.Error.Message)
	}

	s.logger.Info("WhatsApp webhook subscription configured successfully")
	return nil
}

// SendMessage envía un mensaje a través de WhatsApp
func (s *WhatsAppSetupService) SendMessage(ctx context.Context, accessToken, phoneNumberID, recipient, text string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneNumberID)
	
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":               recipient,
		"type":             "text",
		"text": map[string]string{
			"body": text,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp MetaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("meta API error: %s", apiResp.Error.Message)
	}

	s.logger.Info("WhatsApp message sent successfully", map[string]interface{}{
		"recipient": recipient,
		"text":      text,
	})

	return nil
}

// CreateWhatsAppIntegration crea una integración de WhatsApp con configuración completa
func (s *WhatsAppSetupService) CreateWhatsAppIntegration(ctx context.Context, accessToken, phoneNumberID, businessAccountID, webhookURL, tenantID string) (*domain.ChannelIntegration, error) {
	// Verificar información del número de teléfono (obligatorio)
	phoneInfo, err := s.GetPhoneNumberInfo(ctx, accessToken, phoneNumberID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify phone number: %w", err)
	}

	s.logger.Info("Phone number verified successfully", map[string]interface{}{
		"phone_id":            phoneInfo.ID,
		"display_phone":       phoneInfo.DisplayPhoneNumber,
		"verified_name":       phoneInfo.VerifiedName,
		"quality_rating":      phoneInfo.QualityRating,
		"verification_status": phoneInfo.CodeVerificationStatus,
	})

	// Crear configuración base
	config := map[string]interface{}{
		"access_token":          accessToken,
		"phone_number_id":       phoneNumberID,
		"webhook_url":           webhookURL,
		"display_phone_number":  phoneInfo.DisplayPhoneNumber,
		"verified_name":         phoneInfo.VerifiedName,
		"quality_rating":        phoneInfo.QualityRating,
		"platform_type":         phoneInfo.PlatformType,
		"throughput_level":      phoneInfo.ThroughputLevel,
	}

	// Intentar verificar información del negocio (opcional)
	if businessAccountID != "" {
		businessInfo, err := s.GetBusinessInfo(ctx, accessToken, businessAccountID)
		if err != nil {
			s.logger.Warn("Failed to verify business info, continuing without it", map[string]interface{}{
				"business_account_id": businessAccountID,
				"error":              err.Error(),
			})
			// Agregar el ID aunque no podamos verificarlo
			config["business_account_id"] = businessAccountID
			config["business_name"] = "Unknown Business"
		} else {
			s.logger.Info("Business verified successfully", map[string]interface{}{
				"business_id":   businessInfo.ID,
				"business_name": businessInfo.Name,
				"status":        businessInfo.Status,
			})
			config["business_account_id"] = businessAccountID
			config["business_name"] = businessInfo.Name
			config["business_status"] = businessInfo.Status
		}
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	integration := &domain.ChannelIntegration{
		TenantID:    tenantID,
		Platform:    domain.PlatformWhatsApp,
		Provider:    domain.ProviderMeta,
		AccessToken: accessToken,
		WebhookURL:  webhookURL,
		Status:      domain.StatusActive,
		Config:      configJSON,
	}

	return integration, nil
}

// ValidateWebhookToken valida el token de verificación del webhook
func (s *WhatsAppSetupService) ValidateWebhookToken(providedToken, expectedToken string) bool {
	return providedToken == expectedToken
}