package events

import (
	"time"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

type ContractActivated struct {
	ContractID int64
	UserID     int64
	OccurredAt time.Time
}

type CollateralLocked struct {
	ContractID       int64
	CollateralLockID int64
	UserID           int64
	AmountCents      int64
	OccurredAt       time.Time
}

type IssuanceCompleted struct {
	ContractID      int64
	IssuanceBatchID int64
	CreatorUserID   int64
	AboveQuantity   int64
	BelowQuantity   int64
	OccurredAt      time.Time
}

type OrderAccepted struct {
	OrderID    int64
	ContractID int64
	UserID     int64
	OccurredAt time.Time
}

type ExecutionCreated struct {
	ExecutionID            string
	ContractID             int64
	TokenType              string
	BuyOrderID             int64
	SellOrderID            int64
	BuyerUserID            int64
	SellerUserID           int64
	BuyerCashReservationID *int64
	SellerPositionLockID   *int64
	Price                  string
	Quantity               int64
	Sequence               int64
	TraceID                string
	OccurredAt             time.Time
}

type ContractResolved struct {
	EventID    string
	ContractID int64
	Outcome    string
	TraceID    string
	ResolvedAt time.Time
}

type SettlementCompleted struct {
	EventID    string
	ContractID int64
	TraceID    string
	SettledAt  time.Time
}

type ProjectionChangeKind string

const (
	ProjectionChangeContract       ProjectionChangeKind = "contract_changed"
	ProjectionChangeUserState      ProjectionChangeKind = "user_state_changed"
	ProjectionChangeOracleState    ProjectionChangeKind = "oracle_state_changed"
	ProjectionChangeStationCatalog ProjectionChangeKind = "station_catalog_changed"
	ProjectionChangeSettlement     ProjectionChangeKind = "settlement_changed"
	ProjectionChangeMarket         ProjectionChangeKind = "market_changed"
	ProjectionChangeOrderCommand   ProjectionChangeKind = "order_command_changed"
)

type ProjectionChange struct {
	EventID    string               `json:"event_id"`
	Kind       ProjectionChangeKind `json:"kind"`
	TraceID    string               `json:"trace_id,omitempty"`
	ContractID int64                `json:"contract_id,omitempty"`
	UserID     int64                `json:"user_id,omitempty"`
	CommandID  string               `json:"command_id,omitempty"`
	Version    int64                `json:"version"`
	OccurredAt time.Time            `json:"occurred_at"`
}

type ReadModelProjection struct {
	EventID             string                        `json:"event_id"`
	Kind                ProjectionChangeKind          `json:"kind"`
	TraceID             string                        `json:"trace_id,omitempty"`
	ContractID          int64                         `json:"contract_id,omitempty"`
	UserID              int64                         `json:"user_id,omitempty"`
	CommandID           string                        `json:"command_id,omitempty"`
	Contract            *domain.Contract              `json:"contract,omitempty"`
	CashAccounts        []domain.CashAccount          `json:"cash_accounts,omitempty"`
	Positions           []domain.Position             `json:"positions,omitempty"`
	PositionLocks       []domain.PositionLock         `json:"position_locks,omitempty"`
	CollateralLocks     []domain.CollateralLock       `json:"collateral_locks,omitempty"`
	CashReservations    []domain.OrderCashReservation `json:"cash_reservations,omitempty"`
	UserSettlements     []domain.SettlementEntry      `json:"user_settlements,omitempty"`
	ContractSettlements []domain.SettlementEntry      `json:"contract_settlements,omitempty"`
	Observations        []domain.OracleObservation    `json:"observations,omitempty"`
	Resolution          *domain.ContractResolution    `json:"resolution,omitempty"`
	Stations            []domain.WeatherStation       `json:"stations,omitempty"`
	Orders              []domain.Order                `json:"orders,omitempty"`
	Executions          []domain.Execution            `json:"executions,omitempty"`
	Snapshots           []domain.MarketSnapshot       `json:"snapshots,omitempty"`
	OrderCommand        *commands.OrderCommand        `json:"order_command,omitempty"`
	Version             int64                         `json:"version"`
	OccurredAt          time.Time                     `json:"occurred_at"`
}
