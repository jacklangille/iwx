package httpapi

import "net/http"

func (s *Server) handleStationsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}

	activeOnly := r.URL.Query().Get("active") == "true"
	stations, err := s.reads.ListStations(r.Context(), activeOnly)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stations": serializeStations(stations),
	})
}
