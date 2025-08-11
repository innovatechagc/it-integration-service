package handlers

import (
	"net/http"
	"time"

	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// MailchimpSetupHandler maneja las operaciones de configuración de Mailchimp
type MailchimpSetupHandler struct {
	mailchimpService    *services.MailchimpSetupService
	integrationService  services.IntegrationService
	logger              logger.Logger
}

// NewMailchimpSetupHandler crea una nueva instancia del handler de Mailchimp
func NewMailchimpSetupHandler(
	mailchimpService *services.MailchimpSetupService,
	integrationService services.IntegrationService,
	logger logger.Logger,
) *MailchimpSetupHandler {
	return &MailchimpSetupHandler{
		mailchimpService:   mailchimpService,
		integrationService: integrationService,
		logger:             logger,
	}
}

// SetupMailchimpRequest representa la solicitud de configuración de Mailchimp
type SetupMailchimpRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	APIKey      string `json:"api_key" binding:"required"`
	ServerPrefix string `json:"server_prefix" binding:"required"`
	AudienceID  string `json:"audience_id" binding:"required"`
	DataCenter  string `json:"data_center"`
	WebhookURL  string `json:"webhook_url"`
}

// GetAccountInfoResponse representa la respuesta con información de la cuenta
type GetAccountInfoResponse struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	Enabled     bool   `json:"enabled"`
}

// GetAudienceInfoResponse representa la respuesta con información de la audiencia
type GetAudienceInfoResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	EmailType         string `json:"email_type"`
	Status            string `json:"status"`
	SubscriberCount   int    `json:"subscriber_count"`
	UnsubscribeCount  int    `json:"unsubscribe_count"`
	CleanCount        int    `json:"clean_count"`
	MemberCount       int    `json:"member_count"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// GetAccountInfo obtiene información de la cuenta de Mailchimp
func (h *MailchimpSetupHandler) GetAccountInfo(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id es requerido"})
		return
	}

	// Obtener configuración del tenant
	config, err := h.mailchimpService.GetMailchimpConfig(tenantID)
	if err != nil {
		h.logger.Error("Error obteniendo configuración de Mailchimp", "error", err.Error(), "tenant_id", tenantID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuración de Mailchimp no encontrada"})
		return
	}

	// Obtener información de la cuenta
	accountInfo, err := h.mailchimpService.GetAccountInfo(config)
	if err != nil {
		h.logger.Error("Error obteniendo información de cuenta", "error", err.Error(), "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo información de cuenta"})
		return
	}

	response := GetAccountInfoResponse{
		AccountID:   accountInfo.AccountID,
		AccountName: accountInfo.AccountName,
		Email:       accountInfo.Email,
		Username:    accountInfo.Username,
		Role:        accountInfo.Role,
		Enabled:     accountInfo.Enabled,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetAudienceInfo obtiene información de la audiencia de Mailchimp
func (h *MailchimpSetupHandler) GetAudienceInfo(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id es requerido"})
		return
	}

	// Obtener configuración del tenant
	config, err := h.mailchimpService.GetMailchimpConfig(tenantID)
	if err != nil {
		h.logger.Error("Error obteniendo configuración de Mailchimp", "error", err.Error(), "tenant_id", tenantID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuración de Mailchimp no encontrada"})
		return
	}

	// Obtener información de la audiencia
	audienceInfo, err := h.mailchimpService.GetAudienceInfo(config)
	if err != nil {
		h.logger.Error("Error obteniendo información de audiencia", "error", err.Error(), "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo información de audiencia"})
		return
	}

	response := GetAudienceInfoResponse{
		ID:               audienceInfo.ID,
		Name:             audienceInfo.Name,
		EmailType:        audienceInfo.EmailType,
		Status:           audienceInfo.Status,
		SubscriberCount:  audienceInfo.SubscriberCount,
		UnsubscribeCount: audienceInfo.UnsubscribeCount,
		CleanCount:       audienceInfo.CleanCount,
		MemberCount:      audienceInfo.MemberCount,
		CreatedAt:        audienceInfo.CreatedAt,
		UpdatedAt:        audienceInfo.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// SetupMailchimp configura la integración de Mailchimp
func (h *MailchimpSetupHandler) SetupMailchimp(c *gin.Context) {
	var req SetupMailchimpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada inválidos: " + err.Error()})
		return
	}

	// Crear configuración de Mailchimp
	config := &services.MailchimpConfig{
		APIKey:       req.APIKey,
		ServerPrefix: req.ServerPrefix,
		AudienceID:   req.AudienceID,
		DataCenter:   req.DataCenter,
		WebhookURL:   req.WebhookURL,
		UpdatedAt:    time.Now(),
	}

	// Configurar integración
	integration, err := h.mailchimpService.SetupMailchimpIntegration(req.TenantID, config)
	if err != nil {
		h.logger.Error("Error configurando integración de Mailchimp", "error", err.Error(), "tenant_id", req.TenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error configurando integración: " + err.Error()})
		return
	}

	// Obtener información de la cuenta para la respuesta
	accountInfo, err := h.mailchimpService.GetAccountInfo(config)
	if err != nil {
		h.logger.Warn("Error obteniendo información de cuenta para respuesta", "error", err.Error())
	}

	// Obtener información de la audiencia para la respuesta
	audienceInfo, err := h.mailchimpService.GetAudienceInfo(config)
	if err != nil {
		h.logger.Warn("Error obteniendo información de audiencia para respuesta", "error", err.Error())
	}

	response := gin.H{
		"success": true,
		"message": "Integración de Mailchimp configurada exitosamente",
		"data": gin.H{
			"integration_id": integration.ID,
			"platform":       integration.Platform,
			"status":         integration.Status,
			"account": gin.H{
				"account_id":   accountInfo.AccountID,
				"account_name": accountInfo.AccountName,
				"email":        accountInfo.Email,
			},
			"audience": gin.H{
				"id":                audienceInfo.ID,
				"name":              audienceInfo.Name,
				"subscriber_count":  audienceInfo.SubscriberCount,
				"member_count":      audienceInfo.MemberCount,
			},
		},
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateMailchimpConfig actualiza la configuración de Mailchimp
func (h *MailchimpSetupHandler) UpdateMailchimpConfig(c *gin.Context) {
	var req SetupMailchimpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada inválidos: " + err.Error()})
		return
	}

	// Crear configuración de Mailchimp
	config := &services.MailchimpConfig{
		APIKey:       req.APIKey,
		ServerPrefix: req.ServerPrefix,
		AudienceID:   req.AudienceID,
		DataCenter:   req.DataCenter,
		WebhookURL:   req.WebhookURL,
		UpdatedAt:    time.Now(),
	}

	// Actualizar configuración
	err := h.mailchimpService.UpdateMailchimpConfig(req.TenantID, config)
	if err != nil {
		h.logger.Error("Error actualizando configuración de Mailchimp", "error", err.Error(), "tenant_id", req.TenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando configuración: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuración de Mailchimp actualizada exitosamente",
	})
}

// GetMailchimpAnalytics obtiene analytics de Mailchimp
func (h *MailchimpSetupHandler) GetMailchimpAnalytics(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id es requerido"})
		return
	}

	// Parsear fechas
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Use YYYY-MM-DD"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Últimos 30 días por defecto
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Use YYYY-MM-DD"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Obtener analytics
	analytics, err := h.mailchimpService.GetMailchimpAnalytics(tenantID, startDate, endDate)
	if err != nil {
		h.logger.Error("Error obteniendo analytics de Mailchimp", "error", err.Error(), "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo analytics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"analytics":   analytics,
			"start_date":  startDate.Format("2006-01-02"),
			"end_date":    endDate.Format("2006-01-02"),
			"tenant_id":   tenantID,
		},
	})
}

// ProcessMailchimpWebhook procesa los webhooks de Mailchimp
func (h *MailchimpSetupHandler) ProcessMailchimpWebhook(c *gin.Context) {
	// Leer el payload
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Error leyendo payload del webhook", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error leyendo payload"})
		return
	}

	// Obtener firma del header
	signature := c.GetHeader("X-Mailchimp-Signature")

	// Procesar webhook
	normalizedMessage, err := h.mailchimpService.ProcessMailchimpWebhook(payload, signature)
	if err != nil {
		h.logger.Error("Error procesando webhook de Mailchimp", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error procesando webhook: " + err.Error()})
		return
	}

	// Reenviar al servicio de mensajería
	if err := h.integrationService.ProcessMailchimpWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("Error reenviando mensaje al servicio de mensajería", "error", err.Error())
		// No retornamos error aquí para no fallar el webhook
	}

	h.logger.Info("Webhook de Mailchimp procesado exitosamente", map[string]interface{}{
		"message_id": normalizedMessage.MessageID,
		"type":       normalizedMessage.Content.Type,
		"recipient":  normalizedMessage.Recipient,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook procesado exitosamente",
		"data": gin.H{
			"message_id": normalizedMessage.MessageID,
			"type":       normalizedMessage.Content.Type,
		},
	})
}
