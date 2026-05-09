package matching

import (
	"context"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/readprojection"
)

type projectingHandler struct {
	next      Handler
	projector *readprojection.Projector
}

func NewProjectingHandler(next Handler, projector *readprojection.Projector) Handler {
	if projector == nil {
		return next
	}

	return &projectingHandler{
		next:      next,
		projector: projector,
	}
}

func (h *projectingHandler) HandlePlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error) {
	result, err := h.next.HandlePlaceOrder(ctx, envelope)
	if projectErr := h.projector.ProjectOrderCommand(ctx, envelope.CommandID); projectErr != nil {
		if err == nil {
			return result, projectErr
		}
	}
	if err != nil {
		return result, err
	}

	if err := h.projector.ProjectMarket(ctx, result.ContractID); err != nil {
		return result, err
	}

	return result, nil
}
