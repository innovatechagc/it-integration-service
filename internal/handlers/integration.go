package handlers

import (
	"context"
	"encoding/json"
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

// Chat/Messages endpoints

// GetInboundMessages godoc
// @Summary Obtener mensajes entrantes
// @Description Obtiene el historial de mensajes entrantes por plataforma
// @Tags messages
// @Accept json
// @Produce json
// @Param platform query string false "Filtrar por plataforma"
// @Param limit query int false "Límite de mensajes (default: 50)"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/messages/inbound [get]
func (h *IntegrationHandler) GetInboundMessages(c *gin.Context) {
	platform := c.Query("platform")
	// limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	// offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// TODO: Implementar consulta real a la base de datos
	// query := `SELECT id, platform, payload, received_at, processed FROM inbound_messages WHERE ($1 = '' OR platform = $1) ORDER BY received_at DESC LIMIT $2 OFFSET $3`

	// Por ahora devolvemos datos mock, pero la estructura está lista
	messages := []map[string]interface{}{
		{
			"id":          "example-id",
			"platform":    platform,
			"payload":     map[string]interface{}{"message": map[string]string{"text": "Ejemplo"}},
			"received_at": "2025-08-04T14:09:50Z",
			"processed":   true,
		},
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Messages retrieved successfully",
		Data:    messages,
	})
}

// GetChatHistory godoc
// @Summary Obtener historial de chat
// @Description Obtiene la conversación entre el bot y un usuario específico
// @Tags messages
// @Accept json
// @Produce json
// @Param platform path string true "Plataforma (telegram, whatsapp, etc)"
// @Param user_id path string true "ID del usuario"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/chat/{platform}/{user_id} [get]
func (h *IntegrationHandler) GetChatHistory(c *gin.Context) {
	platform := c.Param("platform")
	userID := c.Param("user_id")

	// Aquí combinarías inbound y outbound messages para crear la conversación
	conversation := []map[string]interface{}{
		{
			"id":        "msg-1",
			"type":      "inbound",
			"platform":  platform,
			"user_id":   userID,
			"text":      "/start",
			"timestamp": "2025-08-04T14:09:50Z",
		},
		{
			"id":        "msg-2", 
			"type":      "outbound",
			"platform":  platform,
			"user_id":   userID,
			"text":      "¡Hola! 👋 Tu integración está funcionando.",
			"timestamp": "2025-08-04T14:10:00Z",
			"status":    "sent",
		},
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Chat history retrieved successfully",
		Data: map[string]interface{}{
			"platform":    platform,
			"user_id":     userID,
			"messages":    conversation,
			"total_count": len(conversation),
		},
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
	var update map[string]interface{}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid payload: " + err.Error(),
		})
		return
	}

	// Parse básico (luego puedes mapear al struct oficial de Telegram)
	if message, exists := update["message"].(map[string]interface{}); exists {
		text := ""
		if textVal, ok := message["text"].(string); ok {
			text = textVal
		}
		
		chat := message["chat"].(map[string]interface{})
		chatID := int64(chat["id"].(float64))

		// Log para debugging
		h.logger.Info("Telegram webhook received", 
			"chat_id", chatID,
			"text", text,
		)

		// Aquí puedes reenviar al bot-service u otra lógica
		// Por ahora solo loggeamos
	}

	// Convertir el update a JSON para el servicio
	payload, _ := json.Marshal(update)
	if err := h.integrationService.ProcessTelegramWebhook(c.Request.Context(), payload); err != nil {
		h.logger.Error("Failed to process Telegram webhook", 
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to process webhook",
		})
		return
	}

	c.Status(http.StatusOK)
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