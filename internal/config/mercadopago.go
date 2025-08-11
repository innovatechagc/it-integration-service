package config

import (
	"os"
)

// MercadoPagoConfig contiene la configuración para Mercado Pago
type MercadoPagoConfig struct {
	AccessToken  string
	ClientID     string
	ClientSecret string
	Environment  string
	WebhookURL   string
	SecretKey    string // Clave secreta para validar webhooks
	SDK          interface{}
}

// NewMercadoPagoConfig crea una nueva instancia de configuración de Mercado Pago
func NewMercadoPagoConfig() (*MercadoPagoConfig, error) {
	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, ErrMissingAccessToken
	}

	clientID := os.Getenv("MP_CLIENT_ID")
	clientSecret := os.Getenv("MP_CLIENT_SECRET")
	environment := os.Getenv("MP_ENVIRONMENT")
	if environment == "" {
		environment = "sandbox" // Por defecto usa sandbox
	}

	webhookURL := os.Getenv("MP_WEBHOOK_URL")
	secretKey := os.Getenv("MP_WEBHOOK_SECRET") // Clave secreta para validar webhooks

	// Configurar el SDK de Mercado Pago (placeholder)
	var sdk interface{} = nil

	return &MercadoPagoConfig{
		AccessToken:  accessToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Environment:  environment,
		WebhookURL:   webhookURL,
		SecretKey:    secretKey,
		SDK:          sdk,
	}, nil
}

// IsProduction verifica si está en modo producción
func (c *MercadoPagoConfig) IsProduction() bool {
	return c.Environment == "production"
}

// GetBaseURL retorna la URL base según el ambiente
func (c *MercadoPagoConfig) GetBaseURL() string {
	if c.IsProduction() {
		return "https://api.mercadopago.com"
	}
	return "https://api.mercadopago.com/sandbox"
}
