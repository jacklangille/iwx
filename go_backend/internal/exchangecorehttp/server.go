package exchangecorehttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

type Server struct {
	config  config.Config
	service *exchangecore.Service
	matcher matching.CommandClient
	mux     *http.ServeMux
}

func NewServer(cfg config.Config, service *exchangecore.Service, matcher matching.CommandClient) *Server {
	server := &Server{
		config:  cfg,
		service: service,
		matcher: matcher,
		mux:     http.NewServeMux(),
	}

	server.registerRoutes()
	return server
}

func (s *Server) ListenAndServe() error {
	httpServer := &http.Server{
		Addr:    s.config.ExchangeCoreHTTPAddr,
		Handler: s.loggingMiddleware(s.mux),
	}

	log.Printf("exchange-core server listening addr=%s", s.config.ExchangeCoreHTTPAddr)
	return httpServer.ListenAndServe()
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/internal/contracts/", s.handleInternalContracts)
	s.mux.HandleFunc("/internal/projection/", s.handleInternalProjection)
	s.mux.HandleFunc("/api/accounts/me", s.handleAccountMe)
	s.mux.HandleFunc("/api/accounts/me/ledger", s.handleAccountLedger)
	s.mux.HandleFunc("/api/accounts/me/collateral_locks", s.handleAccountCollateralLocks)
	s.mux.HandleFunc("/api/accounts/me/cash_reservations", s.handleAccountCashReservations)
	s.mux.HandleFunc("/api/accounts/deposits", s.handleAccountDeposits)
	s.mux.HandleFunc("/api/accounts/withdrawals", s.handleAccountWithdrawals)
	s.mux.HandleFunc("/api/accounts/collateral_locks", s.handleCollateralLocks)
	s.mux.HandleFunc("/api/accounts/collateral_locks/", s.handleCollateralLockActions)
	s.mux.HandleFunc("/api/accounts/cash_reservations", s.handleCashReservations)
	s.mux.HandleFunc("/api/accounts/cash_reservations/", s.handleCashReservationActions)
	s.mux.HandleFunc("/api/positions/me", s.handlePositionsMe)
	s.mux.HandleFunc("/api/positions/me/locks", s.handlePositionLocksMe)
	s.mux.HandleFunc("/api/orders", s.handleOrders)
	s.mux.HandleFunc("/api/contracts", s.handleContracts)
	s.mux.HandleFunc("/api/contracts/", s.handleContractLifecycle)
	s.mux.HandleFunc("/api/contract_commands/", s.handleContractCommands)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "iwx-go-exchange-core",
		"status":  "ok",
	})
}

func (s *Server) handleContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleContractsCreate)(w, r)
}

func (s *Server) handleContractsCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	command, err := decodeCreateContractCommand(r.Body)
	if err != nil {
		logging.Error(r.Context(), "exchange_core_contract_decode_failed", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	command.CreatorUserID = claims.UserID

	result, err := s.service.SubmitCreateContract(r.Context(), command)
	if err != nil {
		var validationErr *exchangecore.ValidationError
		if errors.As(err, &validationErr) {
			logging.Error(r.Context(), "exchange_core_contract_validation_failed", err, "name", command.Name, "errors", validationErr.Errors)
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}

		logging.Error(r.Context(), "exchange_core_contract_submit_failed", err, "name", command.Name)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	logging.Info(r.Context(), "exchange_core_contract_submitted", "command_id", result.CommandID, "creator_user_id", command.CreatorUserID, "status", result.Status, "name", command.Name)
	writeJSON(w, http.StatusAccepted, createContractAcceptedResponse{
		CommandID:  result.CommandID,
		Partition:  result.Partition,
		Status:     result.Status,
		EnqueuedAt: result.EnqueuedAt.UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleContractCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleContractCommandShow)(w, r)
}

func (s *Server) handleContractCommandShow(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	commandID := strings.TrimPrefix(r.URL.Path, "/api/contract_commands/")
	commandID = strings.Trim(commandID, "/")
	if commandID == "" {
		http.NotFound(w, r)
		return
	}

	command, err := s.service.GetContractCommand(r.Context(), commandID)
	if err != nil {
		logging.Error(r.Context(), "exchange_core_contract_command_lookup_failed", err, "command_id", commandID)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if command == nil {
		logging.Info(r.Context(), "exchange_core_contract_command_not_found", "command_id", commandID)
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract command not found"})
		return
	}
	if command.CreatorUserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
		return
	}

	logging.Info(r.Context(), "exchange_core_contract_command_fetched", "command_id", command.CommandID, "status", command.CommandStatus)
	writeJSON(w, http.StatusOK, serializeContractCommand(command))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		requestID, traceID := requestctx.ExtractOrGenerate(r)
		requestctx.ApplyHeaders(recorder, requestID, traceID)
		ctx := requestctx.WithHTTPContext(r.Context(), requestID, traceID)
		next.ServeHTTP(recorder, r.WithContext(ctx))
		logging.Info(
			ctx,
			"exchange_core_http_request",
			"method",
			r.Method,
			"path",
			r.URL.RequestURI(),
			"status",
			recorder.status,
			"duration_ms",
			time.Since(startedAt).Milliseconds(),
			"remote",
			r.RemoteAddr,
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
