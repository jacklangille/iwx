package store

import (
	"context"

	"iwx/go_backend/internal/domain"
)

type ContractRepository interface {
	ListContracts(ctx context.Context) ([]domain.Contract, error)
	ContractExists(ctx context.Context, contractID int64) (bool, error)
}

type ContractProjectionSource interface {
	GetContract(ctx context.Context, contractID int64) (*domain.Contract, error)
}

type ContractProjectionTarget interface {
	UpsertContractProjection(ctx context.Context, contract domain.Contract) error
}
