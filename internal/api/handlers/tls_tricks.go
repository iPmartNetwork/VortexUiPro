package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// TLSTricksHandler manages TLS trick profiles for anti-DPI.
type TLSTricksHandler struct {
	svc *service.TLSTricksService
}

// NewTLSTricksHandler creates a new TLS tricks handler.
func NewTLSTricksHandler(svc *service.TLSTricksService) *TLSTricksHandler {
	return &TLSTricksHandler{svc: svc}
}

// ListProfiles returns all TLS trick profiles.
func (h *TLSTricksHandler) ListProfiles(c *gin.Context) {
	profiles := h.svc.ListProfiles()
	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

// GetProfile returns a specific TLS profile.
func (h *TLSTricksHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	profile, err := h.svc.GetProfile(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// SaveProfile creates or updates a TLS profile.
func (h *TLSTricksHandler) SaveProfile(c *gin.Context) {
	var profile service.TLSProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.SaveProfile(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile saved"})
}

// DeleteProfile removes a TLS profile.
func (h *TLSTricksHandler) DeleteProfile(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteProfile(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile deleted"})
}

// EnableProfile enables or disables a TLS profile.
func (h *TLSTricksHandler) EnableProfile(c *gin.Context) {
	id := c.Param("id")
	var req struct{ Enabled bool `json:"enabled"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.EnableProfile(id, req.Enabled); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// GenerateConfig generates Xray config for a TLS profile.
func (h *TLSTricksHandler) GenerateConfig(c *gin.Context) {
	id := c.Query("profile_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile_id required"})
		return
	}
	config, err := h.svc.GenerateXrayConfig(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}
