package domain

type ContractState string

const (
	ContractStateDraft              ContractState = "draft"
	ContractStatePendingApproval    ContractState = "pending_approval"
	ContractStatePendingCollateral  ContractState = "pending_collateral"
	ContractStateActive             ContractState = "active"
	ContractStateTradingClosed      ContractState = "trading_closed"
	ContractStateAwaitingResolution ContractState = "awaiting_resolution"
	ContractStateResolved           ContractState = "resolved"
	ContractStateSettled            ContractState = "settled"
	ContractStateCancelled          ContractState = "cancelled"
)

func ValidContractStates() []ContractState {
	return []ContractState{
		ContractStateDraft,
		ContractStatePendingApproval,
		ContractStatePendingCollateral,
		ContractStateActive,
		ContractStateTradingClosed,
		ContractStateAwaitingResolution,
		ContractStateResolved,
		ContractStateSettled,
		ContractStateCancelled,
	}
}

func IsValidContractState(value string) bool {
	state := ContractState(value)
	for _, allowed := range ValidContractStates() {
		if state == allowed {
			return true
		}
	}

	return false
}

type ClaimSide string

const (
	ClaimSideAbove ClaimSide = "above"
	ClaimSideBelow ClaimSide = "below"
)

func ValidClaimSides() []ClaimSide {
	return []ClaimSide{ClaimSideAbove, ClaimSideBelow}
}

type OrderSide string

const (
	OrderSideBid OrderSide = "bid"
	OrderSideAsk OrderSide = "ask"
)

type OrderStatus string

const (
	OrderStatusOpen            OrderStatus = "open"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusRejected        OrderStatus = "rejected"
)

type LedgerEntryType string

const (
	LedgerEntryDeposit           LedgerEntryType = "deposit"
	LedgerEntryWithdrawal        LedgerEntryType = "withdrawal"
	LedgerEntryCollateralLock    LedgerEntryType = "collateral_lock"
	LedgerEntryCollateralRelease LedgerEntryType = "collateral_release"
	LedgerEntryOrderCashReserve  LedgerEntryType = "order_cash_reserve"
	LedgerEntryOrderCashRelease  LedgerEntryType = "order_cash_release"
	LedgerEntryTradeDebit        LedgerEntryType = "trade_debit"
	LedgerEntryTradeCredit       LedgerEntryType = "trade_credit"
	LedgerEntrySettlementCredit  LedgerEntryType = "settlement_credit"
	LedgerEntrySettlementDebit   LedgerEntryType = "settlement_debit"
)

type CollateralLockStatus string

const (
	CollateralLockStatusPending   CollateralLockStatus = "pending"
	CollateralLockStatusLocked    CollateralLockStatus = "locked"
	CollateralLockStatusReleased  CollateralLockStatus = "released"
	CollateralLockStatusConsumed  CollateralLockStatus = "consumed"
	CollateralLockStatusCancelled CollateralLockStatus = "cancelled"
)

type IssuanceBatchStatus string

const (
	IssuanceBatchStatusPending   IssuanceBatchStatus = "pending"
	IssuanceBatchStatusIssued    IssuanceBatchStatus = "issued"
	IssuanceBatchStatusCancelled IssuanceBatchStatus = "cancelled"
)

type PositionLockStatus string

const (
	PositionLockStatusActive   PositionLockStatus = "active"
	PositionLockStatusReleased PositionLockStatus = "released"
	PositionLockStatusConsumed PositionLockStatus = "consumed"
)

type CashReservationStatus string

const (
	CashReservationStatusActive   CashReservationStatus = "active"
	CashReservationStatusReleased CashReservationStatus = "released"
	CashReservationStatusConsumed CashReservationStatus = "consumed"
)

type ResolutionOutcome string

const (
	ResolutionOutcomeAbove     ResolutionOutcome = "above"
	ResolutionOutcomeBelow     ResolutionOutcome = "below"
	ResolutionOutcomeCancelled ResolutionOutcome = "cancelled"
)

type SettlementEntryType string

const (
	SettlementEntryPayout            SettlementEntryType = "payout"
	SettlementEntryCollateralRelease SettlementEntryType = "collateral_release"
	SettlementEntryRefund            SettlementEntryType = "refund"
)
