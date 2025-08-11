package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
)

// TawkToHandler maneja las rutas de Tawk.to
type TawkToHandler struct {
	tawkToService *services.TawkToService
	logger        logger.Logger
}

// NewTawkToHandler crea una nueva instancia del handler de Tawk.to
func NewTawkToHandler(tawkToService *services.TawkToService, logger logger.Logger) *TawkToHandler {
	return &TawkToHandler{
		tawkToService: tawkToService,
		logger:        logger,
	}
}

// SetupTawkToIntegration configura la integración de Tawk.to
func (h *TawkToHandler) SetupTawkToIntegration(c *gin.Context) {
	var request struct {
		TenantID string `json:"tenant_id" binding:"required"`
		Config   struct {
			WidgetID     string `json:"widget_id" binding:"required"`
			PropertyID   string `json:"property_id" binding:"required"`
			APIKey       string `json:"api_key" binding:"required"`
			BaseURL      string `json:"base_url"`
			CustomCSS    string `json:"custom_css,omitempty"`
			CustomJS     string `json:"custom_js,omitempty"`
			Greeting     string `json:"greeting,omitempty"`
			OfflineMsg   string `json:"offline_msg,omitempty"`
		} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Datos de configuración inválidos",
		})
		return
	}

	// Crear configuración de Tawk.to
	tawkToConfig := &services.TawkToConfig{
		WidgetID:   request.Config.WidgetID,
		PropertyID: request.Config.PropertyID,
		APIKey:     request.Config.APIKey,
		BaseURL:    request.Config.BaseURL,
		CustomCSS:  request.Config.CustomCSS,
		CustomJS:   request.Config.CustomJS,
		Greeting:   request.Config.Greeting,
		OfflineMsg: request.Config.OfflineMsg,
	}

	// Configurar integración
	integration, err := h.tawkToService.SetupTawkToIntegration(request.TenantID, tawkToConfig)
	if err != nil {
		h.logger.Error("Error configurando integración Tawk.to", "error", err, "tenant_id", request.TenantID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SETUP_ERROR",
			"message": "Error configurando integración: " + err.Error(),
		})
		return
	}

	h.logger.Info("Integración Tawk.to configurada exitosamente", "tenant_id", request.TenantID, "integration_id", integration.ID)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"integration_id": integration.ID,
			"status":         integration.Status,
			"message":        "Integración Tawk.to configurada exitosamente",
		},
	})
}

// GetTawkToConfig obtiene la configuración de Tawk.to
func (h *TawkToHandler) GetTawkToConfig(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "tenant_id es requerido",
		})
		return
	}

	config, err := h.tawkToService.GetTawkToConfig(tenantID)
	if err != nil {
		h.logger.Error("Error obteniendo configuración Tawk.to", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "CONFIG_NOT_FOUND",
			"message": "Configuración no encontrada: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateTawkToConfig actualiza la configuración de Tawk.to
func (h *TawkToHandler) UpdateTawkToConfig(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "tenant_id es requerido",
		})
		return
	}

	var request struct {
		WidgetID     string `json:"widget_id,omitempty"`
		PropertyID   string `json:"property_id,omitempty"`
		APIKey       string `json:"api_key,omitempty"`
		BaseURL      string `json:"base_url,omitempty"`
		CustomCSS    string `json:"custom_css,omitempty"`
		CustomJS     string `json:"custom_js,omitempty"`
		Greeting     string `json:"greeting,omitempty"`
		OfflineMsg   string `json:"offline_msg,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Datos de configuración inválidos",
		})
		return
	}

	// Obtener configuración actual
	currentConfig, err := h.tawkToService.GetTawkToConfig(tenantID)
	if err != nil {
		h.logger.Error("Error obteniendo configuración actual", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "CONFIG_NOT_FOUND",
			"message": "Configuración no encontrada: " + err.Error(),
		})
		return
	}

	// Actualizar campos proporcionados
	if request.WidgetID != "" {
		currentConfig.WidgetID = request.WidgetID
	}
	if request.PropertyID != "" {
		currentConfig.PropertyID = request.PropertyID
	}
	if request.APIKey != "" {
		currentConfig.APIKey = request.APIKey
	}
	if request.BaseURL != "" {
		currentConfig.BaseURL = request.BaseURL
	}
	if request.CustomCSS != "" {
		currentConfig.CustomCSS = request.CustomCSS
	}
	if request.CustomJS != "" {
		currentConfig.CustomJS = request.CustomJS
	}
	if request.Greeting != "" {
		currentConfig.Greeting = request.Greeting
	}
	if request.OfflineMsg != "" {
		currentConfig.OfflineMsg = request.OfflineMsg
	}

	// Actualizar configuración
	if err := h.tawkToService.UpdateTawkToConfig(tenantID, currentConfig); err != nil {
		h.logger.Error("Error actualizando configuración Tawk.to", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "UPDATE_ERROR",
			"message": "Error actualizando configuración: " + err.Error(),
		})
		return
	}

	h.logger.Info("Configuración Tawk.to actualizada exitosamente", "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuración actualizada exitosamente",
		"data":    currentConfig,
	})
}

// TawkToWebhookHandler maneja los webhooks de Tawk.to
func (h *TawkToHandler) TawkToWebhookHandler(c *gin.Context) {
	// Leer payload
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Error leyendo payload del webhook", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Error leyendo payload",
		})
		return
	}

	// Obtener firma del webhook
	signature := c.GetHeader("X-Tawk-Signature")

	// Procesar webhook
	message, err := h.tawkToService.ProcessTawkToWebhook(payload, signature)
	if err != nil {
		h.logger.Error("Error procesando webhook de Tawk.to", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "WEBHOOK_ERROR",
			"message": "Error procesando webhook: " + err.Error(),
		})
		return
	}

	h.logger.Info("Webhook de Tawk.to procesado exitosamente", "message_id", message.MessageID, "platform", message.Platform)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook procesado exitosamente",
	})
}

// GetTawkToAnalytics obtiene analytics de Tawk.to
func (h *TawkToHandler) GetTawkToAnalytics(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "tenant_id es requerido",
		})
		return
	}

	// Parsear fechas
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "start_date y end_date son requeridos",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_DATE",
			"message": "Formato de fecha inválido (YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_DATE",
			"message": "Formato de fecha inválido (YYYY-MM-DD)",
		})
		return
	}

	// Obtener analytics
	analytics, err := h.tawkToService.GetTawkToAnalytics(tenantID, startDate, endDate)
	if err != nil {
		h.logger.Error("Error obteniendo analytics de Tawk.to", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ANALYTICS_ERROR",
			"message": "Error obteniendo analytics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetTawkToSessions obtiene sesiones de chat de Tawk.to
func (h *TawkToHandler) GetTawkToSessions(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "tenant_id es requerido",
		})
		return
	}

	// Parsear límite
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_LIMIT",
			"message": "Límite inválido",
		})
		return
	}

	// Obtener sesiones
	sessions, err := h.tawkToService.GetTawkToSessions(tenantID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo sesiones de Tawk.to", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SESSIONS_ERROR",
			"message": "Error obteniendo sesiones: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}
