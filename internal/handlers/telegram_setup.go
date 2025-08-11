package handlers

import (
	"net/http"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type TelegramSetupHandler struct {
	telegramService    *services.TelegramSetupService
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewTelegramSetupHandler(telegramService *services.TelegramSetupService, integrationService services.IntegrationService, logger logger.Logger) *TelegramSetupHandler {
	return &TelegramSetupHandler{
		telegramService:    telegramService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// TelegramSetupRequest representa la solicitud para configurar Telegram
type TelegramSetupRequest struct {
	BotToken   string `json:"bot_token" binding:"required"`
	WebhookURL string `json:"webhook_url" binding:"required"`
	TenantID   string `json:"tenant_id" binding:"required"`
}

// TelegramBotInfoResponse representa la respuesta con información del bot
type TelegramBotInfoResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	IsBot     bool   `json:"is_bot"`
}

// TelegramWebhookInfoResponse representa la respuesta con información del webhook
type TelegramWebhookInfoResponse struct {
	URL                string   `json:"url"`
	PendingUpdateCount int      `json:"pending_update_count"`
	LastErrorMessage   string   `json:"last_error_message,omitempty"`
	AllowedUpdates     []string `json:"allowed_updates,omitempty"`
}

// GetBotInfo godoc
// @Summary Obtener información del bot de Telegram
// @Description Obtiene información básica del bot usando el token
// @Tags telegram
// @Accept json
// @Produce json
// @Param bot_token query string true "Token del bot de Telegram"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/telegram/bot-info [get]
func (h *TelegramSetupHandler) GetBotInfo(c *gin.Context) {
	botToken := c.Query("bot_token")
	if botToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "bot_token is required",
		})
		return
	}

	botInfo, err := h.telegramService.GetBotInfo(c.Request.Context(), botToken)
	if err != nil {
		h.logger.Error("Failed to get bot info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "BOT_ERROR",
			Message: "Failed to get bot info: " + err.Error(),
		})
		return
	}

	response := TelegramBotInfoResponse{
		ID:        botInfo.ID,
		Username:  botInfo.Username,
		FirstName: botInfo.FirstName,
		IsBot:     botInfo.IsBot,
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot info retrieved successfully",
		Data:    response,
	})
}

// SetupTelegramIntegration godoc
// @Summary Configurar integración completa de Telegram
// @Description Configura el bot, webhook y crea la integración en una sola operación
// @Tags telegram
// @Accept json
// @Produce json
// @Param request body TelegramSetupRequest true "Datos de configuración"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/telegram/setup [post]
func (h *TelegramSetupHandler) SetupTelegramIntegration(c *gin.Context) {
	var request TelegramSetupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Crear integración usando el servicio de Telegram
	integration, err := h.telegramService.CreateTelegramIntegration(
		c.Request.Context(),
		request.BotToken,
		request.WebhookURL,
		request.TenantID,
	)
	if err != nil {
		h.logger.Error("Failed to create Telegram integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SETUP_ERROR",
			Message: "Failed to setup Telegram integration: " + err.Error(),
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
		Message: "Telegram integration configured successfully",
		Data:    integration,
	})
}

// GetWebhookInfo godoc
// @Summary Obtener información del webhook configurado
// @Description Obtiene el estado actual del webhook del bot
// @Tags telegram
// @Accept json
// @Produce json
// @Param bot_token query string true "Token del bot de Telegram"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/telegram/webhook-info [get]
func (h *TelegramSetupHandler) GetWebhookInfo(c *gin.Context) {
	botToken := c.Query("bot_token")
	if botToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "bot_token is required",
		})
		return
	}

	webhookInfo, err := h.telegramService.GetWebhookInfo(c.Request.Context(), botToken)
	if err != nil {
		h.logger.Error("Failed to get webhook info", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to get webhook info: " + err.Error(),
		})
		return
	}

	response := TelegramWebhookInfoResponse{
		URL:                webhookInfo.URL,
		PendingUpdateCount: webhookInfo.PendingUpdateCount,
		LastErrorMessage:   webhookInfo.LastErrorMessage,
		AllowedUpdates:     webhookInfo.AllowedUpdates,
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook info retrieved successfully",
		Data:    response,
	})
}

// SetWebhook godoc
// @Summary Configurar webhook del bot
// @Description Configura la URL del webhook para recibir actualizaciones
// @Tags telegram
// @Accept json
// @Produce json
// @Param request body map[string]string true "bot_token y webhook_url"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/telegram/webhook [post]
func (h *TelegramSetupHandler) SetWebhook(c *gin.Context) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	botToken, exists := request["bot_token"]
	if !exists || botToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "bot_token is required",
		})
		return
	}

	webhookURL, exists := request["webhook_url"]
	if !exists || webhookURL == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "webhook_url is required",
		})
		return
	}

	if err := h.telegramService.SetWebhook(c.Request.Context(), botToken, webhookURL); err != nil {
		h.logger.Error("Failed to set webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to set webhook: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook configured successfully",
		Data: map[string]string{
			"webhook_url": webhookURL,
		},
	})
}

// DeleteWebhook godoc
// @Summary Eliminar webhook del bot
// @Description Elimina la configuración del webhook
// @Tags telegram
// @Accept json
// @Produce json
// @Param bot_token query string true "Token del bot de Telegram"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/telegram/webhook [delete]
func (h *TelegramSetupHandler) DeleteWebhook(c *gin.Context) {
	botToken := c.Query("bot_token")
	if botToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "bot_token is required",
		})
		return
	}

	if err := h.telegramService.DeleteWebhook(c.Request.Context(), botToken); err != nil {
		h.logger.Error("Failed to delete webhook", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "WEBHOOK_ERROR",
			Message: "Failed to delete webhook: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Webhook deleted successfully",
	})
}

// ValidateToken godoc
// @Summary Validar token del bot
// @Description Valida que el token del bot sea válido
// @Tags telegram
// @Accept json
// @Produce json
// @Param request body map[string]string true "bot_token"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/telegram/validate-token [post]
func (h *TelegramSetupHandler) ValidateToken(c *gin.Context) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	botToken, exists := request["bot_token"]
	if !exists || botToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "bot_token is required",
		})
		return
	}

	if err := h.telegramService.ValidateBotToken(c.Request.Context(), botToken); err != nil {
		h.logger.Error("Failed to validate bot token", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Failed to validate bot token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot token validated successfully",
	})
}
