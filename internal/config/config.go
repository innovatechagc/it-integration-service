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
			Name:     getEnv("DB_NAME", "microservice"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		ExternalAPI: ExternalAPIConfig{
			BaseURL: getEnv("EXTERNAL_API_URL", "https://api.example.com"),
			APIKey:  getEnv("EXTERNAL_API_KEY", ""),
			Timeout: getEnvAsInt("EXTERNAL_API_TIMEOUT", 30),
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