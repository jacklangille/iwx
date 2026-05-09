package exchangecore

import (
	"context"

	"iwx/go_backend/internal/projectionbundle"
)

func (s *Service) GetContractBundle(ctx context.Context, contractID int64) (*projectionbundle.ContractBundle, error) {
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, nil
	}

	rule, err := s.repo.GetContractRule(ctx, contractID)
	if err != nil {
		return nil, err
	}

	return &projectionbundle.ContractBundle{
		Contract: contract,
		Rule:     rule,
	}, nil
}

func (s *Service) GetUserStateBundle(ctx context.Context, userID int64) (*projectionbundle.UserStateBundle, error) {
	accounts, err := s.repo.ListCashAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	positions, err := s.repo.ListPositions(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	positionLocks, err := s.repo.ListPositionLocks(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	collateralLocks, err := s.repo.ListCollateralLocks(ctx, userID, "USD")
	if err != nil {
		return nil, err
	}
	cashReservations, err := s.repo.ListOrderCashReservations(ctx, userID, "USD")
	if err != nil {
		return nil, err
	}
	settlements, err := s.repo.ListSettlementEntriesByUser(ctx, userID, nil, 500)
	if err != nil {
		return nil, err
	}

	return &projectionbundle.UserStateBundle{
		UserID:           userID,
		CashAccounts:     accounts,
		Positions:        positions,
		PositionLocks:    positionLocks,
		CollateralLocks:  collateralLocks,
		CashReservations: cashReservations,
		Settlements:      settlements,
	}, nil
}

func (s *Service) GetSettlementBundle(ctx context.Context, contractID int64) (*projectionbundle.SettlementBundle, error) {
	entries, err := s.repo.ListSettlementEntriesByContract(ctx, contractID, 500)
	if err != nil {
		return nil, err
	}
	return &projectionbundle.SettlementBundle{
		ContractID: contractID,
		Entries:    entries,
	}, nil
}
