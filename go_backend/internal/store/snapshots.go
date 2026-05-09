package store

import (
	"context"
	"time"

	"iwx/go_backend/internal/domain"
)

type SnapshotRepository interface {
	ListSnapshotsSince(ctx context.Context, contractID int64, windowStart time.Time) ([]domain.MarketSnapshot, error)
	LatestTimestamp(ctx context.Context, contractID int64) (*time.Time, error)
	LatestSequence(ctx context.Context, contractID int64) (*int64, error)
}

type SnapshotProjectionSource interface {
	ListAllSnapshots(ctx context.Context, contractID int64) ([]domain.MarketSnapshot, error)
}

type SnapshotProjectionTarget interface {
	ReplaceSnapshotsProjection(ctx context.Context, contractID int64, snapshots []domain.MarketSnapshot) error
}
