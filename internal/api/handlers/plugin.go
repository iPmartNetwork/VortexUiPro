package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// PluginHandler manages plugin lifecycle.
type PluginHandler struct {
	svc *service.PluginService
}

// NewPluginHandler creates a new plugin handler.
func NewPluginHandler(svc *service.PluginService) *PluginHandler {
	return &PluginHandler{svc: svc}
}

// ListPlugins returns all plugins.
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	plugins := h.svc.ListPlugins()
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// GetPlugin returns a specific plugin.
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	id := c.Param("id")
	plugin, err := h.svc.GetPlugin(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plugin)
}

// LoadPlugin loads a new plugin from a .so file.
func (h *PluginHandler) LoadPlugin(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		ID      string `json:"id" binding:"required"`
		Name    string `json:"name" binding:"required"`
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path, id, and name required"})
		return
	}
	if req.Version == "" {
		req.Version = "1.0.0"
	}
	if err := h.svc.LoadPlugin(req.Path, req.ID, req.Name, req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin loaded"})
}

// UnloadPlugin unloads a plugin.
func (h *PluginHandler) UnloadPlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.UnloadPlugin(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin unloaded"})
}

// EnablePlugin enables or disables a plugin.
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	id := c.Param("id")
	var req struct{ Enabled bool `json:"enabled"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.EnablePlugin(id, req.Enabled); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin updated"})
}
