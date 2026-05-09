package exchangecore

import (
	"context"
	"strings"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/store"
)

const DefaultAccountCurrency = "USD"

func normalizeCurrency(value string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return DefaultAccountCurrency
	}

	return trimmed
}

func validateMoneyInput(userID, amountCents int64, currency string) map[string][]string {
	errors := map[string][]string{}
	if userID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if amountCents <= 0 {
		errors["amount_cents"] = append(errors["amount_cents"], "must be greater than 0")
	}
	if strings.TrimSpace(currency) == "" {
		errors["currency"] = append(errors["currency"], "can't be blank")
	}

	return errors
}

func (s *Service) GetCashAccount(ctx context.Context, userID int64, currency string) (*domain.CashAccount, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.GetCashAccount(ctx, userID, normalizeCurrency(currency))
}

func (s *Service) ListLedgerEntries(ctx context.Context, userID int64, currency string, limit int) ([]domain.LedgerEntry, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.ListLedgerEntries(ctx, userID, normalizeCurrency(currency), limit)
}

func (s *Service) DepositCash(ctx context.Context, input store.DepositCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	input.Currency = normalizeCurrency(input.Currency)
	errors := validateMoneyInput(input.UserID, input.AmountCents, input.Currency)
	if len(errors) > 0 {
		return nil, nil, &ValidationError{Errors: errors}
	}

	account, entry, err := s.repo.DepositCash(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	return account, entry, s.projectUser(ctx, input.UserID)
}

func (s *Service) WithdrawCash(ctx context.Context, input store.WithdrawCashInput) (*domain.CashAccount, *domain.LedgerEntry, error) {
	input.Currency = normalizeCurrency(input.Currency)
	errors := validateMoneyInput(input.UserID, input.AmountCents, input.Currency)
	if len(errors) > 0 {
		return nil, nil, &ValidationError{Errors: errors}
	}

	account, entry, err := s.repo.WithdrawCash(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	return account, entry, s.projectUser(ctx, input.UserID)
}

func (s *Service) ListCollateralLocks(ctx context.Context, userID int64, currency string) ([]domain.CollateralLock, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.ListCollateralLocks(ctx, userID, normalizeCurrency(currency))
}

func (s *Service) CreateCollateralLock(ctx context.Context, input store.CreateCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	input.Currency = normalizeCurrency(input.Currency)
	errors := validateMoneyInput(input.UserID, input.AmountCents, input.Currency)
	if input.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if len(errors) > 0 {
		return nil, nil, nil, &ValidationError{Errors: errors}
	}

	lock, account, entry, err := s.repo.CreateCollateralLock(ctx, input)
	if err != nil {
		return nil, nil, nil, err
	}

	return lock, account, entry, s.projectUser(ctx, input.UserID)
}

func (s *Service) ReleaseCollateralLock(ctx context.Context, input store.ReleaseCollateralLockInput) (*domain.CollateralLock, *domain.CashAccount, *domain.LedgerEntry, error) {
	errors := map[string][]string{}
	if input.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if input.LockID <= 0 {
		errors["lock_id"] = append(errors["lock_id"], "must be present")
	}
	if len(errors) > 0 {
		return nil, nil, nil, &ValidationError{Errors: errors}
	}

	lock, account, entry, err := s.repo.ReleaseCollateralLock(ctx, input)
	if err != nil {
		return nil, nil, nil, err
	}

	return lock, account, entry, s.projectUser(ctx, input.UserID)
}

func (s *Service) ListOrderCashReservations(ctx context.Context, userID int64, currency string) ([]domain.OrderCashReservation, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.ListOrderCashReservations(ctx, userID, normalizeCurrency(currency))
}

func (s *Service) CreateOrderCashReservation(ctx context.Context, input store.CreateOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	input.Currency = normalizeCurrency(input.Currency)
	errors := validateMoneyInput(input.UserID, input.AmountCents, input.Currency)
	if input.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if strings.TrimSpace(input.ReferenceType) == "" {
		errors["reference_type"] = append(errors["reference_type"], "can't be blank")
	}
	if strings.TrimSpace(input.ReferenceID) == "" {
		errors["reference_id"] = append(errors["reference_id"], "can't be blank")
	}
	if len(errors) > 0 {
		return nil, nil, nil, &ValidationError{Errors: errors}
	}

	reservation, account, entry, err := s.repo.CreateOrderCashReservation(ctx, input)
	if err != nil {
		return nil, nil, nil, err
	}

	return reservation, account, entry, s.projectUser(ctx, input.UserID)
}

func (s *Service) ReleaseOrderCashReservation(ctx context.Context, input store.ReleaseOrderCashReservationInput) (*domain.OrderCashReservation, *domain.CashAccount, *domain.LedgerEntry, error) {
	errors := map[string][]string{}
	if input.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if input.ReservationID <= 0 {
		errors["reservation_id"] = append(errors["reservation_id"], "must be present")
	}
	if len(errors) > 0 {
		return nil, nil, nil, &ValidationError{Errors: errors}
	}

	reservation, account, entry, err := s.repo.ReleaseOrderCashReservation(ctx, input)
	if err != nil {
		return nil, nil, nil, err
	}

	return reservation, account, entry, s.projectUser(ctx, input.UserID)
}
