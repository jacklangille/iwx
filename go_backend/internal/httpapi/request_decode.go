package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/market"
)

func decodePlaceOrderCommand(body io.Reader) (commands.PlaceOrder, error) {
	var payload map[string]json.RawMessage

	if body == nil {
		return commands.PlaceOrder{}, errors.New("request body is required")
	}

	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return commands.PlaceOrder{}, err
	}

	command := commands.PlaceOrder{}

	if raw := payload["contract_id"]; raw != nil {
		if err := json.Unmarshal(raw, &command.ContractID); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["token_type"]; raw != nil {
		if err := json.Unmarshal(raw, &command.TokenType); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["order_side"]; raw != nil {
		if err := json.Unmarshal(raw, &command.OrderSide); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["quantity"]; raw != nil {
		if err := json.Unmarshal(raw, &command.Quantity); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["price"]; raw != nil {
		command.Price = decodePrice(raw)
	}

	return command, nil
}

func decodePrice(raw json.RawMessage) string {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}

	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber.String()
	}

	return ""
}

func chartConfigFromRequest(r *http.Request) domain.ChartConfig {
	var lookback *int64
	var bucket *int64

	if raw := r.URL.Query().Get("lookback_seconds"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			lookback = &parsed
		}
	}

	if raw := r.URL.Query().Get("bucket_seconds"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			bucket = &parsed
		}
	}

	return market.NormalizeChartConfig(lookback, bucket)
}
