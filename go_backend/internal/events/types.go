package events

import "time"

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
	ContractID int64
	Outcome    string
	TraceID    string
	ResolvedAt time.Time
}

type SettlementCompleted struct {
	ContractID int64
	TraceID    string
	SettledAt  time.Time
}
