package readmodel

import (
	"context"
	"errors"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/market"
	"iwx/go_backend/internal/store"
)

var ErrContractNotFound = errors.New("contract not found")

type Service struct {
	contracts     store.ContractRepository
	users         store.UserViewRepository
	oracles       store.OracleRepository
	settlements   store.SettlementRepository
	orders        store.OrderRepository
	executions    store.ExecutionRepository
	snapshots     store.SnapshotRepository
	orderCommands store.OrderCommandRepository
	cache         *marketCache
}

func NewService(
	contracts store.ContractRepository,
	users store.UserViewRepository,
	oracles store.OracleRepository,
	settlements store.SettlementRepository,
	orders store.OrderRepository,
	executions store.ExecutionRepository,
	snapshots store.SnapshotRepository,
	orderCommands store.OrderCommandRepository,
) *Service {
	return &Service{
		contracts:     contracts,
		users:         users,
		oracles:       oracles,
		settlements:   settlements,
		orders:        orders,
		executions:    executions,
		snapshots:     snapshots,
		orderCommands: orderCommands,
		cache:         newMarketCache(2 * time.Second),
	}
}

func (s *Service) ListContractSummaries(ctx context.Context) ([]ContractSummary, error) {
	contracts, err := s.contracts.ListContracts(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]ContractSummary, 0, len(contracts))
	for _, contract := range contracts {
		summary, err := s.marketSummary(ctx, contract.ID)
		if err != nil {
			return nil, err
		}

		asOf, err := s.snapshots.LatestTimestamp(ctx, contract.ID)
		if err != nil {
			return nil, err
		}

		sequence, err := s.snapshots.LatestSequence(ctx, contract.ID)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, ContractSummary{
			Contract: contract,
			Market: ContractMarket{
				AsOf:     asOf,
				Sequence: sequence,
				Summary:  summary,
			},
		})
	}

	return summaries, nil
}

func (s *Service) ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error) {
	if s.oracles == nil {
		return []domain.WeatherStation{}, nil
	}

	return s.oracles.ListStations(ctx, activeOnly)
}

func (s *Service) ListOpenOrders(ctx context.Context, contractID *int64) ([]domain.Order, error) {
	return s.orders.ListOpenOrders(ctx, contractID)
}

func (s *Service) ListExecutions(ctx context.Context, contractID int64, limit int) ([]domain.Execution, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return nil, err
	}

	if s.executions == nil {
		return []domain.Execution{}, nil
	}

	if executions, ok := s.cache.getExecutions(contractID, limit); ok {
		return executions, nil
	}

	executions, err := s.executions.ListExecutions(ctx, contractID, limit)
	if err != nil {
		return nil, err
	}
	s.cache.setExecutions(contractID, limit, executions)
	return executions, nil
}

func (s *Service) MarketState(ctx context.Context, contractID int64) (domain.MarketState, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return domain.MarketState{}, err
	}

	if state, ok := s.cache.getState(contractID); ok {
		return state, nil
	}

	orders, err := s.orders.ListOpenOrders(ctx, &contractID)
	if err != nil {
		return domain.MarketState{}, err
	}

	components := market.BuildComponents(orders)
	asOf, err := s.snapshots.LatestTimestamp(ctx, contractID)
	if err != nil {
		return domain.MarketState{}, err
	}

	sequence, err := s.snapshots.LatestSequence(ctx, contractID)
	if err != nil {
		return domain.MarketState{}, err
	}

	state := domain.MarketState{
		ContractID: contractID,
		Sequence:   sequence,
		AsOf:       asOf,
		OrderBook:  components.OrderBook,
		Summary: domain.MarketSummary{
			Best:             components.Best,
			Mid:              components.Mid,
			Liquidity:        market.LiquidityTotals(orders),
			AboveBelowBidGap: market.AboveBelowBidGap(components.Best),
		},
	}
	s.cache.setState(contractID, state)
	return state, nil
}

func (s *Service) ListMarketChartSeries(
	ctx context.Context,
	contractID int64,
	config domain.ChartConfig,
) ([]domain.ChartPoint, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return nil, err
	}

	windowEnd := time.Now().UTC().Truncate(time.Second)
	windowStart := windowEnd.Add(-time.Duration(config.LookbackSeconds) * time.Second)
	firstBucket := market.BucketStart(windowStart, config.BucketSeconds)
	lastBucket := market.BucketStart(windowEnd, config.BucketSeconds)

	snapshots, err := s.snapshots.ListSnapshotsSince(ctx, contractID, windowStart)
	if err != nil {
		return nil, err
	}

	return market.BucketSnapshots(snapshots, config.BucketSeconds, firstBucket, lastBucket), nil
}

func (s *Service) GetOrderCommand(ctx context.Context, commandID string) (*commands.OrderCommand, error) {
	if s.orderCommands == nil {
		return nil, nil
	}

	return s.orderCommands.GetOrderCommand(ctx, commandID)
}

func (s *Service) ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error) {
	if s.users == nil {
		return []domain.CashAccount{}, nil
	}

	return s.users.ListCashAccounts(ctx, userID)
}

func (s *Service) ListUserPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error) {
	if s.users == nil {
		return []domain.Position{}, nil
	}

	return s.users.ListPositions(ctx, userID, contractID)
}

func (s *Service) ListUserPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error) {
	if s.users == nil {
		return []domain.PositionLock{}, nil
	}

	return s.users.ListPositionLocks(ctx, userID, contractID)
}

func (s *Service) ListUserCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error) {
	if s.users == nil {
		return []domain.CollateralLock{}, nil
	}

	return s.users.ListCollateralLocks(ctx, userID, currency)
}

func (s *Service) ListUserCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error) {
	if s.users == nil {
		return []domain.OrderCashReservation{}, nil
	}

	return s.users.ListOrderCashReservations(ctx, userID, currency)
}

func (s *Service) ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return nil, err
	}
	if s.oracles == nil {
		return []domain.OracleObservation{}, nil
	}

	return s.oracles.ListObservations(ctx, contractID, limit)
}

func (s *Service) GetResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return nil, err
	}
	if s.oracles == nil {
		return nil, nil
	}

	return s.oracles.GetLatestResolution(ctx, contractID)
}

func (s *Service) ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error) {
	if err := s.ensureContractExists(ctx, contractID); err != nil {
		return nil, err
	}
	if s.settlements == nil {
		return []domain.SettlementEntry{}, nil
	}

	return s.settlements.ListSettlementEntriesByContract(ctx, contractID, limit)
}

func (s *Service) ListUserSettlementEntries(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error) {
	if s.settlements == nil {
		return []domain.SettlementEntry{}, nil
	}

	return s.settlements.ListSettlementEntriesByUser(ctx, userID, contractID, limit)
}

func (s *Service) marketSummary(ctx context.Context, contractID int64) (domain.MarketSummary, error) {
	orders, err := s.orders.ListOpenOrders(ctx, &contractID)
	if err != nil {
		return domain.MarketSummary{}, err
	}

	return market.Summary(orders), nil
}

func (s *Service) ensureContractExists(ctx context.Context, contractID int64) error {
	exists, err := s.contracts.ContractExists(ctx, contractID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrContractNotFound
	}

	return nil
}

func (s *Service) InvalidateMarket(contractID int64) {
	if s == nil || s.cache == nil || contractID <= 0 {
		return
	}

	s.cache.invalidateContract(contractID)
}
