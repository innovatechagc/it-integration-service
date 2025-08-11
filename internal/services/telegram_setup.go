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

// TelegramSetupService maneja la configuración específica de Telegram
type TelegramSetupService struct {
	logger logger.Logger
}

// NewTelegramSetupService crea una nueva instancia del servicio de configuración de Telegram
func NewTelegramSetupService(logger logger.Logger) *TelegramSetupService {
	return &TelegramSetupService{
		logger: logger,
	}
}

// TelegramBotInfo representa la información del bot de Telegram
type TelegramBotInfo struct {
	ID                      int64  `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	Username                string `json:"username"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

// TelegramWebhookInfo representa la información del webhook configurado
type TelegramWebhookInfo struct {
	URL                  string   `json:"url"`
	HasCustomCertificate bool     `json:"has_custom_certificate"`
	PendingUpdateCount   int      `json:"pending_update_count"`
	LastErrorDate        int64    `json:"last_error_date,omitempty"`
	LastErrorMessage     string   `json:"last_error_message,omitempty"`
	MaxConnections       int      `json:"max_connections,omitempty"`
	AllowedUpdates       []string `json:"allowed_updates,omitempty"`
}

// TelegramAPIResponse representa una respuesta de la API de Telegram
type TelegramAPIResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

// GetBotInfo obtiene información del bot de Telegram
func (s *TelegramSetupService) GetBotInfo(ctx context.Context, botToken string) (*TelegramBotInfo, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	var botInfo TelegramBotInfo
	if err := json.Unmarshal(apiResp.Result, &botInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bot info: %w", err)
	}

	return &botInfo, nil
}

// SetWebhook configura el webhook del bot de Telegram
func (s *TelegramSetupService) SetWebhook(ctx context.Context, botToken, webhookURL string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", botToken)

	payload := map[string]interface{}{
		"url":                  webhookURL,
		"allowed_updates":      []string{"message", "edited_message", "callback_query"},
		"drop_pending_updates": true,
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
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	defer resp.Body.Close()

	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	s.logger.Info("Telegram webhook configured successfully", map[string]interface{}{
		"webhook_url": webhookURL,
	})

	return nil
}

// GetWebhookInfo obtiene información del webhook configurado
func (s *TelegramSetupService) GetWebhookInfo(ctx context.Context, botToken string) (*TelegramWebhookInfo, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", botToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook info: %w", err)
	}
	defer resp.Body.Close()

	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	var webhookInfo TelegramWebhookInfo
	if err := json.Unmarshal(apiResp.Result, &webhookInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook info: %w", err)
	}

	return &webhookInfo, nil
}

// DeleteWebhook elimina el webhook configurado
func (s *TelegramSetupService) DeleteWebhook(ctx context.Context, botToken string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", botToken)

	payload := map[string]interface{}{
		"drop_pending_updates": true,
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
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	defer resp.Body.Close()

	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	s.logger.Info("Telegram webhook deleted successfully")
	return nil
}

// SendMessage envía un mensaje a través de Telegram
func (s *TelegramSetupService) ValidateBotToken(ctx context.Context, botToken string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate bot token: %w", err)
	}
	defer resp.Body.Close()

	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("invalid bot token: %s", apiResp.Description)
	}

	var botInfo TelegramBotInfo
	if err := json.Unmarshal(apiResp.Result, &botInfo); err == nil {
		s.logger.Info("Bot token validated successfully", map[string]interface{}{
			"bot_id":   botInfo.ID,
			"username": botInfo.Username,
		})
	}

	return nil
}

// CreateTelegramIntegration crea una integración de Telegram con configuración completa
func (s *TelegramSetupService) CreateTelegramIntegration(ctx context.Context, botToken, webhookURL, tenantID string) (*domain.ChannelIntegration, error) {
	// Verificar que el bot funcione
	botInfo, err := s.GetBotInfo(ctx, botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify bot: %w", err)
	}

	s.logger.Info("Bot verified successfully", map[string]interface{}{
		"bot_id":       botInfo.ID,
		"bot_username": botInfo.Username,
		"bot_name":     botInfo.FirstName,
	})

	// Configurar webhook
	if err := s.SetWebhook(ctx, botToken, webhookURL); err != nil {
		return nil, fmt.Errorf("failed to set webhook: %w", err)
	}

	// Crear configuración de la integración
	config := map[string]interface{}{
		"bot_token":    botToken,
		"bot_id":       botInfo.ID,
		"bot_username": botInfo.Username,
		"bot_name":     botInfo.FirstName,
		"webhook_url":  webhookURL,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	integration := &domain.ChannelIntegration{
		TenantID:    tenantID,
		Platform:    domain.PlatformTelegram,
		Provider:    domain.ProviderCustom,
		AccessToken: botToken,
		WebhookURL:  webhookURL,
		Status:      domain.StatusActive,
		Config:      configJSON,
	}

	return integration, nil
}
