package services

import (
	"context"
	"encoding/json"

	"it-integration-service/internal/domain"
)

// IntegrationService define las operaciones del servicio de integración
type IntegrationService interface {
	// Gestión de canales
	CreateChannel(ctx context.Context, integration *domain.ChannelIntegration) error
	GetChannel(ctx context.Context, id string) (*domain.ChannelIntegration, error)
	GetChannelsByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error)
	UpdateChannel(ctx context.Context, integration *domain.ChannelIntegration) error
	DeleteChannel(ctx context.Context, id string) error

	// Procesamiento de webhooks (solo recepción)
	ProcessWhatsAppWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessMessengerWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessInstagramWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessTelegramWebhook(ctx context.Context, payload []byte) error
	ProcessWebchatWebhook(ctx context.Context, payload []byte) error

	// Consulta de mensajes entrantes (solo para validación)
	GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error)
}

// WebhookService define las operaciones para procesamiento de webhooks
type WebhookService interface {
	ValidateSignature(payload []byte, signature string, secret string) bool
	NormalizeMessage(platform domain.Platform, payload []byte) (*NormalizedMessage, error)
	ForwardToMessagingService(ctx context.Context, message *NormalizedMessage) error
}

// NormalizedMessage representa un mensaje normalizado entre plataformas
type NormalizedMessage struct {
	Platform   domain.Platform        `json:"platform"`
	Sender     string                 `json:"sender"`
	Recipient  string                 `json:"recipient"`
	Content    *domain.MessageContent `json:"content"`
	Timestamp  int64                  `json:"timestamp"`
	MessageID  string                 `json:"message_id"`
	TenantID   string                 `json:"tenant_id"`
	ChannelID  string                 `json:"channel_id"`
	RawPayload json.RawMessage        `json:"raw_payload"`
}
