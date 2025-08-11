package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/controllers"
	"it-integration-service/internal/handlers"
	"it-integration-service/internal/middleware"
	"it-integration-service/internal/repository"
	"it-integration-service/internal/routes"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// @title Microservice Template API
// @version 1.0
// @description Template para microservicios en Go
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Cargar configuración
	cfg := config.Load()

	// Inicializar logger
	logger := logger.NewLogger(cfg.LogLevel)

	// Inicializar cliente de Vault (comentado para testing)
	// vaultClient, err := vault.NewClient(cfg.VaultConfig)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize Vault client", err)
	// }

	// Inicializar conexión a base de datos
	db, err := repository.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Inicializar repositorios
	channelRepo := repository.NewChannelIntegrationRepository(db)
	inboundRepo := repository.NewInboundMessageRepository(db)

	// Inicializar servicios
	healthService := services.NewHealthService(db.DB, logger)
	webhookService := services.NewWebhookService(cfg.Integration.MessagingServiceURL, logger)
	channelService := services.NewChannelService(channelRepo, logger)

	// Inicializar servicio de encriptación
	// encryptionService, err := services.NewEncryptionService(cfg.Integration.EncryptionKey)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize encryption service", err)
	// }

	// Inicializar servicio de rotación de tokens
	tokenRotationService := services.NewTokenRotationService(channelRepo, logger)

	// Servicio de integración (solo para integraciones, no envío de mensajes)
	integrationService := services.NewIntegrationService(
		channelService,
		inboundRepo,
		webhookService,
		logger,
	)

	// Inicializar configuración de Mercado Pago
	mpConfig, err := config.NewMercadoPagoConfig()
	if err != nil {
		logger.Fatal("Failed to initialize Mercado Pago configuration", err)
	}

	// Inicializar servicios de pago
	paymentService := services.NewPaymentService(mpConfig)
	mpWebhookService := services.NewMercadoPagoWebhookService(mpConfig.SecretKey)
	paymentController := controllers.NewPaymentController(paymentService, mpWebhookService)

	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Metrics())
	router.Use(middleware.RateLimit(cfg.Integration.RateLimitRPS, cfg.Integration.RateLimitBurst))

	// Programar rotación automática de tokens
	tokenConfig := tokenRotationService.GetTokenRotationConfig()
	if err := tokenRotationService.ScheduleTokenRotation(context.Background(), tokenConfig); err != nil {
		logger.Error("Failed to schedule token rotation", err)
	}

	// Rutas
	handlers.SetupRoutes(router, healthService, integrationService, logger, cfg, db)

	// Rutas de pagos
	routes.SetupPaymentRoutes(router, paymentController)

	// Servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		logger.Info("Starting server on port " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
}
