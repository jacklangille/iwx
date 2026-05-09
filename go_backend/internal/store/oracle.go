package store

import (
	"context"
	"time"

	"iwx/go_backend/internal/domain"
)

type OracleRepository interface {
	UpsertStation(ctx context.Context, input UpsertStationInput) (*domain.WeatherStation, error)
	ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error)
	RecordObservation(ctx context.Context, input RecordObservationInput) (*domain.OracleObservation, error)
	ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error)
	GetLatestResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error)
	InsertResolution(ctx context.Context, input domain.ContractResolution) (*domain.ContractResolution, error)
	ResolveContract(ctx context.Context, input ResolveContractInput) (*domain.ContractResolution, error)
}

type OracleProjectionSource interface {
	ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error)
	ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error)
	GetLatestResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error)
}

type StationCatalog interface {
	FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error)
}

type OracleProjectionTarget interface {
	ReplaceStationsProjection(ctx context.Context, stations []domain.WeatherStation) error
	ReplaceObservationsProjection(ctx context.Context, contractID int64, observations []domain.OracleObservation) error
	UpsertResolutionProjection(ctx context.Context, resolution domain.ContractResolution) error
}

type UpsertStationInput struct {
	ProviderName     string
	StationID        string
	DisplayName      string
	Region           string
	Latitude         *float64
	Longitude        *float64
	SupportedMetrics []string
	Active           bool
}

type RecordObservationInput struct {
	ContractID             int64
	ProviderName           string
	StationID              string
	ObservedMetric         string
	ObservationWindowStart time.Time
	ObservationWindowEnd   time.Time
	ObservedValue          string
	NormalizedValue        string
	ObservedAt             time.Time
}

type ResolveContractInput struct {
	ContractID int64
}
