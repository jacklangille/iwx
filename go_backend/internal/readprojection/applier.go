package readprojection

import (
	"context"
	"strconv"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/store"
)

type CheckpointStore interface {
	ShouldApply(ctx context.Context, key, eventID string, version int64) (bool, error)
	RecordApplied(ctx context.Context, key, eventID string, version int64) error
}

type Applier struct {
	checkpoints        CheckpointStore
	contractTarget     store.ContractProjectionTarget
	userTarget         store.UserProjectionTarget
	oracleTarget       store.OracleProjectionTarget
	settlementTarget   store.SettlementProjectionTarget
	orderTarget        store.OrderProjectionTarget
	executionTarget    store.ExecutionProjectionTarget
	snapshotTarget     store.SnapshotProjectionTarget
	orderCommandTarget store.OrderCommandProjectionTarget
}

func NewApplier(
	checkpoints CheckpointStore,
	contractTarget store.ContractProjectionTarget,
	userTarget store.UserProjectionTarget,
	oracleTarget store.OracleProjectionTarget,
	settlementTarget store.SettlementProjectionTarget,
	orderTarget store.OrderProjectionTarget,
	executionTarget store.ExecutionProjectionTarget,
	snapshotTarget store.SnapshotProjectionTarget,
	orderCommandTarget store.OrderCommandProjectionTarget,
) *Applier {
	return &Applier{
		checkpoints:        checkpoints,
		contractTarget:     contractTarget,
		userTarget:         userTarget,
		oracleTarget:       oracleTarget,
		settlementTarget:   settlementTarget,
		orderTarget:        orderTarget,
		executionTarget:    executionTarget,
		snapshotTarget:     snapshotTarget,
		orderCommandTarget: orderCommandTarget,
	}
}

func (a *Applier) Apply(ctx context.Context, event events.ReadModelProjection) error {
	key := projectionCheckpointKey(event)
	if a.checkpoints != nil && key != "" {
		shouldApply, err := a.checkpoints.ShouldApply(ctx, key, event.EventID, event.Version)
		if err != nil {
			return err
		}
		if !shouldApply {
			return nil
		}
	}

	var err error
	switch event.Kind {
	case events.ProjectionChangeContract:
		if a.contractTarget == nil || event.Contract == nil {
			return nil
		}
		err = a.contractTarget.UpsertContractProjection(ctx, *event.Contract)
	case events.ProjectionChangeUserState:
		if a.userTarget == nil || event.UserID <= 0 {
			return nil
		}
		if err = a.userTarget.ReplaceCashAccountsProjection(ctx, event.UserID, event.CashAccounts); err != nil {
			return err
		}
		if err = a.userTarget.ReplacePositionsProjection(ctx, event.UserID, event.Positions); err != nil {
			return err
		}
		if err = a.userTarget.ReplacePositionLocksProjection(ctx, event.UserID, event.PositionLocks); err != nil {
			return err
		}
		for currency, locks := range groupCollateralLocksByCurrency(event.CollateralLocks) {
			if err = a.userTarget.ReplaceCollateralLocksProjection(ctx, event.UserID, currency, locks); err != nil {
				return err
			}
		}
		for currency, reservations := range groupCashReservationsByCurrency(event.CashReservations) {
			if err = a.userTarget.ReplaceOrderCashReservationsProjection(ctx, event.UserID, currency, reservations); err != nil {
				return err
			}
		}
		err = a.userTarget.ReplaceUserSettlementEntriesProjection(ctx, event.UserID, event.UserSettlements)
	case events.ProjectionChangeOracleState:
		if a.oracleTarget == nil || event.ContractID <= 0 {
			return nil
		}
		if err = a.oracleTarget.ReplaceObservationsProjection(ctx, event.ContractID, event.Observations); err != nil {
			return err
		}
		if event.Resolution != nil {
			if err = a.oracleTarget.UpsertResolutionProjection(ctx, *event.Resolution); err != nil {
				return err
			}
		}
	case events.ProjectionChangeStationCatalog:
		if a.oracleTarget == nil {
			return nil
		}
		err = a.oracleTarget.ReplaceStationsProjection(ctx, event.Stations)
	case events.ProjectionChangeSettlement:
		if a.settlementTarget == nil || event.ContractID <= 0 {
			return nil
		}
		err = a.settlementTarget.ReplaceSettlementEntriesProjection(ctx, event.ContractID, event.ContractSettlements)
	case events.ProjectionChangeMarket:
		if event.ContractID <= 0 {
			return nil
		}
		if a.orderTarget != nil {
			if err = a.orderTarget.ReplaceOpenOrdersProjection(ctx, event.ContractID, event.Orders); err != nil {
				return err
			}
		}
		if a.executionTarget != nil {
			if err = a.executionTarget.ReplaceExecutionsProjection(ctx, event.ContractID, event.Executions); err != nil {
				return err
			}
		}
		if a.snapshotTarget != nil {
			if err = a.snapshotTarget.ReplaceSnapshotsProjection(ctx, event.ContractID, event.Snapshots); err != nil {
				return err
			}
		}
	case events.ProjectionChangeOrderCommand:
		if a.orderCommandTarget == nil || event.OrderCommand == nil {
			return nil
		}
		err = a.orderCommandTarget.UpsertOrderCommandProjection(ctx, *event.OrderCommand)
	default:
		return nil
	}

	if err != nil {
		return err
	}
	if a.checkpoints != nil && key != "" {
		return a.checkpoints.RecordApplied(ctx, key, event.EventID, event.Version)
	}
	return nil
}

func groupCollateralLocksByCurrency(locks []domain.CollateralLock) map[string][]domain.CollateralLock {
	grouped := map[string][]domain.CollateralLock{}
	for _, lock := range locks {
		grouped[lock.Currency] = append(grouped[lock.Currency], lock)
	}
	return grouped
}

func groupCashReservationsByCurrency(reservations []domain.OrderCashReservation) map[string][]domain.OrderCashReservation {
	grouped := map[string][]domain.OrderCashReservation{}
	for _, reservation := range reservations {
		grouped[reservation.Currency] = append(grouped[reservation.Currency], reservation)
	}
	return grouped
}

func projectionCheckpointKey(event events.ReadModelProjection) string {
	switch event.Kind {
	case events.ProjectionChangeContract:
		return "contract:" + strconv.FormatInt(event.ContractID, 10)
	case events.ProjectionChangeUserState:
		return "user:" + strconv.FormatInt(event.UserID, 10)
	case events.ProjectionChangeOracleState:
		return "oracle:" + strconv.FormatInt(event.ContractID, 10)
	case events.ProjectionChangeStationCatalog:
		return "stations"
	case events.ProjectionChangeSettlement:
		return "settlement:" + strconv.FormatInt(event.ContractID, 10)
	case events.ProjectionChangeMarket:
		return "market:" + strconv.FormatInt(event.ContractID, 10)
	case events.ProjectionChangeOrderCommand:
		if event.CommandID == "" && event.OrderCommand == nil {
			return ""
		}
		if event.CommandID != "" {
			return "order_command:" + event.CommandID
		}
		return "order_command:" + event.OrderCommand.CommandID
	default:
		return ""
	}
}
