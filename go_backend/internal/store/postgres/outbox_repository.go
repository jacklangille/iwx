package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"iwx/go_backend/internal/outbox"
)

type OutboxRepository struct {
	*baseRepository
}

func NewOutboxRepository(databaseURL string) *OutboxRepository {
	return &OutboxRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *OutboxRepository) EnqueueOutboxEvent(ctx context.Context, event outbox.Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO outbox_events (
			event_id,
			event_type,
			payload
		)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (event_id) DO NOTHING
	`, strings.TrimSpace(event.EventID), strings.TrimSpace(event.EventType), string(event.Payload))
	return err
}

func (r *OutboxRepository) ListPendingOutboxEvents(ctx context.Context, limit int) ([]outbox.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			event_id,
			event_type,
			payload::text,
			created_at,
			published_at,
			attempt_count,
			last_error
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY id ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []outbox.Event{}
	for rows.Next() {
		var event outbox.Event
		var payload string
		var publishedAt sql.NullTime
		var lastError sql.NullString
		if err := rows.Scan(
			&event.ID,
			&event.EventID,
			&event.EventType,
			&payload,
			&event.CreatedAt,
			&publishedAt,
			&event.AttemptCount,
			&lastError,
		); err != nil {
			return nil, err
		}
		event.Payload = []byte(payload)
		event.CreatedAt = event.CreatedAt.UTC()
		event.PublishedAt = nullableTime(publishedAt)
		event.LastError = nullableString(lastError)
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *OutboxRepository) MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET published_at = $2, attempt_count = attempt_count + 1, last_error = NULL
		WHERE event_id = $1
	`, strings.TrimSpace(eventID), publishedAt.UTC())
	return err
}

func (r *OutboxRepository) MarkOutboxEventFailed(ctx context.Context, eventID, lastError string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET attempt_count = attempt_count + 1, last_error = $2
		WHERE event_id = $1
	`, strings.TrimSpace(eventID), strings.TrimSpace(lastError))
	return err
}
