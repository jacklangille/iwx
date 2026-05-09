package exchangecorehttp

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleInternalProjection(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/internal/projection/")
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.Split(trimmed, "/")

	switch {
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "contracts":
		contractID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
			return
		}
		bundle, err := s.service.GetContractBundle(r.Context(), contractID)
		if err != nil {
			writeExchangeCoreError(w, err)
			return
		}
		if bundle == nil || bundle.Contract == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		writeJSON(w, http.StatusOK, bundle)
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "users":
		userID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid user id"})
			return
		}
		bundle, err := s.service.GetUserStateBundle(r.Context(), userID)
		if err != nil {
			writeExchangeCoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bundle)
	case r.Method == http.MethodGet && len(parts) == 3 && parts[0] == "settlements" && parts[1] == "contracts":
		contractID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
			return
		}
		bundle, err := s.service.GetSettlementBundle(r.Context(), contractID)
		if err != nil {
			writeExchangeCoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bundle)
	default:
		methodNotAllowed(w, "GET")
	}
}
