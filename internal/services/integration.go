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
	
	// Envío de mensajes
	SendMessage(ctx context.Context, request *domain.SendMessageRequest) error
	
	// Procesamiento de webhooks
	ProcessWhatsAppWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessMessengerWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessInstagramWebhook(ctx context.Context, payload []byte, signature string) error
	ProcessTelegramWebhook(ctx context.Context, payload []byte) error
	ProcessWebchatWebhook(ctx context.Context, payload []byte) error
	
	// Consulta de mensajes
	GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error)
	GetOutboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.OutboundMessageLog, error)
	GetChatHistory(ctx context.Context, platform, userID string) (*domain.ChatHistory, error)
	
	// Envío masivo
	BroadcastMessage(ctx context.Context, request *domain.BroadcastMessageRequest) (*domain.BroadcastResult, error)
}

// WebhookService define las operaciones para procesamiento de webhooks
type WebhookService interface {
	ValidateSignature(payload []byte, signature string, secret string) bool
	NormalizeMessage(platform domain.Platform, payload []byte) (*NormalizedMessage, error)
	ForwardToMessagingService(ctx context.Context, message *NormalizedMessage) error
}

// MessagingProviderService define las operaciones para proveedores de mensajería
type MessagingProviderService interface {
	SendWhatsAppMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error
	SendMessengerMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error
	SendInstagramMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error
	SendTelegramMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error
	SendWebchatMessage(ctx context.Context, integration *domain.ChannelIntegration, recipient string, content *domain.MessageContent) error
}

// NormalizedMessage representa un mensaje normalizado entre plataformas
type NormalizedMessage struct {
	Platform    domain.Platform `json:"platform"`
	Sender      string          `json:"sender"`
	Recipient   string          `json:"recipient"`
	Content     *domain.MessageContent `json:"content"`
	Timestamp   int64           `json:"timestamp"`
	MessageID   string          `json:"message_id"`
	TenantID    string          `json:"tenant_id"`
	ChannelID   string          `json:"channel_id"`
	RawPayload  json.RawMessage `json:"raw_payload"`
}