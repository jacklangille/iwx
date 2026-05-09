package matching

import (
	"testing"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

func TestValidatePlaceOrderRequiresCashReservationForBid(t *testing.T) {
	t.Parallel()

	err := ValidatePlaceOrder(commands.PlaceOrder{
		ContractID: 1,
		UserID:     2,
		TokenType:  string(domain.ClaimSideAbove),
		OrderSide:  string(domain.OrderSideBid),
		Price:      "10.00",
		Quantity:   1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Errors["cash_reservation_id"]) == 0 {
		t.Fatalf("expected cash_reservation_id validation error, got %#v", validationErr.Errors)
	}
}

func TestValidatePlaceOrderRequiresPositionLockForAsk(t *testing.T) {
	t.Parallel()

	err := ValidatePlaceOrder(commands.PlaceOrder{
		ContractID: 1,
		UserID:     2,
		TokenType:  string(domain.ClaimSideBelow),
		OrderSide:  string(domain.OrderSideAsk),
		Price:      "10.00",
		Quantity:   1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Errors["position_lock_id"]) == 0 {
		t.Fatalf("expected position_lock_id validation error, got %#v", validationErr.Errors)
	}
}
