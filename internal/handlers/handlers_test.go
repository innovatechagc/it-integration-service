package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	healthService := services.NewHealthService()
	logger := logger.NewLogger("debug")
	
	SetupRoutes(router, healthService, logger)
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestReadinessCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	healthService := services.NewHealthService()
	logger := logger.NewLogger("debug")
	
	SetupRoutes(router, healthService, logger)
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ready", nil)
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ready")
}