package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/pkg/logging"
)

const outboxEventTypeContractResolved = "contract_resolved"

type OutboxDispatcher struct {
	repo                outboxRepository
	contractPublisher   EventPublisher
	projectionPublisher projectionChangePublisher
}

type projectionChangePublisher interface {
	PublishProjectionChange(ctx context.Context, event events.ProjectionChange) error
}

type outboxRepository interface {
	ListPendingOutboxEvents(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkOutboxEventFailed(ctx context.Context, eventID, lastError string) error
	MarkResolutionPublished(ctx context.Context, eventID string, publishedAt time.Time) error
}

func NewOutboxDispatcher(repo outboxRepository, contractPublisher EventPublisher, projectionPublisher projectionChangePublisher) *OutboxDispatcher {
	return &OutboxDispatcher{repo: repo, contractPublisher: contractPublisher, projectionPublisher: projectionPublisher}
}

func (d *OutboxDispatcher) Run(ctx context.Context) error {
	if d == nil || d.repo == nil {
		return nil
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := d.dispatchOnce(ctx); err != nil {
			logging.Error(ctx, "oracle_outbox_dispatch_failed", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (d *OutboxDispatcher) dispatchOnce(ctx context.Context) error {
	eventsToDispatch, err := d.repo.ListPendingOutboxEvents(ctx, 100)
	if err != nil {
		return err
	}

	for _, pending := range eventsToDispatch {
		if err := d.dispatchEvent(ctx, pending); err != nil {
			_ = d.repo.MarkOutboxEventFailed(ctx, pending.EventID, err.Error())
			continue
		}
		publishedAt := time.Now().UTC()
		if err := d.repo.MarkOutboxEventPublished(ctx, pending.EventID, publishedAt); err != nil {
			return err
		}
		if pending.EventType == outboxEventTypeContractResolved {
			if err := d.repo.MarkResolutionPublished(ctx, pending.EventID, publishedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *OutboxDispatcher) dispatchEvent(ctx context.Context, pending outbox.Event) error {
	switch pending.EventType {
	case outboxEventTypeContractResolved:
		var event events.ContractResolved
		if err := json.Unmarshal(pending.Payload, &event); err != nil {
			return err
		}
		if d.contractPublisher == nil {
			return fmt.Errorf("contract publisher unavailable")
		}
		return d.contractPublisher.PublishContractResolved(ctx, event)
	case outbox.EventTypeProjectionChange:
		var event events.ProjectionChange
		if err := json.Unmarshal(pending.Payload, &event); err != nil {
			return err
		}
		if d.projectionPublisher == nil {
			return fmt.Errorf("projection publisher unavailable")
		}
		return d.projectionPublisher.PublishProjectionChange(ctx, event)
	default:
		return fmt.Errorf("unsupported outbox event type: %s", pending.EventType)
	}
}
