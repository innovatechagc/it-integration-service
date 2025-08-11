package services

import (
	"context"
	"encoding/json"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"

	"github.com/google/uuid"
)

type integrationService struct {
	channelService ChannelService
	inboundRepo    domain.InboundMessageRepository
	webhookService WebhookService
	logger         logger.Logger
}

// NewIntegrationService crea una nueva instancia del servicio de integración
func NewIntegrationService(
	channelService ChannelService,
	inboundRepo domain.InboundMessageRepository,
	webhookService WebhookService,
	logger logger.Logger,
) IntegrationService {
	return &integrationService{
		channelService: channelService,
		inboundRepo:    inboundRepo,
		webhookService: webhookService,
		logger:         logger,
	}
}

// Gestión de canales
func (s *integrationService) CreateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
	return s.channelService.CreateChannel(ctx, integration)
}

func (s *integrationService) GetChannel(ctx context.Context, id string) (*domain.ChannelIntegration, error) {
	return s.channelService.GetChannel(ctx, id)
}

func (s *integrationService) GetChannelsByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
	return s.channelService.GetChannelsByTenant(ctx, tenantID)
}

func (s *integrationService) UpdateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
	return s.channelService.UpdateChannel(ctx, integration)
}

func (s *integrationService) DeleteChannel(ctx context.Context, id string) error {
	return s.channelService.DeleteChannel(ctx, id)
}

// Procesamiento de webhooks
func (s *integrationService) ProcessWhatsAppWebhook(ctx context.Context, payload []byte, signature string) error {
	return s.processWebhook(ctx, domain.PlatformWhatsApp, payload, signature)
}

func (s *integrationService) ProcessMessengerWebhook(ctx context.Context, payload []byte, signature string) error {
	return s.processWebhook(ctx, domain.PlatformMessenger, payload, signature)
}

func (s *integrationService) ProcessInstagramWebhook(ctx context.Context, payload []byte, signature string) error {
	return s.processWebhook(ctx, domain.PlatformInstagram, payload, signature)
}

func (s *integrationService) ProcessTelegramWebhook(ctx context.Context, payload []byte) error {
	return s.processWebhook(ctx, domain.PlatformTelegram, payload, "")
}

func (s *integrationService) ProcessWebchatWebhook(ctx context.Context, payload []byte) error {
	return s.processWebhook(ctx, domain.PlatformWebchat, payload, "")
}

func (s *integrationService) ProcessMailchimpWebhook(ctx context.Context, payload []byte, signature string) error {
	return s.processWebhook(ctx, domain.PlatformMailchimp, payload, signature)
}

// Consulta de mensajes entrantes
func (s *integrationService) GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error) {
	if s.inboundRepo == nil {
		s.logger.Warn("Inbound repository is nil, returning mock data")
		return s.getMockInboundMessages(platform, limit)
	}
	// Por ahora solo retornamos mensajes no procesados
	return s.inboundRepo.GetUnprocessed(ctx, limit)
}

// Helper functions
func (s *integrationService) processWebhook(ctx context.Context, platform domain.Platform, payload []byte, signature string) error {
	s.logger.Info("Processing webhook", map[string]interface{}{
		"platform":     platform,
		"payload_size": len(payload),
	})

	// Guardar mensaje entrante
	message := &domain.InboundMessage{
		ID:         uuid.New().String(),
		Platform:   platform,
		Payload:    payload,
		ReceivedAt: time.Now(),
		Processed:  false,
	}

	if s.inboundRepo != nil {
		if err := s.inboundRepo.Create(ctx, message); err != nil {
			s.logger.Error("Failed to save inbound message", err)
		}
	}

	// Normalizar mensaje
	normalizedMessage, err := s.webhookService.NormalizeMessage(platform, payload)
	if err != nil {
		s.logger.Error("Failed to normalize message", err)
		return err
	}

	// Marcar como procesado
	if s.inboundRepo != nil {
		if err := s.inboundRepo.MarkAsProcessed(ctx, message.ID); err != nil {
			s.logger.Error("Failed to mark message as processed", err)
		}
	}

	// Reenviar al servicio de mensajería
	if err := s.webhookService.ForwardToMessagingService(ctx, normalizedMessage); err != nil {
		s.logger.Error("Failed to forward message to messaging service", err)
		return err
	}

	s.logger.Info("Webhook processed successfully", map[string]interface{}{
		"platform":   platform,
		"message_id": normalizedMessage.MessageID,
	})

	return nil
}

// Mock data para desarrollo
func (s *integrationService) getMockInboundMessages(platform string, limit int) ([]*domain.InboundMessage, error) {
	mockMessages := []*domain.InboundMessage{
		{
			ID:         "mock-1",
			Platform:   domain.Platform(platform),
			Payload:    json.RawMessage(`{"text": "Mensaje de prueba"}`),
			ReceivedAt: time.Now().Add(-time.Hour),
			Processed:  true,
		},
	}

	if limit > len(mockMessages) {
		return mockMessages, nil
	}
	return mockMessages[:limit], nil
}
