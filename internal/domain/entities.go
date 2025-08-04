package domain

import (
	"time"
	"encoding/json"
)

// ChannelIntegration representa una integración de canal de mensajería
type ChannelIntegration struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Platform    Platform               `json:"platform" db:"platform"`
	Provider    Provider               `json:"provider" db:"provider"`
	AccessToken string                 `json:"access_token,omitempty" db:"access_token"` // Encrypted, allow receiving but don't always show
	WebhookURL  string                 `json:"webhook_url" db:"webhook_url"`
	Status      IntegrationStatus      `json:"status" db:"status"`
	Config      json.RawMessage        `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// InboundMessage representa un mensaje entrante para logs/debug
type InboundMessage struct {
	ID         string          `json:"id" db:"id"`
	Platform   Platform        `json:"platform" db:"platform"`
	Payload    json.RawMessage `json:"payload" db:"payload"`
	ReceivedAt time.Time       `json:"received_at" db:"received_at"`
	Processed  bool            `json:"processed" db:"processed"`
}

// OutboundMessageLog representa el log de mensajes salientes
type OutboundMessageLog struct {
	ID        string          `json:"id" db:"id"`
	ChannelID string          `json:"channel_id" db:"channel_id"`
	Recipient string          `json:"recipient" db:"recipient"`
	Content   json.RawMessage `json:"content" db:"content"`
	Status    MessageStatus   `json:"status" db:"status"`
	Response  json.RawMessage `json:"response" db:"response"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
}

// SendMessageRequest representa una solicitud de envío de mensaje
type SendMessageRequest struct {
	ChannelID string      `json:"channel_id" binding:"required"`
	Recipient string      `json:"recipient" binding:"required"`
	Content   MessageContent `json:"content" binding:"required"`
}

// MessageContent representa el contenido de un mensaje
type MessageContent struct {
	Type string `json:"type" binding:"required"`
	Text string `json:"text,omitempty"`
	// Otros campos para diferentes tipos de contenido
	Media *MediaContent `json:"media,omitempty"`
}

// MediaContent representa contenido multimedia
type MediaContent struct {
	URL      string `json:"url"`
	Caption  string `json:"caption,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// Platform enum para plataformas de mensajería
type Platform string

const (
	PlatformWhatsApp  Platform = "whatsapp"
	PlatformMessenger Platform = "messenger"
	PlatformInstagram Platform = "instagram"
	PlatformTelegram  Platform = "telegram"
	PlatformWebchat   Platform = "webchat"
)

// Provider enum para proveedores de servicios
type Provider string

const (
	ProviderMeta      Provider = "meta"
	ProviderTwilio    Provider = "twilio"
	Provider360Dialog Provider = "360dialog"
	ProviderCustom    Provider = "custom"
)

// IntegrationStatus enum para estado de integración
type IntegrationStatus string

const (
	StatusActive   IntegrationStatus = "active"
	StatusDisabled IntegrationStatus = "disabled"
	StatusError    IntegrationStatus = "error"
)

// MessageStatus enum para estado de mensajes
type MessageStatus string

const (
	MessageStatusSent   MessageStatus = "sent"
	MessageStatusFailed MessageStatus = "failed"
	MessageStatusQueued MessageStatus = "queued"
)

// User representa un usuario del sistema
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Roles     []string  `json:"roles" db:"roles"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog representa un registro de auditoría
type AuditLog struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Action    string                 `json:"action" db:"action"`
	Resource  string                 `json:"resource" db:"resource"`
	Details   map[string]interface{} `json:"details" db:"details"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// APIResponse estructura estándar para respuestas de API
type APIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// HealthStatus representa el estado de salud del servicio
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}