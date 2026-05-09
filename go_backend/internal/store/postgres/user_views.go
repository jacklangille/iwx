package postgres

import (
	"context"
	"database/sql"
	"strings"

	"iwx/go_backend/internal/domain"
)

func (r *ContractRepository) ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			currency,
			available_cents,
			locked_cents,
			total_cents,
			updated_at
		FROM cash_accounts
		WHERE user_id = $1
		ORDER BY currency ASC, id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []domain.CashAccount{}
	for rows.Next() {
		account, err := scanCashAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (r *ContractRepository) ReplaceCashAccountsProjection(ctx context.Context, userID int64, accounts []domain.CashAccount) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM cash_accounts WHERE user_id = $1`, userID); err != nil {
			return struct{}{}, err
		}

		for _, account := range accounts {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO cash_accounts (
					id,
					user_id,
					currency,
					available_cents,
					locked_cents,
					total_cents,
					updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`,
				account.ID,
				account.UserID,
				account.Currency,
				account.AvailableCents,
				account.LockedCents,
				account.TotalCents,
				account.UpdatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) ReplacePositionsProjection(ctx context.Context, userID int64, positions []domain.Position) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM positions WHERE user_id = $1`, userID); err != nil {
			return struct{}{}, err
		}

		for _, position := range positions {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO positions (
					id,
					user_id,
					contract_id,
					side,
					available_quantity,
					locked_quantity,
					total_quantity,
					updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`,
				position.ID,
				position.UserID,
				position.ContractID,
				string(position.Side),
				position.AvailableQuantity,
				position.LockedQuantity,
				position.TotalQuantity,
				position.UpdatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) ReplacePositionLocksProjection(ctx context.Context, userID int64, locks []domain.PositionLock) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM position_locks WHERE user_id = $1`, userID); err != nil {
			return struct{}{}, err
		}

		for _, lock := range locks {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO position_locks (
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
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			`,
				lock.ID,
				lock.UserID,
				lock.ContractID,
				string(lock.Side),
				lock.Quantity,
				string(lock.Status),
				lock.OrderID,
				lock.ReferenceType,
				lock.ReferenceID,
				lock.CorrelationID,
				lock.Description,
				lock.CreatedAt.UTC(),
				lock.UpdatedAt.UTC(),
				lock.ReleasedAt,
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) ReplaceCollateralLocksProjection(ctx context.Context, userID int64, currency string, locks []domain.CollateralLock) error {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM collateral_locks WHERE user_id = $1 AND currency = $2`, userID, normalized); err != nil {
			return struct{}{}, err
		}

		for _, lock := range locks {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO collateral_locks (
					id,
					user_id,
					contract_id,
					currency,
					amount_cents,
					status,
					reference_id,
					description,
					reference_issuance_id,
					created_at,
					updated_at,
					released_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`,
				lock.ID,
				lock.UserID,
				lock.ContractID,
				lock.Currency,
				lock.AmountCents,
				string(lock.Status),
				lock.ReferenceID,
				lock.Description,
				lock.ReferenceIssuanceID,
				lock.CreatedAt.UTC(),
				lock.UpdatedAt.UTC(),
				lock.ReleasedAt,
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) ReplaceOrderCashReservationsProjection(ctx context.Context, userID int64, currency string, reservations []domain.OrderCashReservation) error {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM order_cash_reservations WHERE user_id = $1 AND currency = $2`, userID, normalized); err != nil {
			return struct{}{}, err
		}

		for _, reservation := range reservations {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO order_cash_reservations (
					id,
					user_id,
					contract_id,
					currency,
					amount_cents,
					status,
					reference_type,
					reference_id,
					correlation_id,
					description,
					created_at,
					updated_at,
					released_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`,
				reservation.ID,
				reservation.UserID,
				reservation.ContractID,
				reservation.Currency,
				reservation.AmountCents,
				string(reservation.Status),
				reservation.ReferenceType,
				reservation.ReferenceID,
				reservation.CorrelationID,
				reservation.Description,
				reservation.CreatedAt.UTC(),
				reservation.UpdatedAt.UTC(),
				reservation.ReleasedAt,
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}
