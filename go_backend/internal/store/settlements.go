package store

import (
	"context"

	"iwx/go_backend/internal/domain"
)

type SettlementRepository interface {
	ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error)
	ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error)
	ReplaceSettlementEntriesProjection(ctx context.Context, contractID int64, entries []domain.SettlementEntry) error
}

type SettlementProjectionSource interface {
	ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error)
}

type SettlementProjectionTarget interface {
	ReplaceSettlementEntriesProjection(ctx context.Context, contractID int64, entries []domain.SettlementEntry) error
}
