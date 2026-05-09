package postgres

import (
	"context"
	"database/sql"
	"errors"

	"iwx/go_backend/internal/commands"
)

func (r *MatchingRepository) GetOrderCommand(ctx context.Context, commandID string) (*commands.OrderCommand, error) {
	row := r.db.QueryRowContext(ctx, `
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

func markCommandProcessing(ctx context.Context, tx *sql.Tx, envelope commands.PlaceOrderEnvelope) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO order_commands (
			command_id,
			contract_id,
			user_id,
			token_type,
			order_side,
			price,
			quantity,
			status,
			enqueued_at,
			started_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'processing', $8, NOW(), NOW())
		ON CONFLICT (command_id) DO UPDATE
		SET
			status = EXCLUDED.status,
			started_at = NOW(),
			updated_at = NOW()
	`, envelope.CommandID, envelope.Command.ContractID, envelope.Command.UserID, envelope.Command.TokenType, envelope.Command.OrderSide, envelope.Command.Price, envelope.Command.Quantity, envelope.EnqueuedAt.UTC())
	return err
}

func markCommandSucceeded(ctx context.Context, tx *sql.Tx, commandID string, result commands.PlaceOrderResult) error {
	var orderID any
	if result.Order != nil {
		orderID = result.Order.ID
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE order_commands
		SET
			status = 'succeeded',
			error_message = NULL,
			result_status = $2,
			result_order_id = $3,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE command_id = $1
	`, commandID, result.Status, orderID)
	return err
}

func markCommandFailed(ctx context.Context, tx *sql.Tx, commandID, errorMessage string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE order_commands
		SET
			status = 'failed',
			error_message = $2,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE command_id = $1
	`, commandID, errorMessage)
	return err
}

func (r *MatchingRepository) UpsertOrderCommandProjection(ctx context.Context, command commands.OrderCommand) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO order_commands (
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
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (command_id) DO UPDATE
		SET
			contract_id = EXCLUDED.contract_id,
			user_id = EXCLUDED.user_id,
			token_type = EXCLUDED.token_type,
			order_side = EXCLUDED.order_side,
			price = EXCLUDED.price,
			quantity = EXCLUDED.quantity,
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			result_status = EXCLUDED.result_status,
			result_order_id = EXCLUDED.result_order_id,
			enqueued_at = EXCLUDED.enqueued_at,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`,
		command.CommandID,
		command.ContractID,
		command.UserID,
		command.TokenType,
		command.OrderSide,
		command.Price,
		command.Quantity,
		command.Status,
		command.ErrorMessage,
		command.ResultStatus,
		command.ResultOrderID,
		command.EnqueuedAt.UTC(),
		command.StartedAt,
		command.CompletedAt,
		command.UpdatedAt.UTC(),
	)
	return err
}
