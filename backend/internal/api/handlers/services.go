package handlers

import (
	"net/http"
	"strconv"

	"home-run-backend/internal/logger"
	"home-run-backend/internal/services"
	"home-run-backend/internal/services/federation"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ServicesHandler struct {
	manager    *services.Manager
	aggregator *federation.Aggregator
}

func NewServicesHandler(manager *services.Manager, aggregator *federation.Aggregator) *ServicesHandler {
	return &ServicesHandler{
		manager:    manager,
		aggregator: aggregator,
	}
}

// List returns all services (local + federated)
func (h *ServicesHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	allServices := h.aggregator.GetAllServices(ctx)

	// Count running services
	running := 0
	for _, svc := range allServices {
		if svc.Status == "RUNNING" {
			running++
		}
	}

	logger.WithFields(logrus.Fields{
		"total":   len(allServices),
		"running": running,
	}).Debug("Listed all services")

	c.JSON(http.StatusOK, gin.H{
		"services": allServices,
		"total":    len(allServices),
		"running":  running,
	})
}

// Get returns a single service by ID
func (h *ServicesHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	svc, err := h.manager.GetByID(ctx, id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service_id": id,
			"error":      err.Error(),
		}).Warn("Service not found")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	logger.WithField("service_id", id).Debug("Retrieved service details")
	c.JSON(http.StatusOK, svc)
}

// GetConfig returns the content of a service's config file
func (h *ServicesHandler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()
	serviceID := c.Param("id")
	indexStr := c.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service_id": serviceID,
			"index":      indexStr,
		}).Warn("Invalid config index")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid config index",
		})
		return
	}

	config, err := h.manager.GetConfigContent(ctx, serviceID, index)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service_id": serviceID,
			"index":      index,
			"error":      err.Error(),
		}).Warn("Failed to get config content")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"service_id": serviceID,
		"index":      index,
		"path":       config.Path,
	}).Debug("Retrieved config file")

	c.JSON(http.StatusOK, config)
}
