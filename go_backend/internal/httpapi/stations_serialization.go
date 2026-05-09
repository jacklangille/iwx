package httpapi

import "iwx/go_backend/internal/domain"

func serializeStation(station domain.WeatherStation) map[string]any {
	return map[string]any{
		"id":                station.ID,
		"provider_name":     station.ProviderName,
		"station_id":        station.StationID,
		"display_name":      station.DisplayName,
		"region":            station.Region,
		"latitude":          station.Latitude,
		"longitude":         station.Longitude,
		"supported_metrics": station.SupportedMetrics,
		"active":            station.Active,
		"updated_at":        station.UpdatedAt,
	}
}

func serializeStations(stations []domain.WeatherStation) []map[string]any {
	items := make([]map[string]any, 0, len(stations))
	for _, station := range stations {
		items = append(items, serializeStation(station))
	}
	return items
}
