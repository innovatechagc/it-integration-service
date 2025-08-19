package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/repository"
	"it-integration-service/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarService maneja las operaciones de eventos de Google Calendar
type GoogleCalendarService struct {
	config     *config.GoogleCalendarConfig
	setupSvc   *GoogleCalendarSetupService
	repo       repository.GoogleCalendarRepository
	logger     logger.Logger
	encryption *EncryptionService
}

// EventListResponse representa la respuesta de listado de eventos
type EventListResponse struct {
	Events        []*domain.CalendarEvent `json:"events"`
	NextPageToken string                  `json:"next_page_token,omitempty"`
	TotalEvents   int                     `json:"total_events"`
}

// SyncResult representa el resultado de una sincronización
type SyncResult struct {
	Created   int      `json:"created"`
	Updated   int      `json:"updated"`
	Deleted   int      `json:"deleted"`
	Errors    int      `json:"errors"`
	ErrorList []string `json:"error_list,omitempty"`
}

// NotificationConfig configura las notificaciones para eventos
type NotificationConfig struct {
	SendEmail       bool  `json:"send_email"`
	SendSMS         bool  `json:"send_sms"`
	SendWhatsApp    bool  `json:"send_whatsapp"`
	SendTelegram    bool  `json:"send_telegram"`
	ReminderMinutes []int `json:"reminder_minutes"` // minutos antes del evento
}

// NewGoogleCalendarService crea una nueva instancia del servicio
func NewGoogleCalendarService(cfg *config.GoogleCalendarConfig, setupSvc *GoogleCalendarSetupService, repo repository.GoogleCalendarRepository, logger logger.Logger, encryption *EncryptionService) *GoogleCalendarService {
	return &GoogleCalendarService{
		config:     cfg,
		setupSvc:   setupSvc,
		repo:       repo,
		logger:     logger,
		encryption: encryption,
	}
}

// CreateEvent crea un nuevo evento en Google Calendar
func (s *GoogleCalendarService) CreateEvent(ctx context.Context, req *domain.CreateEventRequest) (*domain.CalendarEvent, error) {
	s.logger.Info("Creando evento en Google Calendar", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"channel_id":  req.ChannelID,
		"calendar_id": req.CalendarID,
		"summary":     req.Summary,
	})

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, req.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.setupSvc.createOAuth2Client(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Convertir request a evento de Google Calendar
	googleEvent := s.convertToGoogleEvent(req)

	// Crear evento en Google Calendar
	createdEvent, err := calendarService.Events.Insert(req.CalendarID, googleEvent).Do()
	if err != nil {
		s.logger.Error("Error al crear evento en Google Calendar", err, map[string]interface{}{
			"calendar_id": req.CalendarID,
			"summary":     req.Summary,
		})
		return nil, fmt.Errorf("error al crear evento en Google Calendar: %w", err)
	}

	// Convertir respuesta a dominio
	event := s.convertFromGoogleEvent(createdEvent, req.TenantID, req.ChannelID, req.CalendarID)

	// Guardar evento en base de datos local
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	err = s.repo.CreateEvent(ctx, event)
	if err != nil {
		s.logger.Error("Error al guardar evento en base de datos", err, map[string]interface{}{
			"event_id": event.ID,
		})
		// No fallar si no se puede guardar localmente
	}

	// Configurar notificaciones si se especifican
	if len(req.Reminders) > 0 {
		err = s.setupEventNotifications(ctx, event, req.Reminders)
		if err != nil {
			s.logger.Warn("Error al configurar notificaciones", map[string]interface{}{
				"event_id": event.ID,
				"error":    err.Error(),
			})
		}
	}

	s.logger.Info("Evento creado exitosamente", map[string]interface{}{
		"event_id":   event.ID,
		"google_id":  event.GoogleID,
		"summary":    event.Summary,
		"start_time": event.StartTime,
	})

	return event, nil
}

// UpdateEvent actualiza un evento existente
func (s *GoogleCalendarService) UpdateEvent(ctx context.Context, eventID string, req *domain.UpdateEventRequest) (*domain.CalendarEvent, error) {
	s.logger.Info("Actualizando evento en Google Calendar", map[string]interface{}{
		"event_id": eventID,
	})

	// Obtener evento de base de datos local
	event, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener evento: %w", err)
	}

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, event.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.setupSvc.createOAuth2Client(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Obtener evento actual de Google Calendar
	googleEvent, err := calendarService.Events.Get(event.CalendarID, event.GoogleID).Do()
	if err != nil {
		return nil, fmt.Errorf("error al obtener evento de Google Calendar: %w", err)
	}

	// Actualizar campos del evento
	s.updateGoogleEvent(googleEvent, req)

	// Actualizar evento en Google Calendar
	updatedEvent, err := calendarService.Events.Update(event.CalendarID, event.GoogleID, googleEvent).Do()
	if err != nil {
		s.logger.Error("Error al actualizar evento en Google Calendar", err, map[string]interface{}{
			"event_id":  eventID,
			"google_id": event.GoogleID,
		})
		return nil, fmt.Errorf("error al actualizar evento en Google Calendar: %w", err)
	}

	// Actualizar evento local
	updatedLocalEvent := s.convertFromGoogleEvent(updatedEvent, event.TenantID, event.ChannelID, event.CalendarID)
	updatedLocalEvent.ID = event.ID
	updatedLocalEvent.UpdatedAt = time.Now()

	err = s.repo.UpdateEvent(ctx, updatedLocalEvent)
	if err != nil {
		s.logger.Error("Error al actualizar evento en base de datos", err, map[string]interface{}{
			"event_id": eventID,
		})
		// No fallar si no se puede actualizar localmente
	}

	s.logger.Info("Evento actualizado exitosamente", map[string]interface{}{
		"event_id":  eventID,
		"google_id": event.GoogleID,
		"summary":   updatedLocalEvent.Summary,
	})

	return updatedLocalEvent, nil
}

// DeleteEvent elimina un evento
func (s *GoogleCalendarService) DeleteEvent(ctx context.Context, eventID string) error {
	s.logger.Info("Eliminando evento de Google Calendar", map[string]interface{}{
		"event_id": eventID,
	})

	// Obtener evento de base de datos local
	event, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("error al obtener evento: %w", err)
	}

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, event.ChannelID)
	if err != nil {
		return fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.setupSvc.createOAuth2Client(ctx, integration)
	if err != nil {
		return fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Eliminar evento de Google Calendar
	err = calendarService.Events.Delete(event.CalendarID, event.GoogleID).Do()
	if err != nil {
		s.logger.Error("Error al eliminar evento de Google Calendar", err, map[string]interface{}{
			"event_id":  eventID,
			"google_id": event.GoogleID,
		})
		return fmt.Errorf("error al eliminar evento de Google Calendar: %w", err)
	}

	// Eliminar evento de base de datos local
	err = s.repo.DeleteEvent(ctx, eventID)
	if err != nil {
		s.logger.Error("Error al eliminar evento de base de datos", err, map[string]interface{}{
			"event_id": eventID,
		})
		// No fallar si no se puede eliminar localmente
	}

	s.logger.Info("Evento eliminado exitosamente", map[string]interface{}{
		"event_id":  eventID,
		"google_id": event.GoogleID,
	})

	return nil
}

// ListEvents lista eventos de Google Calendar
func (s *GoogleCalendarService) ListEvents(ctx context.Context, req *domain.ListEventsRequest) (*EventListResponse, error) {
	s.logger.Info("Listando eventos de Google Calendar", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"channel_id":  req.ChannelID,
		"calendar_id": req.CalendarID,
	})

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, req.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.setupSvc.createOAuth2Client(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Configurar parámetros de búsqueda
	calendarID := req.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	timeMin := time.Now().Format(time.RFC3339)
	if req.StartTime != nil {
		timeMin = req.StartTime.Format(time.RFC3339)
	}

	timeMax := ""
	if req.EndTime != nil {
		timeMax = req.EndTime.Format(time.RFC3339)
	}

	maxResults := int64(10)
	if req.MaxResults > 0 {
		maxResults = int64(req.MaxResults)
	}

	// Listar eventos
	eventsCall := calendarService.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		MaxResults(maxResults).
		OrderBy("startTime")

	if timeMax != "" {
		eventsCall = eventsCall.TimeMax(timeMax)
	}

	if req.PageToken != "" {
		eventsCall = eventsCall.PageToken(req.PageToken)
	}

	events, err := eventsCall.Do()
	if err != nil {
		s.logger.Error("Error al listar eventos de Google Calendar", err, map[string]interface{}{
			"calendar_id": calendarID,
		})
		return nil, fmt.Errorf("error al listar eventos de Google Calendar: %w", err)
	}

	// Convertir eventos a dominio
	domainEvents := make([]*domain.CalendarEvent, 0, len(events.Items))
	for _, googleEvent := range events.Items {
		event := s.convertFromGoogleEvent(googleEvent, req.TenantID, req.ChannelID, calendarID)
		domainEvents = append(domainEvents, event)
	}

	s.logger.Info("Eventos listados exitosamente", map[string]interface{}{
		"total_events":    len(domainEvents),
		"next_page_token": events.NextPageToken,
	})

	return &EventListResponse{
		Events:        domainEvents,
		NextPageToken: events.NextPageToken,
		TotalEvents:   len(domainEvents),
	}, nil
}

// SyncEvents sincroniza eventos entre Google Calendar y base de datos local
func (s *GoogleCalendarService) SyncEvents(ctx context.Context, channelID string) (*SyncResult, error) {
	s.logger.Info("Iniciando sincronización de eventos", map[string]interface{}{
		"channel_id": channelID,
	})

	result := &SyncResult{}

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.setupSvc.createOAuth2Client(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Obtener eventos de Google Calendar
	googleEvents, err := calendarService.Events.List(integration.CalendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)). // Últimos 30 días
		TimeMax(time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)). // Próximo año
		Do()
	if err != nil {
		return nil, fmt.Errorf("error al obtener eventos de Google Calendar: %w", err)
	}

	// Obtener eventos locales
	localEvents, err := s.repo.GetEventsByChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener eventos locales: %w", err)
	}

	// Crear mapas para comparación
	googleEventMap := make(map[string]*calendar.Event)
	for _, event := range googleEvents.Items {
		googleEventMap[event.Id] = event
	}

	localEventMap := make(map[string]*domain.CalendarEvent)
	for _, event := range localEvents {
		localEventMap[event.GoogleID] = event
	}

	// Sincronizar eventos
	for googleID, googleEvent := range googleEventMap {
		if localEvent, exists := localEventMap[googleID]; exists {
			// Evento existe en ambos, verificar si necesita actualización
			if s.needsUpdate(localEvent, googleEvent) {
				updatedEvent := s.convertFromGoogleEvent(googleEvent, localEvent.TenantID, localEvent.ChannelID, localEvent.CalendarID)
				updatedEvent.ID = localEvent.ID
				updatedEvent.UpdatedAt = time.Now()

				err := s.repo.UpdateEvent(ctx, updatedEvent)
				if err != nil {
					result.Errors++
					result.ErrorList = append(result.ErrorList, fmt.Sprintf("Error actualizando evento %s: %v", googleID, err))
				} else {
					result.Updated++
				}
			}
		} else {
			// Evento nuevo en Google Calendar
			newEvent := s.convertFromGoogleEvent(googleEvent, integration.TenantID, channelID, integration.CalendarID)
			newEvent.ID = uuid.New().String()
			newEvent.CreatedAt = time.Now()
			newEvent.UpdatedAt = time.Now()

			err := s.repo.CreateEvent(ctx, newEvent)
			if err != nil {
				result.Errors++
				result.ErrorList = append(result.ErrorList, fmt.Sprintf("Error creando evento %s: %v", googleID, err))
			} else {
				result.Created++
			}
		}
	}

	// Verificar eventos eliminados en Google Calendar
	for googleID, localEvent := range localEventMap {
		if _, exists := googleEventMap[googleID]; !exists {
			// Evento eliminado en Google Calendar
			err := s.repo.DeleteEvent(ctx, localEvent.ID)
			if err != nil {
				result.Errors++
				result.ErrorList = append(result.ErrorList, fmt.Sprintf("Error eliminando evento %s: %v", googleID, err))
			} else {
				result.Deleted++
			}
		}
	}

	s.logger.Info("Sincronización completada", map[string]interface{}{
		"channel_id": channelID,
		"created":    result.Created,
		"updated":    result.Updated,
		"deleted":    result.Deleted,
		"errors":     result.Errors,
	})

	return result, nil
}

// setupEventNotifications configura notificaciones para un evento
func (s *GoogleCalendarService) setupEventNotifications(ctx context.Context, event *domain.CalendarEvent, reminders []domain.EventReminder) error {
	// TODO: Implementar integración con servicios de notificación
	// - Email notifications
	// - SMS notifications
	// - WhatsApp notifications
	// - Telegram notifications

	s.logger.Info("Configurando notificaciones para evento", map[string]interface{}{
		"event_id":  event.ID,
		"reminders": reminders,
	})

	// Por ahora, solo logueamos las notificaciones
	for _, reminder := range reminders {
		s.logger.Info("Recordatorio configurado", map[string]interface{}{
			"event_id": event.ID,
			"method":   reminder.Method,
			"minutes":  reminder.Minutes,
		})
	}

	return nil
}

// convertToGoogleEvent convierte un request de dominio a evento de Google Calendar
func (s *GoogleCalendarService) convertToGoogleEvent(req *domain.CreateEventRequest) *calendar.Event {
	event := &calendar.Event{
		Summary:     req.Summary,
		Description: req.Description,
		Location:    req.Location,
		Start: &calendar.EventDateTime{
			DateTime: req.StartTime.Format(time.RFC3339),
			TimeZone: s.config.DefaultTimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: req.EndTime.Format(time.RFC3339),
			TimeZone: s.config.DefaultTimeZone,
		},
	}

	// Configurar evento de todo el día
	if req.AllDay {
		event.Start = &calendar.EventDateTime{
			Date:     req.StartTime.Format("2006-01-02"),
			TimeZone: s.config.DefaultTimeZone,
		}
		event.End = &calendar.EventDateTime{
			Date:     req.EndTime.Format("2006-01-02"),
			TimeZone: s.config.DefaultTimeZone,
		}
	}

	// Configurar asistentes
	if len(req.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, 0, len(req.Attendees))
		for _, attendee := range req.Attendees {
			attendees = append(attendees, &calendar.EventAttendee{
				Email: attendee.Email,
				Name:  attendee.Name,
			})
		}
		event.Attendees = attendees
	}

	// Configurar recurrencia
	if req.Recurrence != nil {
		event.Recurrence = s.buildRecurrenceRule(req.Recurrence)
	}

	// Configurar visibilidad
	if req.Visibility != "" {
		event.Visibility = string(req.Visibility)
	}

	// Configurar recordatorios
	if len(req.Reminders) > 0 {
		reminders := make([]*calendar.EventReminder, 0, len(req.Reminders))
		for _, reminder := range req.Reminders {
			reminders = append(reminders, &calendar.EventReminder{
				Method:  reminder.Method,
				Minutes: int64(reminder.Minutes),
			})
		}
		event.Reminders = &calendar.EventReminders{
			UseDefault: false,
			Overrides:  reminders,
		}
	}

	return event
}

// convertFromGoogleEvent convierte un evento de Google Calendar a dominio
func (s *GoogleCalendarService) convertFromGoogleEvent(googleEvent *calendar.Event, tenantID, channelID, calendarID string) *domain.CalendarEvent {
	event := &domain.CalendarEvent{
		GoogleID:    googleEvent.Id,
		TenantID:    tenantID,
		ChannelID:   channelID,
		CalendarID:  calendarID,
		Summary:     googleEvent.Summary,
		Description: googleEvent.Description,
		Location:    googleEvent.Location,
		Status:      domain.EventStatus(googleEvent.Status),
		Visibility:  domain.EventVisibility(googleEvent.Visibility),
	}

	// Parsear fechas de inicio y fin
	if googleEvent.Start != nil {
		if googleEvent.Start.DateTime != "" {
			startTime, _ := time.Parse(time.RFC3339, googleEvent.Start.DateTime)
			event.StartTime = startTime
			event.AllDay = false
		} else if googleEvent.Start.Date != "" {
			startTime, _ := time.Parse("2006-01-02", googleEvent.Start.Date)
			event.StartTime = startTime
			event.AllDay = true
		}
	}

	if googleEvent.End != nil {
		if googleEvent.End.DateTime != "" {
			endTime, _ := time.Parse(time.RFC3339, googleEvent.End.DateTime)
			event.EndTime = endTime
		} else if googleEvent.End.Date != "" {
			endTime, _ := time.Parse("2006-01-02", googleEvent.End.Date)
			event.EndTime = endTime
		}
	}

	// Parsear asistentes
	if len(googleEvent.Attendees) > 0 {
		attendees := make([]domain.CalendarAttendee, 0, len(googleEvent.Attendees))
		for _, attendee := range googleEvent.Attendees {
			attendees = append(attendees, domain.CalendarAttendee{
				Email:          attendee.Email,
				Name:           attendee.DisplayName,
				ResponseStatus: attendee.ResponseStatus,
				Organizer:      attendee.Organizer,
				Self:           attendee.Self,
			})
		}
		event.Attendees = attendees
	}

	// Parsear recordatorios
	if googleEvent.Reminders != nil && len(googleEvent.Reminders.Overrides) > 0 {
		reminders := make([]domain.EventReminder, 0, len(googleEvent.Reminders.Overrides))
		for _, reminder := range googleEvent.Reminders.Overrides {
			reminders = append(reminders, domain.EventReminder{
				Method:  reminder.Method,
				Minutes: int(reminder.Minutes),
			})
		}
		event.Reminders = reminders
	}

	return event
}

// updateGoogleEvent actualiza un evento de Google Calendar con los campos del request
func (s *GoogleCalendarService) updateGoogleEvent(googleEvent *calendar.Event, req *domain.UpdateEventRequest) {
	if req.Summary != "" {
		googleEvent.Summary = req.Summary
	}
	if req.Description != "" {
		googleEvent.Description = req.Description
	}
	if req.Location != "" {
		googleEvent.Location = req.Location
	}
	if req.StartTime != nil {
		googleEvent.Start = &calendar.EventDateTime{
			DateTime: req.StartTime.Format(time.RFC3339),
			TimeZone: s.config.DefaultTimeZone,
		}
	}
	if req.EndTime != nil {
		googleEvent.End = &calendar.EventDateTime{
			DateTime: req.EndTime.Format(time.RFC3339),
			TimeZone: s.config.DefaultTimeZone,
		}
	}
	if req.AllDay != nil {
		if *req.AllDay {
			googleEvent.Start = &calendar.EventDateTime{
				Date:     googleEvent.Start.DateTime[:10], // YYYY-MM-DD
				TimeZone: s.config.DefaultTimeZone,
			}
			googleEvent.End = &calendar.EventDateTime{
				Date:     googleEvent.End.DateTime[:10], // YYYY-MM-DD
				TimeZone: s.config.DefaultTimeZone,
			}
		}
	}
	if req.Visibility != "" {
		googleEvent.Visibility = string(req.Visibility)
	}
}

// buildRecurrenceRule construye la regla de recurrencia para Google Calendar
func (s *GoogleCalendarService) buildRecurrenceRule(recurrence *domain.EventRecurrence) []string {
	var rules []string

	// Construir regla básica
	rule := fmt.Sprintf("FREQ=%s", strings.ToUpper(recurrence.Frequency))

	if recurrence.Interval > 1 {
		rule += fmt.Sprintf(";INTERVAL=%d", recurrence.Interval)
	}

	if recurrence.Count > 0 {
		rule += fmt.Sprintf(";COUNT=%d", recurrence.Count)
	}

	if recurrence.Until != nil {
		rule += fmt.Sprintf(";UNTIL=%s", recurrence.Until.Format("20060102T150405Z"))
	}

	if len(recurrence.ByDay) > 0 {
		rule += fmt.Sprintf(";BYDAY=%s", strings.Join(recurrence.ByDay, ","))
	}

	if len(recurrence.ByMonth) > 0 {
		months := make([]string, len(recurrence.ByMonth))
		for i, month := range recurrence.ByMonth {
			months[i] = fmt.Sprintf("%d", month)
		}
		rule += fmt.Sprintf(";BYMONTH=%s", strings.Join(months, ","))
	}

	if len(recurrence.ByMonthDay) > 0 {
		days := make([]string, len(recurrence.ByMonthDay))
		for i, day := range recurrence.ByMonthDay {
			days[i] = fmt.Sprintf("%d", day)
		}
		rule += fmt.Sprintf(";BYMONTHDAY=%s", strings.Join(days, ","))
	}

	rules = append(rules, rule)
	return rules
}

// needsUpdate determina si un evento local necesita actualización
func (s *GoogleCalendarService) needsUpdate(localEvent *domain.CalendarEvent, googleEvent *calendar.Event) bool {
	// Comparar campos principales
	if localEvent.Summary != googleEvent.Summary {
		return true
	}
	if localEvent.Description != googleEvent.Description {
		return true
	}
	if localEvent.Location != googleEvent.Location {
		return true
	}
	if localEvent.Status != domain.EventStatus(googleEvent.Status) {
		return true
	}

	// Comparar fechas de inicio
	if googleEvent.Start != nil && googleEvent.Start.DateTime != "" {
		googleStartTime, _ := time.Parse(time.RFC3339, googleEvent.Start.DateTime)
		if !localEvent.StartTime.Equal(googleStartTime) {
			return true
		}
	}

	// Comparar fechas de fin
	if googleEvent.End != nil && googleEvent.End.DateTime != "" {
		googleEndTime, _ := time.Parse(time.RFC3339, googleEvent.End.DateTime)
		if !localEvent.EndTime.Equal(googleEndTime) {
			return true
		}
	}

	return false
}
