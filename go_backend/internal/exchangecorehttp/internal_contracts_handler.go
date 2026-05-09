package exchangecorehttp

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"iwx/go_backend/internal/store"
)

func (s *Server) handleInternalContracts(w http.ResponseWriter, r *http.Request) {
	contractID, action, err := parseInternalContractPath(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract path"})
		return
	}

	switch {
	case r.Method == http.MethodGet && action == "":
		s.handleInternalContractShow(w, r, contractID)
	case r.Method == http.MethodPost && action == "resolve":
		s.handleInternalContractResolve(w, r, contractID)
	case r.Method == http.MethodPost && action == "settlement":
		s.handleInternalContractSettlement(w, r, contractID)
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleInternalContractShow(w http.ResponseWriter, r *http.Request, contractID int64) {
	contract, err := s.service.GetContractByID(r.Context(), contractID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}
	if contract == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
		return
	}

	rule, err := s.service.GetContractRuleByContractID(r.Context(), contractID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, internalContractResponse{
		Contract: serializeContractLifecycleContract(contract),
		Rule:     serializeContractRule(rule),
	})
}

func (s *Server) handleInternalContractResolve(w http.ResponseWriter, r *http.Request, contractID int64) {
	contract, err := s.service.MarkContractResolved(r.Context(), contractID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, serializeContractLifecycleContract(contract))
}

func (s *Server) handleInternalContractSettlement(w http.ResponseWriter, r *http.Request, contractID int64) {
	request, err := decodeInternalSettlementRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	result, err := s.service.SettleContract(r.Context(), store.SettleContractInput{
		ContractID:    contractID,
		EventID:       request.EventID,
		Outcome:       request.Outcome,
		ResolvedAt:    request.ResolvedAt,
		CorrelationID: request.CorrelationID,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, internalSettlementResponse{
		Contract:      serializeContractLifecycleContract(result.Contract),
		AffectedUsers: result.AffectedUsers,
		SettledAt:     result.SettledAt.UTC().Format(time.RFC3339),
	})
}

func parseInternalContractPath(path string) (int64, string, error) {
	trimmed := strings.TrimPrefix(path, "/internal/contracts/")
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.Split(trimmed, "/")
	contractID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	return contractID, action, nil
}
