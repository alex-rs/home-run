package handlers

import (
	"net/http"
	"strconv"

	"home-run-backend/internal/services"
	"home-run-backend/internal/services/federation"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, svc)
}

// GetConfig returns the content of a service's config file
func (h *ServicesHandler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()
	serviceID := c.Param("id")
	indexStr := c.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid config index",
		})
		return
	}

	config, err := h.manager.GetConfigContent(ctx, serviceID, index)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}
