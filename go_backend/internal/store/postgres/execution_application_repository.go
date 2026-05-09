package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/money"
	"iwx/go_backend/internal/store"
)

func (r *ContractRepository) ApplyExecution(ctx context.Context, event events.ExecutionCreated) (*store.ExecutionApplicationResult, error) {
	return withTransaction(ctx, r.db, func(tx *sql.Tx) (*store.ExecutionApplicationResult, error) {
		if strings.TrimSpace(event.ExecutionID) == "" {
			return nil, fmt.Errorf("execution_id is required")
		}

		exists, err := executionAlreadyAppliedTx(ctx, tx, event.ExecutionID)
		if err != nil {
			return nil, err
		}
		if exists {
			return &store.ExecutionApplicationResult{
				ExecutionID:   event.ExecutionID,
				ContractID:    event.ContractID,
				BuyerUserID:   event.BuyerUserID,
				SellerUserID:  event.SellerUserID,
				AffectedUsers: uniqueInt64s([]int64{event.BuyerUserID, event.SellerUserID}),
				Applied:       false,
				AppliedAt:     time.Now().UTC(),
			}, nil
		}

		if event.BuyerCashReservationID == nil || *event.BuyerCashReservationID <= 0 {
			return nil, fmt.Errorf("buyer_cash_reservation_id is required")
		}
		if event.SellerPositionLockID == nil || *event.SellerPositionLockID <= 0 {
			return nil, fmt.Errorf("seller_position_lock_id is required")
		}
		if strings.TrimSpace(event.TokenType) == "" {
			return nil, fmt.Errorf("token_type is required")
		}
		if event.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than 0")
		}

		fillAmountCents, err := money.ParseCents(event.Price)
		if err != nil {
			return nil, err
		}
		fillAmountCents *= event.Quantity

		buyerReservation, err := getOrderCashReservationForUpdateTx(ctx, tx, *event.BuyerCashReservationID)
		if err != nil {
			return nil, err
		}
		if buyerReservation == nil || buyerReservation.UserID != event.BuyerUserID || buyerReservation.ContractID != event.ContractID {
			return nil, fmt.Errorf("buyer reservation not found for execution %s", event.ExecutionID)
		}
		if buyerReservation.Status != domain.CashReservationStatusActive {
			return nil, fmt.Errorf("buyer reservation is not active for execution %s", event.ExecutionID)
		}
		if buyerReservation.AmountCents < fillAmountCents {
			return nil, fmt.Errorf("buyer reservation underfunded for execution %s", event.ExecutionID)
		}

		buyerAccount, err := getOrCreateCashAccountForUpdateTx(ctx, tx, event.BuyerUserID, buyerReservation.Currency)
		if err != nil {
			return nil, err
		}
		if buyerAccount.LockedCents < fillAmountCents {
			return nil, fmt.Errorf("buyer locked balance underflow for execution %s", event.ExecutionID)
		}
		if err := updateCashAccountBalancesTx(ctx, tx, buyerAccount.ID, buyerAccount.AvailableCents, buyerAccount.LockedCents-fillAmountCents); err != nil {
			return nil, err
		}
		if err := consumeOrderCashReservationTx(ctx, tx, buyerReservation.ID, buyerReservation.AmountCents-fillAmountCents); err != nil {
			return nil, err
		}
		if _, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     buyerAccount.ID,
			UserID:        event.BuyerUserID,
			EntryType:     domain.LedgerEntryTradeDebit,
			AmountCents:   -fillAmountCents,
			ReferenceType: "execution",
			ReferenceID:   event.ExecutionID,
			CorrelationID: event.ExecutionID,
			Description:   "execution debit",
		}); err != nil {
			return nil, err
		}

		sellerLock, err := getPositionLockForUpdateTx(ctx, tx, *event.SellerPositionLockID)
		if err != nil {
			return nil, err
		}
		if sellerLock == nil || sellerLock.UserID != event.SellerUserID || sellerLock.ContractID != event.ContractID {
			return nil, fmt.Errorf("seller position lock not found for execution %s", event.ExecutionID)
		}
		if sellerLock.Status != domain.PositionLockStatusActive {
			return nil, fmt.Errorf("seller position lock is not active for execution %s", event.ExecutionID)
		}
		if sellerLock.Quantity < event.Quantity {
			return nil, fmt.Errorf("seller position lock underfunded for execution %s", event.ExecutionID)
		}

		sellerPosition, err := getPositionForUpdateTx(ctx, tx, event.SellerUserID, event.ContractID, event.TokenType)
		if err != nil {
			return nil, err
		}
		if sellerPosition == nil || sellerPosition.LockedQuantity < event.Quantity {
			return nil, fmt.Errorf("seller locked position underflow for execution %s", event.ExecutionID)
		}
		if err := updatePositionBalancesTx(ctx, tx, sellerPosition.ID, sellerPosition.AvailableQuantity, sellerPosition.LockedQuantity-event.Quantity); err != nil {
			return nil, err
		}
		if err := consumePositionLockTx(ctx, tx, sellerLock.ID, sellerLock.Quantity-event.Quantity); err != nil {
			return nil, err
		}

		buyerPosition, err := upsertPositionQuantityTx(ctx, tx, event.BuyerUserID, event.ContractID, domain.ClaimSide(strings.ToLower(strings.TrimSpace(event.TokenType))), event.Quantity)
		if err != nil {
			return nil, err
		}
		if err := validatePosition(*buyerPosition); err != nil {
			return nil, err
		}

		sellerAccount, err := getOrCreateCashAccountForUpdateTx(ctx, tx, event.SellerUserID, buyerReservation.Currency)
		if err != nil {
			return nil, err
		}
		if err := updateCashAccountBalancesTx(ctx, tx, sellerAccount.ID, sellerAccount.AvailableCents+fillAmountCents, sellerAccount.LockedCents); err != nil {
			return nil, err
		}
		if _, err := insertLedgerEntryTx(ctx, tx, ledgerEntryInput{
			AccountID:     sellerAccount.ID,
			UserID:        event.SellerUserID,
			EntryType:     domain.LedgerEntryTradeCredit,
			AmountCents:   fillAmountCents,
			ReferenceType: "execution",
			ReferenceID:   event.ExecutionID,
			CorrelationID: event.ExecutionID,
			Description:   "execution credit",
		}); err != nil {
			return nil, err
		}

		appliedAt := event.OccurredAt.UTC()
		if appliedAt.IsZero() {
			appliedAt = time.Now().UTC()
		}
		if err := recordExecutionApplicationTx(ctx, tx, event, appliedAt); err != nil {
			return nil, err
		}

		return &store.ExecutionApplicationResult{
			ExecutionID:   event.ExecutionID,
			ContractID:    event.ContractID,
			BuyerUserID:   event.BuyerUserID,
			SellerUserID:  event.SellerUserID,
			AffectedUsers: uniqueInt64s([]int64{event.BuyerUserID, event.SellerUserID}),
			Applied:       true,
			AppliedAt:     appliedAt,
		}, nil
	})
}

func executionAlreadyAppliedTx(ctx context.Context, tx *sql.Tx, executionID string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM execution_applications WHERE execution_id = $1)`, executionID).Scan(&exists)
	return exists, err
}

func recordExecutionApplicationTx(ctx context.Context, tx *sql.Tx, event events.ExecutionCreated, appliedAt time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO execution_applications (
			execution_id,
			contract_id,
			buyer_user_id,
			seller_user_id,
			occurred_at,
			applied_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.ExecutionID, event.ContractID, event.BuyerUserID, event.SellerUserID, event.OccurredAt.UTC(), appliedAt.UTC())
	return err
}

func consumeOrderCashReservationTx(ctx context.Context, tx *sql.Tx, reservationID, remainingAmount int64) error {
	status := "active"
	var releasedAt any
	if remainingAmount == 0 {
		status = "consumed"
		releasedAt = time.Now().UTC()
	} else if remainingAmount < 0 {
		return errors.New("negative remaining reservation amount")
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE order_cash_reservations
		SET amount_cents = $2, status = $3, updated_at = NOW(), released_at = $4
		WHERE id = $1
	`, reservationID, remainingAmount, status, releasedAt)
	return err
}

func consumePositionLockTx(ctx context.Context, tx *sql.Tx, lockID, remainingQuantity int64) error {
	status := "active"
	var releasedAt any
	if remainingQuantity == 0 {
		status = "consumed"
		releasedAt = time.Now().UTC()
	} else if remainingQuantity < 0 {
		return errors.New("negative remaining lock quantity")
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE position_locks
		SET quantity = $2, status = $3, updated_at = NOW(), released_at = $4
		WHERE id = $1
	`, lockID, remainingQuantity, status, releasedAt)
	return err
}
