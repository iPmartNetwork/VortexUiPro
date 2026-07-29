package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/rbac"
)

// PermissionMiddleware checks if the authenticated admin has a specific permission.
func PermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, exists := c.Get("admin_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var id int64
		switch v := adminID.(type) {
		case int64:
			id = v
		case float64:
			id = int64(v)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid admin identity"})
			return
		}

		role, err := database.GetAdminRole(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role assigned"})
			return
		}

		if role.OwnerRole {
			c.Next()
			return
		}

		scope := rbac.PermissionScope(role, resource, action)
		if scope == "none" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Set("perm_scope", scope)
		c.Next()
	}
}

// RoleOrPermissionMiddleware allows access if the user has either a legacy role string
// OR the specified RBAC permission.
func RoleOrPermissionMiddleware(legacyRoles []string, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check legacy role
		role, exists := c.Get("role")
		if exists {
			roleStr := role.(string)
			for _, r := range legacyRoles {
				if strings.EqualFold(roleStr, r) {
					c.Next()
					return
				}
			}
		}

		// Then check RBAC permission
		adminID, exists := c.Get("admin_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var id int64
		switch v := adminID.(type) {
		case int64:
			id = v
		case float64:
			id = int64(v)
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role assigned"})
			return
		}

		adminRole, err := database.GetAdminRole(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role assigned"})
			return
		}

		if adminRole.OwnerRole || rbac.CheckPermission(adminRole, resource, action) {
			scope := rbac.PermissionScope(adminRole, resource, action)
			c.Set("perm_scope", scope)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
