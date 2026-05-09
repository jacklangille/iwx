package matcherhttp

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/projectionbundle"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/internal/store"
	"iwx/go_backend/pkg/logging"
)

type Server struct {
	config     config.Config
	orders     store.OrderRepository
	executions store.ExecutionRepository
	snapshots  store.SnapshotProjectionSource
	orderCmds  store.OrderCommandRepository
	mux        *http.ServeMux
}

func NewServer(cfg config.Config, orders store.OrderRepository, executions store.ExecutionRepository, snapshots store.SnapshotProjectionSource, orderCmds store.OrderCommandRepository) *Server {
	s := &Server{
		config:     cfg,
		orders:     orders,
		executions: executions,
		snapshots:  snapshots,
		orderCmds:  orderCmds,
		mux:        http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/internal/projection/", s.handleInternalProjection)
}

func (s *Server) ListenAndServe() error {
	server := &http.Server{
		Addr:    s.config.MatcherHTTPAddr,
		Handler: s.loggingMiddleware(s.mux),
	}
	log.Printf("matcher http server listening addr=%s", s.config.MatcherHTTPAddr)
	return server.ListenAndServe()
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": "iwx-matcher", "status": "ok"})
}

func (s *Server) handleInternalProjection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/internal/projection/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 3 && parts[0] == "contracts" && parts[2] == "market":
		contractID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
			return
		}
		orders, err := s.orders.ListOpenOrders(r.Context(), &contractID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		executions, err := s.executions.ListExecutions(r.Context(), contractID, 200)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		snapshots, err := s.snapshots.ListAllSnapshots(r.Context(), contractID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, projectionbundle.MarketBundle{
			ContractID: contractID,
			Orders:     orders,
			Executions: executions,
			Snapshots:  snapshots,
		})
	case len(parts) == 2 && parts[0] == "order_commands":
		command, err := s.orderCmds.GetOrderCommand(r.Context(), parts[1])
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if command == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "order command not found"})
			return
		}
		writeJSON(w, http.StatusOK, projectionbundle.OrderCommandBundle{Command: command})
	default:
		http.NotFound(w, r)
	}
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
			"matcher_http_request",
			"method", r.Method,
			"path", r.URL.RequestURI(),
			"status", recorder.status,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"remote", r.RemoteAddr,
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
