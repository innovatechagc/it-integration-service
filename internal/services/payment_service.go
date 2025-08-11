package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/models"
)

// PaymentService maneja la lógica de pagos con Mercado Pago
type PaymentService struct {
	config *config.MercadoPagoConfig
	client *http.Client
}

// NewPaymentService crea una nueva instancia del servicio de pagos
func NewPaymentService(config *config.MercadoPagoConfig) *PaymentService {
	return &PaymentService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePayment crea un nuevo pago en Mercado Pago
func (s *PaymentService) CreatePayment(request *models.PaymentRequest) (*models.PaymentResponse, error) {
	// Validar el monto de la transacción
	if request.TransactionAmount <= 0 {
		return nil, fmt.Errorf("el monto de la transacción debe ser mayor a 0")
	}

	// Validar el número de cuotas
	if request.Installments < 1 {
		request.Installments = 1
	}

	// Preparar la URL de notificación si no está definida
	if request.NotificationURL == "" {
		request.NotificationURL = s.config.WebhookURL
	}

	// Crear el payload para Mercado Pago
	payload := map[string]interface{}{
		"transaction_amount": request.TransactionAmount,
		"token":              request.Token,
		"description":        request.Description,
		"installments":       request.Installments,
		"payment_method_id":  request.PaymentMethodID,
		"payer": map[string]interface{}{
			"email": request.Payer.Email,
			"name":  request.Payer.Name,
		},
		"external_reference": request.ExternalReference,
		"notification_url":   request.NotificationURL,
	}

	// Agregar información adicional si existe
	if len(request.AdditionalInfo.Items) > 0 {
		payload["additional_info"] = map[string]interface{}{
			"items": request.AdditionalInfo.Items,
		}
	}

	// Convertir a JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error al serializar la solicitud: %w", err)
	}

	// Crear la solicitud HTTP
	url := s.getAPIURL() + "/v1/payments"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %w", err)
	}

	// Configurar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	// Ejecutar la solicitud
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer la respuesta: %w", err)
	}

	// Verificar el código de estado
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("error en la respuesta de Mercado Pago (status: %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("error en la respuesta de Mercado Pago: %v", errorResp)
	}

	// Parsear la respuesta
	var paymentResponse models.PaymentResponse
	if err := json.Unmarshal(body, &paymentResponse); err != nil {
		return nil, fmt.Errorf("error al parsear la respuesta: %w", err)
	}

	return &paymentResponse, nil
}

// GetPayment obtiene información de un pago específico
func (s *PaymentService) GetPayment(paymentID int64) (*models.PaymentResponse, error) {
	url := fmt.Sprintf("%s/v1/payments/%d", s.getAPIURL(), paymentID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer la respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener el pago (status: %d): %s", resp.StatusCode, string(body))
	}

	var paymentResponse models.PaymentResponse
	if err := json.Unmarshal(body, &paymentResponse); err != nil {
		return nil, fmt.Errorf("error al parsear la respuesta: %w", err)
	}

	return &paymentResponse, nil
}

// RefundPayment procesa un reembolso
func (s *PaymentService) RefundPayment(paymentID int64, amount float64) error {
	payload := map[string]interface{}{
		"amount": amount,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error al serializar la solicitud: %w", err)
	}

	url := fmt.Sprintf("%s/v1/payments/%d/refunds", s.getAPIURL(), paymentID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error al crear la solicitud HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error al ejecutar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al procesar el reembolso (status: %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ValidateWebhookSignature valida la firma del webhook
func (s *PaymentService) ValidateWebhookSignature(signature, body string) bool {
	// En un entorno de producción, implementar la validación de firma
	// según la documentación de Mercado Pago
	// Por ahora, retornamos true para desarrollo
	return true
}

// getAPIURL retorna la URL base de la API según el entorno
func (s *PaymentService) getAPIURL() string {
	if s.config.Environment == "production" {
		return "https://api.mercadopago.com"
	}
	return "https://api.mercadopago.com/sandbox"
}
