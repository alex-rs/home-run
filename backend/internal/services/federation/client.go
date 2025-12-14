package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"home-run-backend/internal/config"
	"home-run-backend/internal/models"
)

// FederationResponse is the response from a remote host's federation endpoint
type FederationResponse struct {
	Services []models.Service `json:"services"`
	Host     string           `json:"host"`
}

// Client fetches services from a remote host
type Client struct {
	name       string
	endpoint   string
	token      string
	httpClient *http.Client
}

// NewClient creates a new federation client
func NewClient(host config.RemoteHost) *Client {
	return &Client{
		name:     host.Name,
		endpoint: strings.TrimSuffix(host.Endpoint, "/"),
		token:    host.Token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FetchServices retrieves services from the remote host
func (c *Client) FetchServices(ctx context.Context) (*FederationResponse, error) {
	url := fmt.Sprintf("%s/federation/services", c.endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("federation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote host returned status %d", resp.StatusCode)
	}

	var result FederationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Name returns the name of the remote host
func (c *Client) Name() string {
	return c.name
}
