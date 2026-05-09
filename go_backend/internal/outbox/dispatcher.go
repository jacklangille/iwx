package outbox

import (
	"context"
	"time"

	"iwx/go_backend/pkg/logging"
)

type Repository interface {
	EnqueueOutboxEvent(ctx context.Context, event Event) error
	ListPendingOutboxEvents(ctx context.Context, limit int) ([]Event, error)
	MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkOutboxEventFailed(ctx context.Context, eventID, lastError string) error
}

type Dispatcher struct {
	repo     Repository
	interval time.Duration
	handler  func(ctx context.Context, event Event) error
}

func NewDispatcher(repo Repository, interval time.Duration, handler func(ctx context.Context, event Event) error) *Dispatcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Dispatcher{
		repo:     repo,
		interval: interval,
		handler:  handler,
	}
}

func (d *Dispatcher) Run(ctx context.Context) error {
	if d == nil || d.repo == nil || d.handler == nil {
		return nil
	}

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		if err := d.DispatchOnce(ctx); err != nil {
			logging.Error(ctx, "outbox_dispatch_failed", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (d *Dispatcher) DispatchOnce(ctx context.Context) error {
	eventsToDispatch, err := d.repo.ListPendingOutboxEvents(ctx, 100)
	if err != nil {
		return err
	}

	for _, pending := range eventsToDispatch {
		if err := d.handler(ctx, pending); err != nil {
			_ = d.repo.MarkOutboxEventFailed(ctx, pending.EventID, err.Error())
			continue
		}
		if err := d.repo.MarkOutboxEventPublished(ctx, pending.EventID, time.Now().UTC()); err != nil {
			return err
		}
	}

	return nil
}
