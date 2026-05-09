package authhttp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"iwx/go_backend/internal/auth"
	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

type Server struct {
	config config.Config
	auth   *auth.Service
	mux    *http.ServeMux
}

func NewServer(cfg config.Config, authService *auth.Service) *Server {
	server := &Server{
		config: cfg,
		auth:   authService,
		mux:    http.NewServeMux(),
	}

	server.registerRoutes()

	return server
}

func (s *Server) ListenAndServe(_ context.Context) error {
	httpServer := &http.Server{
		Addr:    s.config.AuthHTTPAddr,
		Handler: s.loggingMiddleware(s.mux),
	}

	log.Printf("auth server listening addr=%s", s.config.AuthHTTPAddr)
	return httpServer.ListenAndServe()
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleRoute)
}

func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.handleIndex(w, r)
	case "/api/auth/login":
		s.handleAuthLogin(w, r)
	case "/api/auth/signup":
		s.handleAuthSignup(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "iwx-go-auth",
		"status":  "ok",
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}

	if s.auth == nil || !s.auth.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "auth service is not configured"})
		return
	}

	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}

	if request.Username == "" || request.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "username and password are required"})
		return
	}

	token, claims, err := s.auth.Authenticate(r.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			logging.Info(r.Context(), "auth_login_rejected", "username", request.Username)
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
			return
		}

		if errors.Is(err, auth.ErrNotConfigured) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "auth service is not configured"})
			return
		}

		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   claims.ExpiresAt,
		"subject":      claims.Subject,
		"user_id":      claims.UserID,
	})
}

func (s *Server) handleAuthSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}

	if s.auth == nil || !s.auth.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "auth service is not configured"})
		return
	}

	var request signupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}

	if request.Username == "" || request.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "username and password are required"})
		return
	}

	token, claims, err := s.auth.Signup(r.Context(), request.Username, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUsernameTaken):
			logging.Info(r.Context(), "auth_signup_rejected", "username", request.Username, "reason", "username_taken")
			writeJSON(w, http.StatusConflict, map[string]any{"error": "username is already taken"})
			return
		case errors.Is(err, auth.ErrInvalidSignup):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "username must be at least 3 characters and password at least 8 characters"})
			return
		case errors.Is(err, auth.ErrNotConfigured):
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "auth service is not configured"})
			return
		default:
			writeInternalError(w, err)
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   claims.ExpiresAt,
		"subject":      claims.Subject,
		"user_id":      claims.UserID,
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
			"auth_http_request",
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
