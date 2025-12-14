package handlers

import (
	"net/http"
	"os"

	"home-run-backend/internal/services/federation"

	"github.com/gin-gonic/gin"
)

type FederationHandler struct {
	aggregator *federation.Aggregator
}

func NewFederationHandler(aggregator *federation.Aggregator) *FederationHandler {
	return &FederationHandler{
		aggregator: aggregator,
	}
}

// Services returns local services for remote hosts
func (h *FederationHandler) Services(c *gin.Context) {
	ctx := c.Request.Context()
	services := h.aggregator.GetLocalServices(ctx)

	// Get hostname for identification
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"host":     hostname,
	})
}
