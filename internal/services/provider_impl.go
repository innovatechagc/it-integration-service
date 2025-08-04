package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/company/microservice-template/internal/domain"
	"github.com/company/microservice-template/pkg/logger"
)

type messagingProviderService struct {
	logger logger.Logger
}

// NewMessagingProviderService crea una nueva instancia del servicio de proveedores
func NewMessagingProviderService(logger logger.Logger) MessagingProviderService {
	return &messagingProviderService{
		logger: logger,
	}
}

func (s *messagingProviderService) SendWhatsAppMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	switch integration.Provider {
	case domain.ProviderMeta:
		return s.sendMetaWhatsAppMessage(ctx, integration, recipient, content)
	case domain.Provider360Dialog:
		return s.send360DialogMessage(ctx, integration, recipient, content)
	case domain.ProviderTwilio:
		return s.sendTwilioWhatsAppMessage(ctx, integration, recipient, content)
	default:
		return fmt.Errorf("unsupported WhatsApp provider: %s", integration.Provider)
	}
}

func (s *messagingProviderService) SendMessengerMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	if integration.Provider != domain.ProviderMeta {
		return fmt.Errorf("unsupported Messenger provider: %s", integration.Provider)
	}
	return s.sendMetaMessengerMessage(ctx, integration, recipient, content)
}

func (s *messagingProviderService) SendInstagramMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	if integration.Provider != domain.ProviderMeta {
		return fmt.Errorf("unsupported Instagram provider: %s", integration.Provider)
	}
	return s.sendMetaInstagramMessage(ctx, integration, recipient, content)
}

func (s *messagingProviderService) SendTelegramMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	return s.sendTelegramBotMessage(ctx, integration, recipient, content)
}

func (s *messagingProviderService) SendWebchatMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	return s.sendWebchatMessage(ctx, integration, recipient, content)
}

// Meta WhatsApp Business API
func (s *messagingProviderService) sendMetaWhatsAppMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	var config struct {
		PhoneNumberID string `json:"phone_number_id"`
		BusinessID    string `json:"business_id"`
	}
	
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return fmt.Errorf("failed to parse WhatsApp config: %w", err)
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":               recipient,
		"type":             content.Type,
	}

	if content.Type == "text" {
		payload["text"] = map[string]string{"body": content.Text}
	}

	return s.sendHTTPRequest(ctx, 
		fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", config.PhoneNumberID),
		integration.AccessToken,
		payload,
	)
}

// 360Dialog WhatsApp API
func (s *messagingProviderService) send360DialogMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	payload := map[string]interface{}{
		"to":   recipient,
		"type": content.Type,
	}

	if content.Type == "text" {
		payload["text"] = map[string]string{"body": content.Text}
	}

	return s.sendHTTPRequest(ctx,
		"https://waba.360dialog.io/v1/messages",
		integration.AccessToken,
		payload,
	)
}

// Twilio WhatsApp API
func (s *messagingProviderService) sendTwilioWhatsAppMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	var config struct {
		AccountSID string `json:"account_sid"`
		From       string `json:"from"`
	}
	
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return fmt.Errorf("failed to parse Twilio config: %w", err)
	}

	payload := map[string]interface{}{
		"From": fmt.Sprintf("whatsapp:%s", config.From),
		"To":   fmt.Sprintf("whatsapp:%s", recipient),
		"Body": content.Text,
	}

	return s.sendHTTPRequest(ctx,
		fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", config.AccountSID),
		integration.AccessToken,
		payload,
	)
}

// Meta Messenger API
func (s *messagingProviderService) sendMetaMessengerMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	payload := map[string]interface{}{
		"recipient": map[string]string{"id": recipient},
		"message":   map[string]string{"text": content.Text},
	}

	return s.sendHTTPRequest(ctx,
		"https://graph.facebook.com/v18.0/me/messages",
		integration.AccessToken,
		payload,
	)
}

// Meta Instagram API
func (s *messagingProviderService) sendMetaInstagramMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	payload := map[string]interface{}{
		"recipient": map[string]string{"id": recipient},
		"message":   map[string]string{"text": content.Text},
	}

	return s.sendHTTPRequest(ctx,
		"https://graph.facebook.com/v18.0/me/messages",
		integration.AccessToken,
		payload,
	)
}

// Telegram Bot API
func (s *messagingProviderService) sendTelegramBotMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	var config struct {
		BotToken string `json:"bot_token"`
	}
	
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return fmt.Errorf("failed to parse Telegram config: %w", err)
	}

	payload := map[string]interface{}{
		"chat_id": recipient,
		"text":    content.Text,
	}

	return s.sendHTTPRequest(ctx,
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken),
		"", // Telegram no usa Authorization header
		payload,
	)
}

// Webchat custom API
func (s *messagingProviderService) sendWebchatMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error {
	var config struct {
		WebchatURL string `json:"webchat_url"`
	}
	
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return fmt.Errorf("failed to parse Webchat config: %w", err)
	}

	payload := map[string]interface{}{
		"session_id": recipient,
		"message":    content.Text,
		"type":       content.Type,
	}

	return s.sendHTTPRequest(ctx,
		config.WebchatURL+"/api/messages",
		integration.AccessToken,
		payload,
	)
}

// Helper para enviar requests HTTP
func (s *messagingProviderService) sendHTTPRequest(ctx context.Context, url, token string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Leer el cuerpo de la respuesta para obtener más detalles del error
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		s.logger.Error("Provider API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response_body": errorBody.String(),
			"url": url,
		})
		return fmt.Errorf("provider API returned error: %d - %s", resp.StatusCode, errorBody.String())
	}

	// Leer respuesta exitosa para logging
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	
	s.logger.Info("Message sent successfully", map[string]interface{}{
		"url":           url,
		"status":        resp.StatusCode,
		"response_body": responseBody.String(),
		"payload":       string(jsonData),
	})

	return nil
}