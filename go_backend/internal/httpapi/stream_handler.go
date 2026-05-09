package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/readmodel"
)

func (s *Server) handleStreams(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/stream/")
	path = strings.Trim(path, "/")

	switch {
	case strings.HasPrefix(path, "contracts/"):
		s.handleContractMarketStream(w, r, path)
	case strings.HasPrefix(path, "me/"):
		requireAuth(s.config, s.handleAuthedStream)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleAuthedStream(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/stream/me/")
	path = strings.Trim(path, "/")

	switch {
	case path == "portfolio":
		s.handlePortfolioStream(w, r)
	case strings.HasPrefix(path, "order_commands/"):
		s.handleOrderCommandStream(w, r, path)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleContractMarketStream(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] != "contracts" || parts[2] != "market" {
		http.NotFound(w, r)
		return
	}

	contractID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
		return
	}

	s.streamEvents(w, r, func() (any, error) {
		state, err := s.reads.MarketState(r.Context(), contractID)
		if err != nil {
			return nil, err
		}
		executions, err := s.reads.ListExecutions(r.Context(), contractID, 25)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"contract_id":  contractID,
			"market_state": serializeMarketState(state),
			"executions":   serializeExecutions(executions),
		}, nil
	})
}

func (s *Server) handleOrderCommandStream(w http.ResponseWriter, r *http.Request, path string) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	commandID := strings.TrimPrefix(path, "order_commands/")
	commandID = strings.Trim(commandID, "/")
	if commandID == "" {
		http.NotFound(w, r)
		return
	}

	s.streamEvents(w, r, func() (any, error) {
		command, err := s.reads.GetOrderCommand(r.Context(), commandID)
		if err != nil {
			return nil, err
		}
		if command == nil {
			return map[string]any{"error": "order command not found"}, nil
		}
		if command.UserID != claims.UserID {
			return map[string]any{"error": "forbidden"}, nil
		}

		return serializeOrderCommand(*command), nil
	})
}

func (s *Server) handlePortfolioStream(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	s.streamEvents(w, r, func() (any, error) {
		accounts, err := s.reads.ListCashAccounts(r.Context(), claims.UserID)
		if err != nil {
			return nil, err
		}
		positions, err := s.reads.ListUserPositions(r.Context(), claims.UserID, nil)
		if err != nil {
			return nil, err
		}
		locks, err := s.reads.ListUserPositionLocks(r.Context(), claims.UserID, nil)
		if err != nil {
			return nil, err
		}
		collateralLocks, err := s.reads.ListUserCollateralLocks(r.Context(), claims.UserID, "USD")
		if err != nil {
			return nil, err
		}
		cashReservations, err := s.reads.ListUserCashReservations(r.Context(), claims.UserID, "USD")
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"user_id":           claims.UserID,
			"accounts":          serializeCashAccounts(accounts),
			"positions":         serializePositions(positions),
			"position_locks":    serializePositionLocks(locks),
			"collateral_locks":  serializeCollateralLocks(collateralLocks),
			"cash_reservations": serializeOrderCashReservations(cashReservations),
		}, nil
	})
}

func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request, fetch func() (any, error)) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	lastPayload := ""
	for {
		payload, err := fetch()
		if err != nil {
			if errors.Is(err, readmodel.ErrContractNotFound) {
				_ = writeSSE(w, "error", map[string]any{"error": "contract not found"})
				flusher.Flush()
				return
			}

			_ = writeSSE(w, "error", map[string]any{"error": err.Error()})
			flusher.Flush()
			return
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			_ = writeSSE(w, "error", map[string]any{"error": err.Error()})
			flusher.Flush()
			return
		}

		if string(raw) != lastPayload {
			lastPayload = string(raw)
			_, _ = w.Write([]byte("event: update\n"))
			_, _ = w.Write([]byte("data: " + string(raw) + "\n\n"))
			flusher.Flush()
		}

		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte("event: " + event + "\n")); err != nil {
		return err
	}
	_, err = w.Write([]byte("data: " + string(raw) + "\n\n"))
	return err
}
