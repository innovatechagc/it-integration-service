package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// GoogleCalendarRepository implementa el repositorio para Google Calendar
type GoogleCalendarRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewGoogleCalendarRepository crea una nueva instancia del repositorio
func NewGoogleCalendarRepository(db *sql.DB, logger logger.Logger) *GoogleCalendarRepository {
	return &GoogleCalendarRepository{
		db:     db,
		logger: logger,
	}
}

// CreateIntegration crea una nueva integración de Google Calendar
func (r *GoogleCalendarRepository) CreateIntegration(ctx context.Context, integration *domain.GoogleCalendarIntegration) error {
	query := `
		INSERT INTO google_calendar_integrations (
			id, tenant_id, channel_id, calendar_type, calendar_id, calendar_name,
			access_token, refresh_token, token_expiry, webhook_channel, webhook_resource,
			status, config, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		integration.ID,
		integration.TenantID,
		integration.ChannelID,
		integration.CalendarType,
		integration.CalendarID,
		integration.CalendarName,
		integration.AccessToken,
		integration.RefreshToken,
		integration.TokenExpiry,
		integration.WebhookChannel,
		integration.WebhookResource,
		integration.Status,
		configJSON,
		integration.CreatedAt,
		integration.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Error creating Google Calendar integration", err, map[string]interface{}{
			"tenant_id":     integration.TenantID,
			"channel_id":    integration.ChannelID,
			"calendar_type": integration.CalendarType,
		})
		return fmt.Errorf("error creating integration: %w", err)
	}

	r.logger.Info("Google Calendar integration created", map[string]interface{}{
		"integration_id": integration.ID,
		"tenant_id":      integration.TenantID,
		"channel_id":     integration.ChannelID,
	})

	return nil
}

// GetIntegration obtiene una integración por channel_id
func (r *GoogleCalendarRepository) GetIntegration(ctx context.Context, channelID string) (*domain.GoogleCalendarIntegration, error) {
	query := `
		SELECT id, tenant_id, channel_id, calendar_type, calendar_id, calendar_name,
			   access_token, refresh_token, token_expiry, webhook_channel, webhook_resource,
			   status, config, created_at, updated_at
		FROM google_calendar_integrations
		WHERE channel_id = $1 AND deleted_at IS NULL
	`

	var integration domain.GoogleCalendarIntegration
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, channelID).Scan(
		&integration.ID,
		&integration.TenantID,
		&integration.ChannelID,
		&integration.CalendarType,
		&integration.CalendarID,
		&integration.CalendarName,
		&integration.AccessToken,
		&integration.RefreshToken,
		&integration.TokenExpiry,
		&integration.WebhookChannel,
		&integration.WebhookResource,
		&integration.Status,
		&configJSON,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("integration not found: %s", channelID)
		}
		return nil, fmt.Errorf("error getting integration: %w", err)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			r.logger.Error("Error unmarshaling config", err, map[string]interface{}{
				"channel_id": channelID,
			})
		}
	}

	return &integration, nil
}

// GetIntegrationsByTenant obtiene todas las integraciones de un tenant
func (r *GoogleCalendarRepository) GetIntegrationsByTenant(ctx context.Context, tenantID string) ([]*domain.GoogleCalendarIntegration, error) {
	query := `
		SELECT id, tenant_id, channel_id, calendar_type, calendar_id, calendar_name,
			   access_token, refresh_token, token_expiry, webhook_channel, webhook_resource,
			   status, config, created_at, updated_at
		FROM google_calendar_integrations
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error querying integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*domain.GoogleCalendarIntegration

	for rows.Next() {
		var integration domain.GoogleCalendarIntegration
		var configJSON []byte

		err := rows.Scan(
			&integration.ID,
			&integration.TenantID,
			&integration.ChannelID,
			&integration.CalendarType,
			&integration.CalendarID,
			&integration.CalendarName,
			&integration.AccessToken,
			&integration.RefreshToken,
			&integration.TokenExpiry,
			&integration.WebhookChannel,
			&integration.WebhookResource,
			&integration.Status,
			&configJSON,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Error scanning integration", err, nil)
			continue
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
				r.logger.Error("Error unmarshaling config", err, map[string]interface{}{
					"channel_id": integration.ChannelID,
				})
			}
		}

		integrations = append(integrations, &integration)
	}

	return integrations, nil
}

// UpdateIntegration actualiza una integración existente
func (r *GoogleCalendarRepository) UpdateIntegration(ctx context.Context, integration *domain.GoogleCalendarIntegration) error {
	query := `
		UPDATE google_calendar_integrations
		SET calendar_type = $1, calendar_id = $2, calendar_name = $3,
			access_token = $4, refresh_token = $5, token_expiry = $6,
			webhook_channel = $7, webhook_resource = $8, status = $9,
			config = $10, updated_at = $11
		WHERE channel_id = $12 AND deleted_at IS NULL
	`

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		integration.CalendarType,
		integration.CalendarID,
		integration.CalendarName,
		integration.AccessToken,
		integration.RefreshToken,
		integration.TokenExpiry,
		integration.WebhookChannel,
		integration.WebhookResource,
		integration.Status,
		configJSON,
		time.Now(),
		integration.ChannelID,
	)

	if err != nil {
		r.logger.Error("Error updating Google Calendar integration", err, map[string]interface{}{
			"channel_id": integration.ChannelID,
		})
		return fmt.Errorf("error updating integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("integration not found: %s", integration.ChannelID)
	}

	r.logger.Info("Google Calendar integration updated", map[string]interface{}{
		"channel_id": integration.ChannelID,
		"status":     integration.Status,
	})

	return nil
}

// DeleteIntegration elimina una integración (soft delete)
func (r *GoogleCalendarRepository) DeleteIntegration(ctx context.Context, channelID string) error {
	query := `
		UPDATE google_calendar_integrations
		SET deleted_at = $1, status = 'disabled'
		WHERE channel_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), channelID)
	if err != nil {
		r.logger.Error("Error deleting Google Calendar integration", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error deleting integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("integration not found: %s", channelID)
	}

	r.logger.Info("Google Calendar integration deleted", map[string]interface{}{
		"channel_id": channelID,
	})

	return nil
}

// CreateEvent crea un nuevo evento de calendario
func (r *GoogleCalendarRepository) CreateEvent(ctx context.Context, event *domain.CalendarEvent) error {
	query := `
		INSERT INTO calendar_events (
			id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			location, start_time, end_time, all_day, attendees, recurrence, status,
			visibility, reminders, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	attendeesJSON, err := json.Marshal(event.Attendees)
	if err != nil {
		return fmt.Errorf("error marshaling attendees: %w", err)
	}

	recurrenceJSON, err := json.Marshal(event.Recurrence)
	if err != nil {
		return fmt.Errorf("error marshaling recurrence: %w", err)
	}

	remindersJSON, err := json.Marshal(event.Reminders)
	if err != nil {
		return fmt.Errorf("error marshaling reminders: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.TenantID,
		event.ChannelID,
		event.GoogleID,
		event.CalendarID,
		event.Summary,
		event.Description,
		event.Location,
		event.StartTime,
		event.EndTime,
		event.AllDay,
		attendeesJSON,
		recurrenceJSON,
		event.Status,
		event.Visibility,
		remindersJSON,
		event.CreatedAt,
		event.UpdatedAt,
		nil, // deleted_at
	)

	if err != nil {
		r.logger.Error("Error creating calendar event", err, map[string]interface{}{
			"event_id":   event.ID,
			"google_id":  event.GoogleID,
			"channel_id": event.ChannelID,
		})
		return fmt.Errorf("error creating event: %w", err)
	}

	// Crear registro de auditoría
	r.createEventAuditLog(ctx, event.ID, "created", nil, event)

	r.logger.Info("Calendar event created", map[string]interface{}{
		"event_id":   event.ID,
		"google_id":  event.GoogleID,
		"summary":    event.Summary,
		"channel_id": event.ChannelID,
	})

	return nil
}

// GetEvent obtiene un evento por ID
func (r *GoogleCalendarRepository) GetEvent(ctx context.Context, eventID string) (*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE id = $1 AND deleted_at IS NULL
	`

	var event domain.CalendarEvent
	var attendeesJSON, recurrenceJSON, remindersJSON []byte

	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.TenantID,
		&event.ChannelID,
		&event.GoogleID,
		&event.CalendarID,
		&event.Summary,
		&event.Description,
		&event.Location,
		&event.StartTime,
		&event.EndTime,
		&event.AllDay,
		&attendeesJSON,
		&recurrenceJSON,
		&event.Status,
		&event.Visibility,
		&remindersJSON,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %s", eventID)
		}
		return nil, fmt.Errorf("error getting event: %w", err)
	}

	// Parsear JSON fields
	if err := r.parseEventJSONFields(&event, attendeesJSON, recurrenceJSON, remindersJSON); err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEventsByChannel obtiene eventos por channel_id con paginación
func (r *GoogleCalendarRepository) GetEventsByChannel(ctx context.Context, channelID string, limit, offset int) ([]*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE channel_id = $1 AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetEventsByTenant obtiene eventos por tenant con paginación
func (r *GoogleCalendarRepository) GetEventsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// UpdateEvent actualiza un evento existente
func (r *GoogleCalendarRepository) UpdateEvent(ctx context.Context, eventID string, event *domain.CalendarEvent) error {
	// Obtener evento actual para auditoría
	oldEvent, err := r.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("error getting old event for audit: %w", err)
	}

	query := `
		UPDATE calendar_events
		SET summary = $1, description = $2, location = $3, start_time = $4, end_time = $5,
			all_day = $6, attendees = $7, recurrence = $8, status = $9, visibility = $10,
			reminders = $11, updated_at = $12
		WHERE id = $13 AND deleted_at IS NULL
	`

	attendeesJSON, err := json.Marshal(event.Attendees)
	if err != nil {
		return fmt.Errorf("error marshaling attendees: %w", err)
	}

	recurrenceJSON, err := json.Marshal(event.Recurrence)
	if err != nil {
		return fmt.Errorf("error marshaling recurrence: %w", err)
	}

	remindersJSON, err := json.Marshal(event.Reminders)
	if err != nil {
		return fmt.Errorf("error marshaling reminders: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		event.Summary,
		event.Description,
		event.Location,
		event.StartTime,
		event.EndTime,
		event.AllDay,
		attendeesJSON,
		recurrenceJSON,
		event.Status,
		event.Visibility,
		remindersJSON,
		time.Now(),
		eventID,
	)

	if err != nil {
		r.logger.Error("Error updating calendar event", err, map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("error updating event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	// Crear registro de auditoría
	r.createEventAuditLog(ctx, eventID, "updated", oldEvent, event)

	r.logger.Info("Calendar event updated", map[string]interface{}{
		"event_id": eventID,
		"summary":  event.Summary,
	})

	return nil
}

// DeleteEvent elimina un evento (soft delete)
func (r *GoogleCalendarRepository) DeleteEvent(ctx context.Context, eventID string) error {
	// Obtener evento para auditoría
	oldEvent, err := r.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("error getting event for audit: %w", err)
	}

	query := `
		UPDATE calendar_events
		SET deleted_at = $1, status = 'cancelled'
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), eventID)
	if err != nil {
		r.logger.Error("Error deleting calendar event", err, map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("error deleting event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	// Crear registro de auditoría
	r.createEventAuditLog(ctx, eventID, "deleted", oldEvent, nil)

	r.logger.Info("Calendar event deleted", map[string]interface{}{
		"event_id": eventID,
		"summary":  oldEvent.Summary,
	})

	return nil
}

// GetEventsByDateRange obtiene eventos en un rango de fechas
func (r *GoogleCalendarRepository) GetEventsByDateRange(ctx context.Context, channelID string, startTime, endTime time.Time) ([]*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE channel_id = $1 
		  AND deleted_at IS NULL
		  AND (
			(start_time >= $2 AND start_time <= $3) OR
			(end_time >= $2 AND end_time <= $3) OR
			(start_time <= $2 AND end_time >= $3)
		  )
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("error querying events by date range: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetUpcomingEvents obtiene eventos próximos
func (r *GoogleCalendarRepository) GetUpcomingEvents(ctx context.Context, channelID string, hours int) ([]*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE channel_id = $1 
		  AND deleted_at IS NULL
		  AND start_time >= NOW()
		  AND start_time <= NOW() + INTERVAL '1 hour' * $2
		  AND status = 'confirmed'
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, hours)
	if err != nil {
		return nil, fmt.Errorf("error querying upcoming events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetEventsByAttendee obtiene eventos por asistente
func (r *GoogleCalendarRepository) GetEventsByAttendee(ctx context.Context, channelID, attendeeEmail string) ([]*domain.CalendarEvent, error) {
	query := `
		SELECT id, tenant_id, channel_id, google_id, calendar_id, summary, description,
			   location, start_time, end_time, all_day, attendees, recurrence, status,
			   visibility, reminders, created_at, updated_at
		FROM calendar_events
		WHERE channel_id = $1 
		  AND deleted_at IS NULL
		  AND attendees::text LIKE '%' || $2 || '%'
		ORDER BY start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, attendeeEmail)
	if err != nil {
		return nil, fmt.Errorf("error querying events by attendee: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetEventStats obtiene estadísticas de eventos
func (r *GoogleCalendarRepository) GetEventStats(ctx context.Context, tenantID string) (*domain.EventStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN start_time >= NOW() THEN 1 END) as upcoming_events,
			COUNT(CASE WHEN end_time < NOW() THEN 1 END) as past_events,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_events,
			COUNT(DISTINCT channel_id) as active_channels
		FROM calendar_events
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`

	var stats domain.EventStats
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&stats.TotalEvents,
		&stats.UpcomingEvents,
		&stats.PastEvents,
		&stats.CancelledEvents,
		&stats.ActiveChannels,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting event stats: %w", err)
	}

	return &stats, nil
}

// Helper methods

// scanEvents escanea múltiples eventos desde rows
func (r *GoogleCalendarRepository) scanEvents(rows *sql.Rows) ([]*domain.CalendarEvent, error) {
	var events []*domain.CalendarEvent

	for rows.Next() {
		var event domain.CalendarEvent
		var attendeesJSON, recurrenceJSON, remindersJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&event.ChannelID,
			&event.GoogleID,
			&event.CalendarID,
			&event.Summary,
			&event.Description,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.AllDay,
			&attendeesJSON,
			&recurrenceJSON,
			&event.Status,
			&event.Visibility,
			&remindersJSON,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Error scanning event", err, nil)
			continue
		}

		if err := r.parseEventJSONFields(&event, attendeesJSON, recurrenceJSON, remindersJSON); err != nil {
			r.logger.Error("Error parsing event JSON fields", err, map[string]interface{}{
				"event_id": event.ID,
			})
			continue
		}

		events = append(events, &event)
	}

	return events, nil
}

// parseEventJSONFields parsea los campos JSON de un evento
func (r *GoogleCalendarRepository) parseEventJSONFields(event *domain.CalendarEvent, attendeesJSON, recurrenceJSON, remindersJSON []byte) error {
	if len(attendeesJSON) > 0 {
		if err := json.Unmarshal(attendeesJSON, &event.Attendees); err != nil {
			return fmt.Errorf("error unmarshaling attendees: %w", err)
		}
	}

	if len(recurrenceJSON) > 0 {
		if err := json.Unmarshal(recurrenceJSON, &event.Recurrence); err != nil {
			return fmt.Errorf("error unmarshaling recurrence: %w", err)
		}
	}

	if len(remindersJSON) > 0 {
		if err := json.Unmarshal(remindersJSON, &event.Reminders); err != nil {
			return fmt.Errorf("error unmarshaling reminders: %w", err)
		}
	}

	return nil
}

// createEventAuditLog crea un registro de auditoría para cambios en eventos
func (r *GoogleCalendarRepository) createEventAuditLog(ctx context.Context, eventID, action string, oldEvent, newEvent *domain.CalendarEvent) {
	// TODO: Implementar tabla de auditoría si es necesaria
	// Por ahora solo loggeamos la acción
	r.logger.Info("Event audit log", map[string]interface{}{
		"event_id": eventID,
		"action":   action,
		"old_summary": func() string {
			if oldEvent != nil {
				return oldEvent.Summary
			}
			return ""
		}(),
		"new_summary": func() string {
			if newEvent != nil {
				return newEvent.Summary
			}
			return ""
		}(),
		"timestamp": time.Now(),
	})
}

// CleanupOldEvents limpia eventos antiguos (opcional)
func (r *GoogleCalendarRepository) CleanupOldEvents(ctx context.Context, daysToKeep int) (int, error) {
	query := `
		UPDATE calendar_events
		SET deleted_at = NOW()
		WHERE deleted_at IS NULL
		  AND end_time < NOW() - INTERVAL '1 day' * $1
		  AND status = 'cancelled'
	`

	result, err := r.db.ExecContext(ctx, query, daysToKeep)
	if err != nil {
		return 0, fmt.Errorf("error cleaning up old events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %w", err)
	}

	r.logger.Info("Old events cleaned up", map[string]interface{}{
		"deleted_count": rowsAffected,
		"days_to_keep":  daysToKeep,
	})

	return int(rowsAffected), nil
}
