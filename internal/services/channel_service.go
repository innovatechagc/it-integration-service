package services

import (
	"context"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"

	"github.com/google/uuid"
)

// ChannelService define las operaciones para gestión de canales
type ChannelService interface {
	CreateChannel(ctx context.Context, integration *domain.ChannelIntegration) error
	GetChannel(ctx context.Context, id string) (*domain.ChannelIntegration, error)
	GetChannelsByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error)
	UpdateChannel(ctx context.Context, integration *domain.ChannelIntegration) error
	DeleteChannel(ctx context.Context, id string) error
}

type channelService struct {
	channelRepo domain.ChannelIntegrationRepository
	logger      logger.Logger
}

// NewChannelService crea una nueva instancia del servicio de canales
func NewChannelService(channelRepo domain.ChannelIntegrationRepository, logger logger.Logger) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
		logger:      logger,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
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

func (s *channelService) GetChannel(ctx context.Context, id string) (*domain.ChannelIntegration, error) {
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

func (s *channelService) GetChannelsByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
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

func (s *channelService) UpdateChannel(ctx context.Context, integration *domain.ChannelIntegration) error {
	integration.UpdatedAt = time.Now()
	if s.channelRepo == nil {
		s.logger.Info("Mock: Channel updated", map[string]interface{}{"id": integration.ID})
		return nil
	}
	return s.channelRepo.Update(ctx, integration)
}

func (s *channelService) DeleteChannel(ctx context.Context, id string) error {
	if s.channelRepo == nil {
		s.logger.Info("Mock: Channel deleted", map[string]interface{}{"id": id})
		return nil
	}
	return s.channelRepo.Delete(ctx, id)
}
