package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// ─── Anti-Censorship Suite Types ─────────────────────────────────────

// TLSTrick defines a TLS obfuscation technique.
type TLSTrick struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      any    `json:"config"`
	Enabled     bool   `json:"enabled"`
}

// RealityScanResult holds the result of a REALITY scan.
type RealityScanResult struct {
	Target     string `json:"target"`
	Port       int    `json:"port"`
	Reachable  bool   `json:"reachable"`
	LatencyMS  int    `json:"latency_ms"`
	TLSVersion string `json:"tls_version"`
	ServerName string `json:"server_name"`
	Error      string `json:"error,omitempty"`
}

// DecoyConfig defines a decoy site configuration.
type DecoyConfig struct {
	Domain     string `json:"domain"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	ProxyProto string `json:"proxy_proto"`
	Enable     bool   `json:"enable"`
}

// TLSFingerprint contains available TLS fingerprint options.
type TLSFingerprint struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Security string `json:"security"`
}

// ─── Anti-Censorship Service ─────────────────────────────────────────

// AntiCensorshipService provides anti-censorship tools and configurations.
type AntiCensorshipService struct {
	dataDir string
}

// NewAntiCensorshipService creates a new anti-censorship service.
func NewAntiCensorshipService(dataDir string) *AntiCensorshipService {
	return &AntiCensorshipService{dataDir: dataDir}
}

// ─── TLS Tricks Generator ────────────────────────────────────────────

// GetAvailableTricks returns all available TLS tricks with descriptions.
func (s *AntiCensorshipService) GetAvailableTricks() []TLSTrick {
	return []TLSTrick{
		{
			Name: "random_tls_fingerprint",
			Description: "Randomize TLS fingerprint to avoid deep packet inspection. " +
				"Uses Chrome, Firefox, Safari, or Edge fingerprints randomly.",
			Config: map[string]any{
				"clients": []string{"chrome", "firefox", "safari", "edge", "random"},
				"mode":    "rotate",
			},
			Enabled: true,
		},
		{
			Name: "tls_ech",
			Description: "Enable TLS Encrypted Client Hello (ECH) to hide the SNI. " +
				"Prevents censorship based on domain name inspection.",
			Config: map[string]any{
				"enabled":        true,
				"ech_config":     "",
				"pq_signature":   "mlkem768x25519",
			},
			Enabled: false,
		},
		{
			Name: "padding",
			Description: "Add random padding to TLS handshake packets to bypass " +
				"packet-size-based censorship detection.",
			Config: map[string]any{
				"type":           "random",
				"min_padding":    100,
				"max_padding":    2000,
				"probability":    0.8,
			},
			Enabled: true,
		},
		{
			Name: "mix_https",
			Description: "Mix proxy traffic with normal HTTPS traffic to make it " +
				"indistinguishable from regular web browsing.",
			Config: map[string]any{
				"enabled":            true,
				"https_ratio":        0.3,
				"decoy_sites":        []string{"cloudflare.com", "google.com", "bing.com"},
			},
			Enabled: true,
		},
		{
			Name: "tls_over_tls",
			Description: "Wrap proxy TLS inside another TLS layer. Protects against " +
				"TLS fingerprinting and active probing.",
			Config: map[string]any{
				"enabled":          false,
				"inner_tls":        true,
				"outer_sni":        "www.cloudflare.com",
				"padding_enabled":  true,
			},
			Enabled: false,
		},
		{
			Name: "http_headers_mimic",
			Description: "Mimic standard browser HTTP headers in WebSocket/HTTP upgrades " +
				"to avoid detection by protocol analysis.",
			Config: map[string]any{
				"user_agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"accept_language":  "en-US,en;q=0.9",
				"accept_encoding":  "gzip, deflate, br",
			},
			Enabled: true,
		},
		{
			Name: "fragment",
			Description: "Fragment TLS ClientHello into multiple small packets to bypass " +
				"packet-level censorship (used in China, Iran, etc.).",
			Config: map[string]any{
				"enabled":       true,
				"packets":       2,
				"length":        "100-200",
				"sleep":         "0-25ms",
			},
			Enabled: true,
		},
		{
			Name: "reality_tls",
			Description: "Use REALITY to masquerade as a legitimate website (e.g., " +
				"yahoo.com, microsoft.com) with realistic TLS handshakes.",
			Config: map[string]any{
				"enabled":       true,
				"target":        "www.yahoo.com:443",
				"server_names":  []string{"yahoo.com", "www.yahoo.com"},
				"fingerprint":   "chrome",
				"short_id":      "abcdef",
			},
			Enabled: true,
		},
	}
}

// GetTLSFingerprints returns available TLS fingerprint options.
func (s *AntiCensorshipService) GetTLSFingerprints() []TLSFingerprint {
	return []TLSFingerprint{
		{Name: "Chrome (Auto)", Client: "chrome", Security: "auto"},
		{Name: "Firefox (Auto)", Client: "firefox", Security: "auto"},
		{Name: "Safari (Auto)", Client: "safari", Security: "auto"},
		{Name: "Edge (Auto)", Client: "edge", Security: "auto"},
		{Name: "Random", Client: "random", Security: "auto"},
		{Name: "Chrome 120+", Client: "chrome", Security: "tls13"},
		{Name: "iOS Safari 17+", Client: "safari", Security: "tls13"},
	}
}

// ─── Reality Scanner ─────────────────────────────────────────────────

// ScanTarget checks if a target host is reachable over TLS and REALITY-compatible.
func (s *AntiCensorshipService) ScanTarget(target string, port int) *RealityScanResult {
	result := &RealityScanResult{
		Target: target,
		Port:   port,
	}

	addr := net.JoinHostPort(target, fmt.Sprintf("%d", port))
	start := time.Now()

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	})
	if err != nil {
		result.Error = fmt.Sprintf("connection failed: %v", err)
		return result
	}
	defer conn.Close()

	result.Reachable = true
	result.LatencyMS = int(time.Since(start).Milliseconds())
	result.TLSVersion = tlsVersionString(conn.ConnectionState().Version)
	result.ServerName = conn.ConnectionState().ServerName

	return result
}

func tlsVersionString(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("0x%04X", ver)
	}
}

// ─── Decoy Site Generator ────────────────────────────────────────────

// GenerateDecoyConfig creates a configuration for setting up decoy sites.
func (s *AntiCensorshipService) GenerateDecoyConfig(domain, proxyProto string) *DecoyConfig {
	if domain == "" {
		domain = "cdn.example.com"
	}
	if proxyProto == "" {
		proxyProto = "tls"
	}

	return &DecoyConfig{
		Domain:     domain,
		Port:       443,
		Protocol:   "https",
		ProxyProto: proxyProto,
		Enable:     true,
	}
}

// ─── TLS Certificate Generator (Self-Signed) ─────────────────────────

// GenerateSelfSignedCert creates a self-signed TLS certificate for testing.
func (s *AntiCensorshipService) GenerateSelfSignedCert(domain string) (certPEM, keyPEM []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"VortexUiPro Auto-Generated"},
			CommonName:   domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create cert: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM, nil
}

// ─── Config Generators ───────────────────────────────────────────────

// GenerateFragmentConfig generates xray fragment stream settings.
func (s *AntiCensorshipService) GenerateFragmentConfig() string {
	cfg := map[string]any{
		"packets":  "tlshello",
		"length":   "100-200",
		"interval": "0-25",
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// GeneratePaddingConfig generates xray padding stream settings.
func (s *AntiCensorshipService) GeneratePaddingConfig() string {
	cfg := map[string]any{
		"type":              "random",
		"minPadding":        100,
		"maxPadding":        2000,
		"maxPaddingLen":     2000,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// GenerateMixConfig generates mixed HTTPS configuration.
func (s *AntiCensorshipService) GenerateMixConfig(decoySites []string) string {
	if len(decoySites) == 0 {
		decoySites = []string{"cloudflare.com", "google.com"}
	}
	cfg := map[string]any{
		"decoySites": decoySites,
		"ratio":      0.3,
		"enabled":    true,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// ─── Anti-DPI Enhanced Config ────────────────────────────────────────

// AntiDPIConfig holds the full anti-DPI configuration bundle.
type AntiDPIConfig struct {
	Fragment       map[string]any `json:"fragment"`
	Padding        map[string]any `json:"padding"`
	TLSFingerprint string         `json:"tls_fingerprint"`
	AllowInsecure  bool           `json:"allow_insecure"`
	MixHTTPSTarget string         `json:"mix_https_target"`
}

// GenerateAntiDPIConfig creates a bundled anti-DPI configuration for xray/sing-box.
func (s *AntiCensorshipService) GenerateAntiDPIConfig(transport string) *AntiDPIConfig {
	cfg := &AntiDPIConfig{
		Fragment: map[string]any{
			"packets":  "tlshello",
			"length":   "100-200",
			"interval": "0-25",
		},
		Padding: map[string]any{
			"type":          "random",
			"minPadding":    100,
			"maxPadding":    2000,
		},
		TLSFingerprint: "chrome",
		AllowInsecure:  false,
		MixHTTPSTarget: "www.bing.com:443",
	}

	if transport == "grpc" {
		cfg.Fragment["packets"] = "tlshello"
		cfg.Fragment["length"] = "50-100"
		cfg.Fragment["interval"] = "0-10"
	}

	return cfg
}

// ─── MTProto Proxy Support ───────────────────────────────────────────

// MTProtoConfig defines MTProto proxy configuration.
type MTProtoConfig struct {
	Enabled        bool   `json:"enabled"`
	Port           int    `json:"port"`
	Secret         string `json:"secret"`
	FakeTLS        bool   `json:"fake_tls"`
	FakeTLSDomain  string `json:"fake_tls_domain"`
	Tag            string `json:"tag"`
}

// GenerateMTProtoConfig creates a default MTProto proxy configuration.
func (s *AntiCensorshipService) GenerateMTProtoConfig() *MTProtoConfig {
	return &MTProtoConfig{
		Enabled:       false,
		Port:          1443,
		Secret:        "0123456789abcdef0123456789abcdef",
		FakeTLS:       true,
		FakeTLSDomain: "cloudflare.com",
		Tag:           "mtproto-inbound",
	}
}

// ─── Warp Integration ────────────────────────────────────────────────

// WarpConfig defines Cloudflare WARP integration for routing traffic.
type WarpConfig struct {
	Enabled             bool   `json:"enabled"`
	Mode                string `json:"mode"` // "proxy" or "wireguard"
	WireguardPrivateKey string `json:"wireguard_private_key"`
	WireguardPublicKey  string `json:"wireguard_public_key"`
	WireguardAddressV4  string `json:"wireguard_address_v4"`
	WireguardAddressV6  string `json:"wireguard_address_v6"`
	Endpoint            string `json:"endpoint"`
}

// GenerateWarpConfig creates a default WARP config for routing xray traffic through Cloudflare.
func (s *AntiCensorshipService) GenerateWarpConfig() *WarpConfig {
	return &WarpConfig{
		Enabled:            false,
		Mode:               "wireguard",
		WireguardAddressV4: "172.16.0.2/32",
		WireguardAddressV6: "fd01:5ca1:ab1e:80c9::1/128",
		Endpoint:           "engage.cloudflareclient.com:2408",
	}
}

// SaveCertToFile saves certificate and key to files.
func (s *AntiCensorshipService) SaveCertToFile(certPEM, keyPEM []byte, certPath, keyPath string) error {
	if certPath == "" {
		certPath = s.dataDir + "/vortex.crt"
	}
	if keyPath == "" {
		keyPath = s.dataDir + "/vortex.key"
	}

	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("write key: %w", err)
	}

	return nil
}
