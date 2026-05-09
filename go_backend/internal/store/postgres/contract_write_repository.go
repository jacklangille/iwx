package postgres

import (
	"context"
	"database/sql"
	"strings"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/exchangecore"
)

func (r *ContractRepository) ProcessCreateContract(
	ctx context.Context,
	envelope commands.CreateContractEnvelope,
) (commands.CreateContractResult, error) {
	return withTransaction(ctx, r.db, func(tx *sql.Tx) (commands.CreateContractResult, error) {
		if err := markContractCommandProcessing(ctx, tx, envelope); err != nil {
			return commands.CreateContractResult{}, err
		}

		result, err := executeCreateContract(ctx, tx, envelope.Command)
		if err != nil {
			if updateErr := markContractCommandFailed(ctx, tx, envelope.CommandID, err.Error()); updateErr != nil {
				return commands.CreateContractResult{}, updateErr
			}
			return commands.CreateContractResult{}, err
		}

		if err := markContractCommandSucceeded(ctx, tx, envelope.CommandID, result); err != nil {
			return commands.CreateContractResult{}, err
		}

		return result, nil
	})
}

func executeCreateContract(
	ctx context.Context,
	tx *sql.Tx,
	command commands.CreateContract,
) (commands.CreateContractResult, error) {
	contract, err := insertContract(ctx, tx, command)
	if err != nil {
		return commands.CreateContractResult{}, translateDuplicateContractError(err)
	}
	if err := insertContractRule(ctx, tx, *contract, command); err != nil {
		return commands.CreateContractResult{}, err
	}

	return commands.CreateContractResult{
		Status:   strings.ToLower(strings.TrimSpace(contract.Status)),
		Contract: contract,
	}, nil
}

func insertContract(ctx context.Context, tx *sql.Tx, command commands.CreateContract) (*domain.Contract, error) {
	tradingStart, err := parseNullableDate(command.TradingPeriodStart)
	if err != nil {
		return nil, err
	}
	tradingEnd, err := parseNullableDate(command.TradingPeriodEnd)
	if err != nil {
		return nil, err
	}
	measurementStart, err := parseNullableDate(command.MeasurementPeriodStart)
	if err != nil {
		return nil, err
	}
	measurementEnd, err := parseNullableDate(command.MeasurementPeriodEnd)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO contracts (
			name,
			creator_user_id,
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
		)
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
	`,
		strings.TrimSpace(command.Name),
		command.CreatorUserID,
		strings.TrimSpace(command.Region),
		strings.TrimSpace(command.Metric),
		string(domain.ContractStateDraft),
		command.Threshold,
		command.Multiplier,
		nullableTrimmedString(command.MeasurementUnit),
		tradingStart,
		tradingEnd,
		measurementStart,
		measurementEnd,
		nullableTrimmedString(command.DataProviderName),
		nullableTrimmedString(command.StationID),
		nullableTrimmedString(command.DataProviderStationMode),
		nullableTrimmedString(command.Description),
	)

	contract, err := scanContract(row)
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

func insertContractRule(ctx context.Context, tx *sql.Tx, contract domain.Contract, command commands.CreateContract) error {
	measurementUnit := strings.TrimSpace(command.MeasurementUnit)
	if measurementUnit == "" {
		measurementUnit = contract.MeasurementUnit
	}

	resolutionInclusiveSide := exchangecore.DefaultResolutionInclusiveSide
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO contract_rules (
			contract_id,
			rule_version,
			metric,
			threshold,
			measurement_unit,
			resolution_inclusive_side,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`,
		contract.ID,
		exchangecore.DefaultContractRuleVersion,
		contract.Metric,
		contract.Threshold,
		nullableTrimmedString(measurementUnit),
		resolutionInclusiveSide,
	); err != nil {
		return err
	}

	return nil
}
