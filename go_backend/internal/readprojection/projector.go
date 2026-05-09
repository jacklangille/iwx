package readprojection

import (
	"context"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/store"
)

type Projector struct {
	contractSource     store.ContractProjectionSource
	contractTarget     store.ContractProjectionTarget
	userSource         store.UserProjectionSource
	userTarget         store.UserProjectionTarget
	oracleSource       store.OracleProjectionSource
	oracleTarget       store.OracleProjectionTarget
	settlementSource   store.SettlementProjectionSource
	settlementTarget   store.SettlementProjectionTarget
	orderSource        store.OrderRepository
	orderTarget        store.OrderProjectionTarget
	executionSource    store.ExecutionProjectionSource
	executionTarget    store.ExecutionProjectionTarget
	snapshotSource     store.SnapshotProjectionSource
	snapshotTarget     store.SnapshotProjectionTarget
	orderCommandSource store.OrderCommandRepository
	orderCommandTarget store.OrderCommandProjectionTarget
}

func NewProjector(
	contractSource store.ContractProjectionSource,
	contractTarget store.ContractProjectionTarget,
	userSource store.UserProjectionSource,
	userTarget store.UserProjectionTarget,
	oracleSource store.OracleProjectionSource,
	oracleTarget store.OracleProjectionTarget,
	settlementSource store.SettlementProjectionSource,
	settlementTarget store.SettlementProjectionTarget,
	orderSource store.OrderRepository,
	orderTarget store.OrderProjectionTarget,
	executionSource store.ExecutionProjectionSource,
	executionTarget store.ExecutionProjectionTarget,
	snapshotSource store.SnapshotProjectionSource,
	snapshotTarget store.SnapshotProjectionTarget,
	orderCommandSource store.OrderCommandRepository,
	orderCommandTarget store.OrderCommandProjectionTarget,
) *Projector {
	return &Projector{
		contractSource:     contractSource,
		contractTarget:     contractTarget,
		userSource:         userSource,
		userTarget:         userTarget,
		oracleSource:       oracleSource,
		oracleTarget:       oracleTarget,
		settlementSource:   settlementSource,
		settlementTarget:   settlementTarget,
		orderSource:        orderSource,
		orderTarget:        orderTarget,
		executionSource:    executionSource,
		executionTarget:    executionTarget,
		snapshotSource:     snapshotSource,
		snapshotTarget:     snapshotTarget,
		orderCommandSource: orderCommandSource,
		orderCommandTarget: orderCommandTarget,
	}
}

func (p *Projector) ProjectContract(ctx context.Context, contractID int64) error {
	if p == nil || p.contractSource == nil || p.contractTarget == nil {
		return nil
	}

	contract, err := p.contractSource.GetContract(ctx, contractID)
	if err != nil || contract == nil {
		return err
	}

	return p.contractTarget.UpsertContractProjection(ctx, *contract)
}

func (p *Projector) ProjectUserState(ctx context.Context, userID int64) error {
	if p == nil || p.userSource == nil || p.userTarget == nil || userID <= 0 {
		return nil
	}

	accounts, err := p.userSource.ListCashAccounts(ctx, userID)
	if err != nil {
		return err
	}
	if err := p.userTarget.ReplaceCashAccountsProjection(ctx, userID, accounts); err != nil {
		return err
	}

	positions, err := p.userSource.ListPositions(ctx, userID, nil)
	if err != nil {
		return err
	}
	if err := p.userTarget.ReplacePositionsProjection(ctx, userID, positions); err != nil {
		return err
	}

	locks, err := p.userSource.ListPositionLocks(ctx, userID, nil)
	if err != nil {
		return err
	}
	if err := p.userTarget.ReplacePositionLocksProjection(ctx, userID, locks); err != nil {
		return err
	}

	for _, account := range accounts {
		collateralLocks, err := p.userSource.ListCollateralLocks(ctx, userID, account.Currency)
		if err != nil {
			return err
		}
		if err := p.userTarget.ReplaceCollateralLocksProjection(ctx, userID, account.Currency, collateralLocks); err != nil {
			return err
		}

		cashReservations, err := p.userSource.ListOrderCashReservations(ctx, userID, account.Currency)
		if err != nil {
			return err
		}
		if err := p.userTarget.ReplaceOrderCashReservationsProjection(ctx, userID, account.Currency, cashReservations); err != nil {
			return err
		}
	}

	settlementEntries, err := p.userSource.ListSettlementEntriesByUser(ctx, userID, nil, 500)
	if err != nil {
		return err
	}
	if err := p.userTarget.ReplaceUserSettlementEntriesProjection(ctx, userID, settlementEntries); err != nil {
		return err
	}

	return nil
}

func (p *Projector) ProjectMarket(ctx context.Context, contractID int64) error {
	if p == nil {
		return nil
	}

	if p.orderSource != nil && p.orderTarget != nil {
		orders, err := p.orderSource.ListOpenOrders(ctx, &contractID)
		if err != nil {
			return err
		}
		if err := p.orderTarget.ReplaceOpenOrdersProjection(ctx, contractID, orders); err != nil {
			return err
		}
	}

	if p.executionSource != nil && p.executionTarget != nil {
		executions, err := p.executionSource.ListExecutions(ctx, contractID, 200)
		if err != nil {
			return err
		}
		if err := p.executionTarget.ReplaceExecutionsProjection(ctx, contractID, executions); err != nil {
			return err
		}
	}

	if p.snapshotSource != nil && p.snapshotTarget != nil {
		snapshots, err := p.snapshotSource.ListAllSnapshots(ctx, contractID)
		if err != nil {
			return err
		}
		if err := p.snapshotTarget.ReplaceSnapshotsProjection(ctx, contractID, snapshots); err != nil {
			return err
		}
	}

	return nil
}

func (p *Projector) ProjectOracleState(ctx context.Context, contractID int64) error {
	if p == nil || p.oracleSource == nil || p.oracleTarget == nil {
		return nil
	}

	observations, err := p.oracleSource.ListObservations(ctx, contractID, 500)
	if err != nil {
		return err
	}
	if err := p.oracleTarget.ReplaceObservationsProjection(ctx, contractID, observations); err != nil {
		return err
	}

	resolution, err := p.oracleSource.GetLatestResolution(ctx, contractID)
	if err != nil {
		return err
	}
	if resolution != nil {
		if err := p.oracleTarget.UpsertResolutionProjection(ctx, *resolution); err != nil {
			return err
		}
	}

	return nil
}

func (p *Projector) ProjectStationCatalog(ctx context.Context) error {
	if p == nil || p.oracleSource == nil || p.oracleTarget == nil {
		return nil
	}

	stations, err := p.oracleSource.ListStations(ctx, false)
	if err != nil {
		return err
	}

	return p.oracleTarget.ReplaceStationsProjection(ctx, stations)
}

func (p *Projector) ProjectSettlementState(ctx context.Context, contractID int64) error {
	if p == nil || p.settlementSource == nil || p.settlementTarget == nil {
		return nil
	}

	entries, err := p.settlementSource.ListSettlementEntriesByContract(ctx, contractID, 500)
	if err != nil {
		return err
	}
	return p.settlementTarget.ReplaceSettlementEntriesProjection(ctx, contractID, entries)
}

func (p *Projector) ProjectOrderCommand(ctx context.Context, commandID string) error {
	if p == nil || p.orderCommandSource == nil || p.orderCommandTarget == nil {
		return nil
	}

	command, err := p.orderCommandSource.GetOrderCommand(ctx, commandID)
	if err != nil || command == nil {
		return err
	}

	return p.orderCommandTarget.UpsertOrderCommandProjection(ctx, *command)
}

func ProjectOrderResultContractID(result commands.PlaceOrderResult) int64 {
	return result.ContractID
}
