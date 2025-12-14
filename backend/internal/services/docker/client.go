package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"home-run-backend/internal/logger"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// ContainerInfo holds container status information
type ContainerInfo struct {
	ID        string
	Status    string // RUNNING, STOPPED, ERROR, MAINTENANCE
	State     string // raw docker state
	StartedAt time.Time
}

// Stats holds container resource usage
type Stats struct {
	CPUPercent float64
	MemoryMB   float64
}

// Client wraps the Docker API client
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.WithField("error", err.Error()).Warn("Failed to create Docker client")
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx)
	if err != nil {
		logger.WithField("error", err.Error()).Warn("Failed to connect to Docker daemon")
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	logger.Log.Info("Docker client initialized successfully")
	return &Client{cli: cli}, nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.cli.Close()
}

// GetContainerInfo retrieves status information for a container by name
func (c *Client) GetContainerInfo(ctx context.Context, containerName string) (*ContainerInfo, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"container": containerName,
			"error":     err.Error(),
		}).Error("Failed to list containers")
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Normalize container name (Docker prepends /)
	searchName := containerName
	if !strings.HasPrefix(searchName, "/") {
		searchName = "/" + containerName
	}

	for _, ctr := range containers {
		for _, name := range ctr.Names {
			if name == searchName || name == containerName {
				info := &ContainerInfo{
					ID:     ctr.ID[:12],
					Status: mapDockerState(ctr.State),
					State:  ctr.State,
				}

				// Get detailed inspection for uptime
				inspect, err := c.cli.ContainerInspect(ctx, ctr.ID)
				if err == nil && inspect.State != nil && inspect.State.StartedAt != "" {
					startedAt, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
					if err == nil {
						info.StartedAt = startedAt
					}
				}

				logger.WithFields(logrus.Fields{
					"container": containerName,
					"id":        info.ID,
					"status":    info.Status,
				}).Debug("Retrieved container info")
				return info, nil
			}
		}
	}

	logger.WithField("container", containerName).Warn("Container not found")
	return nil, fmt.Errorf("container '%s' not found", containerName)
}

// GetContainerStats retrieves resource usage for a container
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*Stats, error) {
	statsResp, err := c.cli.ContainerStats(ctx, containerID, false) // false = one-shot
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsResp.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &Stats{
		CPUPercent: calculateCPUPercent(&stats),
		MemoryMB:   float64(stats.MemoryStats.Usage) / (1024 * 1024),
	}, nil
}

// calculateCPUPercent calculates CPU usage percentage
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuCount := float64(stats.CPUStats.OnlineCPUs)
		if cpuCount == 0 {
			cpuCount = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		if cpuCount > 0 {
			return (cpuDelta / systemDelta) * cpuCount * 100.0
		}
	}
	return 0
}

// mapDockerState converts Docker state to our status enum
func mapDockerState(state string) string {
	switch strings.ToLower(state) {
	case "running":
		return "RUNNING"
	case "exited", "dead":
		return "STOPPED"
	case "paused":
		return "MAINTENANCE"
	case "restarting":
		return "MAINTENANCE"
	case "created":
		return "STOPPED"
	default:
		return "ERROR"
	}
}
