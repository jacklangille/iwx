package httpapi

import (
	"errors"
	"net/http"
	"strconv"
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
	switch r.Method {
	case http.MethodGet:
		s.handleOrdersIndex(w, r)
	case http.MethodPost:
		requireAuth(s.config, s.handleOrdersCreate)(w, r)
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleOrdersIndex(w http.ResponseWriter, r *http.Request) {
	var contractID *int64
	if rawID := r.URL.Query().Get("contract_id"); rawID != "" {
		parsed, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
			return
		}

		contractID = &parsed
	}

	orders, err := s.reads.ListOpenOrders(r.Context(), contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeOrders(orders))
}

func (s *Server) handleOrdersCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	command, err := decodePlaceOrderCommand(r.Body)
	if err != nil {
		logging.Error(r.Context(), "api_order_decode_failed", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	command.UserID = claims.UserID

	if err := matching.ValidatePlaceOrder(command); err != nil {
		var validationErr *matching.ValidationError
		if errors.As(err, &validationErr) {
			logging.Error(r.Context(), "api_order_validation_failed", err, "contract_id", command.ContractID, "token_type", command.TokenType, "order_side", command.OrderSide, "errors", validationErr.Errors)
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}

		logging.Error(r.Context(), "api_order_invalid", err, "contract_id", command.ContractID)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	commandID := nuid.Next()
	enqueuedAt := time.Now().UTC().Truncate(time.Millisecond)

	reservation, err := s.exchangeCore.ReserveOrderForMatching(r.Context(), command, commandID)
	if err != nil {
		var validationErr *exchangecore.ValidationError
		if errors.As(err, &validationErr) {
			logging.Error(r.Context(), "api_order_reservation_validation_failed", err, "user_id", command.UserID, "contract_id", command.ContractID, "token_type", command.TokenType, "order_side", command.OrderSide, "errors", validationErr.Errors)
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}

		logging.Error(r.Context(), "api_order_reservation_failed", err, "user_id", command.UserID, "contract_id", command.ContractID, "token_type", command.TokenType, "order_side", command.OrderSide)
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

	envelope := commands.PlaceOrderEnvelope{
		CommandID:  commandID,
		TraceID:    requestctx.TraceID(r.Context()),
		EnqueuedAt: enqueuedAt,
		Command:    command,
	}

	result, err := s.matcher.SubmitPlaceOrder(r.Context(), envelope)
	if err != nil {
		releaseErr := s.exchangeCore.ReleaseOrderReservation(r.Context(), command.UserID, reservation, commandID)
		if releaseErr != nil {
			logging.Error(r.Context(), "api_order_reservation_release_failed", releaseErr, "command_id", commandID, "user_id", command.UserID)
		}

		logging.Error(r.Context(), "api_order_enqueue_failed", err, "user_id", command.UserID, "contract_id", command.ContractID, "token_type", command.TokenType, "order_side", command.OrderSide)
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}

	logging.Info(r.Context(), "api_order_submitted", "command_id", result.CommandID, "user_id", command.UserID, "contract_id", result.ContractID, "partition", result.Partition, "status", result.Status)
	writeJSON(w, http.StatusAccepted, serializePlaceOrderAccepted(result))
}
