package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// QueryService define las operaciones para consultas de mensajes
type QueryService interface {
	GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error)
	GetOutboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.OutboundMessageLog, error)
	GetChatHistory(ctx context.Context, platform, userID string) (*domain.ChatHistory, error)
}

type queryService struct {
	channelRepo  domain.ChannelIntegrationRepository
	inboundRepo  domain.InboundMessageRepository
	outboundRepo domain.OutboundMessageLogRepository
	logger       logger.Logger
}

// NewQueryService crea una nueva instancia del servicio de consultas
func NewQueryService(
	channelRepo domain.ChannelIntegrationRepository,
	inboundRepo domain.InboundMessageRepository,
	outboundRepo domain.OutboundMessageLogRepository,
	logger logger.Logger,
) QueryService {
	return &queryService{
		channelRepo:  channelRepo,
		inboundRepo:  inboundRepo,
		outboundRepo: outboundRepo,
		logger:       logger,
	}
}

// GetInboundMessages obtiene mensajes entrantes con filtros
func (s *queryService) GetInboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.InboundMessage, error) {
	if s.channelRepo == nil {
		// Mock response for development
		return []*domain.InboundMessage{
			{
				ID:         "mock-inbound-1",
				Platform:   domain.Platform(platform),
				Payload:    []byte(`{"message": {"text": "Mensaje de prueba mock"}}`),
				ReceivedAt: time.Now().Add(-2 * time.Hour),
				Processed:  true,
			},
		}, nil
	}

	// Construir query con filtros opcionales
	query := `SELECT id, platform, payload, received_at, processed 
			  FROM inbound_messages 
			  WHERE ($1 = '' OR platform = $1) 
			  ORDER BY received_at DESC 
			  LIMIT $2 OFFSET $3`

	rows, err := s.channelRepo.DB().QueryContext(ctx, query, platform, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbound messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.InboundMessage
	for rows.Next() {
		var msg domain.InboundMessage
		if err := rows.Scan(&msg.ID, &msg.Platform, &msg.Payload, &msg.ReceivedAt, &msg.Processed); err != nil {
			s.logger.Error("Failed to scan inbound message", err)
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// GetOutboundMessages obtiene mensajes salientes con filtros
func (s *queryService) GetOutboundMessages(ctx context.Context, platform string, limit, offset int) ([]*domain.OutboundMessageLog, error) {
	if s.outboundRepo == nil {
		// Mock response for development
		return []*domain.OutboundMessageLog{
			{
				ID:        "mock-outbound-1",
				ChannelID: "mock-channel-1",
				Recipient: "573001234567",
				Status:    domain.MessageStatusSent,
				Timestamp: time.Now().Add(-1 * time.Hour),
			},
		}, nil
	}

	// Construir query con filtros opcionales
	query := `SELECT id, channel_id, recipient, content, status, response, timestamp 
			  FROM outbound_message_logs 
			  WHERE ($1 = '' OR channel_id IN (
				  SELECT id FROM channel_integrations WHERE platform = $1
			  ))
			  ORDER BY timestamp DESC 
			  LIMIT $2 OFFSET $3`

	rows, err := s.channelRepo.DB().QueryContext(ctx, query, platform, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbound messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.OutboundMessageLog
	for rows.Next() {
		var msg domain.OutboundMessageLog
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.Recipient, &msg.Content, &msg.Status, &msg.Response, &msg.Timestamp); err != nil {
			s.logger.Error("Failed to scan outbound message", err)
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// GetChatHistory obtiene el historial de conversación con un usuario específico
func (s *queryService) GetChatHistory(ctx context.Context, platform, userID string) (*domain.ChatHistory, error) {
	// Query para obtener mensajes entrantes del usuario
	inboundQuery := `
		SELECT id, payload, received_at 
		FROM inbound_messages 
		WHERE platform = $1 
		ORDER BY received_at ASC`

	// Query para obtener mensajes salientes al usuario
	outboundQuery := `
		SELECT id, content, timestamp, status 
		FROM outbound_message_logs 
		WHERE recipient = $1 
		AND channel_id IN (
			SELECT id FROM channel_integrations WHERE platform = $2
		)
		ORDER BY timestamp ASC`

	var messages []domain.ChatMessage

	// Obtener mensajes entrantes
	rows, err := s.channelRepo.DB().QueryContext(ctx, inboundQuery, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbound messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var payload []byte
		var receivedAt time.Time

		if err := rows.Scan(&id, &payload, &receivedAt); err != nil {
			s.logger.Error("Failed to scan inbound message", err)
			continue
		}

		// Extraer texto del payload (simplificado)
		var payloadData map[string]interface{}
		if err := json.Unmarshal(payload, &payloadData); err != nil {
			continue
		}

		text := extractTextFromPayload(payloadData, domain.Platform(platform))

		messages = append(messages, domain.ChatMessage{
			ID:        id,
			Type:      "inbound",
			Platform:  domain.Platform(platform),
			UserID:    userID,
			Text:      text,
			Timestamp: receivedAt,
		})
	}

	// Obtener mensajes salientes
	rows, err = s.channelRepo.DB().QueryContext(ctx, outboundQuery, userID, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbound messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var content []byte
		var timestamp time.Time
		var status string

		if err := rows.Scan(&id, &content, &timestamp, &status); err != nil {
			s.logger.Error("Failed to scan outbound message", err)
			continue
		}

		// Extraer texto del contenido
		var contentData map[string]interface{}
		if err := json.Unmarshal(content, &contentData); err != nil {
			continue
		}

		text := ""
		if textVal, ok := contentData["text"].(string); ok {
			text = textVal
		}

		messages = append(messages, domain.ChatMessage{
			ID:        id,
			Type:      "outbound",
			Platform:  domain.Platform(platform),
			UserID:    userID,
			Text:      text,
			Timestamp: timestamp,
			Status:    status,
		})
	}

	// Ordenar mensajes por timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return &domain.ChatHistory{
		Platform:   domain.Platform(platform),
		UserID:     userID,
		Messages:   messages,
		TotalCount: len(messages),
	}, nil
}

// extractTextFromPayload extrae texto de diferentes formatos de payload
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
