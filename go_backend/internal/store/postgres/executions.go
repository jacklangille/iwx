package postgres

import (
	"context"
	"database/sql"

	"iwx/go_backend/internal/domain"
)

type ExecutionRepository struct {
	*baseRepository
}

func NewExecutionRepository(databaseURL string) *ExecutionRepository {
	return &ExecutionRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *ExecutionRepository) ListExecutions(ctx context.Context, contractID int64, limit int) ([]domain.Execution, error) {
	query := `
		SELECT
			id,
			execution_id,
			contract_id,
			token_type,
			buy_order_id,
			sell_order_id,
			buyer_user_id,
			seller_user_id,
			buyer_cash_reservation_id,
			seller_position_lock_id,
			price::text,
			quantity,
			id AS sequence,
			occurred_at
		FROM executions
		WHERE contract_id = $1
		ORDER BY id DESC
	`

	args := []any{contractID}
	if limit > 0 {
		query += ` LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executions := []domain.Execution{}
	for rows.Next() {
		execution, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}

		execution.Price = normalizeNumericString(execution.Price)
		executions = append(executions, execution)
	}

	return executions, rows.Err()
}

func (r *ExecutionRepository) ReplaceExecutionsProjection(ctx context.Context, contractID int64, executions []domain.Execution) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM executions WHERE contract_id = $1`, contractID); err != nil {
			return struct{}{}, err
		}

		for _, execution := range executions {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO executions (
					id,
					execution_id,
					contract_id,
					token_type,
					buy_order_id,
					sell_order_id,
					buyer_user_id,
					seller_user_id,
					buyer_cash_reservation_id,
					seller_position_lock_id,
					price,
					quantity,
					occurred_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`,
				execution.ID,
				execution.ExecutionID,
				execution.ContractID,
				execution.TokenType,
				execution.BuyOrderID,
				execution.SellOrderID,
				execution.BuyerUserID,
				execution.SellerUserID,
				execution.BuyerCashReservationID,
				execution.SellerPositionLockID,
				execution.Price,
				execution.Quantity,
				execution.OccurredAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}
