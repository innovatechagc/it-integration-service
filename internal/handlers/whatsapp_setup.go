package handlers

import (
	"net/http"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type WhatsAppSetupHandler struct {
	whatsappService    *services.WhatsAppSetupService
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewWhatsAppSetupHandler(whatsappService *services.WhatsAppSetupService, integrationService services.IntegrationService, logger logger.Logger) *WhatsAppSetupHandler {
	return &WhatsAppSetupHandler{
		whatsappService:    whatsappService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// WhatsAppSetupRequest representa la solicitud para configurar WhatsApp
type WhatsAppSetupRequest struct {
	AccessToken        string `json:"access_token" binding:"required"`
	PhoneNumberID      string `json:"phone_number_id" binding:"required"`
	BusinessAccountID  string `json:"business_account_id" binding:"required"`
	WebhookURL         string `json:"webhook_url" binding:"required"`
	TenantID           string `json:"tenant_id" binding:"required"`
}

// WhatsAppBusinessInfoResponse representa la respuesta con información del negocio
type WhatsAppBusinessInfoResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
}

// WhatsAppPhoneInfoResponse representa la respuesta con información del teléfono
type WhatsAppPhoneInfoResponse struct {
	ID                     string `json:"id"`
	DisplayPhoneNumber     string `json:"display_phone_number"`
	VerifiedName           string `json:"verified_name"`
	CodeVerificationStatus string `json:"code_verification_status"`
	QualityRating          string `json:"quality_rating"`
	PlatformType           string `json:"platform_type"`
	ThroughputLevel        string `json:"throughput_level"`
}

// GetBusinessInfo godoc
// @Summary Obtener información del negocio de WhatsApp
// @Description Obtiene información de la cuenta de WhatsApp Business
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param access_token query string true "Token de acceso de Meta"
// @Param business_account_id query string true "ID de la cuenta de negocio"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/whatsapp/business-info [get]
func (h *WhatsAppSetupHandler) GetBusinessInfo(c *gin.Context) {
	accessToken := c.Query("access_token")
	businessAccountID := c.Query("business_account_id")

	if accessToken == "" || businessAccountID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "access_token and business_account_id are required",
		})
		return
	}

	businessInfo, err := h.whatsappService.GetBusinessInfo(c.Request.Context(), accessToken, businessAccountID)
	if err != nil {
		h.logger.Error("Failed to get business info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "BUSINESS_ERROR",
			Message: "Failed to get business info: " + err.Error(),
		})
		return
	}

	response := WhatsAppBusinessInfoResponse{
		ID:          businessInfo.ID,
		Name:        businessInfo.Name,
		Category:    businessInfo.Category,
		Description: businessInfo.Description,
		Website:     businessInfo.Website,
		Email:       businessInfo.Email,
		PhoneNumber: businessInfo.PhoneNumber,
		Status:      businessInfo.Status,
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Business info retrieved successfully",
		Data:    response,
	})
}

// GetPhoneNumberInfo godoc
// @Summary Obtener información del número de teléfono
// @Description Obtiene información del número de teléfono de WhatsApp Business
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param access_token query string true "Token de acceso de Meta"
// @Param phone_number_id query string true "ID del número de teléfono"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/whatsapp/phone-info [get]
func (h *WhatsAppSetupHandler) GetPhoneNumberInfo(c *gin.Context) {
	accessToken := c.Query("access_token")
	phoneNumberID := c.Query("phone_number_id")

	if accessToken == "" || phoneNumberID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "access_token and phone_number_id are required",
		})
		return
	}

	phoneInfo, err := h.whatsappService.GetPhoneNumberInfo(c.Request.Context(), accessToken, phoneNumberID)
	if err != nil {
		h.logger.Error("Failed to get phone number info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "PHONE_ERROR",
			Message: "Failed to get phone number info: " + err.Error(),
		})
		return
	}

	response := WhatsAppPhoneInfoResponse{
		ID:                     phoneInfo.ID,
		DisplayPhoneNumber:     phoneInfo.DisplayPhoneNumber,
		VerifiedName:           phoneInfo.VerifiedName,
		CodeVerificationStatus: phoneInfo.CodeVerificationStatus,
		QualityRating:          phoneInfo.QualityRating,
		PlatformType:           phoneInfo.PlatformType,
		ThroughputLevel:        phoneInfo.ThroughputLevel,
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Phone number info retrieved successfully",
		Data:    response,
	})
}

// SetupWhatsAppIntegration godoc
// @Summary Configurar integración completa de WhatsApp
// @Description Configura la cuenta, verifica permisos y crea la integración
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param request body WhatsAppSetupRequest true "Datos de configuración"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/whatsapp/setup [post]
func (h *WhatsAppSetupHandler) SetupWhatsAppIntegration(c *gin.Context) {
	var request WhatsAppSetupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Crear integración usando el servicio de WhatsApp
	integration, err := h.whatsappService.CreateWhatsAppIntegration(
		c.Request.Context(),
		request.AccessToken,
		request.PhoneNumberID,
		request.BusinessAccountID,
		request.WebhookURL,
		request.TenantID,
	)
	if err != nil {
		h.logger.Error("Failed to create WhatsApp integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SETUP_ERROR",
			Message: "Failed to setup WhatsApp integration: " + err.Error(),
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
		Message: "WhatsApp integration configured successfully",
		Data:    integration,
	})
}

// TestMessage godoc
// @Summary Enviar mensaje de prueba por WhatsApp
// @Description Envía un mensaje de prueba a un número específico
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param request body map[string]string true "access_token, phone_number_id, recipient y text"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/whatsapp/test-message [post]
func (h *WhatsAppSetupHandler) TestMessage(c *gin.Context) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	accessToken, exists := request["access_token"]
	if !exists || accessToken == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "access_token is required",
		})
		return
	}

	phoneNumberID, exists := request["phone_number_id"]
	if !exists || phoneNumberID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "phone_number_id is required",
		})
		return
	}

	recipient, exists := request["recipient"]
	if !exists || recipient == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "recipient is required",
		})
		return
	}

	text, exists := request["text"]
	if !exists || text == "" {
		text = "🤖 ¡Hola! Este es un mensaje de prueba de WhatsApp Business desde IT App Chat."
	}

	if err := h.whatsappService.SendMessage(c.Request.Context(), accessToken, phoneNumberID, recipient, text); err != nil {
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
			"recipient": recipient,
			"text":      text,
		},
	})
}

// ValidateWebhook godoc
// @Summary Validar webhook de WhatsApp
// @Description Valida el token de verificación del webhook (usado por Meta)
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param hub.mode query string true "Modo de verificación"
// @Param hub.verify_token query string true "Token de verificación"
// @Param hub.challenge query string true "Challenge de verificación"
// @Success 200 {string} string "Challenge response"
// @Router /integrations/whatsapp/webhook-verify [get]
func (h *WhatsAppSetupHandler) ValidateWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	expectedToken := "wpp-it-app-webhook-verify-token" // Debería venir de configuración

	if mode == "subscribe" && h.whatsappService.ValidateWebhookToken(token, expectedToken) {
		h.logger.Info("WhatsApp webhook verified successfully", map[string]interface{}{
			"verify_token": token,
			"challenge":    challenge,
		})
		c.String(http.StatusOK, challenge)
		return
	}

	h.logger.Warn("WhatsApp webhook verification failed", map[string]interface{}{
		"mode":           mode,
		"provided_token": token,
		"expected_token": expectedToken,
	})

	c.JSON(http.StatusForbidden, domain.APIResponse{
		Code:    "VERIFICATION_FAILED",
		Message: "Webhook verification failed",
	})
}