package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/store"
)

type ContractRepository struct {
	*baseRepository
}

func NewContractRepository(databaseURL string) *ContractRepository {
	return &ContractRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *ContractRepository) ListContracts(ctx context.Context) ([]domain.Contract, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			updated_at
		FROM contracts
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contracts := []domain.Contract{}
	for rows.Next() {
		contract, err := scanContract(rows)
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, contract)
	}

	return contracts, rows.Err()
}

func (r *ContractRepository) GetContract(ctx context.Context, contractID int64) (*domain.Contract, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			updated_at
		FROM contracts
		WHERE id = $1
	`, contractID)

	contract, err := scanContract(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

func (r *ContractRepository) ContractExists(ctx context.Context, contractID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM contracts WHERE id = $1)`,
		contractID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	return exists, err
}

func (r *ContractRepository) FindDuplicateContract(ctx context.Context, input store.FindDuplicateContractInput) (*domain.Contract, error) {
	tradingStart, err := parseNullableDate(input.TradingPeriodStart)
	if err != nil {
		return nil, err
	}
	tradingEnd, err := parseNullableDate(input.TradingPeriodEnd)
	if err != nil {
		return nil, err
	}
	measurementStart, err := parseNullableDate(input.MeasurementPeriodStart)
	if err != nil {
		return nil, err
	}
	measurementEnd, err := parseNullableDate(input.MeasurementPeriodEnd)
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			updated_at
		FROM contracts
		WHERE LOWER(COALESCE(data_provider_name, '')) = LOWER($1)
		  AND LOWER(COALESCE(station_id, '')) = LOWER($2)
		  AND LOWER(metric) = LOWER($3)
		  AND threshold IS NOT DISTINCT FROM $4
		  AND trading_period_start IS NOT DISTINCT FROM $5
		  AND trading_period_end IS NOT DISTINCT FROM $6
		  AND measurement_period_start IS NOT DISTINCT FROM $7
		  AND measurement_period_end IS NOT DISTINCT FROM $8
		  AND status IN ($9, $10, $11, $12, $13, $14, $15)
		ORDER BY id ASC
		LIMIT 1
	`,
		strings.TrimSpace(input.ProviderName),
		strings.TrimSpace(input.StationID),
		strings.TrimSpace(input.Metric),
		input.Threshold,
		tradingStart,
		tradingEnd,
		measurementStart,
		measurementEnd,
		string(domain.ContractStateDraft),
		string(domain.ContractStatePendingApproval),
		string(domain.ContractStatePendingCollateral),
		string(domain.ContractStateActive),
		string(domain.ContractStateTradingClosed),
		string(domain.ContractStateAwaitingResolution),
		string(domain.ContractStateResolved),
	)

	contract, err := scanContract(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

func (r *ContractRepository) GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			contract_id,
			rule_version,
			metric,
			threshold,
			measurement_unit,
			resolution_inclusive_side,
			created_at
		FROM contract_rules
		WHERE contract_id = $1
	`, contractID)

	var rule domain.ContractRule
	var threshold sql.NullInt64
	var measurementUnit sql.NullString
	var inclusiveSide sql.NullString
	if err := row.Scan(
		&rule.ID,
		&rule.ContractID,
		&rule.RuleVersion,
		&rule.Metric,
		&threshold,
		&measurementUnit,
		&inclusiveSide,
		&rule.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rule.Threshold = nullableInt64(threshold)
	rule.MeasurementUnit = measurementUnit.String
	rule.ResolutionInclusiveSide = domain.ClaimSide(inclusiveSide.String)
	rule.CreatedAt = rule.CreatedAt.UTC()

	return &rule, nil
}

func (r *ContractRepository) UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE contracts
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			updated_at
	`, contractID, status)

	contract, err := scanContract(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

func (r *ContractRepository) UpsertContractProjection(ctx context.Context, contract domain.Contract) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contracts (
			id,
			creator_user_id,
			name,
			region,
			metric,
			status,
			threshold,
			multiplier,
			measurement_unit,
			trading_period_start,
			trading_period_end,
			measurement_period_start,
			measurement_period_end,
			data_provider_name,
			station_id,
			data_provider_station_mode,
			description,
			inserted_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, COALESCE($18, NOW()), COALESCE($18, NOW())
		)
		ON CONFLICT (id) DO UPDATE
		SET
			creator_user_id = EXCLUDED.creator_user_id,
			name = EXCLUDED.name,
			region = EXCLUDED.region,
			metric = EXCLUDED.metric,
			status = EXCLUDED.status,
			threshold = EXCLUDED.threshold,
			multiplier = EXCLUDED.multiplier,
			measurement_unit = EXCLUDED.measurement_unit,
			trading_period_start = EXCLUDED.trading_period_start,
			trading_period_end = EXCLUDED.trading_period_end,
			measurement_period_start = EXCLUDED.measurement_period_start,
			measurement_period_end = EXCLUDED.measurement_period_end,
			data_provider_name = EXCLUDED.data_provider_name,
			station_id = EXCLUDED.station_id,
			data_provider_station_mode = EXCLUDED.data_provider_station_mode,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
	`,
		contract.ID,
		contract.CreatorUserID,
		contract.Name,
		contract.Region,
		contract.Metric,
		contract.Status,
		contract.Threshold,
		contract.Multiplier,
		nullableTrimmedString(contract.MeasurementUnit),
		contract.TradingPeriodStart,
		contract.TradingPeriodEnd,
		contract.MeasurementPeriodStart,
		contract.MeasurementPeriodEnd,
		nullableTrimmedString(contract.DataProviderName),
		nullableTrimmedString(contract.StationID),
		nullableTrimmedString(contract.DataProviderStationMode),
		nullableTrimmedString(contract.Description),
		nullIfZeroTime(contract.UpdatedAt),
	)
	return err
}

func translateDuplicateContractError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "contracts_active_duplicate_market_idx" {
		return &exchangecore.ValidationError{Errors: map[string][]string{
			"contract": {"duplicate market already exists in an active or pending lifecycle state"},
		}}
	}

	return err
}
