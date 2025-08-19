package handlers

import (
	"context"
	"net/http"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GoogleCalendarWebhookHandler maneja los webhooks de Google Calendar
type GoogleCalendarWebhookHandler struct {
	notificationService *services.NotificationService
	eventService        *services.GoogleCalendarService
	config              *config.GoogleCalendarConfig
	logger              logger.Logger
}

// NewGoogleCalendarWebhookHandler crea una nueva instancia del handler
func NewGoogleCalendarWebhookHandler(
	notificationService *services.NotificationService,
	eventService *services.GoogleCalendarService,
	config *config.GoogleCalendarConfig,
	logger logger.Logger,
) *GoogleCalendarWebhookHandler {
	return &GoogleCalendarWebhookHandler{
		notificationService: notificationService,
		eventService:        eventService,
		config:              config,
		logger:              logger,
	}
}

// WebhookPayload representa el payload de webhook de Google Calendar
type WebhookPayload struct {
	State       string `json:"state"`
	ResourceID  string `json:"resourceId"`
	ResourceURI string `json:"resourceUri"`
	Expiration  string `json:"expiration"`
	Token       string `json:"token"`
}

// WebhookSyncRequest representa una solicitud de sincronización desde webhook
type WebhookSyncRequest struct {
	ChannelID   string `json:"channel_id"`
	CalendarID  string `json:"calendar_id"`
	SyncToken   string `json:"sync_token,omitempty"`
	EventID     string `json:"event_id,omitempty"`
	Action      string `json:"action"` // created, updated, deleted
}

// HandleWebhook maneja las notificaciones de webhook de Google Calendar
// @Summary Manejar webhook de Google Calendar
// @Description Procesa notificaciones de webhook de Google Calendar y envía notificaciones automáticas
// @Tags Google Calendar Webhooks
// @Accept json
// @Produce json
// @Param payload body WebhookPayload true "Payload del webhook"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /webhooks/google-calendar [post]
func (h *GoogleCalendarWebhookHandler) HandleWebhook(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("Error al validar payload de webhook", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_WEBHOOK_PAYLOAD",
			Message: "Payload de webhook inválido",
			Data:    err.Error(),
		})
		return
	}

	h.logger.Info("Webhook recibido de Google Calendar", map[string]interface{}{
		"state":        payload.State,
		"resource_id":  payload.ResourceID,
		"resource_uri": payload.ResourceURI,
		"expiration":   payload.Expiration,
	})

	// Validar token de webhook si es necesario
	if !h.validateWebhookToken(payload.Token) {
		h.logger.Warn("Token de webhook inválido", map[string]interface{}{
			"token": payload.Token,
		})
		c.JSON(http.StatusUnauthorized, domain.APIResponse{
			Code:    "INVALID_WEBHOOK_TOKEN",
			Message: "Token de webhook inválido",
			Data:    nil,
		})
		return
	}

	// Procesar webhook en background
	go h.processWebhookAsync(c.Request.Context(), &payload)

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "WEBHOOK_PROCESSED",
		Message: "Webhook procesado exitosamente",
		Data: map[string]interface{}{
			"state":       payload.State,
			"resource_id": payload.ResourceID,
			"processed_at": time.Now(),
		},
	})
}

// HandleSyncRequest maneja solicitudes de sincronización desde webhooks
// @Summary Sincronizar eventos desde webhook
// @Description Sincroniza eventos específicos cuando se recibe una notificación de webhook
// @Tags Google Calendar Webhooks
// @Accept json
// @Produce json
// @Param request body WebhookSyncRequest true "Datos de sincronización"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /webhooks/google-calendar/sync [post]
func (h *GoogleCalendarWebhookHandler) HandleSyncRequest(c *gin.Context) {
	var req WebhookSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de sincronización", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_SYNC_REQUEST",
			Message: "Request de sincronización inválido",
			Data:    err.Error(),
		})
		return
	}

	h.logger.Info("Procesando sincronización desde webhook", map[string]interface{}{
		"channel_id": req.ChannelID,
		"action":     req.Action,
		"event_id":   req.EventID,
	})

	// Procesar sincronización según la acción
	switch req.Action {
	case "created":
		err := h.handleEventCreated(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Error procesando evento creado", err, map[string]interface{}{
				"event_id": req.EventID,
			})
			c.JSON(http.StatusInternalServerError, domain.APIResponse{
				Code:    "EVENT_CREATION_ERROR",
				Message: "Error procesando evento creado",
				Data:    err.Error(),
			})
			return
		}

	case "updated":
		err := h.handleEventUpdated(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Error procesando evento actualizado", err, map[string]interface{}{
				"event_id": req.EventID,
			})
			c.JSON(http.StatusInternalServerError, domain.APIResponse{
				Code:    "EVENT_UPDATE_ERROR",
				Message: "Error procesando evento actualizado",
				Data:    err.Error(),
			})
			return
		}

	case "deleted":
		err := h.handleEventDeleted(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Error procesando evento eliminado", err, map[string]interface{}{
				"event_id": req.EventID,
			})
			c.JSON(http.StatusInternalServerError, domain.APIResponse{
				Code:    "EVENT_DELETION_ERROR",
				Message: "Error procesando evento eliminado",
				Data:    err.Error(),
			})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_ACTION",
			Message: "Acción no válida",
			Data:    req.Action,
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SYNC_COMPLETED",
		Message: "Sincronización completada exitosamente",
		Data: map[string]interface{}{
			"channel_id": req.ChannelID,
			"action":     req.Action,
			"event_id":   req.EventID,
			"synced_at":  time.Now(),
		},
	})
}

// HandleNotificationRequest maneja solicitudes de notificación manual
// @Summary Enviar notificación manual
// @Description Envía notificaciones manuales para eventos específicos
// @Tags Google Calendar Webhooks
// @Accept json
// @Produce json
// @Param request body services.NotificationRequest true "Datos de notificación"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /webhooks/google-calendar/notify [post]
func (h *GoogleCalendarWebhookHandler) HandleNotificationRequest(c *gin.Context) {
	var req services.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de notificación", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_NOTIFICATION_REQUEST",
			Message: "Request de notificación inválido",
			Data:    err.Error(),
		})
		return
	}

	h.logger.Info("Enviando notificación manual", map[string]interface{}{
		"event_id":          req.EventID,
		"notification_type": req.NotificationType,
		"attendees_count":   len(req.Attendees),
	})

	var results []*services.NotificationResult
	var err error

	// Enviar notificación según el tipo
	switch req.NotificationType {
	case services.NotificationTypeReminder:
		results, err = h.notificationService.SendEventReminder(c.Request.Context(), &req)
	case services.NotificationTypeConfirmation:
		results, err = h.notificationService.SendEventConfirmation(c.Request.Context(), &req)
	case services.NotificationTypeUpdate:
		results, err = h.notificationService.SendEventUpdate(c.Request.Context(), &req)
	case services.NotificationTypeCancellation:
		results, err = h.notificationService.SendEventCancellation(c.Request.Context(), &req)
	default:
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_NOTIFICATION_TYPE",
			Message: "Tipo de notificación no válido",
			Data:    req.NotificationType,
		})
		return
	}

	if err != nil {
		h.logger.Error("Error enviando notificación", err, map[string]interface{}{
			"event_id": req.EventID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "NOTIFICATION_ERROR",
			Message: "Error enviando notificación",
			Data:    err.Error(),
		})
		return
	}

	// Contar resultados exitosos
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "NOTIFICATION_SENT",
		Message: "Notificación enviada exitosamente",
		Data: map[string]interface{}{
			"event_id":       req.EventID,
			"total_sent":     len(results),
			"success_count":  successCount,
			"failure_count":  len(results) - successCount,
			"results":        results,
		},
	})
}

// Helper methods

// validateWebhookToken valida el token del webhook
func (h *GoogleCalendarWebhookHandler) validateWebhookToken(token string) bool {
	// TODO: Implementar validación real del token
	// Por ahora, aceptar cualquier token no vacío
	return token != ""
}

// processWebhookAsync procesa el webhook de forma asíncrona
func (h *GoogleCalendarWebhookHandler) processWebhookAsync(ctx context.Context, payload *WebhookPayload) {
	h.logger.Info("Procesando webhook de forma asíncrona", map[string]interface{}{
		"resource_id": payload.ResourceID,
		"state":       payload.State,
	})

	// Extraer información del resource_uri
	channelID, calendarID, err := h.extractInfoFromResourceURI(payload.ResourceURI)
	if err != nil {
		h.logger.Error("Error extrayendo información del resource URI", err, map[string]interface{}{
			"resource_uri": payload.ResourceURI,
		})
		return
	}

	// Procesar según el estado
	switch payload.State {
	case "sync":
		err = h.handleSyncState(ctx, channelID, calendarID, payload.ResourceID)
	case "exists":
		err = h.handleExistsState(ctx, channelID, calendarID, payload.ResourceID)
	default:
		h.logger.Warn("Estado de webhook no reconocido", map[string]interface{}{
			"state": payload.State,
		})
		return
	}

	if err != nil {
		h.logger.Error("Error procesando webhook", err, map[string]interface{}{
			"resource_id": payload.ResourceID,
			"state":       payload.State,
		})
	}
}

// extractInfoFromResourceURI extrae información del resource URI
func (h *GoogleCalendarWebhookHandler) extractInfoFromResourceURI(resourceURI string) (string, string, error) {
	// TODO: Implementar parsing del resource URI de Google Calendar
	// Por ahora, retornar valores por defecto
	return "default-channel", "primary", nil
}

// handleSyncState maneja el estado "sync" del webhook
func (h *GoogleCalendarWebhookHandler) handleSyncState(ctx context.Context, channelID, calendarID, resourceID string) error {
	h.logger.Info("Procesando estado sync", map[string]interface{}{
		"channel_id":  channelID,
		"calendar_id": calendarID,
		"resource_id": resourceID,
	})

	// Sincronizar eventos del canal
	_, err := h.eventService.SyncEvents(ctx, channelID)
	if err != nil {
		return err
	}

	// Procesar notificación de webhook
	notification := &domain.WebhookNotification{
		State:       "sync",
		ResourceID:  resourceID,
		ResourceURI: "",
		Expiration:  "",
	}

	return h.notificationService.ProcessWebhookNotification(ctx, notification)
}

// handleExistsState maneja el estado "exists" del webhook
func (h *GoogleCalendarWebhookHandler) handleExistsState(ctx context.Context, channelID, calendarID, resourceID string) error {
	h.logger.Info("Procesando estado exists", map[string]interface{}{
		"channel_id":  channelID,
		"calendar_id": calendarID,
		"resource_id": resourceID,
	})

	// Verificar que el recurso existe
	// TODO: Implementar verificación específica

	return nil
}

// handleEventCreated maneja eventos creados
func (h *GoogleCalendarWebhookHandler) handleEventCreated(ctx context.Context, req *WebhookSyncRequest) error {
	h.logger.Info("Manejando evento creado", map[string]interface{}{
		"event_id": req.EventID,
	})

	// Obtener evento actualizado
	event, err := h.eventService.GetEvent(ctx, req.EventID)
	if err != nil {
		return err
	}

	// Enviar confirmaciones a los asistentes
	notificationReq := &services.NotificationRequest{
		EventID:          event.ID,
		TenantID:         event.TenantID,
		ChannelID:        event.ChannelID,
		EventSummary:     event.Summary,
		EventDescription: event.Description,
		EventLocation:    event.Location,
		StartTime:        event.StartTime,
		EndTime:          event.EndTime,
		Attendees:        event.Attendees,
		NotificationType: services.NotificationTypeConfirmation,
	}

	_, err = h.notificationService.SendEventConfirmation(ctx, notificationReq)
	if err != nil {
		h.logger.Error("Error enviando confirmaciones", err, map[string]interface{}{
			"event_id": req.EventID,
		})
	}

	// Programar recordatorios automáticos
	if len(event.Reminders) > 0 {
		var reminderMinutes []int
		for _, reminder := range event.Reminders {
			reminderMinutes = append(reminderMinutes, reminder.Minutes)
		}

		err = h.notificationService.ScheduleReminders(ctx, event, reminderMinutes)
		if err != nil {
			h.logger.Error("Error programando recordatorios", err, map[string]interface{}{
				"event_id": req.EventID,
			})
		}
	}

	return nil
}

// handleEventUpdated maneja eventos actualizados
func (h *GoogleCalendarWebhookHandler) handleEventUpdated(ctx context.Context, req *WebhookSyncRequest) error {
	h.logger.Info("Manejando evento actualizado", map[string]interface{}{
		"event_id": req.EventID,
	})

	// Obtener evento actualizado
	event, err := h.eventService.GetEvent(ctx, req.EventID)
	if err != nil {
		return err
	}

	// Enviar notificaciones de actualización
	notificationReq := &services.NotificationRequest{
		EventID:          event.ID,
		TenantID:         event.TenantID,
		ChannelID:        event.ChannelID,
		EventSummary:     event.Summary,
		EventDescription: event.Description,
		EventLocation:    event.Location,
		StartTime:        event.StartTime,
		EndTime:          event.EndTime,
		Attendees:        event.Attendees,
		NotificationType: services.NotificationTypeUpdate,
	}

	_, err = h.notificationService.SendEventUpdate(ctx, notificationReq)
	if err != nil {
		h.logger.Error("Error enviando notificaciones de actualización", err, map[string]interface{}{
			"event_id": req.EventID,
		})
	}

	return nil
}

// handleEventDeleted maneja eventos eliminados
func (h *GoogleCalendarWebhookHandler) handleEventDeleted(ctx context.Context, req *WebhookSyncRequest) error {
	h.logger.Info("Manejando evento eliminado", map[string]interface{}{
		"event_id": req.EventID,
	})

	// TODO: Obtener información del evento antes de eliminarlo para las notificaciones
	// Por ahora, enviar notificación genérica

	notificationReq := &services.NotificationRequest{
		EventID:          req.EventID,
		TenantID:         "", // TODO: Obtener del evento
		ChannelID:        req.ChannelID,
		EventSummary:     "Evento cancelado",
		EventDescription: "Este evento ha sido cancelado",
		EventLocation:    "",
		StartTime:        time.Now(),
		EndTime:          time.Now(),
		Attendees:        []domain.CalendarAttendee{}, // TODO: Obtener del evento
		NotificationType: services.NotificationTypeCancellation,
	}

	_, err := h.notificationService.SendEventCancellation(context.Background(), notificationReq)
	if err != nil {
		h.logger.Error("Error enviando notificaciones de cancelación", err, map[string]interface{}{
			"event_id": req.EventID,
		})
	}

	return nil
}
