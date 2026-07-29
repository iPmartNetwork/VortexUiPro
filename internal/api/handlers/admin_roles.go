package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// AdminRoleHandler handles RBAC role management endpoints.
type AdminRoleHandler struct {
	roleSvc *service.AdminRoleService
}

// NewAdminRoleHandler creates a new admin role handler.
func NewAdminRoleHandler(roleSvc *service.AdminRoleService) *AdminRoleHandler {
	return &AdminRoleHandler{roleSvc: roleSvc}
}

// List returns all roles.
func (h *AdminRoleHandler) List(c *gin.Context) {
	roles, err := h.roleSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if roles == nil {
		roles = []*service.AdminRoleView{}
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles, "total": len(roles)})
}

// Get returns a single role by ID.
func (h *AdminRoleHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	role, err := h.roleSvc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// Create creates a new role.
func (h *AdminRoleHandler) Create(c *gin.Context) {
	var req service.AdminRolePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	role, err := h.roleSvc.Create(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// Update updates a role.
func (h *AdminRoleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var req service.AdminRolePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	role, err := h.roleSvc.Update(id, req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

// Duplicate duplicates a role.
func (h *AdminRoleHandler) Duplicate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	role, err := h.roleSvc.Duplicate(id)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// Delete deletes a role.
func (h *AdminRoleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	if err := h.roleSvc.Delete(id); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}
