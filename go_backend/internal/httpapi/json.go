package httpapi

import (
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/readmodel"
)

type contractSummaryResponse struct {
	ID                      int64    `json:"id"`
	CreatorUserID           *int64   `json:"creator_user_id"`
	AsOf                    *string  `json:"as_of"`
	Sequence                *int64   `json:"sequence"`
	Name                    string   `json:"name"`
	Region                  string   `json:"region"`
	Metric                  string   `json:"metric"`
	Status                  string   `json:"status"`
	Threshold               *int64   `json:"threshold"`
	Multiplier              *int64   `json:"multiplier"`
	MeasurementUnit         *string  `json:"measurement_unit"`
	TradingPeriodStart      *string  `json:"trading_period_start"`
	TradingPeriodEnd        *string  `json:"trading_period_end"`
	MeasurementPeriodStart  *string  `json:"measurement_period_start"`
	MeasurementPeriodEnd    *string  `json:"measurement_period_end"`
	DataProviderName        *string  `json:"data_provider_name"`
	StationID               *string  `json:"station_id"`
	DataProviderStationMode *string  `json:"data_provider_station_mode"`
	Description             *string  `json:"description"`
	BestAboveBid            *string `json:"best_above_bid"`
	BestBelowBid            *string `json:"best_below_bid"`
	MidAbove                *string `json:"mid_above"`
	MidBelow                *string `json:"mid_below"`
	MidPrice                *string `json:"mid_price"`
}

type cashAccountResponse struct {
	ID             int64  `json:"id"`
	UserID         int64  `json:"user_id"`
	Currency       string `json:"currency"`
	AvailableCents int64  `json:"available_cents"`
	LockedCents    int64  `json:"locked_cents"`
	TotalCents     int64  `json:"total_cents"`
	UpdatedAt      string `json:"updated_at"`
}

type positionResponse struct {
	ID                int64  `json:"id"`
	UserID            int64  `json:"user_id"`
	ContractID        int64  `json:"contract_id"`
	Side              string `json:"side"`
	AvailableQuantity int64  `json:"available_quantity"`
	LockedQuantity    int64  `json:"locked_quantity"`
	TotalQuantity     int64  `json:"total_quantity"`
	UpdatedAt         string `json:"updated_at"`
}

type positionLockResponse struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	ContractID    int64   `json:"contract_id"`
	Side          string  `json:"side"`
	Quantity      int64   `json:"quantity"`
	Status        string  `json:"status"`
	OrderID       *int64  `json:"order_id"`
	ReferenceType *string `json:"reference_type"`
	ReferenceID   *string `json:"reference_id"`
	CorrelationID *string `json:"correlation_id"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ReleasedAt    *string `json:"released_at"`
}

type collateralLockResponse struct {
	ID                  int64   `json:"id"`
	UserID              int64   `json:"user_id"`
	ContractID          int64   `json:"contract_id"`
	Currency            string  `json:"currency"`
	AmountCents         int64   `json:"amount_cents"`
	Status              string  `json:"status"`
	ReferenceID         *string `json:"reference_id"`
	Description         *string `json:"description"`
	ReferenceIssuanceID *int64  `json:"reference_issuance_id"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	ReleasedAt          *string `json:"released_at"`
}

type orderCashReservationResponse struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	ContractID    int64   `json:"contract_id"`
	Currency      string  `json:"currency"`
	AmountCents   int64   `json:"amount_cents"`
	Status        string  `json:"status"`
	ReferenceType *string `json:"reference_type"`
	ReferenceID   *string `json:"reference_id"`
	CorrelationID *string `json:"correlation_id"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ReleasedAt    *string `json:"released_at"`
}

type settlementEntryResponse struct {
	ID          int64  `json:"id"`
	ContractID  int64  `json:"contract_id"`
	UserID      int64  `json:"user_id"`
	EntryType   string `json:"entry_type"`
	Outcome     string `json:"outcome"`
	AmountCents int64  `json:"amount_cents"`
	Quantity    int64  `json:"quantity"`
	ReferenceID string `json:"reference_id"`
	CreatedAt   string `json:"created_at"`
}

type executionResponse struct {
	ID           int64  `json:"id"`
	ExecutionID  string `json:"execution_id"`
	ContractID   int64  `json:"contract_id"`
	BuyOrderID   int64  `json:"buy_order_id"`
	SellOrderID  int64  `json:"sell_order_id"`
	BuyerUserID  int64  `json:"buyer_user_id"`
	SellerUserID int64  `json:"seller_user_id"`
	Price        string `json:"price"`
	Quantity     int64  `json:"quantity"`
	Sequence     int64  `json:"sequence"`
	OccurredAt   string `json:"occurred_at"`
}

type oracleObservationResponse struct {
	ID                     int64  `json:"id"`
	ContractID             int64  `json:"contract_id"`
	ProviderName           string `json:"provider_name"`
	StationID              string `json:"station_id"`
	ObservedMetric         string `json:"observed_metric"`
	ObservationWindowStart string `json:"observation_window_start"`
	ObservationWindowEnd   string `json:"observation_window_end"`
	ObservedValue          string `json:"observed_value"`
	NormalizedValue        string `json:"normalized_value"`
	ObservedAt             string `json:"observed_at"`
	RecordedAt             string `json:"recorded_at"`
}

type contractResolutionResponse struct {
	ID                     int64  `json:"id"`
	ContractID             int64  `json:"contract_id"`
	ProviderName           string `json:"provider_name"`
	StationID              string `json:"station_id"`
	ObservedMetric         string `json:"observed_metric"`
	ObservationWindowStart string `json:"observation_window_start"`
	ObservationWindowEnd   string `json:"observation_window_end"`
	RuleVersion            string `json:"rule_version"`
	ResolvedValue          string `json:"resolved_value"`
	Outcome                string `json:"outcome"`
	ResolvedAt             string `json:"resolved_at"`
}

type orderResponse struct {
	ID         int64  `json:"id"`
	ContractID int64  `json:"contract_id"`
	UserID     int64  `json:"user_id"`
	TokenType  string `json:"token_type"`
	OrderSide  string `json:"order_side"`
	Price      string `json:"price"`
	Quantity   int64  `json:"quantity"`
	Status     string `json:"status"`
}

type orderCommandResponse struct {
	CommandID     string  `json:"command_id"`
	ContractID    int64   `json:"contract_id"`
	UserID        int64   `json:"user_id"`
	TokenType     string  `json:"token_type"`
	OrderSide     string  `json:"order_side"`
	Price         string  `json:"price"`
	Quantity      int64   `json:"quantity"`
	Status        string  `json:"status"`
	ErrorMessage  *string `json:"error_message"`
	ResultStatus  *string `json:"result_status"`
	ResultOrderID *int64  `json:"result_order_id"`
	EnqueuedAt    string  `json:"enqueued_at"`
	StartedAt     *string `json:"started_at"`
	CompletedAt   *string `json:"completed_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type priceLevelResponse struct {
	Price    string `json:"price"`
	Quantity int64  `json:"quantity"`
}

type chartPointResponse struct {
	BucketStart string  `json:"bucket_start"`
	InsertedAt  string  `json:"inserted_at"`
	MidAbove    *string `json:"mid_above"`
	MidBelow    *string `json:"mid_below"`
	BestAbove   *string `json:"best_above"`
	BestBelow   *string `json:"best_below"`
}

type chartConfigResponse struct {
	LookbackSeconds int64 `json:"lookback_seconds"`
	BucketSeconds   int64 `json:"bucket_seconds"`
}

type marketSnapshotsResponse struct {
	Config chartConfigResponse `json:"config"`
	Points []chartPointResponse `json:"points"`
}

type marketExecutionsResponse struct {
	ContractID int64               `json:"contract_id"`
	Limit      int                 `json:"limit"`
	Executions []executionResponse `json:"executions"`
}

type marketObservationsResponse struct {
	ContractID   int64                      `json:"contract_id"`
	Observations []oracleObservationResponse `json:"observations"`
}

type marketSettlementsResponse struct {
	ContractID int64                    `json:"contract_id"`
	Entries    []settlementEntryResponse `json:"entries"`
}

type marketStateResponse struct {
	ContractID int64                    `json:"contract_id"`
	Sequence   *int64                   `json:"sequence"`
	AsOf       *string                  `json:"as_of"`
	OrderBook  marketOrderBookResponse  `json:"order_book"`
	Summary    marketSummaryResponse    `json:"summary"`
}

type marketOrderBookResponse struct {
	Above marketBookSideResponse `json:"above"`
	Below marketBookSideResponse `json:"below"`
}

type marketBookSideResponse struct {
	Bid []priceLevelResponse `json:"bid"`
	Ask []priceLevelResponse `json:"ask"`
}

type marketSummaryResponse struct {
	Best              marketBestResponse      `json:"best"`
	Mid               marketMidResponse       `json:"mid"`
	Liquidity         marketLiquidityResponse `json:"liquidity"`
	AboveBelowBidGap  *string                 `json:"above_below_bid_gap"`
}

type marketBestResponse struct {
	Above marketQuoteResponse `json:"above"`
	Below marketQuoteResponse `json:"below"`
}

type marketQuoteResponse struct {
	Bid *string `json:"bid"`
	Ask *string `json:"ask"`
}

type marketMidResponse struct {
	Above *string `json:"above"`
	Below *string `json:"below"`
}

type marketLiquidityResponse struct {
	Above int64 `json:"above"`
	Below int64 `json:"below"`
}

type portfolioResponse struct {
	UserID           int64                        `json:"user_id"`
	Accounts         []cashAccountResponse        `json:"accounts"`
	Positions        []positionResponse           `json:"positions"`
	PositionLocks    []positionLockResponse       `json:"position_locks"`
	CollateralLocks  []collateralLockResponse     `json:"collateral_locks"`
	CashReservations []orderCashReservationResponse `json:"cash_reservations"`
}

func serializeContracts(summaries []readmodel.ContractSummary) []contractSummaryResponse {
	rows := make([]contractSummaryResponse, 0, len(summaries))
	for _, summary := range summaries {
		rows = append(rows, serializeContract(summary))
	}

	return rows
}

func serializeCashAccounts(accounts []domain.CashAccount) []cashAccountResponse {
	rows := make([]cashAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		rows = append(rows, serializeCashAccount(account))
	}
	return rows
}

func serializeCashAccount(account domain.CashAccount) cashAccountResponse {
	return cashAccountResponse{
		ID:             account.ID,
		UserID:         account.UserID,
		Currency:       account.Currency,
		AvailableCents: account.AvailableCents,
		LockedCents:    account.LockedCents,
		TotalCents:     account.TotalCents,
		UpdatedAt:      account.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContract(summary readmodel.ContractSummary) contractSummaryResponse {
	return contractSummaryResponse{
		ID:                      summary.Contract.ID,
		CreatorUserID:           summary.Contract.CreatorUserID,
		AsOf:                    timestampString(summary.Market.AsOf),
		Sequence:                summary.Market.Sequence,
		Name:                    summary.Contract.Name,
		Region:                  summary.Contract.Region,
		Metric:                  summary.Contract.Metric,
		Status:                  summary.Contract.Status,
		Threshold:               summary.Contract.Threshold,
		Multiplier:              summary.Contract.Multiplier,
		MeasurementUnit:         stringOrNil(summary.Contract.MeasurementUnit),
		TradingPeriodStart:      dateString(summary.Contract.TradingPeriodStart),
		TradingPeriodEnd:        dateString(summary.Contract.TradingPeriodEnd),
		MeasurementPeriodStart:  dateString(summary.Contract.MeasurementPeriodStart),
		MeasurementPeriodEnd:    dateString(summary.Contract.MeasurementPeriodEnd),
		DataProviderName:        stringOrNil(summary.Contract.DataProviderName),
		StationID:               stringOrNil(summary.Contract.StationID),
		DataProviderStationMode: stringOrNil(summary.Contract.DataProviderStationMode),
		Description:             stringOrNil(summary.Contract.Description),
		BestAboveBid:            summary.Market.Summary.Best.Above.Bid,
		BestBelowBid:            summary.Market.Summary.Best.Below.Bid,
		MidAbove:                summary.Market.Summary.Mid.Above,
		MidBelow:                summary.Market.Summary.Mid.Below,
		MidPrice:                summary.Market.Summary.MidPrice,
	}
}

func serializeOrders(orders []domain.Order) []orderResponse {
	rows := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		rows = append(rows, serializeOrder(order))
	}
	return rows
}

func serializePositions(positions []domain.Position) []positionResponse {
	rows := make([]positionResponse, 0, len(positions))
	for _, position := range positions {
		rows = append(rows, positionResponse{
			ID:                position.ID,
			UserID:            position.UserID,
			ContractID:        position.ContractID,
			Side:              string(position.Side),
			AvailableQuantity: position.AvailableQuantity,
			LockedQuantity:    position.LockedQuantity,
			TotalQuantity:     position.TotalQuantity,
			UpdatedAt:         position.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return rows
}

func serializePositionLocks(locks []domain.PositionLock) []positionLockResponse {
	rows := make([]positionLockResponse, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, positionLockResponse{
			ID:            lock.ID,
			UserID:        lock.UserID,
			ContractID:    lock.ContractID,
			Side:          string(lock.Side),
			Quantity:      lock.Quantity,
			Status:        string(lock.Status),
			OrderID:       lock.OrderID,
			ReferenceType: stringOrNil(lock.ReferenceType),
			ReferenceID:   stringOrNil(lock.ReferenceID),
			CorrelationID: stringOrNil(lock.CorrelationID),
			Description:   stringOrNil(lock.Description),
			CreatedAt:     lock.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:     lock.UpdatedAt.UTC().Format(time.RFC3339),
			ReleasedAt:    timestampRFC3339(lock.ReleasedAt),
		})
	}
	return rows
}

func serializeCollateralLocks(locks []domain.CollateralLock) []collateralLockResponse {
	rows := make([]collateralLockResponse, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, collateralLockResponse{
			ID:                  lock.ID,
			UserID:              lock.UserID,
			ContractID:          lock.ContractID,
			Currency:            lock.Currency,
			AmountCents:         lock.AmountCents,
			Status:              string(lock.Status),
			ReferenceID:         stringOrNil(lock.ReferenceID),
			Description:         stringOrNil(lock.Description),
			ReferenceIssuanceID: lock.ReferenceIssuanceID,
			CreatedAt:           lock.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:           lock.UpdatedAt.UTC().Format(time.RFC3339),
			ReleasedAt:          timestampRFC3339(lock.ReleasedAt),
		})
	}
	return rows
}

func serializeOrderCashReservations(reservations []domain.OrderCashReservation) []orderCashReservationResponse {
	rows := make([]orderCashReservationResponse, 0, len(reservations))
	for _, reservation := range reservations {
		rows = append(rows, orderCashReservationResponse{
			ID:            reservation.ID,
			UserID:        reservation.UserID,
			ContractID:    reservation.ContractID,
			Currency:      reservation.Currency,
			AmountCents:   reservation.AmountCents,
			Status:        string(reservation.Status),
			ReferenceType: stringOrNil(reservation.ReferenceType),
			ReferenceID:   stringOrNil(reservation.ReferenceID),
			CorrelationID: stringOrNil(reservation.CorrelationID),
			Description:   stringOrNil(reservation.Description),
			CreatedAt:     reservation.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:     reservation.UpdatedAt.UTC().Format(time.RFC3339),
			ReleasedAt:    timestampRFC3339(reservation.ReleasedAt),
		})
	}
	return rows
}

func serializeSettlementEntries(entries []domain.SettlementEntry) []settlementEntryResponse {
	rows := make([]settlementEntryResponse, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, settlementEntryResponse{
			ID:          entry.ID,
			ContractID:  entry.ContractID,
			UserID:      entry.UserID,
			EntryType:   string(entry.EntryType),
			Outcome:     string(entry.Outcome),
			AmountCents: entry.AmountCents,
			Quantity:    entry.Quantity,
			ReferenceID: entry.ReferenceID,
			CreatedAt:   entry.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return rows
}

func serializeExecutions(executions []domain.Execution) []executionResponse {
	rows := make([]executionResponse, 0, len(executions))
	for _, execution := range executions {
		rows = append(rows, serializeExecution(execution))
	}

	return rows
}

func serializeOracleObservations(observations []domain.OracleObservation) []oracleObservationResponse {
	rows := make([]oracleObservationResponse, 0, len(observations))
	for _, observation := range observations {
		rows = append(rows, serializeOracleObservation(observation))
	}
	return rows
}

func serializeOracleObservation(observation domain.OracleObservation) oracleObservationResponse {
	return oracleObservationResponse{
		ID:                     observation.ID,
		ContractID:             observation.ContractID,
		ProviderName:           observation.ProviderName,
		StationID:              observation.StationID,
		ObservedMetric:         observation.ObservedMetric,
		ObservationWindowStart: observation.ObservationWindowStart.UTC().Format(time.RFC3339),
		ObservationWindowEnd:   observation.ObservationWindowEnd.UTC().Format(time.RFC3339),
		ObservedValue:          observation.ObservedValue,
		NormalizedValue:        observation.NormalizedValue,
		ObservedAt:             observation.ObservedAt.UTC().Format(time.RFC3339),
		RecordedAt:             observation.RecordedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContractResolution(resolution domain.ContractResolution) contractResolutionResponse {
	return contractResolutionResponse{
		ID:                     resolution.ID,
		ContractID:             resolution.ContractID,
		ProviderName:           resolution.ProviderName,
		StationID:              resolution.StationID,
		ObservedMetric:         resolution.ObservedMetric,
		ObservationWindowStart: resolution.ObservationWindowStart.UTC().Format(time.RFC3339),
		ObservationWindowEnd:   resolution.ObservationWindowEnd.UTC().Format(time.RFC3339),
		RuleVersion:            resolution.RuleVersion,
		ResolvedValue:          resolution.ResolvedValue,
		Outcome:                string(resolution.Outcome),
		ResolvedAt:             resolution.ResolvedAt.UTC().Format(time.RFC3339),
	}
}

func serializeExecution(execution domain.Execution) executionResponse {
	return executionResponse{
		ID:           execution.ID,
		ExecutionID:  execution.ExecutionID,
		ContractID:   execution.ContractID,
		BuyOrderID:   execution.BuyOrderID,
		SellOrderID:  execution.SellOrderID,
		BuyerUserID:  execution.BuyerUserID,
		SellerUserID: execution.SellerUserID,
		Price:        execution.Price,
		Quantity:     execution.Quantity,
		Sequence:     execution.Sequence,
		OccurredAt:   execution.OccurredAt.UTC().Format(time.RFC3339),
	}
}

func serializeOrder(order domain.Order) orderResponse {
	return orderResponse{
		ID:         order.ID,
		ContractID: order.ContractID,
		UserID:     order.UserID,
		TokenType:  order.TokenType,
		OrderSide:  order.OrderSide,
		Price:      order.Price,
		Quantity:   order.Quantity,
		Status:     order.Status,
	}
}

func serializePlaceOrderAccepted(result commands.PlaceOrderAccepted) placeOrderAcceptedResponse {
	return placeOrderAcceptedResponse{
		CommandID:  result.CommandID,
		ContractID: result.ContractID,
		Partition:  result.Partition,
		Status:     result.Status,
		EnqueuedAt: result.EnqueuedAt.UTC().Format(time.RFC3339),
	}
}

type placeOrderAcceptedResponse struct {
	CommandID  string `json:"command_id"`
	ContractID int64  `json:"contract_id"`
	Partition  int    `json:"partition"`
	Status     string `json:"status"`
	EnqueuedAt string `json:"enqueued_at"`
}

func serializeCreateContractAccepted(result commands.CreateContractAccepted) createContractAcceptedResponse {
	return createContractAcceptedResponse{
		CommandID:  result.CommandID,
		Partition:  result.Partition,
		Status:     result.Status,
		EnqueuedAt: result.EnqueuedAt.UTC().Format(time.RFC3339),
	}
}

type createContractAcceptedResponse struct {
	CommandID  string `json:"command_id"`
	Partition  int    `json:"partition"`
	Status     string `json:"status"`
	EnqueuedAt string `json:"enqueued_at"`
}

func serializeOrderCommand(command commands.OrderCommand) orderCommandResponse {
	return orderCommandResponse{
		CommandID:     command.CommandID,
		ContractID:    command.ContractID,
		UserID:        command.UserID,
		TokenType:     command.TokenType,
		OrderSide:     command.OrderSide,
		Price:         command.Price,
		Quantity:      command.Quantity,
		Status:        command.Status,
		ErrorMessage:  command.ErrorMessage,
		ResultStatus:  command.ResultStatus,
		ResultOrderID: command.ResultOrderID,
		EnqueuedAt:    command.EnqueuedAt.UTC().Format(time.RFC3339),
		StartedAt:     timestampRFC3339(command.StartedAt),
		CompletedAt:   timestampRFC3339(command.CompletedAt),
		UpdatedAt:     command.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContractCommand(command commands.ContractCommand) contractCommandResponse {
	return contractCommandResponse{
		CommandID:                 command.CommandID,
		CreatorUserID:             command.CreatorUserID,
		Name:                      command.Name,
		Region:                    command.Region,
		Metric:                    command.Metric,
		Status:                    command.Status,
		Threshold:                 command.Threshold,
		Multiplier:                command.Multiplier,
		MeasurementUnit:           command.MeasurementUnit,
		TradingPeriodStart:        dateString(command.TradingPeriodStart),
		TradingPeriodEnd:          dateString(command.TradingPeriodEnd),
		MeasurementPeriodStart:    dateString(command.MeasurementPeriodStart),
		MeasurementPeriodEnd:      dateString(command.MeasurementPeriodEnd),
		DataProviderName:          command.DataProviderName,
		StationID:                 command.StationID,
		DataProviderStationMode:   command.DataProviderStationMode,
		Description:               command.Description,
		CommandStatus:             command.CommandStatus,
		ErrorMessage:              command.ErrorMessage,
		ResultContractID:          command.ResultContractID,
		EnqueuedAt:                command.EnqueuedAt.UTC().Format(time.RFC3339),
		StartedAt:                 timestampRFC3339(command.StartedAt),
		CompletedAt:               timestampRFC3339(command.CompletedAt),
		UpdatedAt:                 command.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type contractCommandResponse struct {
	CommandID               string  `json:"command_id"`
	CreatorUserID           int64   `json:"creator_user_id"`
	Name                    string  `json:"name"`
	Region                  string  `json:"region"`
	Metric                  string  `json:"metric"`
	Status                  string  `json:"status"`
	Threshold               *int64  `json:"threshold"`
	Multiplier              *int64  `json:"multiplier"`
	MeasurementUnit         *string `json:"measurement_unit"`
	TradingPeriodStart      *string `json:"trading_period_start"`
	TradingPeriodEnd        *string `json:"trading_period_end"`
	MeasurementPeriodStart  *string `json:"measurement_period_start"`
	MeasurementPeriodEnd    *string `json:"measurement_period_end"`
	DataProviderName        *string `json:"data_provider_name"`
	StationID               *string `json:"station_id"`
	DataProviderStationMode *string `json:"data_provider_station_mode"`
	Description             *string `json:"description"`
	CommandStatus           string  `json:"command_status"`
	ErrorMessage            *string `json:"error_message"`
	ResultContractID        *int64  `json:"result_contract_id"`
	EnqueuedAt              string  `json:"enqueued_at"`
	StartedAt               *string `json:"started_at"`
	CompletedAt             *string `json:"completed_at"`
	UpdatedAt               string  `json:"updated_at"`
}

func serializeMarketState(marketState domain.MarketState) marketStateResponse {
	return marketStateResponse{
		ContractID: marketState.ContractID,
		Sequence:   marketState.Sequence,
		AsOf:       timestampString(marketState.AsOf),
		OrderBook: marketOrderBookResponse{
			Above: marketBookSideResponse{
				Bid: serializeLevels(marketState.OrderBook.Above.Bid),
				Ask: serializeLevels(marketState.OrderBook.Above.Ask),
			},
			Below: marketBookSideResponse{
				Bid: serializeLevels(marketState.OrderBook.Below.Bid),
				Ask: serializeLevels(marketState.OrderBook.Below.Ask),
			},
		},
		Summary: marketSummaryResponse{
			Best: marketBestResponse{
				Above: marketQuoteResponse{
					Bid: marketState.Summary.Best.Above.Bid,
					Ask: marketState.Summary.Best.Above.Ask,
				},
				Below: marketQuoteResponse{
					Bid: marketState.Summary.Best.Below.Bid,
					Ask: marketState.Summary.Best.Below.Ask,
				},
			},
			Mid: marketMidResponse{
				Above: marketState.Summary.Mid.Above,
				Below: marketState.Summary.Mid.Below,
			},
			Liquidity: marketLiquidityResponse{
				Above: marketState.Summary.Liquidity.Above,
				Below: marketState.Summary.Liquidity.Below,
			},
			AboveBelowBidGap: marketState.Summary.AboveBelowBidGap,
		},
	}
}

func serializeChartPoints(points []domain.ChartPoint) []chartPointResponse {
	rows := make([]chartPointResponse, 0, len(points))
	for _, point := range points {
		rows = append(rows, chartPointResponse{
			BucketStart: point.BucketStart.UTC().Format(time.RFC3339[:19]),
			InsertedAt:  point.InsertedAt.UTC().Format(time.RFC3339[:19]),
			MidAbove:    point.MidAbove,
			MidBelow:    point.MidBelow,
			BestAbove:   point.BestAbove,
			BestBelow:   point.BestBelow,
		})
	}

	return rows
}

func serializeLevels(levels []domain.PriceLevel) []priceLevelResponse {
	rows := make([]priceLevelResponse, 0, len(levels))
	for _, level := range levels {
		rows = append(rows, priceLevelResponse{
			Price:    level.Price,
			Quantity: level.Quantity,
		})
	}

	return rows
}

func timestampString(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format("2006-01-02T15:04:05")
	return &formatted
}

func dateString(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format("2006-01-02")
	return &formatted
}

func timestampRFC3339(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func stringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}
