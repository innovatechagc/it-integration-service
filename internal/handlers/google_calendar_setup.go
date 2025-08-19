package handlers

import (
	"net/http"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GoogleCalendarSetupHandler maneja las operaciones de configuración de Google Calendar
type GoogleCalendarSetupHandler struct {
	setupService *services.GoogleCalendarSetupService
	config       *config.GoogleCalendarConfig
	logger       logger.Logger
}

// NewGoogleCalendarSetupHandler crea una nueva instancia del handler
func NewGoogleCalendarSetupHandler(setupService *services.GoogleCalendarSetupService, config *config.GoogleCalendarConfig, logger logger.Logger) *GoogleCalendarSetupHandler {
	return &GoogleCalendarSetupHandler{
		setupService: setupService,
		config:       config,
		logger:       logger,
	}
}

// InitiateAuthRequest representa la solicitud de inicio de autenticación
type InitiateAuthRequest struct {
	TenantID     string              `json:"tenant_id" binding:"required"`
	CalendarType domain.CalendarType `json:"calendar_type" binding:"required"`
}

// SetupWebhookRequest representa la solicitud de configuración de webhook
type SetupWebhookRequest struct {
	TenantID   string `json:"tenant_id" binding:"required"`
	ChannelID  string `json:"channel_id" binding:"required"`
	CalendarID string `json:"calendar_id" binding:"required"`
}

// RevokeAccessRequest representa la solicitud de revocación de acceso
type RevokeAccessRequest struct {
	TenantID  string `json:"tenant_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
}

// InitiateAuth inicia el flujo de autenticación OAuth2
// @Summary Iniciar autenticación OAuth2 para Google Calendar
// @Description Inicia el flujo de autenticación OAuth2 para conectar con Google Calendar
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param request body InitiateAuthRequest true "Datos de autenticación"
// @Success 200 {object} services.AuthURLResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/auth [post]
func (h *GoogleCalendarSetupHandler) InitiateAuth(c *gin.Context) {
	var req InitiateAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de autenticación", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Validar tipo de calendario
	if req.CalendarType != domain.CalendarTypePersonal &&
		req.CalendarType != domain.CalendarTypeWork &&
		req.CalendarType != domain.CalendarTypeShared {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_CALENDAR_TYPE",
			Message: "Tipo de calendario inválido. Debe ser 'personal', 'work' o 'shared'",
			Data:    nil,
		})
		return
	}

	// Iniciar autenticación
	response, err := h.setupService.InitiateAuth(c.Request.Context(), req.TenantID, req.CalendarType)
	if err != nil {
		h.logger.Error("Error al iniciar autenticación OAuth2", err, map[string]interface{}{
			"tenant_id":     req.TenantID,
			"calendar_type": req.CalendarType,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "AUTH_INITIATION_ERROR",
			Message: "Error al iniciar autenticación",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "AUTH_INITIATED",
		Message: "Autenticación iniciada exitosamente",
		Data:    response,
	})
}

// HandleCallback maneja el callback de OAuth2
// @Summary Callback de autenticación OAuth2
// @Description Maneja el callback de Google OAuth2 y completa la autenticación
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param code query string true "Código de autorización"
// @Param state query string true "Token de estado"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/callback [get]
func (h *GoogleCalendarSetupHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_PARAMETERS",
			Message: "Faltan parámetros requeridos: code y state",
			Data:    nil,
		})
		return
	}

	// Procesar callback
	err := h.setupService.HandleCallback(c.Request.Context(), code, state)
	if err != nil {
		h.logger.Error("Error al procesar callback OAuth2", err, map[string]interface{}{
			"state": state,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "CALLBACK_ERROR",
			Message: "Error al procesar callback de autenticación",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "AUTH_SUCCESS",
		Message: "Autenticación completada exitosamente",
		Data: map[string]interface{}{
			"channel_id": state,
			"status":     "authenticated",
		},
	})
}

// GetIntegrationStatus obtiene el estado de una integración
// @Summary Obtener estado de integración
// @Description Obtiene el estado actual de una integración de Google Calendar
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param channel_id path string true "ID del canal"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /integrations/google-calendar/status/{channel_id} [get]
func (h *GoogleCalendarSetupHandler) GetIntegrationStatus(c *gin.Context) {
	channelID := c.Param("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_CHANNEL_ID",
			Message: "ID del canal es requerido",
			Data:    nil,
		})
		return
	}

	// Obtener estado de integración
	status, err := h.setupService.GetIntegrationStatus(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Error al obtener estado de integración", err, map[string]interface{}{
			"channel_id": channelID,
		})
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "INTEGRATION_NOT_FOUND",
			Message: "Integración no encontrada",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "INTEGRATION_STATUS",
		Message: "Estado de integración obtenido exitosamente",
		Data:    status,
	})
}

// SetupWebhook configura webhooks para sincronización automática
// @Summary Configurar webhook
// @Description Configura webhooks para recibir notificaciones de cambios en Google Calendar
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param request body SetupWebhookRequest true "Datos de configuración de webhook"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/webhook/setup [post]
func (h *GoogleCalendarSetupHandler) SetupWebhook(c *gin.Context) {
	var req SetupWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de webhook", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Configurar webhook
	err := h.setupService.SetupWebhook(c.Request.Context(), req.ChannelID)
	if err != nil {
		h.logger.Error("Error al configurar webhook", err, map[string]interface{}{
			"channel_id":  req.ChannelID,
			"calendar_id": req.CalendarID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_SETUP_ERROR",
			Message: "Error al configurar webhook",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "WEBHOOK_SETUP_SUCCESS",
		Message: "Webhook configurado exitosamente",
		Data: map[string]interface{}{
			"channel_id":  req.ChannelID,
			"calendar_id": req.CalendarID,
			"webhook_url": h.config.WebhookURL,
		},
	})
}

// ValidateToken valida si el token actual es válido
// @Summary Validar token
// @Description Valida si el token de acceso actual es válido
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param channel_id path string true "ID del canal"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /integrations/google-calendar/validate/{channel_id} [get]
func (h *GoogleCalendarSetupHandler) ValidateToken(c *gin.Context) {
	channelID := c.Param("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_CHANNEL_ID",
			Message: "ID del canal es requerido",
			Data:    nil,
		})
		return
	}

	// Validar token
	isValid, err := h.setupService.ValidateToken(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Error al validar token", err, map[string]interface{}{
			"channel_id": channelID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "TOKEN_VALIDATION_ERROR",
			Message: "Error al validar token",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "TOKEN_VALIDATION_SUCCESS",
		Message: "Token validado exitosamente",
		Data: map[string]interface{}{
			"channel_id": channelID,
			"is_valid":   isValid,
		},
	})
}

// RevokeAccess revoca el acceso a Google Calendar
// @Summary Revocar acceso
// @Description Revoca el acceso a Google Calendar y elimina los tokens
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param request body RevokeAccessRequest true "Datos de revocación"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/revoke [post]
func (h *GoogleCalendarSetupHandler) RevokeAccess(c *gin.Context) {
	var req RevokeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Error al validar request de revocación", err, nil)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos de solicitud inválidos",
			Data:    err.Error(),
		})
		return
	}

	// Revocar acceso
	err := h.setupService.RevokeAccess(c.Request.Context(), req.ChannelID)
	if err != nil {
		h.logger.Error("Error al revocar acceso", err, map[string]interface{}{
			"channel_id": req.ChannelID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "REVOKE_ACCESS_ERROR",
			Message: "Error al revocar acceso",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "ACCESS_REVOKED",
		Message: "Acceso revocado exitosamente",
		Data: map[string]interface{}{
			"channel_id": req.ChannelID,
			"status":     "revoked",
		},
	})
}

// RefreshToken refresca manualmente el token de acceso
// @Summary Refrescar token
// @Description Refresca manualmente el token de acceso
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param channel_id path string true "ID del canal"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/refresh/{channel_id} [post]
func (h *GoogleCalendarSetupHandler) RefreshToken(c *gin.Context) {
	channelID := c.Param("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_CHANNEL_ID",
			Message: "ID del canal es requerido",
			Data:    nil,
		})
		return
	}

	// Refrescar token
	err := h.setupService.RefreshToken(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Error al refrescar token", err, map[string]interface{}{
			"channel_id": channelID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "TOKEN_REFRESH_ERROR",
			Message: "Error al refrescar token",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "TOKEN_REFRESHED",
		Message: "Token refrescado exitosamente",
		Data: map[string]interface{}{
			"channel_id": channelID,
			"status":     "refreshed",
		},
	})
}

// GetIntegrationsByTenant obtiene todas las integraciones de un tenant
// @Summary Obtener integraciones por tenant
// @Description Obtiene todas las integraciones de Google Calendar de un tenant
// @Tags Google Calendar Setup
// @Accept json
// @Produce json
// @Param tenant_id path string true "ID del tenant"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /integrations/google-calendar/tenant/{tenant_id} [get]
func (h *GoogleCalendarSetupHandler) GetIntegrationsByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "MISSING_TENANT_ID",
			Message: "ID del tenant es requerido",
			Data:    nil,
		})
		return
	}

	// Obtener integraciones del tenant
	integrations, err := h.setupService.GetIntegrationsByTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Error al obtener integraciones del tenant", err, map[string]interface{}{
			"tenant_id": tenantID,
		})
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTEGRATIONS_FETCH_ERROR",
			Message: "Error al obtener integraciones",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "INTEGRATIONS_FETCHED",
		Message: "Integraciones obtenidas exitosamente",
		Data: map[string]interface{}{
			"tenant_id":    tenantID,
			"integrations": integrations,
			"total_count":  len(integrations),
		},
	})
}
