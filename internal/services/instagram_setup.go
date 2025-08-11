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

// InstagramSetupService maneja la configuración específica de Instagram
type InstagramSetupService struct {
	logger logger.Logger
}

// NewInstagramSetupService crea una nueva instancia del servicio de configuración de Instagram
func NewInstagramSetupService(logger logger.Logger) *InstagramSetupService {
	return &InstagramSetupService{
		logger: logger,
	}
}

// InstagramAccountInfo representa la información de la cuenta de Instagram Business
type InstagramAccountInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Biography   string `json:"biography"`
	Website     string `json:"website"`
	ProfilePic  string `json:"profile_pic,omitempty"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	MediaCount  int    `json:"media_count"`
	AccountType string `json:"account_type"`
	IsPrivate   bool   `json:"is_private"`
	IsVerified  bool   `json:"is_verified"`
}

// InstagramPageInfo representa la información de la página de Facebook conectada
type InstagramPageInfo struct {
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

// InstagramWebhookSubscription representa una suscripción de webhook
type InstagramWebhookSubscription struct {
	Object string   `json:"object"`
	Fields []string `json:"fields"`
}

// GetInstagramAccountInfo obtiene información de la cuenta de Instagram Business
func (s *InstagramSetupService) GetInstagramAccountInfo(ctx context.Context, pageAccessToken, instagramID string) (*InstagramAccountInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,username,name,biography,website,profile_pic_url,followers_count,follows_count,media_count,account_type,is_private,is_verified&access_token=%s", instagramID, pageAccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get Instagram account info: %w", err)
	}
	defer resp.Body.Close()

	var apiResp MetaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("meta API error: %s", apiResp.Error.Message)
	}

	var accountInfo InstagramAccountInfo
	if err := json.Unmarshal(apiResp.Data, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account info: %w", err)
	}

	return &accountInfo, nil
}

// GetPageInfo obtiene información de la página de Facebook conectada
func (s *InstagramSetupService) GetPageInfo(ctx context.Context, pageAccessToken, pageID string) (*InstagramPageInfo, error) {
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

	var pageInfo InstagramPageInfo
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

// SubscribeToWebhooks suscribe la aplicación a los webhooks de Instagram
func (s *InstagramSetupService) SubscribeToWebhooks(ctx context.Context, pageAccessToken, pageID string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/subscribed_apps", pageID)

	payload := InstagramWebhookSubscription{
		Object: "instagram",
		Fields: []string{"messages", "messaging_postbacks", "messaging_optins"},
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
	req.Header.Set("Authorization", "Bearer "+pageAccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to webhooks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp MetaAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != nil {
			return fmt.Errorf("failed to subscribe to webhooks: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("failed to subscribe to webhooks: status %d", resp.StatusCode)
	}

	return nil
}

// SendMessage envía un mensaje a través de Instagram
func (s *InstagramSetupService) SendMessage(ctx context.Context, pageAccessToken, recipientID, text string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/me/messages?access_token=%s", pageAccessToken)

	payload := map[string]interface{}{
		"recipient": map[string]string{
			"id": recipientID,
		},
		"message": map[string]interface{}{
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp MetaAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != nil {
			return fmt.Errorf("failed to send message: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("failed to send message: status %d", resp.StatusCode)
	}

	return nil
}

// CreateInstagramIntegration crea una integración completa de Instagram
func (s *InstagramSetupService) CreateInstagramIntegration(ctx context.Context, pageAccessToken, instagramID, webhookURL, tenantID string) (*domain.ChannelIntegration, error) {
	// Verificar que la cuenta de Instagram existe
	accountInfo, err := s.GetInstagramAccountInfo(ctx, pageAccessToken, instagramID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Instagram account: %w", err)
	}

	// Suscribir a webhooks
	if err := s.SubscribeToWebhooks(ctx, pageAccessToken, instagramID); err != nil {
		return nil, fmt.Errorf("failed to subscribe to webhooks: %w", err)
	}

	// Crear la integración
	config := map[string]interface{}{
		"page_access_token": pageAccessToken,
		"instagram_id":      instagramID,
		"webhook_url":       webhookURL,
		"username":          accountInfo.Username,
		"account_type":      accountInfo.AccountType,
		"is_verified":       accountInfo.IsVerified,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	integration := &domain.ChannelIntegration{
		ID:          fmt.Sprintf("instagram_%s_%s", tenantID, instagramID),
		Platform:    domain.PlatformInstagram,
		Provider:    domain.ProviderMeta,
		TenantID:    tenantID,
		AccessToken: pageAccessToken,
		WebhookURL:  webhookURL,
		Config:      configJSON,
		Status:      domain.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return integration, nil
}

// ValidateWebhookToken valida el token de verificación del webhook
func (s *InstagramSetupService) ValidateWebhookToken(providedToken, expectedToken string) bool {
	return providedToken == expectedToken
}

// GetInstagramAccounts obtiene la lista de cuentas de Instagram conectadas a una página
func (s *InstagramSetupService) GetInstagramAccounts(ctx context.Context, pageAccessToken, pageID string) ([]InstagramAccountInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/instagram_accounts?access_token=%s", pageID, pageAccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get Instagram accounts: %w", err)
	}
	defer resp.Body.Close()

	var apiResp MetaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("meta API error: %s", apiResp.Error.Message)
	}

	var accounts []InstagramAccountInfo
	if err := json.Unmarshal(apiResp.Data, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return accounts, nil
}
