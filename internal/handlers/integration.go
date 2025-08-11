package handlers

import (
	"net/http"
	"strconv"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewIntegrationHandler(integrationService services.IntegrationService, logger logger.Logger) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		logger:             logger,
	}
}

// Channel Management

// GetChannels godoc
// @Summary Obtener canales de integración
// @Description Obtiene todos los canales de integración para un tenant
// @Tags integrations
// @Accept json
// @Produce json
// @Param tenant_id query string true "ID del tenant"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels [get]
func (h *IntegrationHandler) GetChannels(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "tenant_id is required",
		})
		return
	}

	channels, err := h.integrationService.GetChannelsByTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get channels", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get channels: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channels retrieved successfully",
		Data:    channels,
	})
}

// GetChannel godoc
// @Summary Obtener canal específico
// @Description Obtiene un canal de integración por ID
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID del canal"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [get]
func (h *IntegrationHandler) GetChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Channel ID is required",
		})
		return
	}

	channel, err := h.integrationService.GetChannel(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get channel", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get channel: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channel retrieved successfully",
		Data:    channel,
	})
}

// CreateChannel godoc
// @Summary Crear canal de integración
// @Description Crea un nuevo canal de integración
// @Tags integrations
// @Accept json
// @Produce json
// @Param request body domain.ChannelIntegration true "Datos del canal"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/channels [post]
func (h *IntegrationHandler) CreateChannel(c *gin.Context) {
	var integration domain.ChannelIntegration
	if err := c.ShouldBindJSON(&integration); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.integrationService.CreateChannel(c.Request.Context(), &integration); err != nil {
		h.logger.Error("Failed to create channel", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "CREATE_ERROR",
			Message: "Failed to create channel: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channel created successfully",
		Data:    integration,
	})
}

// UpdateChannel godoc
// @Summary Actualizar canal de integración
// @Description Actualiza un canal de integración existente
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID del canal"
// @Param request body domain.ChannelIntegration true "Datos actualizados del canal"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [patch]
func (h *IntegrationHandler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Channel ID is required",
		})
		return
	}

	var integration domain.ChannelIntegration
	if err := c.ShouldBindJSON(&integration); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	integration.ID = id
	if err := h.integrationService.UpdateChannel(c.Request.Context(), &integration); err != nil {
		h.logger.Error("Failed to update channel", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "UPDATE_ERROR",
			Message: "Failed to update channel: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channel updated successfully",
		Data:    integration,
	})
}

// DeleteChannel godoc
// @Summary Eliminar canal de integración
// @Description Elimina un canal de integración
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID del canal"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [delete]
func (h *IntegrationHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Channel ID is required",
		})
		return
	}

	if err := h.integrationService.DeleteChannel(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete channel", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "DELETE_ERROR",
			Message: "Failed to delete channel: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channel deleted successfully",
	})
}

// Message History (solo para validación)

// GetInboundMessages godoc
// @Summary Obtener mensajes entrantes
// @Description Obtiene mensajes entrantes para validación de integración
// @Tags integrations
// @Accept json
// @Produce json
// @Param platform query string true "Plataforma"
// @Param limit query int false "Límite de resultados" default(10)
// @Success 200 {object} domain.APIResponse
// @Router /integrations/messages/inbound [get]
func (h *IntegrationHandler) GetInboundMessages(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Platform is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	messages, err := h.integrationService.GetInboundMessages(c.Request.Context(), platform, limit, 0)
	if err != nil {
		h.logger.Error("Failed to get inbound messages", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "FETCH_ERROR",
			Message: "Failed to get inbound messages: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Messages retrieved successfully",
		Data:    messages,
	})
}

// Webhook Handlers

// WhatsAppWebhook godoc
// @Summary Webhook de WhatsApp
// @Description Procesa webhooks de WhatsApp
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/whatsapp [post]
func (h *IntegrationHandler) WhatsAppWebhook(c *gin.Context) {
	if c.Request.Method == "GET" {
		// Verificación de webhook
		mode := c.Query("hub.mode")
		token := c.Query("hub.verify_token")
		challenge := c.Query("hub.challenge")

		if mode == "subscribe" && token == "test-token" {
			c.String(http.StatusOK, challenge)
			return
		}

		c.JSON(http.StatusForbidden, domain.APIResponse{
			Code:    "VERIFICATION_FAILED",
			Message: "Webhook verification failed",
		})
		return
	}

	// Procesamiento de webhook
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook payload", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_PAYLOAD",
			Message: "Failed to read webhook payload",
		})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if err := h.integrationService.ProcessWhatsAppWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("Failed to process WhatsApp webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

// MessengerWebhook godoc
// @Summary Webhook de Messenger
// @Description Procesa webhooks de Messenger
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/messenger [post]
func (h *IntegrationHandler) MessengerWebhook(c *gin.Context) {
	if c.Request.Method == "GET" {
		// Verificación de webhook
		mode := c.Query("hub.mode")
		token := c.Query("hub.verify_token")
		challenge := c.Query("hub.challenge")

		if mode == "subscribe" && token == "test-token" {
			c.String(http.StatusOK, challenge)
			return
		}

		c.JSON(http.StatusForbidden, domain.APIResponse{
			Code:    "VERIFICATION_FAILED",
			Message: "Webhook verification failed",
		})
		return
	}

	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook payload", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_PAYLOAD",
			Message: "Failed to read webhook payload",
		})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if err := h.integrationService.ProcessMessengerWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("Failed to process Messenger webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

// InstagramWebhook godoc
// @Summary Webhook de Instagram
// @Description Procesa webhooks de Instagram
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/instagram [post]
func (h *IntegrationHandler) InstagramWebhook(c *gin.Context) {
	if c.Request.Method == "GET" {
		// Verificación de webhook
		mode := c.Query("hub.mode")
		token := c.Query("hub.verify_token")
		challenge := c.Query("hub.challenge")

		if mode == "subscribe" && token == "test-token" {
			c.String(http.StatusOK, challenge)
			return
		}

		c.JSON(http.StatusForbidden, domain.APIResponse{
			Code:    "VERIFICATION_FAILED",
			Message: "Webhook verification failed",
		})
		return
	}

	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook payload", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_PAYLOAD",
			Message: "Failed to read webhook payload",
		})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if err := h.integrationService.ProcessInstagramWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("Failed to process Instagram webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

// TelegramWebhook godoc
// @Summary Webhook de Telegram
// @Description Procesa webhooks de Telegram
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/telegram [post]
func (h *IntegrationHandler) TelegramWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook payload", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_PAYLOAD",
			Message: "Failed to read webhook payload",
		})
		return
	}

	if err := h.integrationService.ProcessTelegramWebhook(c.Request.Context(), payload); err != nil {
		h.logger.Error("Failed to process Telegram webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

// WebchatWebhook godoc
// @Summary Webhook de Webchat
// @Description Procesa webhooks de Webchat
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/webchat [post]
func (h *IntegrationHandler) WebchatWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook payload", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_PAYLOAD",
			Message: "Failed to read webhook payload",
		})
		return
	}

	if err := h.integrationService.ProcessWebchatWebhook(c.Request.Context(), payload); err != nil {
		h.logger.Error("Failed to process Webchat webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}
