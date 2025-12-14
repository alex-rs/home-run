package handlers

import (
	"net/http"

	"home-run-backend/internal/logger"
	"home-run-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type HostHandler struct{}

func NewHostHandler() *HostHandler {
	return &HostHandler{}
}

// Stats returns system resource usage
func (h *HostHandler) Stats(c *gin.Context) {
	stats := models.HostStats{}

	// Get CPU info
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		stats.CPU.Usage = cpuPercent[0]
	} else if err != nil {
		logger.WithField("error", err.Error()).Warn("Failed to get CPU usage")
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		stats.CPU.Cores = len(cpuInfo)
		totalThreads := 0
		for _, info := range cpuInfo {
			totalThreads += int(info.Cores)
		}
		stats.CPU.Threads = totalThreads
	} else if err != nil {
		logger.WithField("error", err.Error()).Warn("Failed to get CPU info")
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		stats.Memory.UsedGB = float64(memInfo.Used) / (1024 * 1024 * 1024)
		stats.Memory.TotalGB = float64(memInfo.Total) / (1024 * 1024 * 1024)
	} else {
		logger.WithField("error", err.Error()).Warn("Failed to get memory info")
	}

	// Get disk info for root partition
	diskInfo, err := disk.Usage("/")
	if err == nil {
		stats.Storage.UsedGB = float64(diskInfo.Used) / (1024 * 1024 * 1024)
		stats.Storage.TotalGB = float64(diskInfo.Total) / (1024 * 1024 * 1024)
	} else {
		logger.WithField("error", err.Error()).Warn("Failed to get disk info")
	}

	logger.Log.Debug("Retrieved host stats")
	c.JSON(http.StatusOK, stats)
}
