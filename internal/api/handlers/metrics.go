package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/metrics"
)

// MetricsHandler serves real-time and historical metrics.
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{collector: collector}
}

// GetMetrics returns a full JSON metrics snapshot.
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	if h.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics collector not initialized"})
		return
	}
	c.JSON(http.StatusOK, h.collector.Snapshot())
}

// CollectorSnapshot returns the latest metrics snapshot (for WebSocket broadcasting).
func (h *MetricsHandler) CollectorSnapshot() map[string]any {
	if h.collector == nil {
		return nil
	}
	return h.collector.Snapshot()
}

// PrometheusMetrics returns metrics in Prometheus text format.
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	if h.collector == nil {
		c.String(http.StatusServiceUnavailable, "")
		return
	}
	c.String(http.StatusOK, h.collector.PrometheusText())
}

// GetHistory returns time-series data for chart rendering.
func (h *MetricsHandler) GetHistory(c *gin.Context) {
	if h.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics collector not initialized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"history": h.collector.History(),
	})
}
