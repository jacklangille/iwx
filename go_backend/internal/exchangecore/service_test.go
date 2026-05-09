package exchangecore

import (
	"context"
	"testing"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
)

func TestSubmitCreateContractDerivesStationBackedFields(t *testing.T) {
	t.Parallel()

	var persisted commands.CreateContract
	service := &Service{
		repo: stubExchangeCoreRepository{
			processCreateContractFn: func(_ context.Context, envelope commands.CreateContractEnvelope) (commands.CreateContractResult, error) {
				persisted = envelope.Command
				return commands.CreateContractResult{}, nil
			},
		},
		stationCatalog: stubStationCatalog{
			findStationFn: func(_ context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
				return &domain.WeatherStation{
					ProviderName:     providerName,
					StationID:        stationID,
					Region:           "Boston",
					SupportedMetrics: []string{"temperature_max", "precipitation"},
					Active:           true,
				}, nil
			},
		},
	}

	_, err := service.SubmitCreateContract(context.Background(), commands.CreateContract{
		CreatorUserID:          17,
		Name:                   "Boston Heat Market",
		Metric:                 "temperature_max",
		DataProviderName:       "NOAA",
		StationID:              "KBOS",
		MeasurementUnit:        "deg_c",
		TradingPeriodStart:     "2026-07-01",
		TradingPeriodEnd:       "2026-07-31",
		MeasurementPeriodStart: "2026-07-01",
		MeasurementPeriodEnd:   "2026-07-31",
	})
	if err != nil {
		t.Fatalf("SubmitCreateContract() error = %v", err)
	}

	if persisted.Region != "Boston" {
		t.Fatalf("expected region to be derived from station, got %q", persisted.Region)
	}
	if persisted.DataProviderStationMode != "single_station" {
		t.Fatalf("expected single_station mode, got %q", persisted.DataProviderStationMode)
	}
}

func TestSubmitCreateContractRejectsUnsupportedStationMetric(t *testing.T) {
	t.Parallel()

	service := &Service{
		repo: stubExchangeCoreRepository{},
		stationCatalog: stubStationCatalog{
			findStationFn: func(_ context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
				return &domain.WeatherStation{
					ProviderName:     providerName,
					StationID:        stationID,
					Region:           "Boston",
					SupportedMetrics: []string{"precipitation"},
					Active:           true,
				}, nil
			},
		},
	}

	_, err := service.SubmitCreateContract(context.Background(), commands.CreateContract{
		CreatorUserID:    17,
		Name:             "Boston Heat Market",
		Metric:           "temperature_max",
		DataProviderName: "NOAA",
		StationID:        "KBOS",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Errors["metric"]) == 0 {
		t.Fatalf("expected metric validation error, got %#v", validationErr.Errors)
	}
}
