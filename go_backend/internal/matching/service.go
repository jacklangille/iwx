package matching

import (
	"context"
	"fmt"
	"strings"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/money"
	"iwx/go_backend/internal/store"
)

type CommandClient interface {
	SubmitPlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderAccepted, error)
	Close()
}

type Handler interface {
	HandlePlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error)
}

type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type Service struct {
	repo store.MatchingRepository
}

func NewService(repo store.MatchingRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HandlePlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderResult, error) {
	if err := ValidatePlaceOrder(envelope.Command); err != nil {
		return commands.PlaceOrderResult{}, err
	}

	return s.repo.ProcessPlaceOrder(ctx, envelope)
}

func ValidatePlaceOrder(command commands.PlaceOrder) error {
	errors := map[string][]string{}

	if command.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if command.UserID <= 0 {
		errors["user_id"] = append(errors["user_id"], "must be present")
	}

	tokenType := strings.TrimSpace(command.TokenType)
	if tokenType == "" {
		errors["token_type"] = append(errors["token_type"], "can't be blank")
	} else if tokenType != string(domain.ClaimSideAbove) && tokenType != string(domain.ClaimSideBelow) {
		errors["token_type"] = append(errors["token_type"], "is invalid")
	}

	orderSide := strings.TrimSpace(command.OrderSide)
	if orderSide == "" {
		errors["order_side"] = append(errors["order_side"], "can't be blank")
	} else if orderSide != string(domain.OrderSideBid) && orderSide != string(domain.OrderSideAsk) {
		errors["order_side"] = append(errors["order_side"], "is invalid")
	}

	price, err := money.ParseCents(command.Price)
	if err != nil {
		errors["price"] = append(errors["price"], fmt.Sprintf("is invalid: %v", err))
	} else if price <= 0 {
		errors["price"] = append(errors["price"], "must be greater than 0")
	}

	if command.Quantity <= 0 {
		errors["quantity"] = append(errors["quantity"], "must be greater than 0")
	}
	if orderSide == string(domain.OrderSideBid) {
		if command.CashReservationID == nil || *command.CashReservationID <= 0 {
			errors["cash_reservation_id"] = append(errors["cash_reservation_id"], "must be present for bid orders")
		}
	}
	if orderSide == string(domain.OrderSideAsk) {
		if command.PositionLockID == nil || *command.PositionLockID <= 0 {
			errors["position_lock_id"] = append(errors["position_lock_id"], "must be present for ask orders")
		}
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}
