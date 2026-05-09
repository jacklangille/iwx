package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

func (r *ContractRepository) ListSettlementEntriesByContract(ctx context.Context, contractID int64, limit int) ([]domain.SettlementEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			user_id,
			entry_type,
			outcome,
			amount_cents,
			quantity,
			reference_id,
			created_at
		FROM settlement_entries
		WHERE contract_id = $1
		ORDER BY id DESC
		LIMIT $2
	`, contractID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []domain.SettlementEntry{}
	for rows.Next() {
		entry, err := scanSettlementEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *ContractRepository) ListSettlementEntriesByUser(ctx context.Context, userID int64, contractID *int64, limit int) ([]domain.SettlementEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
		SELECT
			id,
			contract_id,
			user_id,
			entry_type,
			outcome,
			amount_cents,
			quantity,
			reference_id,
			created_at
		FROM settlement_entries
		WHERE user_id = $1
	`
	args := []any{userID}
	if contractID != nil {
		query += ` AND contract_id = $2`
		args = append(args, *contractID)
		query += ` ORDER BY id DESC LIMIT $3`
		args = append(args, limit)
	} else {
		query += ` ORDER BY id DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []domain.SettlementEntry{}
	for rows.Next() {
		entry, err := scanSettlementEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *ContractRepository) ReplaceSettlementEntriesProjection(ctx context.Context, contractID int64, entries []domain.SettlementEntry) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM settlement_entries WHERE contract_id = $1`, contractID); err != nil {
			return struct{}{}, err
		}

		for _, entry := range entries {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO settlement_entries (
					id,
					contract_id,
					user_id,
					entry_type,
					outcome,
					amount_cents,
					quantity,
					reference_id,
					created_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`,
				entry.ID,
				entry.ContractID,
				entry.UserID,
				string(entry.EntryType),
				string(entry.Outcome),
				entry.AmountCents,
				entry.Quantity,
				entry.ReferenceID,
				entry.CreatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) ReplaceUserSettlementEntriesProjection(ctx context.Context, userID int64, entries []domain.SettlementEntry) error {
	_, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM settlement_entries WHERE user_id = $1`, userID); err != nil {
			return struct{}{}, err
		}

		for _, entry := range entries {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO settlement_entries (
					id,
					contract_id,
					user_id,
					entry_type,
					outcome,
					amount_cents,
					quantity,
					reference_id,
					created_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`,
				entry.ID,
				entry.ContractID,
				entry.UserID,
				string(entry.EntryType),
				string(entry.Outcome),
				entry.AmountCents,
				entry.Quantity,
				entry.ReferenceID,
				entry.CreatedAt.UTC(),
			); err != nil {
				return struct{}{}, err
			}
		}

		return struct{}{}, nil
	})
	return err
}

func (r *ContractRepository) SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
	return withTransaction(ctx, r.db, func(tx *sql.Tx) (*store.SettlementResult, error) {
		eventID := strings.TrimSpace(input.EventID)
		if eventID != "" {
			applied, err := settlementApplicationExistsTx(ctx, tx, eventID)
			if err != nil {
				return nil, err
			}
			if applied {
				contract, err := getContractForUpdateTx(ctx, tx, input.ContractID)
				if err != nil {
					return nil, err
				}
				entries, err := listSettlementEntriesByContractTx(ctx, tx, input.ContractID, 500)
				if err != nil {
					return nil, err
				}
				return &store.SettlementResult{Contract: contract, Entries: entries, AffectedUsers: uniqueSettlementUsers(entries), SettledAt: settlementTimeOrNow(entries)}, nil
			}
		}

		contract, err := getContractForUpdateTx(ctx, tx, input.ContractID)
		if err != nil {
			return nil, err
		}
		if contract == nil {
			return nil, fmt.Errorf("contract not found")
		}
		if contract.Status == string(domain.ContractStateSettled) {
			entries, err := listSettlementEntriesByContractTx(ctx, tx, input.ContractID, 500)
			if err != nil {
				return nil, err
			}
			if eventID != "" {
				if err := recordSettlementApplicationTx(ctx, tx, eventID, input.ContractID, input.CorrelationID, settlementTimeOrNow(entries)); err != nil {
					return nil, err
				}
			}
			return &store.SettlementResult{Contract: contract, Entries: entries, AffectedUsers: uniqueSettlementUsers(entries), SettledAt: settlementTimeOrNow(entries)}, nil
		}

		outcome := domain.ResolutionOutcome(strings.ToLower(strings.TrimSpace(input.Outcome)))
		settlementTime := time.Now().UTC()
		if raw := strings.TrimSpace(input.ResolvedAt); raw != "" {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err == nil {
				settlementTime = parsed.UTC()
			}
		}

		positions, err := listPositionsForContractTx(ctx, tx, input.ContractID)
		if err != nil {
			return nil, err
		}
		if err := releaseContractPositionLocksTx(ctx, tx, input.ContractID, input.CorrelationID, settlementTime); err != nil {
			return nil, err
		}
		if err := releaseContractCashReservationsTx(ctx, tx, input.ContractID, input.CorrelationID, settlementTime); err != nil {
			return nil, err
		}

		contractMultiplier := int64(100)
		if contract.Multiplier != nil && *contract.Multiplier > 0 {
			contractMultiplier = *contract.Multiplier
		}

		locks, err := listContractCollateralLocksForUpdateTx(ctx, tx, input.ContractID)
		if err != nil {
			return nil, err
		}

		result := &store.SettlementResult{
			Contract:      contract,
			Entries:       []domain.SettlementEntry{},
			AffectedUsers: []int64{},
			SettledAt:     settlementTime,
		}

		if outcome == domain.ResolutionOutcomeCancelled {
			for _, lock := range locks {
				if lock.AmountCents <= 0 {
					continue
				}
				if err := releaseCollateralToAvailableTx(ctx, tx, lock, input.CorrelationID, "contract cancelled refund"); err != nil {
					return nil, err
				}
				entry, err := insertSettlementEntryTx(ctx, tx, settlementEntryInput{
					ContractID:  input.ContractID,
					UserID:      lock.UserID,
					EntryType:   domain.SettlementEntryRefund,
					Outcome:     outcome,
					AmountCents: lock.AmountCents,
					Quantity:    0,
					ReferenceID: lock.ReferenceID,
					CreatedAt:   settlementTime,
				})
				if err != nil {
					return nil, err
				}
				result.Entries = append(result.Entries, *entry)
				result.AffectedUsers = append(result.AffectedUsers, lock.UserID)
			}
		} else {
			winningSide := domain.ClaimSideAbove
			if outcome == domain.ResolutionOutcomeBelow {
				winningSide = domain.ClaimSideBelow
			}

			totalCollateral := int64(0)
			for _, lock := range locks {
				totalCollateral += lock.AmountCents
			}

			totalPayout := int64(0)
			for _, position := range positions {
				if position.Side != winningSide || position.TotalQuantity <= 0 {
					continue
				}
				totalPayout += position.TotalQuantity * contractMultiplier
			}
			if totalPayout > totalCollateral {
				return nil, fmt.Errorf("settlement collateral shortfall for contract %d", input.ContractID)
			}

			for _, position := range positions {
				if position.Side != winningSide || position.TotalQuantity <= 0 {
					continue
				}

				payout := position.TotalQuantity * contractMultiplier
				account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, position.UserID, "USD")
				if err != nil {
					return nil, err
				}
				if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+payout, account.LockedCents); err != nil {
					return nil, err
				}
				account, err = getCashAccountByIDTx(ctx, tx, account.ID)
				if err != nil {
					return nil, err
				}
				if _, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
					AccountID:     account.ID,
					UserID:        account.UserID,
					EntryType:     domain.LedgerEntrySettlementCredit,
					AmountCents:   payout,
					ReferenceType: "contract_settlement",
					ReferenceID:   input.CorrelationID,
					CorrelationID: input.CorrelationID,
					Description:   "winning contract payout",
				}); err != nil {
					return nil, err
				}

				entry, err := insertSettlementEntryTx(ctx, tx, settlementEntryInput{
					ContractID:  input.ContractID,
					UserID:      position.UserID,
					EntryType:   domain.SettlementEntryPayout,
					Outcome:     outcome,
					AmountCents: payout,
					Quantity:    position.TotalQuantity,
					ReferenceID: input.CorrelationID,
					CreatedAt:   settlementTime,
				})
				if err != nil {
					return nil, err
				}
				result.Entries = append(result.Entries, *entry)
				result.AffectedUsers = append(result.AffectedUsers, position.UserID)
			}

			payoutRemaining := totalPayout
			for _, lock := range locks {
				if payoutRemaining <= 0 {
					break
				}
				if lock.AmountCents <= 0 {
					continue
				}

				deduction := lock.AmountCents
				if deduction > payoutRemaining {
					deduction = payoutRemaining
				}
				if deduction > 0 {
					if err := consumeCollateralForSettlementTx(ctx, tx, lock, deduction, input.CorrelationID); err != nil {
						return nil, err
					}
					payoutRemaining -= deduction
					result.AffectedUsers = append(result.AffectedUsers, lock.UserID)
				}

				leftover := lock.AmountCents - deduction
				if leftover > 0 {
					remainingLock := lock
					remainingLock.AmountCents = leftover
					if err := releaseCollateralToAvailableTx(ctx, tx, remainingLock, input.CorrelationID, "unused settlement collateral release"); err != nil {
						return nil, err
					}
					entry, err := insertSettlementEntryTx(ctx, tx, settlementEntryInput{
						ContractID:  input.ContractID,
						UserID:      lock.UserID,
						EntryType:   domain.SettlementEntryCollateralRelease,
						Outcome:     outcome,
						AmountCents: leftover,
						Quantity:    0,
						ReferenceID: lock.ReferenceID,
						CreatedAt:   settlementTime,
					})
					if err != nil {
						return nil, err
					}
					result.Entries = append(result.Entries, *entry)
				}
			}
		}

		if err := zeroContractPositionsTx(ctx, tx, input.ContractID); err != nil {
			return nil, err
		}

		updatedContract, err := updateContractStatusTx(ctx, tx, input.ContractID, string(domain.ContractStateSettled))
		if err != nil {
			return nil, err
		}
		if eventID != "" {
			if err := recordSettlementApplicationTx(ctx, tx, eventID, input.ContractID, input.CorrelationID, settlementTime); err != nil {
				return nil, err
			}
		}
		result.Contract = updatedContract
		result.AffectedUsers = uniqueInt64s(result.AffectedUsers)

		return result, nil
	})
}

func settlementApplicationExistsTx(ctx context.Context, tx *sql.Tx, eventID string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM settlement_applications WHERE resolution_event_id = $1)
	`, eventID).Scan(&exists)
	return exists, err
}

func recordSettlementApplicationTx(ctx context.Context, tx *sql.Tx, eventID string, contractID int64, correlationID string, settledAt time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO settlement_applications (
			resolution_event_id,
			contract_id,
			correlation_id,
			settled_at
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (resolution_event_id) DO NOTHING
	`, eventID, contractID, nullableTrimmedString(correlationID), settledAt.UTC())
	return err
}

type settlementEntryInput struct {
	ContractID  int64
	UserID      int64
	EntryType   domain.SettlementEntryType
	Outcome     domain.ResolutionOutcome
	AmountCents int64
	Quantity    int64
	ReferenceID string
	CreatedAt   time.Time
}

func insertSettlementEntryTx(ctx context.Context, tx *sql.Tx, input settlementEntryInput) (*domain.SettlementEntry, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO settlement_entries (
			contract_id,
			user_id,
			entry_type,
			outcome,
			amount_cents,
			quantity,
			reference_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id,
			contract_id,
			user_id,
			entry_type,
			outcome,
			amount_cents,
			quantity,
			reference_id,
			created_at
	`,
		input.ContractID,
		input.UserID,
		string(input.EntryType),
		string(input.Outcome),
		input.AmountCents,
		input.Quantity,
		input.ReferenceID,
		input.CreatedAt.UTC(),
	)

	entry, err := scanSettlementEntry(row)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func listSettlementEntriesByContractTx(ctx context.Context, tx *sql.Tx, contractID int64, limit int) ([]domain.SettlementEntry, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			contract_id,
			user_id,
			entry_type,
			outcome,
			amount_cents,
			quantity,
			reference_id,
			created_at
		FROM settlement_entries
		WHERE contract_id = $1
		ORDER BY id DESC
		LIMIT $2
	`, contractID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []domain.SettlementEntry{}
	for rows.Next() {
		entry, err := scanSettlementEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func listPositionsForContractTx(ctx context.Context, tx *sql.Tx, contractID int64) ([]domain.Position, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			contract_id,
			side,
			available_quantity,
			locked_quantity,
			total_quantity,
			updated_at
		FROM positions
		WHERE contract_id = $1
		ORDER BY user_id ASC, side ASC
		FOR UPDATE
	`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []domain.Position{}
	for rows.Next() {
		position, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}
	return positions, rows.Err()
}

func releaseContractPositionLocksTx(ctx context.Context, tx *sql.Tx, contractID int64, correlationID string, releasedAt time.Time) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			contract_id,
			side,
			quantity,
			status,
			order_id,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at,
			released_at
		FROM position_locks
		WHERE contract_id = $1 AND status = 'active'
		ORDER BY id ASC
		FOR UPDATE
	`, contractID)
	if err != nil {
		return err
	}
	defer rows.Close()

	locks := []domain.PositionLock{}
	for rows.Next() {
		lock, err := scanPositionLock(rows)
		if err != nil {
			return err
		}
		locks = append(locks, lock)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, lock := range locks {
		position, err := getPositionForUpdateTx(ctx, tx, lock.UserID, lock.ContractID, string(lock.Side))
		if err != nil {
			return err
		}
		if position != nil {
			if err := updatePositionBalancesTx(ctx, tx, position.ID, position.AvailableQuantity+lock.Quantity, position.LockedQuantity-lock.Quantity); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE position_locks
			SET status = 'released', updated_at = $2, released_at = $2, correlation_id = $3
			WHERE id = $1
		`, lock.ID, releasedAt.UTC(), correlationID); err != nil {
			return err
		}
	}

	return nil
}

func releaseContractCashReservationsTx(ctx context.Context, tx *sql.Tx, contractID int64, correlationID string, releasedAt time.Time) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			contract_id,
			currency,
			amount_cents,
			status,
			reference_type,
			reference_id,
			correlation_id,
			description,
			created_at,
			updated_at,
			released_at
		FROM order_cash_reservations
		WHERE contract_id = $1 AND status = 'active'
		ORDER BY id ASC
		FOR UPDATE
	`, contractID)
	if err != nil {
		return err
	}
	defer rows.Close()

	reservations := []domain.OrderCashReservation{}
	for rows.Next() {
		reservation, err := scanOrderCashReservation(rows)
		if err != nil {
			return err
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, reservation := range reservations {
		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, reservation.UserID, reservation.Currency)
		if err != nil {
			return err
		}
		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+reservation.AmountCents, account.LockedCents-reservation.AmountCents); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE order_cash_reservations
			SET status = 'released', updated_at = $2, released_at = $2, correlation_id = $3
			WHERE id = $1
		`, reservation.ID, releasedAt.UTC(), correlationID); err != nil {
			return err
		}
		if _, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        account.UserID,
			EntryType:     domain.LedgerEntryOrderCashRelease,
			AmountCents:   reservation.AmountCents,
			ReferenceType: reservation.ReferenceType,
			ReferenceID:   reservation.ReferenceID,
			CorrelationID: correlationID,
			Description:   "order reservation released during settlement",
		}); err != nil {
			return err
		}
	}

	return nil
}

func listContractCollateralLocksForUpdateTx(ctx context.Context, tx *sql.Tx, contractID int64) ([]domain.CollateralLock, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			contract_id,
			currency,
			amount_cents,
			status,
			reference_id,
			description,
			reference_issuance_id,
			created_at,
			updated_at,
			released_at
		FROM collateral_locks
		WHERE contract_id = $1 AND status IN ('locked', 'consumed')
		ORDER BY id ASC
		FOR UPDATE
	`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locks := []domain.CollateralLock{}
	for rows.Next() {
		lock, err := scanCollateralLock(rows)
		if err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}
	return locks, rows.Err()
}

func releaseCollateralToAvailableTx(ctx context.Context, tx *sql.Tx, lock domain.CollateralLock, correlationID, description string) error {
	account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, lock.UserID, lock.Currency)
	if err != nil {
		return err
	}
	if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+lock.AmountCents, account.LockedCents-lock.AmountCents); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE collateral_locks
		SET status = 'released', updated_at = NOW(), released_at = NOW()
		WHERE id = $1
	`, lock.ID); err != nil {
		return err
	}
	_, err = insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
		AccountID:     account.ID,
		UserID:        lock.UserID,
		EntryType:     domain.LedgerEntryCollateralRelease,
		AmountCents:   lock.AmountCents,
		ReferenceType: "collateral_lock",
		ReferenceID:   fmt.Sprintf("%d", lock.ID),
		CorrelationID: correlationID,
		Description:   description,
	})
	return err
}

func consumeCollateralForSettlementTx(ctx context.Context, tx *sql.Tx, lock domain.CollateralLock, deduction int64, correlationID string) error {
	account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, lock.UserID, lock.Currency)
	if err != nil {
		return err
	}
	if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents, account.LockedCents-deduction); err != nil {
		return err
	}
	_, err = insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
		AccountID:     account.ID,
		UserID:        lock.UserID,
		EntryType:     domain.LedgerEntrySettlementDebit,
		AmountCents:   -deduction,
		ReferenceType: "contract_settlement",
		ReferenceID:   correlationID,
		CorrelationID: correlationID,
		Description:   "collateral consumed for settlement payout",
	})
	return err
}

func zeroContractPositionsTx(ctx context.Context, tx *sql.Tx, contractID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE positions
		SET available_quantity = 0, locked_quantity = 0, total_quantity = 0, updated_at = NOW()
		WHERE contract_id = $1
	`, contractID)
	return err
}

func updateContractStatusTx(ctx context.Context, tx *sql.Tx, contractID int64, status string) (*domain.Contract, error) {
	row := tx.QueryRowContext(ctx, `
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
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func uniqueSettlementUsers(entries []domain.SettlementEntry) []int64 {
	userIDs := make([]int64, 0, len(entries))
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}
	return uniqueInt64s(userIDs)
}

func uniqueInt64s(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func settlementTimeOrNow(entries []domain.SettlementEntry) time.Time {
	if len(entries) > 0 {
		return entries[0].CreatedAt.UTC()
	}
	return time.Now().UTC()
}
