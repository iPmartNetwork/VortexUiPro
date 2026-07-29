package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// SecuritySettingsHandler manages security settings endpoints.
type SecuritySettingsHandler struct {
	svc *service.SecuritySettingsService
	adv *service.AdvancedSecurityService
}

// NewSecuritySettingsHandler creates a new handler.
func NewSecuritySettingsHandler(svc *service.SecuritySettingsService, adv *service.AdvancedSecurityService) *SecuritySettingsHandler {
	return &SecuritySettingsHandler{svc: svc, adv: adv}
}

// GetPasswordPolicy returns current password policy.
func (h *SecuritySettingsHandler) GetPasswordPolicy(c *gin.Context) {
	policy := h.svc.GetPasswordPolicy()
	c.JSON(http.StatusOK, policy)
}

// SavePasswordPolicy saves password policy settings.
func (h *SecuritySettingsHandler) SavePasswordPolicy(c *gin.Context) {
	var policy service.PasswordPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password policy"})
		return
	}
	if err := h.svc.SavePasswordPolicy(&policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password policy saved"})
}

// GetGeoBlock returns blocked countries.
func (h *SecuritySettingsHandler) GetGeoBlock(c *gin.Context) {
	countries := h.svc.GetGeoBlock()
	c.JSON(http.StatusOK, gin.H{"blocked_countries": countries})
}

// SetGeoBlock saves blocked countries.
func (h *SecuritySettingsHandler) SetGeoBlock(c *gin.Context) {
	var req struct {
		Countries []string `json:"countries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.SetGeoBlock(req.Countries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "geo-block updated"})
}

// GetBannedIPs returns banned IPs.
func (h *SecuritySettingsHandler) GetBannedIPs(c *gin.Context) {
	ips := h.svc.GetBannedIPs()
	c.JSON(http.StatusOK, gin.H{"banned_ips": ips})
}

// AddBannedIP adds an IP to the ban list.
func (h *SecuritySettingsHandler) AddBannedIP(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip required"})
		return
	}
	if err := h.svc.AddBannedIP(req.IP, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Also record in threat detection
	if h.adv != nil {
		h.adv.RecordFailedLogin(req.IP)
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP banned"})
}

// RemoveBannedIP removes an IP from the ban list.
func (h *SecuritySettingsHandler) RemoveBannedIP(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip query param required"})
		return
	}
	if err := h.svc.RemoveBannedIP(ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP unbanned"})
}

// GetWhitelistedIPs returns whitelisted IPs.
func (h *SecuritySettingsHandler) GetWhitelistedIPs(c *gin.Context) {
	ips := h.svc.GetWhitelistedIPs()
	c.JSON(http.StatusOK, gin.H{"whitelisted_ips": ips})
}

// AddWhitelistedIP adds an IP to whitelist.
func (h *SecuritySettingsHandler) AddWhitelistedIP(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip required"})
		return
	}
	if err := h.svc.AddWhitelistedIP(req.IP, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP whitelisted"})
}

// RemoveWhitelistedIP removes an IP from whitelist.
func (h *SecuritySettingsHandler) RemoveWhitelistedIP(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip query param required"})
		return
	}
	if err := h.svc.RemoveWhitelistedIP(ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP removed from whitelist"})
}

// GetThreatConfig returns threat detection settings.
func (h *SecuritySettingsHandler) GetThreatConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"failed_login_threshold": h.svc.GetFailedLoginThreshold(),
		"ban_duration_minutes":   h.svc.GetBanDuration(),
	})
}
