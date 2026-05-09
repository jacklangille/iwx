package exchangecorehttp

import (
	"io"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/httpjson"
)

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

type internalSettlementRequest struct {
	EventID       string `json:"event_id"`
	Outcome       string `json:"outcome"`
	ResolvedAt    string `json:"resolved_at"`
	CorrelationID string `json:"correlation_id"`
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

type placeOrderAcceptedResponse struct {
	CommandID  string `json:"command_id"`
	ContractID int64  `json:"contract_id"`
	Partition  int    `json:"partition"`
	Status     string `json:"status"`
	EnqueuedAt string `json:"enqueued_at"`
}

type createContractAcceptedResponse struct {
	CommandID  string `json:"command_id"`
	Partition  int    `json:"partition"`
	Status     string `json:"status"`
	EnqueuedAt string `json:"enqueued_at"`
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

type contractDetailsResponse struct {
	Contract *contractResponse     `json:"contract"`
	Rule     *contractRuleResponse `json:"rule"`
}

type contractResponse struct {
	ID                      int64   `json:"id"`
	CreatorUserID           *int64  `json:"creator_user_id"`
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
}

type contractRuleResponse struct {
	ID                      int64   `json:"id"`
	ContractID              int64   `json:"contract_id"`
	RuleVersion             string  `json:"rule_version"`
	Metric                  string  `json:"metric"`
	Threshold               *int64  `json:"threshold"`
	MeasurementUnit         *string `json:"measurement_unit"`
	ResolutionInclusiveSide string  `json:"resolution_inclusive_side"`
	CreatedAt               string  `json:"created_at"`
}

type collateralRequirementResponse struct {
	ContractID          int64  `json:"contract_id"`
	PairedQuantity      int64  `json:"paired_quantity"`
	PerPairCents        int64  `json:"per_pair_cents"`
	RequiredAmountCents int64  `json:"required_amount_cents"`
	Currency            string `json:"currency"`
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

type ledgerEntryResponse struct {
	ID            int64  `json:"id"`
	AccountID     int64  `json:"account_id"`
	UserID        int64  `json:"user_id"`
	EntryType     string `json:"entry_type"`
	AmountCents   int64  `json:"amount_cents"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
	CorrelationID string `json:"correlation_id"`
	Description   string `json:"description"`
	OccurredAt    string `json:"occurred_at"`
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
	ReferenceType string  `json:"reference_type"`
	ReferenceID   string  `json:"reference_id"`
	CorrelationID *string `json:"correlation_id"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ReleasedAt    *string `json:"released_at"`
}

type issuanceBatchResponse struct {
	ID               int64  `json:"id"`
	ContractID       int64  `json:"contract_id"`
	CreatorUserID    int64  `json:"creator_user_id"`
	CollateralLockID int64  `json:"collateral_lock_id"`
	AboveQuantity    int64  `json:"above_quantity"`
	BelowQuantity    int64  `json:"below_quantity"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
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

type accountMutationResponse struct {
	Account     *cashAccountResponse `json:"account"`
	LedgerEntry *ledgerEntryResponse `json:"ledger_entry"`
}

type collateralLockMutationResponse struct {
	CollateralLock *collateralLockResponse `json:"collateral_lock"`
	Account        *cashAccountResponse    `json:"account"`
	LedgerEntry    *ledgerEntryResponse    `json:"ledger_entry"`
}

type cashReservationMutationResponse struct {
	CashReservation *orderCashReservationResponse `json:"cash_reservation"`
	Account         *cashAccountResponse          `json:"account"`
	LedgerEntry     *ledgerEntryResponse          `json:"ledger_entry"`
}

type contractCollateralLockMutationResponse struct {
	Requirement    *collateralRequirementResponse `json:"requirement"`
	CollateralLock *collateralLockResponse        `json:"collateral_lock"`
	Account        *cashAccountResponse           `json:"account"`
	LedgerEntry    *ledgerEntryResponse           `json:"ledger_entry"`
}

type issuanceBatchMutationResponse struct {
	IssuanceBatch  *issuanceBatchResponse    `json:"issuance_batch"`
	CollateralLock *collateralLockResponse   `json:"collateral_lock"`
	Positions      []positionResponse        `json:"positions"`
}

type internalContractResponse struct {
	Contract *contractResponse     `json:"contract"`
	Rule     *contractRuleResponse `json:"rule"`
}

type internalSettlementResponse struct {
	Contract      *contractResponse `json:"contract"`
	AffectedUsers []int64           `json:"affected_users"`
	SettledAt     string            `json:"settled_at"`
}

func decodeCreateContractCommand(body io.Reader) (commands.CreateContract, error) {
	var command commands.CreateContract
	if err := httpjson.DecodeStrict(body, &command); err != nil {
		return commands.CreateContract{}, err
	}
	return command, nil
}

func decodeAccountMoneyRequest(body io.Reader) (accountMoneyRequest, error) {
	var request accountMoneyRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return accountMoneyRequest{}, err
	}
	return request, nil
}

func decodeCollateralLockRequest(body io.Reader) (collateralLockRequest, error) {
	var request collateralLockRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return collateralLockRequest{}, err
	}
	return request, nil
}

func decodeCashReservationRequest(body io.Reader) (cashReservationRequest, error) {
	var request cashReservationRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return cashReservationRequest{}, err
	}
	return request, nil
}

func decodeReleaseRequest(body io.Reader) (releaseRequest, error) {
	var request releaseRequest
	if err := httpjson.DecodeStrictAllowEOF(body, &request); err != nil {
		return releaseRequest{}, err
	}
	return request, nil
}

func decodeContractCollateralLockRequest(body io.Reader) (contractCollateralLockRequest, error) {
	var request contractCollateralLockRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return contractCollateralLockRequest{}, err
	}
	return request, nil
}

func decodeIssuanceBatchRequest(body io.Reader) (issuanceBatchRequest, error) {
	var request issuanceBatchRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return issuanceBatchRequest{}, err
	}
	return request, nil
}

func decodeInternalSettlementRequest(body io.Reader) (internalSettlementRequest, error) {
	var request internalSettlementRequest
	if err := httpjson.DecodeStrict(body, &request); err != nil {
		return internalSettlementRequest{}, err
	}
	return request, nil
}

func serializeContractCommand(command *commands.ContractCommand) *contractCommandResponse {
	if command == nil {
		return nil
	}
	return &contractCommandResponse{
		CommandID:               command.CommandID,
		CreatorUserID:           command.CreatorUserID,
		Name:                    command.Name,
		Region:                  command.Region,
		Metric:                  command.Metric,
		Status:                  command.Status,
		Threshold:               command.Threshold,
		Multiplier:              command.Multiplier,
		MeasurementUnit:         command.MeasurementUnit,
		TradingPeriodStart:      dateString(command.TradingPeriodStart),
		TradingPeriodEnd:        dateString(command.TradingPeriodEnd),
		MeasurementPeriodStart:  dateString(command.MeasurementPeriodStart),
		MeasurementPeriodEnd:    dateString(command.MeasurementPeriodEnd),
		DataProviderName:        command.DataProviderName,
		StationID:               command.StationID,
		DataProviderStationMode: command.DataProviderStationMode,
		Description:             command.Description,
		CommandStatus:           command.CommandStatus,
		ErrorMessage:            command.ErrorMessage,
		ResultContractID:        command.ResultContractID,
		EnqueuedAt:              command.EnqueuedAt.UTC().Format(time.RFC3339),
		StartedAt:               timestampRFC3339(command.StartedAt),
		CompletedAt:             timestampRFC3339(command.CompletedAt),
		UpdatedAt:               command.UpdatedAt.UTC().Format(time.RFC3339),
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

func serializePlaceOrderCommand(command commands.OrderCommand) orderCommandResponse {
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

func serializeContractDetails(details *exchangecore.ContractDetails) *contractDetailsResponse {
	if details == nil {
		return nil
	}
	return &contractDetailsResponse{
		Contract: serializeContractLifecycleContract(details.Contract),
		Rule:     serializeContractRule(details.Rule),
	}
}

func serializeContractLifecycleContract(contract *domain.Contract) *contractResponse {
	if contract == nil {
		return nil
	}
	return &contractResponse{
		ID:                      contract.ID,
		CreatorUserID:           contract.CreatorUserID,
		Name:                    contract.Name,
		Region:                  contract.Region,
		Metric:                  contract.Metric,
		Status:                  contract.Status,
		Threshold:               contract.Threshold,
		Multiplier:              contract.Multiplier,
		MeasurementUnit:         stringOrNil(contract.MeasurementUnit),
		TradingPeriodStart:      dateString(contract.TradingPeriodStart),
		TradingPeriodEnd:        dateString(contract.TradingPeriodEnd),
		MeasurementPeriodStart:  dateString(contract.MeasurementPeriodStart),
		MeasurementPeriodEnd:    dateString(contract.MeasurementPeriodEnd),
		DataProviderName:        stringOrNil(contract.DataProviderName),
		StationID:               stringOrNil(contract.StationID),
		DataProviderStationMode: stringOrNil(contract.DataProviderStationMode),
		Description:             stringOrNil(contract.Description),
	}
}

func serializeContractRule(rule *domain.ContractRule) *contractRuleResponse {
	if rule == nil {
		return nil
	}
	return &contractRuleResponse{
		ID:                      rule.ID,
		ContractID:              rule.ContractID,
		RuleVersion:             rule.RuleVersion,
		Metric:                  rule.Metric,
		Threshold:               rule.Threshold,
		MeasurementUnit:         stringOrNil(rule.MeasurementUnit),
		ResolutionInclusiveSide: string(rule.ResolutionInclusiveSide),
		CreatedAt:               rule.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeCollateralRequirement(requirement *exchangecore.CollateralRequirement) *collateralRequirementResponse {
	if requirement == nil {
		return nil
	}
	return &collateralRequirementResponse{
		ContractID:          requirement.ContractID,
		PairedQuantity:      requirement.PairedQuantity,
		PerPairCents:        requirement.PerPairCents,
		RequiredAmountCents: requirement.RequiredAmountCents,
		Currency:            requirement.Currency,
	}
}

func serializeCashAccount(account *domain.CashAccount) *cashAccountResponse {
	if account == nil {
		return nil
	}
	return &cashAccountResponse{
		ID:             account.ID,
		UserID:         account.UserID,
		Currency:       account.Currency,
		AvailableCents: account.AvailableCents,
		LockedCents:    account.LockedCents,
		TotalCents:     account.TotalCents,
		UpdatedAt:      account.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeLedgerEntries(entries []domain.LedgerEntry) []ledgerEntryResponse {
	rows := make([]ledgerEntryResponse, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, serializeLedgerEntry(&entry))
	}
	return rows
}

func serializeLedgerEntry(entry *domain.LedgerEntry) ledgerEntryResponse {
	if entry == nil {
		return ledgerEntryResponse{}
	}
	return ledgerEntryResponse{
		ID:            entry.ID,
		AccountID:     entry.AccountID,
		UserID:        entry.UserID,
		EntryType:     string(entry.EntryType),
		AmountCents:   entry.AmountCents,
		ReferenceType: entry.ReferenceType,
		ReferenceID:   entry.ReferenceID,
		CorrelationID: entry.CorrelationID,
		Description:   entry.Description,
		OccurredAt:    entry.OccurredAt.UTC().Format(time.RFC3339),
	}
}

func serializeCollateralLocks(locks []domain.CollateralLock) []collateralLockResponse {
	rows := make([]collateralLockResponse, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, *serializeCollateralLock(&lock))
	}
	return rows
}

func serializeCollateralLock(lock *domain.CollateralLock) *collateralLockResponse {
	if lock == nil {
		return nil
	}
	return &collateralLockResponse{
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
	}
}

func serializeOrderCashReservations(reservations []domain.OrderCashReservation) []orderCashReservationResponse {
	rows := make([]orderCashReservationResponse, 0, len(reservations))
	for _, reservation := range reservations {
		rows = append(rows, *serializeOrderCashReservation(&reservation))
	}
	return rows
}

func serializeOrderCashReservation(reservation *domain.OrderCashReservation) *orderCashReservationResponse {
	if reservation == nil {
		return nil
	}
	return &orderCashReservationResponse{
		ID:            reservation.ID,
		UserID:        reservation.UserID,
		ContractID:    reservation.ContractID,
		Currency:      reservation.Currency,
		AmountCents:   reservation.AmountCents,
		Status:        string(reservation.Status),
		ReferenceType: reservation.ReferenceType,
		ReferenceID:   reservation.ReferenceID,
		CorrelationID: stringOrNil(reservation.CorrelationID),
		Description:   stringOrNil(reservation.Description),
		CreatedAt:     reservation.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     reservation.UpdatedAt.UTC().Format(time.RFC3339),
		ReleasedAt:    timestampRFC3339(reservation.ReleasedAt),
	}
}

func serializeIssuanceBatches(batches []domain.IssuanceBatch) []issuanceBatchResponse {
	rows := make([]issuanceBatchResponse, 0, len(batches))
	for _, batch := range batches {
		rows = append(rows, *serializeIssuanceBatch(&batch))
	}
	return rows
}

func serializeIssuanceBatch(batch *domain.IssuanceBatch) *issuanceBatchResponse {
	if batch == nil {
		return nil
	}
	return &issuanceBatchResponse{
		ID:               batch.ID,
		ContractID:       batch.ContractID,
		CreatorUserID:    batch.CreatorUserID,
		CollateralLockID: batch.CollateralLockID,
		AboveQuantity:    batch.AboveQuantity,
		BelowQuantity:    batch.BelowQuantity,
		Status:           string(batch.Status),
		CreatedAt:        batch.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        batch.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializePositions(positions []*domain.Position) []positionResponse {
	rows := make([]positionResponse, 0, len(positions))
	for _, position := range positions {
		rows = append(rows, *serializePosition(position))
	}
	return rows
}

func serializePosition(position *domain.Position) *positionResponse {
	if position == nil {
		return nil
	}
	return &positionResponse{
		ID:                position.ID,
		UserID:            position.UserID,
		ContractID:        position.ContractID,
		Side:              string(position.Side),
		AvailableQuantity: position.AvailableQuantity,
		LockedQuantity:    position.LockedQuantity,
		TotalQuantity:     position.TotalQuantity,
		UpdatedAt:         position.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializePositionsFromValues(positions []domain.Position) []positionResponse {
	rows := make([]positionResponse, 0, len(positions))
	for _, position := range positions {
		positionCopy := position
		rows = append(rows, *serializePosition(&positionCopy))
	}
	return rows
}

func serializePositionLocks(locks []domain.PositionLock) []positionLockResponse {
	rows := make([]positionLockResponse, 0, len(locks))
	for _, lock := range locks {
		lockCopy := lock
		rows = append(rows, *serializePositionLock(&lockCopy))
	}
	return rows
}

func serializePositionLock(lock *domain.PositionLock) *positionLockResponse {
	if lock == nil {
		return nil
	}
	return &positionLockResponse{
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
	}
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

func ptrLedgerEntryResponse(entry ledgerEntryResponse) *ledgerEntryResponse {
	return &entry
}
