package postgres

import (
	"database/sql"
	"strings"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

func scanContract(rows scanner) (domain.Contract, error) {
	var contract domain.Contract
	var creatorUserID sql.NullInt64
	var threshold sql.NullInt64
	var multiplier sql.NullInt64
	var tradingStart sql.NullTime
	var tradingEnd sql.NullTime
	var measurementStart sql.NullTime
	var measurementEnd sql.NullTime
	var dataProviderName sql.NullString
	var stationID sql.NullString
	var dataProviderStationMode sql.NullString
	var description sql.NullString
	var measurementUnit sql.NullString
	var updatedAt sql.NullTime

	err := rows.Scan(
		&contract.ID,
		&creatorUserID,
		&contract.Name,
		&contract.Region,
		&contract.Metric,
		&contract.Status,
		&threshold,
		&multiplier,
		&measurementUnit,
		&tradingStart,
		&tradingEnd,
		&measurementStart,
		&measurementEnd,
		&dataProviderName,
		&stationID,
		&dataProviderStationMode,
		&description,
		&updatedAt,
	)
	if err != nil {
		return domain.Contract{}, err
	}

	contract.Threshold = nullableInt64(threshold)
	contract.CreatorUserID = nullableInt64(creatorUserID)
	contract.Multiplier = nullableInt64(multiplier)
	contract.MeasurementUnit = measurementUnit.String
	contract.TradingPeriodStart = nullableTime(tradingStart)
	contract.TradingPeriodEnd = nullableTime(tradingEnd)
	contract.MeasurementPeriodStart = nullableTime(measurementStart)
	contract.MeasurementPeriodEnd = nullableTime(measurementEnd)
	contract.DataProviderName = dataProviderName.String
	contract.StationID = stationID.String
	contract.DataProviderStationMode = dataProviderStationMode.String
	contract.Description = description.String
	contract.UpdatedAt = zeroableTime(updatedAt)

	return contract, nil
}

func scanCashAccount(rows scanner) (domain.CashAccount, error) {
	var account domain.CashAccount

	err := rows.Scan(
		&account.ID,
		&account.UserID,
		&account.Currency,
		&account.AvailableCents,
		&account.LockedCents,
		&account.TotalCents,
		&account.UpdatedAt,
	)
	if err != nil {
		return domain.CashAccount{}, err
	}

	account.UpdatedAt = account.UpdatedAt.UTC()
	return account, nil
}

func scanLedgerEntry(rows scanner) (domain.LedgerEntry, error) {
	var entry domain.LedgerEntry

	err := rows.Scan(
		&entry.ID,
		&entry.AccountID,
		&entry.UserID,
		&entry.EntryType,
		&entry.AmountCents,
		&entry.ReferenceType,
		&entry.ReferenceID,
		&entry.CorrelationID,
		&entry.Description,
		&entry.OccurredAt,
	)
	if err != nil {
		return domain.LedgerEntry{}, err
	}

	entry.OccurredAt = entry.OccurredAt.UTC()
	return entry, nil
}

func scanCollateralLock(rows scanner) (domain.CollateralLock, error) {
	var lock domain.CollateralLock
	var referenceIssuanceID sql.NullInt64
	var releasedAt sql.NullTime

	err := rows.Scan(
		&lock.ID,
		&lock.UserID,
		&lock.ContractID,
		&lock.Currency,
		&lock.AmountCents,
		&lock.Status,
		&lock.ReferenceID,
		&lock.Description,
		&referenceIssuanceID,
		&lock.CreatedAt,
		&lock.UpdatedAt,
		&releasedAt,
	)
	if err != nil {
		return domain.CollateralLock{}, err
	}

	lock.ReferenceIssuanceID = nullableInt64(referenceIssuanceID)
	lock.CreatedAt = lock.CreatedAt.UTC()
	lock.UpdatedAt = lock.UpdatedAt.UTC()
	lock.ReleasedAt = nullableTime(releasedAt)
	return lock, nil
}

func scanOrderCashReservation(rows scanner) (domain.OrderCashReservation, error) {
	var reservation domain.OrderCashReservation
	var releasedAt sql.NullTime

	err := rows.Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ContractID,
		&reservation.Currency,
		&reservation.AmountCents,
		&reservation.Status,
		&reservation.ReferenceType,
		&reservation.ReferenceID,
		&reservation.CorrelationID,
		&reservation.Description,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&releasedAt,
	)
	if err != nil {
		return domain.OrderCashReservation{}, err
	}

	reservation.CreatedAt = reservation.CreatedAt.UTC()
	reservation.UpdatedAt = reservation.UpdatedAt.UTC()
	reservation.ReleasedAt = nullableTime(releasedAt)
	return reservation, nil
}

func scanIssuanceBatch(rows scanner) (domain.IssuanceBatch, error) {
	var batch domain.IssuanceBatch

	err := rows.Scan(
		&batch.ID,
		&batch.ContractID,
		&batch.CreatorUserID,
		&batch.CollateralLockID,
		&batch.AboveQuantity,
		&batch.BelowQuantity,
		&batch.Status,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return domain.IssuanceBatch{}, err
	}

	batch.CreatedAt = batch.CreatedAt.UTC()
	batch.UpdatedAt = batch.UpdatedAt.UTC()
	return batch, nil
}

func scanPosition(rows scanner) (domain.Position, error) {
	var position domain.Position

	err := rows.Scan(
		&position.ID,
		&position.UserID,
		&position.ContractID,
		&position.Side,
		&position.AvailableQuantity,
		&position.LockedQuantity,
		&position.TotalQuantity,
		&position.UpdatedAt,
	)
	if err != nil {
		return domain.Position{}, err
	}

	position.UpdatedAt = position.UpdatedAt.UTC()
	return position, nil
}

func scanPositionLock(rows scanner) (domain.PositionLock, error) {
	var lock domain.PositionLock
	var orderID sql.NullInt64
	var releasedAt sql.NullTime

	err := rows.Scan(
		&lock.ID,
		&lock.UserID,
		&lock.ContractID,
		&lock.Side,
		&lock.Quantity,
		&lock.Status,
		&orderID,
		&lock.ReferenceType,
		&lock.ReferenceID,
		&lock.CorrelationID,
		&lock.Description,
		&lock.CreatedAt,
		&lock.UpdatedAt,
		&releasedAt,
	)
	if err != nil {
		return domain.PositionLock{}, err
	}

	lock.OrderID = nullableInt64(orderID)
	lock.CreatedAt = lock.CreatedAt.UTC()
	lock.UpdatedAt = lock.UpdatedAt.UTC()
	lock.ReleasedAt = nullableTime(releasedAt)
	return lock, nil
}

func scanOrder(rows scanner) (domain.Order, error) {
	var order domain.Order
	var cashReservationID sql.NullInt64
	var positionLockID sql.NullInt64

	err := rows.Scan(
		&order.ID,
		&order.ContractID,
		&order.UserID,
		&order.TokenType,
		&order.OrderSide,
		&order.Price,
		&order.Quantity,
		&order.Status,
		&cashReservationID,
		&positionLockID,
		&order.ReservationCorrelationID,
		&order.InsertedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, err
	}

	order.CashReservationID = nullableInt64(cashReservationID)
	order.PositionLockID = nullableInt64(positionLockID)
	order.InsertedAt = order.InsertedAt.UTC()
	order.UpdatedAt = order.UpdatedAt.UTC()

	return order, nil
}

func scanSnapshot(rows scanner) (domain.MarketSnapshot, error) {
	var snapshot domain.MarketSnapshot
	var bestAbove sql.NullString
	var bestBelow sql.NullString
	var midAbove sql.NullString
	var midBelow sql.NullString

	err := rows.Scan(
		&snapshot.ID,
		&snapshot.ContractID,
		&bestAbove,
		&bestBelow,
		&midAbove,
		&midBelow,
		&snapshot.InsertedAt,
	)
	if err != nil {
		return domain.MarketSnapshot{}, err
	}

	snapshot.BestAbove = nullableString(bestAbove)
	snapshot.BestBelow = nullableString(bestBelow)
	snapshot.MidAbove = nullableString(midAbove)
	snapshot.MidBelow = nullableString(midBelow)
	snapshot.InsertedAt = snapshot.InsertedAt.UTC()

	return snapshot, nil
}

func scanExecution(rows scanner) (domain.Execution, error) {
	var execution domain.Execution
	var buyerCashReservationID sql.NullInt64
	var sellerPositionLockID sql.NullInt64

	err := rows.Scan(
		&execution.ID,
		&execution.ExecutionID,
		&execution.CommandID,
		&execution.ContractID,
		&execution.TokenType,
		&execution.BuyOrderID,
		&execution.SellOrderID,
		&execution.BuyerUserID,
		&execution.SellerUserID,
		&buyerCashReservationID,
		&sellerPositionLockID,
		&execution.Price,
		&execution.Quantity,
		&execution.Sequence,
		&execution.OccurredAt,
	)
	if err != nil {
		return domain.Execution{}, err
	}

	execution.BuyerCashReservationID = nullableInt64(buyerCashReservationID)
	execution.SellerPositionLockID = nullableInt64(sellerPositionLockID)
	execution.OccurredAt = execution.OccurredAt.UTC()
	return execution, nil
}

func scanOracleObservation(rows scanner) (domain.OracleObservation, error) {
	var observation domain.OracleObservation

	err := rows.Scan(
		&observation.ID,
		&observation.ContractID,
		&observation.ProviderName,
		&observation.StationID,
		&observation.ObservedMetric,
		&observation.ObservationWindowStart,
		&observation.ObservationWindowEnd,
		&observation.ObservedValue,
		&observation.NormalizedValue,
		&observation.ObservedAt,
		&observation.RecordedAt,
	)
	if err != nil {
		return domain.OracleObservation{}, err
	}

	observation.ObservationWindowStart = observation.ObservationWindowStart.UTC()
	observation.ObservationWindowEnd = observation.ObservationWindowEnd.UTC()
	observation.ObservedAt = observation.ObservedAt.UTC()
	observation.RecordedAt = observation.RecordedAt.UTC()
	return observation, nil
}

func scanWeatherStation(rows scanner) (domain.WeatherStation, error) {
	var station domain.WeatherStation
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var supportedMetrics string

	err := rows.Scan(
		&station.ID,
		&station.ProviderName,
		&station.StationID,
		&station.DisplayName,
		&station.Region,
		&latitude,
		&longitude,
		&supportedMetrics,
		&station.Active,
		&station.UpdatedAt,
	)
	if err != nil {
		return domain.WeatherStation{}, err
	}

	if latitude.Valid {
		value := latitude.Float64
		station.Latitude = &value
	}
	if longitude.Valid {
		value := longitude.Float64
		station.Longitude = &value
	}
	if strings.TrimSpace(supportedMetrics) != "" {
		for _, metric := range strings.Split(supportedMetrics, ",") {
			trimmed := strings.TrimSpace(metric)
			if trimmed != "" {
				station.SupportedMetrics = append(station.SupportedMetrics, trimmed)
			}
		}
	}
	station.UpdatedAt = station.UpdatedAt.UTC()
	return station, nil
}

func scanContractResolution(rows scanner) (domain.ContractResolution, error) {
	var resolution domain.ContractResolution
	var publishedAt sql.NullTime

	err := rows.Scan(
		&resolution.ID,
		&resolution.EventID,
		&resolution.ContractID,
		&resolution.ProviderName,
		&resolution.StationID,
		&resolution.ObservedMetric,
		&resolution.ObservationWindowStart,
		&resolution.ObservationWindowEnd,
		&resolution.RuleVersion,
		&resolution.ResolvedValue,
		&resolution.Outcome,
		&resolution.ResolvedAt,
		&publishedAt,
	)
	if err != nil {
		return domain.ContractResolution{}, err
	}

	resolution.ObservationWindowStart = resolution.ObservationWindowStart.UTC()
	resolution.ObservationWindowEnd = resolution.ObservationWindowEnd.UTC()
	resolution.ResolvedAt = resolution.ResolvedAt.UTC()
	resolution.PublishedAt = nullableTime(publishedAt)
	return resolution, nil
}

func scanSettlementEntry(rows scanner) (domain.SettlementEntry, error) {
	var entry domain.SettlementEntry

	err := rows.Scan(
		&entry.ID,
		&entry.ContractID,
		&entry.UserID,
		&entry.EntryType,
		&entry.Outcome,
		&entry.AmountCents,
		&entry.Quantity,
		&entry.ReferenceID,
		&entry.CreatedAt,
	)
	if err != nil {
		return domain.SettlementEntry{}, err
	}

	entry.CreatedAt = entry.CreatedAt.UTC()
	return entry, nil
}

func scanOrderCommand(rows scanner) (commands.OrderCommand, error) {
	var command commands.OrderCommand
	var errorMessage sql.NullString
	var resultStatus sql.NullString
	var resultOrderID sql.NullInt64
	var startedAt sql.NullTime
	var completedAt sql.NullTime

	err := rows.Scan(
		&command.CommandID,
		&command.ContractID,
		&command.UserID,
		&command.TokenType,
		&command.OrderSide,
		&command.Price,
		&command.Quantity,
		&command.Status,
		&errorMessage,
		&resultStatus,
		&resultOrderID,
		&command.EnqueuedAt,
		&startedAt,
		&completedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		return commands.OrderCommand{}, err
	}

	command.ErrorMessage = nullableString(errorMessage)
	command.ResultStatus = nullableString(resultStatus)
	command.ResultOrderID = nullableInt64(resultOrderID)
	command.StartedAt = nullableTime(startedAt)
	command.CompletedAt = nullableTime(completedAt)
	command.EnqueuedAt = command.EnqueuedAt.UTC()
	command.UpdatedAt = command.UpdatedAt.UTC()

	return command, nil
}

func scanContractCommand(rows scanner) (commands.ContractCommand, error) {
	var command commands.ContractCommand
	var threshold sql.NullInt64
	var multiplier sql.NullInt64
	var measurementUnit sql.NullString
	var tradingStart sql.NullTime
	var tradingEnd sql.NullTime
	var measurementStart sql.NullTime
	var measurementEnd sql.NullTime
	var dataProviderName sql.NullString
	var stationID sql.NullString
	var dataProviderStationMode sql.NullString
	var description sql.NullString
	var errorMessage sql.NullString
	var resultContractID sql.NullInt64
	var startedAt sql.NullTime
	var completedAt sql.NullTime

	err := rows.Scan(
		&command.CommandID,
		&command.CreatorUserID,
		&command.Name,
		&command.Region,
		&command.Metric,
		&command.Status,
		&threshold,
		&multiplier,
		&measurementUnit,
		&tradingStart,
		&tradingEnd,
		&measurementStart,
		&measurementEnd,
		&dataProviderName,
		&stationID,
		&dataProviderStationMode,
		&description,
		&command.CommandStatus,
		&errorMessage,
		&resultContractID,
		&command.EnqueuedAt,
		&startedAt,
		&completedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		return commands.ContractCommand{}, err
	}

	command.Threshold = nullableInt64(threshold)
	command.Multiplier = nullableInt64(multiplier)
	command.MeasurementUnit = nullableString(measurementUnit)
	command.TradingPeriodStart = nullableTime(tradingStart)
	command.TradingPeriodEnd = nullableTime(tradingEnd)
	command.MeasurementPeriodStart = nullableTime(measurementStart)
	command.MeasurementPeriodEnd = nullableTime(measurementEnd)
	command.DataProviderName = nullableString(dataProviderName)
	command.StationID = nullableString(stationID)
	command.DataProviderStationMode = nullableString(dataProviderStationMode)
	command.Description = nullableString(description)
	command.ErrorMessage = nullableString(errorMessage)
	command.ResultContractID = nullableInt64(resultContractID)
	command.StartedAt = nullableTime(startedAt)
	command.CompletedAt = nullableTime(completedAt)
	command.EnqueuedAt = command.EnqueuedAt.UTC()
	command.UpdatedAt = command.UpdatedAt.UTC()

	return command, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	copy := value.String
	return &copy
}

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	copy := value.Int64
	return &copy
}

func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	copy := value.Time.UTC()
	return &copy
}

func zeroableTime(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func nullIfZeroTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}
