package handlers

import (
	"net/http"
	"strconv"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type WebchatSetupHandler struct {
	webchatService     *services.WebchatSetupService
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewWebchatSetupHandler(webchatService *services.WebchatSetupService, integrationService services.IntegrationService, logger logger.Logger) *WebchatSetupHandler {
	return &WebchatSetupHandler{
		webchatService:     webchatService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// WebchatSetupRequest representa la solicitud para configurar Webchat
type WebchatSetupRequest struct {
	Config     services.WebchatConfig `json:"config" binding:"required"`
	WebhookURL string                 `json:"webhook_url" binding:"required"`
	TenantID   string                 `json:"tenant_id" binding:"required"`
}

// WebchatConfigRequest representa la solicitud para actualizar configuración
type WebchatConfigRequest struct {
	Config services.WebchatConfig `json:"config" binding:"required"`
}

// WebchatSessionRequest representa la solicitud para crear una sesión
type WebchatSessionRequest struct {
	UserID   string                 `json:"user_id" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WebchatMessageRequest representa la solicitud para enviar un mensaje
type WebchatMessageRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Text      string `json:"text" binding:"required"`
}

// SetupWebchatIntegration godoc
// @Summary Configurar integración completa de Webchat
// @Description Configura el chat web y crea la integración en una sola operación
// @Tags webchat
// @Accept json
// @Produce json
// @Param request body WebchatSetupRequest true "Datos de configuración"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/webchat/setup [post]
func (h *WebchatSetupHandler) SetupWebchatIntegration(c *gin.Context) {
	var request WebchatSetupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Crear integración usando el servicio de Webchat
	integration, err := h.webchatService.CreateWebchatIntegration(
		c.Request.Context(),
		&request.Config,
		request.WebhookURL,
		request.TenantID,
	)
	if err != nil {
		h.logger.Error("Failed to create Webchat integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SETUP_ERROR",
			Message: "Failed to setup Webchat integration: " + err.Error(),
		})
		return
	}

	// Guardar la integración en la base de datos
	if err := h.integrationService.CreateChannel(c.Request.Context(), integration); err != nil {
		h.logger.Error("Failed to save integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to save integration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat integration configured successfully",
		Data:    integration,
	})
}

// GetWebchatConfig godoc
// @Summary Obtener configuración del chat web
// @Description Obtiene la configuración actual del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param webchat_id query string true "ID del chat web"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/config [get]
func (h *WebchatSetupHandler) GetWebchatConfig(c *gin.Context) {
	webchatID := c.Query("webchat_id")
	if webchatID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "webchat_id is required",
		})
		return
	}

	config, err := h.webchatService.GetWebchatConfig(c.Request.Context(), webchatID)
	if err != nil {
		h.logger.Error("Failed to get webchat config", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get webchat config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat config retrieved successfully",
		Data:    config,
	})
}

// UpdateWebchatConfig godoc
// @Summary Actualizar configuración del chat web
// @Description Actualiza la configuración del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param request body WebchatConfigRequest true "Nueva configuración"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/config [put]
func (h *WebchatSetupHandler) UpdateWebchatConfig(c *gin.Context) {
	var request WebchatConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.webchatService.UpdateWebchatConfig(c.Request.Context(), &request.Config); err != nil {
		h.logger.Error("Failed to update webchat config", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "UPDATE_ERROR",
			Message: "Failed to update webchat config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat config updated successfully",
		Data:    request.Config,
	})
}

// CreateWebchatSession godoc
// @Summary Crear sesión de chat web
// @Description Crea una nueva sesión de chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param webchat_id query string true "ID del chat web"
// @Param request body WebchatSessionRequest true "Datos de la sesión"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/webchat/sessions [post]
func (h *WebchatSetupHandler) CreateWebchatSession(c *gin.Context) {
	webchatID := c.Query("webchat_id")
	if webchatID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "webchat_id is required",
		})
		return
	}

	var request WebchatSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	session, err := h.webchatService.CreateWebchatSession(
		c.Request.Context(),
		webchatID,
		request.UserID,
		request.Metadata,
	)
	if err != nil {
		h.logger.Error("Failed to create webchat session", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SESSION_ERROR",
			Message: "Failed to create webchat session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat session created successfully",
		Data:    session,
	})
}

// GetWebchatSessions godoc
// @Summary Obtener sesiones del chat web
// @Description Obtiene las sesiones activas del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param webchat_id query string true "ID del chat web"
// @Param limit query int false "Límite de resultados" default(10)
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/sessions [get]
func (h *WebchatSetupHandler) GetWebchatSessions(c *gin.Context) {
	webchatID := c.Query("webchat_id")
	if webchatID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "webchat_id is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	sessions, err := h.webchatService.GetWebchatSessions(c.Request.Context(), webchatID, limit)
	if err != nil {
		h.logger.Error("Failed to get webchat sessions", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get webchat sessions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat sessions retrieved successfully",
		Data:    sessions,
	})
}

// SendWebchatMessage godoc
// @Summary Enviar mensaje por chat web
// @Description Envía un mensaje a través del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param request body WebchatMessageRequest true "Datos del mensaje"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/messages [post]
func (h *WebchatSetupHandler) SendWebchatMessage(c *gin.Context) {
	var request WebchatMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	message, err := h.webchatService.SendWebchatMessage(
		c.Request.Context(),
		request.SessionID,
		request.UserID,
		request.Text,
	)
	if err != nil {
		h.logger.Error("Failed to send webchat message", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "MESSAGE_ERROR",
			Message: "Failed to send webchat message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat message sent successfully",
		Data:    message,
	})
}

// GetWebchatStats godoc
// @Summary Obtener estadísticas del chat web
// @Description Obtiene estadísticas del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param webchat_id query string true "ID del chat web"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/stats [get]
func (h *WebchatSetupHandler) GetWebchatStats(c *gin.Context) {
	webchatID := c.Query("webchat_id")
	if webchatID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "webchat_id is required",
		})
		return
	}

	stats, err := h.webchatService.GetWebchatStats(c.Request.Context(), webchatID)
	if err != nil {
		h.logger.Error("Failed to get webchat stats", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get webchat stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat stats retrieved successfully",
		Data:    stats,
	})
}

// ValidateWebchatConfig godoc
// @Summary Validar configuración del chat web
// @Description Valida la configuración del chat web
// @Tags webchat
// @Accept json
// @Produce json
// @Param request body services.WebchatConfig true "Configuración a validar"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webchat/validate [post]
func (h *WebchatSetupHandler) ValidateWebchatConfig(c *gin.Context) {
	var config services.WebchatConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.webchatService.ValidateWebchatConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Configuration validation failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webchat configuration is valid",
		Data:    config,
	})
}
