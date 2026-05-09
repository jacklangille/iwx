package projectionchange

import (
	"context"
	"fmt"
	"time"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/requestctx"
)

type Publisher interface {
	PublishProjectionChange(ctx context.Context, event events.ProjectionChange) error
	Close()
}

type Emitter struct {
	publisher Publisher
}

func NewEmitter(publisher Publisher) *Emitter {
	return &Emitter{publisher: publisher}
}

func (e *Emitter) EmitContractChanged(ctx context.Context, contractID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeContract, fmt.Sprintf("contract:%d", contractID), versionTime),
		Kind:       events.ProjectionChangeContract,
		TraceID:    requestctx.TraceID(ctx),
		ContractID: contractID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitUserStateChanged(ctx context.Context, userID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeUserState, fmt.Sprintf("user:%d", userID), versionTime),
		Kind:       events.ProjectionChangeUserState,
		TraceID:    requestctx.TraceID(ctx),
		UserID:     userID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitSettlementChanged(ctx context.Context, contractID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeSettlement, fmt.Sprintf("settlement:%d", contractID), versionTime),
		Kind:       events.ProjectionChangeSettlement,
		TraceID:    requestctx.TraceID(ctx),
		ContractID: contractID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitMarketChanged(ctx context.Context, contractID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeMarket, fmt.Sprintf("market:%d", contractID), versionTime),
		Kind:       events.ProjectionChangeMarket,
		TraceID:    requestctx.TraceID(ctx),
		ContractID: contractID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitOrderCommandChanged(ctx context.Context, commandID string, contractID, userID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeOrderCommand, "order_command:"+commandID, versionTime),
		Kind:       events.ProjectionChangeOrderCommand,
		TraceID:    requestctx.TraceID(ctx),
		ContractID: contractID,
		UserID:     userID,
		CommandID:  commandID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitOracleStateChanged(ctx context.Context, contractID int64, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeOracleState, fmt.Sprintf("oracle:%d", contractID), versionTime),
		Kind:       events.ProjectionChangeOracleState,
		TraceID:    requestctx.TraceID(ctx),
		ContractID: contractID,
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) EmitStationCatalogChanged(ctx context.Context, versionTime time.Time) error {
	return e.emit(ctx, events.ProjectionChange{
		EventID:    eventID(events.ProjectionChangeStationCatalog, "stations", versionTime),
		Kind:       events.ProjectionChangeStationCatalog,
		TraceID:    requestctx.TraceID(ctx),
		Version:    version(versionTime),
		OccurredAt: occurredAt(versionTime),
	})
}

func (e *Emitter) emit(ctx context.Context, event events.ProjectionChange) error {
	if e == nil || e.publisher == nil {
		return nil
	}
	return e.publisher.PublishProjectionChange(ctx, event)
}

func version(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UTC().UnixNano()
}

func occurredAt(t time.Time) time.Time {
	if t.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return t.UTC()
}

func eventID(kind events.ProjectionChangeKind, key string, t time.Time) string {
	return fmt.Sprintf("projection-change:%s:%s:%d", kind, key, version(t))
}
