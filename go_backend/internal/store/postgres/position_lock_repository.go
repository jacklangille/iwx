package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

var (
	errPositionLockNotFound = errors.New("position lock not found")
	errPositionLockInactive = errors.New("position lock is not active")
	errInsufficientPosition = errors.New("insufficient available position")
)

func (r *ContractRepository) ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error) {
	query := `
		SELECT
			id,
			user_id,
			contract_id,
			side,
			quantity,
			status,
			order_id,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at,
			released_at
		FROM position_locks
		WHERE user_id = $1
	`
	args := []any{userID}
	if contractID != nil {
		query += ` AND contract_id = $2`
		args = append(args, *contractID)
	}
	query += ` ORDER BY id DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locks := []domain.PositionLock{}
	for rows.Next() {
		lock, err := scanPositionLock(rows)
		if err != nil {
			return nil, err
		}

		locks = append(locks, lock)
	}

	return locks, rows.Err()
}

func (r *ContractRepository) CreatePositionLock(ctx context.Context, input store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	type result struct {
		lock     *domain.PositionLock
		position *domain.Position
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		position, err := getPositionForUpdateTx(ctx, tx, input.UserID, input.ContractID, input.Side)
		if err != nil {
			return result{}, err
		}
		if position == nil || position.AvailableQuantity < input.Quantity {
			return result{}, errInsufficientPosition
		}

		if err := updatePositionBalancesTx(ctx, tx, position.ID, position.AvailableQuantity-input.Quantity, position.LockedQuantity+input.Quantity); err != nil {
			return result{}, err
		}

		lock, err := insertPositionLockTx(ctx, tx, input)
		if err != nil {
			return result{}, err
		}

		position, err = getPositionByIDTx(ctx, tx, position.ID)
		if err != nil {
			return result{}, err
		}

		return result{lock: lock, position: position}, validatePosition(*position)
	})
	if err != nil {
		return nil, nil, err
	}

	return outcome.lock, outcome.position, nil
}

func (r *ContractRepository) ReleasePositionLock(ctx context.Context, input store.ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	type result struct {
		lock     *domain.PositionLock
		position *domain.Position
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		lock, err := getPositionLockForUpdateTx(ctx, tx, input.LockID)
		if err != nil {
			return result{}, err
		}
		if lock == nil || lock.UserID != input.UserID {
			return result{}, errPositionLockNotFound
		}
		if lock.Status != domain.PositionLockStatusActive {
			return result{}, errPositionLockInactive
		}

		position, err := getPositionForUpdateTx(ctx, tx, lock.UserID, lock.ContractID, string(lock.Side))
		if err != nil {
			return result{}, err
		}
		if position == nil || position.LockedQuantity < lock.Quantity {
			return result{}, fmt.Errorf("position locked quantity underflow for position lock %d", lock.ID)
		}

		if err := updatePositionBalancesTx(ctx, tx, position.ID, position.AvailableQuantity+lock.Quantity, position.LockedQuantity-lock.Quantity); err != nil {
			return result{}, err
		}
		if err := releasePositionLockTx(ctx, tx, lock.ID); err != nil {
			return result{}, err
		}

		lock, err = getPositionLockForUpdateTx(ctx, tx, lock.ID)
		if err != nil {
			return result{}, err
		}
		position, err = getPositionByIDTx(ctx, tx, position.ID)
		if err != nil {
			return result{}, err
		}

		return result{lock: lock, position: position}, validatePosition(*position)
	})
	if err != nil {
		return nil, nil, err
	}

	return outcome.lock, outcome.position, nil
}

func insertPositionLockTx(ctx context.Context, tx *sql.Tx, input store.CreatePositionLockInput) (*domain.PositionLock, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO position_locks (
			user_id,
			contract_id,
			side,
			quantity,
			status,
			order_id,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'active', NULL, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
			id,
			user_id,
			contract_id,
			side,
			quantity,
			status,
			order_id,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at,
			released_at
	`,
		input.UserID,
		input.ContractID,
		strings.ToLower(strings.TrimSpace(input.Side)),
		input.Quantity,
		strings.TrimSpace(input.ReferenceType),
		strings.TrimSpace(input.ReferenceID),
		strings.TrimSpace(input.CorrelationID),
		strings.TrimSpace(input.Description),
	)

	lock, err := scanPositionLock(row)
	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func getPositionLockForUpdateTx(ctx context.Context, tx *sql.Tx, lockID int64) (*domain.PositionLock, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			contract_id,
			side,
			quantity,
			status,
			order_id,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at,
			released_at
		FROM position_locks
		WHERE id = $1
		FOR UPDATE
	`, lockID)

	lock, err := scanPositionLock(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func releasePositionLockTx(ctx context.Context, tx *sql.Tx, lockID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE position_locks
		SET status = 'released', updated_at = NOW(), released_at = NOW()
		WHERE id = $1
	`, lockID)
	return err
}

func getPositionForUpdateTx(ctx context.Context, tx *sql.Tx, userID, contractID int64, side string) (*domain.Position, error) {
	row := tx.QueryRowContext(ctx, `
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
		WHERE user_id = $1 AND contract_id = $2 AND side = $3
		FOR UPDATE
	`, userID, contractID, strings.ToLower(strings.TrimSpace(side)))

	position, err := scanPosition(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &position, nil
}

func getPositionByIDTx(ctx context.Context, tx *sql.Tx, positionID int64) (*domain.Position, error) {
	row := tx.QueryRowContext(ctx, `
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
		WHERE id = $1
	`, positionID)

	position, err := scanPosition(row)
	if err != nil {
		return nil, err
	}

	return &position, nil
}

func updatePositionBalancesTx(ctx context.Context, tx *sql.Tx, positionID, availableQuantity, lockedQuantity int64) error {
	totalQuantity := availableQuantity + lockedQuantity
	_, err := tx.ExecContext(ctx, `
		UPDATE positions
		SET
			available_quantity = $2,
			locked_quantity = $3,
			total_quantity = $4,
			updated_at = NOW()
		WHERE id = $1
	`, positionID, availableQuantity, lockedQuantity, totalQuantity)
	return err
}

func validatePosition(position domain.Position) error {
	if position.AvailableQuantity < 0 || position.LockedQuantity < 0 || position.TotalQuantity < 0 {
		return fmt.Errorf("position %d has negative balances", position.ID)
	}
	if position.AvailableQuantity+position.LockedQuantity != position.TotalQuantity {
		return fmt.Errorf("position %d violates quantity invariant", position.ID)
	}

	return nil
}
