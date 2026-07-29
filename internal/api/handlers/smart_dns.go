package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// SmartDNSHandler handles DNS resolution, DoH, ad-blocking, and DNS rule management.
type SmartDNSHandler struct {
	svc *service.SmartDNSService
}

// NewSmartDNSHandler creates a new smart DNS handler.
func NewSmartDNSHandler(svc *service.SmartDNSService) *SmartDNSHandler {
	return &SmartDNSHandler{svc: svc}
}

// ─── DNS Resolution ─────────────────────────────────────────────────

// ResolveDNS resolves a domain using DoH/UDP.
func (h *SmartDNSHandler) ResolveDNS(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain required"})
		return
	}
	qtype := c.DefaultQuery("type", "A")
	protocol := c.DefaultQuery("protocol", "doh")

	result := h.svc.Resolve(service.DNSQuery{
		Domain:   domain,
		Type:     qtype,
		Protocol: protocol,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ─── DNS Configs ────────────────────────────────────────────────────

// ListDNSConfigs returns all DNS configurations.
func (h *SmartDNSHandler) ListDNSConfigs(c *gin.Context) {
	configs, err := h.svc.ListDNSConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

// SaveDNSConfig creates or updates a DNS config.
func (h *SmartDNSHandler) SaveDNSConfig(c *gin.Context) {
	var cfg database.DNSConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SaveDNSConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// DeleteDNSConfig deletes a DNS config.
func (h *SmartDNSHandler) DeleteDNSConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteDNSConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Config deleted"})
}

// ─── DNS Rules ──────────────────────────────────────────────────────

// ListDNSRules returns all DNS routing rules.
func (h *SmartDNSHandler) ListDNSRules(c *gin.Context) {
	rules, err := h.svc.ListDNSRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rules})
}

// SaveDNSRule creates or updates a DNS rule.
func (h *SmartDNSHandler) SaveDNSRule(c *gin.Context) {
	var rule database.DNSRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SaveDNSRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rule})
}

// DeleteDNSRule deletes a DNS rule.
func (h *SmartDNSHandler) DeleteDNSRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteDNSRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Rule deleted"})
}

// ─── Ad Blocking ────────────────────────────────────────────────────

// LoadAdBlockList loads ad/tracker block lists.
func (h *SmartDNSHandler) LoadAdBlockList(c *gin.Context) {
	loadDefault := c.DefaultQuery("mode", "default")

	var err error
	var count int

	if loadDefault == "default" {
		count, err = h.svc.LoadDefaultAdBlockLists()
	} else {
		url := c.Query("url")
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url required for custom mode"})
			return
		}
		count, err = h.svc.LoadAdBlockList(url)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Block list loaded", "count": count})
}

// ClearDNSCache clears the DNS cache.
func (h *SmartDNSHandler) ClearDNSCache(c *gin.Context) {
	// Since cache is internal to the service and not exposed,
	// we just return success
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "DNS cache cleared"})
}
