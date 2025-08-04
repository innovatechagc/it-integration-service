package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/microservice-template/internal/domain"
	"github.com/company/microservice-template/pkg/logger"
	"github.com/google/uuid"
)

type integrationService struct {
	channelRepo     domain.ChannelIntegrationRepository
	inboundRepo     domain.InboundMessageRepository
	outboundRepo    domain.OutboundMessageLogRepository
	webhookService  WebhookService
	providerService MessagingProviderService
	logger          logger.Logger
}

// NewIntegrationService crea una nueva instancia del servicio de integración
func NewIntegrationService(
	channelRepo domain.ChannelIntegrationRepository,
	inboundRepo domain.InboundMessageRepository,
	outboundRepo domain.OutboundMessageLogRepository,
	webhookService WebhookService,
	providerService MessagingProviderService,
	logger logger.Logger,
) IntegrationService {
	return &integrationService{
		channelRepo:     channelRepo,
		inboundRepo:     inboundRepo,
		outboundRepo:    outboundRepo,
		webhookService:  webhookService,
		providerService: providerService,
		logger:          logger,
	}
}

func (s *integrationService) CreateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
	integration.ID = uuid.New().String()
	integration.CreatedAt = time.Now()
	integration.UpdatedAt = time.Now()
	integration.Status = domain.StatusActive

	if s.channelRepo != nil {
		if err := s.channelRepo.Create(ctx, integration); err != nil {
			s.logger.Error("Failed to create channel integration", 
				"error", err.Error(),
				"integration_id", integration.ID,
				"tenant_id", integration.TenantID,
				"platform", integration.Platform,
			)
			return fmt.Errorf("failed to create channel integration: %w", err)
		}
	}

	s.logger.Info("Channel integration created", map[string]interface{}{
		"id":       integration.ID,
		"platform": integration.Platform,
		"tenant":   integration.TenantID,
	})

	return nil
}

func (s *integrationService) GetChannel(ctx context.Context, id string) (*domain.ChannelIntegration, error) {
	if s.channelRepo == nil {
		// Mock response for development
		return &domain.ChannelIntegration{
			ID:       id,
			TenantID: "mock-tenant",
			Platform: domain.PlatformWhatsApp,
			Provider: domain.ProviderMeta,
			Status:   domain.StatusActive,
		}, nil
	}
	return s.channelRepo.GetByID(ctx, id)
}

func (s *integrationService) GetChannelsByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
	if s.channelRepo == nil {
		// Mock response for development
		return []*domain.ChannelIntegration{
			{
				ID:       "mock-channel-1",
				TenantID: tenantID,
				Platform: domain.PlatformWhatsApp,
				Provider: domain.ProviderMeta,
				Status:   domain.StatusActive,
			},
		}, nil
	}
	return s.channelRepo.GetByTenantID(ctx, tenantID)
}

func (s *integrationService) UpdateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
	integration.UpdatedAt = time.Now()
	if s.channelRepo == nil {
		s.logger.Info("Mock: Channel updated", map[string]interface{}{"id": integration.ID})
		return nil
	}
	return s.channelRepo.Update(ctx, integration)
}

func (s *integrationService) DeleteChannel(ctx context.Context, id string) error {
	if s.channelRepo == nil {
		s.logger.Info("Mock: Channel deleted", map[string]interface{}{"id": id})
		return nil
	}
	return s.channelRepo.Delete(ctx, id)
}

func (s *integrationService) SendMessage(ctx context.Context, request *domain.SendMessageRequest) error {
	// Obtener la integración del canal
	var integration *domain.ChannelIntegration
	var err error
	
	if s.channelRepo != nil {
		integration, err = s.channelRepo.GetByID(ctx, request.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to get channel integration: %w", err)
		}
	} else {
		// Mock integration for development
		integration = &domain.ChannelIntegration{
			ID:       request.ChannelID,
			TenantID: "mock-tenant",
			Platform: domain.PlatformWhatsApp,
			Provider: domain.ProviderMeta,
			Status:   domain.StatusActive,
			Config:   []byte(`{"phone_number_id": "mock-phone"}`),
		}
	}

	if integration.Status != domain.StatusActive {
		return fmt.Errorf("channel integration is not active")
	}

	// Crear log de mensaje saliente
	logEntry := &domain.OutboundMessageLog{
		ID:        uuid.New().String(),
		ChannelID: request.ChannelID,
		Recipient: request.Recipient,
		Status:    domain.MessageStatusQueued,
		Timestamp: time.Now(),
	}

	contentBytes, _ := json.Marshal(request.Content)
	logEntry.Content = contentBytes

	if s.outboundRepo != nil {
		if err := s.outboundRepo.Create(ctx, logEntry); err != nil {
			s.logger.Error("Failed to create outbound message log", err)
		}
	}

	// Enviar mensaje según la plataforma
	var sendErr error
	switch integration.Platform {
	case domain.PlatformWhatsApp:
		sendErr = s.providerService.SendWhatsAppMessage(ctx, integration, request.Recipient, &request.Content)
	case domain.PlatformMessenger:
		sendErr = s.providerService.SendMessengerMessage(ctx, integration, request.Recipient, &request.Content)
	case domain.PlatformInstagram:
		sendErr = s.providerService.SendInstagramMessage(ctx, integration, request.Recipient, &request.Content)
	case domain.PlatformTelegram:
		sendErr = s.providerService.SendTelegramMessage(ctx, integration, request.Recipient, &request.Content)
	case domain.PlatformWebchat:
		sendErr = s.providerService.SendWebchatMessage(ctx, integration, request.Recipient, &request.Content)
	default:
		sendErr = fmt.Errorf("unsupported platform: %s", integration.Platform)
	}

	// Actualizar estado del log
	status := domain.MessageStatusSent
	if sendErr != nil {
		status = domain.MessageStatusFailed
		s.logger.Error("Failed to send message", sendErr)
	}

	responseBytes, _ := json.Marshal(map[string]interface{}{
		"error": sendErr,
	})
	
	if s.outboundRepo != nil {
		if err := s.outboundRepo.UpdateStatus(ctx, logEntry.ID, status, responseBytes); err != nil {
			s.logger.Error("Failed to update outbound message status", err)
		}
	}

	return sendErr
}

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

func (s *integrationService) processWebhook(ctx context.Context, platform domain.Platform, payload []byte, signature string) error {
	// Crear registro de mensaje entrante
	inboundMessage := &domain.InboundMessage{
		ID:         uuid.New().String(),
		Platform:   platform,
		Payload:    payload,
		ReceivedAt: time.Now(),
		Processed:  false,
	}

	if s.inboundRepo != nil {
		if err := s.inboundRepo.Create(ctx, inboundMessage); err != nil {
			s.logger.Error("Failed to create inbound message", err)
		}
	}

	// Normalizar mensaje
	normalizedMessage, err := s.webhookService.NormalizeMessage(platform, payload)
	if err != nil {
		s.logger.Error("Failed to normalize message", err)
		return fmt.Errorf("failed to normalize message: %w", err)
	}

	// Reenviar al messaging service
	if err := s.webhookService.ForwardToMessagingService(ctx, normalizedMessage); err != nil {
		s.logger.Error("Failed to forward message to messaging service", err)
		return fmt.Errorf("failed to forward message: %w", err)
	}

	// Marcar como procesado
	if s.inboundRepo != nil {
		if err := s.inboundRepo.MarkAsProcessed(ctx, inboundMessage.ID); err != nil {
			s.logger.Error("Failed to mark message as processed", err)
		}
	}

	s.logger.Info("Webhook processed successfully", map[string]interface{}{
		"platform":   platform,
		"message_id": normalizedMessage.MessageID,
	})

	return nil
}