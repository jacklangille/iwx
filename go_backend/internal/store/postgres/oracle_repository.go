package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

type OracleRepository struct {
	*baseRepository
}

func NewOracleRepository(databaseURL string) *OracleRepository {
	return &OracleRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *OracleRepository) UpsertStation(ctx context.Context, input store.UpsertStationInput) (*domain.WeatherStation, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO weather_stations (
			provider_name,
			station_id,
			display_name,
			region,
			latitude,
			longitude,
			supported_metrics,
			active,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (provider_name, station_id)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			region = EXCLUDED.region,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			supported_metrics = EXCLUDED.supported_metrics,
			active = EXCLUDED.active,
			updated_at = NOW()
		RETURNING
			id,
			provider_name,
			station_id,
			display_name,
			region,
			latitude,
			longitude,
			supported_metrics,
			active,
			updated_at
	`,
		input.ProviderName,
		input.StationID,
		input.DisplayName,
		input.Region,
		input.Latitude,
		input.Longitude,
		strings.Join(input.SupportedMetrics, ","),
		input.Active,
	)

	station, err := scanWeatherStation(row)
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *OracleRepository) ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error) {
	query := `
		SELECT
			id,
			provider_name,
			station_id,
			display_name,
			region,
			latitude,
			longitude,
			supported_metrics,
			active,
			updated_at
		FROM weather_stations
	`
	args := []any{}
	if activeOnly {
		query += ` WHERE active = TRUE`
	}
	query += ` ORDER BY region ASC, display_name ASC, provider_name ASC, station_id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stations := []domain.WeatherStation{}
	for rows.Next() {
		station, err := scanWeatherStation(rows)
		if err != nil {
			return nil, err
		}
		stations = append(stations, station)
	}
	return stations, rows.Err()
}

func (r *OracleRepository) FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			provider_name,
			station_id,
			display_name,
			region,
			latitude,
			longitude,
			supported_metrics,
			active,
			updated_at
		FROM weather_stations
		WHERE provider_name = $1 AND station_id = $2
	`, strings.TrimSpace(providerName), strings.TrimSpace(stationID))

	station, err := scanWeatherStation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &station, nil
}

func (r *OracleRepository) RecordObservation(ctx context.Context, input store.RecordObservationInput) (*domain.OracleObservation, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO oracle_observations (
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			observed_value,
			normalized_value,
			observed_at,
			recorded_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			observed_value,
			normalized_value,
			observed_at,
			recorded_at
	`,
		input.ContractID,
		input.ProviderName,
		input.StationID,
		input.ObservedMetric,
		input.ObservationWindowStart.UTC(),
		input.ObservationWindowEnd.UTC(),
		input.ObservedValue,
		input.NormalizedValue,
		input.ObservedAt.UTC(),
	)

	observation, err := scanOracleObservation(row)
	if err != nil {
		return nil, err
	}

	return &observation, nil
}

func (r *OracleRepository) ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			observed_value,
			normalized_value,
			observed_at,
			recorded_at
		FROM oracle_observations
		WHERE contract_id = $1
		ORDER BY observed_at DESC, recorded_at DESC, id DESC
		LIMIT $2
	`, contractID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	observations := []domain.OracleObservation{}
	for rows.Next() {
		observation, err := scanOracleObservation(rows)
		if err != nil {
			return nil, err
		}
		observations = append(observations, observation)
	}

	return observations, rows.Err()
}

func (r *OracleRepository) observationsForWindow(ctx context.Context, tx *sql.Tx, contractID int64, windowStart, windowEnd string) ([]domain.OracleObservation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			observed_value,
			normalized_value,
			observed_at,
			recorded_at
		FROM oracle_observations
		WHERE contract_id = $1
		  AND observation_window_start = $2::timestamp
		  AND observation_window_end = $3::timestamp
		ORDER BY observed_at DESC, recorded_at DESC, id DESC
	`, contractID, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	observations := []domain.OracleObservation{}
	for rows.Next() {
		observation, err := scanOracleObservation(rows)
		if err != nil {
			return nil, err
		}
		observations = append(observations, observation)
	}

	return observations, rows.Err()
}

func (r *OracleRepository) GetLatestResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			rule_version,
			resolved_value,
			outcome,
			resolved_at
		FROM contract_resolutions
		WHERE contract_id = $1
		ORDER BY resolved_at DESC, id DESC
		LIMIT 1
	`, contractID)

	resolution, err := scanContractResolution(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &resolution, nil
}

func (r *OracleRepository) ResolveContract(ctx context.Context, input store.ResolveContractInput) (*domain.ContractResolution, error) {
	return nil, errors.New("resolve contract requires oracle service orchestration")
}

func (r *OracleRepository) InsertResolution(ctx context.Context, input domain.ContractResolution) (*domain.ContractResolution, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO contract_resolutions (
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			rule_version,
			resolved_value,
			outcome,
			resolved_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			rule_version,
			resolved_value,
			outcome,
			resolved_at
	`,
		input.ContractID,
		input.ProviderName,
		input.StationID,
		input.ObservedMetric,
		input.ObservationWindowStart.UTC(),
		input.ObservationWindowEnd.UTC(),
		input.RuleVersion,
		input.ResolvedValue,
		string(input.Outcome),
	)

	resolution, err := scanContractResolution(row)
	if err != nil {
		return nil, err
	}

	return &resolution, nil
}

func (r *OracleRepository) ReplaceObservationsProjection(ctx context.Context, contractID int64, observations []domain.OracleObservation) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM oracle_observations WHERE contract_id = $1`, contractID); err != nil {
			return struct{}{}, err
		}

		for _, observation := range observations {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO oracle_observations (
					id,
					contract_id,
					provider_name,
					station_id,
					observed_metric,
					observation_window_start,
					observation_window_end,
					observed_value,
					normalized_value,
					observed_at,
					recorded_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`,
				observation.ID,
				observation.ContractID,
				observation.ProviderName,
				observation.StationID,
				observation.ObservedMetric,
				observation.ObservationWindowStart.UTC(),
				observation.ObservationWindowEnd.UTC(),
				observation.ObservedValue,
				observation.NormalizedValue,
				observation.ObservedAt.UTC(),
				observation.RecordedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *OracleRepository) UpsertResolutionProjection(ctx context.Context, resolution domain.ContractResolution) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contract_resolutions (
			id,
			contract_id,
			provider_name,
			station_id,
			observed_metric,
			observation_window_start,
			observation_window_end,
			rule_version,
			resolved_value,
			outcome,
			resolved_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE
		SET
			contract_id = EXCLUDED.contract_id,
			provider_name = EXCLUDED.provider_name,
			station_id = EXCLUDED.station_id,
			observed_metric = EXCLUDED.observed_metric,
			observation_window_start = EXCLUDED.observation_window_start,
			observation_window_end = EXCLUDED.observation_window_end,
			rule_version = EXCLUDED.rule_version,
			resolved_value = EXCLUDED.resolved_value,
			outcome = EXCLUDED.outcome,
			resolved_at = EXCLUDED.resolved_at
	`,
		resolution.ID,
		resolution.ContractID,
		resolution.ProviderName,
		resolution.StationID,
		resolution.ObservedMetric,
		resolution.ObservationWindowStart.UTC(),
		resolution.ObservationWindowEnd.UTC(),
		resolution.RuleVersion,
		resolution.ResolvedValue,
		string(resolution.Outcome),
		resolution.ResolvedAt.UTC(),
	)
	return err
}

func (r *OracleRepository) ReplaceStationsProjection(ctx context.Context, stations []domain.WeatherStation) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM weather_stations`); err != nil {
			return struct{}{}, err
		}

		for _, station := range stations {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO weather_stations (
					id,
					provider_name,
					station_id,
					display_name,
					region,
					latitude,
					longitude,
					supported_metrics,
					active,
					updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`,
				station.ID,
				station.ProviderName,
				station.StationID,
				station.DisplayName,
				station.Region,
				station.Latitude,
				station.Longitude,
				strings.Join(station.SupportedMetrics, ","),
				station.Active,
				station.UpdatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func parseResolvedFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}
