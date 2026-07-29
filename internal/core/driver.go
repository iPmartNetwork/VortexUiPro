package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"vortexuipro/internal/domain"
)

// CoreStatus represents the state of a proxy core engine.
type CoreStatus string

const (
	CoreStopped  CoreStatus = "stopped"
	CoreRunning  CoreStatus = "running"
	CoreStarting CoreStatus = "starting"
	CoreError    CoreStatus = "error"
)

// TrafficStats represents traffic data from a core.
type TrafficStats struct {
	Up       int64
	Down     int64
	Tag      string
	User     string
	Time     time.Time
}

// EngineDriver is the interface that all proxy core engines must implement.
type EngineDriver interface {
	// Name returns the core name (e.g., "xray", "singbox").
	Name() string

	// Start launches the core process.
	Start(ctx context.Context) error

	// Stop terminates the core process gracefully.
	Stop(ctx context.Context) error

	// Restart stops and starts the core.
	Restart(ctx context.Context) error

	// Status returns the current core status.
	Status(ctx context.Context) CoreStatus

	// ApplyConfig writes and applies a new configuration to the core.
	ApplyConfig(ctx context.Context, config []byte) error

	// GetConfig retrieves the currently running configuration.
	GetConfig(ctx context.Context) ([]byte, error)

	// CollectTraffic returns accumulated traffic statistics.
	CollectTraffic(ctx context.Context) ([]TrafficStats, error)

	// AddUser adds or updates a user on a running core (hot-reload).
	AddUser(ctx context.Context, tag string, user domain.Client) error

	// RemoveUser removes a user from a running core.
	RemoveUser(ctx context.Context, tag string, email string) error

	// HasCapability checks if the core supports a given protocol+transport combination.
	HasCapability(proto domain.Protocol, transport string, security domain.Security) bool
}

// EngineManager manages multiple core engines.
type EngineManager struct {
	mu     sync.RWMutex
	engine map[string]EngineDriver
}

// NewEngineManager creates a new engine manager.
func NewEngineManager() *EngineManager {
	return &EngineManager{
		engine: make(map[string]EngineDriver),
	}
}

// Register adds a core engine to the manager.
func (m *EngineManager) Register(driver EngineDriver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.engine[driver.Name()] = driver
}

// Get returns a registered engine by name.
func (m *EngineManager) Get(name string) (EngineDriver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.engine[name]
	if !ok {
		return nil, fmt.Errorf("core engine %q not registered", name)
	}
	return d, nil
}

// StartAll starts all registered engines.
func (m *EngineManager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for name, d := range m.engine {
		if err := d.Start(ctx); err != nil {
			return fmt.Errorf("failed to start core %q: %w", name, err)
		}
	}
	return nil
}

// StopAll gracefully stops all registered engines.
func (m *EngineManager) StopAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for name, d := range m.engine {
		if err := d.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop core %q: %w", name, err)
		}
	}
	return nil
}

// CollectAllTraffic collects traffic from all cores.
func (m *EngineManager) CollectAllTraffic(ctx context.Context) []TrafficStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []TrafficStats
	for _, d := range m.engine {
		stats, err := d.CollectTraffic(ctx)
		if err == nil {
			all = append(all, stats...)
		}
	}
	return all
}

// ─── Shared capability helpers (used by xray and singbox drivers) ───

// ContainsProto checks if a protocol is in a list.
func ContainsProto(list []domain.Protocol, v domain.Protocol) bool {
	for _, p := range list {
		if p == v {
			return true
		}
	}
	return false
}

// ContainsStr checks if a string is in a list.
func ContainsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// ContainsSec checks if a security type is in a list.
func ContainsSec(list []domain.Security, v domain.Security) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// CapabilityMatrix defines what each core supports.
type CapabilityMatrix struct {
	Protocols  []domain.Protocol
	Transports []string
	Securities []domain.Security
}

// XrayCapabilities returns the xray-core capability matrix.
func XrayCapabilities() CapabilityMatrix {
	return CapabilityMatrix{
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks, domain.ProtoSocks, domain.ProtoHTTP, domain.ProtoDokodemo, domain.ProtoHysteria2},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "xhttp", "kcp"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
	}
}

// SingboxCapabilities returns the sing-box capability matrix.
func SingboxCapabilities() CapabilityMatrix {
	return CapabilityMatrix{
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks, domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard, domain.ProtoHysteria, domain.ProtoShadowTLS, domain.ProtoAnyTLS, domain.ProtoSocks, domain.ProtoHTTP, domain.ProtoNaive},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "quic"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
	}
}
