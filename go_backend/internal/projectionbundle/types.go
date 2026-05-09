package projectionbundle

import (
	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

type ContractBundle struct {
	Contract *domain.Contract     `json:"contract,omitempty"`
	Rule     *domain.ContractRule `json:"rule,omitempty"`
}

type UserStateBundle struct {
	UserID           int64                         `json:"user_id"`
	CashAccounts     []domain.CashAccount          `json:"cash_accounts,omitempty"`
	Positions        []domain.Position             `json:"positions,omitempty"`
	PositionLocks    []domain.PositionLock         `json:"position_locks,omitempty"`
	CollateralLocks  []domain.CollateralLock       `json:"collateral_locks,omitempty"`
	CashReservations []domain.OrderCashReservation `json:"cash_reservations,omitempty"`
	Settlements      []domain.SettlementEntry      `json:"settlements,omitempty"`
}

type SettlementBundle struct {
	ContractID int64                    `json:"contract_id"`
	Entries    []domain.SettlementEntry `json:"entries,omitempty"`
}

type MarketBundle struct {
	ContractID int64                   `json:"contract_id"`
	Orders     []domain.Order          `json:"orders,omitempty"`
	Executions []domain.Execution      `json:"executions,omitempty"`
	Snapshots  []domain.MarketSnapshot `json:"snapshots,omitempty"`
}

type OrderCommandBundle struct {
	Command *commands.OrderCommand `json:"command,omitempty"`
}

type OracleStateBundle struct {
	ContractID   int64                      `json:"contract_id"`
	Observations []domain.OracleObservation `json:"observations,omitempty"`
	Resolution   *domain.ContractResolution `json:"resolution,omitempty"`
}

type StationCatalogBundle struct {
	Stations []domain.WeatherStation `json:"stations,omitempty"`
}
