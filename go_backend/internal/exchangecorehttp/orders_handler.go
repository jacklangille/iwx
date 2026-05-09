package exchangecorehttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nuid"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleOrdersCreate)(w, r)
}

func (s *Server) handleOrdersCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	command, err := decodePlaceOrderCommand(r.Body)
	if err != nil {
		logging.Error(r.Context(), "exchange_core_order_decode_failed", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	command.UserID = claims.UserID

	if err := exchangecore.ValidateOrderReservation(command); err != nil {
		var validationErr *exchangecore.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	commandID := nuid.Next()
	enqueuedAt := time.Now().UTC().Truncate(time.Millisecond)

	reservation, err := s.service.ReserveOrderForMatching(r.Context(), command, commandID)
	if err != nil {
		var validationErr *exchangecore.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
		return
	}

	if reservation != nil {
		if reservation.CashReservation != nil {
			command.CashReservationID = &reservation.CashReservation.ID
		}
		if reservation.PositionLock != nil {
			command.PositionLockID = &reservation.PositionLock.ID
		}
		command.ReservationCorrelationID = commandID
	}

	if err := matching.ValidatePlaceOrder(command); err != nil {
		releaseErr := s.service.ReleaseOrderReservation(r.Context(), command.UserID, reservation, commandID)
		if releaseErr != nil {
			logging.Error(r.Context(), "exchange_core_order_reservation_release_failed", releaseErr, "command_id", commandID, "user_id", command.UserID)
		}

		var validationErr *matching.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	envelope := commands.PlaceOrderEnvelope{
		CommandID:  commandID,
		TraceID:    requestctx.TraceID(r.Context()),
		EnqueuedAt: enqueuedAt,
		Command:    command,
	}

	result, err := s.matcher.SubmitPlaceOrder(r.Context(), envelope)
	if err != nil {
		releaseErr := s.service.ReleaseOrderReservation(r.Context(), command.UserID, reservation, commandID)
		if releaseErr != nil {
			logging.Error(r.Context(), "exchange_core_order_reservation_release_failed", releaseErr, "command_id", commandID, "user_id", command.UserID)
		}

		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, serializePlaceOrderAccepted(result))
}

func decodePlaceOrderCommand(body io.Reader) (commands.PlaceOrder, error) {
	if body == nil {
		return commands.PlaceOrder{}, errors.New("request body is required")
	}

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return commands.PlaceOrder{}, err
	}

	command := commands.PlaceOrder{}
	if raw := payload["contract_id"]; raw != nil {
		if err := json.Unmarshal(raw, &command.ContractID); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["token_type"]; raw != nil {
		if err := json.Unmarshal(raw, &command.TokenType); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["order_side"]; raw != nil {
		if err := json.Unmarshal(raw, &command.OrderSide); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["quantity"]; raw != nil {
		if err := json.Unmarshal(raw, &command.Quantity); err != nil {
			return commands.PlaceOrder{}, err
		}
	}
	if raw := payload["price"]; raw != nil {
		command.Price = decodePrice(raw)
	}

	command.TokenType = strings.ToLower(strings.TrimSpace(command.TokenType))
	command.OrderSide = strings.ToLower(strings.TrimSpace(command.OrderSide))

	return command, nil
}

func decodePrice(raw json.RawMessage) string {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}

	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber.String()
	}

	return ""
}
