package oraclehttp

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"iwx/go_backend/internal/httpjson"
	"iwx/go_backend/internal/oracle"
	"iwx/go_backend/internal/projectionbundle"
	"iwx/go_backend/internal/store"
	"iwx/go_backend/pkg/logging"
)

func (s *Server) handleStations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		activeOnly := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("active")), "true")
		stations, err := s.service.ListStations(r.Context(), activeOnly)
		if err != nil {
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, stationsResponse{Stations: serializeStations(stations)})
	case http.MethodPost:
		var request struct {
			ProviderName     string   `json:"provider_name"`
			StationID        string   `json:"station_id"`
			DisplayName      string   `json:"display_name"`
			Region           string   `json:"region"`
			Latitude         *float64 `json:"latitude"`
			Longitude        *float64 `json:"longitude"`
			SupportedMetrics []string `json:"supported_metrics"`
			Active           *bool    `json:"active"`
		}
		if err := httpjson.DecodeStrict(r.Body, &request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
			return
		}
		active := true
		if request.Active != nil {
			active = *request.Active
		}
		station, err := s.service.UpsertStation(r.Context(), store.UpsertStationInput{
			ProviderName:     request.ProviderName,
			StationID:        request.StationID,
			DisplayName:      request.DisplayName,
			Region:           request.Region,
			Latitude:         request.Latitude,
			Longitude:        request.Longitude,
			SupportedMetrics: request.SupportedMetrics,
			Active:           active,
		})
		if err != nil {
			var validationErr *oracle.ValidationError
			if errors.As(err, &validationErr) {
				writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
				return
			}
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, serializeStation(*station))
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleObservations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleObservationCreate(w, r)
	default:
		methodNotAllowed(w, "POST")
	}
}

func (s *Server) handleInternalStationLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/internal/stations/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	providerName, err := url.PathUnescape(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid provider name"})
		return
	}
	stationID, err := url.PathUnescape(parts[1])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid station id"})
		return
	}

	station, err := s.service.FindStation(r.Context(), providerName, stationID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if station == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "station not found"})
		return
	}

	writeJSON(w, http.StatusOK, serializeStation(*station))
}

func (s *Server) handleContractSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/oracle/contracts/")
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
	case "observations":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		s.handleObservationIndex(w, r, contractID)
	case "resolve":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, "POST")
			return
		}
		s.handleResolveContract(w, r, contractID)
	case "resolution":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		s.handleResolutionShow(w, r, contractID)
	default:
		http.NotFound(w, r)
	}
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
	case len(parts) == 1 && parts[0] == "stations":
		stations, err := s.service.ListStations(r.Context(), false)
		if err != nil {
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, projectionbundle.StationCatalogBundle{Stations: stations})
	case len(parts) == 3 && parts[0] == "contracts" && parts[2] == "oracle":
		contractID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid contract id"})
			return
		}
		observations, err := s.service.ListObservations(r.Context(), contractID, 500)
		if err != nil {
			writeInternalError(w, err)
			return
		}
		resolution, err := s.service.GetLatestResolution(r.Context(), contractID)
		if err != nil {
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, projectionbundle.OracleStateBundle{
			ContractID:   contractID,
			Observations: observations,
			Resolution:   resolution,
		})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleObservationCreate(w http.ResponseWriter, r *http.Request) {
	input, err := decodeObservationInput(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	observation, err := s.service.RecordObservation(r.Context(), input)
	if err != nil {
		var validationErr *oracle.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}

		writeInternalError(w, err)
		return
	}

	logging.Info(r.Context(), "oracle_observation_recorded", "contract_id", observation.ContractID, "observation_id", observation.ID, "provider", observation.ProviderName, "station", observation.StationID)
	writeJSON(w, http.StatusCreated, serializeObservation(*observation))
}

func (s *Server) handleObservationIndex(w http.ResponseWriter, r *http.Request, contractID int64) {
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

	observations, err := s.service.ListObservations(r.Context(), contractID, limit)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, observationsResponse{
		ContractID:   contractID,
		Observations: serializeObservations(observations),
	})
}

func (s *Server) handleResolveContract(w http.ResponseWriter, r *http.Request, contractID int64) {
	resolution, err := s.service.ResolveContract(r.Context(), contractID)
	if err != nil {
		var validationErr *oracle.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Errors})
			return
		}

		writeInternalError(w, err)
		return
	}

	logging.Info(r.Context(), "oracle_contract_resolved", "contract_id", resolution.ContractID, "resolution_id", resolution.ID, "outcome", resolution.Outcome)
	writeJSON(w, http.StatusOK, serializeResolution(*resolution))
}

func (s *Server) handleResolutionShow(w http.ResponseWriter, r *http.Request, contractID int64) {
	resolution, err := s.service.GetLatestResolution(r.Context(), contractID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if resolution == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resolution not found"})
		return
	}

	writeJSON(w, http.StatusOK, serializeResolution(*resolution))
}
