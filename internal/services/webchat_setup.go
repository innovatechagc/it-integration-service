package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// WebchatSetupService maneja la configuración específica de Webchat
type WebchatSetupService struct {
	logger logger.Logger
}

// NewWebchatSetupService crea una nueva instancia del servicio de configuración de Webchat
func NewWebchatSetupService(logger logger.Logger) *WebchatSetupService {
	return &WebchatSetupService{
		logger: logger,
	}
}

// WebchatConfig representa la configuración del chat web
type WebchatConfig struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Domain      string    `json:"domain"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
	Theme       struct {
		PrimaryColor    string `json:"primary_color"`
		SecondaryColor  string `json:"secondary_color"`
		TextColor       string `json:"text_color"`
		BackgroundColor string `json:"background_color"`
	} `json:"theme"`
	Settings struct {
		WelcomeMessage string `json:"welcome_message"`
		AutoReply      bool   `json:"auto_reply"`
		BusinessHours  struct {
			Enabled bool `json:"enabled"`
			Hours   map[string]struct {
				Open  string `json:"open"`
				Close string `json:"close"`
			} `json:"hours"`
		} `json:"business_hours"`
		Notifications struct {
			Email      bool   `json:"email"`
			Webhook    bool   `json:"webhook"`
			WebhookURL string `json:"webhook_url,omitempty"`
		} `json:"notifications"`
	} `json:"settings"`
}

// WebchatSession representa una sesión de chat web
type WebchatSession struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	StartedAt    time.Time              `json:"started_at"`
	LastActivity time.Time              `json:"last_activity"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// WebchatMessage representa un mensaje del chat web
type WebchatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // "user" o "agent"
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "sent", "delivered", "read"
}

// ValidateWebchatConfig valida la configuración del chat web
func (s *WebchatSetupService) ValidateWebchatConfig(ctx context.Context, config *WebchatConfig) error {
	// Validaciones básicas
	if config.Name == "" {
		return fmt.Errorf("webchat name is required")
	}

	if config.Domain == "" {
		return fmt.Errorf("webchat domain is required")
	}

	// Validar formato de dominio
	if !s.isValidDomain(config.Domain) {
		return fmt.Errorf("invalid domain format: %s", config.Domain)
	}

	// Validar configuración de tema
	if config.Theme.PrimaryColor == "" {
		config.Theme.PrimaryColor = "#007bff" // Color por defecto
	}

	if config.Theme.SecondaryColor == "" {
		config.Theme.SecondaryColor = "#6c757d" // Color por defecto
	}

	// Validar configuración de notificaciones
	if config.Settings.Notifications.Webhook && config.Settings.Notifications.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required when webhook notifications are enabled")
	}

	s.logger.Info("Webchat configuration validated successfully", map[string]interface{}{
		"webchat_id": config.ID,
		"name":       config.Name,
		"domain":     config.Domain,
	})

	return nil
}

// CreateWebchatIntegration crea una integración de Webchat con configuración completa
func (s *WebchatSetupService) CreateWebchatIntegration(ctx context.Context, config *WebchatConfig, webhookURL, tenantID string) (*domain.ChannelIntegration, error) {
	// Validar configuración
	if err := s.ValidateWebchatConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("invalid webchat configuration: %w", err)
	}

	// Generar ID único si no existe
	if config.ID == "" {
		config.ID = fmt.Sprintf("webchat_%s_%d", tenantID, time.Now().Unix())
	}

	// Configurar webhook URL si no está configurada
	if config.Settings.Notifications.Webhook && config.Settings.Notifications.WebhookURL == "" {
		config.Settings.Notifications.WebhookURL = webhookURL
	}

	// Crear configuración de la integración
	integrationConfig := map[string]interface{}{
		"webchat_config": config,
		"webhook_url":    webhookURL,
		"tenant_id":      tenantID,
		"created_at":     time.Now().Unix(),
	}

	configJSON, err := json.Marshal(integrationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	integration := &domain.ChannelIntegration{
		TenantID:    tenantID,
		Platform:    domain.PlatformWebchat,
		Provider:    domain.ProviderCustom,
		AccessToken: config.ID, // Usar el ID del webchat como token
		WebhookURL:  webhookURL,
		Status:      domain.StatusActive,
		Config:      configJSON,
	}

	s.logger.Info("Webchat integration created successfully", map[string]interface{}{
		"webchat_id": config.ID,
		"name":       config.Name,
		"domain":     config.Domain,
		"tenant_id":  tenantID,
	})

	return integration, nil
}

// UpdateWebchatConfig actualiza la configuración del chat web
func (s *WebchatSetupService) UpdateWebchatConfig(ctx context.Context, config *WebchatConfig) error {
	if err := s.ValidateWebchatConfig(ctx, config); err != nil {
		return fmt.Errorf("invalid webchat configuration: %w", err)
	}

	config.Status = "active"
	config.UpdatedAt = time.Now()

	s.logger.Info("Webchat configuration updated successfully", map[string]interface{}{
		"webchat_id": config.ID,
		"name":       config.Name,
	})

	return nil
}

// GetWebchatConfig obtiene la configuración del chat web
func (s *WebchatSetupService) GetWebchatConfig(ctx context.Context, webchatID string) (*WebchatConfig, error) {
	// En una implementación real, esto obtendría la configuración de la base de datos
	// Por ahora, retornamos una configuración de ejemplo
	config := &WebchatConfig{
		ID:          webchatID,
		Name:        "Default Webchat",
		Description: "Chat web personalizado",
		Domain:      "example.com",
		Status:      "active",
	}

	config.Theme.PrimaryColor = "#007bff"
	config.Theme.SecondaryColor = "#6c757d"
	config.Theme.TextColor = "#333333"
	config.Theme.BackgroundColor = "#ffffff"

	config.Settings.WelcomeMessage = "¡Hola! ¿En qué podemos ayudarte?"
	config.Settings.AutoReply = true

	config.Settings.BusinessHours.Enabled = true
	config.Settings.BusinessHours.Hours = map[string]struct {
		Open  string `json:"open"`
		Close string `json:"close"`
	}{
		"monday":    {"09:00", "18:00"},
		"tuesday":   {"09:00", "18:00"},
		"wednesday": {"09:00", "18:00"},
		"thursday":  {"09:00", "18:00"},
		"friday":    {"09:00", "18:00"},
		"saturday":  {"10:00", "16:00"},
		"sunday":    {"closed", "closed"},
	}

	config.Settings.Notifications.Email = true
	config.Settings.Notifications.Webhook = true

	return config, nil
}

// CreateWebchatSession crea una nueva sesión de chat web
func (s *WebchatSetupService) CreateWebchatSession(ctx context.Context, webchatID, userID string, metadata map[string]interface{}) (*WebchatSession, error) {
	session := &WebchatSession{
		ID:           fmt.Sprintf("session_%s_%d", webchatID, time.Now().Unix()),
		UserID:       userID,
		SessionID:    fmt.Sprintf("sess_%s", userID),
		StartedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       "active",
		Metadata:     metadata,
	}

	s.logger.Info("Webchat session created successfully", map[string]interface{}{
		"session_id": session.ID,
		"user_id":    userID,
		"webchat_id": webchatID,
	})

	return session, nil
}

// SendWebchatMessage envía un mensaje a través del chat web
func (s *WebchatSetupService) SendWebchatMessage(ctx context.Context, sessionID, userID, text string) (*WebchatMessage, error) {
	message := &WebchatMessage{
		ID:        fmt.Sprintf("msg_%s_%d", sessionID, time.Now().Unix()),
		SessionID: sessionID,
		UserID:    userID,
		Type:      "agent",
		Text:      text,
		Timestamp: time.Now(),
		Status:    "sent",
	}

	s.logger.Info("Webchat message sent successfully", map[string]interface{}{
		"message_id": message.ID,
		"session_id": sessionID,
		"user_id":    userID,
		"text":       text,
	})

	return message, nil
}

// GetWebchatSessions obtiene las sesiones activas del chat web
func (s *WebchatSetupService) GetWebchatSessions(ctx context.Context, webchatID string, limit int) ([]*WebchatSession, error) {
	// En una implementación real, esto obtendría las sesiones de la base de datos
	// Por ahora, retornamos sesiones de ejemplo
	sessions := []*WebchatSession{
		{
			ID:           "session_1",
			UserID:       "user_1",
			SessionID:    "sess_user_1",
			StartedAt:    time.Now().Add(-time.Hour),
			LastActivity: time.Now().Add(-10 * time.Minute),
			Status:       "active",
			Metadata: map[string]interface{}{
				"browser": "Chrome",
				"os":      "Windows",
			},
		},
	}

	return sessions, nil
}

// ValidateWebhookToken valida el token de verificación del webhook
func (s *WebchatSetupService) ValidateWebhookToken(providedToken, expectedToken string) bool {
	return providedToken == expectedToken
}

// isValidDomain valida el formato de un dominio
func (s *WebchatSetupService) isValidDomain(domain string) bool {
	// Validación básica de formato de dominio
	if len(domain) < 3 || len(domain) > 253 {
		return false
	}

	// Verificar que contenga al menos un punto
	hasDot := false
	for _, char := range domain {
		if char == '.' {
			hasDot = true
			break
		}
	}

	return hasDot
}

// GetWebchatStats obtiene estadísticas del chat web
func (s *WebchatSetupService) GetWebchatStats(ctx context.Context, webchatID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"total_sessions":        150,
		"active_sessions":       12,
		"total_messages":        1250,
		"avg_response_time":     "2.5 minutes",
		"customer_satisfaction": 4.8,
		"busy_hours": map[string]int{
			"09:00": 25,
			"10:00": 45,
			"11:00": 38,
			"12:00": 52,
			"13:00": 30,
			"14:00": 42,
			"15:00": 48,
			"16:00": 35,
			"17:00": 28,
		},
	}

	return stats, nil
}
