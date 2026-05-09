package oraclestationhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/projectionbundle"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) FindStation(ctx context.Context, providerName, stationID string) (*domain.WeatherStation, error) {
	if c == nil || c.baseURL == "" {
		return nil, fmt.Errorf("oracle station client not configured")
	}

	endpoint := c.baseURL + "/internal/stations/" + url.PathEscape(strings.TrimSpace(providerName)) + "/" + url.PathEscape(strings.TrimSpace(stationID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oracle station lookup failed: status %d", resp.StatusCode)
	}

	var payload struct {
		ID               int64    `json:"id"`
		ProviderName     string   `json:"provider_name"`
		StationID        string   `json:"station_id"`
		DisplayName      string   `json:"display_name"`
		Region           string   `json:"region"`
		Latitude         *float64 `json:"latitude"`
		Longitude        *float64 `json:"longitude"`
		SupportedMetrics []string `json:"supported_metrics"`
		Active           bool     `json:"active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return &domain.WeatherStation{
		ID:               payload.ID,
		ProviderName:     payload.ProviderName,
		StationID:        payload.StationID,
		DisplayName:      payload.DisplayName,
		Region:           payload.Region,
		Latitude:         payload.Latitude,
		Longitude:        payload.Longitude,
		SupportedMetrics: payload.SupportedMetrics,
		Active:           payload.Active,
	}, nil
}

func (c *Client) GetOracleStateBundle(ctx context.Context, contractID int64) (*projectionbundle.OracleStateBundle, error) {
	endpoint := c.baseURL + "/internal/projection/contracts/" + url.PathEscape(strings.TrimSpace(fmt.Sprintf("%d", contractID))) + "/oracle"
	var bundle projectionbundle.OracleStateBundle
	status, err := c.getJSON(ctx, endpoint, &bundle)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &bundle, nil
}

func (c *Client) GetStationCatalogBundle(ctx context.Context) (*projectionbundle.StationCatalogBundle, error) {
	var bundle projectionbundle.StationCatalogBundle
	_, err := c.getJSON(ctx, c.baseURL+"/internal/projection/stations", &bundle)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) (int, error) {
	if c == nil || c.baseURL == "" {
		return 0, fmt.Errorf("oracle station client not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return http.StatusNotFound, nil
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("oracle request failed: status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}
