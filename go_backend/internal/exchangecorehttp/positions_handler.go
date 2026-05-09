package exchangecorehttp

import (
	"net/http"
	"strconv"

	"iwx/go_backend/internal/authcontext"
)

func (s *Server) handlePositionsMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handlePositionsMeIndex)(w, r)
}

func (s *Server) handlePositionLocksMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handlePositionLocksMeIndex)(w, r)
}

func (s *Server) handlePositionsMeIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	var contractID *int64
	if rawContractID := r.URL.Query().Get("contract_id"); rawContractID != "" {
		parsed, err := strconv.ParseInt(rawContractID, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
			return
		}
		contractID = &parsed
	}

	positions, err := s.service.ListPositions(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializePositionsFromValues(positions))
}

func (s *Server) handlePositionLocksMeIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	var contractID *int64
	if rawContractID := r.URL.Query().Get("contract_id"); rawContractID != "" {
		parsed, err := strconv.ParseInt(rawContractID, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract_id"})
			return
		}
		contractID = &parsed
	}

	locks, err := s.service.ListPositionLocks(r.Context(), claims.UserID, contractID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializePositionLocks(locks))
}
