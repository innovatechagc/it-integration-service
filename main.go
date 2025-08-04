package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/microservice-template/internal/config"
	"github.com/company/microservice-template/internal/handlers"
	"github.com/company/microservice-template/internal/middleware"
	"github.com/company/microservice-template/internal/repository"
	"github.com/company/microservice-template/internal/services"
	"github.com/company/microservice-template/pkg/logger"
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
	outboundRepo := repository.NewOutboundMessageLogRepository(db)
	
	// Inicializar servicios
	healthService := services.NewHealthService()
	webhookService := services.NewWebhookService(cfg.Integration.MessagingServiceURL, logger)
	providerService := services.NewMessagingProviderService(logger)
	
	// Servicio de integración con repositorios reales
	integrationService := services.NewIntegrationService(
		channelRepo,
		inboundRepo,
		outboundRepo,
		webhookService,
		providerService,
		logger,
	)
	
	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Metrics())
	
	// Rutas
	handlers.SetupRoutes(router, healthService, integrationService, logger)
	
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