package oracle

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/internal/store"
)

func TestResolveContractPublishesTraceIDAndOutcome(t *testing.T) {
	t.Parallel()

	threshold := int64(30)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	observedAt := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	repo := stubOracleRepository{
		listObservationsFn: func(context.Context, int64, int) ([]domain.OracleObservation, error) {
			return []domain.OracleObservation{{
				ContractID:             1,
				ProviderName:           "NOAA",
				StationID:              "HALI",
				ObservedMetric:         "temperature_max",
				ObservationWindowStart: start,
				ObservationWindowEnd:   end,
				NormalizedValue:        "31.5",
				ObservedAt:             observedAt,
			}}, nil
		},
		insertResolutionFn: func(_ context.Context, input domain.ContractResolution) (*domain.ContractResolution, error) {
			input.ID = 22
			input.ResolvedAt = observedAt
			return &input, nil
		},
	}
	var enqueued *events.ContractResolved
	repo.enqueueOutboxFn = func(_ context.Context, event outbox.Event) error {
		var payload events.ContractResolved
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		enqueued = &payload
		return nil
	}

	contracts := stubContractRepository{
		getContractFn: func(context.Context, int64) (*domain.Contract, error) {
			return &domain.Contract{
				ID:                     1,
				MeasurementPeriodStart: &start,
				MeasurementPeriodEnd:   &end,
			}, nil
		},
		getContractRuleFn: func(context.Context, int64) (*domain.ContractRule, error) {
			return &domain.ContractRule{
				ContractID:              1,
				Threshold:               &threshold,
				ResolutionInclusiveSide: domain.ClaimSideBelow,
				RuleVersion:             "v1",
			}, nil
		},
		updateContractStatusFn: func(_ context.Context, contractID int64, status string) (*domain.Contract, error) {
			return &domain.Contract{ID: contractID, Status: status}, nil
		},
	}

	service := NewService(repo, contracts, nil, &stubOracleEventPublisher{})

	ctx := requestctx.WithTraceID(context.Background(), "trace-123")
	resolution, err := service.ResolveContract(ctx, 1)
	if err != nil {
		t.Fatalf("ResolveContract() error = %v", err)
	}
	if resolution.Outcome != domain.ResolutionOutcomeAbove {
		t.Fatalf("expected above outcome, got %q", resolution.Outcome)
	}
	if enqueued == nil {
		t.Fatal("expected enqueued outbox event")
	}
	if enqueued.TraceID != "trace-123" {
		t.Fatalf("expected trace id trace-123, got %q", enqueued.TraceID)
	}
	if enqueued.Outcome != string(domain.ResolutionOutcomeAbove) {
		t.Fatalf("expected enqueued above outcome, got %q", enqueued.Outcome)
	}
}

type stubOracleRepository struct {
	upsertStationFn       func(context.Context, store.UpsertStationInput) (*domain.WeatherStation, error)
	listStationsFn        func(context.Context, bool) ([]domain.WeatherStation, error)
	findStationFn         func(context.Context, string, string) (*domain.WeatherStation, error)
	recordObservationFn   func(context.Context, store.RecordObservationInput) (*domain.OracleObservation, error)
	listObservationsFn    func(context.Context, int64, int) ([]domain.OracleObservation, error)
	getLatestResolutionFn func(context.Context, int64) (*domain.ContractResolution, error)
	insertResolutionFn    func(context.Context, domain.ContractResolution) (*domain.ContractResolution, error)
	enqueueOutboxFn       func(context.Context, outbox.Event) error
	resolveContractRepoFn func(context.Context, store.ResolveContractInput) (*domain.ContractResolution, error)
}

func (s stubOracleRepository) UpsertStation(ctx context.Context, input store.UpsertStationInput) (*domain.WeatherStation, error) {
	if s.upsertStationFn != nil {
		return s.upsertStationFn(ctx, input)
	}
	return nil, nil
}

func (s stubOracleRepository) ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error) {
	if s.listStationsFn != nil {
		return s.listStationsFn(ctx, activeOnly)
	}
	return nil, nil
}

func (s stubOracleRepository) FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
	if s.findStationFn != nil {
		return s.findStationFn(ctx, providerName, stationID)
	}
	return nil, nil
}

func (s stubOracleRepository) RecordObservation(ctx context.Context, input store.RecordObservationInput) (*domain.OracleObservation, error) {
	if s.recordObservationFn != nil {
		return s.recordObservationFn(ctx, input)
	}
	return nil, nil
}

func (s stubOracleRepository) ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error) {
	if s.listObservationsFn != nil {
		return s.listObservationsFn(ctx, contractID, limit)
	}
	return nil, nil
}

func (s stubOracleRepository) GetLatestResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error) {
	if s.getLatestResolutionFn != nil {
		return s.getLatestResolutionFn(ctx, contractID)
	}
	return nil, nil
}

func (s stubOracleRepository) InsertResolution(ctx context.Context, input domain.ContractResolution) (*domain.ContractResolution, error) {
	if s.insertResolutionFn != nil {
		return s.insertResolutionFn(ctx, input)
	}
	return &input, nil
}

func (s stubOracleRepository) MarkResolutionPublished(context.Context, string, time.Time) error {
	return nil
}

func (s stubOracleRepository) EnqueueOutboxEvent(ctx context.Context, event outbox.Event) error {
	if s.enqueueOutboxFn != nil {
		return s.enqueueOutboxFn(ctx, event)
	}
	return nil
}

func (s stubOracleRepository) ListPendingOutboxEvents(context.Context, int) ([]outbox.Event, error) {
	return nil, nil
}

func (s stubOracleRepository) MarkOutboxEventPublished(context.Context, string, time.Time) error {
	return nil
}

func (s stubOracleRepository) MarkOutboxEventFailed(context.Context, string, string) error {
	return nil
}

func (s stubOracleRepository) ResolveContract(ctx context.Context, input store.ResolveContractInput) (*domain.ContractResolution, error) {
	if s.resolveContractRepoFn != nil {
		return s.resolveContractRepoFn(ctx, input)
	}
	return nil, nil
}

type stubContractRepository struct {
	getContractFn          func(context.Context, int64) (*domain.Contract, error)
	getContractRuleFn      func(context.Context, int64) (*domain.ContractRule, error)
	updateContractStatusFn func(context.Context, int64, string) (*domain.Contract, error)
}

func (s stubContractRepository) GetContract(ctx context.Context, contractID int64) (*domain.Contract, error) {
	return s.getContractFn(ctx, contractID)
}

func (s stubContractRepository) GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error) {
	return s.getContractRuleFn(ctx, contractID)
}

func (s stubContractRepository) UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error) {
	return s.updateContractStatusFn(ctx, contractID, status)
}

type stubOracleEventPublisher struct {
	published *events.ContractResolved
}

func (s *stubOracleEventPublisher) PublishContractResolved(_ context.Context, event events.ContractResolved) error {
	copy := event
	s.published = &copy
	return nil
}

func (s *stubOracleEventPublisher) Close() {}
