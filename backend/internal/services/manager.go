package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"home-run-backend/internal/cache"
	"home-run-backend/internal/config"
	"home-run-backend/internal/models"
	"home-run-backend/internal/services/docker"
	"home-run-backend/internal/services/kuma"
)

// Manager manages local services and their status
type Manager struct {
	cfg            *config.Config
	dockerClient   *docker.Client
	dockerStats    *docker.StatsCollector
	kumaClient     *kuma.Client
	statsCache     *cache.Cache
	dockerDisabled bool
}

// NewManager creates a new service manager
func NewManager(cfg *config.Config) (*Manager, error) {
	m := &Manager{
		cfg:        cfg,
		statsCache: cache.New(30 * time.Second),
	}

	// Initialize Docker client (optional - may not be available)
	dockerClient, err := docker.NewClient()
	if err != nil {
		// Docker not available, disable Docker backend
		m.dockerDisabled = true
	} else {
		m.dockerClient = dockerClient

		// Initialize stats collector
		m.dockerStats = docker.NewStatsCollector(dockerClient, m.statsCache, cfg.Services, 10*time.Second)
	}

	// Initialize Uptime Kuma client if configured
	if cfg.UptimeKuma != nil {
		m.kumaClient = kuma.NewClient(cfg.UptimeKuma)
	}

	return m, nil
}

// Start starts background processes (stats collection)
func (m *Manager) Start(ctx context.Context) {
	if m.dockerStats != nil {
		m.dockerStats.Start(ctx)
	}
}

// Stop stops background processes
func (m *Manager) Stop() {
	if m.dockerStats != nil {
		m.dockerStats.Stop()
	}
	if m.dockerClient != nil {
		m.dockerClient.Close()
	}
}

// GetAll returns all configured services with their current status
func (m *Manager) GetAll(ctx context.Context) []models.Service {
	var result []models.Service

	for _, svcCfg := range m.cfg.Services {
		svc := m.buildService(ctx, svcCfg)
		result = append(result, svc)
	}

	return result
}

// GetByID returns a single service by ID
func (m *Manager) GetByID(ctx context.Context, id string) (*models.Service, error) {
	for _, svcCfg := range m.cfg.Services {
		if generateID(svcCfg.Name) == id {
			svc := m.buildService(ctx, svcCfg)
			return &svc, nil
		}
	}
	return nil, fmt.Errorf("service not found: %s", id)
}

// GetConfigContent returns the content of a service's config file
func (m *Manager) GetConfigContent(ctx context.Context, serviceID string, configIndex int) (*models.ServiceConfig, error) {
	for _, svcCfg := range m.cfg.Services {
		if generateID(svcCfg.Name) != serviceID {
			continue
		}

		if configIndex < 0 || configIndex >= len(svcCfg.Configs) {
			return nil, fmt.Errorf("config index out of range")
		}

		configPath := svcCfg.Configs[configIndex]
		content, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		return &models.ServiceConfig{
			Type:       detectConfigType(configPath),
			Path:       configPath,
			Content:    string(content),
			LastEdited: getFileModTime(configPath),
		}, nil
	}

	return nil, fmt.Errorf("service not found: %s", serviceID)
}

// buildService constructs a Service model from config and live data
func (m *Manager) buildService(ctx context.Context, cfg config.ServiceConfig) models.Service {
	svc := models.Service{
		ID:   generateID(cfg.Name),
		Name: cfg.Name,
		URL:  cfg.URL,
		Port: cfg.Port,
	}

	// Build configs list (without content - lazy loaded)
	for _, path := range cfg.Configs {
		svc.Configs = append(svc.Configs, models.ServiceConfig{
			Path:       path,
			Type:       detectConfigType(path),
			LastEdited: getFileModTime(path),
		})
	}

	// Get status based on backend
	switch cfg.Backend {
	case "docker":
		m.populateDockerStatus(ctx, &svc, cfg.ContainerName)
	case "uptime_kuma":
		m.populateKumaStatus(ctx, &svc, cfg.KumaMonitorID)
	default:
		svc.Status = "ERROR"
	}

	return svc
}

// populateDockerStatus fills in status from Docker
func (m *Manager) populateDockerStatus(ctx context.Context, svc *models.Service, containerName string) {
	if m.dockerDisabled || m.dockerClient == nil {
		svc.Status = "ERROR"
		return
	}

	// Try cache first
	if cached := m.dockerStats.GetCachedStats(containerName); cached != nil {
		svc.Status = cached.Status
		svc.CPUUsage = cached.CPUPercent
		svc.MemoryUsage = cached.MemoryMB
		svc.Uptime = formatUptime(cached.StartedAt)
		return
	}

	// Fallback to live query
	info, err := m.dockerClient.GetContainerInfo(ctx, containerName)
	if err != nil {
		svc.Status = "ERROR"
		return
	}

	svc.Status = info.Status
	svc.Uptime = formatUptime(info.StartedAt)

	// Try to get stats
	if info.Status == "RUNNING" {
		if stats, err := m.dockerClient.GetContainerStats(ctx, info.ID); err == nil {
			svc.CPUUsage = stats.CPUPercent
			svc.MemoryUsage = stats.MemoryMB
		}
	}
}

// populateKumaStatus fills in status from Uptime Kuma
func (m *Manager) populateKumaStatus(ctx context.Context, svc *models.Service, monitorID int) {
	if m.kumaClient == nil {
		svc.Status = "ERROR"
		return
	}

	status, err := m.kumaClient.GetMonitorStatus(ctx, monitorID)
	if err != nil {
		svc.Status = "ERROR"
		return
	}

	svc.Status = status.Status
	if status.Uptime > 0 {
		svc.Uptime = fmt.Sprintf("%.1f%% uptime", status.Uptime)
	}
}

// generateID creates a unique ID from a service name
func generateID(name string) string {
	hash := md5.Sum([]byte(name))
	return fmt.Sprintf("%x", hash[:6])
}

// detectConfigType determines the config file type from extension
func detectConfigType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return "YAML"
	case ".json":
		return "JSON"
	case ".ini", ".conf", ".cfg":
		return "INI"
	case ".dockerfile":
		return "DOCKERFILE"
	default:
		base := strings.ToLower(filepath.Base(path))
		if strings.Contains(base, "dockerfile") {
			return "DOCKERFILE"
		}
		return "YAML" // Default
	}
}

// getFileModTime returns the modification time of a file
func getFileModTime(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "Unknown"
	}
	return info.ModTime().Format("2006-01-02 15:04")
}

// formatUptime formats a start time as a human-readable uptime string
func formatUptime(startedAt time.Time) string {
	if startedAt.IsZero() {
		return "Unknown"
	}

	duration := time.Since(startedAt)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
