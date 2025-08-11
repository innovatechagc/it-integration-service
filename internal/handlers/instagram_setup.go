package handlers

import (
	"net/http"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type InstagramSetupHandler struct {
	instagramService   *services.InstagramSetupService
	integrationService services.IntegrationService
	logger             logger.Logger
}

func NewInstagramSetupHandler(instagramService *services.InstagramSetupService, integrationService services.IntegrationService, logger logger.Logger) *InstagramSetupHandler {
	return &InstagramSetupHandler{
		instagramService:   instagramService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// InstagramSetupRequest representa la solicitud para configurar Instagram
type InstagramSetupRequest struct {
	PageAccessToken string `json:"page_access_token" binding:"required"`
	InstagramID     string `json:"instagram_id" binding:"required"`
	WebhookURL      string `json:"webhook_url" binding:"required"`
	TenantID        string `json:"tenant_id" binding:"required"`
}

// InstagramAccountInfoResponse representa la respuesta con información de la cuenta de Instagram
type InstagramAccountInfoResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Biography   string `json:"biography"`
	Website     string `json:"website"`
	ProfilePic  string `json:"profile_pic,omitempty"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	MediaCount  int    `json:"media_count"`
	AccountType string `json:"account_type"`
	IsPrivate   bool   `json:"is_private"`
	IsVerified  bool   `json:"is_verified"`
}

// InstagramPageInfoResponse representa la respuesta con información de la página de Facebook conectada
type InstagramPageInfoResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	About    string `json:"about"`
	Website  string `json:"website"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Picture  string `json:"picture,omitempty"`
}

// GetInstagramAccountInfo godoc
// @Summary Obtener información de la cuenta de Instagram
// @Description Obtiene información de la cuenta de Instagram Business
// @Tags instagram
// @Accept json
// @Produce json
// @Param page_access_token query string true "Token de acceso de la página de Facebook"
// @Param instagram_id query string true "ID de la cuenta de Instagram"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/instagram/account-info [get]
func (h *InstagramSetupHandler) GetInstagramAccountInfo(c *gin.Context) {
	pageAccessToken := c.Query("page_access_token")
	instagramID := c.Query("instagram_id")

	if pageAccessToken == "" || instagramID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "page_access_token and instagram_id are required",
		})
		return
	}

	accountInfo, err := h.instagramService.GetInstagramAccountInfo(c.Request.Context(), pageAccessToken, instagramID)
	if err != nil {
		h.logger.Error("Failed to get Instagram account info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "ACCOUNT_ERROR",
			Message: "Failed to get Instagram account info: " + err.Error(),
		})
		return
	}

	response := InstagramAccountInfoResponse{
		ID:          accountInfo.ID,
		Username:    accountInfo.Username,
		Name:        accountInfo.Name,
		Biography:   accountInfo.Biography,
		Website:     accountInfo.Website,
		Followers:   accountInfo.Followers,
		Following:   accountInfo.Following,
		MediaCount:  accountInfo.MediaCount,
		AccountType: accountInfo.AccountType,
		IsPrivate:   accountInfo.IsPrivate,
		IsVerified:  accountInfo.IsVerified,
	}

	if accountInfo.ProfilePic != "" {
		response.ProfilePic = accountInfo.ProfilePic
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Instagram account info retrieved successfully",
		Data:    response,
	})
}

// GetPageInfo godoc
// @Summary Obtener información de la página de Facebook conectada
// @Description Obtiene información de la página de Facebook que está conectada a Instagram
// @Tags instagram
// @Accept json
// @Produce json
// @Param page_access_token query string true "Token de acceso de la página"
// @Param page_id query string true "ID de la página de Facebook"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/instagram/page-info [get]
func (h *InstagramSetupHandler) GetPageInfo(c *gin.Context) {
	pageAccessToken := c.Query("page_access_token")
	pageID := c.Query("page_id")

	if pageAccessToken == "" || pageID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "page_access_token and page_id are required",
		})
		return
	}

	pageInfo, err := h.instagramService.GetPageInfo(c.Request.Context(), pageAccessToken, pageID)
	if err != nil {
		h.logger.Error("Failed to get page info", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "PAGE_ERROR",
			Message: "Failed to get page info: " + err.Error(),
		})
		return
	}

	response := InstagramPageInfoResponse{
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

// SetupInstagramIntegration godoc
// @Summary Configurar integración completa de Instagram
// @Description Configura la página, verifica permisos y crea la integración
// @Tags instagram
// @Accept json
// @Produce json
// @Param request body InstagramSetupRequest true "Datos de configuración"
// @Success 201 {object} domain.APIResponse
// @Router /integrations/instagram/setup [post]
func (h *InstagramSetupHandler) SetupInstagramIntegration(c *gin.Context) {
	var request InstagramSetupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Crear integración usando el servicio de Instagram
	integration, err := h.instagramService.CreateInstagramIntegration(
		c.Request.Context(),
		request.PageAccessToken,
		request.InstagramID,
		request.WebhookURL,
		request.TenantID,
	)
	if err != nil {
		h.logger.Error("Failed to create Instagram integration", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "SETUP_ERROR",
			Message: "Failed to setup Instagram integration: " + err.Error(),
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
		Message: "Instagram integration configured successfully",
		Data:    integration,
	})
}

// TestMessage godoc
// @Summary Enviar mensaje de prueba por Instagram
// @Description Envía un mensaje de prueba a un usuario específico
// @Tags instagram
// @Accept json
// @Produce json
// @Param request body map[string]string true "page_access_token, recipient_id y text"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/instagram/test-message [post]
func (h *InstagramSetupHandler) TestMessage(c *gin.Context) {
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
		text = "📸 ¡Hola! Este es un mensaje de prueba desde Instagram API."
	}

	if err := h.instagramService.SendMessage(c.Request.Context(), pageAccessToken, recipientID, text); err != nil {
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
// @Summary Validar webhook de Instagram
// @Description Valida el token de verificación del webhook (usado por Facebook)
// @Tags instagram
// @Accept json
// @Produce json
// @Param hub.mode query string true "Modo de verificación"
// @Param hub.verify_token query string true "Token de verificación"
// @Param hub.challenge query string true "Challenge de verificación"
// @Success 200 {string} string "Challenge response"
// @Router /integrations/instagram/webhook-verify [get]
func (h *InstagramSetupHandler) ValidateWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	expectedToken := "instagram-it-app-webhook-verify-token" // Debería venir de configuración

	if mode == "subscribe" && h.instagramService.ValidateWebhookToken(token, expectedToken) {
		h.logger.Info("Instagram webhook verified successfully", map[string]interface{}{
			"verify_token": token,
			"challenge":    challenge,
		})
		c.String(http.StatusOK, challenge)
		return
	}

	h.logger.Warn("Instagram webhook verification failed", map[string]interface{}{
		"mode":           mode,
		"provided_token": token,
		"expected_token": expectedToken,
	})

	c.JSON(http.StatusForbidden, domain.APIResponse{
		Code:    "VERIFICATION_FAILED",
		Message: "Webhook verification failed",
	})
}

// GetInstagramAccounts godoc
// @Summary Obtener cuentas de Instagram conectadas
// @Description Obtiene la lista de cuentas de Instagram conectadas a una página
// @Tags instagram
// @Accept json
// @Produce json
// @Param page_access_token query string true "Token de acceso de la página"
// @Param page_id query string true "ID de la página de Facebook"
// @Success 200 {object} domain.APIResponse
// @Router /integrations/instagram/accounts [get]
func (h *InstagramSetupHandler) GetInstagramAccounts(c *gin.Context) {
	pageAccessToken := c.Query("page_access_token")
	pageID := c.Query("page_id")

	if pageAccessToken == "" || pageID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "page_access_token and page_id are required",
		})
		return
	}

	accounts, err := h.instagramService.GetInstagramAccounts(c.Request.Context(), pageAccessToken, pageID)
	if err != nil {
		h.logger.Error("Failed to get Instagram accounts", err)
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "ACCOUNTS_ERROR",
			Message: "Failed to get Instagram accounts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Instagram accounts retrieved successfully",
		Data:    accounts,
	})
}
