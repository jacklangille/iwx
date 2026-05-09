package oraclehttp

import (
	"time"

	"iwx/go_backend/internal/domain"
)

type stationResponse struct {
	ID               int64    `json:"id"`
	ProviderName     string   `json:"provider_name"`
	StationID        string   `json:"station_id"`
	DisplayName      string   `json:"display_name"`
	Region           string   `json:"region"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	SupportedMetrics []string `json:"supported_metrics"`
	Active           bool     `json:"active"`
	UpdatedAt        string   `json:"updated_at"`
}

type stationsResponse struct {
	Stations []stationResponse `json:"stations"`
}

func serializeStation(station domain.WeatherStation) stationResponse {
	return stationResponse{
		ID:               station.ID,
		ProviderName:     station.ProviderName,
		StationID:        station.StationID,
		DisplayName:      station.DisplayName,
		Region:           station.Region,
		Latitude:         station.Latitude,
		Longitude:        station.Longitude,
		SupportedMetrics: station.SupportedMetrics,
		Active:           station.Active,
		UpdatedAt:        station.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeStations(stations []domain.WeatherStation) []stationResponse {
	items := make([]stationResponse, 0, len(stations))
	for _, station := range stations {
		items = append(items, serializeStation(station))
	}
	return items
}
