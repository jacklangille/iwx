package domain

import "testing"

func TestIsValidContractState(t *testing.T) {
	valid := []string{
		"draft",
		"pending_approval",
		"pending_collateral",
		"active",
		"trading_closed",
		"awaiting_resolution",
		"resolved",
		"settled",
		"cancelled",
	}

	for _, state := range valid {
		if !IsValidContractState(state) {
			t.Fatalf("expected valid contract state: %s", state)
		}
	}

	if IsValidContractState("open") {
		t.Fatal("expected legacy state 'open' to be invalid")
	}
}
