package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

var errIssuanceQuantityMismatch = errors.New("issuance quantity does not match collateral lock")

const defaultIssuanceCollateralPerPairCents int64 = 100

func (r *ContractRepository) ListIssuanceBatches(ctx context.Context, userID, contractID int64) ([]domain.IssuanceBatch, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			creator_user_id,
			collateral_lock_id,
			above_quantity,
			below_quantity,
			status,
			created_at,
			updated_at
		FROM issuance_batches
		WHERE creator_user_id = $1 AND contract_id = $2
		ORDER BY id DESC
	`, userID, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := []domain.IssuanceBatch{}
	for rows.Next() {
		batch, err := scanIssuanceBatch(rows)
		if err != nil {
			return nil, err
		}

		batches = append(batches, batch)
	}

	return batches, rows.Err()
}

func (r *ContractRepository) ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error) {
	query := `
		SELECT
			id,
			user_id,
			contract_id,
			side,
			available_quantity,
			locked_quantity,
			total_quantity,
			updated_at
		FROM positions
		WHERE user_id = $1
	`
	args := []any{userID}
	if contractID != nil {
		query += ` AND contract_id = $2`
		args = append(args, *contractID)
	}
	query += ` ORDER BY contract_id ASC, side ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []domain.Position{}
	for rows.Next() {
		position, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}

		positions = append(positions, position)
	}

	return positions, rows.Err()
}

func (r *ContractRepository) CreateIssuanceBatch(ctx context.Context, input store.CreateIssuanceBatchInput) (*domain.IssuanceBatch, *domain.CollateralLock, []*domain.Position, error) {
	type result struct {
		batch     *domain.IssuanceBatch
		lock      *domain.CollateralLock
		positions []*domain.Position
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		lock, err := getCollateralLockForUpdateTx(ctx, tx, input.CollateralLockID)
		if err != nil {
			return result{}, err
		}
		if lock == nil {
			return result{}, errCollateralLockNotFound
		}
		if lock.UserID != input.UserID || lock.ContractID != input.ContractID {
			return result{}, errCollateralLockNotFound
		}
		if lock.Status != domain.CollateralLockStatusLocked {
			return result{}, errCollateralLockInactive
		}

		contract, err := getContractForUpdateTx(ctx, tx, input.ContractID)
		if err != nil {
			return result{}, err
		}
		if contract == nil {
			return result{}, fmt.Errorf("contract not found")
		}

		expectedAmount := defaultIssuanceCollateralPerPairCents * input.PairedQuantity
		if contract.Multiplier != nil && *contract.Multiplier > 0 {
			expectedAmount = *contract.Multiplier * input.PairedQuantity
		}
		if lock.AmountCents != expectedAmount {
			return result{}, errIssuanceQuantityMismatch
		}

		batch, err := insertIssuanceBatchTx(ctx, tx, input)
		if err != nil {
			return result{}, err
		}

		if err := consumeCollateralLockTx(ctx, tx, lock.ID, batch.ID); err != nil {
			return result{}, err
		}

		abovePosition, err := upsertPositionQuantityTx(ctx, tx, input.UserID, input.ContractID, domain.ClaimSideAbove, input.PairedQuantity)
		if err != nil {
			return result{}, err
		}
		belowPosition, err := upsertPositionQuantityTx(ctx, tx, input.UserID, input.ContractID, domain.ClaimSideBelow, input.PairedQuantity)
		if err != nil {
			return result{}, err
		}

		lock, err = getCollateralLockForUpdateTx(ctx, tx, lock.ID)
		if err != nil {
			return result{}, err
		}

		return result{
			batch:     batch,
			lock:      lock,
			positions: []*domain.Position{abovePosition, belowPosition},
		}, nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return outcome.batch, outcome.lock, outcome.positions, nil
}

func insertIssuanceBatchTx(ctx context.Context, tx *sql.Tx, input store.CreateIssuanceBatchInput) (*domain.IssuanceBatch, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO issuance_batches (
			contract_id,
			creator_user_id,
			collateral_lock_id,
			above_quantity,
			below_quantity,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'issued', NOW(), NOW())
		RETURNING
			id,
			contract_id,
			creator_user_id,
			collateral_lock_id,
			above_quantity,
			below_quantity,
			status,
			created_at,
			updated_at
	`, input.ContractID, input.UserID, input.CollateralLockID, input.PairedQuantity, input.PairedQuantity)

	batch, err := scanIssuanceBatch(row)
	if err != nil {
		return nil, err
	}

	return &batch, nil
}

func consumeCollateralLockTx(ctx context.Context, tx *sql.Tx, lockID, issuanceBatchID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE collateral_locks
		SET
			status = 'consumed',
			reference_issuance_id = $2,
			updated_at = NOW()
		WHERE id = $1
	`, lockID, issuanceBatchID)
	return err
}

func upsertPositionQuantityTx(ctx context.Context, tx *sql.Tx, userID, contractID int64, side domain.ClaimSide, deltaQuantity int64) (*domain.Position, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO positions (
			user_id,
			contract_id,
			side,
			available_quantity,
			locked_quantity,
			total_quantity,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 0, $4, NOW())
		ON CONFLICT (user_id, contract_id, side) DO UPDATE
		SET
			available_quantity = positions.available_quantity + EXCLUDED.available_quantity,
			total_quantity = positions.total_quantity + EXCLUDED.total_quantity,
			updated_at = NOW()
		RETURNING
			id,
			user_id,
			contract_id,
			side,
			available_quantity,
			locked_quantity,
			total_quantity,
			updated_at
	`, userID, contractID, string(side), deltaQuantity)

	position, err := scanPosition(row)
	if err != nil {
		return nil, err
	}

	return &position, nil
}

func getContractForUpdateTx(ctx context.Context, tx *sql.Tx, contractID int64) (*domain.Contract, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			updated_at
		FROM contracts
		WHERE id = $1
		FOR UPDATE
	`, contractID)

	contract, err := scanContract(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &contract, nil
}
