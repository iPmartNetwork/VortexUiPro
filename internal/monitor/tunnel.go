package monitor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
}

// RecoveryFunc performs recovery after repeated failures.
type RecoveryFunc func(context.Context) error

// Monitor periodically probes a URL and triggers recovery on failures.
type Monitor struct {
	cfg          Config
	client       *http.Client
	recover      RecoveryFunc
	failures     int
	lastRecovery time.Time
	now          func() time.Time
}

// DefaultConfig returns the default (disabled) monitor configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:          false,
		URL:              defaultHealthURL,
		Interval:         defaultInterval,
		Timeout:          defaultTimeout,
		FailureThreshold: defaultFailureThreshold,
		Cooldown:         defaultCooldown,
	}
}

// ConfigFromEnv builds Config from VORTEX_TUNNEL_MONITOR_* environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	cfg.Enabled = parseBool(os.Getenv("VORTEX_TUNNEL_MONITOR"))
	cfg.URL = firstNonEmpty(os.Getenv("VORTEX_TUNNEL_MONITOR_URL"), cfg.URL)
	cfg.ProxyURL = strings.TrimSpace(os.Getenv("VORTEX_TUNNEL_MONITOR_PROXY"))
	cfg.Interval = parseDurationEnv("VORTEX_TUNNEL_MONITOR_INTERVAL", cfg.Interval)
	cfg.Timeout = parseDurationEnv("VORTEX_TUNNEL_MONITOR_TIMEOUT", cfg.Timeout)
	cfg.Cooldown = parseDurationEnv("VORTEX_TUNNEL_MONITOR_COOLDOWN", cfg.Cooldown)
	cfg.FailureThreshold = parseIntEnv("VORTEX_TUNNEL_MONITOR_FAILURES", cfg.FailureThreshold)

	return cfg.Normalize()
}

// Normalize applies safe bounds and defaults.
func (c Config) Normalize() Config {
	if strings.TrimSpace(c.URL) == "" {
		c.URL = defaultHealthURL
	}
	c.URL = strings.TrimSpace(c.URL)
	c.ProxyURL = strings.TrimSpace(c.ProxyURL)

	if c.Interval < time.Second {
		c.Interval = defaultInterval
	}
	if c.Timeout < time.Second {
		c.Timeout = defaultTimeout
	}
	if c.FailureThreshold < 1 {
		c.FailureThreshold = defaultFailureThreshold
	}
	if c.Cooldown < time.Second {
		c.Cooldown = defaultCooldown
	}

	return c
}

// New creates a monitor with an HTTP client.
func New(cfg Config, recover RecoveryFunc) (*Monitor, error) {
	cfg = cfg.Normalize()

	transport := &http.Transport{}
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", cfg.ProxyURL, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &Monitor{
		cfg:     cfg,
		client:  client,
		recover: recover,
		now:     time.Now,
	}, nil
}

// Run starts the monitor loop until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	if m == nil || !m.cfg.Enabled {
		return
	}

	fmt.Printf("Tunnel health monitor enabled: %s\n", m.cfg.URL)

	ticker := time.NewTicker(m.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Tunnel health monitor stopped")
			return
		case <-ticker.C:
			recovered, err := m.Step(ctx)
			if err != nil {
				fmt.Printf("Tunnel health monitor check failed: %v\n", err)
			}
			if recovered {
				fmt.Println("Tunnel health monitor triggered core restart")
			}
		}
	}
}

// Step performs one probe and maybe triggers recovery.
func (m *Monitor) Step(ctx context.Context) (bool, error) {
	if m == nil {
		return false, errors.New("nil monitor")
	}

	if err := m.probe(ctx); err != nil {
		m.failures++

		if m.failures < m.cfg.FailureThreshold {
			return false, fmt.Errorf("probe failed %d/%d: %w", m.failures, m.cfg.FailureThreshold, err)
		}

		now := m.now()
		if !m.lastRecovery.IsZero() && now.Sub(m.lastRecovery) < m.cfg.Cooldown {
			m.failures = m.cfg.FailureThreshold
			return false, fmt.Errorf("probe failed %d/%d; recovery cooldown active: %w",
				m.failures, m.cfg.FailureThreshold, err)
		}

		if m.recover == nil {
			m.failures = m.cfg.FailureThreshold
			return false, errors.New("recovery function is not configured")
		}

		if recErr := m.recover(ctx); recErr != nil {
			return false, fmt.Errorf("recovery failed after probe error %w: %w", err, recErr)
		}

		m.lastRecovery = now
		m.failures = 0
		return true, err
	}

	if m.failures > 0 {
		fmt.Println("Tunnel health monitor recovered after successful probe")
	}
	m.failures = 0
	return false, nil
}

func (m *Monitor) probe(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.URL, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	return nil
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enable", "enabled":
		return true
	default:
		return false
	}
}

func parseDurationEnv(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func parseIntEnv(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func firstNonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
