package rbac

import (
	"encoding/json"
	"strings"

	"vortexuipro/internal/database"

	"gorm.io/gorm"
)

// ClientAccessMode defines how an admin can access clients.
type ClientAccessMode string

const (
	ClientAccessAll  ClientAccessMode = "all"
	ClientAccessOwn  ClientAccessMode = "own"
	ClientAccessNone ClientAccessMode = "none"
)

// ClientAccessScope contains the resolved access scope for an admin.
type ClientAccessScope struct {
	AdminID          int64
	Mode             ClientAccessMode
	RestrictGroups   bool
	AllowAllGroups   bool
	AllowedGroups    []string
	RestrictInbounds bool
	AllowAllInbounds bool
	AllowedInboundIDs []int
}

// HasClientAccess returns true if the scope allows any client access.
func (s ClientAccessScope) HasClientAccess() bool {
	return s.Mode != ClientAccessNone
}

// ClientAllowed checks if a specific client record is accessible under this scope.
func (s ClientAccessScope) ClientAllowed(client *database.Client) bool {
	if client == nil || s.Mode == ClientAccessNone {
		return false
	}

	// Group restriction check
	if s.RestrictGroups && !s.AllowAllGroups {
		if len(s.AllowedGroups) == 0 {
			return false
		}
		groupKey := strings.ToLower(strings.TrimSpace(client.Group))
		allowed := false
		for _, g := range s.AllowedGroups {
			if strings.ToLower(strings.TrimSpace(g)) == groupKey {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Inbound restriction check
	if s.RestrictInbounds && !s.AllowAllInbounds {
		if len(s.AllowedInboundIDs) == 0 {
			return false
		}
		inboundAllowed := false
		for _, id := range s.AllowedInboundIDs {
			if int64(id) == client.InboundID {
				inboundAllowed = true
				break
			}
		}
		if !inboundAllowed {
			return false
		}
	}

	// Ownership check
	switch s.Mode {
	case ClientAccessAll:
		return true
	case ClientAccessOwn:
		if client.OwnerAdminID > 0 {
			return client.OwnerAdminID == s.AdminID
		}
		// Fall back to User.AdminID chain
		return true // We filter at the query level instead
	default:
		return false
	}
}

// ApplyClientScope adds WHERE clauses to a query to restrict by access scope.
func ApplyClientScope(db *gorm.DB, scope ClientAccessScope) *gorm.DB {
	if scope.Mode == ClientAccessNone {
		return db.Where("1 = 0")
	}

	// Group restriction
	if scope.RestrictGroups && !scope.AllowAllGroups {
		if len(scope.AllowedGroups) == 0 {
			return db.Where("1 = 0")
		}
		lowerGroups := make([]string, len(scope.AllowedGroups))
		for i, g := range scope.AllowedGroups {
			lowerGroups[i] = strings.ToLower(strings.TrimSpace(g))
		}
		db = db.Where("LOWER(\"group\") IN ?", lowerGroups)
	}

	// Inbound restriction
	if scope.RestrictInbounds && !scope.AllowAllInbounds {
		if len(scope.AllowedInboundIDs) == 0 {
			return db.Where("1 = 0")
		}
		db = db.Where("inbound_id IN ?", scope.AllowedInboundIDs)
	}

	// Ownership
	switch scope.Mode {
	case ClientAccessOwn:
		db = db.Where("owner_admin_id = ?", scope.AdminID)
	}

	return db
}

// GetClientAccessScopeForAdmin computes the ClientAccessScope from an admin's role.
func GetClientAccessScopeForAdmin(adminID int64, permission string) ClientAccessScope {
	role, err := database.GetAdminRole(adminID)
	if err != nil {
		return ClientAccessScope{Mode: ClientAccessNone, RestrictGroups: true, RestrictInbounds: true}
	}

	scope := ClientAccessScope{
		AdminID: adminID,
	}

	// Owner has full access
	if role.OwnerRole {
		scope.Mode = ClientAccessAll
		scope.AllowAllGroups = true
		scope.AllowAllInbounds = true
		return scope
	}

	// Resolve permission scope
	permScope := PermissionScope(role, "users", permission)
	switch permScope {
	case "all":
		scope.Mode = ClientAccessAll
	case "own":
		scope.Mode = ClientAccessOwn
	default:
		scope.Mode = ClientAccessNone
	}

	// Non-owner roles see only own clients (security hardening)
	if scope.Mode == ClientAccessAll {
		scope.Mode = ClientAccessOwn
	}

	// Resolve access from role's AccessJSON
	var accessRoot map[string]any
	if err := json.Unmarshal([]byte(role.AccessJSON), &accessRoot); err != nil {
		scope.RestrictGroups = true
		scope.RestrictInbounds = true
		return scope
	}

	// Groups
	scope.RestrictGroups = true
	if v, ok := accessRoot["allowAllGroups"]; ok && boolFeatureValue(v) {
		scope.AllowAllGroups = true
	}
	if v, ok := accessRoot["allowedGroups"]; ok {
		switch g := v.(type) {
		case []string:
			scope.AllowedGroups = g
		case []any:
			for _, item := range g {
				if s, ok := item.(string); ok {
					scope.AllowedGroups = append(scope.AllowedGroups, s)
				}
			}
		}
	}
	if scope.AllowAllGroups {
		scope.AllowedGroups = nil
	}

	// Inbounds
	scope.RestrictInbounds = true
	if v, ok := accessRoot["allowAllInbounds"]; ok && boolFeatureValue(v) {
		scope.AllowAllInbounds = true
	}
	if v, ok := accessRoot["allowedInboundIds"]; ok {
		switch ids := v.(type) {
		case []float64:
			for _, id := range ids {
				scope.AllowedInboundIDs = append(scope.AllowedInboundIDs, int(id))
			}
		case []int:
			scope.AllowedInboundIDs = ids
		case []any:
			for _, item := range ids {
				if f, ok := item.(float64); ok {
					scope.AllowedInboundIDs = append(scope.AllowedInboundIDs, int(f))
				}
			}
		}
	}
	if scope.AllowAllInbounds {
		scope.AllowedInboundIDs = nil
	}

	return scope
}

// CanCreateClient checks whether the admin has permission to create clients.
func CanCreateClient(adminID int64) bool {
	role, err := database.GetAdminRole(adminID)
	if err != nil {
		return false
	}
	if role.OwnerRole {
		return true
	}
	return CheckPermission(role, "users", "create")
}

// CanFilterClientOwners checks whether the admin can filter clients by owner.
func CanFilterClientOwners(adminID int64) bool {
	role, err := database.GetAdminRole(adminID)
	if err != nil {
		return false
	}
	if role.OwnerRole {
		return true
	}
	return CheckPermission(role, "users", "adminFilter")
}

// RestrictedInboundIDs returns the allowed inbound IDs if restricted.
func RestrictedInboundIDs(adminID int64) (restricted bool, ids []int) {
	scope := GetClientAccessScopeForAdmin(adminID, "view")
	if !scope.RestrictInbounds || scope.AllowAllInbounds {
		return false, nil
	}
	ids = scope.AllowedInboundIDs
	if len(ids) == 0 {
		return false, nil
	}
	return true, ids
}
