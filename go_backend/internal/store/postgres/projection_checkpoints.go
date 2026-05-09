package postgres

import (
	"context"
	"database/sql"
)

type ProjectionCheckpointRepository struct {
	*baseRepository
}

func NewProjectionCheckpointRepository(databaseURL string) *ProjectionCheckpointRepository {
	return &ProjectionCheckpointRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *ProjectionCheckpointRepository) ShouldApply(ctx context.Context, key, eventID string, version int64) (bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT event_id, version
		FROM projection_checkpoints
		WHERE projection_key = $1
	`, key)

	var currentEventID string
	var currentVersion int64
	switch err := row.Scan(&currentEventID, &currentVersion); err {
	case nil:
		if currentVersion > version {
			return false, nil
		}
		if currentVersion == version {
			return false, nil
		}
		return true, nil
	case sql.ErrNoRows:
		return true, nil
	default:
		return false, err
	}
}

func (r *ProjectionCheckpointRepository) RecordApplied(ctx context.Context, key, eventID string, version int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projection_checkpoints (
			projection_key,
			event_id,
			version,
			updated_at
		)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (projection_key) DO UPDATE
		SET
			event_id = EXCLUDED.event_id,
			version = EXCLUDED.version,
			updated_at = NOW()
		WHERE projection_checkpoints.version <= EXCLUDED.version
	`, key, eventID, version)
	return err
}
