package routes

import (
	"it-integration-service/internal/config"
	"it-integration-service/internal/handlers"
	"it-integration-service/internal/middleware"
	"it-integration-service/internal/repository"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupGoogleCalendarRoutes configura las rutas de Google Calendar
func SetupGoogleCalendarRoutes(
	router *gin.Engine,
	cfg *config.Config,
	logger logger.Logger,
	googleCalendarRepo repository.GoogleCalendarRepository,
	encryptionService *services.EncryptionService,
) {
	// Crear servicios
	setupService := services.NewGoogleCalendarSetupService(
		&cfg.GoogleCalendar,
		googleCalendarRepo,
		logger,
		encryptionService,
	)

	eventService := services.NewGoogleCalendarService(
		&cfg.GoogleCalendar,
		setupService,
		googleCalendarRepo,
		logger,
		encryptionService,
	)

	// Crear handlers
	setupHandler := handlers.NewGoogleCalendarSetupHandler(setupService, &cfg.GoogleCalendar, logger)
	eventsHandler := handlers.NewGoogleCalendarEventsHandler(eventService, &cfg.GoogleCalendar, logger)

	// Grupo de rutas para Google Calendar
	googleCalendar := router.Group("/api/v1/integrations/google-calendar")
	{
		// Rutas de configuración OAuth2
		googleCalendar.POST("/auth", setupHandler.InitiateAuth)
		googleCalendar.GET("/callback", setupHandler.HandleCallback)
		googleCalendar.GET("/status/:channel_id", setupHandler.GetIntegrationStatus)
		googleCalendar.GET("/validate/:channel_id", setupHandler.ValidateToken)
		googleCalendar.POST("/refresh/:channel_id", setupHandler.RefreshToken)
		googleCalendar.POST("/webhook/setup", setupHandler.SetupWebhook)
		googleCalendar.POST("/revoke", setupHandler.RevokeAccess)
		googleCalendar.GET("/tenant/:tenant_id", setupHandler.GetIntegrationsByTenant)

		// Rutas de eventos
		events := googleCalendar.Group("/events")
		{
			events.GET("", eventsHandler.ListEvents)
			events.POST("", eventsHandler.CreateEvent)
			events.GET("/:event_id", eventsHandler.GetEvent)
			events.PUT("/:event_id", eventsHandler.UpdateEvent)
			events.DELETE("/:event_id", eventsHandler.DeleteEvent)
			events.POST("/sync", eventsHandler.SyncEvents)
			events.GET("/range/:channel_id", eventsHandler.GetEventsByDateRange)
			events.GET("/tenant/:tenant_id", eventsHandler.GetEventsByTenant)
		}
	}

	// Webhook endpoint (fuera del grupo de integraciones)
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/google-calendar", middleware.WebhookValidation(), eventsHandler.HandleWebhook)
	}

	logger.Info("Rutas de Google Calendar configuradas", map[string]interface{}{
		"base_path":    "/api/v1/integrations/google-calendar",
		"webhook_path": "/api/v1/webhooks/google-calendar",
	})
}

// SetupGoogleCalendarRoutesWithAuth configura las rutas con autenticación
func SetupGoogleCalendarRoutesWithAuth(
	router *gin.Engine,
	cfg *config.Config,
	logger logger.Logger,
	googleCalendarRepo repository.GoogleCalendarRepository,
	encryptionService *services.EncryptionService,
	authMiddleware gin.HandlerFunc,
) {
	// Crear servicios
	setupService := services.NewGoogleCalendarSetupService(
		&cfg.GoogleCalendar,
		googleCalendarRepo,
		logger,
		encryptionService,
	)

	eventService := services.NewGoogleCalendarService(
		&cfg.GoogleCalendar,
		setupService,
		googleCalendarRepo,
		logger,
		encryptionService,
	)

	// Crear handlers
	setupHandler := handlers.NewGoogleCalendarSetupHandler(setupService, &cfg.GoogleCalendar, logger)
	eventsHandler := handlers.NewGoogleCalendarEventsHandler(eventService, &cfg.GoogleCalendar, logger)

	// Grupo de rutas para Google Calendar con autenticación
	googleCalendar := router.Group("/api/v1/integrations/google-calendar")
	googleCalendar.Use(authMiddleware) // Aplicar middleware de autenticación
	{
		// Rutas de configuración OAuth2 (protegidas)
		googleCalendar.POST("/auth", setupHandler.InitiateAuth)
		googleCalendar.GET("/callback", setupHandler.HandleCallback)
		googleCalendar.GET("/status/:channel_id", setupHandler.GetIntegrationStatus)
		googleCalendar.GET("/validate/:channel_id", setupHandler.ValidateToken)
		googleCalendar.POST("/refresh/:channel_id", setupHandler.RefreshToken)
		googleCalendar.POST("/webhook/setup", setupHandler.SetupWebhook)
		googleCalendar.POST("/revoke", setupHandler.RevokeAccess)
		googleCalendar.GET("/tenant/:tenant_id", setupHandler.GetIntegrationsByTenant)

		// Rutas de eventos (protegidas)
		events := googleCalendar.Group("/events")
		{
			events.GET("", eventsHandler.ListEvents)
			events.POST("", eventsHandler.CreateEvent)
			events.GET("/:event_id", eventsHandler.GetEvent)
			events.PUT("/:event_id", eventsHandler.UpdateEvent)
			events.DELETE("/:event_id", eventsHandler.DeleteEvent)
			events.POST("/sync", eventsHandler.SyncEvents)
			events.GET("/range/:channel_id", eventsHandler.GetEventsByDateRange)
			events.GET("/tenant/:tenant_id", eventsHandler.GetEventsByTenant)
		}
	}

	// Webhook endpoint (sin autenticación, solo validación de webhook)
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/google-calendar", middleware.WebhookValidation(), eventsHandler.HandleWebhook)
	}

	logger.Info("Rutas de Google Calendar configuradas con autenticación", map[string]interface{}{
		"base_path":     "/api/v1/integrations/google-calendar",
		"webhook_path":  "/api/v1/webhooks/google-calendar",
		"auth_required": true,
	})
}
