package kuma

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"

	"github.com/sirupsen/logrus"
)

// MonitorStatus holds the status of an Uptime Kuma monitor
type MonitorStatus struct {
	ID      int
	Name    string
	Status  string  // RUNNING, STOPPED, ERROR, MAINTENANCE
	Uptime  float64 // percentage (0-100)
	Latency float64 // ms
}

// Client is the Uptime Kuma API client
type Client struct {
	baseURL    string
	username   string
	password   string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Uptime Kuma client
func NewClient(cfg *config.UptimeKumaConfig) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(cfg.URL, "/"),
		username: cfg.Username,
		password: cfg.Password,
		apiKey:   cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMonitorStatus fetches the status of a specific monitor from the /metrics endpoint
func (c *Client) GetMonitorStatus(ctx context.Context, monitorID int) (*MonitorStatus, error) {
	url := fmt.Sprintf("%s/metrics", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"monitor_id": monitorID,
			"error":      err.Error(),
		}).Error("Failed to create Kuma request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication
	if c.apiKey != "" {
		req.SetBasicAuth("", c.apiKey)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	logger.WithField("monitor_id", monitorID).Debug("Fetching monitor status from Uptime Kuma")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"monitor_id": monitorID,
			"error":      err.Error(),
		}).Warn("Failed to fetch Kuma metrics")
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(logrus.Fields{
			"monitor_id": monitorID,
			"status":     resp.StatusCode,
		}).Warn("Kuma returned non-200 status")
		return nil, fmt.Errorf("kuma returned status %d", resp.StatusCode)
	}

	status, err := c.parseMetrics(resp.Body, monitorID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"monitor_id": monitorID,
			"error":      err.Error(),
		}).Error("Failed to parse Kuma metrics")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"monitor_id": monitorID,
		"status":     status.Status,
	}).Debug("Retrieved monitor status from Kuma")

	return status, nil
}

// parseMetrics parses the Prometheus format metrics response
func (c *Client) parseMetrics(body io.Reader, targetID int) (*MonitorStatus, error) {
	scanner := bufio.NewScanner(body)
	result := &MonitorStatus{ID: targetID}
	found := false

	// Regex to extract monitor_id from metric lines
	// Example: monitor_status{monitor_name="Nginx",monitor_type="http",monitor_url="https://example.com",monitor_hostname="null",monitor_port="null"} 1
	statusRegex := regexp.MustCompile(`monitor_status\{[^}]*monitor_name="([^"]*)"[^}]*\}\s+(\d+)`)
	responseTimeRegex := regexp.MustCompile(`monitor_response_time\{[^}]*\}\s+([\d.]+)`)

	// Uptime Kuma uses a different format - need to match by order or specific attributes
	// Let's use a simpler approach: look for lines containing the monitor ID
	monitorIDStr := strconv.Itoa(targetID)
	idRegex := regexp.MustCompile(fmt.Sprintf(`monitor_\w+\{[^}]*monitor_id="%s"[^}]*\}`, monitorIDStr))

	var statusValue int = -1

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Check if this line is for our monitor
		if idRegex.MatchString(line) {
			found = true

			// Extract monitor status
			if strings.HasPrefix(line, "monitor_status{") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if val, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
						statusValue = val
					}
				}
				// Extract name from the line
				if matches := statusRegex.FindStringSubmatch(line); len(matches) > 1 {
					result.Name = matches[1]
				}
			}

			// Extract response time
			if strings.HasPrefix(line, "monitor_response_time{") {
				if matches := responseTimeRegex.FindStringSubmatch(line); len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						result.Latency = val
					}
				}
			}

		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading metrics: %w", err)
	}

	if !found {
		return nil, fmt.Errorf("monitor ID %d not found in metrics", targetID)
	}

	// Map status value to our status enum
	result.Status = mapKumaStatus(statusValue)

	return result, nil
}

// mapKumaStatus converts Uptime Kuma status values to our status enum
// Uptime Kuma status values:
// 0 = DOWN
// 1 = UP
// 2 = PENDING
// 3 = MAINTENANCE
func mapKumaStatus(value int) string {
	switch value {
	case 1:
		return "RUNNING" // UP
	case 0:
		return "STOPPED" // DOWN
	case 2:
		return "ERROR" // PENDING
	case 3:
		return "MAINTENANCE"
	default:
		return "ERROR"
	}
}

// Ping tests connectivity to the Uptime Kuma instance
func (c *Client) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/metrics", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.SetBasicAuth("", c.apiKey)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("uptime kuma returned status %d", resp.StatusCode)
	}

	return nil
}
