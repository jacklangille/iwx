package matcherhttpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (c *Client) GetMarketBundle(ctx context.Context, contractID int64) (*projectionbundle.MarketBundle, error) {
	var bundle projectionbundle.MarketBundle
	_, err := c.getJSON(ctx, c.baseURL+"/internal/projection/contracts/"+strconv.FormatInt(contractID, 10)+"/market", &bundle)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (c *Client) GetOrderCommandBundle(ctx context.Context, commandID string) (*projectionbundle.OrderCommandBundle, error) {
	var bundle projectionbundle.OrderCommandBundle
	status, err := c.getJSON(ctx, c.baseURL+"/internal/projection/order_commands/"+commandID, &bundle)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &bundle, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) (int, error) {
	if c == nil || c.baseURL == "" {
		return 0, fmt.Errorf("matcher client not configured")
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
		return resp.StatusCode, fmt.Errorf("matcher request failed: status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}
