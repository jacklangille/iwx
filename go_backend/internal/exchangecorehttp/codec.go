package exchangecorehttp

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/exchangecore"
)

func decodeCreateContractCommand(body io.Reader) (commands.CreateContract, error) {
	if body == nil {
		return commands.CreateContract{}, errors.New("request body is required")
	}

	var command commands.CreateContract
	if err := json.NewDecoder(body).Decode(&command); err != nil {
		return commands.CreateContract{}, err
	}

	return command, nil
}

type accountMoneyRequest struct {
	Currency      string `json:"currency"`
	AmountCents   int64  `json:"amount_cents"`
	ReferenceID   string `json:"reference_id"`
	CorrelationID string `json:"correlation_id"`
	Description   string `json:"description"`
}

type collateralLockRequest struct {
	ContractID    int64  `json:"contract_id"`
	Currency      string `json:"currency"`
	AmountCents   int64  `json:"amount_cents"`
	ReferenceID   string `json:"reference_id"`
	CorrelationID string `json:"correlation_id"`
	Description   string `json:"description"`
}

type cashReservationRequest struct {
	ContractID    int64  `json:"contract_id"`
	Currency      string `json:"currency"`
	AmountCents   int64  `json:"amount_cents"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
	CorrelationID string `json:"correlation_id"`
	Description   string `json:"description"`
}

type releaseRequest struct {
	CorrelationID string `json:"correlation_id"`
	Description   string `json:"description"`
}

type contractCollateralLockRequest struct {
	PairedQuantity int64  `json:"paired_quantity"`
	Currency       string `json:"currency"`
	CorrelationID  string `json:"correlation_id"`
	Description    string `json:"description"`
}

type issuanceBatchRequest struct {
	CollateralLockID int64 `json:"collateral_lock_id"`
	PairedQuantity   int64 `json:"paired_quantity"`
}

func decodeAccountMoneyRequest(body io.Reader) (accountMoneyRequest, error) {
	if body == nil {
		return accountMoneyRequest{}, errors.New("request body is required")
	}

	var request accountMoneyRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return accountMoneyRequest{}, err
	}

	return request, nil
}

func decodeCollateralLockRequest(body io.Reader) (collateralLockRequest, error) {
	if body == nil {
		return collateralLockRequest{}, errors.New("request body is required")
	}

	var request collateralLockRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return collateralLockRequest{}, err
	}

	return request, nil
}

func decodeCashReservationRequest(body io.Reader) (cashReservationRequest, error) {
	if body == nil {
		return cashReservationRequest{}, errors.New("request body is required")
	}

	var request cashReservationRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return cashReservationRequest{}, err
	}

	return request, nil
}

func decodeReleaseRequest(body io.Reader) (releaseRequest, error) {
	if body == nil {
		return releaseRequest{}, nil
	}

	var request releaseRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil && !errors.Is(err, io.EOF) {
		return releaseRequest{}, err
	}

	return request, nil
}

func decodeContractCollateralLockRequest(body io.Reader) (contractCollateralLockRequest, error) {
	if body == nil {
		return contractCollateralLockRequest{}, errors.New("request body is required")
	}

	var request contractCollateralLockRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return contractCollateralLockRequest{}, err
	}

	return request, nil
}

func decodeIssuanceBatchRequest(body io.Reader) (issuanceBatchRequest, error) {
	if body == nil {
		return issuanceBatchRequest{}, errors.New("request body is required")
	}

	var request issuanceBatchRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return issuanceBatchRequest{}, err
	}

	return request, nil
}

func serializeContractCommand(command *commands.ContractCommand) map[string]any {
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

func serializePlaceOrderAccepted(result commands.PlaceOrderAccepted) map[string]any {
	return map[string]any{
		"command_id":  result.CommandID,
		"contract_id": result.ContractID,
		"partition":   result.Partition,
		"status":      result.Status,
		"enqueued_at": result.EnqueuedAt.UTC().Format(time.RFC3339),
	}
}

func serializePlaceOrderCommand(command commands.OrderCommand) map[string]any {
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

func serializeContractDetails(details *exchangecore.ContractDetails) map[string]any {
	if details == nil {
		return nil
	}

	return map[string]any{
		"contract": serializeContractLifecycleContract(details.Contract),
		"rule":     serializeContractRule(details.Rule),
	}
}

func serializeContractLifecycleContract(contract *domain.Contract) map[string]any {
	if contract == nil {
		return nil
	}

	return map[string]any{
		"id":                         contract.ID,
		"creator_user_id":            contract.CreatorUserID,
		"name":                       contract.Name,
		"region":                     contract.Region,
		"metric":                     contract.Metric,
		"status":                     contract.Status,
		"threshold":                  contract.Threshold,
		"multiplier":                 contract.Multiplier,
		"measurement_unit":           blankToNil(contract.MeasurementUnit),
		"trading_period_start":       dateString(contract.TradingPeriodStart),
		"trading_period_end":         dateString(contract.TradingPeriodEnd),
		"measurement_period_start":   dateString(contract.MeasurementPeriodStart),
		"measurement_period_end":     dateString(contract.MeasurementPeriodEnd),
		"data_provider_name":         blankToNil(contract.DataProviderName),
		"station_id":                 blankToNil(contract.StationID),
		"data_provider_station_mode": blankToNil(contract.DataProviderStationMode),
		"description":                blankToNil(contract.Description),
	}
}

func serializeContractRule(rule *domain.ContractRule) map[string]any {
	if rule == nil {
		return nil
	}

	return map[string]any{
		"id":                        rule.ID,
		"contract_id":               rule.ContractID,
		"rule_version":              rule.RuleVersion,
		"metric":                    rule.Metric,
		"threshold":                 rule.Threshold,
		"measurement_unit":          blankToNil(rule.MeasurementUnit),
		"resolution_inclusive_side": rule.ResolutionInclusiveSide,
		"created_at":                rule.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeCollateralRequirement(requirement *exchangecore.CollateralRequirement) map[string]any {
	if requirement == nil {
		return nil
	}

	return map[string]any{
		"contract_id":           requirement.ContractID,
		"paired_quantity":       requirement.PairedQuantity,
		"per_pair_cents":        requirement.PerPairCents,
		"required_amount_cents": requirement.RequiredAmountCents,
		"currency":              requirement.Currency,
	}
}

func serializeCashAccount(account *domain.CashAccount) map[string]any {
	if account == nil {
		return nil
	}

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

func serializeLedgerEntries(entries []domain.LedgerEntry) []map[string]any {
	rows := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, serializeLedgerEntry(&entry))
	}

	return rows
}

func serializeLedgerEntry(entry *domain.LedgerEntry) map[string]any {
	if entry == nil {
		return nil
	}

	return map[string]any{
		"id":             entry.ID,
		"account_id":     entry.AccountID,
		"user_id":        entry.UserID,
		"entry_type":     entry.EntryType,
		"amount_cents":   entry.AmountCents,
		"reference_type": entry.ReferenceType,
		"reference_id":   entry.ReferenceID,
		"correlation_id": entry.CorrelationID,
		"description":    entry.Description,
		"occurred_at":    entry.OccurredAt.UTC().Format(time.RFC3339),
	}
}

func serializeCollateralLocks(locks []domain.CollateralLock) []map[string]any {
	rows := make([]map[string]any, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, serializeCollateralLock(&lock))
	}

	return rows
}

func serializeCollateralLock(lock *domain.CollateralLock) map[string]any {
	if lock == nil {
		return nil
	}

	return map[string]any{
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
	}
}

func serializeOrderCashReservations(reservations []domain.OrderCashReservation) []map[string]any {
	rows := make([]map[string]any, 0, len(reservations))
	for _, reservation := range reservations {
		rows = append(rows, serializeOrderCashReservation(&reservation))
	}

	return rows
}

func serializeOrderCashReservation(reservation *domain.OrderCashReservation) map[string]any {
	if reservation == nil {
		return nil
	}

	return map[string]any{
		"id":             reservation.ID,
		"user_id":        reservation.UserID,
		"contract_id":    reservation.ContractID,
		"currency":       reservation.Currency,
		"amount_cents":   reservation.AmountCents,
		"status":         reservation.Status,
		"reference_type": reservation.ReferenceType,
		"reference_id":   reservation.ReferenceID,
		"correlation_id": blankToNil(reservation.CorrelationID),
		"description":    blankToNil(reservation.Description),
		"created_at":     reservation.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":     reservation.UpdatedAt.UTC().Format(time.RFC3339),
		"released_at":    timestampRFC3339(reservation.ReleasedAt),
	}
}

func serializeIssuanceBatches(batches []domain.IssuanceBatch) []map[string]any {
	rows := make([]map[string]any, 0, len(batches))
	for _, batch := range batches {
		rows = append(rows, serializeIssuanceBatch(&batch))
	}

	return rows
}

func serializeIssuanceBatch(batch *domain.IssuanceBatch) map[string]any {
	if batch == nil {
		return nil
	}

	return map[string]any{
		"id":                 batch.ID,
		"contract_id":        batch.ContractID,
		"creator_user_id":    batch.CreatorUserID,
		"collateral_lock_id": batch.CollateralLockID,
		"above_quantity":     batch.AboveQuantity,
		"below_quantity":     batch.BelowQuantity,
		"status":             batch.Status,
		"created_at":         batch.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":         batch.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializePositions(positions []*domain.Position) []map[string]any {
	rows := make([]map[string]any, 0, len(positions))
	for _, position := range positions {
		rows = append(rows, serializePosition(position))
	}

	return rows
}

func serializePosition(position *domain.Position) map[string]any {
	if position == nil {
		return nil
	}

	return map[string]any{
		"id":                 position.ID,
		"user_id":            position.UserID,
		"contract_id":        position.ContractID,
		"side":               position.Side,
		"available_quantity": position.AvailableQuantity,
		"locked_quantity":    position.LockedQuantity,
		"total_quantity":     position.TotalQuantity,
		"updated_at":         position.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializePositionsFromValues(positions []domain.Position) []map[string]any {
	rows := make([]map[string]any, 0, len(positions))
	for _, position := range positions {
		positionCopy := position
		rows = append(rows, serializePosition(&positionCopy))
	}

	return rows
}

func serializePositionLocks(locks []domain.PositionLock) []map[string]any {
	rows := make([]map[string]any, 0, len(locks))
	for _, lock := range locks {
		lockCopy := lock
		rows = append(rows, serializePositionLock(&lockCopy))
	}

	return rows
}

func serializePositionLock(lock *domain.PositionLock) map[string]any {
	if lock == nil {
		return nil
	}

	return map[string]any{
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
	}
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
