package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	LogLevel    string
	VaultConfig VaultConfig
	Database    DatabaseConfig
	ExternalAPI ExternalAPIConfig
	Integration IntegrationConfig
	MercadoPago MercadoPagoConfig
	TawkTo      TawkToConfig
	Mailchimp   MailchimpConfig
}

type VaultConfig struct {
	Address string
	Token   string
	Path    string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ExternalAPIConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
}

type IntegrationConfig struct {
	MessagingServiceURL string
	EncryptionKey       string
	RateLimitRPS        int
	RateLimitBurst      int
	WebhookSecrets      map[string]string
	WebhookVerifyTokens map[string]string
}

type TawkToConfig struct {
	APIKey        string `envconfig:"TAWKTO_API_KEY" required:"true"`
	BaseURL       string `envconfig:"TAWKTO_BASE_URL" default:"https://api.tawk.to"`
	WebhookSecret string `envconfig:"TAWKTO_WEBHOOK_SECRET"`
	WidgetID      string `envconfig:"TAWKTO_WIDGET_ID"`
	PropertyID    string `envconfig:"TAWKTO_PROPERTY_ID"`
}

type MailchimpConfig struct {
	APIKey        string `envconfig:"MAILCHIMP_API_KEY" required:"true"`
	ServerPrefix  string `envconfig:"MAILCHIMP_SERVER_PREFIX" required:"true"`
	BaseURL       string `envconfig:"MAILCHIMP_BASE_URL" default:"https://us1.api.mailchimp.com"`
	WebhookSecret string `envconfig:"MAILCHIMP_WEBHOOK_SECRET"`
	AudienceID    string `envconfig:"MAILCHIMP_AUDIENCE_ID"`
	DataCenter    string `envconfig:"MAILCHIMP_DATA_CENTER"`
}

func Load() *Config {
	// Cargar variables de entorno desde .env si existe
	_ = godotenv.Load()

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		VaultConfig: VaultConfig{
			Address: getEnv("VAULT_ADDR", "http://localhost:8200"),
			Token:   getEnv("VAULT_TOKEN", ""),
			Path:    getEnv("VAULT_PATH", "secret/microservice"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "it_db_chatbot"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		ExternalAPI: ExternalAPIConfig{
			BaseURL: getEnv("EXTERNAL_API_URL", "https://api.example.com"),
			APIKey:  getEnv("EXTERNAL_API_KEY", ""),
			Timeout: getEnvAsInt("EXTERNAL_API_TIMEOUT", 30),
		},
		Integration: IntegrationConfig{
			MessagingServiceURL: getEnv("MESSAGING_SERVICE_URL", "http://localhost:8081"),
			EncryptionKey:       getEnv("ENCRYPTION_KEY", "default-key-change-in-production"),
			RateLimitRPS:        getEnvAsInt("RATE_LIMIT_RPS", 100),
			RateLimitBurst:      getEnvAsInt("RATE_LIMIT_BURST", 200),
			WebhookSecrets: map[string]string{
				"whatsapp":  getEnv("WHATSAPP_WEBHOOK_SECRET", ""),
				"messenger": getEnv("MESSENGER_WEBHOOK_SECRET", ""),
				"instagram": getEnv("INSTAGRAM_WEBHOOK_SECRET", ""),
				"telegram":  getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
				"webchat":   getEnv("WEBCHAT_WEBHOOK_SECRET", ""),
				"tawkto":    getEnv("TAWKTO_WEBHOOK_SECRET", ""),
				"mailchimp": getEnv("MAILCHIMP_WEBHOOK_SECRET", ""),
			},
			WebhookVerifyTokens: map[string]string{
				"whatsapp":  getEnv("WHATSAPP_VERIFY_TOKEN", ""),
				"messenger": getEnv("MESSENGER_VERIFY_TOKEN", ""),
				"instagram": getEnv("INSTAGRAM_VERIFY_TOKEN", ""),
				"telegram":  getEnv("TELEGRAM_VERIFY_TOKEN", ""),
				"webchat":   getEnv("WEBCHAT_VERIFY_TOKEN", ""),
				"tawkto":    getEnv("TAWKTO_VERIFY_TOKEN", ""),
				"mailchimp": getEnv("MAILCHIMP_VERIFY_TOKEN", ""),
			},
		},
		MercadoPago: MercadoPagoConfig{
			AccessToken:  getEnv("MP_ACCESS_TOKEN", ""),
			ClientID:     getEnv("MP_CLIENT_ID", ""),
			ClientSecret: getEnv("MP_CLIENT_SECRET", ""),
			Environment:  getEnv("MP_ENVIRONMENT", "sandbox"),
			WebhookURL:   getEnv("MP_WEBHOOK_URL", ""),
			SecretKey:    getEnv("MP_WEBHOOK_SECRET", ""),
		},
		TawkTo: TawkToConfig{
			APIKey:        getEnv("TAWKTO_API_KEY", ""),
			BaseURL:       getEnv("TAWKTO_BASE_URL", "https://api.tawk.to"),
			WebhookSecret: getEnv("TAWKTO_WEBHOOK_SECRET", ""),
			WidgetID:      getEnv("TAWKTO_WIDGET_ID", ""),
			PropertyID:    getEnv("TAWKTO_PROPERTY_ID", ""),
		},
		Mailchimp: MailchimpConfig{
			APIKey:        getEnv("MAILCHIMP_API_KEY", ""),
			ServerPrefix:  getEnv("MAILCHIMP_SERVER_PREFIX", ""),
			BaseURL:       getEnv("MAILCHIMP_BASE_URL", "https://us1.api.mailchimp.com"),
			WebhookSecret: getEnv("MAILCHIMP_WEBHOOK_SECRET", ""),
			AudienceID:    getEnv("MAILCHIMP_AUDIENCE_ID", ""),
			DataCenter:    getEnv("MAILCHIMP_DATA_CENTER", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
