package httpapi

import "net/http"

func (s *Server) handleContractsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}

	summaries, err := s.reads.ListContractSummaries(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeContracts(summaries))
}
