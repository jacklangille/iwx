package oraclehttp

import (
	"io"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/httpjson"
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
	if err := httpjson.DecodeStrict(body, &request); err != nil {
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

type observationResponse struct {
	ID                     int64  `json:"id"`
	ContractID             int64  `json:"contract_id"`
	ProviderName           string `json:"provider_name"`
	StationID              string `json:"station_id"`
	ObservedMetric         string `json:"observed_metric"`
	ObservationWindowStart string `json:"observation_window_start"`
	ObservationWindowEnd   string `json:"observation_window_end"`
	ObservedValue          string `json:"observed_value"`
	NormalizedValue        string `json:"normalized_value"`
	ObservedAt             string `json:"observed_at"`
	RecordedAt             string `json:"recorded_at"`
}

type observationsResponse struct {
	ContractID   int64                 `json:"contract_id"`
	Observations []observationResponse `json:"observations"`
}

type resolutionResponse struct {
	ID                     int64  `json:"id"`
	ContractID             int64  `json:"contract_id"`
	ProviderName           string `json:"provider_name"`
	StationID              string `json:"station_id"`
	ObservedMetric         string `json:"observed_metric"`
	ObservationWindowStart string `json:"observation_window_start"`
	ObservationWindowEnd   string `json:"observation_window_end"`
	RuleVersion            string `json:"rule_version"`
	ResolvedValue          string `json:"resolved_value"`
	Outcome                string `json:"outcome"`
	ResolvedAt             string `json:"resolved_at"`
}

func serializeObservations(observations []domain.OracleObservation) []observationResponse {
	rows := make([]observationResponse, 0, len(observations))
	for _, observation := range observations {
		rows = append(rows, serializeObservation(observation))
	}
	return rows
}

func serializeObservation(observation domain.OracleObservation) observationResponse {
	return observationResponse{
		ID:                     observation.ID,
		ContractID:             observation.ContractID,
		ProviderName:           observation.ProviderName,
		StationID:              observation.StationID,
		ObservedMetric:         observation.ObservedMetric,
		ObservationWindowStart: observation.ObservationWindowStart.UTC().Format(time.RFC3339),
		ObservationWindowEnd:   observation.ObservationWindowEnd.UTC().Format(time.RFC3339),
		ObservedValue:          observation.ObservedValue,
		NormalizedValue:        observation.NormalizedValue,
		ObservedAt:             observation.ObservedAt.UTC().Format(time.RFC3339),
		RecordedAt:             observation.RecordedAt.UTC().Format(time.RFC3339),
	}
}

func serializeResolution(resolution domain.ContractResolution) resolutionResponse {
	return resolutionResponse{
		ID:                     resolution.ID,
		ContractID:             resolution.ContractID,
		ProviderName:           resolution.ProviderName,
		StationID:              resolution.StationID,
		ObservedMetric:         resolution.ObservedMetric,
		ObservationWindowStart: resolution.ObservationWindowStart.UTC().Format(time.RFC3339),
		ObservationWindowEnd:   resolution.ObservationWindowEnd.UTC().Format(time.RFC3339),
		RuleVersion:            resolution.RuleVersion,
		ResolvedValue:          resolution.ResolvedValue,
		Outcome:                string(resolution.Outcome),
		ResolvedAt:             resolution.ResolvedAt.UTC().Format(time.RFC3339),
	}
}

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}
