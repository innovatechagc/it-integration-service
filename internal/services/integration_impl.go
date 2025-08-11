package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"

	"github.com/google/uuid"
)

type integrationService struct {
	channelService  ChannelService
	queryService    QueryService
	webhookService  WebhookService
	providerService MessagingProviderService
	inboundRepo     domain.InboundMessageRepository
	outboundRepo    domain.OutboundMessageLogRepository
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
	channelService := NewChannelService(channelRepo, logger)
	queryService := NewQueryService(channelRepo, inboundRepo, outboundRepo, logger)

	return &integrationService{
		channelService:  channelService,
		queryService:    queryService,
		webhookService:  webhookService,
		providerService: providerService,
		inboundRepo:     inboundRepo,
		outboundRepo:    outboundRepo,
		logger:          logger,
	}
}

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

func (s *integrationService) SendMessage(ctx context.Context, request *domain.SendMessageRequest) error {
	s.logger.Info("SendMessage called", map[string]interface{}{
		"channel_id": request.ChannelID,
		"recipient":  request.Recipient,
	})

	// Obtener la integración del canal
	integration, err := s.channelService.GetChannel(ctx, request.ChannelID)
	if err != nil {
		s.logger.Error("Failed to get channel integration", err)
		return fmt.Errorf("failed to get channel integration: %w", err)
	}

	s.logger.Info("Channel integration found", map[string]interface{}{
		"platform": integration.Platform,
		"status":   integration.Status,
	})

	if integration.Status != domain.StatusActive {
		s.logger.Error("Channel integration is not active", fmt.Errorf("status: %s", integration.Status))
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

	s.logger.Info("Creating outbound message log", map[string]interface{}{
		"log_id":     logEntry.ID,
		"channel_id": logEntry.ChannelID,
		"recipient":  logEntry.Recipient,
	})

	if s.outboundRepo != nil {
		if err := s.outboundRepo.Create(ctx, logEntry); err != nil {
			s.logger.Error("Failed to create outbound message log", err)
		} else {
			s.logger.Info("Outbound message log created successfully", map[string]interface{}{
				"log_id": logEntry.ID,
			})
		}
	} else {
		s.logger.Warn("Outbound repository is nil, cannot create log")
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
	s.logger.Info("Processing webhook", map[string]interface{}{
		"platform":     platform,
		"payload_size": len(payload),
	})

	// Crear registro de mensaje entrante
	inboundMessage := &domain.InboundMessage{
		ID:         uuid.New().String(),
		Platform:   platform,
		Payload:    payload,
		ReceivedAt: time.Now(),
		Processed:  false,
	}

	s.logger.Info("Created inbound message", map[string]interface{}{
		"message_id": inboundMessage.ID,
		"platform":   platform,
	})

	if s.inboundRepo != nil {
		if err := s.inboundRepo.Create(ctx, inboundMessage); err != nil {
			s.logger.Error("Failed to create inbound message", err)
		} else {
			s.logger.Info("Inbound message saved to database")
		}
	}

	// Normalizar mensaje
	s.logger.Info("Normalizing message...")
	normalizedMessage, err := s.webhookService.NormalizeMessage(platform, payload)
	if err != nil {
		s.logger.Error("Failed to normalize message", err)
		return fmt.Errorf("failed to normalize message: %w", err)
	}

	s.logger.Info("Message normalized successfully", map[string]interface{}{
		"message_id": normalizedMessage.MessageID,
		"sender":     normalizedMessage.Sender,
		"text":       normalizedMessage.Content.Text,
	})

	// Reenviar al messaging service
	s.logger.Info("Forwarding to messaging service...")
	if err := s.webhookService.ForwardToMessagingService(ctx, normalizedMessage); err != nil {
		s.logger.Error("Failed to forward message to messaging service", err)
		return fmt.Errorf("failed to forward message: %w", err)
	}

	// Marcar como procesado
	if s.inboundRepo != nil {
		if err := s.inboundRepo.MarkAsProcessed(ctx, inboundMessage.ID); err != nil {
			s.logger.Error("Failed to mark message as processed", err)
		} else {
			s.logger.Info("Message marked as processed")
		}
	}

	s.logger.Info("Webhook processed successfully", map[string]interface{}{
		"platform":   platform,
		"message_id": normalizedMessage.MessageID,
	})

	return nil
}

// GetInboundMessages obtiene mensajes entrantes con filtros
func (s *integrationService) GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error) {
	return s.queryService.GetInboundMessages(ctx, platform, limit, offset)
}

// GetChatHistory obtiene el historial de conversación con un usuario específico
func (s *integrationService) GetChatHistory(ctx context.Context, platform, userID string) (*domain.ChatHistory, error) {
	return s.queryService.GetChatHistory(ctx, platform, userID)
}

// GetOutboundMessages obtiene mensajes salientes con filtros
func (s *integrationService) GetOutboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.OutboundMessageLog, error) {
	return s.queryService.GetOutboundMessages(ctx, platform, limit, offset)
}

// BroadcastMessage envía un mensaje a múltiples destinatarios en diferentes plataformas
func (s *integrationService) BroadcastMessage(ctx context.Context, request *domain.BroadcastMessageRequest) (*domain.BroadcastResult, error) {
	result := &domain.BroadcastResult{
		Results: make([]domain.BroadcastItemResult, 0),
	}

	// Obtener integraciones activas para el tenant y las plataformas solicitadas
	channels, err := s.channelService.GetChannelsByTenant(ctx, request.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	// Filtrar canales por plataformas solicitadas
	platformChannels := make(map[domain.Platform]*domain.ChannelIntegration)
	for _, channel := range channels {
		if channel.Status != domain.StatusActive {
			continue
		}
		for _, platform := range request.Platforms {
			if channel.Platform == platform {
				platformChannels[platform] = channel
				break
			}
		}
	}

	// Enviar mensaje a cada destinatario
	for _, recipient := range request.Recipients {
		for _, platform := range request.Platforms {
			channel, exists := platformChannels[platform]
			if !exists {
				result.Results = append(result.Results, domain.BroadcastItemResult{
					Platform:  platform,
					Recipient: recipient,
					Success:   false,
					Error:     fmt.Sprintf("No active channel found for platform %s", platform),
				})
				result.TotalFailed++
				continue
			}

			// Crear solicitud de envío individual
			sendRequest := &domain.SendMessageRequest{
				ChannelID: channel.ID,
				Recipient: recipient,
				Content:   request.Content,
			}

			// Enviar mensaje
			err := s.SendMessage(ctx, sendRequest)
			if err != nil {
				result.Results = append(result.Results, domain.BroadcastItemResult{
					Platform:  platform,
					Recipient: recipient,
					Success:   false,
					Error:     err.Error(),
				})
				result.TotalFailed++
			} else {
				result.Results = append(result.Results, domain.BroadcastItemResult{
					Platform:  platform,
					Recipient: recipient,
					Success:   true,
					MessageID: fmt.Sprintf("broadcast-%s-%s", platform, recipient),
				})
				result.TotalSent++
			}
		}
	}

	s.logger.Info("Broadcast completed", map[string]interface{}{
		"tenant_id":    request.TenantID,
		"total_sent":   result.TotalSent,
		"total_failed": result.TotalFailed,
		"platforms":    request.Platforms,
		"recipients":   len(request.Recipients),
	})

	return result, nil
}

// Helper function para extraer texto de diferentes formatos de payload
func extractTextFromPayload(payload map[string]interface{}, platform domain.Platform) string {
	switch platform {
	case domain.PlatformWhatsApp:
		if entry, ok := payload["entry"].([]interface{}); ok && len(entry) > 0 {
			if entryObj, ok := entry[0].(map[string]interface{}); ok {
				if changes, ok := entryObj["changes"].([]interface{}); ok && len(changes) > 0 {
					if changeObj, ok := changes[0].(map[string]interface{}); ok {
						if value, ok := changeObj["value"].(map[string]interface{}); ok {
							if messages, ok := value["messages"].([]interface{}); ok && len(messages) > 0 {
								if msgObj, ok := messages[0].(map[string]interface{}); ok {
									if text, ok := msgObj["text"].(map[string]interface{}); ok {
										if body, ok := text["body"].(string); ok {
											return body
										}
									}
								}
							}
						}
					}
				}
			}
		}
	case domain.PlatformTelegram:
		if message, ok := payload["message"].(map[string]interface{}); ok {
			if text, ok := message["text"].(string); ok {
				return text
			}
		}
	}
	return ""
}
