package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nats-io/nuid"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/market"
)

func findMatch(ctx context.Context, tx *sql.Tx, command commands.PlaceOrder, excludeOrderID int64) (*domain.Order, error) {
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
		WHERE
			id <> $1 AND
			contract_id = $2 AND
			token_type = $3 AND
			order_side = $4 AND
			status = 'open' AND
			price %s $5
		ORDER BY %s, inserted_at ASC
		LIMIT 1
		FOR UPDATE
	`

	otherSide := "ask"
	comparator := "<="
	ordering := "price ASC"
	if command.OrderSide == "ask" {
		otherSide = "bid"
		comparator = ">="
		ordering = "price DESC"
	}

	row := tx.QueryRowContext(ctx, fmt.Sprintf(query, comparator, ordering),
		excludeOrderID,
		command.ContractID,
		command.TokenType,
		otherSide,
		command.Price,
	)

	order, err := scanOrder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func updateMatchedOrder(ctx context.Context, tx *sql.Tx, matchedOrder domain.Order, matchedQuantity int64) error {
	nextQuantity := matchedOrder.Quantity - matchedQuantity
	nextStatus := "open"
	if nextQuantity == 0 {
		nextStatus = "filled"
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE orders
		SET quantity = $2, status = $3, updated_at = NOW()
		WHERE id = $1
	`, matchedOrder.ID, nextQuantity, nextStatus)
	return err
}

func insertIncomingOrder(ctx context.Context, tx *sql.Tx, command commands.PlaceOrder) (*domain.Order, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO orders (
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
		VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $8, $9, NOW(), NOW())
		RETURNING
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
	`, command.ContractID, command.UserID, command.TokenType, command.OrderSide, command.Price, command.Quantity, command.CashReservationID, command.PositionLockID, command.ReservationCorrelationID)

	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func updateIncomingOrder(ctx context.Context, tx *sql.Tx, orderID int64, remainingQuantity int64) (*domain.Order, error) {
	nextStatus := "open"
	if remainingQuantity == 0 {
		nextStatus = "filled"
	}

	row := tx.QueryRowContext(ctx, `
		UPDATE orders
		SET quantity = $2, status = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING
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
	`, orderID, remainingQuantity, nextStatus)

	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func getOrderByIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*domain.Order, error) {
	row := tx.QueryRowContext(ctx, `
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
		WHERE id = $1
	`, orderID)

	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func insertExecution(
	ctx context.Context,
	tx *sql.Tx,
	commandID string,
	incomingOrder domain.Order,
	matchedOrder domain.Order,
	matchedQuantity int64,
) (*domain.Execution, error) {
	executionID := nuid.Next()

	buyOrderID := incomingOrder.ID
	sellOrderID := matchedOrder.ID
	buyerUserID := incomingOrder.UserID
	sellerUserID := matchedOrder.UserID
	if incomingOrder.OrderSide == string(domain.OrderSideAsk) {
		buyOrderID = matchedOrder.ID
		sellOrderID = incomingOrder.ID
		buyerUserID = matchedOrder.UserID
		sellerUserID = incomingOrder.UserID
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO executions (
			execution_id,
			command_id,
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING
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
	`,
		executionID,
		commandID,
		incomingOrder.ContractID,
		incomingOrder.TokenType,
		buyOrderID,
		sellOrderID,
		buyerUserID,
		sellerUserID,
		buyOrderCashReservationID(incomingOrder, matchedOrder),
		sellOrderPositionLockID(incomingOrder, matchedOrder),
		matchedOrder.Price,
		matchedQuantity,
	)

	execution, err := scanExecution(row)
	if err != nil {
		return nil, err
	}

	return &execution, nil
}

func recordMarketSnapshot(ctx context.Context, tx *sql.Tx, contractID int64) error {
	orders, err := listOpenOrdersTx(ctx, tx, contractID)
	if err != nil {
		return err
	}

	summary := market.Summary(orders)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO market_snapshots (contract_id, best_above, best_below, mid_above, mid_below, inserted_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, contractID, summary.Best.Above.Bid, summary.Best.Below.Bid, summary.Mid.Above, summary.Mid.Below)

	return err
}

func listOpenOrdersTx(ctx context.Context, tx *sql.Tx, contractID int64) ([]domain.Order, error) {
	rows, err := tx.QueryContext(ctx, `
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
		WHERE contract_id = $1 AND status = 'open'
		ORDER BY inserted_at ASC, id ASC
	`, contractID)
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

		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func getOrderCommandTx(ctx context.Context, tx *sql.Tx, commandID string) (*commands.OrderCommand, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			command_id,
			contract_id,
			user_id,
			token_type,
			order_side,
			price,
			quantity,
			status,
			error_message,
			result_status,
			result_order_id,
			enqueued_at,
			started_at,
			completed_at,
			updated_at
		FROM order_commands
		WHERE command_id = $1
		FOR UPDATE
	`, commandID)

	command, err := scanOrderCommand(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &command, nil
}

func listExecutionsByCommandTx(ctx context.Context, tx *sql.Tx, commandID string) ([]domain.Execution, error) {
	rows, err := tx.QueryContext(ctx, `
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
		WHERE command_id = $1
		ORDER BY id ASC
	`, commandID)
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

func buyOrderCashReservationID(incomingOrder, matchedOrder domain.Order) *int64 {
	if incomingOrder.OrderSide == string(domain.OrderSideBid) {
		return incomingOrder.CashReservationID
	}

	return matchedOrder.CashReservationID
}

func sellOrderPositionLockID(incomingOrder, matchedOrder domain.Order) *int64 {
	if incomingOrder.OrderSide == string(domain.OrderSideAsk) {
		return incomingOrder.PositionLockID
	}

	return matchedOrder.PositionLockID
}
