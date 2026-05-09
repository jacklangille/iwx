package exchangecore

import (
	"context"
	"testing"

	"iwx/go_backend/internal/store"
)

func TestDepositCashRejectsInvalidInput(t *testing.T) {
	service := &Service{}

	_, _, err := service.DepositCash(context.Background(), store.DepositCashInput{
		UserID:      0,
		Currency:    "",
		AmountCents: 0,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(validationErr.Errors["user_id"]) == 0 {
		t.Fatalf("expected user_id validation error, got %#v", validationErr.Errors)
	}
	if len(validationErr.Errors["amount_cents"]) == 0 {
		t.Fatalf("expected amount_cents validation error, got %#v", validationErr.Errors)
	}
}

func TestCreateOrderCashReservationRejectsMissingReferences(t *testing.T) {
	service := &Service{}

	_, _, _, err := service.CreateOrderCashReservation(context.Background(), store.CreateOrderCashReservationInput{
		UserID:      1,
		ContractID:  1,
		Currency:    "usd",
		AmountCents: 100,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(validationErr.Errors["reference_type"]) == 0 {
		t.Fatalf("expected reference_type validation error, got %#v", validationErr.Errors)
	}
	if len(validationErr.Errors["reference_id"]) == 0 {
		t.Fatalf("expected reference_id validation error, got %#v", validationErr.Errors)
	}
}

func TestNormalizeCurrencyDefaultsToUSD(t *testing.T) {
	if actual := normalizeCurrency(""); actual != DefaultAccountCurrency {
		t.Fatalf("expected default currency %q, got %q", DefaultAccountCurrency, actual)
	}

	if actual := normalizeCurrency(" cad "); actual != "CAD" {
		t.Fatalf("expected normalized currency CAD, got %q", actual)
	}
}
