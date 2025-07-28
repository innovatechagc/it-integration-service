package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/company/microservice-template/internal/domain"
	"github.com/company/microservice-template/pkg/logger"
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
		Platform:    domain.PlatformWhatsApp,
		Sender:      msg.From,
		Recipient:   whatsappPayload.Entry[0].Changes[0].Value.Metadata.PhoneNumberID,
		Content:     content,
		Timestamp:   timestamp,
		MessageID:   msg.ID,
		RawPayload:  payload,
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
		Platform:    domain.PlatformMessenger,
		Sender:      msg.Sender.ID,
		Recipient:   msg.Recipient.ID,
		Content:     content,
		Timestamp:   msg.Timestamp,
		MessageID:   msg.Message.Mid,
		RawPayload:  payload,
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
		Platform:    domain.PlatformTelegram,
		Sender:      strconv.FormatInt(telegramPayload.Message.From.ID, 10),
		Recipient:   strconv.FormatInt(telegramPayload.Message.Chat.ID, 10),
		Content:     content,
		Timestamp:   telegramPayload.Message.Date,
		MessageID:   strconv.FormatInt(telegramPayload.Message.MessageID, 10),
		RawPayload:  payload,
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
		Platform:    domain.PlatformWebchat,
		Sender:      webchatPayload.UserID,
		Recipient:   webchatPayload.SessionID,
		Content:     content,
		Timestamp:   webchatPayload.Timestamp,
		MessageID:   webchatPayload.MessageID,
		RawPayload:  payload,
	}, nil
}

func (s *webhookService) ForwardToMessagingService(ctx context.Context, message *NormalizedMessage) error {
	if s.messagingServiceURL == "" {
		s.logger.Warn("Messaging service URL not configured, skipping forward")
		return nil
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.messagingServiceURL+"/api/v1/messages/inbound", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("messaging service returned error: %d", resp.StatusCode)
	}

	s.logger.Info("Message forwarded to messaging service", map[string]interface{}{
		"message_id": message.MessageID,
		"platform":   message.Platform,
	})

	return nil
}