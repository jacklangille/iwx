package market

import (
	"sort"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/money"
)

type Components struct {
	Orders    []domain.Order
	OrderBook domain.OrderBook
	Best      domain.BestQuotes
	Mid       domain.MidQuotes
}

func BuildComponents(orders []domain.Order) Components {
	orderBook := domain.OrderBook{
		Above: domain.OrderBookSide{
			Bid: aggregateLevels(orders, "above", "bid", true),
			Ask: aggregateLevels(orders, "above", "ask", false),
		},
		Below: domain.OrderBookSide{
			Bid: aggregateLevels(orders, "below", "bid", true),
			Ask: aggregateLevels(orders, "below", "ask", false),
		},
	}

	best := bestQuotes(orderBook)

	return Components{
		Orders:    orders,
		OrderBook: orderBook,
		Best:      best,
		Mid: domain.MidQuotes{
			Above: midpoint(best.Above.Bid, best.Above.Ask),
			Below: midpoint(best.Below.Bid, best.Below.Ask),
		},
	}
}

func Summary(orders []domain.Order) domain.MarketSummary {
	components := BuildComponents(orders)

	return domain.MarketSummary{
		Best:             components.Best,
		Mid:              components.Mid,
		Liquidity:        LiquidityTotals(orders),
		AboveBelowBidGap: AboveBelowBidGap(components.Best),
		MidPrice:         midPrice(components.Mid),
	}
}

func LiquidityTotals(orders []domain.Order) domain.LiquidityTotals {
	totals := domain.LiquidityTotals{}

	for _, order := range orders {
		switch order.TokenType {
		case "above":
			totals.Above += order.Quantity
		case "below":
			totals.Below += order.Quantity
		}
	}

	return totals
}

func AboveBelowBidGap(best domain.BestQuotes) *string {
	if best.Above.Bid == nil || best.Below.Bid == nil {
		return nil
	}

	above, err := money.ParseCents(*best.Above.Bid)
	if err != nil {
		return nil
	}

	below, err := money.ParseCents(*best.Below.Bid)
	if err != nil {
		return nil
	}

	value := money.FormatCents(money.Abs(above - below))
	return &value
}

func midpoint(bid, ask *string) *string {
	switch {
	case bid != nil && ask != nil:
		bidCents, err := money.ParseCents(*bid)
		if err != nil {
			return nil
		}

		askCents, err := money.ParseCents(*ask)
		if err != nil {
			return nil
		}

		value := money.FormatCents(money.Average(bidCents, askCents))
		return &value
	case bid != nil:
		return bid
	case ask != nil:
		return ask
	default:
		return nil
	}
}

func midPrice(mid domain.MidQuotes) *string {
	switch {
	case mid.Above != nil && mid.Below != nil:
		above, err := money.ParseCents(*mid.Above)
		if err != nil {
			return nil
		}

		below, err := money.ParseCents(*mid.Below)
		if err != nil {
			return nil
		}

		value := money.FormatCents(money.Average(above, below))
		return &value
	case mid.Above != nil:
		return mid.Above
	case mid.Below != nil:
		return mid.Below
	default:
		return nil
	}
}

func bestQuotes(orderBook domain.OrderBook) domain.BestQuotes {
	return domain.BestQuotes{
		Above: domain.QuotePair{
			Bid: bestLevelPrice(orderBook.Above.Bid),
			Ask: bestLevelPrice(orderBook.Above.Ask),
		},
		Below: domain.QuotePair{
			Bid: bestLevelPrice(orderBook.Below.Bid),
			Ask: bestLevelPrice(orderBook.Below.Ask),
		},
	}
}

func bestLevelPrice(levels []domain.PriceLevel) *string {
	if len(levels) == 0 {
		return nil
	}

	value := levels[0].Price
	return &value
}

func aggregateLevels(orders []domain.Order, tokenType, orderSide string, descending bool) []domain.PriceLevel {
	levels := map[string]domain.PriceLevel{}

	for _, order := range orders {
		if order.TokenType != tokenType || order.OrderSide != orderSide {
			continue
		}

		level := levels[order.Price]
		level.Price = order.Price
		level.Quantity += order.Quantity
		levels[order.Price] = level
	}

	rows := make([]domain.PriceLevel, 0, len(levels))
	for _, level := range levels {
		rows = append(rows, level)
	}

	sort.Slice(rows, func(i, j int) bool {
		left, leftErr := money.ParseCents(rows[i].Price)
		right, rightErr := money.ParseCents(rows[j].Price)
		if leftErr != nil || rightErr != nil {
			if descending {
				return rows[i].Price > rows[j].Price
			}
			return rows[i].Price < rows[j].Price
		}

		if descending {
			return left > right
		}
		return left < right
	})

	return rows
}
