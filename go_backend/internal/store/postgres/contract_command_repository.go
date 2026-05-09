package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"iwx/go_backend/internal/commands"
)

func (r *ContractRepository) GetContractCommand(ctx context.Context, commandID string) (*commands.ContractCommand, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			command_id,
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
			command_status,
			error_message,
			result_contract_id,
			enqueued_at,
			started_at,
			completed_at,
			updated_at
		FROM contract_commands
		WHERE command_id = $1
	`, commandID)

	command, err := scanContractCommand(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &command, nil
}

func markContractCommandProcessing(ctx context.Context, tx *sql.Tx, envelope commands.CreateContractEnvelope) error {
	tradingStart, _ := parseNullableDate(envelope.Command.TradingPeriodStart)
	tradingEnd, _ := parseNullableDate(envelope.Command.TradingPeriodEnd)
	measurementStart, _ := parseNullableDate(envelope.Command.MeasurementPeriodStart)
	measurementEnd, _ := parseNullableDate(envelope.Command.MeasurementPeriodEnd)

	_, err := tx.ExecContext(ctx, `
		INSERT INTO contract_commands (
			command_id,
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
			command_status,
			enqueued_at,
			started_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 'processing', $18, NOW(), NOW()
		)
		ON CONFLICT (command_id) DO UPDATE
		SET
			command_status = EXCLUDED.command_status,
			started_at = NOW(),
			updated_at = NOW()
	`,
		envelope.CommandID,
		envelope.Command.CreatorUserID,
		strings.TrimSpace(envelope.Command.Name),
		strings.TrimSpace(envelope.Command.Region),
		strings.TrimSpace(envelope.Command.Metric),
		strings.TrimSpace(envelope.Command.Status),
		envelope.Command.Threshold,
		envelope.Command.Multiplier,
		nullableTrimmedString(envelope.Command.MeasurementUnit),
		tradingStart,
		tradingEnd,
		measurementStart,
		measurementEnd,
		nullableTrimmedString(envelope.Command.DataProviderName),
		nullableTrimmedString(envelope.Command.StationID),
		nullableTrimmedString(envelope.Command.DataProviderStationMode),
		nullableTrimmedString(envelope.Command.Description),
		envelope.EnqueuedAt.UTC(),
	)
	return err
}

func markContractCommandSucceeded(ctx context.Context, tx *sql.Tx, commandID string, result commands.CreateContractResult) error {
	var contractID any
	if result.Contract != nil {
		contractID = result.Contract.ID
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE contract_commands
		SET
			command_status = 'succeeded',
			error_message = NULL,
			result_contract_id = $2,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE command_id = $1
	`, commandID, contractID)
	return err
}

func markContractCommandFailed(ctx context.Context, tx *sql.Tx, commandID, errorMessage string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE contract_commands
		SET
			command_status = 'failed',
			error_message = $2,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE command_id = $1
	`, commandID, errorMessage)
	return err
}

func parseNullableDate(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}

	parsed = parsed.UTC()
	return &parsed, nil
}

func nullableTrimmedString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}
