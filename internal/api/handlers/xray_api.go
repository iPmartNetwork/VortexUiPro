package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/core"
	"vortexuipro/internal/core/xray"
	"vortexuipro/internal/service"
)

// XrayAPIHandler provides REST endpoints for xray gRPC API integration.
type XrayAPIHandler struct {
	xraySvc *service.XrayService
}

// NewXrayAPIHandler creates a new XrayAPI handler.
func NewXrayAPIHandler(xraySvc *service.XrayService) *XrayAPIHandler {
	return &XrayAPIHandler{xraySvc: xraySvc}
}

// GetProcessInfo returns runtime info about the xray process.
func (h *XrayAPIHandler) GetProcessInfo(c *gin.Context) {
	info := h.xraySvc.GetProcessInfo()
	c.JSON(http.StatusOK, gin.H{"data": info})
}

// GetLogs returns recent log lines from the xray process.
func (h *XrayAPIHandler) GetLogs(c *gin.Context) {
	n := 50
	if nStr := c.Query("lines"); nStr != "" {
		if parsed, err := strconv.Atoi(nStr); err == nil && parsed > 0 && parsed <= 500 {
			n = parsed
		}
	}
	logs := h.xraySvc.GetLogs(n)
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// ValidateConfig validates an xray JSON config.
func (h *XrayAPIHandler) ValidateConfig(c *gin.Context) {
	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "config field required"})
		return
	}

	if err := h.xraySvc.ValidateConfig([]byte(req.Config)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "valid": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// GetOnlineUsers returns users with live connections via xray's gRPC StatsService.
func (h *XrayAPIHandler) GetOnlineUsers(c *gin.Context) {
	users, err := h.xraySvc.GetOnlineUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if users == nil {
		users = []xray.OnlineUser{}
	}
	c.JSON(http.StatusOK, gin.H{"data": users, "count": len(users)})
}

// GetTrafficStats returns collected traffic stats (inbound + client).
func (h *XrayAPIHandler) GetTrafficStats(c *gin.Context) {
	traffic, err := h.xraySvc.CollectTraffic(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientTraffic, _ := h.xraySvc.CollectClientTraffic(c.Request.Context())

	if traffic == nil {
		traffic = []core.TrafficStats{}
	}
	if clientTraffic == nil {
		clientTraffic = []xray.ClientTraffic{}
	}
	c.JSON(http.StatusOK, gin.H{
		"inbound_traffic": traffic,
		"client_traffic":  clientTraffic,
	})
}

// TestRoute tests a route through the running xray core's router.
func (h *XrayAPIHandler) TestRoute(c *gin.Context) {
	var req xray.RouteTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.xraySvc.TestRoute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetBalancerInfo queries a balancer's live state.
func (h *XrayAPIHandler) GetBalancerInfo(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag required"})
		return
	}

	info, err := h.xraySvc.GetBalancerInfo(c.Request.Context(), tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": info})
}

// SetBalancerTarget forces a balancer to a specific outbound.
func (h *XrayAPIHandler) SetBalancerTarget(c *gin.Context) {
	var req struct {
		BalancerTag string `json:"balancer_tag" binding:"required"`
		Target      string `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.xraySvc.SetBalancerTarget(c.Request.Context(), req.BalancerTag, req.Target); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "balancer target updated"})
}
