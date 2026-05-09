package oraclehttp

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

type recordObservationRequest struct {
	ContractID             int64  `json:"contract_id"`
	ProviderName           string `json:"provider_name"`
	StationID              string `json:"station_id"`
	ObservedMetric         string `json:"observed_metric"`
	ObservationWindowStart string `json:"observation_window_start"`
	ObservationWindowEnd   string `json:"observation_window_end"`
	ObservedValue          string `json:"observed_value"`
	NormalizedValue        string `json:"normalized_value"`
	ObservedAt             string `json:"observed_at"`
}

func decodeObservationInput(body io.Reader) (store.RecordObservationInput, error) {
	var request recordObservationRequest
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		return store.RecordObservationInput{}, err
	}

	windowStart, err := parseRFC3339(request.ObservationWindowStart)
	if err != nil {
		return store.RecordObservationInput{}, err
	}
	windowEnd, err := parseRFC3339(request.ObservationWindowEnd)
	if err != nil {
		return store.RecordObservationInput{}, err
	}
	observedAt, err := parseRFC3339(request.ObservedAt)
	if err != nil {
		return store.RecordObservationInput{}, err
	}

	return store.RecordObservationInput{
		ContractID:             request.ContractID,
		ProviderName:           strings.TrimSpace(request.ProviderName),
		StationID:              strings.TrimSpace(request.StationID),
		ObservedMetric:         strings.TrimSpace(request.ObservedMetric),
		ObservationWindowStart: windowStart,
		ObservationWindowEnd:   windowEnd,
		ObservedValue:          strings.TrimSpace(request.ObservedValue),
		NormalizedValue:        strings.TrimSpace(request.NormalizedValue),
		ObservedAt:             observedAt,
	}, nil
}

func serializeObservations(observations []domain.OracleObservation) []map[string]any {
	rows := make([]map[string]any, 0, len(observations))
	for _, observation := range observations {
		rows = append(rows, serializeObservation(observation))
	}
	return rows
}

func serializeObservation(observation domain.OracleObservation) map[string]any {
	return map[string]any{
		"id":                       observation.ID,
		"contract_id":              observation.ContractID,
		"provider_name":            observation.ProviderName,
		"station_id":               observation.StationID,
		"observed_metric":          observation.ObservedMetric,
		"observation_window_start": observation.ObservationWindowStart.UTC().Format(time.RFC3339),
		"observation_window_end":   observation.ObservationWindowEnd.UTC().Format(time.RFC3339),
		"observed_value":           observation.ObservedValue,
		"normalized_value":         observation.NormalizedValue,
		"observed_at":              observation.ObservedAt.UTC().Format(time.RFC3339),
		"recorded_at":              observation.RecordedAt.UTC().Format(time.RFC3339),
	}
}

func serializeResolution(resolution domain.ContractResolution) map[string]any {
	return map[string]any{
		"id":                       resolution.ID,
		"contract_id":              resolution.ContractID,
		"provider_name":            resolution.ProviderName,
		"station_id":               resolution.StationID,
		"observed_metric":          resolution.ObservedMetric,
		"observation_window_start": resolution.ObservationWindowStart.UTC().Format(time.RFC3339),
		"observation_window_end":   resolution.ObservationWindowEnd.UTC().Format(time.RFC3339),
		"rule_version":             resolution.RuleVersion,
		"resolved_value":           resolution.ResolvedValue,
		"outcome":                  resolution.Outcome,
		"resolved_at":              resolution.ResolvedAt.UTC().Format(time.RFC3339),
	}
}

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}
