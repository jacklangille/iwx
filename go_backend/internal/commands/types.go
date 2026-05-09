package commands

import (
	"time"

	"iwx/go_backend/internal/domain"
)

type PlaceOrder struct {
	ContractID               int64  `json:"contract_id"`
	UserID                   int64  `json:"user_id,omitempty"`
	TokenType                string `json:"token_type"`
	OrderSide                string `json:"order_side"`
	Price                    string `json:"price"`
	Quantity                 int64  `json:"quantity"`
	CashReservationID        *int64 `json:"cash_reservation_id,omitempty"`
	PositionLockID           *int64 `json:"position_lock_id,omitempty"`
	ReservationCorrelationID string `json:"reservation_correlation_id,omitempty"`
}

type CreateContract struct {
	CreatorUserID           int64  `json:"creator_user_id,omitempty"`
	Name                    string `json:"name"`
	Region                  string `json:"region"`
	Metric                  string `json:"metric"`
	Status                  string `json:"status"`
	Threshold               *int64 `json:"threshold,omitempty"`
	Multiplier              *int64 `json:"multiplier,omitempty"`
	MeasurementUnit         string `json:"measurement_unit"`
	TradingPeriodStart      string `json:"trading_period_start"`
	TradingPeriodEnd        string `json:"trading_period_end"`
	MeasurementPeriodStart  string `json:"measurement_period_start"`
	MeasurementPeriodEnd    string `json:"measurement_period_end"`
	DataProviderName        string `json:"data_provider_name"`
	StationID               string `json:"station_id"`
	DataProviderStationMode string `json:"data_provider_station_mode"`
	Description             string `json:"description"`
}

type PlaceOrderEnvelope struct {
	CommandID  string     `json:"command_id"`
	TraceID    string     `json:"trace_id,omitempty"`
	EnqueuedAt time.Time  `json:"enqueued_at"`
	Command    PlaceOrder `json:"command"`
}

type CreateContractEnvelope struct {
	CommandID  string         `json:"command_id"`
	TraceID    string         `json:"trace_id,omitempty"`
	EnqueuedAt time.Time      `json:"enqueued_at"`
	Command    CreateContract `json:"command"`
}

type PlaceOrderAccepted struct {
	CommandID  string    `json:"command_id"`
	ContractID int64     `json:"contract_id"`
	Partition  int       `json:"partition"`
	Status     string    `json:"status"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

type CreateContractAccepted struct {
	CommandID  string    `json:"command_id"`
	Partition  int       `json:"partition"`
	Status     string    `json:"status"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

type PlaceOrderResult struct {
	Status     string             `json:"status"`
	ContractID int64              `json:"contract_id"`
	Order      *domain.Order      `json:"order,omitempty"`
	Executions []domain.Execution `json:"executions,omitempty"`
	Sequence   *int64             `json:"sequence,omitempty"`
	AsOf       *string            `json:"as_of,omitempty"`
}

type CreateContractResult struct {
	Status   string           `json:"status"`
	Contract *domain.Contract `json:"contract,omitempty"`
}

type OrderCommand struct {
	CommandID     string
	ContractID    int64
	UserID        int64
	TokenType     string
	OrderSide     string
	Price         string
	Quantity      int64
	Status        string
	ErrorMessage  *string
	ResultStatus  *string
	ResultOrderID *int64
	EnqueuedAt    time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	UpdatedAt     time.Time
}

type ContractCommand struct {
	CommandID               string
	CreatorUserID           int64
	Name                    string
	Region                  string
	Metric                  string
	Status                  string
	Threshold               *int64
	Multiplier              *int64
	MeasurementUnit         *string
	TradingPeriodStart      *time.Time
	TradingPeriodEnd        *time.Time
	MeasurementPeriodStart  *time.Time
	MeasurementPeriodEnd    *time.Time
	DataProviderName        *string
	StationID               *string
	DataProviderStationMode *string
	Description             *string
	CommandStatus           string
	ErrorMessage            *string
	ResultContractID        *int64
	EnqueuedAt              time.Time
	StartedAt               *time.Time
	CompletedAt             *time.Time
	UpdatedAt               time.Time
}
