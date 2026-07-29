package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// AnalyticsHandler provides dashboard analytics endpoints.
type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(analyticsSvc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

// Stats returns aggregated dashboard statistics.
func (h *AnalyticsHandler) Stats(c *gin.Context) {
	stats, err := h.analyticsSvc.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Traffic returns traffic history data.
func (h *AnalyticsHandler) Traffic(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	data, err := h.analyticsSvc.GetTrafficHistory(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"traffic": data, "days": days})
}

// UserGrowth returns user registration growth data.
func (h *AnalyticsHandler) UserGrowth(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.analyticsSvc.GetUserGrowth(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"growth": data, "days": days})
}

// Revenue returns revenue history data.
func (h *AnalyticsHandler) Revenue(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.analyticsSvc.GetRevenueHistory(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revenue": data, "days": days})
}

// Online returns current online count.
func (h *AnalyticsHandler) Online(c *gin.Context) {
	count := h.analyticsSvc.GetOnlineCount()
	c.JSON(http.StatusOK, gin.H{"online": count})
}
