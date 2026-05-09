package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"iwx/go_backend/internal/readmodel"
)

func (s *Server) handleContractSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/contracts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	contractID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
		return
	}

	switch parts[1] {
	case "market_state":
		s.handleMarketState(w, r, contractID)
	case "market_snapshots":
		s.handleMarketSnapshots(w, r, contractID)
	case "executions":
		s.handleExecutions(w, r, contractID)
	case "observations":
		s.handleObservations(w, r, contractID)
	case "resolution":
		s.handleResolution(w, r, contractID)
	case "settlements":
		s.handleSettlements(w, r, contractID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleMarketState(w http.ResponseWriter, r *http.Request, contractID int64) {
	marketState, err := s.reads.MarketState(r.Context(), contractID)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}

		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeMarketState(marketState))
}

func (s *Server) handleMarketSnapshots(w http.ResponseWriter, r *http.Request, contractID int64) {
	config := chartConfigFromRequest(r)
	points, err := s.reads.ListMarketChartSeries(r.Context(), contractID, config)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}

		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"config": map[string]any{
			"lookback_seconds": config.LookbackSeconds,
			"bucket_seconds":   config.BucketSeconds,
		},
		"points": serializeChartPoints(points),
	})
}

func (s *Server) handleExecutions(w http.ResponseWriter, r *http.Request, contractID int64) {
	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	executions, err := s.reads.ListExecutions(r.Context(), contractID, limit)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}

		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"contract_id": contractID,
		"limit":       limit,
		"executions":  serializeExecutions(executions),
	})
}

func (s *Server) handleObservations(w http.ResponseWriter, r *http.Request, contractID int64) {
	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	observations, err := s.reads.ListObservations(r.Context(), contractID, limit)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"contract_id":  contractID,
		"observations": serializeOracleObservations(observations),
	})
}

func (s *Server) handleResolution(w http.ResponseWriter, r *http.Request, contractID int64) {
	resolution, err := s.reads.GetResolution(r.Context(), contractID)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		writeInternalError(w, err)
		return
	}
	if resolution == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resolution not found"})
		return
	}

	writeJSON(w, http.StatusOK, serializeContractResolution(*resolution))
}

func (s *Server) handleSettlements(w http.ResponseWriter, r *http.Request, contractID int64) {
	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	entries, err := s.reads.ListSettlementEntriesByContract(r.Context(), contractID, limit)
	if err != nil {
		if errors.Is(err, readmodel.ErrContractNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"contract_id": contractID,
		"entries":     serializeSettlementEntries(entries),
	})
}
