package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"it-integration-service/internal/models"
	"it-integration-service/internal/services"

	"github.com/gin-gonic/gin"
)

// PaymentController maneja las rutas HTTP para los pagos
type PaymentController struct {
	paymentService *services.PaymentService
	webhookService *services.MercadoPagoWebhookService
}

// NewPaymentController crea una nueva instancia del controlador de pagos
func NewPaymentController(paymentService *services.PaymentService, webhookService *services.MercadoPagoWebhookService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
		webhookService: webhookService,
	}
}

// CreatePayment maneja la creación de un nuevo pago
// @Summary Crear un nuevo pago
// @Description Crea un nuevo pago usando Mercado Pago Checkout Pro
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body models.PaymentRequest true "Información del pago"
// @Success 201 {object} models.PaymentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments [post]
func (pc *PaymentController) CreatePayment(c *gin.Context) {
	var request models.PaymentRequest

	// Validar el cuerpo de la solicitud
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Datos de pago inválidos: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Crear el pago
	payment, err := pc.paymentService.CreatePayment(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Error al procesar el pago: " + err.Error(),
			Code:    "PAYMENT_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPayment maneja la obtención de información de un pago
// @Summary Obtener información de un pago
// @Description Obtiene la información detallada de un pago específico
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "ID del pago"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/{id} [get]
func (pc *PaymentController) GetPayment(c *gin.Context) {
	// Obtener el ID del pago de los parámetros de la URL
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "ID de pago inválido",
			Code:    "INVALID_PAYMENT_ID",
		})
		return
	}

	// Obtener el pago
	payment, err := pc.paymentService.GetPayment(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Error al obtener el pago: " + err.Error(),
			Code:    "PAYMENT_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// RefundPayment maneja el reembolso de un pago
// @Summary Reembolsar un pago
// @Description Procesa un reembolso total o parcial de un pago
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "ID del pago"
// @Param amount body map[string]float64 true "Monto a reembolsar"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/{id}/refund [post]
func (pc *PaymentController) RefundPayment(c *gin.Context) {
	// Obtener el ID del pago
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "ID de pago inválido",
			Code:    "INVALID_PAYMENT_ID",
		})
		return
	}

	// Obtener el monto del reembolso
	var refundRequest struct {
		Amount float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refundRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Monto de reembolso inválido: " + err.Error(),
			Code:    "INVALID_REFUND_AMOUNT",
		})
		return
	}

	// Validar el monto
	if refundRequest.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "El monto del reembolso debe ser mayor a 0",
			Code:    "INVALID_REFUND_AMOUNT",
		})
		return
	}

	// Procesar el reembolso
	err = pc.paymentService.RefundPayment(paymentID, refundRequest.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Error al procesar el reembolso: " + err.Error(),
			Code:    "REFUND_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Reembolso procesado exitosamente",
		"payment_id": paymentID,
		"amount":     refundRequest.Amount,
	})
}

// WebhookHandler maneja las notificaciones de webhook de Mercado Pago
// @Summary Webhook de Mercado Pago
// @Description Maneja las notificaciones de webhook de Mercado Pago
// @Tags webhooks
// @Accept json
// @Produce json
// @Param notification body models.WebhookNotification true "Notificación del webhook"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/mercadopago [post]
func (pc *PaymentController) WebhookHandler(c *gin.Context) {
	// Leer el cuerpo de la solicitud
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Error al leer el cuerpo de la solicitud: " + err.Error(),
			Code:    "INVALID_REQUEST_BODY",
		})
		return
	}

	// Validar la firma del webhook si está configurada
	if pc.webhookService != nil {
		valid, err := pc.webhookService.ValidateWebhookSignature(c.Request, body)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "Error al validar la firma del webhook: " + err.Error(),
				Code:    "WEBHOOK_SIGNATURE_ERROR",
			})
			return
		}
		if !valid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Firma del webhook inválida",
				Code:    "INVALID_WEBHOOK_SIGNATURE",
			})
			return
		}
	}

	// Parsear la notificación
	var notification map[string]interface{}
	if err := json.Unmarshal(body, &notification); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Error al parsear la notificación: " + err.Error(),
			Code:    "INVALID_NOTIFICATION_FORMAT",
		})
		return
	}

	// Procesar la notificación
	webhookNotification, err := pc.webhookService.ProcessWebhookNotification(notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Error al procesar la notificación: " + err.Error(),
			Code:    "NOTIFICATION_PROCESSING_ERROR",
		})
		return
	}

	// Procesar según el tipo de notificación
	switch webhookNotification.Type {
	case "payment":
		// Procesar notificación de pago
		if err := pc.processPaymentNotification(webhookNotification); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Error al procesar notificación de pago: " + err.Error(),
				Code:    "PAYMENT_NOTIFICATION_ERROR",
			})
			return
		}
	case "merchant_order":
		// Procesar notificación de orden
		if err := pc.processMerchantOrderNotification(webhookNotification); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Error al procesar notificación de orden: " + err.Error(),
				Code:    "ORDER_NOTIFICATION_ERROR",
			})
			return
		}
	default:
		// Tipo de notificación no soportado
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Tipo de notificación no soportado: " + webhookNotification.Type,
			Code:    "UNSUPPORTED_NOTIFICATION_TYPE",
		})
		return
	}

	// Responder con éxito
	c.JSON(http.StatusOK, gin.H{
		"message": "Notificación procesada exitosamente",
		"id":      webhookNotification.ID,
		"type":    webhookNotification.Type,
	})
}

// processPaymentNotification procesa una notificación de pago
func (pc *PaymentController) processPaymentNotification(notification *services.WebhookNotification) error {
	// Obtener el ID del pago
	paymentID, ok := notification.Data["id"].(string)
	if !ok {
		return fmt.Errorf("payment ID not found in notification data")
	}

	// Aquí puedes implementar la lógica específica para procesar pagos
	// Por ejemplo, actualizar el estado en tu base de datos, enviar emails, etc.
	
	// Log de la notificación
	fmt.Printf("Procesando notificación de pago: ID=%s, Action=%s\n", paymentID, notification.Action)
	
	return nil
}

// processMerchantOrderNotification procesa una notificación de orden
func (pc *PaymentController) processMerchantOrderNotification(notification *services.WebhookNotification) error {
	// Obtener el ID de la orden
	orderID, ok := notification.Data["id"].(string)
	if !ok {
		return fmt.Errorf("order ID not found in notification data")
	}

	// Aquí puedes implementar la lógica específica para procesar órdenes
	// Por ejemplo, actualizar el estado en tu base de datos, enviar emails, etc.
	
	// Log de la notificación
	fmt.Printf("Procesando notificación de orden: ID=%s, Action=%s\n", orderID, notification.Action)
	
	return nil
}
