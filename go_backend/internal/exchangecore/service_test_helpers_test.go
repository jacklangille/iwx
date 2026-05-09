package exchangecore

import (
	"context"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/store"
)

type stubExchangeCoreRepository struct {
	processCreateContractFn       func(context.Context, commands.CreateContractEnvelope) (commands.CreateContractResult, error)
	findDuplicateContractFn       func(context.Context, store.FindDuplicateContractInput) (*domain.Contract, error)
	getContractCommandFn          func(context.Context, string) (*commands.ContractCommand, error)
	getContractFn                 func(context.Context, int64) (*domain.Contract, error)
	getContractRuleFn             func(context.Context, int64) (*domain.ContractRule, error)
	updateContractStatusFn        func(context.Context, int64, string) (*domain.Contract, error)
	listContractCollateralLocksFn func(context.Context, int64, int64) ([]domain.CollateralLock, error)
	listIssuanceBatchesFn         func(context.Context, int64, int64) ([]domain.IssuanceBatch, error)
	createPositionLockFn          func(context.Context, store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error)
	releasePositionLockFn         func(context.Context, store.ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error)
	createOrderCashReservationFn  func(context.Context, store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error)
	releaseOrderCashReservationFn func(context.Context, store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error)
	releaseCollateralLockFn       func(context.Context, store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error)
	createCollateralLockFn        func(context.Context, store.CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error)
	createIssuanceBatchFn         func(context.Context, store.CreateIssuanceBatchInput) (*domain.IssuanceBatch, *domain.CollateralLock, []*domain.Position, error)
	listCashAccountsFn            func(context.Context, int64) ([]domain.CashAccount, error)
	listPositionsFn               func(context.Context, int64, *int64) ([]domain.Position, error)
	listPositionLocksFn           func(context.Context, int64, *int64) ([]domain.PositionLock, error)
	getCashAccountFn              func(context.Context, int64, string) (*domain.CashAccount, error)
	listLedgerEntriesFn           func(context.Context, int64, string, int) ([]domain.LedgerEntry, error)
	depositCashFn                 func(context.Context, store.DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error)
	withdrawCashFn                func(context.Context, store.WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error)
	listCollateralLocksFn         func(context.Context, int64, string) ([]domain.CollateralLock, error)
	listOrderCashReservationsFn   func(context.Context, int64, string) ([]domain.OrderCashReservation, error)
	listSettlementEntriesByUserFn func(context.Context, int64, *int64, int) ([]domain.SettlementEntry, error)
	listSettlementEntriesByContFn func(context.Context, int64, int) ([]domain.SettlementEntry, error)
	applyExecutionFn              func(context.Context, events.ExecutionCreated) (*store.ExecutionApplicationResult, error)
	settleContractFn              func(context.Context, store.SettleContractInput) (*store.SettlementResult, error)
}

type stubStationCatalog struct {
	findStationFn func(context.Context, string, string) (*domain.WeatherStation, error)
}

func (s stubStationCatalog) FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
	if s.findStationFn != nil {
		return s.findStationFn(ctx, providerName, stationID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ProcessCreateContract(ctx context.Context, envelope commands.CreateContractEnvelope) (commands.CreateContractResult, error) {
	if s.processCreateContractFn != nil {
		return s.processCreateContractFn(ctx, envelope)
	}
	return commands.CreateContractResult{}, nil
}

func (s stubExchangeCoreRepository) FindDuplicateContract(ctx context.Context, input store.FindDuplicateContractInput) (*domain.Contract, error) {
	if s.findDuplicateContractFn != nil {
		return s.findDuplicateContractFn(ctx, input)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) GetContractCommand(ctx context.Context, commandID string) (*commands.ContractCommand, error) {
	if s.getContractCommandFn != nil {
		return s.getContractCommandFn(ctx, commandID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) GetContract(ctx context.Context, contractID int64) (*domain.Contract, error) {
	if s.getContractFn != nil {
		return s.getContractFn(ctx, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error) {
	if s.getContractRuleFn != nil {
		return s.getContractRuleFn(ctx, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error) {
	if s.updateContractStatusFn != nil {
		return s.updateContractStatusFn(ctx, contractID, status)
	}
	return &domain.Contract{ID: contractID, Status: status}, nil
}

func (s stubExchangeCoreRepository) ListContractCollateralLocks(ctx context.Context, userID, contractID int64) ([]domain.CollateralLock, error) {
	if s.listContractCollateralLocksFn != nil {
		return s.listContractCollateralLocksFn(ctx, userID, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListIssuanceBatches(ctx context.Context, userID, contractID int64) ([]domain.IssuanceBatch, error) {
	if s.listIssuanceBatchesFn != nil {
		return s.listIssuanceBatchesFn(ctx, userID, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) CreateIssuanceBatch(ctx context.Context, input store.CreateIssuanceBatchInput) (*domain.IssuanceBatch, *domain.CollateralLock, []*domain.Position, error) {
	if s.createIssuanceBatchFn != nil {
		return s.createIssuanceBatchFn(ctx, input)
	}
	return nil, nil, nil, nil
}

func (s stubExchangeCoreRepository) ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error) {
	if s.listPositionsFn != nil {
		return s.listPositionsFn(ctx, userID, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error) {
	if s.listCashAccountsFn != nil {
		return s.listCashAccountsFn(ctx, userID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error) {
	if s.listPositionLocksFn != nil {
		return s.listPositionLocksFn(ctx, userID, contractID)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error) {
	if s.listSettlementEntriesByContFn != nil {
		return s.listSettlementEntriesByContFn(ctx, contractID, limit)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error) {
	if s.listSettlementEntriesByUserFn != nil {
		return s.listSettlementEntriesByUserFn(ctx, userID, contractID, limit)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) CreatePositionLock(ctx context.Context, input store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	if s.createPositionLockFn != nil {
		return s.createPositionLockFn(ctx, input)
	}
	return &domain.PositionLock{ID: 1, ContractID: input.ContractID, UserID: input.UserID, Side: domain.ClaimSide(input.Side), Quantity: input.Quantity}, nil, nil
}

func (s stubExchangeCoreRepository) ReleasePositionLock(ctx context.Context, input store.ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	if s.releasePositionLockFn != nil {
		return s.releasePositionLockFn(ctx, input)
	}
	return &domain.PositionLock{ID: input.LockID}, nil, nil
}

func (s stubExchangeCoreRepository) GetCashAccount(ctx context.Context, userID int64, currency string) (*domain.CashAccount, error) {
	if s.getCashAccountFn != nil {
		return s.getCashAccountFn(ctx, userID, currency)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) ListLedgerEntries(ctx context.Context, userID int64, currency string, limit int) ([]domain.LedgerEntry, error) {
	if s.listLedgerEntriesFn != nil {
		return s.listLedgerEntriesFn(ctx, userID, currency, limit)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) DepositCash(ctx context.Context, input store.DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	if s.depositCashFn != nil {
		return s.depositCashFn(ctx, input)
	}
	return nil, nil, nil
}

func (s stubExchangeCoreRepository) WithdrawCash(ctx context.Context, input store.WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	if s.withdrawCashFn != nil {
		return s.withdrawCashFn(ctx, input)
	}
	return nil, nil, nil
}

func (s stubExchangeCoreRepository) ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error) {
	if s.listCollateralLocksFn != nil {
		return s.listCollateralLocksFn(ctx, userID, currency)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) CreateCollateralLock(ctx context.Context, input store.CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	if s.createCollateralLockFn != nil {
		return s.createCollateralLockFn(ctx, input)
	}
	return nil, nil, nil, nil
}

func (s stubExchangeCoreRepository) ReleaseCollateralLock(ctx context.Context, input store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	if s.releaseCollateralLockFn != nil {
		return s.releaseCollateralLockFn(ctx, input)
	}
	return &domain.CollateralLock{ID: input.LockID}, nil, nil, nil
}

func (s stubExchangeCoreRepository) ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error) {
	if s.listOrderCashReservationsFn != nil {
		return s.listOrderCashReservationsFn(ctx, userID, currency)
	}
	return nil, nil
}

func (s stubExchangeCoreRepository) CreateOrderCashReservation(ctx context.Context, input store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	if s.createOrderCashReservationFn != nil {
		return s.createOrderCashReservationFn(ctx, input)
	}
	return &domain.OrderCashReservation{ID: 1, UserID: input.UserID, ContractID: input.ContractID, AmountCents: input.AmountCents}, nil, nil, nil
}

func (s stubExchangeCoreRepository) ReleaseOrderCashReservation(ctx context.Context, input store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	if s.releaseOrderCashReservationFn != nil {
		return s.releaseOrderCashReservationFn(ctx, input)
	}
	return &domain.OrderCashReservation{ID: input.ReservationID}, nil, nil, nil
}

func (s stubExchangeCoreRepository) ApplyExecution(ctx context.Context, event events.ExecutionCreated) (*store.ExecutionApplicationResult, error) {
	if s.applyExecutionFn != nil {
		return s.applyExecutionFn(ctx, event)
	}
	return &store.ExecutionApplicationResult{
		ExecutionID:   event.ExecutionID,
		ContractID:    event.ContractID,
		BuyerUserID:   event.BuyerUserID,
		SellerUserID:  event.SellerUserID,
		AffectedUsers: []int64{event.BuyerUserID, event.SellerUserID},
		Applied:       true,
	}, nil
}

func (s stubExchangeCoreRepository) SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
	if s.settleContractFn != nil {
		return s.settleContractFn(ctx, input)
	}
	return &store.SettlementResult{SettledAt: time.Now().UTC()}, nil
}
