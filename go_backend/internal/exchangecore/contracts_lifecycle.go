package exchangecore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

const (
	DefaultContractRuleVersion            = "v1"
	DefaultResolutionInclusiveSide string = "below"
	DefaultCollateralPerPairCents  int64  = 100
	MinimumContractCollateralCents int64  = 100000
)

var (
	ErrContractNotFound     = errors.New("contract not found")
	ErrContractForbidden    = errors.New("forbidden")
	ErrInvalidContractState = errors.New("invalid contract state")
)

type ContractDetails struct {
	Contract *domain.Contract
	Rule     *domain.ContractRule
}

type CollateralRequirement struct {
	ContractID          int64
	PairedQuantity      int64
	PerPairCents        int64
	RequiredAmountCents int64
	Currency            string
}

func normalizeClaimSide(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return DefaultResolutionInclusiveSide
	}

	return trimmed
}

func (s *Service) GetContractDetails(ctx context.Context, contractID, userID int64) (*ContractDetails, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}

	rule, err := s.repo.GetContractRule(ctx, contractID)
	if err != nil {
		return nil, err
	}

	return &ContractDetails{Contract: contract, Rule: rule}, nil
}

func (s *Service) SubmitContractForApproval(ctx context.Context, contractID, userID int64) (*domain.Contract, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}
	if contract.Status != string(domain.ContractStateDraft) {
		return nil, fmt.Errorf("%w: contract must be in draft state", ErrInvalidContractState)
	}

	updated, err := s.repo.UpdateContractStatus(ctx, contractID, string(domain.ContractStatePendingApproval))
	if err != nil {
		return nil, err
	}

	return updated, s.projectContract(ctx, contractID)
}

func (s *Service) ApproveContract(ctx context.Context, contractID, userID int64) (*domain.Contract, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}
	if contract.Status != string(domain.ContractStatePendingApproval) {
		return nil, fmt.Errorf("%w: contract must be pending approval", ErrInvalidContractState)
	}

	updated, err := s.repo.UpdateContractStatus(ctx, contractID, string(domain.ContractStatePendingCollateral))
	if err != nil {
		return nil, err
	}

	return updated, s.projectContract(ctx, contractID)
}

func (s *Service) CalculateCollateralRequirement(ctx context.Context, contractID, userID, pairedQuantity int64, currency string) (*CollateralRequirement, error) {
	if pairedQuantity <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"paired_quantity": {"must be greater than 0"},
		}}
	}

	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}
	if contract.Status != string(domain.ContractStatePendingCollateral) && contract.Status != string(domain.ContractStateActive) {
		return nil, fmt.Errorf("%w: contract must be pending collateral or active", ErrInvalidContractState)
	}

	perPair := DefaultCollateralPerPairCents
	if contract.Multiplier != nil && *contract.Multiplier > 0 {
		perPair = *contract.Multiplier
	}

	return &CollateralRequirement{
		ContractID:          contractID,
		PairedQuantity:      pairedQuantity,
		PerPairCents:        perPair,
		RequiredAmountCents: maxInt64(perPair*pairedQuantity, MinimumContractCollateralCents),
		Currency:            normalizeCurrency(currency),
	}, nil
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

type LockContractCollateralResult struct {
	Requirement *CollateralRequirement
	Lock        *domain.CollateralLock
	Account     *domain.CashAccount
	LedgerEntry *domain.LedgerEntry
}

func (s *Service) LockContractCollateral(ctx context.Context, contractID, userID, pairedQuantity int64, currency, correlationID, description string) (*LockContractCollateralResult, error) {
	requirement, err := s.CalculateCollateralRequirement(ctx, contractID, userID, pairedQuantity, currency)
	if err != nil {
		return nil, err
	}

	lock, account, entry, err := s.repo.CreateCollateralLock(ctx, store.CreateCollateralLockInput{
		UserID:        userID,
		ContractID:    contractID,
		Currency:      requirement.Currency,
		AmountCents:   requirement.RequiredAmountCents,
		ReferenceID:   fmt.Sprintf("contract:%d:pairs:%d", contractID, pairedQuantity),
		CorrelationID: strings.TrimSpace(correlationID),
		Description:   strings.TrimSpace(description),
	})
	if err != nil {
		return nil, err
	}

	result := &LockContractCollateralResult{
		Requirement: requirement,
		Lock:        lock,
		Account:     account,
		LedgerEntry: entry,
	}

	return result, s.projectUser(ctx, userID)
}

func (s *Service) ActivateContract(ctx context.Context, contractID, userID int64) (*domain.Contract, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}
	if contract.Status != string(domain.ContractStatePendingCollateral) {
		return nil, fmt.Errorf("%w: contract must be pending collateral", ErrInvalidContractState)
	}

	batches, err := s.repo.ListIssuanceBatches(ctx, userID, contractID)
	if err != nil {
		return nil, err
	}

	hasIssuedSupply := false
	for _, batch := range batches {
		if batch.Status == domain.IssuanceBatchStatusIssued {
			hasIssuedSupply = true
			break
		}
	}
	if !hasIssuedSupply {
		return nil, &ValidationError{Errors: map[string][]string{
			"issuance": {"issued supply is required before activation"},
		}}
	}

	updated, err := s.repo.UpdateContractStatus(ctx, contractID, string(domain.ContractStateActive))
	if err != nil {
		return nil, err
	}

	return updated, s.projectContract(ctx, contractID)
}

func (s *Service) CancelContract(ctx context.Context, contractID, userID int64) (*domain.Contract, error) {
	contract, err := s.requireOwnedContract(ctx, contractID, userID)
	if err != nil {
		return nil, err
	}

	switch contract.Status {
	case string(domain.ContractStateDraft), string(domain.ContractStatePendingApproval), string(domain.ContractStatePendingCollateral):
	default:
		return nil, fmt.Errorf("%w: contract cannot be cancelled from status %s", ErrInvalidContractState, contract.Status)
	}

	batches, err := s.repo.ListIssuanceBatches(ctx, userID, contractID)
	if err != nil {
		return nil, err
	}
	for _, batch := range batches {
		if batch.Status == domain.IssuanceBatchStatusIssued {
			return nil, fmt.Errorf("%w: contract cannot be cancelled after issuance", ErrInvalidContractState)
		}
	}

	locks, err := s.repo.ListContractCollateralLocks(ctx, userID, contractID)
	if err != nil {
		return nil, err
	}
	for _, lock := range locks {
		if lock.Status != domain.CollateralLockStatusLocked {
			continue
		}
		if _, _, _, err := s.repo.ReleaseCollateralLock(ctx, store.ReleaseCollateralLockInput{
			UserID:        userID,
			LockID:        lock.ID,
			CorrelationID: fmt.Sprintf("contract-%d-cancel", contractID),
			Description:   "contract cancelled before activation",
		}); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.UpdateContractStatus(ctx, contractID, string(domain.ContractStateCancelled))
	if err != nil {
		return nil, err
	}

	if err := s.projectUser(ctx, userID); err != nil {
		return nil, err
	}

	return updated, s.projectContract(ctx, contractID)
}

func (s *Service) requireOwnedContract(ctx context.Context, contractID, userID int64) (*domain.Contract, error) {
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, ErrContractNotFound
	}
	if contract.CreatorUserID == nil || *contract.CreatorUserID != userID {
		return nil, ErrContractForbidden
	}

	return contract, nil
}

func (s *Service) projectContract(ctx context.Context, contractID int64) error {
	if s.projector == nil {
		return nil
	}

	return s.projector.ProjectContract(ctx, contractID)
}
