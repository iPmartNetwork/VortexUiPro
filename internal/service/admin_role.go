package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"vortexuipro/internal/database"

	"gorm.io/gorm"
)

// AdminRoleService manages role-based access control roles.
type AdminRoleService struct {
	db *gorm.DB
}

// NewAdminRoleService creates a new role service.
func NewAdminRoleService() *AdminRoleService {
	return &AdminRoleService{db: database.DB}
}

// AdminRolePayload is the input for creating/updating a role.
type AdminRolePayload struct {
	Name        string         `json:"name"`
	Permissions map[string]any `json:"permissions"`
	Limits      map[string]any `json:"limits"`
	Features    map[string]any `json:"features"`
	Access      map[string]any `json:"access"`
}

// AdminRoleView is the public representation of a role.
type AdminRoleView struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	BuiltIn     bool           `json:"builtIn"`
	OwnerRole   bool           `json:"ownerRole"`
	Permissions map[string]any `json:"permissions"`
	Limits      map[string]any `json:"limits"`
	Features    map[string]any `json:"features"`
	Access      map[string]any `json:"access"`
	AdminCount  int64          `json:"adminCount"`
	CreatedAt   int64          `json:"createdAt"`
	UpdatedAt   int64          `json:"updatedAt"`
}

// normalizeRoleSlug creates a URL-safe slug from a name.
func normalizeRoleSlug(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
		if !ok {
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
			continue
		}
		if r == '-' {
			if lastDash {
				continue
			}
			lastDash = true
		} else {
			lastDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func marshalJSONMap(v map[string]any) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func decodeJSONMap(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]any{}
	}
	return out
}

func roleToView(row *database.AdminRole, adminCount int64) *AdminRoleView {
	if row == nil {
		return nil
	}
	return &AdminRoleView{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		BuiltIn:     row.BuiltIn,
		OwnerRole:   row.OwnerRole,
		Permissions: decodeJSONMap(row.PermissionsJSON),
		Limits:      decodeJSONMap(row.LimitsJSON),
		Features:    decodeJSONMap(row.FeaturesJSON),
		Access:      decodeJSONMap(row.AccessJSON),
		AdminCount:  adminCount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// List returns all roles with admin counts.
func (s *AdminRoleService) List() ([]*AdminRoleView, error) {
	var rows []database.AdminRole
	if err := s.db.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	roleIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		roleIDs = append(roleIDs, row.ID)
	}

	counts := map[int]int64{}
	if len(roleIDs) > 0 {
		type CountRow struct {
			RoleID int
			Count  int64
		}
		var grouped []CountRow
		if err := s.db.Table("admins").
			Select("role_id, COUNT(*) AS count").
			Where("role_id IN ?", roleIDs).
			Group("role_id").
			Scan(&grouped).Error; err != nil {
			return nil, err
		}
		for _, row := range grouped {
			counts[row.RoleID] = row.Count
		}
	}

	out := make([]*AdminRoleView, 0, len(rows))
	for _, row := range rows {
		out = append(out, roleToView(&row, counts[row.ID]))
	}
	return out, nil
}

// Get returns a single role by ID.
func (s *AdminRoleService) Get(id int) (*AdminRoleView, error) {
	if id <= 0 {
		return nil, errors.New("invalid role id")
	}
	var row database.AdminRole
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	var count int64
	s.db.Model(&database.Admin{}).Where("role_id = ?", id).Count(&count)
	return roleToView(&row, count), nil
}

// defaultOperatorRoleMaps returns the default permissions/limits/features/access for a custom role.
func defaultOperatorRoleMaps() (map[string]any, map[string]any, map[string]any, map[string]any) {
	return map[string]any{
			"users": map[string]any{
				"read":        map[string]any{"scope": 1},
				"create":      true,
				"update":      map[string]any{"scope": 1},
				"delete":      map[string]any{"scope": 1},
				"reset_usage": map[string]any{"scope": 1},
			},
			"inbounds": map[string]any{
				"read_simple": true,
			},
			"settings": map[string]any{
				"read_general": true,
			},
		},
		map[string]any{
			"maxUsers":     nil,
			"minDataLimit": nil,
			"maxDataLimit": nil,
		},
		map[string]any{
			"blockLimitedAdmins":          false,
			"disconnectUsersWhenLimited":  true,
			"disconnectUsersWhenDisabled": true,
			"useResetStrategy":            true,
			"useNextPlan":                 true,
		},
		map[string]any{
			"allowAllGroups":   true,
			"allowAllInbounds": true,
		}
}

// Create creates a new custom role.
func (s *AdminRoleService) Create(payload AdminRolePayload) (*AdminRoleView, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return nil, errors.New("role name is required")
	}
	if len(name) > 64 {
		return nil, errors.New("role name must be 64 characters or fewer")
	}
	slug := normalizeRoleSlug(name)
	if slug == "" {
		return nil, errors.New("role slug is invalid")
	}

	var count int64
	if err := s.db.Model(&database.AdminRole{}).
		Where("name = ? OR slug = ?", name, slug).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("role already exists")
	}

	defaultPerms, defaultLimits, defaultFeatures, defaultAccess := defaultOperatorRoleMaps()

	row := &database.AdminRole{
		Name:            name,
		Slug:            slug,
		PermissionsJSON: marshalJSONMap(nonEmptyMap(payload.Permissions, defaultPerms)),
		LimitsJSON:      marshalJSONMap(nonEmptyMap(payload.Limits, defaultLimits)),
		FeaturesJSON:    marshalJSONMap(nonEmptyMap(payload.Features, defaultFeatures)),
		AccessJSON:      marshalJSONMap(nonEmptyMap(payload.Access, defaultAccess)),
	}
	if err := s.db.Create(row).Error; err != nil {
		return nil, err
	}
	return roleToView(row, 0), nil
}

// Update updates a custom role. Built-in roles can only update permissions/limits/features/access.
func (s *AdminRoleService) Update(id int, payload AdminRolePayload) (*AdminRoleView, error) {
	if id <= 0 {
		return nil, errors.New("invalid role id")
	}
	var row database.AdminRole
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	if row.OwnerRole {
		return nil, errors.New("owner role is read-only")
	}

	updates := map[string]any{
		"permissions": marshalJSONMap(payload.Permissions),
		"limits":      marshalJSONMap(payload.Limits),
		"features":    marshalJSONMap(payload.Features),
		"access":      marshalJSONMap(payload.Access),
	}

	if !row.BuiltIn {
		name := strings.TrimSpace(payload.Name)
		if name == "" {
			return nil, errors.New("role name is required")
		}
		if len(name) > 64 {
			return nil, errors.New("role name must be 64 characters or fewer")
		}
		slug := normalizeRoleSlug(name)
		if slug == "" {
			return nil, errors.New("role slug is invalid")
		}
		var count int64
		if err := s.db.Model(&database.AdminRole{}).
			Where("(name = ? OR slug = ?) AND id <> ?", name, slug, id).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("role already exists")
		}
		updates["name"] = name
		updates["slug"] = slug
	}

	if err := s.db.Model(&database.AdminRole{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

// Duplicate creates a copy of a role.
func (s *AdminRoleService) Duplicate(id int) (*AdminRoleView, error) {
	if id <= 0 {
		return nil, errors.New("invalid role id")
	}
	var src database.AdminRole
	if err := s.db.First(&src, id).Error; err != nil {
		return nil, err
	}

	baseName := strings.TrimSpace(src.Name) + " (copy)"
	name := baseName
	for i := 2; ; i++ {
		slug := normalizeRoleSlug(name)
		var count int64
		if err := s.db.Model(&database.AdminRole{}).
			Where("name = ? OR slug = ?", name, slug).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			row := &database.AdminRole{
				Name:            name,
				Slug:            slug,
				PermissionsJSON: src.PermissionsJSON,
				LimitsJSON:      src.LimitsJSON,
				FeaturesJSON:    src.FeaturesJSON,
				AccessJSON:      src.AccessJSON,
			}
			if err := s.db.Create(row).Error; err != nil {
				return nil, err
			}
			return roleToView(row, 0), nil
		}
		name = fmt.Sprintf("%s %d", baseName, i)
	}
}

// Delete removes a custom role. Built-in and owner roles cannot be deleted.
func (s *AdminRoleService) Delete(id int) error {
	if id <= 0 {
		return errors.New("invalid role id")
	}
	var row database.AdminRole
	if err := s.db.First(&row, id).Error; err != nil {
		return err
	}
	if row.OwnerRole {
		return errors.New("owner role cannot be deleted")
	}
	if row.BuiltIn {
		return errors.New("built-in roles cannot be deleted")
	}

	var count int64
	if err := s.db.Model(&database.Admin{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("role is assigned to admins")
	}

	res := s.db.Delete(&database.AdminRole{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("role not found")
	}
	return nil
}

// SeedDefaultRoles creates the three built-in roles if they don't exist.
func SeedDefaultRoles(db *gorm.DB) error {
	var count int64
	db.Model(&database.AdminRole{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaults := []database.AdminRole{
		{
			Name:            "Owner",
			Slug:            "owner",
			BuiltIn:         true,
			OwnerRole:       true,
			PermissionsJSON: `{"users":{"view":"all","create":true,"update":"all","delete":"all"},"admins":{"view":true,"create":true,"update":true,"delete":true},"roles":{"view":true,"create":true,"update":true,"delete":true},"inbounds":{"view":true,"create":true,"update":true,"delete":true},"nodes":{"view":true,"create":true,"update":true,"delete":true},"settings":{"view":true,"update":true},"system":{"view":true}}`,
			LimitsJSON:      `{}`,
			FeaturesJSON:    `{"blockLimitedAdmins":true,"disconnectUsersWhenLimited":true,"disconnectUsersWhenDisabled":true,"useResetStrategy":true,"useNextPlan":true}`,
			AccessJSON:      `{"allowAllGroups":true,"allowAllInbounds":true}`,
		},
		{
			Name:            "Administrator",
			Slug:            "administrator",
			BuiltIn:         true,
			OwnerRole:       false,
			PermissionsJSON: `{"users":{"read":{"scope":1},"create":true,"update":{"scope":1},"delete":{"scope":1},"reset_usage":{"scope":1}},"inbounds":{"read":true,"create":true,"update":true,"delete":true},"nodes":{"read_simple":true},"admins":{"read_simple":true},"settings":{"read_general":true,"update":true}}`,
			LimitsJSON:      `{}`,
			FeaturesJSON:    `{"blockLimitedAdmins":false,"disconnectUsersWhenLimited":true,"disconnectUsersWhenDisabled":true,"useResetStrategy":true,"useNextPlan":true}`,
			AccessJSON:      `{"allowAllGroups":true,"allowAllInbounds":true}`,
		},
		{
			Name:            "Operator",
			Slug:            "operator",
			BuiltIn:         true,
			OwnerRole:       false,
			PermissionsJSON: `{"users":{"read":{"scope":1},"create":true,"update":{"scope":1},"delete":{"scope":1}},"inbounds":{"read_simple":true},"settings":{"read_general":true}}`,
			LimitsJSON:      `{}`,
			FeaturesJSON:    `{"blockLimitedAdmins":false,"disconnectUsersWhenLimited":true,"disconnectUsersWhenDisabled":true,"useResetStrategy":true,"useNextPlan":true}`,
			AccessJSON:      `{"allowAllGroups":true,"allowAllInbounds":true}`,
		},
	}

	for _, role := range defaults {
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

func nonEmptyMap(v map[string]any, fallback map[string]any) map[string]any {
	if len(v) > 0 {
		return v
	}
	return fallback
}

// Ensure interfaces match

