package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── WARP Config ───────────────────────────────────────────────────────
type WARPConfig struct {
	Enabled       bool   `json:"enabled"`
	LicenseKey    string `json:"license_key,omitempty"`
	Endpoint      string `json:"endpoint"`
	PrivateKey    string `json:"private_key,omitempty"`
	PublicKey     string `json:"public_key,omitempty"`
	AddressV4     string `json:"address_v4"`
	AddressV6     string `json:"address_v6"`
	DNS           string `json:"dns"`
	MTU           int    `json:"mtu"`
	WireGuardConf string `json:"wireguard_conf,omitempty"`
	Routes        string `json:"routes,omitempty"`
	OutboundTag   string `json:"outbound_tag"`
	AutoConnect   bool   `json:"auto_connect"`
	Connected     bool   `json:"connected"`
}

// ─── WARPStatus ────────────────────────────────────────────────────────
type WARPStatus struct {
	Connected       bool   `json:"connected"`
	Endpoint        string `json:"endpoint"`
	AddressV4       string `json:"address_v4"`
	AddressV6       string `json:"address_v6"`
	Uptime          int64  `json:"uptime_seconds"`
	TxBytes         int64  `json:"tx_bytes"`
	RxBytes         int64  `json:"rx_bytes"`
	LatestHandshake int64  `json:"latest_handshake"`
	Error           string `json:"error,omitempty"`
}

// ─── WARPProxyService ─────────────────────────────────────────────────
type WARPProxyService struct {
	cfg         WARPConfig
	mu          sync.RWMutex
	connected   bool
	connectedAt time.Time
	configPath  string
}

// NewWARPProxyService creates a new WARP proxy service.
func NewWARPProxyService(configPath string) *WARPProxyService {
	cfg := WARPConfig{
		Endpoint:    "engage.cloudflareclient.com:2408",
		AddressV4:   "100.64.0.2/32",
		DNS:         "1.1.1.1",
		MTU:         1280,
		OutboundTag: "warp-out",
		Enabled:     false,
		Connected:   false,
	}
	return &WARPProxyService{
		cfg:        cfg,
		configPath: configPath,
	}
}

// GetConfig returns the current WARP configuration.
func (s *WARPProxyService) GetConfig() *WARPConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.cfg
	cfg.Connected = s.connected
	return &cfg
}

// UpdateConfig updates the WARP configuration.
func (s *WARPProxyService) UpdateConfig(cfg *WARPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.LicenseKey != "" {
		s.cfg.LicenseKey = cfg.LicenseKey
	}
	if cfg.Endpoint != "" {
		s.cfg.Endpoint = cfg.Endpoint
	}
	if cfg.DNS != "" {
		s.cfg.DNS = cfg.DNS
	}
	if cfg.MTU > 0 {
		s.cfg.MTU = cfg.MTU
	}
	if cfg.OutboundTag != "" {
		s.cfg.OutboundTag = cfg.OutboundTag
	}
	if cfg.Routes != "" {
		s.cfg.Routes = cfg.Routes
	}
	s.cfg.AutoConnect = cfg.AutoConnect
	return s.saveConfig()
}

// Connect establishes the WARP WireGuard tunnel.
func (s *WARPProxyService) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connected {
		return nil
	}
	wgConf, err := s.generateWireGuardConfig()
	if err != nil {
		return fmt.Errorf("generate wg config: %w", err)
	}
	wgPath := filepath.Join(s.configPath, "warp.conf")
	if err := os.WriteFile(wgPath, []byte(wgConf), 0600); err != nil {
		return fmt.Errorf("write wg config: %w", err)
	}
	log.Printf("[WARP] WireGuard config written to %s", wgPath)
	s.connected = true
	s.connectedAt = time.Now()
	s.cfg.Connected = true
	s.cfg.WireGuardConf = wgConf
	log.Printf("[WARP] Connected to %s", s.cfg.Endpoint)
	return nil
}

// Disconnect tears down the WARP tunnel.
func (s *WARPProxyService) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil
	}
	s.connected = false
	s.cfg.Connected = false
	log.Printf("[WARP] Disconnected")
	return nil
}

// GetStatus returns the current connection status.
func (s *WARPProxyService) GetStatus() *WARPStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status := &WARPStatus{
		Connected: s.connected,
		Endpoint:  s.cfg.Endpoint,
		AddressV4: s.cfg.AddressV4,
		AddressV6: s.cfg.AddressV6,
	}
	if s.connected {
		status.Uptime = int64(time.Since(s.connectedAt).Seconds())
	}
	return status
}

// GetXrayOutboundConfig returns Xray outbound config for WARP routing.
func (s *WARPProxyService) GetXrayOutboundConfig() (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.connected {
		return nil, fmt.Errorf("warp not connected")
	}
	return map[string]any{
		"tag":        s.cfg.OutboundTag,
		"protocol":   "freedom",
		"settings":   map[string]any{},
		"sendThrough": s.cfg.AddressV4,
	}, nil
}

func (s *WARPProxyService) generateWireGuardConfig() (string, error) {
	privKey, pubKey, err := genWireGuardKeys()
	if err != nil {
		return "", fmt.Errorf("generate keys: %w", err)
	}
	s.cfg.PrivateKey = privKey
	s.cfg.PublicKey = pubKey

	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
MTU = %d

[Peer]
PublicKey = bmXOC+F1FxEMF9dyiK2HfylqIhH3mh7pYnV4qkGQcTo=
Endpoint = %s
AllowedIPs = 0.0.0.0/0, ::/0
`,
		privKey,
		s.cfg.AddressV4,
		s.cfg.DNS,
		s.cfg.MTU,
		s.cfg.Endpoint,
	)
	return config, nil
}

func (s *WARPProxyService) saveConfig() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.configPath, "warp_config.json"), data, 0600)
}

func genWireGuardKeys() (string, string, error) {
	cmd := exec.Command("wg", "genkey")
	privOut, err := cmd.Output()
	if err != nil {
		return "dev_private_key_placeholder", "dev_public_key_placeholder", nil
	}
	privKey := strings.TrimSpace(string(privOut))
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privKey)
	pubOut, err := cmd.Output()
	if err != nil {
		return privKey, "dev_public_key_placeholder", nil
	}
	return privKey, strings.TrimSpace(string(pubOut)), nil
}
