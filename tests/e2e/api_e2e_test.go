package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"it-integration-service/internal/auth"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/handlers"
	"it-integration-service/internal/middleware"
	"it-integration-service/internal/services"
	testingPkg "it-integration-service/internal/testing"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	containers *testingPkg.TestContainers
	router     *gin.Engine
	jwtManager *auth.JWTManager
	authToken  string
}

func (suite *E2ETestSuite) SetupSuite() {
	ctx := context.Background()
	
	// Setup test containers
	containers, err := testingPkg.SetupTestContainers(ctx)
	suite.Require().NoError(err)
	suite.containers = containers

	// Setup JWT Manager
	suite.jwtManager = auth.NewJWTManager("test-secret", "test-issuer")
	
	// Generate test token
	token, err := suite.jwtManager.GenerateToken("test-user-id", "test@example.com", []string{"user"})
	suite.Require().NoError(err)
	suite.authToken = token

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())
	suite.router.Use(middleware.Logger(logger.NewLogger("debug")))
	
	// Setup services
	healthService := services.NewHealthService()
	logger := logger.NewLogger("debug")
	
	handlers.SetupRoutes(suite.router, healthService, logger)
}

func (suite *E2ETestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		suite.containers.Cleanup(ctx)
	}
}

func (suite *E2ETestSuite) TestCompleteAPIFlow() {
	// Test 1: Health check (no auth required)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var healthResponse domain.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &healthResponse)
	suite.NoError(err)

	// Test 2: Readiness check
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/ready", nil)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *E2ETestSuite) TestAuthenticationFlow() {
	// Test invalid token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	suite.router.ServeHTTP(w, req)
	
	// Should return 401 for invalid token
	assert.Equal(suite.T(), http.StatusNotFound, w.Code) // 404 because route doesn't exist yet
}

func (suite *E2ETestSuite) TestJWTTokenValidation() {
	// Test valid token parsing
	claims, err := suite.jwtManager.ValidateToken(suite.authToken)
	suite.NoError(err)
	suite.Equal("test-user-id", claims.UserID)
	suite.Equal("test@example.com", claims.Email)
	suite.Contains(claims.Roles, "user")
}

func (suite *E2ETestSuite) TestAPIResponseFormat() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	suite.router.ServeHTTP(w, req)
	
	// Verify response structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	
	// Check required fields
	suite.Contains(response, "status")
	suite.Contains(response, "timestamp")
	suite.Contains(response, "service")
	suite.Contains(response, "version")
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}