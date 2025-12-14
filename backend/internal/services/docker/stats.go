package docker

import (
	"context"
	"sync"
	"time"

	"home-run-backend/internal/cache"
	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"

	"github.com/sirupsen/logrus"
)

// CachedStats holds cached container stats
type CachedStats struct {
	Status     string
	StartedAt  time.Time
	CPUPercent float64
	MemoryMB   float64
	LastUpdate time.Time
}

// StatsCollector collects Docker container stats periodically
type StatsCollector struct {
	client   *Client
	cache    *cache.Cache
	services []config.ServiceConfig
	interval time.Duration
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector(dockerClient *Client, statsCache *cache.Cache, services []config.ServiceConfig, interval time.Duration) *StatsCollector {
	return &StatsCollector{
		client:   dockerClient,
		cache:    statsCache,
		services: services,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins collecting stats in the background
func (sc *StatsCollector) Start(ctx context.Context) {
	sc.mu.Lock()
	if sc.running {
		sc.mu.Unlock()
		return
	}
	sc.running = true
	sc.mu.Unlock()

	logger.WithField("interval", sc.interval).Info("Starting Docker stats collector")

	// Collect immediately on start
	sc.collectAll(ctx)

	go func() {
		ticker := time.NewTicker(sc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("Docker stats collector stopped (context cancelled)")
				return
			case <-sc.stopCh:
				logger.Log.Info("Docker stats collector stopped")
				return
			case <-ticker.C:
				sc.collectAll(ctx)
			}
		}
	}()
}

// Stop stops the stats collector
func (sc *StatsCollector) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.running {
		close(sc.stopCh)
		sc.running = false
	}
}

// collectAll collects stats for all Docker services
func (sc *StatsCollector) collectAll(ctx context.Context) {
	logger.Log.Debug("Collecting Docker stats for all containers")
	for _, svc := range sc.services {
		if svc.Backend != "docker" {
			continue
		}

		info, err := sc.client.GetContainerInfo(ctx, svc.ContainerName)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"container": svc.ContainerName,
				"error":     err.Error(),
			}).Warn("Failed to get container info")
			sc.cache.Set(svc.ContainerName, &CachedStats{
				Status:     "ERROR",
				LastUpdate: time.Now(),
			})
			continue
		}

		cached := &CachedStats{
			Status:     info.Status,
			StartedAt:  info.StartedAt,
			LastUpdate: time.Now(),
		}

		// Only get stats if container is running
		if info.Status == "RUNNING" {
			stats, err := sc.client.GetContainerStats(ctx, info.ID)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"container": svc.ContainerName,
					"error":     err.Error(),
				}).Warn("Failed to get container stats")
			} else {
				cached.CPUPercent = stats.CPUPercent
				cached.MemoryMB = stats.MemoryMB
				logger.WithFields(logrus.Fields{
					"container": svc.ContainerName,
					"cpu":       stats.CPUPercent,
					"memory_mb": stats.MemoryMB,
				}).Debug("Collected container stats")
			}
		}

		sc.cache.Set(svc.ContainerName, cached)
	}
}

// GetCachedStats retrieves cached stats for a container
func (sc *StatsCollector) GetCachedStats(containerName string) *CachedStats {
	if val, ok := sc.cache.Get(containerName); ok {
		if stats, ok := val.(*CachedStats); ok {
			return stats
		}
	}
	return nil
}
