// Package tunnelmonitor provides outbound tunnel health monitoring with automatic recovery.
package tunnelmonitor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"vortexuipro/internal/util/netsafe"
	"vortexuipro/internal/util/random"
)

const (
	defaultHealthURL        = "https://www.cloudflare.com/cdn-cgi/trace"
	defaultInterval         = 30 * time.Second
	defaultTimeout          = 10 * time.Second
	defaultFailureThreshold = 3
	defaultCooldown         = 5 * time.Minute
)

// Config controls the optional tunnel health monitor.
type Config struct {
	Enabled          bool
	URL              string
	ProxyURL         string
	Interval         time.Duration
	Timeout          time.Duration
	FailureThreshold int
	Cooldown         time.Duration

	// RecoveryFunc performs recovery after the failure threshold is reached.
	// Typically wired to restart the Xray core.
	RecoveryFunc func(ctx context.Context) error
}

// Monitor periodically checks the tunnel health and triggers recovery on failure.
type Monitor struct {
	cfg    Config
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu            sync.Mutex
	failCount     int
	cooldownUntil time.Time
	healthy       bool

	// Reusable HTTP client (created once in New)
	httpClient *http.Client

	// OnStatusChange is called whenever the overall status flips.
	OnStatusChange func(healthy bool, msg string)

	// OnRecovery is called when recovery is triggered.
	OnRecovery func(err error)
}

// New creates a new tunnel monitor.
func New(cfg Config) *Monitor {
	if cfg.URL == "" {
		cfg.URL = defaultHealthURL
	}
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInterval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = defaultFailureThreshold
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = defaultCooldown
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		cfg:    cfg,
		id:     "tm-" + random.NumLower(8),
		ctx:    ctx,
		cancel: cancel,
		healthy: true,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: netsafe.SSRFGuardedDialContext,
			},
			Timeout: cfg.Timeout,
		},
	}
}

// Start begins the health check loop.
func (m *Monitor) Start() {
	if !m.cfg.Enabled {
		return
	}
	m.wg.Add(1)
	go m.loop()
}

// Stop terminates the health check loop.
func (m *Monitor) Stop() {
	m.cancel()
	m.wg.Wait()
}

// Healthy returns the current tunnel health status.
func (m *Monitor) Healthy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.healthy
}

// Status returns a detailed status snapshot.
func (m *Monitor) Status() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]any{
		"id":              m.id,
		"enabled":         m.cfg.Enabled,
		"healthy":         m.healthy,
		"fail_count":      m.failCount,
		"failure_threshold": m.cfg.FailureThreshold,
		"cooldown_until":  m.cooldownUntil.UnixMilli(),
		"url":             m.cfg.URL,
		"interval_ms":     m.cfg.Interval.Milliseconds(),
	}
}

func (m *Monitor) loop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cfg.Interval)
	defer ticker.Stop()

	// Run first check immediately
	m.check()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.check()
		}
	}
}

func (m *Monitor) check() {
	ctx, cancel := context.WithTimeout(m.ctx, m.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.URL, nil)
	if err != nil {
		m.recordFailure(fmt.Sprintf("request creation failed: %v", err))
		return
	}

	// Reuse the SSRF-guarded HTTP client created in New()
	client := m.httpClient

	if m.cfg.ProxyURL != "" {
		proxyReq, err := http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.ProxyURL, nil)
		if err == nil {
			resp, proxyErr := client.Do(proxyReq)
			if proxyErr == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				// Check if proxy is alive
				if resp.StatusCode >= 200 && resp.StatusCode < 500 {
					m.recordSuccess("proxy tunnel healthy")
					return
				}
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			m.recordFailure(fmt.Sprintf("health check timed out (%v)", m.cfg.Timeout))
		} else if strings.Contains(err.Error(), "blocked private") {
			m.recordFailure(fmt.Sprintf("SSRF guard blocked: %v", err))
		} else {
			m.recordFailure(fmt.Sprintf("health check failed: %v", err))
		}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		m.recordFailure(fmt.Sprintf("response read failed: %v", err))
		return
	}

	_ = body // Body content available for parsing if needed
	m.recordSuccess(fmt.Sprintf("health check OK (%d)", resp.StatusCode))
}

func (m *Monitor) recordSuccess(msg string) {
	m.mu.Lock()
	m.failCount = 0
	wasHealthy := m.healthy
	m.healthy = true
	m.mu.Unlock()

	if !wasHealthy {
		if m.OnStatusChange != nil {
			m.OnStatusChange(true, msg)
		}
	}
}

func (m *Monitor) recordFailure(msg string) {
	m.mu.Lock()
	m.failCount++
	if m.failCount >= m.cfg.FailureThreshold {
		m.healthy = false
		fc := m.failCount
		m.mu.Unlock()

		// Trigger recovery if not in cooldown
		m.mu.Lock()
		now := time.Now()
		if now.After(m.cooldownUntil) {
			m.cooldownUntil = now.Add(m.cfg.Cooldown)
			m.mu.Unlock()
			go m.recover()
		} else {
			m.mu.Unlock()
		}

		if m.OnStatusChange != nil {
			m.OnStatusChange(false, fmt.Sprintf("tunnel down: %d failures", fc))
		}
	} else {
		m.mu.Unlock()
	}
}

func (m *Monitor) recover() {
	if m.cfg.RecoveryFunc == nil {
		return
	}
	if err := m.cfg.RecoveryFunc(m.ctx); err != nil {
		if m.OnRecovery != nil {
			m.OnRecovery(err)
		}
		_ = fmt.Errorf("recovery failed: %w", err)
	}
}

// DefaultRecovery creates a recovery function that restarts xray-core.
func DefaultRecovery(restartFn func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if restartFn == nil {
			return errors.New("no restart function provided")
		}
		return restartFn(ctx)
	}
}

// CheckIfRunning checks if this panel instance is likely an x-ui variant
// by looking for common environment indicators.
func CheckIfRunning() bool {
	// Check for common x-ui or panel environment variables
	if os.Getenv("X_UI_HOST") != "" || os.Getenv("X_UI_PORT") != "" {
		return true
	}
	if os.Getenv("VORTEX_CLUSTER_ENABLED") != "" {
		return true
	}
	return false
}

// ParseHealthURL returns the URL string constructed from scheme and host.
func ParseHealthURL(scheme, host string) string {
	if host == "" {
		return defaultHealthURL
	}
	s := scheme
	if s == "" {
		s = "https"
	}
	return s + "://" + host + "/"
}

// ExtractCFTrace parses the cloudflare /cdn-cgi/trace response body
// and returns selected fields.
func ExtractCFTrace(body string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "ip", "loc", "colo", "tls", "sni", "warp", "gateway", "h2", "http2", "http3":
			result[key] = val
		}
	}
	return result
}

// ParseRate parses a string representation of a rate (e.g., "5m") to a duration.
// Supported suffixes: s (seconds), m (minutes), h (hours).
func ParseRate(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultInterval, nil
	}
	suffix := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid rate %q: %w", s, err)
	}
	switch suffix {
	case 's':
		return time.Duration(num) * time.Second, nil
	case 'm':
		return time.Duration(num) * time.Minute, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid rate suffix %q in %q", suffix, s)
	}
}
