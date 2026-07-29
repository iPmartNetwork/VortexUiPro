package singbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"vortexuipro/internal/core"
	"vortexuipro/internal/domain"
)

// Config configures the sing-box core driver.
type Config struct {
	BinaryPath string
	ConfigPath string
	WorkDir    string
}

// Driver implements core.EngineDriver for sing-box.
type Driver struct {
	cfg    Config
	cmd    *exec.Cmd
	mu     sync.RWMutex
	status core.CoreStatus
}

// New creates a new sing-box core driver.
func New(cfg Config) *Driver {
	if cfg.BinaryPath == "" {
		cfg.BinaryPath = "/usr/local/bin/sing-box"
	}
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = "/etc/vortex/singbox.json"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = "/etc/vortex"
	}
	return &Driver{
		cfg:    cfg,
		status: core.CoreStopped,
	}
}

// Name returns the engine name.
func (d *Driver) Name() string { return "singbox" }

// Start launches the sing-box process.
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.status == core.CoreRunning {
		return nil
	}

	if _, err := os.Stat(d.cfg.BinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("sing-box binary not found at %s", d.cfg.BinaryPath)
	}

	d.cmd = exec.CommandContext(ctx, d.cfg.BinaryPath, "run", "-c", d.cfg.ConfigPath, "-D", d.cfg.WorkDir)
	d.cmd.Stdout = os.Stdout
	d.cmd.Stderr = os.Stderr

	if err := d.cmd.Start(); err != nil {
		d.status = core.CoreError
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	d.status = core.CoreRunning
	go d.waitForExit()
	return nil
}

func (d *Driver) waitForExit() {
	err := d.cmd.Wait()
	d.mu.Lock()
	d.status = core.CoreStopped
	d.mu.Unlock()
	if err != nil {
		fmt.Printf("sing-box process exited: %v\n", err)
	}
}

// Stop terminates sing-box gracefully.
func (d *Driver) Stop(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.status != core.CoreRunning || d.cmd == nil || d.cmd.Process == nil {
		d.status = core.CoreStopped
		return nil
	}

	// Graceful shutdown: SIGTERM
	if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		_ = d.cmd.Process.Kill()
	}

	done := make(chan struct{})
	go func() {
		_ = d.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		_ = d.cmd.Process.Kill()
		<-done
	case <-time.After(10 * time.Second):
		_ = d.cmd.Process.Kill()
	}

	d.status = core.CoreStopped
	return nil
}

// Restart stops and starts sing-box.
func (d *Driver) Restart(ctx context.Context) error {
	if err := d.Stop(ctx); err != nil {
		return err
	}
	return d.Start(ctx)
}

// Status returns the current sing-box status.
func (d *Driver) Status(_ context.Context) core.CoreStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status
}

// ApplyConfig writes config to disk and reloads sing-box.
// Sing-box doesn't support hot-reload natively, so we restart.
func (d *Driver) ApplyConfig(ctx context.Context, config []byte) error {
	if err := os.WriteFile(d.cfg.ConfigPath, config, 0644); err != nil {
		return fmt.Errorf("failed to write sing-box config: %w", err)
	}
	// Sing-box requires restart to apply config
	return d.Restart(ctx)
}

// GetConfig reads the current config from disk.
func (d *Driver) GetConfig(_ context.Context) ([]byte, error) {
	data, err := os.ReadFile(d.cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sing-box config: %w", err)
	}
	return data, nil
}

// CollectTraffic retrieves traffic stats from sing-box.
// Sing-box doesn't have a native gRPC API, returns empty for now.
func (d *Driver) CollectTraffic(_ context.Context) ([]core.TrafficStats, error) {
	// Sing-box has no native stats API comparable to xray.
	// Traffic data is gathered via database records instead.
	return nil, nil
}

// AddUser adds a user. Sing-box requires config rewrite + restart.
func (d *Driver) AddUser(_ context.Context, _ string, _ domain.Client) error {
	return fmt.Errorf("sing-box does not support hot user management; restart required")
}

// RemoveUser removes a user. Sing-box requires config rewrite + restart.
func (d *Driver) RemoveUser(_ context.Context, _ string, _ string) error {
	return fmt.Errorf("sing-box does not support hot user management; restart required")
}

// HasCapability checks if sing-box supports the given protocol+transport+security.
func (d *Driver) HasCapability(proto domain.Protocol, transport string, security domain.Security) bool {
	caps := core.SingboxCapabilities()
	return core.ContainsProto(caps.Protocols, proto) &&
		core.ContainsStr(caps.Transports, transport) &&
		core.ContainsSec(caps.Securities, security)
}

// ─── Sing-box JSON Config Builder ────────────────────────────────────

// JSONConfig builds a complete sing-box JSON config from domain models.
func JSONConfig(inbounds []domain.Inbound, outbounds []domain.Outbound, routing *domain.RoutingConfig) ([]byte, error) {
	cfg := map[string]any{
		"log": map[string]any{
			"level":   "warn",
			"output":  "",
			"timestamp": true,
		},
		"inbounds":  buildSingboxInbounds(inbounds),
		"outbounds": buildSingboxOutbounds(outbounds),
		"route":     buildSingboxRouting(routing),
		"experimental": map[string]any{
			"cache_file": map[string]any{
				"enabled": true,
				"path":    "/etc/vortex/singbox_cache.db",
			},
		},
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func buildSingboxInbounds(inbounds []domain.Inbound) []map[string]any {
	out := make([]map[string]any, 0, len(inbounds))
	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}
		inbound := map[string]any{
			"type":        string(ib.Protocol),
			"tag":         ib.Tag,
			"listen":      ib.Listen,
			"listen_port": ib.Port,
		}
		// Parse settings JSON if available
		if ib.Settings != "" {
			var settings map[string]any
			if err := json.Unmarshal([]byte(ib.Settings), &settings); err == nil {
				inbound["users"] = settings["clients"]
				if password, ok := settings["password"]; ok {
					inbound["password"] = password
				}
			}
		}
		// Parse stream settings
		if ib.StreamSettings != "" {
			var stream map[string]any
			if err := json.Unmarshal([]byte(ib.StreamSettings), &stream); err == nil {
				inbound["tls"] = stream
			}
		}
		out = append(out, inbound)
	}
	return out
}

func buildSingboxOutbounds(outbounds []domain.Outbound) []map[string]any {
	out := make([]map[string]any, 0, len(outbounds)+3)

	// Add default outbounds
	defaultOutbounds := []map[string]any{
		{"type": "direct", "tag": "direct"},
		{"type": "block", "tag": "block"},
		{"type": "dns", "tag": "dns-out"},
	}
	out = append(out, defaultOutbounds...)

	for _, ob := range outbounds {
		if !ob.Enable {
			continue
		}
		outbound := map[string]any{
			"type": ob.Protocol,
			"tag":  ob.Tag,
		}
		if ob.Settings != "" {
			var settings map[string]any
			if err := json.Unmarshal([]byte(ob.Settings), &settings); err == nil {
				outbound["server"] = settings["address"]
				outbound["server_port"] = settings["port"]
			}
		}
		out = append(out, outbound)
	}
	return out
}

func buildSingboxRouting(routing *domain.RoutingConfig) map[string]any {
	defaultRules := []map[string]any{
		{
			"rule_set":      []string{"geosite-cn"},
			"outbound":      "block",
			"domain_strategy": "",
		},
		{
			"inbound":       []string{"dns-in"},
			"outbound":      "dns-out",
		},
	}

	if routing == nil {
		return map[string]any{
			"rules":      defaultRules,
			"final":      "direct",
			"auto_detect_interface": true,
		}
	}

	rules := defaultRules
	for _, rule := range routing.Rules {
		if !rule.Enabled {
			continue
		}
		r := map[string]any{
			"outbound": rule.OutboundTag,
		}
		if len(rule.Domain) > 0 {
			r["domain"] = rule.Domain
		}
		if len(rule.IP) > 0 {
			r["ip_cidr"] = rule.IP
		}
		rules = append(rules, r)
	}

	return map[string]any{
		"rules":                rules,
		"final":                "direct",
		"auto_detect_interface": true,
	}
}
