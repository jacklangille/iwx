package httpapi

import (
	"net/http"
	"strconv"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/market"
)

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
