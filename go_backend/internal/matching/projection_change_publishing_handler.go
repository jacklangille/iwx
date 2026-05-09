package matching

import (
	"context"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/projectionchange"
	"iwx/go_backend/internal/store"
)

type projectionChangePublishingHandler struct {
	next             Handler
	emitter          *projectionchange.Emitter
	orderCommandRepo store.OrderCommandRepository
}

func NewProjectionChangePublishingHandler(next Handler, emitter *projectionchange.Emitter, orderCommandRepo store.OrderCommandRepository) Handler {
	if emitter == nil {
		return next
	}

	return &projectionChangePublishingHandler{
		next:             next,
		emitter:          emitter,
		orderCommandRepo: orderCommandRepo,
	}
}

func (h *projectionChangePublishingHandler) HandlePlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error) {
	result, err := h.next.HandlePlaceOrder(ctx, envelope)

	commandVersion := envelope.EnqueuedAt
	commandUserID := envelope.Command.UserID
	commandContractID := envelope.Command.ContractID
	if h.orderCommandRepo != nil {
		command, lookupErr := h.orderCommandRepo.GetOrderCommand(ctx, envelope.CommandID)
		if lookupErr != nil {
			if err == nil {
				return result, lookupErr
			}
		} else if command != nil {
			commandVersion = command.UpdatedAt
			commandUserID = command.UserID
			commandContractID = command.ContractID
		}
	}
	if projectErr := h.emitter.EmitOrderCommandChanged(ctx, envelope.CommandID, commandContractID, commandUserID, commandVersion); projectErr != nil {
		if err == nil {
			return result, projectErr
		}
	}
	if err != nil {
		return result, err
	}

	if result.ContractID > 0 {
		if projectErr := h.emitter.EmitMarketChanged(ctx, result.ContractID, latestMarketVersion(result)); projectErr != nil {
			return result, projectErr
		}
	}

	return result, nil
}

func latestMarketVersion(result commands.PlaceOrderResult) time.Time {
	latest := time.Time{}
	if result.Order != nil && result.Order.UpdatedAt.After(latest) {
		latest = result.Order.UpdatedAt
	}
	for _, execution := range result.Executions {
		if execution.OccurredAt.After(latest) {
			latest = execution.OccurredAt
		}
	}
	if latest.IsZero() {
		latest = time.Now().UTC()
	}
	return latest
}
