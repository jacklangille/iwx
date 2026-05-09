package projector

import (
	"context"
	"fmt"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/projectionbundle"
	"iwx/go_backend/internal/readprojection"
)

type ExchangeCoreReader interface {
	GetContractBundle(ctx context.Context, contractID int64) (*projectionbundle.ContractBundle, error)
	GetUserStateBundle(ctx context.Context, userID int64) (*projectionbundle.UserStateBundle, error)
	GetSettlementBundle(ctx context.Context, contractID int64) (*projectionbundle.SettlementBundle, error)
}

type MatcherReader interface {
	GetMarketBundle(ctx context.Context, contractID int64) (*projectionbundle.MarketBundle, error)
	GetOrderCommandBundle(ctx context.Context, commandID string) (*projectionbundle.OrderCommandBundle, error)
}

type OracleReader interface {
	GetOracleStateBundle(ctx context.Context, contractID int64) (*projectionbundle.OracleStateBundle, error)
	GetStationCatalogBundle(ctx context.Context) (*projectionbundle.StationCatalogBundle, error)
}

type Service struct {
	applier       *readprojection.Applier
	exchangeCore  ExchangeCoreReader
	matcher       MatcherReader
	oracle        OracleReader
	onMarketApply func(contractID int64)
}

func NewService(applier *readprojection.Applier, exchangeCore ExchangeCoreReader, matcher MatcherReader, oracle OracleReader, onMarketApply func(contractID int64)) *Service {
	return &Service{
		applier:       applier,
		exchangeCore:  exchangeCore,
		matcher:       matcher,
		oracle:        oracle,
		onMarketApply: onMarketApply,
	}
}

func (s *Service) HandleProjectionChange(ctx context.Context, event events.ProjectionChange) error {
	if s.applier == nil {
		return nil
	}

	projection, err := s.buildProjection(ctx, event)
	if err != nil {
		return err
	}
	if projection == nil {
		return nil
	}
	if err := s.applier.Apply(ctx, *projection); err != nil {
		return err
	}
	if s.onMarketApply != nil && event.ContractID > 0 {
		s.onMarketApply(event.ContractID)
	}
	return nil
}

func (s *Service) buildProjection(ctx context.Context, event events.ProjectionChange) (*events.ReadModelProjection, error) {
	base := &events.ReadModelProjection{
		EventID:    event.EventID,
		Kind:       event.Kind,
		TraceID:    event.TraceID,
		ContractID: event.ContractID,
		UserID:     event.UserID,
		CommandID:  event.CommandID,
		Version:    event.Version,
		OccurredAt: event.OccurredAt,
	}

	switch event.Kind {
	case events.ProjectionChangeContract:
		bundle, err := s.exchangeCore.GetContractBundle(ctx, event.ContractID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.Contract = bundle.Contract
		}
	case events.ProjectionChangeUserState:
		bundle, err := s.exchangeCore.GetUserStateBundle(ctx, event.UserID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.CashAccounts = bundle.CashAccounts
			base.Positions = bundle.Positions
			base.PositionLocks = bundle.PositionLocks
			base.CollateralLocks = bundle.CollateralLocks
			base.CashReservations = bundle.CashReservations
			base.UserSettlements = bundle.Settlements
		}
	case events.ProjectionChangeSettlement:
		bundle, err := s.exchangeCore.GetSettlementBundle(ctx, event.ContractID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.ContractSettlements = bundle.Entries
		}
	case events.ProjectionChangeMarket:
		bundle, err := s.matcher.GetMarketBundle(ctx, event.ContractID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.Orders = bundle.Orders
			base.Executions = bundle.Executions
			base.Snapshots = bundle.Snapshots
		}
	case events.ProjectionChangeOrderCommand:
		bundle, err := s.matcher.GetOrderCommandBundle(ctx, event.CommandID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.OrderCommand = bundle.Command
		}
	case events.ProjectionChangeOracleState:
		bundle, err := s.oracle.GetOracleStateBundle(ctx, event.ContractID)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.Observations = bundle.Observations
			base.Resolution = bundle.Resolution
		}
	case events.ProjectionChangeStationCatalog:
		bundle, err := s.oracle.GetStationCatalogBundle(ctx)
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			base.Stations = bundle.Stations
		}
	default:
		return nil, fmt.Errorf("unsupported projection change kind: %s", event.Kind)
	}

	return base, nil
}
