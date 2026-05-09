package domain

import "time"

type Contract struct {
	ID                      int64
	CreatorUserID           *int64
	Name                    string
	Region                  string
	Metric                  string
	Status                  string
	Threshold               *int64
	Multiplier              *int64
	MeasurementUnit         string
	TradingPeriodStart      *time.Time
	TradingPeriodEnd        *time.Time
	MeasurementPeriodStart  *time.Time
	MeasurementPeriodEnd    *time.Time
	DataProviderName        string
	StationID               string
	DataProviderStationMode string
	Description             string
	UpdatedAt               time.Time
}

type Order struct {
	ID                       int64
	ContractID               int64
	UserID                   int64
	TokenType                string
	OrderSide                string
	Price                    string
	Quantity                 int64
	Status                   string
	CashReservationID        *int64
	PositionLockID           *int64
	ReservationCorrelationID string
	InsertedAt               time.Time
	UpdatedAt                time.Time
}

type MarketSnapshot struct {
	ID         int64
	ContractID int64
	BestAbove  *string
	BestBelow  *string
	MidAbove   *string
	MidBelow   *string
	InsertedAt time.Time
}

type PriceLevel struct {
	Price    string
	Quantity int64
}

type OrderBookSide struct {
	Bid []PriceLevel
	Ask []PriceLevel
}

type OrderBook struct {
	Above OrderBookSide
	Below OrderBookSide
}

type BestQuotes struct {
	Above QuotePair
	Below QuotePair
}

type QuotePair struct {
	Bid *string
	Ask *string
}

type MidQuotes struct {
	Above *string
	Below *string
}

type LiquidityTotals struct {
	Above int64
	Below int64
}

type MarketSummary struct {
	Best             BestQuotes
	Mid              MidQuotes
	Liquidity        LiquidityTotals
	AboveBelowBidGap *string
	MidPrice         *string
}

type MarketState struct {
	ContractID int64
	Sequence   *int64
	AsOf       *time.Time
	OrderBook  OrderBook
	Summary    MarketSummary
}

type ChartPoint struct {
	BucketStart time.Time
	InsertedAt  time.Time
	MidAbove    *string
	MidBelow    *string
	BestAbove   *string
	BestBelow   *string
}

type ChartConfig struct {
	LookbackSeconds int64
	BucketSeconds   int64
}
