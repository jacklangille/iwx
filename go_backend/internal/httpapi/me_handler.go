package httpapi

import (
	"net/http"
	"strconv"

	"iwx/go_backend/internal/authcontext"
)

func (s *Server) handleMeAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMeAccountShow)(w, r)
}

func (s *Server) handleMePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMePositionsIndex)(w, r)
}

func (s *Server) handleMePositionLocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMePositionLocksIndex)(w, r)
}

func (s *Server) handleMeCollateralLocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMeCollateralLocksIndex)(w, r)
}

func (s *Server) handleMeCashReservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMeCashReservationsIndex)(w, r)
}

func (s *Server) handleMePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMePortfolioShow)(w, r)
}

func (s *Server) handleMeSettlements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleMeSettlementsIndex)(w, r)
}

func (s *Server) handleMeAccountShow(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	accounts, err := s.reads.ListCashAccounts(r.Context(), claims.UserID)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeCashAccounts(accounts))
}

func (s *Server) handleMePositionsIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	contractID, ok := optionalContractID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
		return
	}

	positions, err := s.reads.ListUserPositions(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializePositions(positions))
}

func (s *Server) handleMePositionLocksIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	contractID, ok := optionalContractID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
		return
	}

	locks, err := s.reads.ListUserPositionLocks(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializePositionLocks(locks))
}

func (s *Server) handleMeCollateralLocksIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	locks, err := s.reads.ListUserCollateralLocks(r.Context(), claims.UserID, normalizeCurrencyQuery(r))
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeCollateralLocks(locks))
}

func (s *Server) handleMeCashReservationsIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	reservations, err := s.reads.ListUserCashReservations(r.Context(), claims.UserID, normalizeCurrencyQuery(r))
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeOrderCashReservations(reservations))
}

func (s *Server) handleMePortfolioShow(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	contractID, ok := optionalContractID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
		return
	}

	accounts, err := s.reads.ListCashAccounts(r.Context(), claims.UserID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	positions, err := s.reads.ListUserPositions(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	positionLocks, err := s.reads.ListUserPositionLocks(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	currency := normalizeCurrencyQuery(r)
	collateralLocks, err := s.reads.ListUserCollateralLocks(r.Context(), claims.UserID, currency)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	cashReservations, err := s.reads.ListUserCashReservations(r.Context(), claims.UserID, currency)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":           claims.UserID,
		"accounts":          serializeCashAccounts(accounts),
		"positions":         serializePositions(positions),
		"position_locks":    serializePositionLocks(positionLocks),
		"collateral_locks":  serializeCollateralLocks(collateralLocks),
		"cash_reservations": serializeOrderCashReservations(cashReservations),
	})
}

func (s *Server) handleMeSettlementsIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	contractID, ok := optionalContractID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
		return
	}

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	entries, err := s.reads.ListUserSettlementEntries(r.Context(), claims.UserID, contractID, limit)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeSettlementEntries(entries))
}

func optionalContractID(r *http.Request) (*int64, bool) {
	rawContractID := r.URL.Query().Get("contract_id")
	if rawContractID == "" {
		return nil, true
	}

	parsed, err := strconv.ParseInt(rawContractID, 10, 64)
	if err != nil {
		return nil, false
	}

	return &parsed, true
}

func normalizeCurrencyQuery(r *http.Request) string {
	currency := r.URL.Query().Get("currency")
	if currency == "" {
		return "USD"
	}

	return currency
}
