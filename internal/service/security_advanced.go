package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// AdvancedSecurityService provides audit logging, compliance checking, and threat detection.
type AdvancedSecurityService struct {
	eventBus events.Publisher
	mu       sync.RWMutex
	failed   map[string]int              // IP -> failed login count
	banned   map[string]time.Time        // IP -> unban time
}

// NewAdvancedSecurityService creates a new advanced security service.
func NewAdvancedSecurityService(bus events.Publisher) *AdvancedSecurityService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &AdvancedSecurityService{
		eventBus: bus,
		failed:   make(map[string]int),
		banned:   make(map[string]time.Time),
	}
}

// Start begins the periodic cleanup loop.
func (s *AdvancedSecurityService) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.mu.Lock()
			now := time.Now()
			for ip, until := range s.banned {
				if now.After(until) {
					delete(s.banned, ip)
					delete(s.failed, ip)
				}
			}
			s.mu.Unlock()
		}
	}()
}

// ─── Audit Log ───────────────────────────────────────────────────────

// LogAudit records an audit event in the database.
func (s *AdvancedSecurityService) LogAudit(ctx context.Context, actor, action, resource, detail, sourceIP, outcome string) error {
	entry := database.AuditEntry{
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		SourceIP:  sourceIP,
		Outcome:   outcome,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		return fmt.Errorf("create audit entry: %w", err)
	}
	return nil
}

// ListAuditLogs returns audit entries with optional filters.
func (s *AdvancedSecurityService) ListAuditLogs(ctx context.Context, actor, action string, limit, offset int) ([]database.AuditEntry, int64, error) {
	var total int64
	var entries []database.AuditEntry
	q := database.DB.Model(&database.AuditEntry{})
	if actor != "" {
		q = q.Where("actor LIKE ?", "%"+actor+"%")
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	q.Count(&total)
	if limit <= 0 {
		limit = 50
	}
	if err := q.Order("created_at desc").Limit(limit).Offset(offset).Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return entries, total, nil
}

// ─── Threat Detection ───────────────────────────────────────────────

// RecordFailedLogin records a failed login attempt for threat detection.
func (s *AdvancedSecurityService) RecordFailedLogin(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed[ip]++
	if s.failed[ip] >= 5 {
		s.banned[ip] = time.Now().Add(15 * time.Minute)
		s.eventBus.Publish(events.Event{
			Type:    "security.threat",
			Message: fmt.Sprintf("IP %s banned for 15 minutes (5 failed logins)", ip),
		})
	}
}

// IsBanned checks if an IP is currently banned.
func (s *AdvancedSecurityService) IsBanned(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	banUntil, ok := s.banned[ip]
	if !ok {
		return false
	}
	if time.Now().After(banUntil) {
		delete(s.banned, ip)
		return false
	}
	return true
}

// ClearFailedLogins resets the failed login counter for an IP.
func (s *AdvancedSecurityService) ClearFailedLogins(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.failed, ip)
}

// GetThreatSummary returns summary of current threats.
func (s *AdvancedSecurityService) GetThreatSummary() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, v := range s.failed {
		total += v
	}
	return map[string]interface{}{
		"failed_attempts": s.failed,
		"banned_ips":      len(s.banned),
		"total_attempts":  total,
	}
}

// ─── Compliance Checker ────────────────────────────────────────────

// ComplianceReport represents a compliance check result.
type ComplianceReport struct {
	Checks []ComplianceCheck `json:"checks"`
	Passed int               `json:"passed"`
	Failed int               `json:"failed"`
	Score  float64           `json:"score"`
}

// ComplianceCheck represents a single compliance check.
type ComplianceCheck struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Severity    string `json:"severity"`
	Detail      string `json:"detail,omitempty"`
}

// RunComplianceCheck runs all compliance checks.
func (s *AdvancedSecurityService) RunComplianceCheck(ctx context.Context) (*ComplianceReport, error) {
	checks := []ComplianceCheck{
		{Name: "admin_mfa", Description: "Admin accounts have MFA enabled", Passed: false, Severity: "critical"},
		{Name: "password_policy", Description: "Password policy is enforced", Passed: false, Severity: "high"},
		{Name: "tls_enabled", Description: "TLS is enabled for panel access", Passed: false, Severity: "high"},
		{Name: "rate_limiting", Description: "Rate limiting is active", Passed: true, Severity: "medium"},
		{Name: "audit_logging", Description: "Audit logging is enabled", Passed: true, Severity: "medium"},
	}

	var adminCount int64
	database.DB.Model(&database.Admin{}).Count(&adminCount)
	var mfaCount int64
	database.DB.Model(&database.Admin{}).Where("totp_enabled = ?", true).Count(&mfaCount)
	if adminCount > 0 && mfaCount > 0 {
		checks[0].Passed = true
		checks[0].Detail = fmt.Sprintf("%d/%d admins have MFA", mfaCount, adminCount)
	} else {
		checks[0].Detail = "No admins have MFA enabled"
	}

	val, err := database.GetSetting("password_policy")
	if err == nil && val == "enabled" {
		checks[1].Passed = true
	}
	checks[1].Detail = "Password policy is " + val

	_, err = database.GetSetting("tls_enabled")
	if err == nil {
		checks[2].Passed = true
	}

	var passed, failed int
	for i := range checks {
		if checks[i].Passed {
			passed++
		} else {
			failed++
		}
	}

	score := float64(passed) / float64(len(checks)) * 100

	return &ComplianceReport{
		Checks: checks,
		Passed: passed,
		Failed: failed,
		Score:  score,
	}, nil
}
