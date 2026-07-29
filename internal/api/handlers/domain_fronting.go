package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// DomainFrontingHandler handles CDN fronting and proxy config generation.
type DomainFrontingHandler struct {
	svc *service.DomainFrontingService
}

// NewDomainFrontingHandler creates a new domain fronting handler.
func NewDomainFrontingHandler(svc *service.DomainFrontingService) *DomainFrontingHandler {
	return &DomainFrontingHandler{svc: svc}
}

// ListProviders returns known CDN frontable domains.
func (h *DomainFrontingHandler) ListProviders(c *gin.Context) {
	providers := h.svc.ListProviders()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": providers})
}

// ScanDomain scans a specific domain for fronting.
func (h *DomainFrontingHandler) ScanDomain(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		domain = c.Param("domain")
	}
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain required"})
		return
	}
	provider := c.DefaultQuery("provider", "")
	result := h.svc.ScanCustomDomain(domain)
	if provider != "" {
		result.Provider = provider
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ScanAll scans all known frontable domains.
func (h *DomainFrontingHandler) ScanAll(c *gin.Context) {
	results := h.svc.ScanAllKnown()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ListFrontable returns all scanned domains from the database.
func (h *DomainFrontingHandler) ListFrontable(c *gin.Context) {
	domains, err := h.svc.ListFrontableDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": domains})
}

// GenerateConfig generates a CDN-fronted proxy configuration.
func (h *DomainFrontingHandler) GenerateConfig(c *gin.Context) {
	frontDomain := c.Query("front_domain")
	hiddenDomain := c.DefaultQuery("hidden_domain", "REPLACE_ME")
	provider := c.DefaultQuery("provider", "cloudflare")

	if frontDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "front_domain required"})
		return
	}

	cfg := h.svc.GenerateProxyConfig(frontDomain, hiddenDomain, provider)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// DeleteDomain deletes a CDN domain from the database.
func (h *DomainFrontingHandler) DeleteDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := database.DB.Delete(&database.CDNDomain{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Domain deleted"})
}
