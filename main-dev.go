package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/handlers"
	"it-integration-service/internal/middleware"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// @title Integration Service API
// @version 1.0
// @description Servicio de integraciones para múltiples plataformas de mensajería
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Cargar configuración
	cfg := config.Load()
	
	// Inicializar logger
	logger := logger.NewLogger(cfg.LogLevel)
	
	logger.Info("Starting Integration Service in development mode (without database)")
	
	// Inicializar servicios sin base de datos (usando mocks)
	healthService := services.NewHealthService()
	
	// Servicios de integración (usando mocks para desarrollo inicial)
	webhookService := services.NewWebhookService(cfg.Integration.MessagingServiceURL, logger)
	providerService := services.NewMessagingProviderService(logger)
	
	// Servicio de integración sin repositorios (usando mocks)
	integrationService := services.NewIntegrationService(
		nil, // channelRepo - usando mock interno
		nil, // inboundRepo - usando mock interno  
		nil, // outboundRepo - usando mock interno
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
		logger.Info("Available endpoints:")
		logger.Info("  - Health: http://localhost:" + cfg.Port + "/api/v1/health")
		logger.Info("  - Swagger: http://localhost:" + cfg.Port + "/swagger/index.html")
		logger.Info("  - Webhook Simulator: http://localhost:8081")
		
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