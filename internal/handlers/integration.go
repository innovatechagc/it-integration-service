package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/company/microservice-template/internal/domain"
	"github.com/company/microservice-template/internal/services"
	"github.com/company/microservice-template/pkg/logger"
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

// GetChannels godoc
// @Summary Listar integraciones de canales
// @Description Obtiene todas las integraciones activas por tenant
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
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get channels",
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
// @Summary Obtener detalles de integración
// @Description Obtiene los detalles de una integración específica
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID de la integración"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [get]
func (h *IntegrationHandler) GetChannel(c *gin.Context) {
	id := c.Param("id")
	
	channel, err := h.integrationService.GetChannel(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get channel", err)
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Channel not found",
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
// @Summary Registrar nueva integración
// @Description Registra una nueva integración de canal
// @Tags integrations
// @Accept json
// @Produce json
// @Param integration body domain.ChannelIntegration true "Datos de la integración"
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
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create channel",
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
// @Summary Actualizar integración
// @Description Actualiza una integración existente
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID de la integración"
// @Param integration body domain.ChannelIntegration true "Datos actualizados"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [patch]
func (h *IntegrationHandler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	
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
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update channel",
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
// @Summary Eliminar integración
// @Description Desactiva o elimina una integración
// @Tags integrations
// @Accept json
// @Produce json
// @Param id path string true "ID de la integración"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/channels/{id} [delete]
func (h *IntegrationHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.integrationService.DeleteChannel(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete channel", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete channel",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Channel deleted successfully",
	})
}

// SendMessage godoc
// @Summary Enviar mensaje
// @Description Envía un mensaje a través de un canal específico
// @Tags integrations
// @Accept json
// @Produce json
// @Param request body domain.SendMessageRequest true "Datos del mensaje"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/send [post]
func (h *IntegrationHandler) SendMessage(c *gin.Context) {
	var request domain.SendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.integrationService.SendMessage(c.Request.Context(), &request); err != nil {
		h.logger.Error("Failed to send message", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SEND_ERROR",
			Message: "Failed to send message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Message sent successfully",
	})
}

// Webhook handlers

// WhatsAppWebhook godoc
// @Summary Webhook para WhatsApp
// @Description Procesa webhooks entrantes de WhatsApp
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/whatsapp [post]
func (h *IntegrationHandler) WhatsAppWebhook(c *gin.Context) {
	h.processWebhook(c, h.integrationService.ProcessWhatsAppWebhook)
}

// MessengerWebhook godoc
// @Summary Webhook para Messenger
// @Description Procesa webhooks entrantes de Facebook Messenger
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/messenger [post]
func (h *IntegrationHandler) MessengerWebhook(c *gin.Context) {
	// Verificación de webhook de Facebook
	if c.Request.Method == "GET" {
		h.verifyFacebookWebhook(c)
		return
	}
	h.processWebhook(c, h.integrationService.ProcessMessengerWebhook)
}

// InstagramWebhook godoc
// @Summary Webhook para Instagram
// @Description Procesa webhooks entrantes de Instagram
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/instagram [post]
func (h *IntegrationHandler) InstagramWebhook(c *gin.Context) {
	// Verificación de webhook de Facebook
	if c.Request.Method == "GET" {
		h.verifyFacebookWebhook(c)
		return
	}
	h.processWebhook(c, h.integrationService.ProcessInstagramWebhook)
}

// TelegramWebhook godoc
// @Summary Webhook para Telegram
// @Description Procesa webhooks entrantes de Telegram
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/telegram [post]
func (h *IntegrationHandler) TelegramWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Failed to read request body",
		})
		return
	}

	if err := h.integrationService.ProcessTelegramWebhook(c.Request.Context(), payload); err != nil {
		h.logger.Error("Failed to process Telegram webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
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
// @Summary Webhook para Webchat
// @Description Procesa webhooks entrantes del webchat embebido
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /integrations/webhooks/webchat [post]
func (h *IntegrationHandler) WebchatWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Failed to read request body",
		})
		return
	}

	if err := h.integrationService.ProcessWebchatWebhook(c.Request.Context(), payload); err != nil {
		h.logger.Error("Failed to process Webchat webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

// Helper functions

func (h *IntegrationHandler) processWebhook(c *gin.Context, processor func(ctx context.Context, payload []byte, signature string) error) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Failed to read request body",
		})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		signature = c.GetHeader("X-Hub-Signature")
	}

	if err := processor(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("Failed to process webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook processed successfully",
	})
}

func (h *IntegrationHandler) verifyFacebookWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	// TODO: Validar el verify_token contra la configuración
	if mode == "subscribe" && token != "" {
		challengeInt, err := strconv.Atoi(challenge)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, challengeInt)
		return
	}

	c.Status(http.StatusForbidden)
}