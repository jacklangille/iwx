package exchangecore

import (
	"context"
	"testing"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

func TestCalculateCollateralRequirementAppliesMinimumFloor(t *testing.T) {
	t.Parallel()

	userID := int64(12)
	service := &Service{
		repo: stubExchangeCoreRepository{
			getContractFn: func(_ context.Context, contractID int64) (*domain.Contract, error) {
				return &domain.Contract{
					ID:            contractID,
					CreatorUserID: &userID,
					Status:        string(domain.ContractStatePendingCollateral),
				}, nil
			},
		},
	}

	requirement, err := service.CalculateCollateralRequirement(context.Background(), 44, userID, 10, "usd")
	if err != nil {
		t.Fatalf("CalculateCollateralRequirement() error = %v", err)
	}
	if requirement.RequiredAmountCents != MinimumContractCollateralCents {
		t.Fatalf("expected minimum collateral floor %d, got %d", MinimumContractCollateralCents, requirement.RequiredAmountCents)
	}
	if requirement.PerPairCents != DefaultCollateralPerPairCents {
		t.Fatalf("expected per-pair collateral %d, got %d", DefaultCollateralPerPairCents, requirement.PerPairCents)
	}
}

func TestSubmitContractForApprovalTransitionsDraftToPendingApproval(t *testing.T) {
	t.Parallel()

	userID := int64(7)
	service := &Service{
		repo: stubExchangeCoreRepository{
			getContractFn: func(_ context.Context, contractID int64) (*domain.Contract, error) {
				return &domain.Contract{ID: contractID, CreatorUserID: &userID, Status: string(domain.ContractStateDraft)}, nil
			},
			updateContractStatusFn: func(_ context.Context, contractID int64, status string) (*domain.Contract, error) {
				if status != string(domain.ContractStatePendingApproval) {
					t.Fatalf("expected status %q, got %q", domain.ContractStatePendingApproval, status)
				}
				return &domain.Contract{ID: contractID, CreatorUserID: &userID, Status: status}, nil
			},
		},
	}

	contract, err := service.SubmitContractForApproval(context.Background(), 42, userID)
	if err != nil {
		t.Fatalf("SubmitContractForApproval() error = %v", err)
	}
	if contract.Status != string(domain.ContractStatePendingApproval) {
		t.Fatalf("expected pending approval status, got %q", contract.Status)
	}
}

func TestActivateContractRequiresIssuedSupply(t *testing.T) {
	t.Parallel()

	userID := int64(8)
	service := &Service{
		repo: stubExchangeCoreRepository{
			getContractFn: func(_ context.Context, contractID int64) (*domain.Contract, error) {
				return &domain.Contract{ID: contractID, CreatorUserID: &userID, Status: string(domain.ContractStatePendingCollateral)}, nil
			},
			listIssuanceBatchesFn: func(context.Context, int64, int64) ([]domain.IssuanceBatch, error) {
				return []domain.IssuanceBatch{{ID: 1, Status: domain.IssuanceBatchStatusPending}}, nil
			},
		},
	}

	_, err := service.ActivateContract(context.Background(), 9, userID)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Errors["issuance"]) == 0 {
		t.Fatalf("expected issuance validation error, got %#v", validationErr.Errors)
	}
}

func TestCancelContractReleasesOnlyLockedCollateral(t *testing.T) {
	t.Parallel()

	userID := int64(9)
	released := make([]store.ReleaseCollateralLockInput, 0, 1)
	service := &Service{
		repo: stubExchangeCoreRepository{
			getContractFn: func(_ context.Context, contractID int64) (*domain.Contract, error) {
				return &domain.Contract{ID: contractID, CreatorUserID: &userID, Status: string(domain.ContractStatePendingCollateral)}, nil
			},
			listIssuanceBatchesFn: func(context.Context, int64, int64) ([]domain.IssuanceBatch, error) {
				return nil, nil
			},
			listContractCollateralLocksFn: func(context.Context, int64, int64) ([]domain.CollateralLock, error) {
				return []domain.CollateralLock{
					{ID: 11, Status: domain.CollateralLockStatusLocked},
					{ID: 12, Status: domain.CollateralLockStatusReleased},
				}, nil
			},
			releaseCollateralLockFn: func(_ context.Context, input store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
				released = append(released, input)
				return &domain.CollateralLock{ID: input.LockID, Status: domain.CollateralLockStatusReleased}, nil, nil, nil
			},
			updateContractStatusFn: func(_ context.Context, contractID int64, status string) (*domain.Contract, error) {
				return &domain.Contract{ID: contractID, CreatorUserID: &userID, Status: status}, nil
			},
		},
	}

	contract, err := service.CancelContract(context.Background(), 77, userID)
	if err != nil {
		t.Fatalf("CancelContract() error = %v", err)
	}
	if contract.Status != string(domain.ContractStateCancelled) {
		t.Fatalf("expected cancelled status, got %q", contract.Status)
	}
	if len(released) != 1 {
		t.Fatalf("expected 1 collateral release, got %d", len(released))
	}
	if released[0].LockID != 11 {
		t.Fatalf("expected lock 11 to be released, got %d", released[0].LockID)
	}
}
