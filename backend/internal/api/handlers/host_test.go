package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home-run-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostHandler_Stats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHostHandler()
	router := gin.New()
	router.GET("/stats", handler.Stats)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats models.HostStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Basic sanity checks - values should be populated
	// CPU stats
	assert.GreaterOrEqual(t, stats.CPU.Usage, 0.0)
	assert.GreaterOrEqual(t, stats.CPU.Cores, 0)

	// Memory stats
	assert.GreaterOrEqual(t, stats.Memory.TotalGB, 0.0)

	// Storage stats
	assert.GreaterOrEqual(t, stats.Storage.TotalGB, 0.0)
}
