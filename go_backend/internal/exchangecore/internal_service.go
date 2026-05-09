package exchangecore

import (
	"context"
	"fmt"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

func (s *Service) MarkContractResolved(ctx context.Context, contractID int64) (*domain.Contract, error) {
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, ErrContractNotFound
	}

	switch contract.Status {
	case string(domain.ContractStateResolved), string(domain.ContractStateSettled):
		return contract, nil
	case string(domain.ContractStateActive), string(domain.ContractStateTradingClosed), string(domain.ContractStateAwaitingResolution):
	default:
		return nil, fmt.Errorf("%w: contract cannot be resolved from status %s", ErrInvalidContractState, contract.Status)
	}

	updated, err := s.repo.UpdateContractStatus(ctx, contractID, string(domain.ContractStateResolved))
	if err != nil {
		return nil, err
	}
	return updated, s.projectContract(ctx, contractID)
}

func (s *Service) SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
	result, err := s.repo.SettleContract(ctx, input)
	if err != nil {
		return nil, err
	}

	if s.emitter != nil {
		if err := s.projectContract(ctx, input.ContractID); err != nil {
			return nil, err
		}
		if err := s.emitter.EmitSettlementChanged(ctx, input.ContractID, result.SettledAt); err != nil {
			return nil, err
		}
		for _, userID := range result.AffectedUsers {
			if err := s.emitter.EmitUserStateChanged(ctx, userID, result.SettledAt); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
