package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// ─── Constants ───────────────────────────────────────────────────────

const (
	ProbeTypeTCP  = "tcp"
	ProbeTypeHTTP = "http"
	ProbeTypePing = "ping"
	ProbeTypeGRPC = "grpc"

	RecoveryActionRestartCore  = "restart_core"
	RecoveryActionRestartNode  = "restart_node"
	RecoveryActionReboot       = "reboot"
	RecoveryActionWebhook     = "webhook"
	RecoveryActionScript      = "script"

	HealthStatusHealthy   = "healthy"
	HealthStatusWarning   = "warning"
	HealthStatusCritical  = "critical"
	HealthStatusDown      = "down"
	HealthStatusUnknown   = "unknown"
)

// ─── Data Types ──────────────────────────────────────────────────────

// HealthStatus represents the current health state of a target.
type HealthStatus struct {
	ConfigID     int64   `json:"config_id"`
	Name         string  `json:"name"`
	Target       string  `json:"target"`
	ProbeType    string  `json:"probe_type"`
	Status       string  `json:"status"`
	SuccessCount int     `json:"success_count"`
	FailureCount int     `json:"failure_count"`
	Consecutive  int     `json:"consecutive_failures"`
	LastLatency  float64 `json:"last_latency_ms"`
	AvgLatency   float64 `json:"avg_latency_ms"`
	UptimePct    float64 `json:"uptime_pct"`
	LastCheckAt  int64   `json:"last_check_at"`
	LastSuccessAt int64  `json:"last_success_at"`
	LastError    string  `json:"last_error,omitempty"`
	Enabled      bool    `json:"enabled"`
}

// RecoveryResult represents the outcome of a recovery action.
type RecoveryResult struct {
	RuleID    int64  `json:"rule_id"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Status    string `json:"status"` // success, failed
	Message   string `json:"message,omitempty"`
	LatencyMs float64 `json:"latency_ms"`
	Timestamp int64  `json:"timestamp"`
}

// ─── HealthCheckService ──────────────────────────────────────────────

type HealthCheckService struct {
	mu         sync.RWMutex
	eventBus   *events.Bus
	statuses   map[int64]*HealthStatus // configID -> status
	stopCh     chan struct{}
	started    bool

	httpClient *http.Client
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(eventBus *events.Bus) *HealthCheckService {
	return &HealthCheckService{
		eventBus:   eventBus,
		statuses:   make(map[int64]*HealthStatus),
		stopCh:     make(chan struct{}),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Start begins all enabled health checks.
func (s *HealthCheckService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	// Load all enabled configs
	var configs []database.HealthCheckConfig
	if err := database.DB.Where("enabled = ?", true).Find(&configs).Error; err != nil {
		return fmt.Errorf("load health check configs: %w", err)
	}

	for _, cfg := range configs {
		s.statuses[cfg.ID] = &HealthStatus{
			ConfigID:  cfg.ID,
			Name:      cfg.Name,
			Target:    cfg.Target,
			ProbeType: cfg.ProbeType,
			Status:    HealthStatusUnknown,
			Enabled:   cfg.Enabled,
		}
		go s.runCheckLoop(cfg)
	}

	s.started = true
	log.Printf("[Phase 12] Health check service started with %d checks", len(configs))
	return nil
}

// Stop stops all health check loops.
func (s *HealthCheckService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		close(s.stopCh)
		s.started = false
		log.Println("[Phase 12] Health check service stopped")
	}
}

// ─── Health Check Loop ───────────────────────────────────────────────

func (s *HealthCheckService) runCheckLoop(cfg database.HealthCheckConfig) {
	interval := time.Duration(cfg.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first check immediately
	s.runSingleCheck(cfg)

	for {
		select {
		case <-ticker.C:
			s.runSingleCheck(cfg)
		case <-s.stopCh:
			return
		}
	}
}

func (s *HealthCheckService) runSingleCheck(cfg database.HealthCheckConfig) {
	start := time.Now()

	result := &database.HealthCheckResult{
		ConfigID:  cfg.ID,
		Target:    cfg.Target,
		ProbeType: cfg.ProbeType,
	}

	var err error
	switch cfg.ProbeType {
	case ProbeTypeTCP:
		err = s.probeTCP(cfg.Target, cfg.Timeout)
	case ProbeTypeHTTP:
		err = s.probeHTTP(cfg.Target, cfg.Timeout, cfg.ExpectedCode)
	case ProbeTypePing:
		err = s.probePing(cfg.Target, cfg.Timeout)
	default:
		err = s.probeTCP(cfg.Target, cfg.Timeout)
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	result.LatencyMs = latency
	result.Success = err == nil
	if err != nil {
		result.ErrorMsg = err.Error()
	}

	// Save result to database
	database.DB.Create(result)

	// Update in-memory status
	s.mu.Lock()
	status, exists := s.statuses[cfg.ID]
	if !exists {
		status = &HealthStatus{
			ConfigID:  cfg.ID,
			Name:      cfg.Name,
			Target:    cfg.Target,
			ProbeType: cfg.ProbeType,
			Enabled:   cfg.Enabled,
		}
		s.statuses[cfg.ID] = status
	}

	status.LastCheckAt = time.Now().UnixMilli()
	status.LastLatency = latency

	if err == nil {
		status.SuccessCount++
		status.Consecutive = 0
		status.LastSuccessAt = time.Now().UnixMilli()
		status.LastError = ""

		if status.FailureCount == 0 {
			status.Status = HealthStatusHealthy
		} else {
			status.Status = HealthStatusHealthy
		}

		// If was in failure, trigger recovery success event
		if status.FailureCount > 0 {
			s.eventBus.Publish(events.Event{
				Type:    "health.recovered",
				Message: fmt.Sprintf("Health check recovered: %s (%s)", cfg.Name, cfg.Target),
				Data:    map[string]any{"config_id": cfg.ID, "latency_ms": latency},
			})
		}
		status.FailureCount = 0
	} else {
		status.FailureCount++
		status.Consecutive++
		status.LastError = err.Error()

		if status.Consecutive >= cfg.Threshold {
			status.Status = HealthStatusCritical

			// Publish failure event
			s.eventBus.Publish(events.Event{
				Type:    "health.failed",
				Message: fmt.Sprintf("Health check failed: %s (%s) - %s", cfg.Name, cfg.Target, err.Error()),
				Data:    map[string]any{"config_id": cfg.ID, "failures": status.Consecutive, "threshold": cfg.Threshold},
			})

			// Attempt auto-recovery
			go s.attemptRecovery(cfg, err.Error())
		} else {
			status.Status = HealthStatusWarning
		}
	}

	// Calculate uptime
	total := status.SuccessCount + status.FailureCount
	if total > 0 {
		status.UptimePct = float64(status.SuccessCount) / float64(total) * 100
	}

	// Calculate average latency (simple moving average)
	if status.AvgLatency == 0 {
		status.AvgLatency = latency
	} else {
		status.AvgLatency = status.AvgLatency*0.9 + latency*0.1
	}

	s.mu.Unlock()

	// Cleanup old results (keep last 1000 per config)
	var count int64
	database.DB.Model(&database.HealthCheckResult{}).Where("config_id = ?", cfg.ID).Count(&count)
	if count > 1000 {
		database.DB.Where("config_id = ? AND id NOT IN (SELECT id FROM (SELECT id FROM health_check_results WHERE config_id = ? ORDER BY id DESC LIMIT 500) AS tmp)", cfg.ID, cfg.ID).Delete(&database.HealthCheckResult{})
	}
}

// ─── Probe Implementations ───────────────────────────────────────────

func (s *HealthCheckService) probeTCP(target string, timeout int) error {
	timeoutDur := time.Duration(timeout) * time.Second
	conn, err := net.DialTimeout("tcp", target, timeoutDur)
	if err != nil {
		return fmt.Errorf("tcp probe failed: %w", err)
	}
	defer conn.Close()
	return nil
}

func (s *HealthCheckService) probeHTTP(target string, timeout int, expectedCode int) error {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	timeoutDur := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("http probe req failed: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http probe failed: %w", err)
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	if expectedCode > 0 && resp.StatusCode != expectedCode {
		return fmt.Errorf("http probe unexpected status: %d (expected %d)", resp.StatusCode, expectedCode)
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("http probe server error: %d", resp.StatusCode)
	}
	return nil
}

func (s *HealthCheckService) probePing(target string, timeout int) error {
	timeoutDur := time.Duration(timeout) * time.Second
	conn, err := net.DialTimeout("ip:icmp", target, timeoutDur)
	if err != nil {
		// Fallback to TCP ping on common ports
		return s.probeTCP(target+":80", timeout)
	}
	defer conn.Close()
	return nil
}

// ─── Auto-Recovery ──────────────────────────────────────────────────

func (s *HealthCheckService) attemptRecovery(cfg database.HealthCheckConfig, errorMsg string) {
	// Find matching recovery rules
	var rules []database.AutoRecoveryRule
	database.DB.Where("enabled = ?", true).Find(&rules)

	matched := false
	for _, rule := range rules {
		if rule.MatchLabel != "" && !strings.Contains(cfg.Name, rule.MatchLabel) && !strings.Contains(cfg.Target, rule.MatchLabel) {
			continue
		}

		matched = true
		action := &database.AutoRecoveryAction{
			RuleID:     rule.ID,
			CheckID:    cfg.ID,
			ActionType: rule.ActionType,
			Target:     cfg.Target,
			Status:     "pending",
		}

		start := time.Now()
		err := s.executeRecoveryAction(rule.ActionType, rule.ActionParams, cfg.Target)
		latency := float64(time.Since(start).Microseconds()) / 1000.0
		action.LatencyMs = latency

		if err != nil {
			action.Status = "failed"
			action.Result = err.Error()
			log.Printf("[Phase 12] Recovery failed for %s (%s): %v", cfg.Name, rule.ActionType, err)
		} else {
			action.Status = "success"
			action.Result = fmt.Sprintf("Recovery action %s completed successfully", rule.ActionType)
			log.Printf("[Phase 12] Recovery success for %s (%s)", cfg.Name, rule.ActionType)
		}

		database.DB.Create(action)

		s.eventBus.Publish(events.Event{
			Type:    "recovery.executed",
			Message: fmt.Sprintf("Recovery action %s: %s for %s", action.Status, rule.ActionType, cfg.Name),
			Data: map[string]any{
				"rule_id": rule.ID, "action": rule.ActionType,
				"target": cfg.Target, "status": action.Status,
			},
		})
	}

	if !matched {
		log.Printf("[Phase 12] No recovery rule matched for %s, skipping auto-recovery", cfg.Name)
	}
}

func (s *HealthCheckService) executeRecoveryAction(actionType, actionParams, target string) error {
	switch actionType {
	case RecoveryActionRestartCore:
		// Placeholder - would call xray/singbox restart
		log.Printf("[Phase 12] Restarting core for %s", target)
		return nil

	case RecoveryActionRestartNode:
		// Placeholder
		log.Printf("[Phase 12] Restarting node for %s", target)
		return nil

	case RecoveryActionWebhook:
		if actionParams != "" {
			var params map[string]string
			if err := json.Unmarshal([]byte(actionParams), &params); err == nil {
				if url, ok := params["url"]; ok {
					resp, err := s.httpClient.Post(url, "application/json", nil)
					if err != nil {
						return fmt.Errorf("webhook failed: %w", err)
					}
					defer resp.Body.Close()
					io.Copy(io.Discard, resp.Body)
				}
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown recovery action: %s", actionType)
	}
}

// ─── Config Management ──────────────────────────────────────────────

// CreateCheckConfig creates a new health check configuration.
func (s *HealthCheckService) CreateCheckConfig(cfg *database.HealthCheckConfig) error {
	cfg.CreatedAt = time.Now().UnixMilli()
	cfg.UpdatedAt = cfg.CreatedAt
	if err := database.DB.Create(cfg).Error; err != nil {
		return err
	}

	if cfg.Enabled {
		go s.runCheckLoop(*cfg)
	}

	s.eventBus.Publish(events.Event{
		Type:    "health.config_created",
		Message: fmt.Sprintf("Health check config created: %s", cfg.Name),
	})
	return nil
}

// ListCheckConfigs returns all health check configurations.
func (s *HealthCheckService) ListCheckConfigs() ([]database.HealthCheckConfig, error) {
	var configs []database.HealthCheckConfig
	if err := database.DB.Order("name asc").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// UpdateCheckConfig updates a health check configuration.
func (s *HealthCheckService) UpdateCheckConfig(id int64, cfg *database.HealthCheckConfig) error {
	cfg.UpdatedAt = time.Now().UnixMilli()
	return database.DB.Model(&database.HealthCheckConfig{}).Where("id = ?", id).Updates(cfg).Error
}

// DeleteCheckConfig deletes a health check configuration.
func (s *HealthCheckService) DeleteCheckConfig(id int64) error {
	database.DB.Where("config_id = ?", id).Delete(&database.HealthCheckResult{})
	return database.DB.Delete(&database.HealthCheckConfig{}, id).Error
}

// ─── Recovery Rule Management ───────────────────────────────────────

// CreateRecoveryRule creates a new auto-recovery rule.
func (s *HealthCheckService) CreateRecoveryRule(rule *database.AutoRecoveryRule) error {
	rule.CreatedAt = time.Now().UnixMilli()
	rule.UpdatedAt = rule.CreatedAt
	return database.DB.Create(rule).Error
}

// ListRecoveryRules returns all auto-recovery rules.
func (s *HealthCheckService) ListRecoveryRules() ([]database.AutoRecoveryRule, error) {
	var rules []database.AutoRecoveryRule
	if err := database.DB.Order("name asc").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// UpdateRecoveryRule updates an auto-recovery rule.
func (s *HealthCheckService) UpdateRecoveryRule(id int64, rule *database.AutoRecoveryRule) error {
	rule.UpdatedAt = time.Now().UnixMilli()
	return database.DB.Model(&database.AutoRecoveryRule{}).Where("id = ?", id).Updates(rule).Error
}

// DeleteRecoveryRule deletes an auto-recovery rule.
func (s *HealthCheckService) DeleteRecoveryRule(id int64) error {
	return database.DB.Delete(&database.AutoRecoveryRule{}, id).Error
}

// ─── Status & History ───────────────────────────────────────────────

// GetStatuses returns the current health status of all checks.
func (s *HealthCheckService) GetStatuses() []*HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]*HealthStatus, 0, len(s.statuses))
	for _, st := range s.statuses {
		statuses = append(statuses, st)
	}
	return statuses
}

// GetStatus returns the health status for a specific check.
func (s *HealthCheckService) GetStatus(configID int64) *HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statuses[configID]
}

// GetHistory returns recent health check results for a config.
func (s *HealthCheckService) GetHistory(configID int64, limit int) ([]database.HealthCheckResult, error) {
	if limit <= 0 {
		limit = 50
	}
	var results []database.HealthCheckResult
	if err := database.DB.Where("config_id = ?", configID).Order("id desc").Limit(limit).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GetRecoveryHistory returns recent recovery actions.
func (s *HealthCheckService) GetRecoveryHistory(limit int) ([]database.AutoRecoveryAction, error) {
	if limit <= 0 {
		limit = 50
	}
	var actions []database.AutoRecoveryAction
	if err := database.DB.Order("id desc").Limit(limit).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// RunManualCheck triggers a manual health check.
func (s *HealthCheckService) RunManualCheck(configID int64) (*database.HealthCheckResult, error) {
	var cfg database.HealthCheckConfig
	if err := database.DB.First(&cfg, configID).Error; err != nil {
		return nil, err
	}

	start := time.Now()
	result := &database.HealthCheckResult{
		ConfigID:  cfg.ID,
		Target:    cfg.Target,
		ProbeType: cfg.ProbeType,
	}

	var err error
	switch cfg.ProbeType {
	case ProbeTypeTCP:
		err = s.probeTCP(cfg.Target, cfg.Timeout)
	case ProbeTypeHTTP:
		err = s.probeHTTP(cfg.Target, cfg.Timeout, cfg.ExpectedCode)
	default:
		err = s.probeTCP(cfg.Target, cfg.Timeout)
	}

	result.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
	result.Success = err == nil
	if err != nil {
		result.ErrorMsg = err.Error()
	}

	database.DB.Create(result)
	return result, nil
}
