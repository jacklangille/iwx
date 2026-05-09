package oraclehttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/oracle"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/internal/store"
)

func TestInternalStationLookupReturnsStation(t *testing.T) {
	t.Parallel()

	service := oracle.NewService(stubOracleRepository{
		findStationFn: func(_ context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
			if providerName != "NOAA" {
				t.Fatalf("expected provider NOAA, got %q", providerName)
			}
			if stationID != "HALI" {
				t.Fatalf("expected station HALI, got %q", stationID)
			}
			return &domain.WeatherStation{
				ID:               7,
				ProviderName:     providerName,
				StationID:        stationID,
				DisplayName:      "Halifax Stanfield",
				Region:           "Nova Scotia",
				SupportedMetrics: []string{"temperature_max"},
				Active:           true,
			}, nil
		},
	}, nil, nil, nil)
	server := NewServer(config.Config{}, service)

	request := httptest.NewRequest(http.MethodGet, "/internal/stations/NOAA/HALI", nil)
	recorder := httptest.NewRecorder()

	server.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		ProviderName string `json:"provider_name"`
		StationID    string `json:"station_id"`
		DisplayName  string `json:"display_name"`
		Region       string `json:"region"`
		Active       bool   `json:"active"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ProviderName != "NOAA" || payload.StationID != "HALI" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.DisplayName != "Halifax Stanfield" || payload.Region != "Nova Scotia" || !payload.Active {
		t.Fatalf("unexpected payload fields: %#v", payload)
	}
}

func TestInternalStationLookupReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := oracle.NewService(stubOracleRepository{}, nil, nil, nil)
	server := NewServer(config.Config{}, service)

	request := httptest.NewRequest(http.MethodGet, "/internal/stations/NOAA/MISSING", nil)
	recorder := httptest.NewRecorder()

	server.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

type stubOracleRepository struct {
	findStationFn func(context.Context, string, string) (*domain.WeatherStation, error)
}

func (s stubOracleRepository) UpsertStation(context.Context, store.UpsertStationInput) (*domain.WeatherStation, error) {
	return nil, nil
}

func (s stubOracleRepository) ListStations(context.Context, bool) ([]domain.WeatherStation, error) {
	return nil, nil
}

func (s stubOracleRepository) FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
	if s.findStationFn != nil {
		return s.findStationFn(ctx, providerName, stationID)
	}
	return nil, nil
}

func (s stubOracleRepository) RecordObservation(context.Context, store.RecordObservationInput) (*domain.OracleObservation, error) {
	return nil, nil
}

func (s stubOracleRepository) ListObservations(context.Context, int64, int) ([]domain.OracleObservation, error) {
	return nil, nil
}

func (s stubOracleRepository) GetLatestResolution(context.Context, int64) (*domain.ContractResolution, error) {
	return nil, nil
}

func (s stubOracleRepository) InsertResolution(context.Context, domain.ContractResolution) (*domain.ContractResolution, error) {
	return nil, nil
}

func (s stubOracleRepository) MarkResolutionPublished(context.Context, string, time.Time) error {
	return nil
}

func (s stubOracleRepository) EnqueueOutboxEvent(context.Context, outbox.Event) error {
	return nil
}

func (s stubOracleRepository) ListPendingOutboxEvents(context.Context, int) ([]outbox.Event, error) {
	return nil, nil
}

func (s stubOracleRepository) MarkOutboxEventPublished(context.Context, string, time.Time) error {
	return nil
}

func (s stubOracleRepository) MarkOutboxEventFailed(context.Context, string, string) error {
	return nil
}

func (s stubOracleRepository) ResolveContract(context.Context, store.ResolveContractInput) (*domain.ContractResolution, error) {
	return nil, nil
}
