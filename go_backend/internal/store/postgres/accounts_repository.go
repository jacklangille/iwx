package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

var (
	errInsufficientAvailableBalance = errors.New("insufficient available balance")
	errCollateralLockNotFound       = errors.New("collateral lock not found")
	errCollateralLockInactive       = errors.New("collateral lock is not active")
	errCashReservationNotFound      = errors.New("cash reservation not found")
	errCashReservationInactive      = errors.New("cash reservation is not active")
)

func (r *ContractRepository) GetCashAccount(ctx context.Context, userID int64, currency string) (*domain.CashAccount, error) {
	return withTransaction(ctx, r.db, func(tx *sql.Tx) (*domain.CashAccount, error) {
		account, err := getOrCreateCashAccountTx(ctx, tx, userID, currency)
		if err != nil {
			return nil, err
		}

		return account, validateCashAccount(*account)
	})
}

func (r *ContractRepository) ListLedgerEntries(ctx context.Context, userID int64, currency string, limit int) ([]domain.LedgerEntry, error) {
	if limit <= 0 || limit > 250 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			le.id,
			le.account_id,
			le.user_id,
			le.entry_type,
			le.amount_cents,
			le.reference_type,
			le.reference_id,
			le.correlation_id,
			le.description,
			le.occurred_at
		FROM ledger_entries le
		INNER JOIN cash_accounts ca ON ca.id = le.account_id
		WHERE le.user_id = $1 AND ca.currency = $2
		ORDER BY le.id DESC
		LIMIT $3
	`, userID, strings.ToUpper(strings.TrimSpace(currency)), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []domain.LedgerEntry{}
	for rows.Next() {
		entry, err := scanLedgerEntry(rows)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *ContractRepository) DepositCash(ctx context.Context, input store.DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		account *domain.CashAccount
		entry   *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, input.UserID, input.Currency)
		if err != nil {
			return result{}, err
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+input.AmountCents, account.LockedCents); err != nil {
			return result{}, err
		}

		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        input.UserID,
			EntryType:     domain.LedgerEntryDeposit,
			AmountCents:   input.AmountCents,
			ReferenceType: "cash_account",
			ReferenceID:   defaultReferenceID(input.ReferenceID, input.CorrelationID),
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, err
	}

	return outcome.account, outcome.entry, nil
}

func (r *ContractRepository) WithdrawCash(ctx context.Context, input store.WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		account *domain.CashAccount
		entry   *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, input.UserID, input.Currency)
		if err != nil {
			return result{}, err
		}
		if account.AvailableCents < input.AmountCents {
			return result{}, errInsufficientAvailableBalance
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents-input.AmountCents, account.LockedCents); err != nil {
			return result{}, err
		}

		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        input.UserID,
			EntryType:     domain.LedgerEntryWithdrawal,
			AmountCents:   -input.AmountCents,
			ReferenceType: "cash_account",
			ReferenceID:   defaultReferenceID(input.ReferenceID, input.CorrelationID),
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, err
	}

	return outcome.account, outcome.entry, nil
}

func (r *ContractRepository) ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error) {
	rows, err := r.db.QueryContext(ctx, `
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
		WHERE user_id = $1 AND currency = $2
		ORDER BY id DESC
	`, userID, strings.ToUpper(strings.TrimSpace(currency)))
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

func (r *ContractRepository) ListContractCollateralLocks(ctx context.Context, userID, contractID int64) ([]domain.CollateralLock, error) {
	rows, err := r.db.QueryContext(ctx, `
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
		WHERE user_id = $1 AND contract_id = $2
		ORDER BY id DESC
	`, userID, contractID)
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

func (r *ContractRepository) CreateCollateralLock(ctx context.Context, input store.CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		lock    *domain.CollateralLock
		account *domain.CashAccount
		entry   *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, input.UserID, input.Currency)
		if err != nil {
			return result{}, err
		}
		if account.AvailableCents < input.AmountCents {
			return result{}, errInsufficientAvailableBalance
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents-input.AmountCents, account.LockedCents+input.AmountCents); err != nil {
			return result{}, err
		}

		lock, err := insertCollateralLockTx(ctx, tx, input)
		if err != nil {
			return result{}, err
		}

		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        input.UserID,
			EntryType:     domain.LedgerEntryCollateralLock,
			AmountCents:   input.AmountCents,
			ReferenceType: "collateral_lock",
			ReferenceID:   fmt.Sprintf("%d", lock.ID),
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{lock: lock, account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return outcome.lock, outcome.account, outcome.entry, nil
}

func (r *ContractRepository) ReleaseCollateralLock(ctx context.Context, input store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		lock    *domain.CollateralLock
		account *domain.CashAccount
		entry   *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		lock, err := getCollateralLockForUpdateTx(ctx, tx, input.LockID)
		if err != nil {
			return result{}, err
		}
		if lock == nil {
			return result{}, errCollateralLockNotFound
		}
		if lock.UserID != input.UserID {
			return result{}, errCollateralLockNotFound
		}
		if lock.Status != domain.CollateralLockStatusLocked {
			return result{}, errCollateralLockInactive
		}

		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, lock.UserID, lock.Currency)
		if err != nil {
			return result{}, err
		}
		if account.LockedCents < lock.AmountCents {
			return result{}, fmt.Errorf("cash account locked balance underflow for collateral lock %d", lock.ID)
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+lock.AmountCents, account.LockedCents-lock.AmountCents); err != nil {
			return result{}, err
		}
		if err := releaseCollateralLockTx(ctx, tx, lock.ID); err != nil {
			return result{}, err
		}

		lock, err = getCollateralLockForUpdateTx(ctx, tx, lock.ID)
		if err != nil {
			return result{}, err
		}
		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        account.UserID,
			EntryType:     domain.LedgerEntryCollateralRelease,
			AmountCents:   lock.AmountCents,
			ReferenceType: "collateral_lock",
			ReferenceID:   fmt.Sprintf("%d", lock.ID),
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{lock: lock, account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return outcome.lock, outcome.account, outcome.entry, nil
}

func (r *ContractRepository) ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error) {
	rows, err := r.db.QueryContext(ctx, `
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
		WHERE user_id = $1 AND currency = $2
		ORDER BY id DESC
	`, userID, strings.ToUpper(strings.TrimSpace(currency)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reservations := []domain.OrderCashReservation{}
	for rows.Next() {
		reservation, err := scanOrderCashReservation(rows)
		if err != nil {
			return nil, err
		}

		reservations = append(reservations, reservation)
	}

	return reservations, rows.Err()
}

func (r *ContractRepository) CreateOrderCashReservation(ctx context.Context, input store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		reservation *domain.OrderCashReservation
		account     *domain.CashAccount
		entry       *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, input.UserID, input.Currency)
		if err != nil {
			return result{}, err
		}
		if account.AvailableCents < input.AmountCents {
			return result{}, errInsufficientAvailableBalance
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents-input.AmountCents, account.LockedCents+input.AmountCents); err != nil {
			return result{}, err
		}

		reservation, err := insertOrderCashReservationTx(ctx, tx, input)
		if err != nil {
			return result{}, err
		}

		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        account.UserID,
			EntryType:     domain.LedgerEntryOrderCashReserve,
			AmountCents:   input.AmountCents,
			ReferenceType: strings.TrimSpace(input.ReferenceType),
			ReferenceID:   strings.TrimSpace(input.ReferenceID),
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{reservation: reservation, account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return outcome.reservation, outcome.account, outcome.entry, nil
}

func (r *ContractRepository) ReleaseOrderCashReservation(ctx context.Context, input store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	type result struct {
		reservation *domain.OrderCashReservation
		account     *domain.CashAccount
		entry       *domain.LedgerEntry
	}

	outcome, err := withTransaction(ctx, r.db, func(tx *sql.Tx) (result, error) {
		reservation, err := getOrderCashReservationForUpdateTx(ctx, tx, input.ReservationID)
		if err != nil {
			return result{}, err
		}
		if reservation == nil {
			return result{}, errCashReservationNotFound
		}
		if reservation.UserID != input.UserID {
			return result{}, errCashReservationNotFound
		}
		if reservation.Status != domain.CashReservationStatusActive {
			return result{}, errCashReservationInactive
		}

		account, err := getOrCreateCashAccountForUpdateTx(ctx, tx, reservation.UserID, reservation.Currency)
		if err != nil {
			return result{}, err
		}
		if account.LockedCents < reservation.AmountCents {
			return result{}, fmt.Errorf("cash account locked balance underflow for reservation %d", reservation.ID)
		}

		if err := updateCashAccountBalancesTx(ctx, tx, account.ID, account.AvailableCents+reservation.AmountCents, account.LockedCents-reservation.AmountCents); err != nil {
			return result{}, err
		}
		if err := releaseOrderCashReservationTx(ctx, tx, reservation.ID); err != nil {
			return result{}, err
		}

		reservation, err = getOrderCashReservationForUpdateTx(ctx, tx, reservation.ID)
		if err != nil {
			return result{}, err
		}
		account, err = getCashAccountByIDTx(ctx, tx, account.ID)
		if err != nil {
			return result{}, err
		}

		entry, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     account.ID,
			UserID:        account.UserID,
			EntryType:     domain.LedgerEntryOrderCashRelease,
			AmountCents:   reservation.AmountCents,
			ReferenceType: reservation.ReferenceType,
			ReferenceID:   reservation.ReferenceID,
			CorrelationID: strings.TrimSpace(input.CorrelationID),
			Description:   strings.TrimSpace(input.Description),
		})
		if err != nil {
			return result{}, err
		}

		return result{reservation: reservation, account: account, entry: entry}, validateCashAccount(*account)
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return outcome.reservation, outcome.account, outcome.entry, nil
}

type ledgerEntryInput struct {
	AccountID     int64
	UserID        int64
	EntryType     domain.LedgerEntryType
	AmountCents   int64
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
}

func insertLedgerEntryTx(ctx context.Context, tx *sql.Tx, input ledgerEntryInput) (*domain.LedgerEntry, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO ledger_entries (
			account_id,
			user_id,
			entry_type,
			amount_cents,
			reference_type,
			reference_id,
			correlation_id,
			description,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING
			id,
			account_id,
			user_id,
			entry_type,
			amount_cents,
			reference_type,
			reference_id,
			correlation_id,
			description,
			occurred_at
	`,
		input.AccountID,
		input.UserID,
		input.EntryType,
		input.AmountCents,
		defaultReferenceType(input.ReferenceType),
		defaultReferenceID(input.ReferenceID, input.CorrelationID),
		strings.TrimSpace(input.CorrelationID),
		strings.TrimSpace(input.Description),
	)

	entry, err := scanLedgerEntry(row)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func insertCollateralLockTx(ctx context.Context, tx *sql.Tx, input store.CreateCollateralLockInput) (*domain.CollateralLock, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO collateral_locks (
			user_id,
			contract_id,
			currency,
			amount_cents,
			status,
			reference_id,
			description,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'locked', $5, $6, NOW(), NOW())
		RETURNING
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
	`, input.UserID, input.ContractID, input.Currency, input.AmountCents, strings.TrimSpace(input.ReferenceID), strings.TrimSpace(input.Description))

	lock, err := scanCollateralLock(row)
	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func releaseCollateralLockTx(ctx context.Context, tx *sql.Tx, lockID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE collateral_locks
		SET status = 'released', updated_at = NOW(), released_at = NOW()
		WHERE id = $1
	`, lockID)
	return err
}

func getCollateralLockForUpdateTx(ctx context.Context, tx *sql.Tx, lockID int64) (*domain.CollateralLock, error) {
	row := tx.QueryRowContext(ctx, `
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
		WHERE id = $1
		FOR UPDATE
	`, lockID)

	lock, err := scanCollateralLock(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func insertOrderCashReservationTx(ctx context.Context, tx *sql.Tx, input store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO order_cash_reservations (
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
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $8, NOW(), NOW())
		RETURNING
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
	`,
		input.UserID,
		input.ContractID,
		input.Currency,
		input.AmountCents,
		strings.TrimSpace(input.ReferenceType),
		strings.TrimSpace(input.ReferenceID),
		strings.TrimSpace(input.CorrelationID),
		strings.TrimSpace(input.Description),
	)

	reservation, err := scanOrderCashReservation(row)
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

func releaseOrderCashReservationTx(ctx context.Context, tx *sql.Tx, reservationID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE order_cash_reservations
		SET status = 'released', updated_at = NOW(), released_at = NOW()
		WHERE id = $1
	`, reservationID)
	return err
}

func getOrderCashReservationForUpdateTx(ctx context.Context, tx *sql.Tx, reservationID int64) (*domain.OrderCashReservation, error) {
	row := tx.QueryRowContext(ctx, `
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
		WHERE id = $1
		FOR UPDATE
	`, reservationID)

	reservation, err := scanOrderCashReservation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

func getOrCreateCashAccountTx(ctx context.Context, tx *sql.Tx, userID int64, currency string) (*domain.CashAccount, error) {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO cash_accounts (
			user_id,
			currency,
			available_cents,
			locked_cents,
			total_cents,
			updated_at
		)
		VALUES ($1, $2, 0, 0, 0, NOW())
		ON CONFLICT (user_id, currency) DO NOTHING
	`, userID, strings.ToUpper(strings.TrimSpace(currency))); err != nil {
		return nil, err
	}

	return getCashAccountTx(ctx, tx, userID, currency, false)
}

func getOrCreateCashAccountForUpdateTx(ctx context.Context, tx *sql.Tx, userID int64, currency string) (*domain.CashAccount, error) {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO cash_accounts (
			user_id,
			currency,
			available_cents,
			locked_cents,
			total_cents,
			updated_at
		)
		VALUES ($1, $2, 0, 0, 0, NOW())
		ON CONFLICT (user_id, currency) DO NOTHING
	`, userID, strings.ToUpper(strings.TrimSpace(currency))); err != nil {
		return nil, err
	}

	return getCashAccountTx(ctx, tx, userID, currency, true)
}

func getCashAccountTx(ctx context.Context, tx *sql.Tx, userID int64, currency string, forUpdate bool) (*domain.CashAccount, error) {
	query := `
		SELECT
			id,
			user_id,
			currency,
			available_cents,
			locked_cents,
			total_cents,
			updated_at
		FROM cash_accounts
		WHERE user_id = $1 AND currency = $2
	`
	if forUpdate {
		query += ` FOR UPDATE`
	}

	row := tx.QueryRowContext(ctx, query, userID, strings.ToUpper(strings.TrimSpace(currency)))
	account, err := scanCashAccount(row)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func getCashAccountByIDTx(ctx context.Context, tx *sql.Tx, accountID int64) (*domain.CashAccount, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			currency,
			available_cents,
			locked_cents,
			total_cents,
			updated_at
		FROM cash_accounts
		WHERE id = $1
	`, accountID)

	account, err := scanCashAccount(row)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func updateCashAccountBalancesTx(ctx context.Context, tx *sql.Tx, accountID, availableCents, lockedCents int64) error {
	totalCents := availableCents + lockedCents
	_, err := tx.ExecContext(ctx, `
		UPDATE cash_accounts
		SET
			available_cents = $2,
			locked_cents = $3,
			total_cents = $4,
			updated_at = NOW()
		WHERE id = $1
	`, accountID, availableCents, lockedCents, totalCents)
	return err
}

func validateCashAccount(account domain.CashAccount) error {
	if account.AvailableCents < 0 || account.LockedCents < 0 || account.TotalCents < 0 {
		return fmt.Errorf("cash account %d has negative balances", account.ID)
	}
	if account.AvailableCents+account.LockedCents != account.TotalCents {
		return fmt.Errorf("cash account %d violates balance invariant", account.ID)
	}

	return nil
}

func defaultReferenceType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "cash_account"
	}

	return trimmed
}

func defaultReferenceID(value, correlationID string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	if trimmed := strings.TrimSpace(correlationID); trimmed != "" {
		return trimmed
	}

	return "manual"
}
