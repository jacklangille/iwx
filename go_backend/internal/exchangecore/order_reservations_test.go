package exchangecore

import (
	"context"
	"testing"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

func TestReserveOrderForMatchingBidCreatesCashReservation(t *testing.T) {
	t.Parallel()

	var captured store.CreateOrderCashReservationInput
	service := &Service{
		repo: stubExchangeCoreRepository{
			createOrderCashReservationFn: func(_ context.Context, input store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
				captured = input
				return &domain.OrderCashReservation{ID: 15, AmountCents: input.AmountCents}, nil, nil, nil
			},
		},
	}

	reservation, err := service.ReserveOrderForMatching(context.Background(), commands.PlaceOrder{
		UserID:     3,
		ContractID: 4,
		TokenType:  string(domain.ClaimSideAbove),
		OrderSide:  string(domain.OrderSideBid),
		Price:      "55.00",
		Quantity:   10,
	}, "cmd-1")
	if err != nil {
		t.Fatalf("ReserveOrderForMatching() error = %v", err)
	}
	if reservation.CashReservation == nil {
		t.Fatal("expected cash reservation")
	}
	if captured.AmountCents != 55000 {
		t.Fatalf("expected amount cents 55000, got %d", captured.AmountCents)
	}
	if captured.ReferenceID != "cmd-1" || captured.CorrelationID != "cmd-1" {
		t.Fatalf("expected command correlation to be preserved, got %+v", captured)
	}
}

func TestReserveOrderForMatchingAskCreatesPositionLock(t *testing.T) {
	t.Parallel()

	var captured store.CreatePositionLockInput
	service := &Service{
		repo: stubExchangeCoreRepository{
			createPositionLockFn: func(_ context.Context, input store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
				captured = input
				return &domain.PositionLock{ID: 21, Side: domain.ClaimSide(input.Side), Quantity: input.Quantity}, nil, nil
			},
		},
	}

	reservation, err := service.ReserveOrderForMatching(context.Background(), commands.PlaceOrder{
		UserID:     5,
		ContractID: 6,
		TokenType:  string(domain.ClaimSideBelow),
		OrderSide:  string(domain.OrderSideAsk),
		Price:      "42.00",
		Quantity:   7,
	}, "cmd-ask")
	if err != nil {
		t.Fatalf("ReserveOrderForMatching() error = %v", err)
	}
	if reservation.PositionLock == nil {
		t.Fatal("expected position lock")
	}
	if captured.Side != string(domain.ClaimSideBelow) {
		t.Fatalf("expected below side, got %q", captured.Side)
	}
	if captured.ReferenceID != "cmd-ask" {
		t.Fatalf("expected reference id cmd-ask, got %q", captured.ReferenceID)
	}
}

func TestReleaseOrderReservationReleasesCashReservation(t *testing.T) {
	t.Parallel()

	var captured store.ReleaseOrderCashReservationInput
	service := &Service{
		repo: stubExchangeCoreRepository{
			releaseOrderCashReservationFn: func(_ context.Context, input store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
				captured = input
				return &domain.OrderCashReservation{ID: input.ReservationID}, nil, nil, nil
			},
		},
	}

	err := service.ReleaseOrderReservation(context.Background(), 44, &OrderReservation{
		CashReservation: &domain.OrderCashReservation{ID: 99},
	}, "enqueue-failed")
	if err != nil {
		t.Fatalf("ReleaseOrderReservation() error = %v", err)
	}
	if captured.ReservationID != 99 || captured.CorrelationID != "enqueue-failed" {
		t.Fatalf("unexpected release input: %+v", captured)
	}
}
