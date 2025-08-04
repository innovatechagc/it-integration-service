package repository

import (
	"context"
	"fmt"

	"github.com/company/microservice-template/internal/domain"
)

type outboundMessageLogRepository struct {
	db *PostgresDB
}

// NewOutboundMessageLogRepository creates a new outbound message log repository
func NewOutboundMessageLogRepository(db *PostgresDB) domain.OutboundMessageLogRepository {
	return &outboundMessageLogRepository{db: db}
}

func (r *outboundMessageLogRepository) Create(ctx context.Context, log *domain.OutboundMessageLog) error {
	query := `
		INSERT INTO outbound_message_logs (id, channel_id, recipient, content, status, response, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.DB.ExecContext(ctx, query,
		log.ID,
		log.ChannelID,
		log.Recipient,
		log.Content,
		log.Status,
		log.Response,
		log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create outbound message log: %w", err)
	}

	return nil
}

func (r *outboundMessageLogRepository) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*domain.OutboundMessageLog, error) {
	query := `
		SELECT id, channel_id, recipient, content, status, response, timestamp
		FROM outbound_message_logs
		WHERE channel_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.DB.QueryContext(ctx, query, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbound message logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.OutboundMessageLog

	for rows.Next() {
		var log domain.OutboundMessageLog

		err := rows.Scan(
			&log.ID,
			&log.ChannelID,
			&log.Recipient,
			&log.Content,
			&log.Status,
			&log.Response,
			&log.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan outbound message log: %w", err)
		}

		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

func (r *outboundMessageLogRepository) GetByStatus(ctx context.Context, status domain.MessageStatus, limit int) ([]*domain.OutboundMessageLog, error) {
	query := `
		SELECT id, channel_id, recipient, content, status, response, timestamp
		FROM outbound_message_logs
		WHERE status = $1
		ORDER BY timestamp ASC
		LIMIT $2`

	rows, err := r.db.DB.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbound message logs by status: %w", err)
	}
	defer rows.Close()

	var logs []*domain.OutboundMessageLog

	for rows.Next() {
		var log domain.OutboundMessageLog

		err := rows.Scan(
			&log.ID,
			&log.ChannelID,
			&log.Recipient,
			&log.Content,
			&log.Status,
			&log.Response,
			&log.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan outbound message log: %w", err)
		}

		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

func (r *outboundMessageLogRepository) UpdateStatus(ctx context.Context, id string, status domain.MessageStatus, response []byte) error {
	query := `
		UPDATE outbound_message_logs
		SET status = $2, response = $3
		WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id, status, response)
	if err != nil {
		return fmt.Errorf("failed to update outbound message log status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outbound message log not found")
	}

	return nil
}