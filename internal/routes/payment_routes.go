package routes

import (
	"it-integration-service/internal/controllers"

	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes configura las rutas para los pagos
func SetupPaymentRoutes(router *gin.Engine, paymentController *controllers.PaymentController) {
	// Grupo de rutas para pagos
	payments := router.Group("/api/v1/payments")
	{
		// Crear un nuevo pago
		payments.POST("/", paymentController.CreatePayment)

		// Obtener información de un pago específico
		payments.GET("/:id", paymentController.GetPayment)

		// Reembolsar un pago
		payments.POST("/:id/refund", paymentController.RefundPayment)
	}

	// Grupo de rutas para webhooks
	webhooks := router.Group("/api/v1/webhooks")
	{
		// Webhook de Mercado Pago
		webhooks.POST("/mercadopago", paymentController.WebhookHandler)
	}
}
