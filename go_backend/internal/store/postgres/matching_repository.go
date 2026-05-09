package postgres

import (
	"context"
	"database/sql"
	"errors"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

type MatchingRepository struct {
	*baseRepository
}

func NewMatchingRepository(databaseURL string) *MatchingRepository {
	return &MatchingRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *MatchingRepository) ProcessPlaceOrder(
	ctx context.Context,
	envelope commands.PlaceOrderEnvelope,
) (commands.PlaceOrderResult, error) {
	return withTransaction(ctx, r.db, func(tx *sql.Tx) (commands.PlaceOrderResult, error) {
		existing, err := getOrderCommandTx(ctx, tx, envelope.CommandID)
		if err != nil {
			return commands.PlaceOrderResult{}, err
		}
		if existing != nil && existing.Status == "succeeded" {
			return replayPlaceOrderResult(ctx, tx, *existing)
		}

		if err := markCommandProcessing(ctx, tx, envelope); err != nil {
			return commands.PlaceOrderResult{}, err
		}

		result, err := executePlaceOrder(ctx, tx, envelope.CommandID, envelope.Command)
		if err != nil {
			if updateErr := markCommandFailed(ctx, tx, envelope.CommandID, err.Error()); updateErr != nil {
				return commands.PlaceOrderResult{}, updateErr
			}
			return commands.PlaceOrderResult{}, err
		}

		if err := markCommandSucceeded(ctx, tx, envelope.CommandID, result); err != nil {
			return commands.PlaceOrderResult{}, err
		}

		return result, nil
	})
}

func executePlaceOrder(
	ctx context.Context,
	tx *sql.Tx,
	commandID string,
	command commands.PlaceOrder,
) (commands.PlaceOrderResult, error) {
	incomingOrder, err := insertIncomingOrder(ctx, tx, command)
	if err != nil {
		return commands.PlaceOrderResult{}, err
	}

	remainingQuantity := command.Quantity
	executions := make([]domain.Execution, 0)

	for remainingQuantity > 0 {
		matchedOrder, err := findMatch(ctx, tx, command, incomingOrder.ID)
		if err != nil {
			return commands.PlaceOrderResult{}, err
		}

		if matchedOrder == nil {
			break
		}

		matchedQuantity := remainingQuantity
		if matchedOrder.Quantity < matchedQuantity {
			matchedQuantity = matchedOrder.Quantity
		}

		execution, err := insertExecution(ctx, tx, commandID, *incomingOrder, *matchedOrder, matchedQuantity)
		if err != nil {
			return commands.PlaceOrderResult{}, err
		}

		if err := updateMatchedOrder(ctx, tx, *matchedOrder, matchedQuantity); err != nil {
			return commands.PlaceOrderResult{}, err
		}

		executions = append(executions, *execution)
		remainingQuantity -= matchedQuantity
	}

	var openOrder *domain.Order
	updatedOrder, err := updateIncomingOrder(ctx, tx, incomingOrder.ID, remainingQuantity)
	if err != nil {
		return commands.PlaceOrderResult{}, err
	}
	if updatedOrder.Status == "open" {
		openOrder = updatedOrder
	}

	if err := recordMarketSnapshot(ctx, tx, command.ContractID); err != nil {
		return commands.PlaceOrderResult{}, err
	}

	result := commands.PlaceOrderResult{
		Status:     updatedOrder.Status,
		ContractID: command.ContractID,
		Executions: executions,
	}
	if openOrder != nil {
		result.Order = openOrder
	}

	return result, nil
}

func replayPlaceOrderResult(
	ctx context.Context,
	tx *sql.Tx,
	command commands.OrderCommand,
) (commands.PlaceOrderResult, error) {
	result := commands.PlaceOrderResult{
		Status:     stringValue(command.ResultStatus, "succeeded"),
		ContractID: command.ContractID,
	}
	executions, err := listExecutionsByCommandTx(ctx, tx, command.CommandID)
	if err != nil {
		return commands.PlaceOrderResult{}, err
	}
	result.Executions = executions

	if command.ResultOrderID != nil {
		order, err := getOrderByIDTx(ctx, tx, *command.ResultOrderID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return commands.PlaceOrderResult{}, err
		}
		if err == nil {
			result.Order = order
		}
	}

	return result, nil
}

func stringValue(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}

	return *value
}
