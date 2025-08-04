package repository

import (
	"context"
	"fmt"

	"github.com/company/microservice-template/internal/domain"
)

type inboundMessageRepository struct {
	db *PostgresDB
}

// NewInboundMessageRepository creates a new inbound message repository
func NewInboundMessageRepository(db *PostgresDB) domain.InboundMessageRepository {
	return &inboundMessageRepository{db: db}
}

func (r *inboundMessageRepository) Create(ctx context.Context, message *domain.InboundMessage) error {
	query := `
		INSERT INTO inbound_messages (id, platform, payload, received_at, processed)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.DB.ExecContext(ctx, query,
		message.ID,
		message.Platform,
		message.Payload,
		message.ReceivedAt,
		message.Processed,
	)

	if err != nil {
		return fmt.Errorf("failed to create inbound message: %w", err)
	}

	return nil
}

func (r *inboundMessageRepository) GetUnprocessed(ctx context.Context, limit int) ([]*domain.InboundMessage, error) {
	query := `
		SELECT id, platform, payload, received_at, processed
		FROM inbound_messages
		WHERE processed = false
		ORDER BY received_at ASC
		LIMIT $1`

	rows, err := r.db.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.InboundMessage

	for rows.Next() {
		var message domain.InboundMessage

		err := rows.Scan(
			&message.ID,
			&message.Platform,
			&message.Payload,
			&message.ReceivedAt,
			&message.Processed,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan inbound message: %w", err)
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return messages, nil
}

func (r *inboundMessageRepository) MarkAsProcessed(ctx context.Context, id string) error {
	query := `UPDATE inbound_messages SET processed = true WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inbound message not found")
	}

	return nil
}