package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleOrdersIndex(w, r)
	default:
		methodNotAllowed(w, "GET")
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
