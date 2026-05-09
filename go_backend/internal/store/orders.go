package store

import (
	"context"

	"iwx/go_backend/internal/domain"
)

type OrderRepository interface {
	ListOpenOrders(ctx context.Context, contractID *int64) ([]domain.Order, error)
}

type OrderProjectionTarget interface {
	ReplaceOpenOrdersProjection(ctx context.Context, contractID int64, orders []domain.Order) error
}
