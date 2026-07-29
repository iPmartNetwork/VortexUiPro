package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// SubProfileHandler manages subscription profiles, hosts, and formats.
type SubProfileHandler struct {
	svc *service.SubProfileService
}

// NewSubProfileHandler creates a new handler.
func NewSubProfileHandler(svc *service.SubProfileService) *SubProfileHandler {
	return &SubProfileHandler{svc: svc}
}

// ─── Profiles ──────────────────────────────────────────────────────

// ListProfiles returns all subscription profiles for an inbound.
func (h *SubProfileHandler) ListProfiles(c *gin.Context) {
	inboundID, _ := strconv.ParseInt(c.Query("inbound_id"), 10, 64)
	profiles, err := h.svc.ListProfiles(inboundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profiles": profiles, "total": len(profiles)})
}

// CreateProfile adds a new subscription profile.
func (h *SubProfileHandler) CreateProfile(c *gin.Context) {
	var req struct {
		InboundID   int64  `json:"inbound_id" binding:"required"`
		Dest        string `json:"dest" binding:"required"`
		Port        int    `json:"port" binding:"required"`
		Remark      string `json:"remark,omitempty"`
		Enabled     bool   `json:"enabled"`
		Network     string `json:"network,omitempty"`
		Security    string `json:"security,omitempty"`
		SNI         string `json:"sni,omitempty"`
		ALPN        string `json:"alpn,omitempty"`
		Fingerprint string `json:"fingerprint,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inbound_id, dest, port required"})
		return
	}
	profile := &database.SubscriptionProfile{
		InboundID:   req.InboundID,
		Dest:        req.Dest,
		Port:        req.Port,
		Remark:      req.Remark,
		Enabled:     req.Enabled,
		Network:     req.Network,
		Security:    req.Security,
		SNI:         req.SNI,
		ALPN:        req.ALPN,
		Fingerprint: req.Fingerprint,
	}
	created, err := h.svc.CreateProfile(c.Request.Context(), profile)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// DeleteProfile removes a subscription profile.
func (h *SubProfileHandler) DeleteProfile(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteProfile(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile deleted"})
}

// ─── Hosts ─────────────────────────────────────────────────────────

// ListHosts returns all subscription hosts.
func (h *SubProfileHandler) ListHosts(c *gin.Context) {
	hosts, err := h.svc.ListHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}

// CreateHost adds a new subscription host.
func (h *SubProfileHandler) CreateHost(c *gin.Context) {
	var req struct {
		Remark   string `json:"remark" binding:"required"`
		Domain   string `json:"domain" binding:"required"`
		CertFile string `json:"cert_file,omitempty"`
		KeyFile  string `json:"key_file,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "remark and domain required"})
		return
	}
	host := &database.SubscriptionHost{
		Remark:   req.Remark,
		Domain:   req.Domain,
		CertFile: req.CertFile,
		KeyFile:  req.KeyFile,
		Enable:   true,
	}
	created, err := h.svc.CreateHost(c.Request.Context(), host)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// DeleteHost removes a subscription host.
func (h *SubProfileHandler) DeleteHost(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteHost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "host deleted"})
}

// ─── Formats & Vars ────────────────────────────────────────────────

// ListFormats returns all supported subscription formats.
func (h *SubProfileHandler) ListFormats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"formats": h.svc.ListFormats()})
}

// ListRemarkVars returns all available remark template variables.
func (h *SubProfileHandler) ListRemarkVars(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"vars": h.svc.ListRemarkVars()})
}
