package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// MercadoPagoWebhookService maneja la validación de webhooks de Mercado Pago
type MercadoPagoWebhookService struct {
	secretKey string
}

// NewMercadoPagoWebhookService crea una nueva instancia del servicio de webhooks
func NewMercadoPagoWebhookService(secretKey string) *MercadoPagoWebhookService {
	return &MercadoPagoWebhookService{
		secretKey: secretKey,
	}
}

// ValidateWebhookSignature valida la firma del webhook según la documentación de Mercado Pago
func (s *MercadoPagoWebhookService) ValidateWebhookSignature(r *http.Request, body []byte) (bool, error) {
	// Obtener headers necesarios
	xSignature := r.Header.Get("x-signature")
	xRequestId := r.Header.Get("x-request-id")

	if xSignature == "" {
		return false, fmt.Errorf("x-signature header is missing")
	}

	// Extraer parámetros de la URL
	queryParams := r.URL.Query()
	dataID := queryParams.Get("data.id")

	// Parsear x-signature
	ts, hash, err := s.parseXSignature(xSignature)
	if err != nil {
		return false, fmt.Errorf("failed to parse x-signature: %w", err)
	}

	// Generar el template de firma
	manifest := s.generateManifest(dataID, xRequestId, ts)

	// Calcular HMAC
	expectedHash := s.calculateHMAC(manifest)

	// Comparar hashes
	if expectedHash != hash {
		return false, fmt.Errorf("signature validation failed")
	}

	// Validar timestamp (opcional: verificar que no sea muy antiguo)
	if err := s.validateTimestamp(ts); err != nil {
		return false, fmt.Errorf("timestamp validation failed: %w", err)
	}

	return true, nil
}

// parseXSignature extrae timestamp y hash del header x-signature
func (s *MercadoPagoWebhookService) parseXSignature(xSignature string) (string, string, error) {
	parts := strings.Split(xSignature, ",")
	var ts, hash string

	for _, part := range parts {
		keyValue := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			switch key {
			case "ts":
				ts = value
			case "v1":
				hash = value
			}
		}
	}

	if ts == "" || hash == "" {
		return "", "", fmt.Errorf("invalid x-signature format")
	}

	return ts, hash, nil
}

// generateManifest genera el template de firma según la documentación de Mercado Pago
func (s *MercadoPagoWebhookService) generateManifest(dataID, xRequestId, ts string) string {
	// Template: id:[data.id_url];request-id:[x-request-id_header];ts:[ts_header];
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, xRequestId, ts)
	return manifest
}

// calculateHMAC calcula el HMAC SHA256
func (s *MercadoPagoWebhookService) calculateHMAC(manifest string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(manifest))
	return hex.EncodeToString(h.Sum(nil))
}

// validateTimestamp valida que el timestamp no sea muy antiguo
func (s *MercadoPagoWebhookService) validateTimestamp(ts string) error {
	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Verificar que el timestamp no sea más antiguo que 5 minutos
	now := time.Now().Unix()
	if now-timestamp > 300 { // 5 minutos = 300 segundos
		return fmt.Errorf("timestamp is too old")
	}

	return nil
}

// ProcessWebhookNotification procesa una notificación de webhook
func (s *MercadoPagoWebhookService) ProcessWebhookNotification(notification map[string]interface{}) (*WebhookNotification, error) {
	// Validar campos requeridos
	id, ok := notification["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid notification id")
	}

	notificationType, ok := notification["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid notification type")
	}

	action, ok := notification["action"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid notification action")
	}

	data, ok := notification["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid notification data")
	}

	// Crear objeto de notificación
	webhookNotification := &WebhookNotification{
		ID:        int64(id),
		Type:      notificationType,
		Action:    action,
		Data:      data,
		Timestamp: time.Now(),
	}

	return webhookNotification, nil
}

// WebhookNotification representa una notificación de webhook
type WebhookNotification struct {
	ID        int64                  `json:"id"`
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// =============================================================================
// TESTS
// =============================================================================

// TestMercadoPagoWebhookService_ValidateWebhookSignature tests the ValidateWebhookSignature method
func TestMercadoPagoWebhookService_ValidateWebhookSignature(t *testing.T) {
	secretKey := "test_secret_key_123"
	service := NewMercadoPagoWebhookService(secretKey)

	tests := []struct {
		name          string
		xSignature    string
		xRequestId    string
		dataID        string
		body          string
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "Valid signature",
			xSignature:    "ts=1704908010,v1=618c85345248dd820d5fd456117c2ab2ef8eda45a0282ff693eac24131a5e839",
			xRequestId:    "test-request-id",
			dataID:        "123456789",
			body:          `{"test": "data"}`,
			expectedValid: false, // Will be false because we're using a test secret
			expectedError: false,
		},
		{
			name:          "Missing x-signature",
			xSignature:    "",
			xRequestId:    "test-request-id",
			dataID:        "123456789",
			body:          `{"test": "data"}`,
			expectedValid: false,
			expectedError: true,
		},
		{
			name:          "Invalid x-signature format",
			xSignature:    "invalid-format",
			xRequestId:    "test-request-id",
			dataID:        "123456789",
			body:          `{"test": "data"}`,
			expectedValid: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear request de prueba
			req := httptest.NewRequest("POST", "/webhook", strings.NewReader(tt.body))
			req.Header.Set("x-signature", tt.xSignature)
			req.Header.Set("x-request-id", tt.xRequestId)

			// Agregar query params
			q := req.URL.Query()
			q.Add("data.id", tt.dataID)
			req.URL.RawQuery = q.Encode()

			// Validar firma
			valid, err := service.ValidateWebhookSignature(req, []byte(tt.body))

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if valid != tt.expectedValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectedValid, valid)
			}
		})
	}
}

// TestMercadoPagoWebhookService_ParseXSignature tests the parseXSignature method
func TestMercadoPagoWebhookService_ParseXSignature(t *testing.T) {
	service := &MercadoPagoWebhookService{}

	tests := []struct {
		name         string
		xSignature   string
		expectedTs   string
		expectedHash string
		expectError  bool
	}{
		{
			name:         "Valid signature",
			xSignature:   "ts=1704908010,v1=618c85345248dd820d5fd456117c2ab2ef8eda45a0282ff693eac24131a5e839",
			expectedTs:   "1704908010",
			expectedHash: "618c85345248dd820d5fd456117c2ab2ef8eda45a0282ff693eac24131a5e839",
			expectError:  false,
		},
		{
			name:         "Invalid format",
			xSignature:   "invalid-format",
			expectedTs:   "",
			expectedHash: "",
			expectError:  true,
		},
		{
			name:         "Missing ts",
			xSignature:   "v1=618c85345248dd820d5fd456117c2ab2ef8eda45a0282ff693eac24131a5e839",
			expectedTs:   "",
			expectedHash: "",
			expectError:  true,
		},
		{
			name:         "Missing v1",
			xSignature:   "ts=1704908010",
			expectedTs:   "",
			expectedHash: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, hash, err := service.parseXSignature(tt.xSignature)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if ts != tt.expectedTs {
				t.Errorf("Expected ts=%s, got %s", tt.expectedTs, ts)
			}
			if hash != tt.expectedHash {
				t.Errorf("Expected hash=%s, got %s", tt.expectedHash, hash)
			}
		})
	}
}

// TestMercadoPagoWebhookService_GenerateManifest tests the generateManifest method
func TestMercadoPagoWebhookService_GenerateManifest(t *testing.T) {
	service := &MercadoPagoWebhookService{}

	dataID := "123456789"
	xRequestId := "test-request-id"
	ts := "1704908010"

	expected := "id:123456789;request-id:test-request-id;ts:1704908010;"
	result := service.generateManifest(dataID, xRequestId, ts)

	if result != expected {
		t.Errorf("Expected manifest=%s, got %s", expected, result)
	}
}

// TestMercadoPagoWebhookService_ProcessWebhookNotification tests the ProcessWebhookNotification method
func TestMercadoPagoWebhookService_ProcessWebhookNotification(t *testing.T) {
	service := &MercadoPagoWebhookService{}

	tests := []struct {
		name           string
		notification   map[string]interface{}
		expectError    bool
		expectedType   string
		expectedAction string
	}{
		{
			name: "Valid payment notification",
			notification: map[string]interface{}{
				"id":     float64(12345),
				"type":   "payment",
				"action": "payment.created",
				"data": map[string]interface{}{
					"id": "123456789",
				},
			},
			expectError:    false,
			expectedType:   "payment",
			expectedAction: "payment.created",
		},
		{
			name: "Invalid notification - missing id",
			notification: map[string]interface{}{
				"type":   "payment",
				"action": "payment.created",
				"data": map[string]interface{}{
					"id": "123456789",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid notification - missing type",
			notification: map[string]interface{}{
				"id":     float64(12345),
				"action": "payment.created",
				"data": map[string]interface{}{
					"id": "123456789",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ProcessWebhookNotification(tt.notification)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if result.Type != tt.expectedType {
					t.Errorf("Expected type=%s, got %s", tt.expectedType, result.Type)
				}
				if result.Action != tt.expectedAction {
					t.Errorf("Expected action=%s, got %s", tt.expectedAction, result.Action)
				}
			}
		})
	}
}
