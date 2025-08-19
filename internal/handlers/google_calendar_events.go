package handlers

import (
	"net/http"
	"strconv"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GoogleCalendarEventsHandler maneja las operaciones de eventos de Google Calendar
type GoogleCalendarEventsHandler struct {
	eventService *services.GoogleCalendarService
	config       *config.GoogleCalendarConfig
	logger       logger.Logger
}

// NewGoogleCalendarEventsHandler crea una nueva instancia del handler
func NewGoogleCalendarEventsHandler(eventService *services.GoogleCalendarService, config *config.GoogleCalendarConfig, logger logger.Logger) *GoogleCalendarEventsHandler {
	return &GoogleCalendarEventsHandler{
		eventService: eventService,
		config:       config,
		logger:       logger,
	}
}

// ListEventsRequest representa la solicitud de listado de eventos con filtros
type ListEventsRequest struct {
	TenantID   string     `json:"tenant_id" binding:"required"`
	ChannelID  string     `json:"channel_id" binding:"required"`
	CalendarID string     `json:"calendar_id"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Status     string     `json:"status"`
	Attendee   string     `json:"attendee"`
	MaxResults int        `json:"max_results"`
	PageToken  string     `json:"page_token"`
}

// SyncEventsRequest representa la solicitud de sincronización
type SyncEventsRequest struct {
	TenantID  string `json:"tenant_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
}

// WebhookNotification representa una notificación de webhook de Google Calendar
type WebhookNotification struct {
	State       string `json:"state"`
	ResourceID  string `json:"resourceId"`
	ResourceURI string `json:"resourceUri"`
	Expiration  string `json:"expiration"`
}

// ListEvents lista eventos de Google Calendar con filtros y paginación
// @Summary Listar eventos
// @Description Lista eventos de Google Calendar con filtros opcionales y paginación
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param request body ListEventsRequest true "Filtros de búsqueda"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events [get]
func (h *GoogleCalendarEventsHandler) ListEvents(c *gin.Context) {
	var req ListEventsRequest

	// Parsear parámetros de query string
	req.TenantID = c.Query("tenant_id")
	req.ChannelID = c.Query("channel_id")
	req.CalendarID = c.Query("calendar_id")
	req.Status = c.Query("status")
	req.Attendee = c.Query("attendee")
	req.PageToken = c.Query("page_token")

	// Parsear fechas
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	// Parsear max_results
	if maxResultsStr := c.Query("max_results"); maxResultsStr != "" {
		if maxResults, err := strconv.Atoi(maxResultsStr); err == nil {
			req.MaxResults = maxResults
		}
	}

	// Validar parámetros requeridos
	if req.TenantID == "" || req.ChannelID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_REQUIRED_PARAMS",
			Message: "tenant_id y channel_id son requeridos",
			Data:    nil,
		})
		return
	}

	// Convertir a dominio
	listReq := &domain.ListEventsRequest{
		TenantID:   req.TenantID,
		ChannelID:  req.ChannelID,
		CalendarID: req.CalendarID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		MaxResults: req.MaxResults,
		PageToken:  req.PageToken,
	}

	// Listar eventos
	response, err := h.eventService.ListEvents(c.Request.Context(), listReq)
	if err != nil {
		h.logger.Error("Error al listar eventos", err, map[string]interface{}{
			"tenant_id":   req.TenantID,
			"channel_id":  req.ChannelID,
			"calendar_id": req.CalendarID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENTS_LIST_ERROR",
			Message: "Error al listar eventos",
			Data:    err.Error(),
		})
		return
	}

	// Aplicar filtros adicionales si se especifican
	if req.Status != "" || req.Attendee != "" {
		filteredEvents := make([]*domain.CalendarEvent, 0)
		for _, event := range response.Events {
			// Filtrar por estado
			if req.Status != "" && string(event.Status) != req.Status {
				continue
			}

			// Filtrar por asistente
			if req.Attendee != "" {
				found := false
				for _, attendee := range event.Attendees {
					if attendee.Email == req.Attendee || attendee.Name == req.Attendee {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			filteredEvents = append(filteredEvents, event)
		}
		response.Events = filteredEvents
		response.TotalEvents = len(filteredEvents)
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENTS_LISTED",
		Message: "Eventos listados exitosamente",
		Data:    response,
	})
}

// CreateEvent crea un nuevo evento en Google Calendar
// @Summary Crear evento
// @Description Crea un nuevo evento en Google Calendar
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param request body domain.CreateEventRequest true "Datos del evento"
// @Success 201 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events [post]
func (h *GoogleCalendarEventsHandler) CreateEvent(c *gin.Context) {
	var req domain.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de creación de evento", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Validar campos requeridos
	if req.Summary == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_SUMMARY",
			Message: "El campo summary es requerido",
			Data:    nil,
		})
		return
	}

	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_DATES",
			Message: "Las fechas de inicio y fin son requeridas",
			Data:    nil,
		})
		return
	}

	if req.StartTime.After(req.EndTime) {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_DATES",
			Message: "La fecha de inicio debe ser anterior a la fecha de fin",
			Data:    nil,
		})
		return
	}

	// Crear evento
	event, err := h.eventService.CreateEvent(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Error al crear evento", err, map[string]interface{}{
			"tenant_id":  req.TenantID,
			"channel_id": req.ChannelID,
			"summary":    req.Summary,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENT_CREATION_ERROR",
			Message: "Error al crear evento",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "EVENT_CREATED",
		Message: "Evento creado exitosamente",
		Data:    event,
	})
}

// GetEvent obtiene un evento específico
// @Summary Obtener evento
// @Description Obtiene un evento específico por ID
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param event_id path string true "ID del evento"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/{event_id} [get]
func (h *GoogleCalendarEventsHandler) GetEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_EVENT_ID",
			Message: "ID del evento es requerido",
			Data:    nil,
		})
		return
	}

	// Obtener evento
	event, err := h.eventService.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		h.logger.Error("Error al obtener evento", err, map[string]interface{}{
			"event_id": eventID,
		})
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "EVENT_NOT_FOUND",
			Message: "Evento no encontrado",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENT_FOUND",
		Message: "Evento obtenido exitosamente",
		Data:    event,
	})
}

// UpdateEvent actualiza un evento existente
// @Summary Actualizar evento
// @Description Actualiza un evento existente en Google Calendar
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param event_id path string true "ID del evento"
// @Param request body domain.UpdateEventRequest true "Datos de actualización"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/{event_id} [put]
func (h *GoogleCalendarEventsHandler) UpdateEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_EVENT_ID",
			Message: "ID del evento es requerido",
			Data:    nil,
		})
		return
	}

	var req domain.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de actualización", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Validar fechas si se proporcionan
	if req.StartTime != nil && req.EndTime != nil {
		if req.StartTime.After(*req.EndTime) {
			c.JSON(http.StatusBadRequest, domain.APIResponse{
				Code:    "INVALID_DATES",
				Message: "La fecha de inicio debe ser anterior a la fecha de fin",
				Data:    nil,
			})
			return
		}
	}

	// Actualizar evento
	event, err := h.eventService.UpdateEvent(c.Request.Context(), eventID, &req)
	if err != nil {
		h.logger.Error("Error al actualizar evento", err, map[string]interface{}{
			"event_id": eventID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENT_UPDATE_ERROR",
			Message: "Error al actualizar evento",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENT_UPDATED",
		Message: "Evento actualizado exitosamente",
		Data:    event,
	})
}

// DeleteEvent elimina un evento
// @Summary Eliminar evento
// @Description Elimina un evento de Google Calendar
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param event_id path string true "ID del evento"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/{event_id} [delete]
func (h *GoogleCalendarEventsHandler) DeleteEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_EVENT_ID",
			Message: "ID del evento es requerido",
			Data:    nil,
		})
		return
	}

	// Eliminar evento
	err := h.eventService.DeleteEvent(c.Request.Context(), eventID)
	if err != nil {
		h.logger.Error("Error al eliminar evento", err, map[string]interface{}{
			"event_id": eventID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENT_DELETION_ERROR",
			Message: "Error al eliminar evento",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENT_DELETED",
		Message: "Evento eliminado exitosamente",
		Data: map[string]interface{}{
			"event_id": eventID,
		},
	})
}

// SyncEvents sincroniza eventos entre Google Calendar y base de datos local
// @Summary Sincronizar eventos
// @Description Sincroniza eventos entre Google Calendar y la base de datos local
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param request body SyncEventsRequest true "Datos de sincronización"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/sync [post]
func (h *GoogleCalendarEventsHandler) SyncEvents(c *gin.Context) {
	var req SyncEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de sincronización", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Sincronizar eventos
	result, err := h.eventService.SyncEvents(c.Request.Context(), req.ChannelID)
	if err != nil {
		h.logger.Error("Error al sincronizar eventos", err, map[string]interface{}{
			"tenant_id":  req.TenantID,
			"channel_id": req.ChannelID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SYNC_ERROR",
			Message: "Error al sincronizar eventos",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SYNC_COMPLETED",
		Message: "Sincronización completada exitosamente",
		Data:    result,
	})
}

// HandleWebhook maneja las notificaciones de webhook de Google Calendar
// @Summary Manejar webhook
// @Description Maneja las notificaciones de webhook de Google Calendar
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param notification body WebhookNotification true "Notificación de webhook"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /webhooks/google-calendar [post]
func (h *GoogleCalendarEventsHandler) HandleWebhook(c *gin.Context) {
	var notification WebhookNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		h.logger.Error("Error al validar notificación de webhook", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_WEBHOOK",
			Message: "Notificación de webhook inválida",
			Data:    err.Error(),
		})
		return
	}

	h.logger.Info("Webhook recibido de Google Calendar", map[string]interface{}{
		"state":        notification.State,
		"resource_id":  notification.ResourceID,
		"resource_uri": notification.ResourceURI,
		"expiration":   notification.Expiration,
	})

	// TODO: Implementar procesamiento de webhook
	// - Extraer channel_id del resource_uri
	// - Sincronizar eventos del canal específico
	// - Enviar notificaciones si es necesario

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "WEBHOOK_PROCESSED",
		Message: "Webhook procesado exitosamente",
		Data: map[string]interface{}{
			"state":       notification.State,
			"resource_id": notification.ResourceID,
		},
	})
}

// GetEventsByDateRange obtiene eventos en un rango de fechas específico
// @Summary Obtener eventos por rango de fechas
// @Description Obtiene eventos en un rango de fechas específico
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param channel_id path string true "ID del canal"
// @Param start_time query string true "Fecha de inicio (RFC3339)"
// @Param end_time query string true "Fecha de fin (RFC3339)"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/range/{channel_id} [get]
func (h *GoogleCalendarEventsHandler) GetEventsByDateRange(c *gin.Context) {
	channelID := c.Param("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_CHANNEL_ID",
			Message: "ID del canal es requerido",
			Data:    nil,
		})
		return
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_DATES",
			Message: "start_time y end_time son requeridos",
			Data:    nil,
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_START_TIME",
			Message: "Formato de fecha de inicio inválido (RFC3339)",
			Data:    err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_END_TIME",
			Message: "Formato de fecha de fin inválido (RFC3339)",
			Data:    err.Error(),
		})
		return
	}

	if startTime.After(endTime) {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_DATE_RANGE",
			Message: "La fecha de inicio debe ser anterior a la fecha de fin",
			Data:    nil,
		})
		return
	}

	// Obtener eventos por rango de fechas
	events, err := h.eventService.GetEventsByDateRange(c.Request.Context(), channelID, startTime, endTime)
	if err != nil {
		h.logger.Error("Error al obtener eventos por rango de fechas", err, map[string]interface{}{
			"channel_id": channelID,
			"start_time": startTime,
			"end_time":   endTime,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENTS_RANGE_ERROR",
			Message: "Error al obtener eventos por rango de fechas",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENTS_RANGE_FOUND",
		Message: "Eventos obtenidos exitosamente",
		Data: map[string]interface{}{
			"channel_id":   channelID,
			"start_time":   startTime,
			"end_time":     endTime,
			"events":       events,
			"total_events": len(events),
		},
	})
}

// GetEventsByTenant obtiene eventos de un tenant con paginación
// @Summary Obtener eventos por tenant
// @Description Obtiene eventos de un tenant con paginación
// @Tags Google Calendar Events
// @Accept json
// @Produce json
// @Param tenant_id path string true "ID del tenant"
// @Param limit query int false "Límite de resultados (default: 10)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/events/tenant/{tenant_id} [get]
func (h *GoogleCalendarEventsHandler) GetEventsByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_TENANT_ID",
			Message: "ID del tenant es requerido",
			Data:    nil,
		})
		return
	}

	// Parsear parámetros de paginación
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Obtener eventos del tenant
	events, err := h.eventService.GetEventsByTenant(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Error al obtener eventos del tenant", err, map[string]interface{}{
			"tenant_id": tenantID,
			"limit":     limit,
			"offset":    offset,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "EVENTS_TENANT_ERROR",
			Message: "Error al obtener eventos del tenant",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "EVENTS_TENANT_FOUND",
		Message: "Eventos del tenant obtenidos exitosamente",
		Data: map[string]interface{}{
			"tenant_id":    tenantID,
			"events":       events,
			"total_events": len(events),
			"limit":        limit,
			"offset":       offset,
		},
	})
}
