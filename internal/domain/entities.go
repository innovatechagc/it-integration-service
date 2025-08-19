package domain

import (
	"encoding/json"
	"time"
)

// ChannelIntegration representa una integración de canal de mensajería
type ChannelIntegration struct {
	ID          string            `json:"id" db:"id"`
	TenantID    string            `json:"tenant_id" db:"tenant_id"`
	Platform    Platform          `json:"platform" db:"platform"`
	Provider    Provider          `json:"provider" db:"provider"`
	AccessToken string            `json:"access_token,omitempty" db:"access_token"` // Encrypted, allow receiving but don't always show
	WebhookURL  string            `json:"webhook_url" db:"webhook_url"`
	Status      IntegrationStatus `json:"status" db:"status"`
	Config      json.RawMessage   `json:"config" db:"config"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
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
	ChannelID string         `json:"channel_id" binding:"required"`
	Recipient string         `json:"recipient" binding:"required"`
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
	PlatformWhatsApp       Platform = "whatsapp"
	PlatformMessenger      Platform = "messenger"
	PlatformInstagram      Platform = "instagram"
	PlatformTelegram       Platform = "telegram"
	PlatformWebchat        Platform = "webchat"
	PlatformMailchimp      Platform = "mailchimp"
	PlatformGoogleCalendar Platform = "google_calendar"
)

// Provider enum para proveedores de servicios
type Provider string

const (
	ProviderMeta      Provider = "meta"
	ProviderTwilio    Provider = "twilio"
	Provider360Dialog Provider = "360dialog"
	ProviderCustom    Provider = "custom"
	ProviderMailchimp Provider = "mailchimp"
	ProviderGoogle    Provider = "google"
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

// CalendarType enum para tipos de calendario de Google
type CalendarType string

const (
	CalendarTypePersonal CalendarType = "personal"
	CalendarTypeWork     CalendarType = "work"
	CalendarTypeShared   CalendarType = "shared"
)

// EventStats representa estadísticas de eventos
type EventStats struct {
	TotalEvents     int `json:"total_events"`
	UpcomingEvents  int `json:"upcoming_events"`
	PastEvents      int `json:"past_events"`
	CancelledEvents int `json:"cancelled_events"`
	ActiveChannels  int `json:"active_channels"`
}

// CalendarEvent representa un evento de calendario
type CalendarEvent struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenant_id"`
	ChannelID   string             `json:"channel_id"`
	GoogleID    string             `json:"google_id"`
	CalendarID  string             `json:"calendar_id"`
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Location    string             `json:"location"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     time.Time          `json:"end_time"`
	AllDay      bool               `json:"all_day"`
	Attendees   []CalendarAttendee `json:"attendees"`
	Recurrence  *EventRecurrence   `json:"recurrence,omitempty"`
	Status      EventStatus        `json:"status"`
	Visibility  EventVisibility    `json:"visibility"`
	Reminders   []EventReminder    `json:"reminders"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty"` // Soft delete
}

// CalendarAttendee representa un asistente a un evento
type CalendarAttendee struct {
	Email          string `json:"email" db:"email"`
	Name           string `json:"name" db:"name"`
	ResponseStatus string `json:"response_status" db:"response_status"` // accepted, declined, tentative, needsAction
	Organizer      bool   `json:"organizer" db:"organizer"`
	Self           bool   `json:"self" db:"self"`
}

// EventRecurrence representa la recurrencia de un evento
type EventRecurrence struct {
	Frequency  string     `json:"frequency" db:"frequency"`                 // daily, weekly, monthly, yearly
	Interval   int        `json:"interval" db:"interval"`                   // cada cuántos días/semanas/meses/años
	Count      int        `json:"count" db:"count"`                         // número de ocurrencias
	Until      *time.Time `json:"until,omitempty" db:"until"`               // fecha hasta cuándo
	ByDay      []string   `json:"by_day,omitempty" db:"by_day"`             // días de la semana (MO, TU, WE, etc.)
	ByMonth    []int      `json:"by_month,omitempty" db:"by_month"`         // meses del año
	ByMonthDay []int      `json:"by_month_day,omitempty" db:"by_month_day"` // días del mes
}

// EventStatus enum para estado de eventos
type EventStatus string

const (
	EventStatusConfirmed EventStatus = "confirmed"
	EventStatusTentative EventStatus = "tentative"
	EventStatusCancelled EventStatus = "cancelled"
)

// EventVisibility enum para visibilidad de eventos
type EventVisibility string

const (
	EventVisibilityDefault EventVisibility = "default"
	EventVisibilityPublic  EventVisibility = "public"
	EventVisibilityPrivate EventVisibility = "private"
)

// EventReminder representa un recordatorio de evento
type EventReminder struct {
	Method  string `json:"method" db:"method"`   // email, popup, sms
	Minutes int    `json:"minutes" db:"minutes"` // minutos antes del evento
}

// GoogleCalendarIntegration representa una integración de Google Calendar
type GoogleCalendarIntegration struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	ChannelID       string                 `json:"channel_id"`
	CalendarType    CalendarType           `json:"calendar_type"`
	CalendarID      string                 `json:"calendar_id"`
	CalendarName    string                 `json:"calendar_name"`
	AccessToken     string                 `json:"access_token"`
	RefreshToken    string                 `json:"refresh_token"`
	TokenExpiry     time.Time              `json:"token_expiry"`
	WebhookChannel  string                 `json:"webhook_channel"`
	WebhookResource string                 `json:"webhook_resource"`
	Status          IntegrationStatus      `json:"status"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty"` // Soft delete
}

// CreateEventRequest representa una solicitud de creación de evento
type CreateEventRequest struct {
	TenantID    string             `json:"tenant_id" binding:"required"`
	ChannelID   string             `json:"channel_id" binding:"required"`
	CalendarID  string             `json:"calendar_id" binding:"required"`
	Summary     string             `json:"summary" binding:"required"`
	Description string             `json:"description"`
	Location    string             `json:"location"`
	StartTime   time.Time          `json:"start_time" binding:"required"`
	EndTime     time.Time          `json:"end_time" binding:"required"`
	AllDay      bool               `json:"all_day"`
	Attendees   []CalendarAttendee `json:"attendees"`
	Recurrence  *EventRecurrence   `json:"recurrence"`
	Visibility  EventVisibility    `json:"visibility"`
	Reminders   []EventReminder    `json:"reminders"`
}

// UpdateEventRequest representa una solicitud de actualización de evento
type UpdateEventRequest struct {
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Location    string             `json:"location"`
	StartTime   *time.Time         `json:"start_time"`
	EndTime     *time.Time         `json:"end_time"`
	AllDay      *bool              `json:"all_day"`
	Attendees   []CalendarAttendee `json:"attendees"`
	Recurrence  *EventRecurrence   `json:"recurrence"`
	Visibility  EventVisibility    `json:"visibility"`
	Reminders   []EventReminder    `json:"reminders"`
}

// ListEventsRequest representa una solicitud de listado de eventos
type ListEventsRequest struct {
	TenantID   string     `json:"tenant_id" binding:"required"`
	ChannelID  string     `json:"channel_id" binding:"required"`
	CalendarID string     `json:"calendar_id"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	MaxResults int        `json:"max_results"`
	PageToken  string     `json:"page_token"`
}

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

// ChatMessage representa un mensaje en una conversación
type ChatMessage struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "inbound" o "outbound"
	Platform  Platform  `json:"platform"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status,omitempty"` // Para mensajes outbound
}

// ChatHistory representa el historial de conversación con un usuario
type ChatHistory struct {
	Platform   Platform      `json:"platform"`
	UserID     string        `json:"user_id"`
	Messages   []ChatMessage `json:"messages"`
	TotalCount int           `json:"total_count"`
}

// BroadcastMessageRequest representa una solicitud de mensaje masivo
type BroadcastMessageRequest struct {
	TenantID   string         `json:"tenant_id" binding:"required"`
	Platforms  []Platform     `json:"platforms" binding:"required"`
	Recipients []string       `json:"recipients" binding:"required"`
	Content    MessageContent `json:"content" binding:"required"`
}

// BroadcastResult representa el resultado de un envío masivo
type BroadcastResult struct {
	TotalSent   int                   `json:"total_sent"`
	TotalFailed int                   `json:"total_failed"`
	Results     []BroadcastItemResult `json:"results"`
}

// BroadcastItemResult representa el resultado de un envío individual
type BroadcastItemResult struct {
	Platform  Platform `json:"platform"`
	Recipient string   `json:"recipient"`
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	MessageID string   `json:"message_id,omitempty"`
}
