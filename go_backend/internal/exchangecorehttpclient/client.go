package exchangecorehttpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/projectionbundle"
	"iwx/go_backend/internal/store"
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

func (c *Client) GetContract(ctx context.Context, contractID int64) (*domain.Contract, error) {
	payload, err := c.getContractPayload(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return payload.contract.toDomain(), nil
}

func (c *Client) GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error) {
	payload, err := c.getContractPayload(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return payload.rule.toDomain(), nil
}

func (c *Client) UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error) {
	if strings.EqualFold(strings.TrimSpace(status), string(domain.ContractStateResolved)) {
		return c.markContractResolved(ctx, contractID)
	}
	return nil, fmt.Errorf("unsupported remote contract status update: %s", status)
}

func (c *Client) SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error) {
	if c == nil || c.baseURL == "" {
		return nil, fmt.Errorf("exchange-core client not configured")
	}

	body, err := json.Marshal(map[string]any{
		"event_id":       input.EventID,
		"outcome":        input.Outcome,
		"resolved_at":    input.ResolvedAt,
		"correlation_id": input.CorrelationID,
	})
	if err != nil {
		return nil, err
	}

	endpoint := c.baseURL + "/internal/contracts/" + strconv.FormatInt(input.ContractID, 10) + "/settlement"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange-core settlement failed: status %d", resp.StatusCode)
	}

	var payload struct {
		Contract      *contractPayload `json:"contract"`
		AffectedUsers []int64          `json:"affected_users"`
		SettledAt     string           `json:"settled_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var settledAt time.Time
	if strings.TrimSpace(payload.SettledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, payload.SettledAt)
		if err == nil {
			settledAt = parsed.UTC()
		}
	}

	return &store.SettlementResult{
		Contract:      payload.Contract.toDomain(),
		AffectedUsers: payload.AffectedUsers,
		SettledAt:     settledAt,
	}, nil
}

func (c *Client) GetContractBundle(ctx context.Context, contractID int64) (*projectionbundle.ContractBundle, error) {
	var bundle projectionbundle.ContractBundle
	status, err := c.getJSON(ctx, c.baseURL+"/internal/projection/contracts/"+strconv.FormatInt(contractID, 10), &bundle)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &bundle, nil
}

func (c *Client) GetUserStateBundle(ctx context.Context, userID int64) (*projectionbundle.UserStateBundle, error) {
	var bundle projectionbundle.UserStateBundle
	_, err := c.getJSON(ctx, c.baseURL+"/internal/projection/users/"+strconv.FormatInt(userID, 10), &bundle)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (c *Client) GetSettlementBundle(ctx context.Context, contractID int64) (*projectionbundle.SettlementBundle, error) {
	var bundle projectionbundle.SettlementBundle
	_, err := c.getJSON(ctx, c.baseURL+"/internal/projection/settlements/contracts/"+strconv.FormatInt(contractID, 10), &bundle)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

type contractEnvelope struct {
	contract *contractPayload     `json:"contract"`
	rule     *contractRulePayload `json:"rule"`
}

func (c *Client) getContractPayload(ctx context.Context, contractID int64) (*contractEnvelope, error) {
	var payload contractEnvelope
	status, err := c.getJSON(ctx, c.baseURL+"/internal/contracts/"+strconv.FormatInt(contractID, 10), &payload)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return &contractEnvelope{}, nil
	}
	return &payload, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) (int, error) {
	if c == nil || c.baseURL == "" {
		return 0, fmt.Errorf("exchange-core client not configured")
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
		return resp.StatusCode, fmt.Errorf("exchange-core request failed: status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

func (c *Client) markContractResolved(ctx context.Context, contractID int64) (*domain.Contract, error) {
	if c == nil || c.baseURL == "" {
		return nil, fmt.Errorf("exchange-core client not configured")
	}

	endpoint := c.baseURL + "/internal/contracts/" + strconv.FormatInt(contractID, 10) + "/resolve"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, http.NoBody)
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
		return nil, fmt.Errorf("exchange-core contract resolve failed: status %d", resp.StatusCode)
	}

	var payload contractPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

type contractPayload struct {
	ID                      int64   `json:"id"`
	CreatorUserID           *int64  `json:"creator_user_id"`
	Name                    string  `json:"name"`
	Region                  string  `json:"region"`
	Metric                  string  `json:"metric"`
	Status                  string  `json:"status"`
	Threshold               *int64  `json:"threshold"`
	Multiplier              *int64  `json:"multiplier"`
	MeasurementUnit         *string `json:"measurement_unit"`
	TradingPeriodStart      *string `json:"trading_period_start"`
	TradingPeriodEnd        *string `json:"trading_period_end"`
	MeasurementPeriodStart  *string `json:"measurement_period_start"`
	MeasurementPeriodEnd    *string `json:"measurement_period_end"`
	DataProviderName        *string `json:"data_provider_name"`
	StationID               *string `json:"station_id"`
	DataProviderStationMode *string `json:"data_provider_station_mode"`
	Description             *string `json:"description"`
}

func (p *contractPayload) toDomain() *domain.Contract {
	if p == nil {
		return nil
	}
	return &domain.Contract{
		ID:                      p.ID,
		CreatorUserID:           p.CreatorUserID,
		Name:                    p.Name,
		Region:                  p.Region,
		Metric:                  p.Metric,
		Status:                  p.Status,
		Threshold:               p.Threshold,
		Multiplier:              p.Multiplier,
		MeasurementUnit:         stringValue(p.MeasurementUnit),
		TradingPeriodStart:      parseDatePtr(p.TradingPeriodStart),
		TradingPeriodEnd:        parseDatePtr(p.TradingPeriodEnd),
		MeasurementPeriodStart:  parseDatePtr(p.MeasurementPeriodStart),
		MeasurementPeriodEnd:    parseDatePtr(p.MeasurementPeriodEnd),
		DataProviderName:        stringValue(p.DataProviderName),
		StationID:               stringValue(p.StationID),
		DataProviderStationMode: stringValue(p.DataProviderStationMode),
		Description:             stringValue(p.Description),
	}
}

type contractRulePayload struct {
	ID                      int64   `json:"id"`
	ContractID              int64   `json:"contract_id"`
	RuleVersion             string  `json:"rule_version"`
	Metric                  string  `json:"metric"`
	Threshold               *int64  `json:"threshold"`
	MeasurementUnit         *string `json:"measurement_unit"`
	ResolutionInclusiveSide string  `json:"resolution_inclusive_side"`
	CreatedAt               string  `json:"created_at"`
}

func (p *contractRulePayload) toDomain() *domain.ContractRule {
	if p == nil {
		return nil
	}
	rule := &domain.ContractRule{
		ID:                      p.ID,
		ContractID:              p.ContractID,
		RuleVersion:             p.RuleVersion,
		Metric:                  p.Metric,
		Threshold:               p.Threshold,
		MeasurementUnit:         stringValue(p.MeasurementUnit),
		ResolutionInclusiveSide: domain.ClaimSide(p.ResolutionInclusiveSide),
	}
	if strings.TrimSpace(p.CreatedAt) != "" {
		if parsed, err := time.Parse(time.RFC3339, p.CreatedAt); err == nil {
			rule.CreatedAt = parsed.UTC()
		}
	}
	return rule
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func parseDatePtr(value *string) *time.Time {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}
