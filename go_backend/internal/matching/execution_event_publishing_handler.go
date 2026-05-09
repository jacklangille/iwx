package matching

import (
	"context"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/requestctx"
)

type ExecutionEventPublisher interface {
	PublishExecutionCreated(ctx context.Context, event events.ExecutionCreated) error
	Close()
}

type executionEventPublishingHandler struct {
	next      Handler
	publisher ExecutionEventPublisher
}

func NewExecutionEventPublishingHandler(next Handler, publisher ExecutionEventPublisher) Handler {
	if publisher == nil {
		return next
	}

	return &executionEventPublishingHandler{
		next:      next,
		publisher: publisher,
	}
}

func (h *executionEventPublishingHandler) HandlePlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error) {
	result, err := h.next.HandlePlaceOrder(ctx, envelope)
	if err != nil {
		return result, err
	}

	for _, execution := range result.Executions {
		if err := h.publisher.PublishExecutionCreated(ctx, events.ExecutionCreated{
			ExecutionID:            execution.ExecutionID,
			ContractID:             execution.ContractID,
			TokenType:              execution.TokenType,
			BuyOrderID:             execution.BuyOrderID,
			SellOrderID:            execution.SellOrderID,
			BuyerUserID:            execution.BuyerUserID,
			SellerUserID:           execution.SellerUserID,
			BuyerCashReservationID: execution.BuyerCashReservationID,
			SellerPositionLockID:   execution.SellerPositionLockID,
			Price:                  execution.Price,
			Quantity:               execution.Quantity,
			Sequence:               execution.Sequence,
			TraceID:                requestctx.TraceID(ctx),
			OccurredAt:             execution.OccurredAt,
		}); err != nil {
			return result, err
		}
	}

	return result, nil
}
