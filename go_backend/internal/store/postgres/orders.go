package postgres

import (
	"context"
	"database/sql"
	"strings"

	"iwx/go_backend/internal/domain"
)

type OrderRepository struct {
	*baseRepository
}

func NewOrderRepository(databaseURL string) *OrderRepository {
	return &OrderRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *OrderRepository) ListOpenOrders(ctx context.Context, contractID *int64) ([]domain.Order, error) {
	query := `
		SELECT
			id,
			contract_id,
			user_id,
			token_type,
			order_side,
			price::text,
			quantity,
			status,
			cash_reservation_id,
			position_lock_id,
			reservation_correlation_id,
			inserted_at,
			updated_at
		FROM orders
		WHERE status = 'open'
	`

	args := []any{}
	if contractID != nil {
		query += ` AND contract_id = $1`
		args = append(args, *contractID)
	}

	query += ` ORDER BY inserted_at ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []domain.Order{}
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}

		order.Price = normalizeNumericString(order.Price)
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func normalizeNumericString(value string) string {
	if strings.Contains(value, ".") {
		parts := strings.SplitN(value, ".", 2)
		if len(parts[1]) == 1 {
			return value + "0"
		}
	}

	return value
}

func (r *OrderRepository) ReplaceOpenOrdersProjection(ctx context.Context, contractID int64, orders []domain.Order) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM orders WHERE contract_id = $1`, contractID); err != nil {
			return struct{}{}, err
		}

		for _, order := range orders {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO orders (
					id,
					contract_id,
					user_id,
					token_type,
					order_side,
					price,
					quantity,
					status,
					cash_reservation_id,
					position_lock_id,
					reservation_correlation_id,
					inserted_at,
					updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`,
				order.ID,
				order.ContractID,
				order.UserID,
				order.TokenType,
				order.OrderSide,
				order.Price,
				order.Quantity,
				order.Status,
				order.CashReservationID,
				order.PositionLockID,
				order.ReservationCorrelationID,
				order.InsertedAt.UTC(),
				order.UpdatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}
