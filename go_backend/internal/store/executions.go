package store

import (
	"context"

	"iwx/go_backend/internal/domain"
)

type ExecutionRepository interface {
	ListExecutions(ctx context.Context, contractID int64, limit int) ([]domain.Execution, error)
}

type ExecutionProjectionSource interface {
	ListExecutions(ctx context.Context, contractID int64, limit int) ([]domain.Execution, error)
}

type ExecutionProjectionTarget interface {
	ReplaceExecutionsProjection(ctx context.Context, contractID int64, executions []domain.Execution) error
}
