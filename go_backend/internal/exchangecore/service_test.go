package exchangecore

import (
	"context"
	"testing"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
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

func TestSubmitCreateContractRejectsDuplicateMarket(t *testing.T) {
	t.Parallel()

	service := &Service{
		repo: stubExchangeCoreRepository{
			findDuplicateContractFn: func(_ context.Context, input store.FindDuplicateContractInput) (*domain.Contract, error) {
				if input.ProviderName != "NOAA" || input.StationID != "KBOS" || input.Metric != "temperature_max" {
					t.Fatalf("duplicate check got unexpected input: %#v", input)
				}
				return &domain.Contract{ID: 42, Status: string(domain.ContractStateActive)}, nil
			},
		},
		stationCatalog: stubStationCatalog{
			findStationFn: func(_ context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
				return &domain.WeatherStation{
					ProviderName:     providerName,
					StationID:        stationID,
					Region:           "Boston",
					SupportedMetrics: []string{"temperature_max"},
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
		TradingPeriodStart:     "2026-07-01",
		TradingPeriodEnd:       "2026-07-31",
		MeasurementPeriodStart: "2026-07-01",
		MeasurementPeriodEnd:   "2026-07-31",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Errors["contract"]) == 0 {
		t.Fatalf("expected duplicate-market validation error, got %#v", validationErr.Errors)
	}
}
