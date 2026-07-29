package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/auth"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// AdminManagementHandler handles admin user management.
type AdminManagementHandler struct {
	adminSvc *service.AdminService
}

// NewAdminManagementHandler creates a new admin management handler.
func NewAdminManagementHandler(adminSvc *service.AdminService) *AdminManagementHandler {
	return &AdminManagementHandler{adminSvc: adminSvc}
}

// List returns all admins.
func (h *AdminManagementHandler) List(c *gin.Context) {
	admins, err := database.ListAdmins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Attach role names
	type AdminView struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
		Role     string `json:"role"`
		RoleID   int    `json:"roleId"`
		RoleName string `json:"roleName,omitempty"`
		Status   string `json:"status"`
		CreatedAt int64 `json:"created_at"`
	}
	views := make([]AdminView, 0, len(admins))
	for _, a := range admins {
		roleName := ""
		if a.RoleID > 0 {
			var role database.AdminRole
			if err := database.DB.First(&role, a.RoleID).Error; err == nil {
				roleName = role.Name
			}
		}
		views = append(views, AdminView{
			ID:       a.ID,
			Username: a.Username,
			Email:    a.Email,
			Role:     a.Role,
			RoleID:   a.RoleID,
			RoleName: roleName,
			Status:   a.Status,
			CreatedAt: a.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"admins": views, "total": len(views)})
}

// Get returns a single admin by ID.
func (h *AdminManagementHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}
	admin, err := database.GetAdminByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}
	roleName := ""
	if admin.RoleID > 0 {
		var role database.AdminRole
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			roleName = role.Name
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"admin": admin,
		"role_name": roleName,
	})
}

// CreateRequest is the payload for creating an admin.
type AdminCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email,omitempty"`
	RoleID   int    `json:"roleId"`
	Status   string `json:"status,omitempty"`
}

// Create creates a new admin.
func (h *AdminManagementHandler) Create(c *gin.Context) {
	var req AdminCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Default role
	roleID := req.RoleID
	if roleID <= 0 {
		// Find operator role as default
		var opRole database.AdminRole
		if err := database.DB.Where("slug = ?", "operator").First(&opRole).Error; err == nil {
			roleID = opRole.ID
		}
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	admin := &database.Admin{
		Username:     strings.TrimSpace(req.Username),
		PasswordHash: hash,
		Email:        strings.TrimSpace(req.Email),
		Role:         "admin",
		RoleID:       roleID,
		Status:       status,
	}
	if err := database.CreateAdmin(admin); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, admin)
}

// UpdateRequest is the payload for updating an admin.
type AdminUpdateRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	RoleID   int    `json:"roleId,omitempty"`
	Status   string `json:"status,omitempty"`
}

// Update updates an admin.
func (h *AdminManagementHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}

	admin, err := database.GetAdminByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	var req AdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.Username != "" {
		admin.Username = strings.TrimSpace(req.Username)
	}
	if req.Email != "" {
		admin.Email = strings.TrimSpace(req.Email)
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		admin.PasswordHash = hash
		admin.LoginEpoch++
	}
	if req.RoleID > 0 {
		// Verify role exists
		var role database.AdminRole
		if err := database.DB.First(&role, req.RoleID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role not found"})
			return
		}
		if role.OwnerRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "owner role cannot be assigned here"})
			return
		}
		admin.RoleID = req.RoleID
	}
	if req.Status != "" {
		admin.Status = req.Status
	}

	if err := database.UpdateAdmin(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admin)
}

// Delete deletes an admin.
func (h *AdminManagementHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}

	admin, err := database.GetAdminByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	// Prevent deleting owner
	var ownerRole database.AdminRole
	if err := database.DB.Where("slug = ?", "owner").First(&ownerRole).Error; err == nil {
		if admin.RoleID == ownerRole.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "owner admin cannot be deleted"})
			return
		}
	}

	if err := database.DeleteAdmin(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "admin deleted"})
}

// SetEnabled enables or disables an admin.
func (h *AdminManagementHandler) SetEnabled(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	admin, err := database.GetAdminByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	var ownerRole database.AdminRole
	if err := database.DB.Where("slug = ?", "owner").First(&ownerRole).Error; err == nil {
		if admin.RoleID == ownerRole.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "owner admin cannot be disabled"})
			return
		}
	}

	status := "active"
	if !req.Enabled {
		status = "disabled"
	}
	admin.Status = status
	admin.LoginEpoch++
	database.UpdateAdmin(admin)

	c.JSON(http.StatusOK, gin.H{"message": "admin status updated"})
}
