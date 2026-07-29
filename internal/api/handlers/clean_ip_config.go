package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// CleanIPHandler handles clean IP scanning endpoints.
type CleanIPHandler struct {
	svc *service.CleanIPScanner
}

// NewCleanIPHandler creates a new handler.
func NewCleanIPHandler(svc *service.CleanIPScanner) *CleanIPHandler {
	return &CleanIPHandler{svc: svc}
}

// GetResults returns the latest scan results.
func (h *CleanIPHandler) GetResults(c *gin.Context) {
	results := h.svc.GetResults()
	c.JSON(http.StatusOK, gin.H{"results": results, "total": len(results)})
}

// ScanNow triggers an immediate scan.
func (h *CleanIPHandler) ScanNow(c *gin.Context) {
	h.svc.ScanNow()
	c.JSON(http.StatusOK, gin.H{"message": "Scan triggered"})
}

// ConfigVersionHandler handles config versioning endpoints.
type ConfigVersionHandler struct {
	svc *service.ConfigVersionService
}

// NewConfigVersionHandler creates a new handler.
func NewConfigVersionHandler(svc *service.ConfigVersionService) *ConfigVersionHandler {
	return &ConfigVersionHandler{svc: svc}
}

// ListVersions returns versions for a resource.
func (h *ConfigVersionHandler) ListVersions(c *gin.Context) {
	resource := c.Query("resource")
	resourceID, _ := strconv.ParseInt(c.Query("resource_id"), 10, 64)
	if resource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource query param required"})
		return
	}
	versions, err := h.svc.ListVersions(resource, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"versions": versions, "total": len(versions)})
}

// Rollback restores a previous version.
func (h *ConfigVersionHandler) Rollback(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	actor := c.GetString("username")
	if actor == "" {
		actor = "admin"
	}
	ver, err := h.svc.Rollback(c.Request.Context(), id, actor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rolled back", "version": ver})
}
