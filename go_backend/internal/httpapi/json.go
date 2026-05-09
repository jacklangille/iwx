package httpapi

import (
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/readmodel"
)

func serializeContracts(summaries []readmodel.ContractSummary) []map[string]any {
	rows := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		rows = append(rows, serializeContract(summary))
	}

	return rows
}

func serializeCashAccounts(accounts []domain.CashAccount) []map[string]any {
	rows := make([]map[string]any, 0, len(accounts))
	for _, account := range accounts {
		rows = append(rows, serializeCashAccount(account))
	}
	return rows
}

func serializeCashAccount(account domain.CashAccount) map[string]any {
	return map[string]any{
		"id":              account.ID,
		"user_id":         account.UserID,
		"currency":        account.Currency,
		"available_cents": account.AvailableCents,
		"locked_cents":    account.LockedCents,
		"total_cents":     account.TotalCents,
		"updated_at":      account.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContract(summary readmodel.ContractSummary) map[string]any {
	return map[string]any{
		"id":                         summary.Contract.ID,
		"creator_user_id":            summary.Contract.CreatorUserID,
		"as_of":                      timestampString(summary.Market.AsOf),
		"sequence":                   summary.Market.Sequence,
		"name":                       summary.Contract.Name,
		"region":                     summary.Contract.Region,
		"metric":                     summary.Contract.Metric,
		"status":                     summary.Contract.Status,
		"threshold":                  summary.Contract.Threshold,
		"multiplier":                 summary.Contract.Multiplier,
		"measurement_unit":           blankToNil(summary.Contract.MeasurementUnit),
		"trading_period_start":       dateString(summary.Contract.TradingPeriodStart),
		"trading_period_end":         dateString(summary.Contract.TradingPeriodEnd),
		"measurement_period_start":   dateString(summary.Contract.MeasurementPeriodStart),
		"measurement_period_end":     dateString(summary.Contract.MeasurementPeriodEnd),
		"data_provider_name":         blankToNil(summary.Contract.DataProviderName),
		"station_id":                 blankToNil(summary.Contract.StationID),
		"data_provider_station_mode": blankToNil(summary.Contract.DataProviderStationMode),
		"description":                blankToNil(summary.Contract.Description),
		"best_above_bid":             summary.Market.Summary.Best.Above.Bid,
		"best_below_bid":             summary.Market.Summary.Best.Below.Bid,
		"mid_above":                  summary.Market.Summary.Mid.Above,
		"mid_below":                  summary.Market.Summary.Mid.Below,
		"mid_price":                  summary.Market.Summary.MidPrice,
	}
}

func serializeOrders(orders []domain.Order) []map[string]any {
	rows := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		rows = append(rows, serializeOrder(order))
	}

	return rows
}

func serializePositions(positions []domain.Position) []map[string]any {
	rows := make([]map[string]any, 0, len(positions))
	for _, position := range positions {
		rows = append(rows, map[string]any{
			"id":                 position.ID,
			"user_id":            position.UserID,
			"contract_id":        position.ContractID,
			"side":               position.Side,
			"available_quantity": position.AvailableQuantity,
			"locked_quantity":    position.LockedQuantity,
			"total_quantity":     position.TotalQuantity,
			"updated_at":         position.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return rows
}

func serializePositionLocks(locks []domain.PositionLock) []map[string]any {
	rows := make([]map[string]any, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, map[string]any{
			"id":             lock.ID,
			"user_id":        lock.UserID,
			"contract_id":    lock.ContractID,
			"side":           lock.Side,
			"quantity":       lock.Quantity,
			"status":         lock.Status,
			"order_id":       lock.OrderID,
			"reference_type": blankToNil(lock.ReferenceType),
			"reference_id":   blankToNil(lock.ReferenceID),
			"correlation_id": blankToNil(lock.CorrelationID),
			"description":    blankToNil(lock.Description),
			"created_at":     lock.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":     lock.UpdatedAt.UTC().Format(time.RFC3339),
			"released_at":    timestampRFC3339(lock.ReleasedAt),
		})
	}
	return rows
}

func serializeCollateralLocks(locks []domain.CollateralLock) []map[string]any {
	rows := make([]map[string]any, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, map[string]any{
			"id":                    lock.ID,
			"user_id":               lock.UserID,
			"contract_id":           lock.ContractID,
			"currency":              lock.Currency,
			"amount_cents":          lock.AmountCents,
			"status":                lock.Status,
			"reference_id":          blankToNil(lock.ReferenceID),
			"description":           blankToNil(lock.Description),
			"reference_issuance_id": lock.ReferenceIssuanceID,
			"created_at":            lock.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":            lock.UpdatedAt.UTC().Format(time.RFC3339),
			"released_at":           timestampRFC3339(lock.ReleasedAt),
		})
	}
	return rows
}

func serializeOrderCashReservations(reservations []domain.OrderCashReservation) []map[string]any {
	rows := make([]map[string]any, 0, len(reservations))
	for _, reservation := range reservations {
		rows = append(rows, map[string]any{
			"id":             reservation.ID,
			"user_id":        reservation.UserID,
			"contract_id":    reservation.ContractID,
			"currency":       reservation.Currency,
			"amount_cents":   reservation.AmountCents,
			"status":         reservation.Status,
			"reference_type": blankToNil(reservation.ReferenceType),
			"reference_id":   blankToNil(reservation.ReferenceID),
			"correlation_id": blankToNil(reservation.CorrelationID),
			"description":    blankToNil(reservation.Description),
			"created_at":     reservation.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":     reservation.UpdatedAt.UTC().Format(time.RFC3339),
			"released_at":    timestampRFC3339(reservation.ReleasedAt),
		})
	}
	return rows
}

func serializeSettlementEntries(entries []domain.SettlementEntry) []map[string]any {
	rows := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, map[string]any{
			"id":           entry.ID,
			"contract_id":  entry.ContractID,
			"user_id":      entry.UserID,
			"entry_type":   entry.EntryType,
			"outcome":      entry.Outcome,
			"amount_cents": entry.AmountCents,
			"quantity":     entry.Quantity,
			"reference_id": entry.ReferenceID,
			"created_at":   entry.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return rows
}

func serializeExecutions(executions []domain.Execution) []map[string]any {
	rows := make([]map[string]any, 0, len(executions))
	for _, execution := range executions {
		rows = append(rows, serializeExecution(execution))
	}

	return rows
}

func serializeOracleObservations(observations []domain.OracleObservation) []map[string]any {
	rows := make([]map[string]any, 0, len(observations))
	for _, observation := range observations {
		rows = append(rows, serializeOracleObservation(observation))
	}
	return rows
}

func serializeOracleObservation(observation domain.OracleObservation) map[string]any {
	return map[string]any{
		"id":                       observation.ID,
		"contract_id":              observation.ContractID,
		"provider_name":            observation.ProviderName,
		"station_id":               observation.StationID,
		"observed_metric":          observation.ObservedMetric,
		"observation_window_start": observation.ObservationWindowStart.UTC().Format(time.RFC3339),
		"observation_window_end":   observation.ObservationWindowEnd.UTC().Format(time.RFC3339),
		"observed_value":           observation.ObservedValue,
		"normalized_value":         observation.NormalizedValue,
		"observed_at":              observation.ObservedAt.UTC().Format(time.RFC3339),
		"recorded_at":              observation.RecordedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContractResolution(resolution domain.ContractResolution) map[string]any {
	return map[string]any{
		"id":                       resolution.ID,
		"contract_id":              resolution.ContractID,
		"provider_name":            resolution.ProviderName,
		"station_id":               resolution.StationID,
		"observed_metric":          resolution.ObservedMetric,
		"observation_window_start": resolution.ObservationWindowStart.UTC().Format(time.RFC3339),
		"observation_window_end":   resolution.ObservationWindowEnd.UTC().Format(time.RFC3339),
		"rule_version":             resolution.RuleVersion,
		"resolved_value":           resolution.ResolvedValue,
		"outcome":                  resolution.Outcome,
		"resolved_at":              resolution.ResolvedAt.UTC().Format(time.RFC3339),
	}
}

func serializeExecution(execution domain.Execution) map[string]any {
	return map[string]any{
		"id":             execution.ID,
		"execution_id":   execution.ExecutionID,
		"contract_id":    execution.ContractID,
		"buy_order_id":   execution.BuyOrderID,
		"sell_order_id":  execution.SellOrderID,
		"buyer_user_id":  execution.BuyerUserID,
		"seller_user_id": execution.SellerUserID,
		"price":          execution.Price,
		"quantity":       execution.Quantity,
		"sequence":       execution.Sequence,
		"occurred_at":    execution.OccurredAt.UTC().Format(time.RFC3339),
	}
}

func serializeOrder(order domain.Order) map[string]any {
	return map[string]any{
		"id":          order.ID,
		"contract_id": order.ContractID,
		"user_id":     order.UserID,
		"token_type":  order.TokenType,
		"order_side":  order.OrderSide,
		"price":       order.Price,
		"quantity":    order.Quantity,
		"status":      order.Status,
	}
}

func serializePlaceOrderResult(result commands.PlaceOrderResult) map[string]any {
	payload := map[string]any{
		"status":      result.Status,
		"contract_id": result.ContractID,
		"executions":  serializeExecutions(result.Executions),
	}
	if result.Order != nil {
		payload["order"] = serializeOrder(*result.Order)
	}

	return payload
}

func serializePlaceOrderAccepted(result commands.PlaceOrderAccepted) map[string]any {
	return map[string]any{
		"command_id":  result.CommandID,
		"contract_id": result.ContractID,
		"partition":   result.Partition,
		"status":      result.Status,
		"enqueued_at": result.EnqueuedAt.UTC().Format(time.RFC3339),
	}
}

func serializeCreateContractAccepted(result commands.CreateContractAccepted) map[string]any {
	return map[string]any{
		"command_id":  result.CommandID,
		"partition":   result.Partition,
		"status":      result.Status,
		"enqueued_at": result.EnqueuedAt.UTC().Format(time.RFC3339),
	}
}

func serializeOrderCommand(command commands.OrderCommand) map[string]any {
	return map[string]any{
		"command_id":      command.CommandID,
		"contract_id":     command.ContractID,
		"user_id":         command.UserID,
		"token_type":      command.TokenType,
		"order_side":      command.OrderSide,
		"price":           command.Price,
		"quantity":        command.Quantity,
		"status":          command.Status,
		"error_message":   command.ErrorMessage,
		"result_status":   command.ResultStatus,
		"result_order_id": command.ResultOrderID,
		"enqueued_at":     command.EnqueuedAt.UTC().Format(time.RFC3339),
		"started_at":      timestampRFC3339(command.StartedAt),
		"completed_at":    timestampRFC3339(command.CompletedAt),
		"updated_at":      command.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeContractCommand(command commands.ContractCommand) map[string]any {
	return map[string]any{
		"command_id":                 command.CommandID,
		"creator_user_id":            command.CreatorUserID,
		"name":                       command.Name,
		"region":                     command.Region,
		"metric":                     command.Metric,
		"status":                     command.Status,
		"threshold":                  command.Threshold,
		"multiplier":                 command.Multiplier,
		"measurement_unit":           command.MeasurementUnit,
		"trading_period_start":       dateString(command.TradingPeriodStart),
		"trading_period_end":         dateString(command.TradingPeriodEnd),
		"measurement_period_start":   dateString(command.MeasurementPeriodStart),
		"measurement_period_end":     dateString(command.MeasurementPeriodEnd),
		"data_provider_name":         command.DataProviderName,
		"station_id":                 command.StationID,
		"data_provider_station_mode": command.DataProviderStationMode,
		"description":                command.Description,
		"command_status":             command.CommandStatus,
		"error_message":              command.ErrorMessage,
		"result_contract_id":         command.ResultContractID,
		"enqueued_at":                command.EnqueuedAt.UTC().Format(time.RFC3339),
		"started_at":                 timestampRFC3339(command.StartedAt),
		"completed_at":               timestampRFC3339(command.CompletedAt),
		"updated_at":                 command.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeMarketState(marketState domain.MarketState) map[string]any {
	return map[string]any{
		"contract_id": marketState.ContractID,
		"sequence":    marketState.Sequence,
		"as_of":       timestampString(marketState.AsOf),
		"order_book": map[string]any{
			"above": map[string]any{
				"bid": serializeLevels(marketState.OrderBook.Above.Bid),
				"ask": serializeLevels(marketState.OrderBook.Above.Ask),
			},
			"below": map[string]any{
				"bid": serializeLevels(marketState.OrderBook.Below.Bid),
				"ask": serializeLevels(marketState.OrderBook.Below.Ask),
			},
		},
		"summary": map[string]any{
			"best": map[string]any{
				"above": map[string]any{
					"bid": marketState.Summary.Best.Above.Bid,
					"ask": marketState.Summary.Best.Above.Ask,
				},
				"below": map[string]any{
					"bid": marketState.Summary.Best.Below.Bid,
					"ask": marketState.Summary.Best.Below.Ask,
				},
			},
			"mid": map[string]any{
				"above": marketState.Summary.Mid.Above,
				"below": marketState.Summary.Mid.Below,
			},
			"liquidity": map[string]any{
				"above": marketState.Summary.Liquidity.Above,
				"below": marketState.Summary.Liquidity.Below,
			},
			"above_below_bid_gap": marketState.Summary.AboveBelowBidGap,
		},
	}
}

func serializeChartPoints(points []domain.ChartPoint) []map[string]any {
	rows := make([]map[string]any, 0, len(points))
	for _, point := range points {
		rows = append(rows, serializeChartPoint(point))
	}

	return rows
}

func serializeChartPoint(point domain.ChartPoint) map[string]any {
	return map[string]any{
		"bucket_start": point.BucketStart.UTC().Format(time.RFC3339[:19]),
		"inserted_at":  point.InsertedAt.UTC().Format(time.RFC3339[:19]),
		"mid_above":    point.MidAbove,
		"mid_below":    point.MidBelow,
		"best_above":   point.BestAbove,
		"best_below":   point.BestBelow,
	}
}

func serializeLevels(levels []domain.PriceLevel) []map[string]any {
	rows := make([]map[string]any, 0, len(levels))
	for _, level := range levels {
		rows = append(rows, map[string]any{
			"price":    level.Price,
			"quantity": level.Quantity,
		})
	}

	return rows
}

func timestampString(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC().Format("2006-01-02T15:04:05")
}

func dateString(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC().Format("2006-01-02")
}

func timestampRFC3339(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC().Format(time.RFC3339)
}

func blankToNil(value string) any {
	if value == "" {
		return nil
	}

	return value
}
