package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/internal/repository"
	"it-integration-service/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarSetupService maneja la configuración OAuth2 para Google Calendar
type GoogleCalendarSetupService struct {
	config     *config.GoogleCalendarConfig
	repo       repository.GoogleCalendarRepository
	logger     logger.Logger
	encryption *EncryptionService
}

// OAuth2State representa el estado del flujo OAuth2
type OAuth2State struct {
	TenantID     string              `json:"tenant_id"`
	ChannelID    string              `json:"channel_id"`
	CalendarType domain.CalendarType `json:"calendar_type"`
	StateToken   string              `json:"state_token"`
	ExpiresAt    time.Time           `json:"expires_at"`
}

// AuthURLResponse representa la respuesta con URL de autenticación
type AuthURLResponse struct {
	AuthURL    string `json:"auth_url"`
	StateToken string `json:"state_token"`
	ExpiresAt  string `json:"expires_at"`
}

// IntegrationStatusResponse representa el estado de la integración
type IntegrationStatusResponse struct {
	ChannelID       string                   `json:"channel_id"`
	CalendarType    domain.CalendarType      `json:"calendar_type"`
	CalendarID      string                   `json:"calendar_id"`
	CalendarName    string                   `json:"calendar_name"`
	Status          domain.IntegrationStatus `json:"status"`
	IsAuthenticated bool                     `json:"is_authenticated"`
	TokenExpiry     *time.Time               `json:"token_expiry,omitempty"`
	LastSync        *time.Time               `json:"last_sync,omitempty"`
}

// NewGoogleCalendarSetupService crea una nueva instancia del servicio
func NewGoogleCalendarSetupService(cfg *config.GoogleCalendarConfig, repo repository.GoogleCalendarRepository, logger logger.Logger, encryption *EncryptionService) *GoogleCalendarSetupService {
	return &GoogleCalendarSetupService{
		config:     cfg,
		repo:       repo,
		logger:     logger,
		encryption: encryption,
	}
}

// InitiateAuth inicia el flujo de autenticación OAuth2
func (s *GoogleCalendarSetupService) InitiateAuth(ctx context.Context, tenantID string, calendarType domain.CalendarType) (*AuthURLResponse, error) {
	s.logger.Info("Iniciando autenticación OAuth2 para Google Calendar", map[string]interface{}{
		"tenant_id":     tenantID,
		"calendar_type": calendarType,
	})

	// Generar state token único
	stateToken := uuid.New().String()

	// Crear o actualizar integración
	channelID := uuid.New().String()
	integration := &domain.GoogleCalendarIntegration{
		ID:           channelID,
		TenantID:     tenantID,
		ChannelID:    channelID,
		CalendarType: calendarType,
		Status:       domain.StatusDisabled,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Guardar integración en base de datos
	err := s.repo.CreateIntegration(ctx, integration)
	if err != nil {
		s.logger.Error("Error al crear integración de Google Calendar", err, map[string]interface{}{
			"tenant_id":     tenantID,
			"calendar_type": calendarType,
		})
		return nil, fmt.Errorf("error al crear integración: %w", err)
	}

	// Configurar OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  s.config.RedirectURL,
		Scopes:       s.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.AuthURL,
			TokenURL: s.config.TokenURL,
		},
	}

	// Generar URL de autenticación
	authURL := oauth2Config.AuthCodeURL(stateToken, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	// Guardar state token temporalmente (en producción usar Redis)
	state := &OAuth2State{
		TenantID:     tenantID,
		ChannelID:    channelID,
		CalendarType: calendarType,
		StateToken:   stateToken,
		ExpiresAt:    time.Now().Add(10 * time.Minute), // 10 minutos de expiración
	}

	// En producción, guardar en Redis o base de datos temporal
	s.logger.Info("State token generado", map[string]interface{}{
		"state_token": stateToken,
		"expires_at":  state.ExpiresAt,
	})

	return &AuthURLResponse{
		AuthURL:    authURL,
		StateToken: stateToken,
		ExpiresAt:  state.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// HandleCallback maneja el callback de OAuth2
func (s *GoogleCalendarSetupService) HandleCallback(ctx context.Context, code, stateToken string) error {
	s.logger.Info("Procesando callback OAuth2", map[string]interface{}{
		"state_token": stateToken,
	})

	// En producción, recuperar state de Redis o base de datos
	// Por ahora, asumimos que el state es válido
	// TODO: Implementar validación de state token

	// Configurar OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  s.config.RedirectURL,
		Scopes:       s.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.AuthURL,
			TokenURL: s.config.TokenURL,
		},
	}

	// Intercambiar código por token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("Error al intercambiar código por token", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return fmt.Errorf("error al intercambiar código por token: %w", err)
	}

	// Crear cliente HTTP con token
	client := oauth2Config.Client(ctx, token)

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		s.logger.Error("Error al crear servicio de Google Calendar", err, nil)
		return fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Obtener información del calendario principal
	calendarList, err := calendarService.CalendarList.Get("primary").Do()
	if err != nil {
		s.logger.Error("Error al obtener información del calendario", err, nil)
		return fmt.Errorf("error al obtener información del calendario: %w", err)
	}

	// Encriptar tokens
	encryptedAccessToken, err := s.encryption.Encrypt(token.AccessToken)
	if err != nil {
		s.logger.Error("Error al encriptar access token", err, nil)
		return fmt.Errorf("error al encriptar access token: %w", err)
	}

	encryptedRefreshToken := ""
	if token.RefreshToken != "" {
		encryptedRefreshToken, err = s.encryption.Encrypt(token.RefreshToken)
		if err != nil {
			s.logger.Error("Error al encriptar refresh token", err, nil)
			return fmt.Errorf("error al encriptar refresh token: %w", err)
		}
	}

	// Actualizar integración con tokens
	integration := &domain.GoogleCalendarIntegration{
		ChannelID:    stateToken, // Usar stateToken como ChannelID temporal
		CalendarID:   "primary",
		CalendarName: calendarList.Summary,
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
		TokenExpiry:  token.Expiry,
		Status:       domain.StatusActive,
		UpdatedAt:    time.Now(),
	}

	// Guardar integración actualizada
	err = s.repo.UpdateIntegration(ctx, integration)
	if err != nil {
		s.logger.Error("Error al actualizar integración con tokens", err, map[string]interface{}{
			"channel_id": integration.ChannelID,
		})
		return fmt.Errorf("error al actualizar integración: %w", err)
	}

	s.logger.Info("Autenticación OAuth2 completada exitosamente", map[string]interface{}{
		"channel_id":    integration.ChannelID,
		"calendar_name": integration.CalendarName,
		"token_expiry":  integration.TokenExpiry,
	})

	return nil
}

// RefreshToken refresca el token de acceso automáticamente
func (s *GoogleCalendarSetupService) RefreshToken(ctx context.Context, channelID string) error {
	s.logger.Info("Refrescando token de acceso", map[string]interface{}{
		"channel_id": channelID,
	})

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		s.logger.Error("Error al obtener integración para refresh", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al obtener integración: %w", err)
	}

	// Desencriptar refresh token
	refreshToken, err := s.encryption.Decrypt(integration.RefreshToken)
	if err != nil {
		s.logger.Error("Error al desencriptar refresh token", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al desencriptar refresh token: %w", err)
	}

	// Configurar OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  s.config.RedirectURL,
		Scopes:       s.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.AuthURL,
			TokenURL: s.config.TokenURL,
		},
	}

	// Crear token para refresh
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Refrescar token
	newToken, err := oauth2Config.TokenSource(ctx, token).Token()
	if err != nil {
		s.logger.Error("Error al refrescar token", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al refrescar token: %w", err)
	}

	// Encriptar nuevo access token
	encryptedAccessToken, err := s.encryption.Encrypt(newToken.AccessToken)
	if err != nil {
		s.logger.Error("Error al encriptar nuevo access token", err, nil)
		return fmt.Errorf("error al encriptar nuevo access token: %w", err)
	}

	// Actualizar integración con nuevo token
	integration.AccessToken = encryptedAccessToken
	integration.TokenExpiry = newToken.Expiry
	integration.UpdatedAt = time.Now()

	err = s.repo.UpdateIntegration(ctx, integration)
	if err != nil {
		s.logger.Error("Error al actualizar integración con nuevo token", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al actualizar integración: %w", err)
	}

	s.logger.Info("Token refrescado exitosamente", map[string]interface{}{
		"channel_id": channelID,
		"new_expiry": newToken.Expiry,
		"expires_in": newToken.Expiry.Sub(time.Now()),
	})

	return nil
}

// GetIntegrationStatus obtiene el estado de la integración
func (s *GoogleCalendarSetupService) GetIntegrationStatus(ctx context.Context, channelID string) (*IntegrationStatusResponse, error) {
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		s.logger.Error("Error al obtener estado de integración", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return nil, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Verificar si el token está próximo a expirar
	isAuthenticated := integration.Status == domain.StatusActive
	tokenExpiry := &integration.TokenExpiry

	if integration.TokenExpiry.Before(time.Now().Add(5 * time.Minute)) {
		// Token expira en menos de 5 minutos, intentar refresh
		err := s.RefreshToken(ctx, channelID)
		if err != nil {
			s.logger.Warn("No se pudo refrescar token", map[string]interface{}{
				"channel_id": channelID,
				"error":      err.Error(),
			})
			isAuthenticated = false
			tokenExpiry = nil
		} else {
			// Obtener integración actualizada
			integration, _ = s.repo.GetIntegration(ctx, channelID)
			tokenExpiry = &integration.TokenExpiry
		}
	}

	return &IntegrationStatusResponse{
		ChannelID:       integration.ChannelID,
		CalendarType:    integration.CalendarType,
		CalendarID:      integration.CalendarID,
		CalendarName:    integration.CalendarName,
		Status:          integration.Status,
		IsAuthenticated: isAuthenticated,
		TokenExpiry:     tokenExpiry,
		LastSync:        &integration.UpdatedAt,
	}, nil
}

// SetupWebhook configura webhooks para sincronización automática
func (s *GoogleCalendarSetupService) SetupWebhook(ctx context.Context, channelID string) error {
	s.logger.Info("Configurando webhook para Google Calendar", map[string]interface{}{
		"channel_id": channelID,
	})

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		return fmt.Errorf("error al obtener integración: %w", err)
	}

	// Crear cliente OAuth2
	client, err := s.createOAuth2Client(ctx, integration)
	if err != nil {
		return fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	// Crear servicio de Google Calendar
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Configurar webhook
	webhook := &calendar.Channel{
		Id:         uuid.New().String(),
		Type:       "web_hook",
		Address:    s.config.WebhookURL,
		Token:      s.config.WebhookSecret,
		Expiration: time.Now().Add(24*time.Hour).UnixNano() / 1e6, // 24 horas en milisegundos
	}

	// Registrar webhook
	_, err = calendarService.Events.Watch("primary", webhook).Do()
	if err != nil {
		s.logger.Error("Error al configurar webhook", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al configurar webhook: %w", err)
	}

	// Actualizar integración con información del webhook
	integration.WebhookChannel = webhook.Id
	integration.WebhookResource = fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", integration.CalendarID)
	integration.UpdatedAt = time.Now()

	err = s.repo.UpdateIntegration(ctx, integration)
	if err != nil {
		s.logger.Error("Error al actualizar integración con webhook", err, map[string]interface{}{
			"channel_id": channelID,
		})
		return fmt.Errorf("error al actualizar integración: %w", err)
	}

	s.logger.Info("Webhook configurado exitosamente", map[string]interface{}{
		"channel_id":      channelID,
		"webhook_id":      webhook.Id,
		"webhook_address": webhook.Address,
		"expiration":      webhook.Expiration,
	})

	return nil
}

// createOAuth2Client crea un cliente OAuth2 con refresh automático
func (s *GoogleCalendarSetupService) createOAuth2Client(ctx context.Context, integration *domain.GoogleCalendarIntegration) (*http.Client, error) {
	// Desencriptar access token
	accessToken, err := s.encryption.Decrypt(integration.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error al desencriptar access token: %w", err)
	}

	// Configurar OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  s.config.RedirectURL,
		Scopes:       s.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.AuthURL,
			TokenURL: s.config.TokenURL,
		},
	}

	// Crear token
	token := &oauth2.Token{
		AccessToken: accessToken,
		Expiry:      integration.TokenExpiry,
	}

	// Si hay refresh token, agregarlo
	if integration.RefreshToken != "" {
		refreshToken, err := s.encryption.Decrypt(integration.RefreshToken)
		if err == nil {
			token.RefreshToken = refreshToken
		}
	}

	// Crear token source con refresh automático
	tokenSource := oauth2Config.TokenSource(ctx, token)

	// Crear cliente HTTP
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
			Base:   http.DefaultTransport,
		},
	}

	return client, nil
}

// ValidateToken valida si el token actual es válido
func (s *GoogleCalendarSetupService) ValidateToken(ctx context.Context, channelID string) (bool, error) {
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		return false, fmt.Errorf("error al obtener integración: %w", err)
	}

	// Verificar si el token ha expirado
	if integration.TokenExpiry.Before(time.Now()) {
		return false, nil
	}

	// Intentar hacer una llamada de prueba
	client, err := s.createOAuth2Client(ctx, integration)
	if err != nil {
		return false, fmt.Errorf("error al crear cliente OAuth2: %w", err)
	}

	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return false, fmt.Errorf("error al crear servicio de Google Calendar: %w", err)
	}

	// Hacer llamada de prueba
	_, err = calendarService.CalendarList.Get("primary").Do()
	if err != nil {
		return false, nil
	}

	return true, nil
}

// RevokeAccess revoca el acceso a Google Calendar
func (s *GoogleCalendarSetupService) RevokeAccess(ctx context.Context, channelID string) error {
	s.logger.Info("Revocando acceso a Google Calendar", map[string]interface{}{
		"channel_id": channelID,
	})

	// Obtener integración
	integration, err := s.repo.GetIntegration(ctx, channelID)
	if err != nil {
		return fmt.Errorf("error al obtener integración: %w", err)
	}

	// Desencriptar access token
	accessToken, err := s.encryption.Decrypt(integration.AccessToken)
	if err != nil {
		return fmt.Errorf("error al desencriptar access token: %w", err)
	}

	// Revocar token en Google
	revokeURL := "https://oauth2.googleapis.com/revoke"
	req, err := http.NewRequest("POST", revokeURL, nil)
	if err != nil {
		return fmt.Errorf("error al crear request de revocación: %w", err)
	}

	q := req.URL.Query()
	q.Add("token", accessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error al revocar token: %w", err)
	}
	defer resp.Body.Close()

	// Actualizar estado de integración
	integration.Status = domain.StatusDisabled
	integration.AccessToken = ""
	integration.RefreshToken = ""
	integration.UpdatedAt = time.Now()

	err = s.repo.UpdateIntegration(ctx, integration)
	if err != nil {
		return fmt.Errorf("error al actualizar integración: %w", err)
	}

	s.logger.Info("Acceso revocado exitosamente", map[string]interface{}{
		"channel_id": channelID,
	})

	return nil
}
