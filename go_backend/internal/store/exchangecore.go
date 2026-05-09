package store

import (
	"context"
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
)

type ExchangeCoreRepository interface {
	ProcessCreateContract(ctx context.Context, envelope commands.CreateContractEnvelope) (commands.CreateContractResult, error)
	FindDuplicateContract(ctx context.Context, input FindDuplicateContractInput) (*domain.Contract, error)
	GetContractCommand(ctx context.Context, commandID string) (*commands.ContractCommand, error)
	GetContract(ctx context.Context, contractID int64) (*domain.Contract, error)
	GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error)
	UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error)
	ListContractCollateralLocks(ctx context.Context, userID, contractID int64) ([]domain.CollateralLock, error)
	ListIssuanceBatches(ctx context.Context, userID, contractID int64) ([]domain.IssuanceBatch, error)
	CreateIssuanceBatch(ctx context.Context, input CreateIssuanceBatchInput) (*domain.IssuanceBatch, *domain.CollateralLock, []*domain.Position, error)
	ListPositions(ctx context.Context, userID int64, contractID *int64) ([]domain.Position, error)
	ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error)
	ListCashAccounts(ctx context.Context, userID int64) ([]domain.CashAccount, error)
	ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error)
	ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error)
	CreatePositionLock(ctx context.Context, input CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error)
	ReleasePositionLock(ctx context.Context, input ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error)
	GetCashAccount(ctx context.Context, userID int64, currency string) (*domain.CashAccount, error)
	ListLedgerEntries(ctx context.Context, userID int64, currency string, limit int) ([]domain.LedgerEntry, error)
	DepositCash(ctx context.Context, input DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error)
	WithdrawCash(ctx context.Context, input WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error)
	ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error)
	CreateCollateralLock(ctx context.Context, input CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error)
	ReleaseCollateralLock(ctx context.Context, input ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error)
	ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error)
	CreateOrderCashReservation(ctx context.Context, input CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error)
	ReleaseOrderCashReservation(ctx context.Context, input ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error)
	ApplyExecution(ctx context.Context, event events.ExecutionCreated) (*ExecutionApplicationResult, error)
	SettleContract(ctx context.Context, input SettleContractInput) (*SettlementResult, error)
}

type DepositCashInput struct {
	UserID        int64
	Currency      string
	AmountCents   int64
	ReferenceID   string
	CorrelationID string
	Description   string
}

type WithdrawCashInput struct {
	UserID        int64
	Currency      string
	AmountCents   int64
	ReferenceID   string
	CorrelationID string
	Description   string
}

type CreateCollateralLockInput struct {
	UserID        int64
	ContractID    int64
	Currency      string
	AmountCents   int64
	ReferenceID   string
	CorrelationID string
	Description   string
}

type ReleaseCollateralLockInput struct {
	UserID        int64
	LockID        int64
	CorrelationID string
	Description   string
}

type CreateOrderCashReservationInput struct {
	UserID        int64
	ContractID    int64
	Currency      string
	AmountCents   int64
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
}

type ReleaseOrderCashReservationInput struct {
	UserID        int64
	ReservationID int64
	CorrelationID string
	Description   string
}

type CreateIssuanceBatchInput struct {
	UserID           int64
	ContractID       int64
	CollateralLockID int64
	PairedQuantity   int64
}

type FindDuplicateContractInput struct {
	ProviderName           string
	StationID              string
	Metric                 string
	Threshold              *int64
	TradingPeriodStart     string
	TradingPeriodEnd       string
	MeasurementPeriodStart string
	MeasurementPeriodEnd   string
}

type CreatePositionLockInput struct {
	UserID        int64
	ContractID    int64
	Side          string
	Quantity      int64
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
}

type ReleasePositionLockInput struct {
	UserID        int64
	LockID        int64
	CorrelationID string
	Description   string
}

type SettleContractInput struct {
	ContractID    int64
	EventID       string
	Outcome       string
	ResolvedAt    string
	CorrelationID string
}

type SettlementResult struct {
	Contract      *domain.Contract
	Entries       []domain.SettlementEntry
	AffectedUsers []int64
	SettledAt     time.Time
}

type ExecutionApplicationResult struct {
	ExecutionID   string
	ContractID    int64
	BuyerUserID   int64
	SellerUserID  int64
	AffectedUsers []int64
	Applied       bool
	AppliedAt     time.Time
}
