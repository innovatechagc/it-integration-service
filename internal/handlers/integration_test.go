package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"it-integration-service/internal/domain"
	"it-integration-service/internal/services"
	"it-integration-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock IntegrationService
type MockIntegrationService struct {
	mock.Mock
}

func (m *MockIntegrationService) CreateChannel(ctx gin.Context, integration *domain.ChannelIntegration) error {
	args := m.Called(ctx, integration)
	return args.Error(0)
}

func (m *MockIntegrationService) GetChannel(ctx gin.Context, id string) (*domain.ChannelIntegration, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.ChannelIntegration), args.Error(1)
}

func (m *MockIntegrationService) GetChannelsByTenant(ctx gin.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*domain.ChannelIntegration), args.Error(1)
}

func (m *MockIntegrationService) UpdateChannel(ctx gin.Context, integration *domain.ChannelIntegration) error {
	args := m.Called(ctx, integration)
	return args.Error(0)
}

func (m *MockIntegrationService) DeleteChannel(ctx gin.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIntegrationService) SendMessage(ctx gin.Context, request *domain.SendMessageRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockIntegrationService) ProcessWhatsAppWebhook(ctx gin.Context, payload []byte, signature string) error {
	args := m.Called(ctx, payload, signature)
	return args.Error(0)
}

func (m *MockIntegrationService) ProcessMessengerWebhook(ctx gin.Context, payload []byte, signature string) error {
	args := m.Called(ctx, payload, signature)
	return args.Error(0)
}

func (m *MockIntegrationService) ProcessInstagramWebhook(ctx gin.Context, payload []byte, signature string) error {
	args := m.Called(ctx, payload, signature)
	return args.Error(0)
}

func (m *MockIntegrationService) ProcessTelegramWebhook(ctx gin.Context, payload []byte) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockIntegrationService) ProcessWebchatWebhook(ctx gin.Context, payload []byte) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func setupTestRouter() (*gin.Engine, *MockIntegrationService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockService := &MockIntegrationService{}
	mockLogger := logger.NewLogger("debug")
	
	handler := NewIntegrationHandler(mockService, mockLogger)
	
	api := router.Group("/api/v1/integrations")
	{
		api.GET("/channels", handler.GetChannels)
		api.GET("/channels/:id", handler.GetChannel)
		api.POST("/channels", handler.CreateChannel)
		api.PATCH("/channels/:id", handler.UpdateChannel)
		api.DELETE("/channels/:id", handler.DeleteChannel)
		api.POST("/send", handler.SendMessage)
		
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/whatsapp", handler.WhatsAppWebhook)
			webhooks.POST("/messenger", handler.MessengerWebhook)
			webhooks.POST("/telegram", handler.TelegramWebhook)
		}
	}
	
	return router, mockService
}

func TestGetChannels(t *testing.T) {
	router, mockService := setupTestRouter()
	
	expectedChannels := []*domain.ChannelIntegration{
		{
			ID:       "channel-1",
			TenantID: "tenant-1",
			Platform: domain.PlatformWhatsApp,
			Provider: domain.ProviderMeta,
			Status:   domain.StatusActive,
		},
	}
	
	mockService.On("GetChannelsByTenant", mock.Anything, "tenant-1").Return(expectedChannels, nil)
	
	req, _ := http.NewRequest("GET", "/api/v1/integrations/channels?tenant_id=tenant-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response domain.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", response.Code)
	
	mockService.AssertExpectations(t)
}

func TestCreateChannel(t *testing.T) {
	router, mockService := setupTestRouter()
	
	integration := &domain.ChannelIntegration{
		TenantID: "tenant-1",
		Platform: domain.PlatformWhatsApp,
		Provider: domain.ProviderMeta,
		Status:   domain.StatusActive,
	}
	
	mockService.On("CreateChannel", mock.Anything, mock.AnythingOfType("*domain.ChannelIntegration")).Return(nil)
	
	jsonData, _ := json.Marshal(integration)
	req, _ := http.NewRequest("POST", "/api/v1/integrations/channels", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response domain.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", response.Code)
	
	mockService.AssertExpectations(t)
}

func TestSendMessage(t *testing.T) {
	router, mockService := setupTestRouter()
	
	sendRequest := &domain.SendMessageRequest{
		ChannelID: "channel-1",
		Recipient: "573001112233",
		Content: domain.MessageContent{
			Type: "text",
			Text: "Hello, World!",
		},
	}
	
	mockService.On("SendMessage", mock.Anything, mock.AnythingOfType("*domain.SendMessageRequest")).Return(nil)
	
	jsonData, _ := json.Marshal(sendRequest)
	req, _ := http.NewRequest("POST", "/api/v1/integrations/send", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response domain.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", response.Code)
	
	mockService.AssertExpectations(t)
}

func TestWhatsAppWebhook(t *testing.T) {
	router, mockService := setupTestRouter()
	
	webhookPayload := `{
		"entry": [{
			"changes": [{
				"value": {
					"messages": [{
						"id": "msg-123",
						"from": "573001112233",
						"timestamp": "1640995200",
						"text": {"body": "Hello"},
						"type": "text"
					}],
					"metadata": {
						"phone_number_id": "123456789"
					}
				}
			}]
		}]
	}`
	
	mockService.On("ProcessWhatsAppWebhook", mock.Anything, mock.AnythingOfType("[]uint8"), mock.AnythingOfType("string")).Return(nil)
	
	req, _ := http.NewRequest("POST", "/api/v1/integrations/webhooks/whatsapp", bytes.NewBufferString(webhookPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256=test-signature")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response domain.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", response.Code)
	
	mockService.AssertExpectations(t)
}