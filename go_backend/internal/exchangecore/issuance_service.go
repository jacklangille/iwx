package exchangecore

import (
	"context"
	"fmt"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

type IssuanceDetails struct {
	Batch     *domain.IssuanceBatch
	Lock      *domain.CollateralLock
	Positions []*domain.Position
}

func (s *Service) ListIssuanceBatches(ctx context.Context, contractID, userID int64) ([]domain.IssuanceBatch, error) {
	_, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListIssuanceBatches(ctx, userID, contractID)
}

func (s *Service) ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.ListPositions(ctx, userID, contractID)
}

func (s *Service) IssueContractSupply(ctx context.Context, contractID, userID, collateralLockID, pairedQuantity int64) (*IssuanceDetails, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}
	if contract.Status != string(domain.ContractStatePendingCollateral) {
		return nil, fmt.Errorf("%w: contract must be pending collateral", ErrInvalidContractState)
	}
	if collateralLockID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"collateral_lock_id": {"must be present"},
		}}
	}
	if pairedQuantity <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"paired_quantity": {"must be greater than 0"},
		}}
	}

	batch, lock, positions, err := s.repo.CreateIssuanceBatch(ctx, store.CreateIssuanceBatchInput{
		UserID:           userID,
		ContractID:       contractID,
		CollateralLockID: collateralLockID,
		PairedQuantity:   pairedQuantity,
	})
	if err != nil {
		return nil, err
	}

	details := &IssuanceDetails{
		Batch:     batch,
		Lock:      lock,
		Positions: positions,
	}

	return details, s.projectUser(ctx, userID)
}
