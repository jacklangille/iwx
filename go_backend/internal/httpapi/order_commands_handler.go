package httpapi

import (
	"net/http"
	"strings"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/pkg/logging"
)

func (s *Server) handleOrderCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleOrderCommandShow)(w, r)
}

func (s *Server) handleOrderCommandShow(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	commandID := strings.TrimPrefix(r.URL.Path, "/api/order_commands/")
	commandID = strings.Trim(commandID, "/")
	if commandID == "" {
		http.NotFound(w, r)
		return
	}

	command, err := s.reads.GetOrderCommand(r.Context(), commandID)
	if err != nil {
		logging.Error(r.Context(), "api_order_command_lookup_failed", err, "command_id", commandID)
		writeInternalError(w, err)
		return
	}
	if command == nil {
		logging.Info(r.Context(), "api_order_command_not_found", "command_id", commandID)
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "order command not found"})
		return
	}
	if command.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
		return
	}

	logging.Info(r.Context(), "api_order_command_fetched", "command_id", command.CommandID, "status", command.Status)
	writeJSON(w, http.StatusOK, serializeOrderCommand(*command))
}
