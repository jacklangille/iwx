package app

import (
	"context"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/exchangecore"
)

type exchangeCoreExecutionHandler struct {
	service *exchangecore.Service
}

func (h *exchangeCoreExecutionHandler) HandleExecutionCreated(ctx context.Context, event events.ExecutionCreated) error {
	return h.service.ApplyExecution(ctx, event)
}
