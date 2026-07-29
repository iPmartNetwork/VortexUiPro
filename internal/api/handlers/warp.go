package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// WARPHandler manages WARP+ proxy outbound.
type WARPHandler struct {
	svc *service.WARPProxyService
}

// NewWARPHandler creates a new WARP handler.
func NewWARPHandler(svc *service.WARPProxyService) *WARPHandler {
	return &WARPHandler{svc: svc}
}

// GetConfig returns the WARP configuration.
func (h *WARPHandler) GetConfig(c *gin.Context) {
	cfg := h.svc.GetConfig()
	c.JSON(http.StatusOK, cfg)
}

// UpdateConfig updates the WARP configuration.
func (h *WARPHandler) UpdateConfig(c *gin.Context) {
	var cfg service.WARPConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.UpdateConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "config updated"})
}

// Connect establishes the WARP tunnel.
func (h *WARPHandler) Connect(c *gin.Context) {
	if err := h.svc.Connect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "connected"})
}

// Disconnect tears down the WARP tunnel.
func (h *WARPHandler) Disconnect(c *gin.Context) {
	if err := h.svc.Disconnect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "disconnected"})
}

// GetStatus returns the WARP connection status.
func (h *WARPHandler) GetStatus(c *gin.Context) {
	status := h.svc.GetStatus()
	c.JSON(http.StatusOK, status)
}

// GetXrayOutbound returns the Xray outbound config for WARP.
func (h *WARPHandler) GetXrayOutbound(c *gin.Context) {
	config, err := h.svc.GetXrayOutboundConfig()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}
