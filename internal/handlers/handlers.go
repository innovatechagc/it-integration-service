package handlers

import (
	"net/http"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/middleware"
	"it-integration-service/internal/repository"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	healthService services.HealthService
	logger        logger.Logger
}

func SetupRoutes(router *gin.Engine, healthService services.HealthService, integrationService services.IntegrationService, logger logger.Logger, cfg *config.Config, db *repository.PostgresDB) {
	h := &Handler{
		healthService: healthService,
		logger:        logger,
	}

	// Integration handler
	integrationHandler := NewIntegrationHandler(integrationService, logger)

	// Setup handlers para configuración específica de plataformas
	telegramSetupService := services.NewTelegramSetupService(logger)
	telegramSetupHandler := NewTelegramSetupHandler(telegramSetupService, integrationService, logger)

	whatsappSetupService := services.NewWhatsAppSetupService(logger)
	whatsappSetupHandler := NewWhatsAppSetupHandler(whatsappSetupService, integrationService, logger)

	messengerSetupService := services.NewMessengerSetupService(logger)
	messengerSetupHandler := NewMessengerSetupHandler(messengerSetupService, integrationService, logger)

	instagramSetupService := services.NewInstagramSetupService(logger)
	instagramSetupHandler := NewInstagramSetupHandler(instagramSetupService, integrationService, logger)

	webchatSetupService := services.NewWebchatSetupService(logger)
	webchatSetupHandler := NewWebchatSetupHandler(webchatSetupService, integrationService, logger)

	// Tawk.to service (usando el repositorio directamente)
	channelRepo := repository.NewChannelIntegrationRepository(db)
	tawkToSetupService := services.NewTawkToService(&cfg.TawkTo, channelRepo, logger)
	tawkToSetupHandler := NewTawkToHandler(tawkToSetupService, logger)

	// Mailchimp service
	mailchimpSetupService := services.NewMailchimpSetupService(&cfg.Mailchimp, channelRepo, logger)
	mailchimpSetupHandler := NewMailchimpSetupHandler(mailchimpSetupService, integrationService, logger)

	// Webhook validation middleware
	webhookValidation := middleware.NewWebhookValidationMiddleware(cfg, logger)

	// Swagger documentation (protegido en producción)
	router.GET("/swagger/*any", middleware.SwaggerAuth(), ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Métricas de Prometheus
	router.GET("/metrics", middleware.MetricsHandler())

	// API routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", h.HealthCheck)
		api.GET("/ready", h.ReadinessCheck)

		// Integration routes
		integrations := api.Group("/integrations")
		{
			// Channel management
			integrations.GET("/channels", integrationHandler.GetChannels)
			integrations.GET("/channels/:id", integrationHandler.GetChannel)
			integrations.POST("/channels", integrationHandler.CreateChannel)
			integrations.PATCH("/channels/:id", integrationHandler.UpdateChannel)
			integrations.DELETE("/channels/:id", integrationHandler.DeleteChannel)

			// Message validation (solo para validar integraciones)
			integrations.GET("/messages/inbound", integrationHandler.GetInboundMessages)

			// Platform-specific setup routes
			telegram := integrations.Group("/telegram")
			{
				telegram.GET("/bot-info", telegramSetupHandler.GetBotInfo)
				telegram.POST("/setup", telegramSetupHandler.SetupTelegramIntegration)
				telegram.GET("/webhook-info", telegramSetupHandler.GetWebhookInfo)
				telegram.POST("/webhook", telegramSetupHandler.SetWebhook)
				telegram.DELETE("/webhook", telegramSetupHandler.DeleteWebhook)
				telegram.POST("/validate-token", telegramSetupHandler.ValidateToken)
			}

			whatsapp := integrations.Group("/whatsapp")
			{
				whatsapp.GET("/business-info", whatsappSetupHandler.GetBusinessInfo)
				whatsapp.GET("/phone-info", whatsappSetupHandler.GetPhoneNumberInfo)
				whatsapp.POST("/setup", whatsappSetupHandler.SetupWhatsAppIntegration)
				whatsapp.POST("/test-message", whatsappSetupHandler.TestMessage)
				whatsapp.GET("/webhook-verify", whatsappSetupHandler.ValidateWebhook)
			}

			messenger := integrations.Group("/messenger")
			{
				messenger.GET("/page-info", messengerSetupHandler.GetPageInfo)
				messenger.POST("/setup", messengerSetupHandler.SetupMessengerIntegration)
				messenger.POST("/test-message", messengerSetupHandler.TestMessage)
				messenger.GET("/webhook-verify", messengerSetupHandler.ValidateWebhook)
			}

			instagram := integrations.Group("/instagram")
			{
				instagram.GET("/account-info", instagramSetupHandler.GetInstagramAccountInfo)
				instagram.GET("/page-info", instagramSetupHandler.GetPageInfo)
				instagram.GET("/accounts", instagramSetupHandler.GetInstagramAccounts)
				instagram.POST("/setup", instagramSetupHandler.SetupInstagramIntegration)
				instagram.POST("/test-message", instagramSetupHandler.TestMessage)
				instagram.GET("/webhook-verify", instagramSetupHandler.ValidateWebhook)
			}

			webchat := integrations.Group("/webchat")
			{
				webchat.POST("/setup", webchatSetupHandler.SetupWebchatIntegration)
				webchat.GET("/config", webchatSetupHandler.GetWebchatConfig)
				webchat.PUT("/config", webchatSetupHandler.UpdateWebchatConfig)
				webchat.POST("/sessions", webchatSetupHandler.CreateWebchatSession)
				webchat.GET("/sessions", webchatSetupHandler.GetWebchatSessions)
				webchat.POST("/messages", webchatSetupHandler.SendWebchatMessage)
				webchat.GET("/stats", webchatSetupHandler.GetWebchatStats)
				webchat.POST("/validate", webchatSetupHandler.ValidateWebchatConfig)
			}

			tawkto := integrations.Group("/tawkto")
			{
				tawkto.POST("/setup", tawkToSetupHandler.SetupTawkToIntegration)
				tawkto.GET("/config/:tenant_id", tawkToSetupHandler.GetTawkToConfig)
				tawkto.PUT("/config/:tenant_id", tawkToSetupHandler.UpdateTawkToConfig)
				tawkto.GET("/analytics/:tenant_id", tawkToSetupHandler.GetTawkToAnalytics)
				tawkto.GET("/sessions/:tenant_id", tawkToSetupHandler.GetTawkToSessions)
			}

			mailchimp := integrations.Group("/mailchimp")
			{
				mailchimp.GET("/account-info", mailchimpSetupHandler.GetAccountInfo)
				mailchimp.GET("/audience-info", mailchimpSetupHandler.GetAudienceInfo)
				mailchimp.POST("/setup", mailchimpSetupHandler.SetupMailchimp)
				mailchimp.PUT("/config", mailchimpSetupHandler.UpdateMailchimpConfig)
				mailchimp.GET("/analytics", mailchimpSetupHandler.GetMailchimpAnalytics)
			}

			// Webhooks
			webhooks := integrations.Group("/webhooks")
			{
				// WhatsApp webhooks con validación
				webhooks.GET("/whatsapp", webhookValidation.ValidateWebhookVerification("whatsapp"), integrationHandler.WhatsAppWebhook)
				webhooks.POST("/whatsapp", webhookValidation.ValidateWebhookSignature("whatsapp"), integrationHandler.WhatsAppWebhook)

				// Messenger webhooks con validación
				webhooks.GET("/messenger", webhookValidation.ValidateWebhookVerification("messenger"), integrationHandler.MessengerWebhook)
				webhooks.POST("/messenger", webhookValidation.ValidateWebhookSignature("messenger"), integrationHandler.MessengerWebhook)

				// Instagram webhooks con validación
				webhooks.GET("/instagram", webhookValidation.ValidateWebhookVerification("instagram"), integrationHandler.InstagramWebhook)
				webhooks.POST("/instagram", webhookValidation.ValidateWebhookSignature("instagram"), integrationHandler.InstagramWebhook)

				// Telegram webhooks con validación
				webhooks.POST("/telegram", webhookValidation.ValidateTelegramWebhook(), integrationHandler.TelegramWebhook)

				// Webchat webhooks (sin validación específica por ahora)
				webhooks.POST("/webchat", integrationHandler.WebchatWebhook)

				// Tawk.to webhooks con validación
				webhooks.POST("/tawkto", webhookValidation.ValidateWebhookSignature("tawkto"), tawkToSetupHandler.TawkToWebhookHandler)

				// Mailchimp webhooks con validación
				webhooks.POST("/mailchimp", webhookValidation.ValidateWebhookSignature("mailchimp"), integrationHandler.MailchimpWebhook)
			}
		}
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Verifica el estado del servicio
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	status := h.healthService.CheckHealth()

	response := domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Service is healthy",
		Data:    status,
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Verifica si el servicio está listo para recibir tráfico
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ready [get]
func (h *Handler) ReadinessCheck(c *gin.Context) {
	status := h.healthService.CheckReadiness()

	if status.Status == "ready" {
		response := domain.APIResponse{
			Code:    "SUCCESS",
			Message: "Service is ready",
			Data:    status,
		}
		c.JSON(http.StatusOK, response)
	} else {
		response := domain.APIResponse{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Service is not ready",
			Data:    status,
		}
		c.JSON(http.StatusServiceUnavailable, response)
	}
}
