package exchangecorehttp

import (
	"net/http"
	"strings"

	"iwx/go_backend/internal/auth"
	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/requestctx"
)

func requireAuth(cfg config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing bearer token"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := auth.VerifyToken(token, cfg)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
			return
		}

		ctx := authcontext.WithClaims(r.Context(), claims)
		ctx = requestctx.WithUserID(ctx, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}
