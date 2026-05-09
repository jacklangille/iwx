package exchangecorehttp

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/exchangecore"
)

func (s *Server) handleContractLifecycle(w http.ResponseWriter, r *http.Request) {
	requireAuth(s.config, s.handleContractLifecycleAuthed)(w, r)
}

func (s *Server) handleContractLifecycleAuthed(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	contractID, action, err := parseContractLifecyclePath(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract path"})
		return
	}

	switch {
	case r.Method == http.MethodGet && action == "":
		s.handleContractDetailsShow(w, r, claims.UserID, contractID)
	case r.Method == http.MethodGet && action == "collateral_requirement":
		s.handleContractCollateralRequirement(w, r, claims.UserID, contractID)
	case r.Method == http.MethodGet && action == "issuance_batches":
		s.handleContractIssuanceBatchesIndex(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "submit_for_approval":
		s.handleContractSubmitForApproval(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "approve":
		s.handleContractApprove(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "collateral_locks":
		s.handleContractCollateralLockCreate(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "issuance_batches":
		s.handleContractIssuanceBatchCreate(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "activate":
		s.handleContractActivate(w, r, claims.UserID, contractID)
	case r.Method == http.MethodPost && action == "cancel":
		s.handleContractCancel(w, r, claims.UserID, contractID)
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleContractDetailsShow(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	details, err := s.service.GetContractDetails(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeContractDetails(details))
}

func (s *Server) handleContractSubmitForApproval(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	contract, err := s.service.SubmitContractForApproval(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core contract submitted_for_approval contract_id=%d user_id=%d", contractID, userID)
	writeJSON(w, http.StatusOK, serializeContractLifecycleContract(contract))
}

func (s *Server) handleContractApprove(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	contract, err := s.service.ApproveContract(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core contract approved contract_id=%d user_id=%d", contractID, userID)
	writeJSON(w, http.StatusOK, serializeContractLifecycleContract(contract))
}

func (s *Server) handleContractCollateralRequirement(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	rawPairs := strings.TrimSpace(r.URL.Query().Get("paired_quantity"))
	if rawPairs == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "paired_quantity is required"})
		return
	}

	pairedQuantity, err := strconv.ParseInt(rawPairs, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid paired_quantity"})
		return
	}

	requirement, err := s.service.CalculateCollateralRequirement(r.Context(), contractID, userID, pairedQuantity, r.URL.Query().Get("currency"))
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeCollateralRequirement(requirement))
}

func (s *Server) handleContractCollateralLockCreate(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	request, err := decodeContractCollateralLockRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	result, err := s.service.LockContractCollateral(
		r.Context(),
		contractID,
		userID,
		request.PairedQuantity,
		request.Currency,
		request.CorrelationID,
		request.Description,
	)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core contract collateral_locked contract_id=%d user_id=%d lock_id=%d paired_quantity=%d", contractID, userID, result.Lock.ID, request.PairedQuantity)
	writeJSON(w, http.StatusCreated, contractCollateralLockMutationResponse{
		Requirement:    serializeCollateralRequirement(result.Requirement),
		CollateralLock: serializeCollateralLock(result.Lock),
		Account:        serializeCashAccount(result.Account),
		LedgerEntry:    ptrLedgerEntryResponse(serializeLedgerEntry(result.LedgerEntry)),
	})
}

func (s *Server) handleContractActivate(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	contract, err := s.service.ActivateContract(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core contract activated contract_id=%d user_id=%d", contractID, userID)
	writeJSON(w, http.StatusOK, serializeContractLifecycleContract(contract))
}

func (s *Server) handleContractIssuanceBatchesIndex(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	batches, err := s.service.ListIssuanceBatches(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeIssuanceBatches(batches))
}

func (s *Server) handleContractIssuanceBatchCreate(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	request, err := decodeIssuanceBatchRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	details, err := s.service.IssueContractSupply(r.Context(), contractID, userID, request.CollateralLockID, request.PairedQuantity)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core issuance_batch_created contract_id=%d user_id=%d batch_id=%d collateral_lock_id=%d paired_quantity=%d", contractID, userID, details.Batch.ID, request.CollateralLockID, request.PairedQuantity)
	writeJSON(w, http.StatusCreated, issuanceBatchMutationResponse{
		IssuanceBatch:  serializeIssuanceBatch(details.Batch),
		CollateralLock: serializeCollateralLock(details.Lock),
		Positions:      serializePositions(details.Positions),
	})
}

func (s *Server) handleContractCancel(w http.ResponseWriter, r *http.Request, userID, contractID int64) {
	contract, err := s.service.CancelContract(r.Context(), contractID, userID)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core contract cancelled contract_id=%d user_id=%d", contractID, userID)
	writeJSON(w, http.StatusOK, serializeContractLifecycleContract(contract))
}

func parseContractLifecyclePath(path string) (int64, string, error) {
	trimmed := strings.TrimPrefix(path, "/api/contracts/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return 0, "", errors.New("missing contract id")
	}

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

func writeExchangeCoreError(w http.ResponseWriter, err error) {
	var validationErr *exchangecore.ValidationError
	if errors.As(err, &validationErr) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
		return
	}

	switch {
	case errors.Is(err, exchangecore.ErrContractNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, exchangecore.ErrContractForbidden):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, exchangecore.ErrInvalidContractState):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
	case strings.Contains(err.Error(), "insufficient available balance"):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
	case strings.Contains(err.Error(), "not found"):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case strings.Contains(err.Error(), "not active"), strings.Contains(err.Error(), "underflow"):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}
