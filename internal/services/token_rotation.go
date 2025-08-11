package services

import (
	"context"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// TokenRotationService maneja la rotación automática de tokens
type TokenRotationService struct {
	channelRepo domain.ChannelIntegrationRepository
	logger       logger.Logger
}

// NewTokenRotationService crea una nueva instancia del servicio de rotación de tokens
func NewTokenRotationService(channelRepo domain.ChannelIntegrationRepository, logger logger.Logger) *TokenRotationService {
	return &TokenRotationService{
		channelRepo: channelRepo,
		logger:       logger,
	}
}

// TokenRotationConfig representa la configuración de rotación de tokens
type TokenRotationConfig struct {
	Enabled           bool          `json:"enabled"`
	RotationInterval  time.Duration `json:"rotation_interval"`
	WarningDays       int           `json:"warning_days"`
	AutoRotation      bool          `json:"auto_rotation"`
	NotificationEmail string        `json:"notification_email"`
}

// TokenStatus representa el estado de un token
type TokenStatus struct {
	ChannelID     string    `json:"channel_id"`
	Platform      string    `json:"platform"`
	TenantID      string    `json:"tenant_id"`
	TokenExpiry   time.Time `json:"token_expiry"`
	DaysUntilExpiry int     `json:"days_until_expiry"`
	Status        string    `json:"status"` // "valid", "expiring_soon", "expired"
	LastRotated   time.Time `json:"last_rotated"`
}

// RotateToken rota un token específico
func (s *TokenRotationService) RotateToken(ctx context.Context, channelID string, newToken string) error {
	// Obtener la integración actual
	integration, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel integration: %w", err)
	}

	// Validar que el nuevo token sea válido
	if err := s.validateNewToken(ctx, integration.Platform, newToken); err != nil {
		return fmt.Errorf("invalid new token: %w", err)
	}

	// Actualizar el token
	integration.AccessToken = newToken
	integration.UpdatedAt = time.Now()

	// Guardar en la base de datos
	if err := s.channelRepo.Update(ctx, integration); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	s.logger.Info("Token rotated successfully", map[string]interface{}{
		"channel_id": channelID,
		"platform":   integration.Platform,
		"tenant_id":  integration.TenantID,
	})

	return nil
}

// GetExpiringTokens obtiene tokens que están por expirar
func (s *TokenRotationService) GetExpiringTokens(ctx context.Context, daysThreshold int) ([]*TokenStatus, error) {
	// En una implementación real, esto consultaría la base de datos
	// Por ahora, retornamos datos de ejemplo
	expiringTokens := []*TokenStatus{
		{
			ChannelID:       "whatsapp_tenant1_123",
			Platform:        "whatsapp",
			TenantID:        "tenant1",
			TokenExpiry:     time.Now().AddDate(0, 0, 5), // 5 días
			DaysUntilExpiry: 5,
			Status:          "expiring_soon",
			LastRotated:     time.Now().AddDate(0, -1, 0), // 1 mes atrás
		},
		{
			ChannelID:       "telegram_tenant1_456",
			Platform:        "telegram",
			TenantID:        "tenant1",
			TokenExpiry:     time.Now().AddDate(0, 0, 2), // 2 días
			DaysUntilExpiry: 2,
			Status:          "expiring_soon",
			LastRotated:     time.Now().AddDate(0, -2, 0), // 2 meses atrás
		},
	}

	return expiringTokens, nil
}

// ScheduleTokenRotation programa la rotación automática de tokens
func (s *TokenRotationService) ScheduleTokenRotation(ctx context.Context, config TokenRotationConfig) error {
	if !config.Enabled {
		s.logger.Info("Token rotation is disabled")
		return nil
	}

	// Programar rotación automática
	go s.runTokenRotationScheduler(ctx, config)

	s.logger.Info("Token rotation scheduler started", map[string]interface{}{
		"rotation_interval": config.RotationInterval,
		"warning_days":      config.WarningDays,
		"auto_rotation":     config.AutoRotation,
	})

	return nil
}

// runTokenRotationScheduler ejecuta el scheduler de rotación de tokens
func (s *TokenRotationService) runTokenRotationScheduler(ctx context.Context, config TokenRotationConfig) {
	ticker := time.NewTicker(config.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Token rotation scheduler stopped")
			return
		case <-ticker.C:
			if err := s.processTokenRotation(ctx, config); err != nil {
				s.logger.Error("Failed to process token rotation", err)
			}
		}
	}
}

// processTokenRotation procesa la rotación de tokens
func (s *TokenRotationService) processTokenRotation(ctx context.Context, config TokenRotationConfig) error {
	// Obtener tokens que están por expirar
	expiringTokens, err := s.GetExpiringTokens(ctx, config.WarningDays)
	if err != nil {
		return fmt.Errorf("failed to get expiring tokens: %w", err)
	}

	for _, token := range expiringTokens {
		if token.Status == "expired" {
			// Token expirado - desactivar integración
			if err := s.deactivateExpiredIntegration(ctx, token.ChannelID); err != nil {
				s.logger.Error("Failed to deactivate expired integration", err)
			}
		} else if token.Status == "expiring_soon" {
			// Token por expirar - enviar notificación
			if err := s.sendExpiryNotification(ctx, token, config); err != nil {
				s.logger.Error("Failed to send expiry notification", err)
			}

			// Rotación automática si está habilitada
			if config.AutoRotation {
				if err := s.autoRotateToken(ctx, token.ChannelID); err != nil {
					s.logger.Error("Failed to auto-rotate token", err)
				}
			}
		}
	}

	return nil
}

// validateNewToken valida que un nuevo token sea válido
func (s *TokenRotationService) validateNewToken(ctx context.Context, platform domain.Platform, token string) error {
	switch platform {
	case domain.PlatformWhatsApp:
		return s.validateWhatsAppToken(ctx, token)
	case domain.PlatformTelegram:
		return s.validateTelegramToken(ctx, token)
	case domain.PlatformMessenger:
		return s.validateMessengerToken(ctx, token)
	case domain.PlatformInstagram:
		return s.validateInstagramToken(ctx, token)
	default:
		return fmt.Errorf("unsupported platform for token validation: %s", platform)
	}
}

// validateWhatsAppToken valida un token de WhatsApp
func (s *TokenRotationService) validateWhatsAppToken(ctx context.Context, token string) error {
	// Implementar validación específica de WhatsApp
	// Por ahora, solo verificar que no esté vacío
	if token == "" {
		return fmt.Errorf("whatsapp token cannot be empty")
	}
	return nil
}

// validateTelegramToken valida un token de Telegram
func (s *TokenRotationService) validateTelegramToken(ctx context.Context, token string) error {
	// Implementar validación específica de Telegram
	if token == "" {
		return fmt.Errorf("telegram token cannot be empty")
	}
	return nil
}

// validateMessengerToken valida un token de Messenger
func (s *TokenRotationService) validateMessengerToken(ctx context.Context, token string) error {
	// Implementar validación específica de Messenger
	if token == "" {
		return fmt.Errorf("messenger token cannot be empty")
	}
	return nil
}

// validateInstagramToken valida un token de Instagram
func (s *TokenRotationService) validateInstagramToken(ctx context.Context, token string) error {
	// Implementar validación específica de Instagram
	if token == "" {
		return fmt.Errorf("instagram token cannot be empty")
	}
	return nil
}

// deactivateExpiredIntegration desactiva una integración con token expirado
func (s *TokenRotationService) deactivateExpiredIntegration(ctx context.Context, channelID string) error {
	integration, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get integration: %w", err)
	}

	integration.Status = domain.StatusError
	integration.UpdatedAt = time.Now()

	if err := s.channelRepo.Update(ctx, integration); err != nil {
		return fmt.Errorf("failed to deactivate integration: %w", err)
	}

	s.logger.Warn("Integration deactivated due to expired token", map[string]interface{}{
		"channel_id": channelID,
		"platform":   integration.Platform,
		"tenant_id":  integration.TenantID,
	})

	return nil
}

// sendExpiryNotification envía notificación de expiración de token
func (s *TokenRotationService) sendExpiryNotification(ctx context.Context, token *TokenStatus, config TokenRotationConfig) error {
	// En una implementación real, esto enviaría un email o webhook
	s.logger.Warn("Token expiring soon", map[string]interface{}{
		"channel_id":       token.ChannelID,
		"platform":         token.Platform,
		"tenant_id":        token.TenantID,
		"days_until_expiry": token.DaysUntilExpiry,
		"notification_email": config.NotificationEmail,
	})

	return nil
}

// autoRotateToken rota automáticamente un token
func (s *TokenRotationService) autoRotateToken(ctx context.Context, channelID string) error {
	// En una implementación real, esto obtendría un nuevo token de la API correspondiente
	// Por ahora, solo loggeamos la acción
	s.logger.Info("Auto-rotating token", map[string]interface{}{
		"channel_id": channelID,
	})

	return nil
}

// GetTokenRotationConfig obtiene la configuración de rotación de tokens
func (s *TokenRotationService) GetTokenRotationConfig() TokenRotationConfig {
	return TokenRotationConfig{
		Enabled:           true,
		RotationInterval:  24 * time.Hour, // Revisar cada 24 horas
		WarningDays:       7,              // Advertir 7 días antes
		AutoRotation:      false,          // No rotar automáticamente por defecto
		NotificationEmail: "admin@company.com",
	}
}
