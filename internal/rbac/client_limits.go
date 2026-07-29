package rbac

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vortexuipro/internal/database"
)

// ClientRoleLimits defines constraints from the admin's role for client operations.
type ClientRoleLimits struct {
	MaxUsers         int64
	MinDataLimit     int64 // bytes
	MaxDataLimit     int64 // bytes
	MinExpireDays    int64
	MaxExpireDays    int64
	MinOnHoldTimeout int64 // days
	MaxOnHoldTimeout int64 // days
}

// GetClientRoleLimits extracts limits from the admin's role.
func GetClientRoleLimits(adminID int64) ClientRoleLimits {
	role, err := database.GetAdminRole(adminID)
	if err != nil {
		return ClientRoleLimits{}
	}

	if role.OwnerRole {
		return ClientRoleLimits{}
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(role.LimitsJSON), &root); err != nil {
		return ClientRoleLimits{}
	}

	var limits ClientRoleLimits

	if v, ok := root["maxUsers"]; ok {
		limits.MaxUsers = toInt64(v)
	}

	// Legacy style (GB/days)
	if v, ok := root["minDataLimit"]; ok && toFloat64(v) > 0 {
		limits.MinDataLimit = int64(toFloat64(v) * 1024 * 1024 * 1024)
	}
	if v, ok := root["maxDataLimit"]; ok && toFloat64(v) > 0 {
		limits.MaxDataLimit = int64(toFloat64(v) * 1024 * 1024 * 1024)
	}
	if v, ok := root["minExpireDays"]; ok {
		limits.MinExpireDays = toInt64(v)
	}
	if v, ok := root["maxExpireDays"]; ok {
		limits.MaxExpireDays = toInt64(v)
	}

	// New UI style (bytes/seconds)
	if v, ok := root["data_limit_min"]; ok && toInt64(v) > 0 {
		limits.MinDataLimit = toInt64(v)
	}
	if v, ok := root["data_limit_max"]; ok && toInt64(v) > 0 {
		limits.MaxDataLimit = toInt64(v)
	}
	if v, ok := root["expire_min"]; ok && toInt64(v) > 0 {
		limits.MinExpireDays = toInt64(v) / (24 * 60 * 60)
	}
	if v, ok := root["expire_max"]; ok && toInt64(v) > 0 {
		limits.MaxExpireDays = toInt64(v) / (24 * 60 * 60)
	}

	return limits
}

// ValidateClientCreate validates a client creation/update against role limits.
// Returns nil if valid, or an error describing the violation.
func ValidateClientCreate(adminID int64, totalGB int64, expiryTime int64) error {
	limits := GetClientRoleLimits(adminID)
	if limits.MaxUsers > 0 {
		var count int64
		if err := database.DB.Model(&database.Client{}).Where("owner_admin_id = ?", adminID).Count(&count).Error; err == nil {
			if count >= limits.MaxUsers {
				return fmt.Errorf("client limit reached: max %d clients allowed", limits.MaxUsers)
			}
		}
	}

	hasDataLimit := limits.MinDataLimit > 0 || limits.MaxDataLimit > 0
	hasExpireLimit := limits.MinExpireDays > 0 || limits.MaxExpireDays > 0

	if hasDataLimit {
		if totalGB <= 0 {
			return fmt.Errorf("data limit is required for this account")
		}
		if limits.MinDataLimit > 0 && totalGB < limits.MinDataLimit {
			return fmt.Errorf("minimum data limit is %s", formatBytes(limits.MinDataLimit))
		}
		if limits.MaxDataLimit > 0 && totalGB > limits.MaxDataLimit {
			return fmt.Errorf("maximum data limit is %s", formatBytes(limits.MaxDataLimit))
		}
	}

	if hasExpireLimit {
		if expiryTime <= 0 {
			return fmt.Errorf("expiry time is required for this account")
		}
		now := time.Now().UnixMilli()
		durationDays := (expiryTime - now) / (24 * 60 * 60 * 1000)
		if durationDays <= 0 {
			return fmt.Errorf("expiry time must be in the future")
		}
		if limits.MinExpireDays > 0 && durationDays < limits.MinExpireDays {
			return fmt.Errorf("minimum expiry is %d days", limits.MinExpireDays)
		}
		if limits.MaxExpireDays > 0 && durationDays > limits.MaxExpireDays {
			return fmt.Errorf("maximum expiry is %d days", limits.MaxExpireDays)
		}
	}

	return nil
}

func formatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d %s", bytes, units[i])
	}
	return fmt.Sprintf("%.1f %s", size, units[i])
}

// ValidateGroupAccess checks if the admin's role allows accessing the given group.
func ValidateGroupAccess(adminID int64, group string) error {
	scope := GetClientAccessScopeForAdmin(adminID, "create")
	if !scope.RestrictGroups || scope.AllowAllGroups {
		return nil
	}
	if len(scope.AllowedGroups) == 0 {
		return fmt.Errorf("no groups allowed for this account")
	}
	groupKey := strings.ToLower(strings.TrimSpace(group))
	for _, g := range scope.AllowedGroups {
		if strings.ToLower(strings.TrimSpace(g)) == groupKey {
			return nil
		}
	}
	return fmt.Errorf("group %q is not allowed for this account", group)
}

// ValidateInboundAccess checks if the admin's role allows using the given inbound.
func ValidateInboundAccess(adminID int64, inboundID int64) error {
	scope := GetClientAccessScopeForAdmin(adminID, "create")
	if !scope.RestrictInbounds || scope.AllowAllInbounds {
		return nil
	}
	if len(scope.AllowedInboundIDs) == 0 {
		return fmt.Errorf("no inbounds allowed for this account")
	}
	for _, id := range scope.AllowedInboundIDs {
		if int64(id) == inboundID {
			return nil
		}
	}
	return fmt.Errorf("inbound %d is not allowed for this account", inboundID)
}
