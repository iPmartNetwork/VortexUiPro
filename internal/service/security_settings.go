package service

import (
	"fmt"
	"strings"
	"sync"

	"vortexuipro/internal/database"
)

// SecuritySettingsService manages security-related panel settings.
type SecuritySettingsService struct {
	mu sync.RWMutex
}

// NewSecuritySettingsService creates a new security settings service.
func NewSecuritySettingsService() *SecuritySettingsService {
	return &SecuritySettingsService{}
}

// ─── Password Policy ───────────────────────────────────────────────

// PasswordPolicy defines the requirements for user passwords.
type PasswordPolicy struct {
	MinLength      int  `json:"min_length"`
	RequireUpper   bool `json:"require_upper"`
	RequireLower   bool `json:"require_lower"`
	RequireNumber  bool `json:"require_number"`
	RequireSpecial bool `json:"require_special"`
}

// GetPasswordPolicy returns the current password policy.
func (s *SecuritySettingsService) GetPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:      getSettingInt("password_min_length", 8),
		RequireUpper:   getSettingBool("password_require_upper", true),
		RequireLower:   getSettingBool("password_require_lower", true),
		RequireNumber:  getSettingBool("password_require_number", true),
		RequireSpecial: getSettingBool("password_require_special", false),
	}
}

// ValidatePassword checks a password against the current policy.
func (s *SecuritySettingsService) ValidatePassword(password string) error {
	policy := s.GetPasswordPolicy()
	var errors []string

	if len(password) < policy.MinLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters", policy.MinLength))
	}
	if policy.RequireUpper && !hasUpper(password) {
		errors = append(errors, "password must contain uppercase letter")
	}
	if policy.RequireLower && !hasLower(password) {
		errors = append(errors, "password must contain lowercase letter")
	}
	if policy.RequireNumber && !hasNumber(password) {
		errors = append(errors, "password must contain number")
	}
	if policy.RequireSpecial && !hasSpecial(password) {
		errors = append(errors, "password must contain special character")
	}

	if len(errors) > 0 {
		return fmt.Errorf("password policy: %s", strings.Join(errors, "; "))
	}
	return nil
}

// SavePasswordPolicy saves the password policy settings.
func (s *SecuritySettingsService) SavePasswordPolicy(policy *PasswordPolicy) error {
	settings := map[string]string{
		"password_min_length":    fmt.Sprintf("%d", policy.MinLength),
		"password_require_upper": fmt.Sprintf("%t", policy.RequireUpper),
		"password_require_lower": fmt.Sprintf("%t", policy.RequireLower),
		"password_require_number": fmt.Sprintf("%t", policy.RequireNumber),
		"password_require_special": fmt.Sprintf("%t", policy.RequireSpecial),
	}
	for k, v := range settings {
		if err := database.SetSetting(k, v); err != nil {
			return fmt.Errorf("save %s: %w", k, err)
		}
	}
	return nil
}

// ─── Geo-Blocking ──────────────────────────────────────────────────

// GetGeoBlock returns the list of blocked country codes.
func (s *SecuritySettingsService) GetGeoBlock() []string {
	val, err := database.GetSetting("geo_block_countries")
	if err != nil || val == "" {
		return []string{}
	}
	return strings.Split(val, ",")
}

// SetGeoBlock saves the list of blocked country codes.
func (s *SecuritySettingsService) SetGeoBlock(countries []string) error {
	val := strings.Join(countries, ",")
	return database.SetSetting("geo_block_countries", val)
}

// IsGeoBlocked checks if a country code is blocked.
func (s *SecuritySettingsService) IsGeoBlocked(countryCode string) bool {
	for _, c := range s.GetGeoBlock() {
		if strings.EqualFold(c, countryCode) {
			return true
		}
	}
	return false
}

// ─── IP Ban / Whitelist ────────────────────────────────────────────

// GetBannedIPs returns the list of banned IPs.
func (s *SecuritySettingsService) GetBannedIPs() []string {
	return s.getIPList("banned_ips")
}

// GetWhitelistedIPs returns the list of whitelisted IPs.
func (s *SecuritySettingsService) GetWhitelistedIPs() []string {
	return s.getIPList("whitelisted_ips")
}

// AddBannedIP adds an IP to the ban list.
func (s *SecuritySettingsService) AddBannedIP(ip, reason string) error {
	return s.addIP("banned_ips", ip, reason)
}

// RemoveBannedIP removes an IP from the ban list.
func (s *SecuritySettingsService) RemoveBannedIP(ip string) error {
	return s.removeIP("banned_ips", ip)
}

// AddWhitelistedIP adds an IP to the whitelist.
func (s *SecuritySettingsService) AddWhitelistedIP(ip, reason string) error {
	return s.addIP("whitelisted_ips", ip, reason)
}

// RemoveWhitelistedIP removes an IP from the whitelist.
func (s *SecuritySettingsService) RemoveWhitelistedIP(ip string) error {
	return s.removeIP("whitelisted_ips", ip)
}

// GetFailedLoginThreshold returns the threshold for auto-ban.
func (s *SecuritySettingsService) GetFailedLoginThreshold() int {
	return getSettingInt("failed_login_threshold", 5)
}

// GetBanDuration returns the ban duration in minutes.
func (s *SecuritySettingsService) GetBanDuration() int {
	return getSettingInt("ban_duration_minutes", 15)
}

// ─── Helpers ───────────────────────────────────────────────────────

func (s *SecuritySettingsService) getIPList(key string) []string {
	val, err := database.GetSetting(key)
	if err != nil || val == "" {
		return []string{}
	}
	var ips []string
	for _, item := range strings.Split(val, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			ips = append(ips, item)
		}
	}
	return ips
}

func (s *SecuritySettingsService) addIP(key, ip, reason string) error {
	existing := s.getIPList(key)
	for _, e := range existing {
		if strings.EqualFold(e, ip) {
			return nil // already exists
		}
	}
	entry := ip
	if reason != "" {
		entry = ip + ":" + reason
	}
	existing = append(existing, entry)
	return database.SetSetting(key, strings.Join(existing, ","))
}

func (s *SecuritySettingsService) removeIP(key, ip string) error {
	existing := s.getIPList(key)
	var updated []string
	for _, e := range existing {
		parts := strings.SplitN(e, ":", 2)
		if !strings.EqualFold(parts[0], ip) {
			updated = append(updated, e)
		}
	}
	return database.SetSetting(key, strings.Join(updated, ","))
}

// ─── Utility ───────────────────────────────────────────────────────

func getSettingInt(key string, defaultVal int) int {
	val, err := database.GetSetting(key)
	if err != nil {
		return defaultVal
	}
	var result int
	fmt.Sscanf(val, "%d", &result)
	return result
}

func getSettingBool(key string, defaultVal bool) bool {
	val, err := database.GetSetting(key)
	if err != nil {
		return defaultVal
	}
	return val == "true"
}

func hasUpper(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func hasLower(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func hasNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func hasSpecial(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return true
		}
	}
	return false
}
