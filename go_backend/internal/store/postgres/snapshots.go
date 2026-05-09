package postgres

import (
	"context"
	"database/sql"
	"time"

	"iwx/go_backend/internal/domain"
)

type SnapshotRepository struct {
	*baseRepository
}

func NewSnapshotRepository(databaseURL string) *SnapshotRepository {
	return &SnapshotRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *SnapshotRepository) ListSnapshotsSince(
	ctx context.Context,
	contractID int64,
	windowStart time.Time,
) ([]domain.MarketSnapshot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			best_above::text,
			best_below::text,
			mid_above::text,
			mid_below::text,
			inserted_at
		FROM market_snapshots
		WHERE contract_id = $1 AND inserted_at >= $2
		ORDER BY inserted_at ASC
	`, contractID, windowStart.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := []domain.MarketSnapshot{}
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

func (r *SnapshotRepository) ListAllSnapshots(ctx context.Context, contractID int64) ([]domain.MarketSnapshot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			best_above::text,
			best_below::text,
			mid_above::text,
			mid_below::text,
			inserted_at
		FROM market_snapshots
		WHERE contract_id = $1
		ORDER BY inserted_at ASC
	`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := []domain.MarketSnapshot{}
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

func (r *SnapshotRepository) LatestTimestamp(ctx context.Context, contractID int64) (*time.Time, error) {
	var insertedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT inserted_at
		FROM market_snapshots
		WHERE contract_id = $1
		ORDER BY inserted_at DESC
		LIMIT 1
	`, contractID).Scan(&insertedAt)
	if err == sql.ErrNoRows || !insertedAt.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	value := insertedAt.Time.UTC()
	return &value, nil
}

func (r *SnapshotRepository) LatestSequence(ctx context.Context, contractID int64) (*int64, error) {
	var sequence sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM market_snapshots
		WHERE contract_id = $1
		ORDER BY inserted_at DESC
		LIMIT 1
	`, contractID).Scan(&sequence)
	if err == sql.ErrNoRows || !sequence.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	value := sequence.Int64
	return &value, nil
}

func (r *SnapshotRepository) ReplaceSnapshotsProjection(ctx context.Context, contractID int64, snapshots []domain.MarketSnapshot) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM market_snapshots WHERE contract_id = $1`, contractID); err != nil {
			return struct{}{}, err
		}

		for _, snapshot := range snapshots {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO market_snapshots (
					id,
					contract_id,
					best_above,
					best_below,
					mid_above,
					mid_below,
					inserted_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`,
				snapshot.ID,
				snapshot.ContractID,
				snapshot.BestAbove,
				snapshot.BestBelow,
				snapshot.MidAbove,
				snapshot.MidBelow,
				snapshot.InsertedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}
