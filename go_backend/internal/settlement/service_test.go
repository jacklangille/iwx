package settlement

import (
	"context"
	"testing"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/store"
)

func TestHandleContractResolvedUsesTraceIDAsCorrelationID(t *testing.T) {
	t.Parallel()

	settledAt := time.Date(2026, 5, 7, 5, 0, 0, 0, time.UTC)
	var captured store.SettleContractInput
	publisher := &stubSettlementPublisher{}
	service := NewService(stubSettlementRepository{
		settleContractFn: func(_ context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
			captured = input
			return &store.SettlementResult{SettledAt: settledAt}, nil
		},
	}, nil, publisher)

	_, err := service.HandleContractResolved(context.Background(), events.ContractResolved{
		ContractID: 17,
		Outcome:    "above",
		TraceID:    "trace-17",
		ResolvedAt: settledAt,
	})
	if err != nil {
		t.Fatalf("HandleContractResolved() error = %v", err)
	}
	if captured.CorrelationID != "trace-17" {
		t.Fatalf("expected correlation id trace-17, got %q", captured.CorrelationID)
	}
	if publisher.published == nil || publisher.published.TraceID != "trace-17" {
		t.Fatalf("expected settlement completed trace id to be preserved, got %#v", publisher.published)
	}
}

func TestHandleContractResolvedFallsBackToGeneratedCorrelationID(t *testing.T) {
	t.Parallel()

	resolvedAt := time.Date(2026, 5, 7, 5, 1, 2, 0, time.UTC)
	var captured store.SettleContractInput
	service := NewService(stubSettlementRepository{
		settleContractFn: func(_ context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
			captured = input
			return &store.SettlementResult{SettledAt: resolvedAt}, nil
		},
	}, nil, nil)

	_, err := service.HandleContractResolved(context.Background(), events.ContractResolved{
		ContractID: 18,
		Outcome:    "below",
		ResolvedAt: resolvedAt,
	})
	if err != nil {
		t.Fatalf("HandleContractResolved() error = %v", err)
	}
	expected := "contract-resolved-20260507050102-below"
	if captured.CorrelationID != expected {
		t.Fatalf("expected fallback correlation id %q, got %q", expected, captured.CorrelationID)
	}
}

type stubSettlementRepository struct {
	settleContractFn func(context.Context, store.SettleContractInput) (*store.SettlementResult, error)
}

func (s stubSettlementRepository) ProcessCreateContract(context.Context, commands.CreateContractEnvelope) (commands.CreateContractResult, error) {
	return commands.CreateContractResult{}, nil
}
func (s stubSettlementRepository) GetContractCommand(context.Context, string) (*commands.ContractCommand, error) {
	return nil, nil
}
func (s stubSettlementRepository) GetContract(context.Context, int64) (*domain.Contract, error) {
	return nil, nil
}
func (s stubSettlementRepository) GetContractRule(context.Context, int64) (*domain.ContractRule, error) {
	return nil, nil
}
func (s stubSettlementRepository) UpdateContractStatus(context.Context, int64, string) (*domain.Contract, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListContractCollateralLocks(context.Context, int64, int64) ([]domain.CollateralLock, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListIssuanceBatches(context.Context, int64, int64) ([]domain.IssuanceBatch, error) {
	return nil, nil
}
func (s stubSettlementRepository) CreateIssuanceBatch(context.Context, store.CreateIssuanceBatchInput) (*domain.IssuanceBatch, *domain.CollateralLock, []*domain.Position, error) {
	return nil, nil, nil, nil
}
func (s stubSettlementRepository) ListPositions(context.Context, int64, *int64) ([]domain.Position, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListPositionLocks(context.Context, int64, *int64) ([]domain.PositionLock, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListSettlementEntriesByContract(context.Context, int64, int) ([]domain.SettlementEntry, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListSettlementEntriesByUser(context.Context, int64, *int64, int) ([]domain.SettlementEntry, error) {
	return nil, nil
}
func (s stubSettlementRepository) CreatePositionLock(context.Context, store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	return nil, nil, nil
}
func (s stubSettlementRepository) ReleasePositionLock(context.Context, store.ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	return nil, nil, nil
}
func (s stubSettlementRepository) GetCashAccount(context.Context, int64, string) (*domain.CashAccount, error) {
	return nil, nil
}
func (s stubSettlementRepository) ListLedgerEntries(context.Context, int64, string, int) ([]domain.LedgerEntry, error) {
	return nil, nil
}
func (s stubSettlementRepository) DepositCash(context.Context, store.DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil
}
func (s stubSettlementRepository) WithdrawCash(context.Context, store.WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil
}
func (s stubSettlementRepository) ListCollateralLocks(context.Context, int64, string) ([]domain.CollateralLock, error) {
	return nil, nil
}
func (s stubSettlementRepository) CreateCollateralLock(context.Context, store.CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil, nil
}
func (s stubSettlementRepository) ReleaseCollateralLock(context.Context, store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil, nil
}
func (s stubSettlementRepository) ListOrderCashReservations(context.Context, int64, string) ([]domain.OrderCashReservation, error) {
	return nil, nil
}
func (s stubSettlementRepository) CreateOrderCashReservation(context.Context, store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil, nil
}
func (s stubSettlementRepository) ReleaseOrderCashReservation(context.Context, store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	return nil, nil, nil, nil
}
func (s stubSettlementRepository) ApplyExecution(context.Context, events.ExecutionCreated) (*store.ExecutionApplicationResult, error) {
	return nil, nil
}
func (s stubSettlementRepository) SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
	return s.settleContractFn(ctx, input)
}

type stubSettlementPublisher struct {
	published *events.SettlementCompleted
}

func (s *stubSettlementPublisher) PublishSettlementCompleted(_ context.Context, event events.SettlementCompleted) error {
	copy := event
	s.published = &copy
	return nil
}

func (s *stubSettlementPublisher) Close() {}
