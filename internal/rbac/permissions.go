package rbac

import (
	"encoding/json"
	"strconv"
	"strings"

	"vortexuipro/internal/database"
)

// ─── Permission Helpers ──────────────────────────────────────────────

// PermissionValue extracts a nested permission value from the role's permissions JSON.
// Keys follow Heimdall convention: "users.view", "inbounds.create", etc.
func PermissionValue(role *database.AdminRole, resource, action string) any {
	if role == nil || role.PermissionsJSON == "" {
		return nil
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(role.PermissionsJSON), &root); err != nil {
		return nil
	}

	res, ok := root[resource].(map[string]any)
	if !ok {
		return nil
	}

	// Try known aliases for the action
	for _, key := range permissionActionKeys(action) {
		if v, ok := res[key]; ok {
			return v
		}
	}
	return nil
}

func permissionActionKeys(action string) []string {
	switch action {
	case "view":
		return []string{"view", "read"}
	case "viewSimple":
		return []string{"viewSimple", "viewSimpleList", "read_simple"}
	case "create":
		return []string{"create"}
	case "update":
		return []string{"update"}
	case "delete":
		return []string{"delete"}
	case "resetUsage":
		return []string{"resetUsage", "reset_usage"}
	case "revokeSubscription":
		return []string{"revokeSubscription", "revoke_sub", "revokeSub"}
	case "activateNextPlan":
		return []string{"activateNextPlan", "activate_next_plan"}
	case "adminFilter":
		return []string{"adminFilter", "admin_filter"}
	case "setOwner":
		return []string{"setOwner", "set_owner"}
	default:
		return []string{action}
	}
}

// CheckPermission returns true if the role has the specified permission.
func CheckPermission(role *database.AdminRole, resource, action string) bool {
	if role == nil {
		return false
	}
	if role.OwnerRole {
		return true
	}

	v := PermissionValue(role, resource, action)
	if v == nil {
		return false
	}

	switch t := v.(type) {
	case bool:
		return t
	case string:
		t = strings.ToLower(strings.TrimSpace(t))
		return t == "all" || t == "true" || t == "yes"
	case float64:
		return t > 0
	}
	return false
}

// PermissionScope returns the scope of a permission (all/own/none).
func PermissionScope(role *database.AdminRole, resource, action string) string {
	if role == nil {
		return "none"
	}
	if role.OwnerRole {
		return "all"
	}

	v := PermissionValue(role, resource, action)
	if v == nil {
		return "none"
	}

	switch t := v.(type) {
	case bool:
		if t {
			return "all"
		}
		return "none"
	case string:
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "all" || t == "true" || t == "yes" {
			return "all"
		}
		if t == "own" || t == "self" || t == "1" {
			return "own"
		}
		return "none"
	case map[string]any:
		if scope, ok := t["scope"]; ok {
			switch s := scope.(type) {
			case float64:
				if int(s) == 2 {
					return "all"
				} else if int(s) == 1 {
					return "own"
				}
			case string:
				s = strings.ToLower(strings.TrimSpace(s))
				if s == "all" || s == "2" {
					return "all"
				} else if s == "own" || s == "1" {
					return "own"
				}
			}
		}
		return "none"
	case float64:
		if int(t) == 2 {
			return "all"
		} else if int(t) == 1 {
			return "own"
		}
		return "none"
	}
	return "none"
}

// ─── Feature Flags ───────────────────────────────────────────────────

// FeatureEnabled checks if a feature flag is enabled in the role's features JSON.
func FeatureEnabled(role *database.AdminRole, key string) bool {
	if role == nil || strings.TrimSpace(key) == "" {
		return false
	}
	if role.OwnerRole {
		return true
	}
	if role.FeaturesJSON == "" {
		return false
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(role.FeaturesJSON), &root); err != nil {
		return false
	}

	return boolFeatureValue(root[key])
}

func boolFeatureValue(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "true", "yes", "1", "on", "enabled", "all":
			return true
		}
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return false
}

// ─── Access Control ──────────────────────────────────────────────────

// AccessScope represents what an admin can access.
type AccessScope struct {
	AllowAllGroups    bool
	AllowedGroups     []string
	AllowAllInbounds  bool
	AllowedInboundIDs []int
}

// GetAccessScope extracts access scope from role JSON.
func GetAccessScope(role *database.AdminRole) AccessScope {
	var scope AccessScope
	if role == nil {
		return scope
	}
	if role.OwnerRole {
		scope.AllowAllGroups = true
		scope.AllowAllInbounds = true
		return scope
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(role.AccessJSON), &root); err != nil {
		return scope
	}

	// Groups
	if v, ok := root["allowAllGroups"]; ok {
		scope.AllowAllGroups = boolFeatureValue(v)
	}
	if v, ok := root["allowedGroups"]; ok {
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

	// Inbounds
	if v, ok := root["allowAllInbounds"]; ok {
		scope.AllowAllInbounds = boolFeatureValue(v)
	}
	if v, ok := root["allowedInboundIds"]; ok {
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

	return scope
}

// GroupAllowed checks if a group is allowed by the access scope.
func (s AccessScope) GroupAllowed(group string) bool {
	if s.AllowAllGroups || len(s.AllowedGroups) == 0 {
		return true
	}
	key := strings.ToLower(strings.TrimSpace(group))
	for _, g := range s.AllowedGroups {
		if strings.ToLower(strings.TrimSpace(g)) == key {
			return true
		}
	}
	return false
}

// InboundAllowed checks if an inbound ID is allowed by the access scope.
func (s AccessScope) InboundAllowed(inboundID int) bool {
	if s.AllowAllInbounds || len(s.AllowedInboundIDs) == 0 {
		return true
	}
	for _, id := range s.AllowedInboundIDs {
		if id == inboundID {
			return true
		}
	}
	return false
}

// ─── Role Limits ─────────────────────────────────────────────────────

// RoleLimits defines constraints for admins with this role.
type RoleLimits struct {
	MaxUsers        int64 // 0 = unlimited
	MinDataLimit    int64 // bytes
	MaxDataLimit    int64 // bytes
	MinExpireDays   int64
	MaxExpireDays   int64
}

// GetRoleLimits extracts limits from role JSON.
func GetRoleLimits(role *database.AdminRole) RoleLimits {
	var limits RoleLimits
	if role == nil || role.LimitsJSON == "" {
		return limits
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(role.LimitsJSON), &root); err != nil {
		return limits
	}

	if v, ok := root["maxUsers"]; ok {
		limits.MaxUsers = toInt64(v)
	}
	if v, ok := root["minDataLimit"]; ok {
		limits.MinDataLimit = int64(toFloat64(v) * 1024 * 1024 * 1024) // GB to bytes
	}
	if v, ok := root["maxDataLimit"]; ok {
		limits.MaxDataLimit = int64(toFloat64(v) * 1024 * 1024 * 1024)
	}
	if v, ok := root["minExpireDays"]; ok {
		limits.MinExpireDays = toInt64(v)
	}
	if v, ok := root["maxExpireDays"]; ok {
		limits.MaxExpireDays = toInt64(v)
	}
	// New UI-style keys (bytes, seconds)
	if v, ok := root["data_limit_min"]; ok {
		limits.MinDataLimit = toInt64(v)
	}
	if v, ok := root["data_limit_max"]; ok {
		limits.MaxDataLimit = toInt64(v)
	}
	if v, ok := root["expire_min"]; ok {
		limits.MinExpireDays = toInt64(v) / (24 * 60 * 60) // seconds to days
	}
	if v, ok := root["expire_max"]; ok {
		limits.MaxExpireDays = toInt64(v) / (24 * 60 * 60)
	}

	return limits
}

func toFloat64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	}
	return 0
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	}
	return 0
}
