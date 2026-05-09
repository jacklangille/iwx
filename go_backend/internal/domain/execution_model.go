package domain

import "time"

type Execution struct {
	ID                     int64
	ExecutionID            string
	CommandID              string
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
	OccurredAt             time.Time
}
