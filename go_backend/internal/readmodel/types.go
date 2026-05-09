package readmodel

import (
	"time"

	"iwx/go_backend/internal/domain"
)

type ContractSummary struct {
	Contract domain.Contract
	Market   ContractMarket
}

type ContractMarket struct {
	AsOf     *time.Time
	Sequence *int64
	Summary  domain.MarketSummary
}
