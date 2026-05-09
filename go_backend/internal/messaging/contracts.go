package messaging

import "iwx/go_backend/internal/commands"

type PlaceOrderRequest struct {
	Envelope commands.PlaceOrderEnvelope `json:"envelope"`
}
