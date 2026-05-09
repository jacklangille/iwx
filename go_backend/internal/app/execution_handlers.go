package app

import (
	"context"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/readmodel"
	"iwx/go_backend/internal/readprojection"
)

type exchangeCoreExecutionHandler struct {
	service *exchangecore.Service
}

func (h *exchangeCoreExecutionHandler) HandleExecutionCreated(ctx context.Context, event events.ExecutionCreated) error {
	return h.service.ApplyExecution(ctx, event)
}

type readAPIExecutionHandler struct {
	reads     *readmodel.Service
	projector *readprojection.Projector
}

func (h *readAPIExecutionHandler) HandleExecutionCreated(ctx context.Context, event events.ExecutionCreated) error {
	if h.projector != nil {
		if err := h.projector.ProjectMarket(ctx, event.ContractID); err != nil {
			return err
		}
	}
	if h.reads != nil {
		h.reads.InvalidateMarket(event.ContractID)
	}
	return nil
}
