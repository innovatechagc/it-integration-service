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

// MessengerSetupService maneja la configuración específica de Messenger
type MessengerSetupService struct {
	logger logger.Logger
}

// NewMessengerSetupService crea una nueva instancia del servicio de configuración de Messenger
func NewMessengerSetupService(logger logger.Logger) *MessengerSetupService {
	return &MessengerSetupService{
		logger: logger,
	}
}

// MessengerPageInfo representa la información de la página de Facebook
type MessengerPageInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	About    string `json:"about"`
	Website  string `json:"website"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Picture  struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// MessengerWebhookSubscription representa una suscripción de webhook
type MessengerWebhookSubscription struct {
	Object string   `json:"object"`
	Fields []string `json:"fields"`
}



// GetPageInfo obtiene información de la página de Facebook
func (s *MessengerSetupService) GetPageInfo(ctx context.Context, pageAccessToken, pageID string) (*MessengerPageInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,name,category,about,website,phone,email,picture&access_token=%s", pageID, pageAccessToken)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}
	defer resp.Body.Close()

	var pageInfo MessengerPageInfo
	if err := json.NewDecoder(resp.Body).Decode(&pageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Verificar si hay error en la respuesta
	if pageInfo.ID == "" {
		var errorResp MetaAPIResponse
		resp.Body.Close()
		
		// Hacer la petición de nuevo para obtener el error
		resp2, _ := client.Do(req)
		if resp2 != nil {
			defer resp2.Body.Close()
			json.NewDecoder(resp2.Body).Decode(&errorResp)
			if errorResp.Error != nil {
				return nil, fmt.Errorf("facebook API error: %s", errorResp.Error.Message)
			}
		}
		return nil, fmt.Errorf("invalid page response")
	}

	return &pageInfo, nil
}

// SubscribeToWebhooks suscribe la página a webhooks de Messenger
func (s *MessengerSetupService) SubscribeToWebhooks(ctx context.Context, pageAccessToken, pageID string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/subscribed_apps", pageID)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Agregar el access token como parámetro
	q := req.URL.Query()
	q.Add("access_token", pageAccessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to webhooks: %w", err)
	}
	defer resp.Body.Close()

	var apiResp FacebookAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("facebook API error: %s", apiResp.Error.Message)
	}

	s.logger.Info("Messenger webhook subscription configured successfully")
	return nil
}

// SendMessage envía un mensaje a través de Messenger
func (s *MessengerSetupService) SendMessage(ctx context.Context, pageAccessToken, recipientID, text string) error {
	url := "https://graph.facebook.com/v18.0/me/messages"
	
	payload := map[string]interface{}{
		"recipient": map[string]string{
			"id": recipientID,
		},
		"message": map[string]string{
			"text": text,
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

	req.Header.Set("Content-Type", "application/json")
	
	// Agregar el access token como parámetro
	q := req.URL.Query()
	q.Add("access_token", pageAccessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp FacebookAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("facebook API error: %s", apiResp.Error.Message)
	}

	s.logger.Info("Messenger message sent successfully", map[string]interface{}{
		"recipient": recipientID,
		"text":      text,
	})

	return nil
}

// CreateMessengerIntegration crea una integración de Messenger con configuración completa
func (s *MessengerSetupService) CreateMessengerIntegration(ctx context.Context, pageAccessToken, pageID, webhookURL, tenantID string) (*domain.ChannelIntegration, error) {
	// Verificar información de la página
	pageInfo, err := s.GetPageInfo(ctx, pageAccessToken, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify page: %w", err)
	}

	s.logger.Info("Page verified successfully", map[string]interface{}{
		"page_id":   pageInfo.ID,
		"page_name": pageInfo.Name,
		"category":  pageInfo.Category,
	})

	// Suscribir a webhooks
	if err := s.SubscribeToWebhooks(ctx, pageAccessToken, pageID); err != nil {
		s.logger.Warn("Failed to subscribe to webhooks, continuing without it", map[string]interface{}{
			"page_id": pageID,
			"error":   err.Error(),
		})
	}

	// Crear configuración de la integración
	config := map[string]interface{}{
		"page_access_token": pageAccessToken,
		"page_id":          pageID,
		"webhook_url":      webhookURL,
		"page_name":        pageInfo.Name,
		"page_category":    pageInfo.Category,
		"page_about":       pageInfo.About,
		"page_website":     pageInfo.Website,
		"page_phone":       pageInfo.Phone,
		"page_email":       pageInfo.Email,
	}

	if pageInfo.Picture.Data.URL != "" {
		config["page_picture"] = pageInfo.Picture.Data.URL
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	integration := &domain.ChannelIntegration{
		TenantID:    tenantID,
		Platform:    domain.PlatformMessenger,
		Provider:    domain.ProviderMeta,
		AccessToken: pageAccessToken,
		WebhookURL:  webhookURL,
		Status:      domain.StatusActive,
		Config:      configJSON,
	}

	return integration, nil
}

// ValidateWebhookToken valida el token de verificación del webhook
func (s *MessengerSetupService) ValidateWebhookToken(providedToken, expectedToken string) bool {
	return providedToken == expectedToken
}