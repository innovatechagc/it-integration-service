package handlers

import (
	"net/http"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type MessengerSetupHandler struct {
	messengerService   *services.MessengerSetupService
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewMessengerSetupHandler(messengerService *services.MessengerSetupService, integrationService services.IntegrationService, logger logger.Logger) *MessengerSetupHandler {
	return &MessengerSetupHandler{
		messengerService:   messengerService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// MessengerSetupRequest representa la solicitud para configurar Messenger
type MessengerSetupRequest struct {
	PageAccessToken string `json:"page_access_token" binding:"required"`
	PageID          string `json:"page_id" binding:"required"`
	WebhookURL      string `json:"webhook_url" binding:"required"`
	TenantID        string `json:"tenant_id" binding:"required"`
}

// MessengerPageInfoResponse representa la respuesta con información de la página
type MessengerPageInfoResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	About    string `json:"about"`
	Website  string `json:"website"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Picture  string `json:"picture,omitempty"`
}

// GetPageInfo godoc
// @Summary Obtener información de la página de Facebook
// @Description Obtiene información de la página de Facebook para Messenger
// @Tags messenger
// @Accept json
// @Produce json
// @Param page_access_token query string true "Token de acceso de la página"
// @Param page_id query string true "ID de la página de Facebook"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/messenger/page-info [get]
func (h *MessengerSetupHandler) GetPageInfo(c *gin.Context) {
	pageAccessToken := c.Query("page_access_token")
	pageID := c.Query("page_id")

	if pageAccessToken == "" || pageID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "page_access_token and page_id are required",
		})
		return
	}

	pageInfo, err := h.messengerService.GetPageInfo(c.Request.Context(), pageAccessToken, pageID)
	if err != nil {
		h.logger.Error("Failed to get page info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "PAGE_ERROR",
			Message: "Failed to get page info: " + err.Error(),
		})
		return
	}

	response := MessengerPageInfoResponse{
		ID:       pageInfo.ID,
		Name:     pageInfo.Name,
		Category: pageInfo.Category,
		About:    pageInfo.About,
		Website:  pageInfo.Website,
		Phone:    pageInfo.Phone,
		Email:    pageInfo.Email,
	}

	if pageInfo.Picture.Data.URL != "" {
		response.Picture = pageInfo.Picture.Data.URL
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Page info retrieved successfully",
		Data:    response,
	})
}

// SetupMessengerIntegration godoc
// @Summary Configurar integración completa de Messenger
// @Description Configura la página, verifica permisos y crea la integración
// @Tags messenger
// @Accept json
// @Produce json
// @Param request body MessengerSetupRequest true "Datos de configuración"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/messenger/setup [post]
func (h *MessengerSetupHandler) SetupMessengerIntegration(c *gin.Context) {
	var request MessengerSetupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Crear integración usando el servicio de Messenger
	integration, err := h.messengerService.CreateMessengerIntegration(
		c.Request.Context(),
		request.PageAccessToken,
		request.PageID,
		request.WebhookURL,
		request.TenantID,
	)
	if err != nil {
		h.logger.Error("Failed to create Messenger integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SETUP_ERROR",
			Message: "Failed to setup Messenger integration: " + err.Error(),
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
		Message: "Messenger integration configured successfully",
		Data:    integration,
	})
}

// TestMessage godoc
// @Summary Enviar mensaje de prueba por Messenger
// @Description Envía un mensaje de prueba a un usuario específico
// @Tags messenger
// @Accept json
// @Produce json
// @Param request body map[string]string true "page_access_token, recipient_id y text"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/messenger/test-message [post]
func (h *MessengerSetupHandler) TestMessage(c *gin.Context) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	pageAccessToken, exists := request["page_access_token"]
	if !exists || pageAccessToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "page_access_token is required",
		})
		return
	}

	recipientID, exists := request["recipient_id"]
	if !exists || recipientID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "recipient_id is required",
		})
		return
	}

	text, exists := request["text"]
	if !exists || text == "" {
		text = "🤖 ¡Hola! Este es un mensaje de prueba desde Messenger API."
	}

	if err := h.messengerService.SendMessage(c.Request.Context(), pageAccessToken, recipientID, text); err != nil {
		h.logger.Error("Failed to send test message", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "MESSAGE_ERROR",
			Message: "Failed to send test message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Test message sent successfully",
		Data: map[string]string{
			"recipient_id": recipientID,
			"text":         text,
		},
	})
}

// ValidateWebhook godoc
// @Summary Validar webhook de Messenger
// @Description Valida el token de verificación del webhook (usado por Facebook)
// @Tags messenger
// @Accept json
// @Produce json
// @Param hub.mode query string true "Modo de verificación"
// @Param hub.verify_token query string true "Token de verificación"
// @Param hub.challenge query string true "Challenge de verificación"
// @Success 200 {string} string "Challenge response"
// @Router /integrations/messenger/webhook-verify [get]
func (h *MessengerSetupHandler) ValidateWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	expectedToken := "messenger-it-app-webhook-verify-token" // Debería venir de configuración

	if mode == "subscribe" && h.messengerService.ValidateWebhookToken(token, expectedToken) {
		h.logger.Info("Messenger webhook verified successfully", map[string]interface{}{
			"verify_token": token,
			"challenge":    challenge,
		})
		c.String(http.StatusOK, challenge)
		return
	}

	h.logger.Warn("Messenger webhook verification failed", map[string]interface{}{
		"mode":           mode,
		"provided_token": token,
		"expected_token": expectedToken,
	})

	c.JSON(http.StatusForbidden, domain.APIResponse{
		Code:    "VERIFICATION_FAILED",
		Message: "Webhook verification failed",
	})
}