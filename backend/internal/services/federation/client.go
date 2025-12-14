package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"
	"home-run-backend/internal/models"

	"github.com/sirupsen/logrus"
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
		logger.WithFields(logrus.Fields{
			"host":  c.name,
			"error": err.Error(),
		}).Error("Failed to create federation request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	logger.WithFields(logrus.Fields{
		"host": c.name,
		"url":  url,
	}).Debug("Fetching services from remote host")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"host":  c.name,
			"error": err.Error(),
		}).Warn("Federation request failed")
		return nil, fmt.Errorf("federation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(logrus.Fields{
			"host":   c.name,
			"status": resp.StatusCode,
		}).Warn("Remote host returned non-200 status")
		return nil, fmt.Errorf("remote host returned status %d", resp.StatusCode)
	}

	var result FederationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.WithFields(logrus.Fields{
			"host":  c.name,
			"error": err.Error(),
		}).Error("Failed to decode federation response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":  c.name,
		"count": len(result.Services),
	}).Info("Successfully fetched services from remote host")

	return &result, nil
}

// Name returns the name of the remote host
func (c *Client) Name() string {
	return c.name
}
