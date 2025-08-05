package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"it-integration-service/internal/domain"
)

type channelIntegrationRepository struct {
	db *PostgresDB
}

// NewChannelIntegrationRepository creates a new channel integration repository
func NewChannelIntegrationRepository(db *PostgresDB) domain.ChannelIntegrationRepository {
	return &channelIntegrationRepository{db: db}
}

func (r *channelIntegrationRepository) Create(ctx context.Context, integration *domain.ChannelIntegration) error {
	query := `
		INSERT INTO channel_integrations (id, tenant_id, platform, provider, access_token, webhook_url, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = r.db.DB.ExecContext(ctx, query,
		integration.ID,
		integration.TenantID,
		string(integration.Platform),
		string(integration.Provider),
		integration.AccessToken,
		integration.WebhookURL,
		string(integration.Status),
		configJSON,
		integration.CreatedAt,
		integration.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create channel integration (query: %s): %w", query, err)
	}

	return nil
}

func (r *channelIntegrationRepository) GetByID(ctx context.Context, id string) (*domain.ChannelIntegration, error) {
	query := `
		SELECT id, tenant_id, platform, provider, access_token, webhook_url, status, config, created_at, updated_at
		FROM channel_integrations
		WHERE id = $1`

	var integration domain.ChannelIntegration
	var configJSON []byte

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&integration.ID,
		&integration.TenantID,
		&integration.Platform,
		&integration.Provider,
		&integration.AccessToken,
		&integration.WebhookURL,
		&integration.Status,
		&configJSON,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("channel integration not found")
		}
		return nil, fmt.Errorf("failed to get channel integration: %w", err)
	}

	if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &integration, nil
}

func (r *channelIntegrationRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
	query := `
		SELECT id, tenant_id, platform, provider, access_token, webhook_url, status, config, created_at, updated_at
		FROM channel_integrations
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query channel integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*domain.ChannelIntegration

	for rows.Next() {
		var integration domain.ChannelIntegration
		var configJSON []byte

		err := rows.Scan(
			&integration.ID,
			&integration.TenantID,
			&integration.Platform,
			&integration.Provider,
			&integration.AccessToken,
			&integration.WebhookURL,
			&integration.Status,
			&configJSON,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan channel integration: %w", err)
		}

		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		integrations = append(integrations, &integration)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return integrations, nil
}

func (r *channelIntegrationRepository) Update(ctx context.Context, integration *domain.ChannelIntegration) error {
	query := `
		UPDATE channel_integrations
		SET tenant_id = $2, platform = $3, provider = $4, access_token = $5, webhook_url = $6, status = $7, config = $8, updated_at = $9
		WHERE id = $1`

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	result, err := r.db.DB.ExecContext(ctx, query,
		integration.ID,
		integration.TenantID,
		integration.Platform,
		integration.Provider,
		integration.AccessToken,
		integration.WebhookURL,
		integration.Status,
		configJSON,
		integration.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update channel integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("channel integration not found")
	}

	return nil
}

func (r *channelIntegrationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_integrations WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete channel integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("channel integration not found")
	}

	return nil
}

func (r *channelIntegrationRepository) GetByPlatformAndTenant(ctx context.Context, platform domain.Platform, tenantID string) (*domain.ChannelIntegration, error) {
	query := `
		SELECT id, tenant_id, platform, provider, access_token, webhook_url, status, config, created_at, updated_at
		FROM channel_integrations
		WHERE platform = $1 AND tenant_id = $2
		LIMIT 1`

	var integration domain.ChannelIntegration
	var configJSON []byte

	err := r.db.DB.QueryRowContext(ctx, query, platform, tenantID).Scan(
		&integration.ID,
		&integration.TenantID,
		&integration.Platform,
		&integration.Provider,
		&integration.AccessToken,
		&integration.WebhookURL,
		&integration.Status,
		&configJSON,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("channel integration not found")
		}
		return nil, fmt.Errorf("failed to get channel integration: %w", err)
	}

	if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &integration, nil
}

// DB returns the database connection for direct queries
func (r *channelIntegrationRepository) DB() *sql.DB {
	return r.db.DB
}