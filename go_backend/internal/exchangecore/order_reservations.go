package exchangecore

import (
	"context"
	"strings"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/money"
	"iwx/go_backend/internal/store"
)

type OrderReservation struct {
	CashReservation *domain.OrderCashReservation
	PositionLock    *domain.PositionLock
}

func (s *Service) ListPositionLocks(ctx context.Context, userID int64, contractID *int64) ([]domain.PositionLock, error) {
	if userID <= 0 {
		return nil, &ValidationError{Errors: map[string][]string{
			"user_id": {"must be present"},
		}}
	}

	return s.repo.ListPositionLocks(ctx, userID, contractID)
}

func (s *Service) CreatePositionLock(ctx context.Context, input store.CreatePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	errors := map[string][]string{}
	if input.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if input.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if input.Quantity <= 0 {
		errors["quantity"] = append(errors["quantity"], "must be greater than 0")
	}
	if side := strings.ToLower(strings.TrimSpace(input.Side)); side != string(domain.ClaimSideAbove) && side != string(domain.ClaimSideBelow) {
		errors["side"] = append(errors["side"], "is invalid")
	}
	if len(errors) > 0 {
		return nil, nil, &ValidationError{Errors: errors}
	}

	input.Side = strings.ToLower(strings.TrimSpace(input.Side))
	lock, position, err := s.repo.CreatePositionLock(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	return lock, position, s.projectUser(ctx, input.UserID)
}

func (s *Service) ReleasePositionLock(ctx context.Context, input store.ReleasePositionLockInput) (*domain.PositionLock, *domain.Position, error) {
	errors := map[string][]string{}
	if input.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if input.LockID <= 0 {
		errors["lock_id"] = append(errors["lock_id"], "must be present")
	}
	if len(errors) > 0 {
		return nil, nil, &ValidationError{Errors: errors}
	}

	lock, position, err := s.repo.ReleasePositionLock(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	return lock, position, s.projectUser(ctx, input.UserID)
}

func (s *Service) ReserveOrderForMatching(ctx context.Context, command commands.PlaceOrder, commandID string) (*OrderReservation, error) {
	if err := ValidateOrderReservation(command); err != nil {
		return nil, err
	}

	if command.OrderSide == string(domain.OrderSideBid) {
		priceCents, err := money.ParseCents(command.Price)
		if err != nil {
			return nil, &ValidationError{Errors: map[string][]string{
				"price": {err.Error()},
			}}
		}

		reservation, _, _, err := s.repo.CreateOrderCashReservation(ctx, store.CreateOrderCashReservationInput{
			UserID:        command.UserID,
			ContractID:    command.ContractID,
			Currency:      DefaultAccountCurrency,
			AmountCents:   priceCents * command.Quantity,
			ReferenceType: "order_command",
			ReferenceID:   commandID,
			CorrelationID: commandID,
			Description:   "reserved cash for bid order",
		})
		if err != nil {
			return nil, err
		}

		return &OrderReservation{CashReservation: reservation}, s.projectUser(ctx, command.UserID)
	}

	lock, _, err := s.repo.CreatePositionLock(ctx, store.CreatePositionLockInput{
		UserID:        command.UserID,
		ContractID:    command.ContractID,
		Side:          command.TokenType,
		Quantity:      command.Quantity,
		ReferenceType: "order_command",
		ReferenceID:   commandID,
		CorrelationID: commandID,
		Description:   "reserved inventory for ask order",
	})
	if err != nil {
		return nil, err
	}

	return &OrderReservation{PositionLock: lock}, s.projectUser(ctx, command.UserID)
}

func (s *Service) ReleaseOrderReservation(ctx context.Context, userID int64, reservation *OrderReservation, correlationID string) error {
	if reservation == nil {
		return nil
	}

	if reservation.CashReservation != nil {
		_, _, _, err := s.repo.ReleaseOrderCashReservation(ctx, store.ReleaseOrderCashReservationInput{
			UserID:        userID,
			ReservationID: reservation.CashReservation.ID,
			CorrelationID: correlationID,
			Description:   "order enqueue failed",
		})
		if err != nil {
			return err
		}
		return s.projectUser(ctx, userID)
	}

	if reservation.PositionLock != nil {
		_, _, err := s.repo.ReleasePositionLock(ctx, store.ReleasePositionLockInput{
			UserID:        userID,
			LockID:        reservation.PositionLock.ID,
			CorrelationID: correlationID,
			Description:   "order enqueue failed",
		})
		if err != nil {
			return err
		}
		return s.projectUser(ctx, userID)
	}

	return nil
}

func ValidateOrderReservation(command commands.PlaceOrder) error {
	errors := map[string][]string{}
	if command.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}
	if command.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if command.Quantity <= 0 {
		errors["quantity"] = append(errors["quantity"], "must be greater than 0")
	}
	tokenType := strings.ToLower(strings.TrimSpace(command.TokenType))
	if tokenType != string(domain.ClaimSideAbove) && tokenType != string(domain.ClaimSideBelow) {
		errors["token_type"] = append(errors["token_type"], "is invalid")
	}
	orderSide := strings.ToLower(strings.TrimSpace(command.OrderSide))
	if orderSide != string(domain.OrderSideBid) && orderSide != string(domain.OrderSideAsk) {
		errors["order_side"] = append(errors["order_side"], "is invalid")
	}
	if _, err := money.ParseCents(command.Price); err != nil {
		errors["price"] = append(errors["price"], err.Error())
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}
