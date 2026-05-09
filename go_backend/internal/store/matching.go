package store

import (
	"context"

	"iwx/go_backend/internal/commands"
)

type MatchingRepository interface {
	ProcessPlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error)
}

type OrderCommandRepository interface {
	GetOrderCommand(ctx context.Context, commandID string) (*commands.OrderCommand, error)
}

type OrderCommandProjectionTarget interface {
	UpsertOrderCommandProjection(ctx context.Context, command commands.OrderCommand) error
}
