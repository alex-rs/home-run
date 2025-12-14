package docker

import (
	"context"
	"log"
	"sync"
	"time"

	"home-run-backend/internal/cache"
	"home-run-backend/internal/config"
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

	// Collect immediately on start
	sc.collectAll(ctx)

	go func() {
		ticker := time.NewTicker(sc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sc.stopCh:
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
	for _, svc := range sc.services {
		if svc.Backend != "docker" {
			continue
		}

		info, err := sc.client.GetContainerInfo(ctx, svc.ContainerName)
		if err != nil {
			log.Printf("Failed to get container info for %s: %v", svc.ContainerName, err)
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
				log.Printf("Failed to get stats for %s: %v", svc.ContainerName, err)
			} else {
				cached.CPUPercent = stats.CPUPercent
				cached.MemoryMB = stats.MemoryMB
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
