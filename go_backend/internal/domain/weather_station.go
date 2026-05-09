package domain

import "time"

type WeatherStation struct {
	ID               int64
	ProviderName     string
	StationID        string
	DisplayName      string
	Region           string
	Latitude         *float64
	Longitude        *float64
	SupportedMetrics []string
	Active           bool
	UpdatedAt        time.Time
}
