package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// DockerHandler manages Docker containers from within the panel.
type DockerHandler struct {
	svc *service.DockerService
}

// NewDockerHandler creates a new Docker handler.
func NewDockerHandler(svc *service.DockerService) *DockerHandler {
	return &DockerHandler{svc: svc}
}

// Status returns whether Docker is available.
func (h *DockerHandler) Status(c *gin.Context) {
	stats, err := h.svc.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// ListContainers returns all containers.
func (h *DockerHandler) ListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "false") == "true"
	containers, err := h.svc.ListContainers(all)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": containers})
}

// CreateContainer creates a new Docker container.
func (h *DockerHandler) CreateContainer(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Image   string `json:"image" binding:"required"`
		EnvVars string `json:"env_vars"`
		Port    int    `json:"port"`
		Network string `json:"network"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	containerID, err := h.svc.CreateContainer(req.Name, req.Image, req.EnvVars, req.Port, req.Network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "container_id": containerID})
}

// StartContainer starts a container.
func (h *DockerHandler) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StartContainer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Container started"})
}

// StopContainer stops a container.
func (h *DockerHandler) StopContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StopContainer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Container stopped"})
}

// RestartContainer restarts a container.
func (h *DockerHandler) RestartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.RestartContainer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Container restarted"})
}

// RemoveContainer removes a container.
func (h *DockerHandler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"
	if err := h.svc.RemoveContainer(id, force); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Container removed"})
}

// GetContainerLogs returns container logs.
func (h *DockerHandler) GetContainerLogs(c *gin.Context) {
	id := c.Param("id")
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "100"))
	logs, err := h.svc.GetContainerLogs(id, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": logs})
}

// ListImages returns Docker images.
func (h *DockerHandler) ListImages(c *gin.Context) {
	images, err := h.svc.ListImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": images})
}

// PullImage pulls a Docker image.
func (h *DockerHandler) PullImage(c *gin.Context) {
	var req struct {
		Image string `json:"image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}
	if err := h.svc.PullImage(req.Image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image pulled"})
}

// RemoveImage removes a Docker image.
func (h *DockerHandler) RemoveImage(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"
	if err := h.svc.RemoveImage(id, force); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image removed"})
}

// PruneImages removes unused Docker images.
func (h *DockerHandler) PruneImages(c *gin.Context) {
	output, err := h.svc.PruneImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Images pruned", "output": output})
}

// GetContainerStats returns stats for a container.
func (h *DockerHandler) GetContainerStats(c *gin.Context) {
	id := c.Param("id")
	stats, err := h.svc.GetContainerStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
