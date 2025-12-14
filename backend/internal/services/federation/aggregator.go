package federation

import (
	"context"
	"fmt"
	"sync"

	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"
	"home-run-backend/internal/models"

	"github.com/sirupsen/logrus"
)

// ServiceProvider is an interface for getting local services
type ServiceProvider interface {
	GetAll(ctx context.Context) []models.Service
}

// Aggregator aggregates services from local and remote hosts
type Aggregator struct {
	localProvider ServiceProvider
	remoteHosts   []config.RemoteHost
}

// NewAggregator creates a new service aggregator
func NewAggregator(localProvider ServiceProvider, remoteHosts []config.RemoteHost) *Aggregator {
	return &Aggregator{
		localProvider: localProvider,
		remoteHosts:   remoteHosts,
	}
}

// GetAllServices returns all services from local and remote hosts
func (a *Aggregator) GetAllServices(ctx context.Context) []models.Service {
	// Start with local services
	localServices := a.localProvider.GetAll(ctx)
	result := make([]models.Service, len(localServices))
	copy(result, localServices)

	// Tag local services
	for i := range result {
		result[i].Host = "local"
	}

	// If no remote hosts, return local only
	if len(a.remoteHosts) == 0 {
		return result
	}

	// Fetch from remote hosts concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	logger.WithField("remote_hosts", len(a.remoteHosts)).Debug("Fetching services from remote hosts")

	for _, host := range a.remoteHosts {
		wg.Add(1)
		go func(h config.RemoteHost) {
			defer wg.Done()

			client := NewClient(h)
			resp, err := client.FetchServices(ctx)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"host":  h.Name,
					"error": err.Error(),
				}).Warn("Failed to fetch services from remote host")
				return
			}

			// Tag with host name and ensure unique IDs
			mu.Lock()
			for _, svc := range resp.Services {
				svc.Host = h.Name
				svc.ID = fmt.Sprintf("%s-%s", h.Name, svc.ID)
				result = append(result, svc)
			}
			mu.Unlock()
		}(host)
	}

	wg.Wait()
	logger.WithField("total_services", len(result)).Debug("Aggregated all services")
	return result
}

// GetLocalServices returns only local services (for federation endpoint)
func (a *Aggregator) GetLocalServices(ctx context.Context) []models.Service {
	services := a.localProvider.GetAll(ctx)
	for i := range services {
		services[i].Host = "local"
	}
	return services
}
