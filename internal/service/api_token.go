package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"vortexuipro/internal/database"

	"gorm.io/gorm"
)

const (
	apiTokenLength             = 48
	maxPresentedAPITokenLength = 256

	ApiTokenScopeClientsRead       = "clients:read"
	ApiTokenScopeClientsCreate     = "clients:create"
	ApiTokenScopeCustomPanelManage = "custom-panel:manage"
)

var (
	ErrInvalidAPIToken = errors.New("invalid API token")
	apiTokenScopes     = map[string]struct{}{
		ApiTokenScopeClientsRead:       {},
		ApiTokenScopeClientsCreate:     {},
		ApiTokenScopeCustomPanelManage: {},
	}
)

// ApiTokenCreateOptions is the input for creating an API token.
type ApiTokenCreateOptions struct {
	Name             string
	Kind             string
	SubjectAdminID   int
	CreatedByAdminID int
	Scopes           []string
	ExpiresAt        int64
}

// ApiTokenAuthentication contains the resolved identity from a presented token.
type ApiTokenAuthentication struct {
	TokenID   int
	TokenName string
	Kind      string
	Scopes    []string
	Subject   *database.Admin
}

// ApiTokenView is the public representation of an API token.
type ApiTokenView struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Token           string   `json:"token,omitempty"`
	Kind            string   `json:"kind"`
	SubjectAdminID  *int     `json:"subjectAdminId,omitempty"`
	SubjectUsername string   `json:"subjectUsername,omitempty"`
	SubjectRoleName string   `json:"subjectRoleName,omitempty"`
	CreatedByAdminID *int    `json:"createdByAdminId,omitempty"`
	Scopes          []string `json:"scopes"`
	ExpiresAt       int64    `json:"expiresAt"`
	Expired         bool     `json:"expired"`
	Enabled         bool     `json:"enabled"`
	CreatedAt       int64    `json:"createdAt"`
}

// ApiTokenService manages API tokens.
type ApiTokenService struct {
	db *gorm.DB
}

// NewApiTokenService creates a new API token service.
func NewApiTokenService() *ApiTokenService {
	return &ApiTokenService{db: database.DB}
}

func hashTokenSHA256(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func randomSeq(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func normalizeAPITokenKind(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" {
		return database.ApiTokenKindService
	}
	return kind
}

func normalizeAPITokenScopes(scopes []string) ([]string, error) {
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, raw := range scopes {
		scope := strings.ToLower(strings.TrimSpace(raw))
		if scope == "" {
			continue
		}
		if _, allowed := apiTokenScopes[scope]; !allowed {
			return nil, errors.New("unsupported API token scope: " + scope)
		}
		if _, dup := seen[scope]; dup {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	if len(out) == 0 {
		return nil, errors.New("at least one scope is required")
	}
	sort.Strings(out)
	return out, nil
}

func apiTokenCreatedAtSeconds(createdAt int64) int64 {
	if createdAt >= database.ApiTokenMillisecondsThreshold {
		return createdAt / 1000
	}
	return createdAt
}

func apiTokenToView(row *database.ApiToken, subject *database.Admin, role *database.AdminRole) *ApiTokenView {
	kind := normalizeAPITokenKind(row.Kind)
	scopes := apiTokenStoredScopes(row)

	view := &ApiTokenView{
		ID:              row.ID,
		Name:            row.Name,
		Kind:            kind,
		SubjectAdminID:  row.SubjectAdminID,
		CreatedByAdminID: row.CreatedByAdminID,
		Scopes:          scopes,
		ExpiresAt:       row.ExpiresAt,
		Expired:         row.ExpiresAt > 0 && row.ExpiresAt <= time.Now().Unix(),
		Enabled:         row.Enabled,
		CreatedAt:       apiTokenCreatedAtSeconds(row.CreatedAt),
	}
	if subject != nil {
		view.SubjectUsername = subject.Username
	}
	if role != nil {
		view.SubjectRoleName = role.Name
	}
	return view
}

func apiTokenStoredScopes(row *database.ApiToken) []string {
	if row == nil {
		return nil
	}
	if normalizeAPITokenKind(row.Kind) == database.ApiTokenKindService {
		return []string{"*"}
	}
	if row.ScopesJSON == "" {
		return nil
	}
	var scopes []string
	if err := json.Unmarshal([]byte(row.ScopesJSON), &scopes); err != nil {
		return nil
	}
	normalized, err := normalizeAPITokenScopes(scopes)
	if err != nil {
		return nil
	}
	return normalized
}

// List returns all API tokens with resolved subject/role info.
func (s *ApiTokenService) List() ([]*ApiTokenView, error) {
	var rows []database.ApiToken
	if err := s.db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}

	subjectIDs := make([]int, 0, len(rows))
	seenIDs := make(map[int]struct{})
	for _, row := range rows {
		if row.SubjectAdminID == nil || *row.SubjectAdminID <= 0 {
			continue
		}
		if _, seen := seenIDs[*row.SubjectAdminID]; seen {
			continue
		}
		seenIDs[*row.SubjectAdminID] = struct{}{}
		subjectIDs = append(subjectIDs, *row.SubjectAdminID)
	}

	usersByID := make(map[int]*database.Admin, len(subjectIDs))
	rolesByID := make(map[int]*database.AdminRole)
	if len(subjectIDs) > 0 {
		var users []database.Admin
		if err := s.db.Where("id IN ?", subjectIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		roleIDs := make([]int, 0, len(users))
		seenRoleIDs := make(map[int]struct{})
		for i := range users {
			u := &users[i]
			usersByID[int(u.ID)] = u
			if u.RoleID <= 0 {
				continue
			}
			if _, seen := seenRoleIDs[u.RoleID]; seen {
				continue
			}
			seenRoleIDs[u.RoleID] = struct{}{}
			roleIDs = append(roleIDs, u.RoleID)
		}
		if len(roleIDs) > 0 {
			var roles []database.AdminRole
			if err := s.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return nil, err
			}
			for i := range roles {
				rolesByID[roles[i].ID] = &roles[i]
			}
		}
	}

	out := make([]*ApiTokenView, 0, len(rows))
	for _, row := range rows {
		var subject *database.Admin
		var role *database.AdminRole
		if row.SubjectAdminID != nil {
			subject = usersByID[*row.SubjectAdminID]
			if subject != nil {
				role = rolesByID[int(subject.ID)]
			}
		}
		out = append(out, apiTokenToView(&row, subject, role))
	}
	return out, nil
}

// ListDelegatedSubjects returns admins eligible for delegated token assignment.
func (s *ApiTokenService) ListDelegatedSubjects() ([]map[string]any, error) {
	var rows []struct {
		ID       int
		Username string
		RoleID   int
		RoleName string
	}
	err := s.db.Table("admins AS a").
		Select("a.id, a.username, a.role_id, r.name AS role_name").
		Joins("LEFT JOIN admin_roles AS r ON r.id = a.role_id").
		Where("a.status = ? AND (r.owner_role = ? OR r.owner_role IS NULL)", "active", false).
		Order("LOWER(a.username) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id":       row.ID,
			"username": row.Username,
			"roleId":   row.RoleID,
			"roleName": row.RoleName,
		})
	}
	return out, nil
}

// Create creates a service API token (backward compatible).
func (s *ApiTokenService) Create(name string) (*ApiTokenView, error) {
	return s.CreateWithOptions(ApiTokenCreateOptions{
		Name: name,
		Kind: database.ApiTokenKindService,
	})
}

// CreateWithOptions creates an API token with full options.
func (s *ApiTokenService) CreateWithOptions(opts ApiTokenCreateOptions) (*ApiTokenView, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, errors.New("token name is required")
	}
	if utf8.RuneCountInString(name) > 64 {
		return nil, errors.New("token name must be 64 characters or fewer")
	}

	kind := normalizeAPITokenKind(opts.Kind)
	if kind != database.ApiTokenKindService && kind != database.ApiTokenKindDelegated {
		return nil, errors.New("unsupported API token kind")
	}
	if opts.ExpiresAt < 0 || (opts.ExpiresAt > 0 && opts.ExpiresAt <= time.Now().Unix()) {
		return nil, errors.New("token expiry must be in the future")
	}

	var subject *database.Admin
	var role *database.AdminRole
	var subjectID *int
	var scopes []string

	if kind == database.ApiTokenKindDelegated {
		if opts.SubjectAdminID <= 0 {
			return nil, errors.New("delegated token subject is required")
		}
		var sub database.Admin
		if err := s.db.Where("id = ?", opts.SubjectAdminID).First(&sub).Error; err != nil {
			return nil, errors.New("delegated subject not found")
		}
		var r database.AdminRole
		if err := s.db.Where("id = ?", sub.RoleID).First(&r).Error; err != nil {
			return nil, errors.New("subject role not found")
		}
		if r.OwnerRole {
			return nil, errors.New("owner cannot be a delegated token subject")
		}
		subject = &sub
		role = &r

		var scopeErr error
		scopes, scopeErr = normalizeAPITokenScopes(opts.Scopes)
		if scopeErr != nil {
			return nil, scopeErr
		}
		idInt := int(subject.ID)
		subjectID = &idInt
	} else {
		if opts.SubjectAdminID != 0 {
			return nil, errors.New("service tokens cannot have a delegated subject")
		}
		scopes = []string{"*"}
	}

	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return nil, err
	}

	var count int64
	if err := s.db.Model(&database.ApiToken{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("a token with that name already exists")
	}

	prefix := "vrt_s_"
	if kind == database.ApiTokenKindDelegated {
		prefix = "vrt_d_"
	}
	plaintext := prefix + randomSeq(apiTokenLength)

	var createdByID *int
	if opts.CreatedByAdminID > 0 {
		id := opts.CreatedByAdminID
		createdByID = &id
	}

	row := &database.ApiToken{
		Name:            name,
		Token:           hashTokenSHA256(plaintext),
		Kind:            kind,
		SubjectAdminID:  subjectID,
		CreatedByAdminID: createdByID,
		ScopesJSON:      string(scopesJSON),
		ExpiresAt:       opts.ExpiresAt,
		Enabled:         true,
	}
	if err := s.db.Create(row).Error; err != nil {
		return nil, err
	}

	view := apiTokenToView(row, subject, role)
	view.Token = plaintext
	return view, nil
}

// Delete removes an API token.
func (s *ApiTokenService) Delete(id int) error {
	if id <= 0 {
		return errors.New("invalid token id")
	}
	return s.db.Delete(&database.ApiToken{}, id).Error
}

// SetEnabled enables or disables a token.
func (s *ApiTokenService) SetEnabled(id int, enabled bool) error {
	if id <= 0 {
		return errors.New("invalid token id")
	}
	res := s.db.Model(&database.ApiToken{}).Where("id = ?", id).Update("enabled", enabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("token not found")
	}
	return nil
}

// Authenticate validates a presented token and returns the authentication context.
func (s *ApiTokenService) Authenticate(presented string) (*ApiTokenAuthentication, error) {
	if presented == "" || len(presented) > maxPresentedAPITokenLength {
		return nil, ErrInvalidAPIToken
	}

	hash := hashTokenSHA256(presented)
	var matches []database.ApiToken
	if err := s.db.Where("token = ? AND enabled = ?", hash, true).Limit(2).Find(&matches).Error; err != nil {
		return nil, err
	}

	if len(matches) != 1 {
		return nil, ErrInvalidAPIToken
	}

	row := matches[0]
	if row.ExpiresAt > 0 && row.ExpiresAt <= time.Now().Unix() {
		return nil, ErrInvalidAPIToken
	}

	kind := normalizeAPITokenKind(row.Kind)
	scopes := apiTokenStoredScopes(&row)
	if scopes == nil {
		return nil, ErrInvalidAPIToken
	}

	auth := &ApiTokenAuthentication{
		TokenID:   row.ID,
		TokenName: row.Name,
		Kind:      kind,
		Scopes:    scopes,
	}

	if kind == database.ApiTokenKindService {
		return auth, nil
	}
	if kind != database.ApiTokenKindDelegated || row.SubjectAdminID == nil {
		return nil, ErrInvalidAPIToken
	}

	var subject database.Admin
	if err := s.db.Where("id = ?", *row.SubjectAdminID).First(&subject).Error; err != nil {
		return nil, ErrInvalidAPIToken
	}
	auth.Subject = &subject
	return auth, nil
}

// Match is a compatibility wrapper.
func (s *ApiTokenService) Match(presented string) bool {
	_, err := s.Authenticate(presented)
	return err == nil
}


