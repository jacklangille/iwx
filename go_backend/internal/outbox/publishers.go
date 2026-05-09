package outbox

import (
	"context"
	"encoding/json"

	"iwx/go_backend/internal/events"
)

type ExecutionCreatedPublisher struct {
	repo Repository
}

func NewExecutionCreatedPublisher(repo Repository) *ExecutionCreatedPublisher {
	return &ExecutionCreatedPublisher{repo: repo}
}

func (p *ExecutionCreatedPublisher) PublishExecutionCreated(ctx context.Context, event events.ExecutionCreated) error {
	if p == nil || p.repo == nil {
		return nil
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.repo.EnqueueOutboxEvent(ctx, Event{
		EventID:   event.ExecutionID,
		EventType: EventTypeExecutionCreated,
		Payload:   body,
	})
}

func (p *ExecutionCreatedPublisher) Close() {}

type ProjectionChangePublisher struct {
	repo Repository
}

func NewProjectionChangePublisher(repo Repository) *ProjectionChangePublisher {
	return &ProjectionChangePublisher{repo: repo}
}

func (p *ProjectionChangePublisher) PublishProjectionChange(ctx context.Context, event events.ProjectionChange) error {
	if p == nil || p.repo == nil {
		return nil
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.repo.EnqueueOutboxEvent(ctx, Event{
		EventID:   event.EventID,
		EventType: EventTypeProjectionChange,
		Payload:   body,
	})
}

func (p *ProjectionChangePublisher) Close() {}
