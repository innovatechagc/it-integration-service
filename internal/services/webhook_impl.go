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

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

type webhookService struct {
	messagingServiceURL string
	logger              logger.Logger
}

// NewWebhookService crea una nueva instancia del servicio de webhook
func NewWebhookService(messagingServiceURL string, logger logger.Logger) WebhookService {
	return &webhookService{
		messagingServiceURL: messagingServiceURL,
		logger:              logger,
	}
}

func (s *webhookService) ValidateSignature(payload []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Remover prefijo si existe (ej: "sha256=")
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (s *webhookService) NormalizeMessage(platform domain.Platform, payload []byte) (*NormalizedMessage, error) {
	switch platform {
	case domain.PlatformWhatsApp:
		return s.normalizeWhatsAppMessage(payload)
	case domain.PlatformMessenger:
		return s.normalizeMessengerMessage(payload)
	case domain.PlatformInstagram:
		return s.normalizeInstagramMessage(payload)
	case domain.PlatformTelegram:
		return s.normalizeTelegramMessage(payload)
	case domain.PlatformWebchat:
		return s.normalizeWebchatMessage(payload)
	case domain.PlatformMailchimp:
		return s.normalizeMailchimpMessage(payload)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (s *webhookService) normalizeWhatsAppMessage(payload []byte) (*NormalizedMessage, error) {
	var whatsappPayload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						ID        string `json:"id"`
						From      string `json:"from"`
						Timestamp string `json:"timestamp"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
						Type string `json:"type"`
					} `json:"messages"`
					Metadata struct {
						PhoneNumberID string `json:"phone_number_id"`
					} `json:"metadata"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(payload, &whatsappPayload); err != nil {
		return nil, fmt.Errorf("failed to parse WhatsApp payload: %w", err)
	}

	if len(whatsappPayload.Entry) == 0 || len(whatsappPayload.Entry[0].Changes) == 0 ||
		len(whatsappPayload.Entry[0].Changes[0].Value.Messages) == 0 {
		return nil, fmt.Errorf("no messages found in WhatsApp payload")
	}

	msg := whatsappPayload.Entry[0].Changes[0].Value.Messages[0]
	timestamp, _ := strconv.ParseInt(msg.Timestamp, 10, 64)

	content := &domain.MessageContent{
		Type: msg.Type,
		Text: msg.Text.Body,
	}

	return &NormalizedMessage{
		Platform:   domain.PlatformWhatsApp,
		Sender:     msg.From,
		Recipient:  whatsappPayload.Entry[0].Changes[0].Value.Metadata.PhoneNumberID,
		Content:    content,
		Timestamp:  timestamp,
		MessageID:  msg.ID,
		RawPayload: payload,
	}, nil
}

func (s *webhookService) normalizeMessengerMessage(payload []byte) (*NormalizedMessage, error) {
	var messengerPayload struct {
		Entry []struct {
			Messaging []struct {
				Sender struct {
					ID string `json:"id"`
				} `json:"sender"`
				Recipient struct {
					ID string `json:"id"`
				} `json:"recipient"`
				Timestamp int64 `json:"timestamp"`
				Message   struct {
					Mid  string `json:"mid"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"messaging"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(payload, &messengerPayload); err != nil {
		return nil, fmt.Errorf("failed to parse Messenger payload: %w", err)
	}

	if len(messengerPayload.Entry) == 0 || len(messengerPayload.Entry[0].Messaging) == 0 {
		return nil, fmt.Errorf("no messages found in Messenger payload")
	}

	msg := messengerPayload.Entry[0].Messaging[0]

	content := &domain.MessageContent{
		Type: "text",
		Text: msg.Message.Text,
	}

	return &NormalizedMessage{
		Platform:   domain.PlatformMessenger,
		Sender:     msg.Sender.ID,
		Recipient:  msg.Recipient.ID,
		Content:    content,
		Timestamp:  msg.Timestamp,
		MessageID:  msg.Message.Mid,
		RawPayload: payload,
	}, nil
}

func (s *webhookService) normalizeInstagramMessage(payload []byte) (*NormalizedMessage, error) {
	// Instagram usa el mismo formato que Messenger
	normalized, err := s.normalizeMessengerMessage(payload)
	if err != nil {
		return nil, err
	}
	normalized.Platform = domain.PlatformInstagram
	return normalized, nil
}

func (s *webhookService) normalizeTelegramMessage(payload []byte) (*NormalizedMessage, error) {
	var telegramPayload struct {
		Message struct {
			MessageID int64 `json:"message_id"`
			From      struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			Date int64  `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	}

	if err := json.Unmarshal(payload, &telegramPayload); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram payload: %w", err)
	}

	content := &domain.MessageContent{
		Type: "text",
		Text: telegramPayload.Message.Text,
	}

	return &NormalizedMessage{
		Platform:   domain.PlatformTelegram,
		Sender:     strconv.FormatInt(telegramPayload.Message.From.ID, 10),
		Recipient:  strconv.FormatInt(telegramPayload.Message.Chat.ID, 10),
		Content:    content,
		Timestamp:  telegramPayload.Message.Date,
		MessageID:  strconv.FormatInt(telegramPayload.Message.MessageID, 10),
		RawPayload: payload,
	}, nil
}

func (s *webhookService) normalizeWebchatMessage(payload []byte) (*NormalizedMessage, error) {
	var webchatPayload struct {
		MessageID string `json:"message_id"`
		UserID    string `json:"user_id"`
		SessionID string `json:"session_id"`
		Text      string `json:"text"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(payload, &webchatPayload); err != nil {
		return nil, fmt.Errorf("failed to parse Webchat payload: %w", err)
	}

	content := &domain.MessageContent{
		Type: "text",
		Text: webchatPayload.Text,
	}

	return &NormalizedMessage{
		Platform:   domain.PlatformWebchat,
		Sender:     webchatPayload.UserID,
		Recipient:  webchatPayload.SessionID,
		Content:    content,
		Timestamp:  webchatPayload.Timestamp,
		MessageID:  webchatPayload.MessageID,
		RawPayload: payload,
	}, nil
}

func (s *webhookService) ForwardToMessagingService(ctx context.Context, message *NormalizedMessage) error {
	if s.messagingServiceURL == "" {
		s.logger.Warn("Messaging service URL not configured, skipping forward")
		return nil
	}

	// Preparar el payload para el servicio de mensajería
	payload := map[string]interface{}{
		"platform":    message.Platform,
		"sender":      message.Sender,
		"recipient":   message.Recipient,
		"content":     message.Content,
		"timestamp":   message.Timestamp,
		"message_id":  message.MessageID,
		"raw_payload": message.RawPayload,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	// Crear request HTTP
	url := s.messagingServiceURL + "/api/v1/webhooks/inbound"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "it-integration-service/1.0")

	// Realizar la llamada HTTP
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to forward message to messaging service: %w", err)
	}
	defer resp.Body.Close()

	// Verificar la respuesta
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Messaging service returned error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"message_id":  message.MessageID,
		})
		return fmt.Errorf("messaging service returned status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("Message forwarded successfully to messaging service", map[string]interface{}{
		"message_id":  message.MessageID,
		"platform":    message.Platform,
		"sender":      message.Sender,
		"status_code": resp.StatusCode,
	})

	return nil
}

func (s *webhookService) normalizeMailchimpMessage(payload []byte) (*NormalizedMessage, error) {
	var mailchimpPayload struct {
		Type    string                 `json:"type"`
		FiredAt string                 `json:"fired_at"`
		Data    map[string]interface{} `json:"data"`
		ListID  string                 `json:"list_id"`
	}

	if err := json.Unmarshal(payload, &mailchimpPayload); err != nil {
		return nil, fmt.Errorf("failed to parse mailchimp payload: %w", err)
	}

	// Extraer información del payload
	var sender, recipient, content string
	var messageType string

	switch mailchimpPayload.Type {
	case "subscribe":
		messageType = "subscription"
		if data, ok := mailchimpPayload.Data["email"].(string); ok {
			recipient = data
		}
		content = "Usuario suscrito a la lista"
	case "unsubscribe":
		messageType = "unsubscription"
		if data, ok := mailchimpPayload.Data["email"].(string); ok {
			recipient = data
		}
		content = "Usuario desuscrito de la lista"
	case "profile":
		messageType = "profile_update"
		if data, ok := mailchimpPayload.Data["email"].(string); ok {
			recipient = data
		}
		content = "Perfil de usuario actualizado"
	case "cleaned":
		messageType = "email_cleaned"
		if data, ok := mailchimpPayload.Data["email"].(string); ok {
			recipient = data
		}
		content = "Email limpiado de la lista"
	case "upemail":
		messageType = "email_changed"
		if data, ok := mailchimpPayload.Data["new_email"].(string); ok {
			recipient = data
		}
		content = "Email de usuario cambiado"
	case "campaign":
		messageType = "campaign_event"
		if data, ok := mailchimpPayload.Data["campaign_id"].(string); ok {
			content = fmt.Sprintf("Evento de campaña: %s", data)
		}
	default:
		messageType = "unknown"
		content = fmt.Sprintf("Evento desconocido: %s", mailchimpPayload.Type)
	}

	// Parsear timestamp
	timestamp := time.Now().Unix()
	if mailchimpPayload.FiredAt != "" {
		if ts, err := time.Parse(time.RFC3339, mailchimpPayload.FiredAt); err == nil {
			timestamp = ts.Unix()
		}
	}

	return &NormalizedMessage{
		Platform:  domain.PlatformMailchimp,
		MessageID: fmt.Sprintf("mailchimp_%s_%d", mailchimpPayload.Type, timestamp),
		Sender:    sender,
		Recipient: recipient,
		Content: &domain.MessageContent{
			Type: messageType,
			Text: content,
		},
		Timestamp:  timestamp,
		RawPayload: payload,
	}, nil
}
