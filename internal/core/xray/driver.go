package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"vortexuipro/internal/core"
	"vortexuipro/internal/domain"
)

// Config configures the xray-core driver.
type Config struct {
	BinaryPath string
	ConfigPath string
	APIPort    int
}

// Driver implements core.EngineDriver for xray-core with full gRPC integration.
type Driver struct {
	cfg    Config
	cmd    *exec.Cmd
	mu     sync.RWMutex
	status core.CoreStatus
	stopCh chan struct{}
	done   chan struct{}

	// Process management
	logWriter   *LogWriter
	exitErr     error
	startTime   time.Time
	intentional bool
	onlineTrk   *OnlineTracker

	// Config tracking
	currentConfig *Config
	apiPort       int
	statsLast     map[string]int64
}

// New creates a new xray-core driver with full gRPC and process management.
func New(cfg Config) *Driver {
	if cfg.BinaryPath == "" {
		cfg.BinaryPath = "/usr/local/bin/xray"
	}
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = "/etc/vortex/xray.json"
	}
	if cfg.APIPort == 0 {
		cfg.APIPort = 10085
	}
	return &Driver{
		cfg:         cfg,
		status:      core.CoreStopped,
		stopCh:      make(chan struct{}),
		logWriter:   NewLogWriter(),
		onlineTrk:   NewOnlineTracker(2 * time.Minute),
		statsLast:   make(map[string]int64),
		apiPort:     cfg.APIPort,
	}
}

// Name returns the engine name.
func (d *Driver) Name() string { return "xray" }

// Start launches the xray-core process with proper binary detection and logging.
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.status == core.CoreRunning {
		return nil
	}

	binaryPath := d.resolveBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("xray binary not found at %s", binaryPath)
	}

	configPath := d.cfg.ConfigPath
	// Ensure config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("[xray] Config not found at %s, generating default...", configPath)
		defaultCfg := d.generateDefaultConfig()
		if err := os.WriteFile(configPath, defaultCfg, 0644); err != nil {
			return fmt.Errorf("write default config: %w", err)
		}
	}

	cmd := exec.CommandContext(ctx, binaryPath, "run", "-c", configPath)
	cmd.Stdout = d.logWriter
	cmd.Stderr = d.logWriter

	if err := cmd.Start(); err != nil {
		d.status = core.CoreError
		return fmt.Errorf("failed to start xray: %w", err)
	}

	d.cmd = cmd
	d.done = make(chan struct{})
	d.status = core.CoreRunning
	d.startTime = time.Now()
	d.intentional = false

	go d.waitForExit()

	// Wait a moment for xray to initialize, then try gRPC
	time.Sleep(500 * time.Millisecond)

	// Verify gRPC is accessible
	go d.verifyGRPC()

	log.Printf("[xray] Started from %s (pid %d, api port %d)", binaryPath, cmd.Process.Pid, d.cfg.APIPort)
	return nil
}

func (d *Driver) verifyGRPC() {
	time.Sleep(2 * time.Second)
	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := client.queryStats(ctx, "inbound>>>", false)
	if err != nil {
		log.Printf("[xray] gRPC verification: %v (stats may be delayed)", err)
	} else {
		log.Printf("[xray] gRPC API verified on %s", addr)
	}
}

func (d *Driver) resolveBinaryPath() string {
	if d.cfg.BinaryPath != "" {
		if _, err := os.Stat(d.cfg.BinaryPath); err == nil {
			return d.cfg.BinaryPath
		}
	}

	// Search common locations
	paths := []string{
		d.cfg.BinaryPath,
		"/usr/local/bin/xray",
		"/usr/bin/xray",
		"/opt/xray/xray",
	}

	// Auto-detect based on OS/arch
	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "arm32"
	}
	autoPath := fmt.Sprintf("xray-%s-%s", runtime.GOOS, arch)

	// Search in binary folder
	if binDir := filepath.Dir(d.cfg.BinaryPath); binDir != "." {
		paths = append(paths, filepath.Join(binDir, autoPath))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return d.cfg.BinaryPath
}

func (d *Driver) generateDefaultConfig() []byte {
	cfg := map[string]any{
		"log": map[string]any{"loglevel": "warning"},
		"inbounds": []map[string]any{{
			"tag":      "api",
			"port":     d.cfg.APIPort,
			"listen":   "127.0.0.1",
			"protocol": "dokodemo-door",
			"settings": map[string]any{
				"address": "127.0.0.1",
			},
		}},
		"outbounds": []map[string]any{
			{"tag": "direct", "protocol": "freedom", "settings": map[string]any{}},
			{"tag": "block", "protocol": "blackhole", "settings": map[string]any{
				"response": map[string]any{"type": "http"},
			}},
		},
		"routing": map[string]any{
			"domainStrategy": "AsIs",
			"rules": []map[string]any{
				{"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api"},
			},
		},
		"stats": map[string]any{"enabled": true, "interval": "5s"},
		"api": map[string]any{
			"tag":      "api",
			"services": []string{"HandlerService", "StatsService", "RoutingService"},
		},
		"policy": map[string]any{
			"levels": map[string]any{
				"0": map[string]any{
					"statsUserUplink": true, "statsUserDownlink": true,
				},
			},
			"system": map[string]any{
				"statsInboundUplink": true, "statsInboundDownlink": true,
				"statsOutboundUplink": true, "statsOutboundDownlink": true,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return data
}

func (d *Driver) waitForExit() {
	err := d.cmd.Wait()
	d.mu.Lock()
	d.status = core.CoreStopped
	d.exitErr = err
	if d.done != nil {
		close(d.done)
	}
	d.mu.Unlock()

	if err != nil && !d.intentional {
		log.Printf("[xray] Process exited unexpectedly: %v", err)
		log.Printf("[xray] Last logs: %s", d.logWriter.GetCrashReport())
	}
}

// Stop terminates xray-core gracefully (SIGTERM, then SIGKILL after timeout).
func (d *Driver) Stop(ctx context.Context) error {
	d.mu.Lock()
	d.intentional = true
	d.mu.Unlock()

	if d.status != core.CoreRunning || d.cmd == nil || d.cmd.Process == nil {
		d.mu.Lock()
		d.status = core.CoreStopped
		d.mu.Unlock()
		return nil
	}

	// Snapshot cmd without lock
	d.mu.RLock()
	cmd := d.cmd
	done := d.done
	d.mu.RUnlock()

	if runtime.GOOS == "windows" {
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("kill xray: %w", err)
		}
		return d.waitForExitTimeout(2 * time.Second)
	}

	// SIGTERM first
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		if err.Error() == "os: process already finished" {
			d.mu.Lock()
			d.status = core.CoreStopped
			d.mu.Unlock()
			return nil
		}
		// Force kill
		_ = cmd.Process.Kill()
	}

	// Wait for graceful stop
	if err := d.waitForDone(done, 5*time.Second); err == nil {
		d.mu.Lock()
		d.status = core.CoreStopped
		d.mu.Unlock()
		return nil
	}

	// Force kill
	log.Printf("[xray] SIGTERM timed out, killing process")
	_ = cmd.Process.Kill()
	return d.waitForExitTimeout(2 * time.Second)
}

func (d *Driver) waitForDone(done chan struct{}, timeout time.Duration) error {
	if done == nil {
		return nil
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}
}

func (d *Driver) waitForExitTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		d.cmd.Wait()
		close(done)
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		d.mu.Lock()
		d.status = core.CoreStopped
		d.mu.Unlock()
		return nil
	case <-timer.C:
		return fmt.Errorf("timed out waiting for xray exit")
	}
}

// Restart stops and starts xray-core.
func (d *Driver) Restart(ctx context.Context) error {
	if err := d.Stop(ctx); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return d.Start(ctx)
}

// Status returns the current xray-core status.
func (d *Driver) Status(_ context.Context) core.CoreStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status
}

// ApplyConfig writes config to disk and applies hot-reload via gRPC.
func (d *Driver) ApplyConfig(ctx context.Context, config []byte) error {
	if err := os.WriteFile(d.cfg.ConfigPath, config, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Try gRPC hot-reload
	if d.status == core.CoreRunning {
		var cfgMap map[string]any
		if err := json.Unmarshal(config, &cfgMap); err == nil {
			client := newGRPCClient(fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort))
			defer client.close()

			// Apply routing via gRPC if routing section exists
			if routing, ok := cfgMap["routing"]; ok {
				routingJSON, _ := json.Marshal(routing)
				_ = client.applyRouting(ctx, routingJSON)
			}
		}

		// SIGHUP fallback
		d.mu.RLock()
		cmd := d.cmd
		d.mu.RUnlock()
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGHUP)
		}
	}
	return nil
}

// GetConfig reads the current config from disk.
func (d *Driver) GetConfig(_ context.Context) ([]byte, error) {
	return os.ReadFile(d.cfg.ConfigPath)
}

// CollectTraffic retrieves traffic stats from xray's gRPC API with delta tracking.
func (d *Driver) CollectTraffic(ctx context.Context) ([]core.TrafficStats, error) {
	if d.status != core.CoreRunning {
		return nil, nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	stats, err := client.queryStats(ctx, "", false)
	if err != nil {
		return nil, nil // graceful degradation
	}

	// Calculate deltas
	deltas := make(map[string]int64)
	for name, value := range stats {
		lastVal, ok := d.statsLast[name]
		d.statsLast[name] = value
		if ok && value >= lastVal {
			delta := value - lastVal
			if delta > 0 {
				deltas[name] = delta
			}
		}
	}

	// Group by tag
	trafficByTag := statToTrafficStats(deltas)
	var result []core.TrafficStats
	for tag, t := range trafficByTag {
		result = append(result, core.TrafficStats{
			Up:   t.Up,
			Down: t.Down,
			Tag:  tag,
			Time: time.Now(),
		})
	}
	return result, nil
}

// CollectClientTraffic retrieves per-client traffic stats.
func (d *Driver) CollectClientTraffic(ctx context.Context) ([]ClientTraffic, error) {
	if d.status != core.CoreRunning {
		return nil, nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	stats, err := client.queryStats(ctx, "user>>>", false)
	if err != nil {
		return nil, nil
	}

	return getClientTrafficFromStats(stats), nil
}

// GetOnlineUsers returns users with live connections via gRPC.
func (d *Driver) GetOnlineUsers(ctx context.Context) ([]OnlineUser, error) {
	if d.status != core.CoreRunning {
		return nil, fmt.Errorf("xray not running")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.getUsersStats(ctx)
}

// GetBalancerInfo queries a balancer's live state.
func (d *Driver) GetBalancerInfo(ctx context.Context, tag string) (*BalancerInfo, error) {
	if d.status != core.CoreRunning {
		return nil, fmt.Errorf("xray not running")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.getBalancerInfo(ctx, tag)
}

// SetBalancerTarget forces a balancer to a specific outbound.
func (d *Driver) SetBalancerTarget(ctx context.Context, balancerTag, target string) error {
	if d.status != core.CoreRunning {
		return fmt.Errorf("xray not running")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.setBalancerTarget(ctx, balancerTag, target)
}

// TestRoute tests a route through the running core's router.
func (d *Driver) TestRoute(ctx context.Context, req *RouteTestRequest) (*RouteTestResult, error) {
	if d.status != core.CoreRunning {
		return nil, fmt.Errorf("xray not running")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.testRoute(ctx, req)
}

// AddUser adds a user via xray's gRPC API (hot-reload).
func (d *Driver) AddUser(ctx context.Context, tag string, user domain.Client) error {
	if d.status != core.CoreRunning {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	// Determine protocol from the inbound
	protocol := "vmess" // default

	if err := client.addUser(ctx, tag, user.Email, protocol, user.ID); err != nil {
		log.Printf("[xray] gRPC AddUser failed for %s: %v (falling back to SIGHUP)", user.Email, err)
		return d.signalReload()
	}
	return nil
}

// RemoveUser removes a user via xray's gRPC API.
func (d *Driver) RemoveUser(ctx context.Context, tag string, email string) error {
	if d.status != core.CoreRunning {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	if err := client.removeUser(ctx, tag, email); err != nil {
		log.Printf("[xray] gRPC RemoveUser failed for %s: %v (falling back to SIGHUP)", email, err)
		return d.signalReload()
	}
	return nil
}

// ApplyRoutingConfig replaces the routing rules of the running core via gRPC.
func (d *Driver) ApplyRoutingConfig(ctx context.Context, routingJSON []byte) error {
	if d.status != core.CoreRunning {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.applyRouting(ctx, routingJSON)
}

// AddInbound adds an inbound to the running core via gRPC.
func (d *Driver) AddInbound(ctx context.Context, inboundJSON []byte) error {
	if d.status != core.CoreRunning {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.addInbound(ctx, inboundJSON)
}

// RemoveInbound removes an inbound from the running core via gRPC.
func (d *Driver) RemoveInbound(ctx context.Context, tag string) error {
	if d.status != core.CoreRunning {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.APIPort)
	client := newGRPCClient(addr)
	defer client.close()

	return client.removeInbound(ctx, tag)
}

// signalReload sends SIGHUP to trigger config reload.
func (d *Driver) signalReload() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.cmd != nil && d.cmd.Process != nil {
		return d.cmd.Process.Signal(syscall.SIGHUP)
	}
	return nil
}

// HasCapability checks if xray supports the given protocol+transport+security.
func (d *Driver) HasCapability(proto domain.Protocol, transport string, security domain.Security) bool {
	caps := core.XrayCapabilities()
	return core.ContainsProto(caps.Protocols, proto) &&
		core.ContainsStr(caps.Transports, transport) &&
		core.ContainsSec(caps.Securities, security)
}

// GetProcessInfo returns runtime information about the xray process.
func (d *Driver) GetProcessInfo() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	uptime := time.Since(d.startTime).Seconds()

	info := map[string]any{
		"pid":     0,
		"status":  string(d.status),
		"uptime":  fmt.Sprintf("%.0fs", uptime),
		"apiPort": d.cfg.APIPort,
		"binary":  d.cfg.BinaryPath,
		"config":  d.cfg.ConfigPath,
		"online":  d.onlineTrk.GetOnlineCount(),
	}
	if d.cmd != nil && d.cmd.Process != nil {
		info["pid"] = d.cmd.Process.Pid
	}
	if d.exitErr != nil {
		info["lastError"] = d.exitErr.Error()
	}
	return info
}

// GetLogs returns recent log lines from the xray process.
func (d *Driver) GetLogs(n int) []string {
	return d.logWriter.LastLines(n)
}

// ValidateConfig validates an xray JSON config by parsing it.
func ValidateConfig(config []byte) error {
	var v map[string]any
	if err := json.Unmarshal(config, &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required sections
	required := []string{"inbounds", "outbounds"}
	for _, key := range required {
		if _, ok := v[key]; !ok {
			return fmt.Errorf("missing required section: %s", key)
		}
	}

	// Validate inbounds
	if inbounds, ok := v["inbounds"].([]any); ok {
		for i, ib := range inbounds {
			if ibMap, ok := ib.(map[string]any); ok {
				if _, ok := ibMap["protocol"]; !ok {
					return fmt.Errorf("inbound[%d]: missing protocol", i)
				}
			}
		}
	}

	// Validate outbounds
	if outbounds, ok := v["outbounds"].([]any); ok {
		for i, ob := range outbounds {
			if obMap, ok := ob.(map[string]any); ok {
				if _, ok := obMap["protocol"]; !ok {
					return fmt.Errorf("outbound[%d]: missing protocol", i)
				}
			}
		}
	}

	return nil
}

// ─── Utility Helpers ─────────────────────────────────────────────────

// WriteConfigFile atomically writes config data to a file.
func WriteConfigFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".xray-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err := os.Rename(tmpPath, path); err != nil {
			return err
		}
	} else {
		// Windows: remove first then rename
		os.Remove(path)
		if err := os.Rename(tmpPath, path); err != nil {
			return err
		}
	}
	return nil
}

// ─── Xray JSON Config Builder ────────────────────────────────────────

// XrayJSONConfig builds a complete xray JSON config from domain models.
// This is a compatibility wrapper that builds the config using domain types.
func XrayJSONConfig(inbounds []domain.Inbound, outbounds []domain.Outbound, routing *domain.RoutingConfig) ([]byte, error) {
	cfg := buildFullConfig(inbounds, outbounds, routing)
	return json.MarshalIndent(cfg, "", "  ")
}

// buildFullConfig constructs the full xray JSON configuration.
func buildFullConfig(inbounds []domain.Inbound, outbounds []domain.Outbound, routing *domain.RoutingConfig) map[string]any {
	cfg := map[string]any{
		"log": map[string]any{
			"loglevel": "warning",
		},
		"inbounds":  buildInbounds(inbounds),
		"outbounds": buildOutbounds(outbounds),
		"routing":   buildRouting(routing),
		"stats": map[string]any{
			"enabled":   true,
			"interval":  "5s",
		},
		"api": map[string]any{
			"tag":      "api",
			"services": []string{"HandlerService", "StatsService", "RoutingService"},
		},
		"policy": map[string]any{
			"levels": map[string]any{
				"0": map[string]any{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]any{
				"statsInboundUplink":   true,
				"statsInboundDownlink": true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		},
		"dns": map[string]any{
			"servers": []string{"https://1.1.1.1/dns-query", "https://8.8.8.8/dns-query"},
		},
	}

	return cfg
}

func buildInbounds(inbounds []domain.Inbound) []map[string]any {
	out := make([]map[string]any, 0, len(inbounds)+1)
	// Always add API inbound
	out = append(out, map[string]any{
		"tag":    "api",
		"port":   getAPIPort(),
		"listen": "127.0.0.1",
		"protocol": "dokodemo-door",
		"settings": map[string]any{"address": "127.0.0.1"},
	})
	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}
		inbound := map[string]any{
			"tag":             ib.Tag,
			"port":            ib.Port,
			"listen":          ib.Listen,
			"protocol":        string(ib.Protocol),
		}
		if ib.Settings != "" {
			var settings any
			if err := json.Unmarshal([]byte(ib.Settings), &settings); err == nil {
				inbound["settings"] = settings
			} else {
				inbound["settings"] = ib.Settings
			}
		}
		if ib.StreamSettings != "" {
			var stream any
			if err := json.Unmarshal([]byte(ib.StreamSettings), &stream); err == nil {
				inbound["streamSettings"] = stream
			}
		}
		if ib.Sniffing != "" {
			var sniff any
			if err := json.Unmarshal([]byte(ib.Sniffing), &sniff); err == nil {
				inbound["sniffing"] = sniff
			}
		}
		out = append(out, inbound)
	}
	return out
}

func buildOutbounds(outbounds []domain.Outbound) []map[string]any {
	out := make([]map[string]any, 0, len(outbounds)+2)
	for _, ob := range outbounds {
		if !ob.Enable {
			continue
		}
		outbound := map[string]any{
			"tag":      ob.Tag,
			"protocol": string(ob.Protocol),
		}
		if ob.Settings != "" {
			var settings any
			if err := json.Unmarshal([]byte(ob.Settings), &settings); err == nil {
				outbound["settings"] = settings
			}
		}
		if ob.StreamSettings != "" {
			var stream any
			if err := json.Unmarshal([]byte(ob.StreamSettings), &stream); err == nil {
				outbound["streamSettings"] = stream
			}
		}
		out = append(out, outbound)
	}

	// Default outbounds
	out = append(out,
		map[string]any{"tag": "direct", "protocol": "freedom", "settings": map[string]any{}},
		map[string]any{"tag": "block", "protocol": "blackhole", "settings": map[string]any{
			"response": map[string]any{"type": "http"},
		}},
	)
	return out
}

func buildRouting(routing *domain.RoutingConfig) map[string]any {
	if routing == nil {
		return map[string]any{
			"domainStrategy": "AsIs",
			"rules": []map[string]any{
				{"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api"},
			},
		}
	}

	rules := make([]map[string]any, 0, len(routing.Rules)+1)
	// Always add API rule
	rules = append(rules, map[string]any{
		"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api",
	})

	for _, rule := range routing.Rules {
		if !rule.Enabled {
			continue
		}
		r := map[string]any{"type": "field"}
		if len(rule.Domain) > 0 {
			r["domain"] = rule.Domain
		}
		if len(rule.IP) > 0 {
			r["ip"] = rule.IP
		}
		if len(rule.Port) > 0 {
			r["port"] = rule.Port
		}
		if rule.OutboundTag != "" {
			r["outboundTag"] = rule.OutboundTag
		}
		if len(rule.InboundTags) > 0 {
			r["inboundTag"] = rule.InboundTags
		}
		rules = append(rules, r)
	}

	return map[string]any{
		"domainStrategy": routing.DomainStrategy,
		"rules":          rules,
	}
}

var getAPIPort = func() int { return 10085 }
