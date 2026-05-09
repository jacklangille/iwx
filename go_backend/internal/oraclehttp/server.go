package oraclehttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/oracle"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

type Server struct {
	config  config.Config
	service *oracle.Service
	mux     *http.ServeMux
}

func NewServer(cfg config.Config, service *oracle.Service) *Server {
	s := &Server{
		config:  cfg,
		service: service,
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) ListenAndServe() error {
	server := &http.Server{
		Addr:    s.config.OracleHTTPAddr,
		Handler: s.loggingMiddleware(s.mux),
	}

	log.Printf("oracle http server listening addr=%s", s.config.OracleHTTPAddr)
	return server.ListenAndServe()
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/oracle/stations", s.handleStations)
	s.mux.HandleFunc("/api/oracle/observations", s.handleObservations)
	s.mux.HandleFunc("/api/oracle/contracts/", s.handleContractSubroutes)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": "iwx-oracle", "status": "ok"})
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeInternalError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
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
			"oracle_http_request",
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
