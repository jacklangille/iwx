package store

import (
	"context"

	"iwx/go_backend/internal/domain"
)

type UserViewRepository interface {
	ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error)
	ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error)
	ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error)
	ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error)
	ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error)
	ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error)
}

type UserProjectionSource interface {
	ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error)
	ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error)
	ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error)
	ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error)
	ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error)
	ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error)
}

type UserProjectionTarget interface {
	ReplaceCashAccountsProjection(ctx context.Context, userID int64, accounts []domain.CashAccount) error
	ReplacePositionsProjection(ctx context.Context, userID int64, positions []domain.Position) error
	ReplacePositionLocksProjection(ctx context.Context, userID int64, locks []domain.PositionLock) error
	ReplaceCollateralLocksProjection(ctx context.Context, userID int64, currency string, locks []domain.CollateralLock) error
	ReplaceOrderCashReservationsProjection(ctx context.Context, userID int64, currency string, reservations []domain.OrderCashReservation) error
	ReplaceUserSettlementEntriesProjection(ctx context.Context, userID int64, entries []domain.SettlementEntry) error
}
