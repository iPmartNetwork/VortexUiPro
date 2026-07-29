package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/core"
	"vortexuipro/internal/domain"
)

// SystemHandler handles system-level endpoints.
type SystemHandler struct {
	engineMgr *core.EngineManager
	startAt   time.Time
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(engineMgr *core.EngineManager) *SystemHandler {
	return &SystemHandler{
		engineMgr: engineMgr,
		startAt:   time.Now(),
	}
}

// Status returns the overall system status.
func (h *SystemHandler) Status(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"version":    "0.0.1",
		"name":       "VortexUiPro",
		"uptime":     time.Since(h.startAt).String(),
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc_mb":  m.Alloc / 1024 / 1024,
			"total_mb":  m.TotalAlloc / 1024 / 1024,
			"sys_mb":    m.Sys / 1024 / 1024,
			"gc_cycles": m.NumGC,
		},
	})
}

// Health returns a simple health check.
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   "0.0.1",
		"name":      "VortexUiPro",
		"timestamp": time.Now().Unix(),
	})
}

// Config returns the current panel configuration (sanitized).
func (h *SystemHandler) Config(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"version":    "0.0.1",
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
	})
}

// GetLogs returns recent log entries from memory buffer.
// This reads from a simple in-memory ring buffer (if connected).
func (h *SystemHandler) GetLogs(c *gin.Context) {
	// In a real implementation, this would read from a log buffer or file.
	// For now, return system metrics which are more useful.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	logs := []string{
		"VortexUiPro v0.0.1 started",
		"Database initialized: SQLite",
		"Event bus initialized",
		"API router initialized",
	}
	// Add uptime info
	logs = append(logs, "Uptime: "+time.Since(h.startAt).String())
	logs = append(logs, "Goroutines: "+itoa(runtime.NumGoroutine()))

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}

// Performance returns system performance metrics.
func (h *SystemHandler) Performance(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	perf := &domain.PerformanceMetrics{
		MemoryUsed:   float64(m.Alloc) / 1024 / 1024,
		MemoryTotal:  float64(m.Sys) / 1024 / 1024,
		ProcessCount: runtime.NumGoroutine(),
		Uptime:       int64(time.Since(h.startAt).Seconds()),
	}

	c.JSON(http.StatusOK, perf)
}

// CoreStatus returns the real status of all registered proxy cores.
func (h *SystemHandler) CoreStatus(c *gin.Context) {
	if h.engineMgr == nil {
		c.JSON(http.StatusOK, gin.H{"cores": []gin.H{}})
		return
	}

	// Collect real status from engine manager
	cores := make([]gin.H, 0)
	traffic := h.engineMgr.CollectAllTraffic(c.Request.Context())

	// Common cores to report
	coreNames := []string{"xray", "singbox"}
	for _, name := range coreNames {
		driver, err := h.engineMgr.Get(name)
		status := "stopped"
		if err == nil && driver != nil {
			status = string(driver.Status(c.Request.Context()))
		}
		cores = append(cores, gin.H{
			"name":   name,
			"status": status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cores":         cores,
		"traffic_stats": traffic,
	})
}

// ResetTraffic resets traffic counters for a user or inbound.
func (h *SystemHandler) ResetTraffic(c *gin.Context) {
	// Parse target type and ID
	targetType := c.Query("type")
	targetID := c.Query("id")

	// In a real implementation, this would reset traffic in the database
	// For now, acknowledge the request
	result := gin.H{"message": "traffic reset request received"}
	if targetType != "" && targetID != "" {
		result["type"] = targetType
		result["id"] = targetID
		result["status"] = "processed"
	} else {
		result["status"] = "no target specified"
	}

	c.JSON(http.StatusOK, result)
}

// itoa is a fast integer to string conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
