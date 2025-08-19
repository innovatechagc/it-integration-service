package domain

import (
	"context"
	"database/sql"
	"time"
)

// ChannelIntegrationRepository define las operaciones de persistencia para integraciones
type ChannelIntegrationRepository interface {
	GetByID(ctx context.Context, id string) (*ChannelIntegration, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ChannelIntegration, error)
	GetByPlatform(ctx context.Context, platform Platform) ([]*ChannelIntegration, error)
	GetActiveByTenant(ctx context.Context, tenantID string) ([]*ChannelIntegration, error)
	Create(ctx context.Context, integration *ChannelIntegration) error
	Update(ctx context.Context, integration *ChannelIntegration) error
	Delete(ctx context.Context, id string) error
	GetByPlatformAndTenant(ctx context.Context, platform Platform, tenantID string) (*ChannelIntegration, error)
	DB() *sql.DB // Para consultas directas
}

// InboundMessageRepository define las operaciones para mensajes entrantes
type InboundMessageRepository interface {
	Create(ctx context.Context, message *InboundMessage) error
	GetUnprocessed(ctx context.Context, limit int) ([]*InboundMessage, error)
	MarkAsProcessed(ctx context.Context, id string) error
}

// OutboundMessageLogRepository define las operaciones para logs de mensajes salientes
type OutboundMessageLogRepository interface {
	Create(ctx context.Context, log *OutboundMessageLog) error
	GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*OutboundMessageLog, error)
	GetByStatus(ctx context.Context, status MessageStatus, limit int) ([]*OutboundMessageLog, error)
	UpdateStatus(ctx context.Context, id string, status MessageStatus, response []byte) error
}

// UserRepository define las operaciones de persistencia para usuarios
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// AuditRepository define las operaciones de persistencia para auditoría
type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
}

// HealthRepository define las operaciones para health checks
type HealthRepository interface {
	CheckDatabase(ctx context.Context) error
	CheckExternalServices(ctx context.Context) map[string]error
}

// GoogleCalendarRepository define las operaciones del repositorio de Google Calendar
type GoogleCalendarRepository interface {
	// Operaciones de integración
	CreateIntegration(ctx context.Context, integration *GoogleCalendarIntegration) error
	GetIntegration(ctx context.Context, channelID string) (*GoogleCalendarIntegration, error)
	GetIntegrationsByTenant(ctx context.Context, tenantID string) ([]*GoogleCalendarIntegration, error)
	UpdateIntegration(ctx context.Context, integration *GoogleCalendarIntegration) error
	DeleteIntegration(ctx context.Context, channelID string) error

	// Operaciones de eventos
	CreateEvent(ctx context.Context, event *CalendarEvent) error
	GetEvent(ctx context.Context, eventID string) (*CalendarEvent, error)
	GetEventsByChannel(ctx context.Context, channelID string) ([]*CalendarEvent, error)
	GetEventsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*CalendarEvent, error)
	UpdateEvent(ctx context.Context, event *CalendarEvent) error
	DeleteEvent(ctx context.Context, eventID string) error
	GetEventsByDateRange(ctx context.Context, channelID string, startTime, endTime time.Time) ([]*CalendarEvent, error)
}
