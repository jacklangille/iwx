package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/readmodel"
	"iwx/go_backend/internal/realtime"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

type Server struct {
	config       config.Config
	reads        *readmodel.Service
	exchangeCore *exchangecore.Service
	hub          *realtime.Hub
	matcher      matching.CommandClient
	mux          *http.ServeMux
}

func NewServer(
	cfg config.Config,
	reads *readmodel.Service,
	exchangeCore *exchangecore.Service,
	hub *realtime.Hub,
	matcher matching.CommandClient,
) *Server {
	server := &Server{
		config:       cfg,
		reads:        reads,
		exchangeCore: exchangeCore,
		hub:          hub,
		matcher:      matcher,
		mux:          http.NewServeMux(),
	}

	server.registerRoutes()

	return server
}

func (s *Server) ListenAndServe(_ context.Context) error {
	httpServer := &http.Server{
		Addr:    s.config.HTTPAddr,
		Handler: s.loggingMiddleware(s.mux),
	}

	log.Printf("http server listening addr=%s", s.config.HTTPAddr)
	return httpServer.ListenAndServe()
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/contracts", s.handleContractsIndex)
	s.mux.HandleFunc("/api/stations", s.handleStationsIndex)
	s.mux.HandleFunc("/api/orders", s.handleOrders)
	s.mux.HandleFunc("/api/order_commands/", s.handleOrderCommands)
	s.mux.HandleFunc("/api/contracts/", s.handleContractSubroutes)
	s.mux.HandleFunc("/api/me/account", s.handleMeAccount)
	s.mux.HandleFunc("/api/me/positions", s.handleMePositions)
	s.mux.HandleFunc("/api/me/position_locks", s.handleMePositionLocks)
	s.mux.HandleFunc("/api/me/collateral_locks", s.handleMeCollateralLocks)
	s.mux.HandleFunc("/api/me/cash_reservations", s.handleMeCashReservations)
	s.mux.HandleFunc("/api/me/portfolio", s.handleMePortfolio)
	s.mux.HandleFunc("/api/me/settlements", s.handleMeSettlements)
	s.mux.HandleFunc("/api/stream/", s.handleStreams)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "iwx-go-api",
		"status":  "ok",
	})
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
}

func writeInternalError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
			"http_request",
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
