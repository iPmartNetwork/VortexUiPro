package wireguard

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ─── Key Generation ─────────────────────────────────────────────────────

// GeneratePrivateKey generates a new WireGuard private key.
func GeneratePrivateKey() (string, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", fmt.Errorf("generate wireguard key: %w", err)
	}
	return key.String(), nil
}

// GeneratePublicKey derives a public key from a private key.
func GeneratePublicKey(privateKey string) (string, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("parse wireguard private key: %w", err)
	}
	return key.PublicKey().String(), nil
}

// GeneratePSK generates a pre-shared key.
func GeneratePSK() (string, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("generate wireguard psk: %w", err)
	}
	return key.String(), nil
}

// ─── Key Validation ─────────────────────────────────────────────────────

// IsValidPrivateKey validates a WireGuard private key (base64, 32 bytes).
func IsValidPrivateKey(key string) bool {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}

// IsValidPublicKey validates a WireGuard public key (base64, 32 bytes).
func IsValidPublicKey(key string) bool {
	return IsValidPrivateKey(key) // same format
}

// ─── Peer Configuration ────────────────────────────────────────────────

// PeerConfig holds WireGuard peer configuration.
type PeerConfig struct {
	PublicKey    string `json:"publicKey"`
	PresharedKey string `json:"presharedKey,omitempty"`
	AllowedIPs   string `json:"allowedIPs"`
	Endpoint     string `json:"endpoint,omitempty"`
	Persistent   bool   `json:"persistentKeepalive,omitempty"`
}

// InterfaceConfig holds WireGuard interface configuration.
type InterfaceConfig struct {
	PrivateKey string `json:"privateKey"`
	Address    string `json:"address"`
	ListenPort int    `json:"listenPort,omitempty"`
	DNS        string `json:"dns,omitempty"`
	MTU        int    `json:"mtu,omitempty"`
}

// ─── Encode / Decode Helpers ───────────────────────────────────────────

// DecodeHexKey decodes a hex-encoded WireGuard key to base64.
func DecodeHexKey(hexKey string) (string, error) {
	decoded, err := hex.DecodeString(strings.TrimSpace(hexKey))
	if err != nil {
		return "", fmt.Errorf("decode hex key: %w", err)
	}
	if len(decoded) != 32 {
		return "", fmt.Errorf("invalid key length: %d (expected 32)", len(decoded))
	}
	return base64.StdEncoding.EncodeToString(decoded), nil
}

// EncodeHexKey encodes a base64 WireGuard key to hex.
func EncodeHexKey(b64Key string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Key))
	if err != nil {
		return "", fmt.Errorf("decode base64 key: %w", err)
	}
	return hex.EncodeToString(decoded), nil
}

// ─── Xray Config Builder ───────────────────────────────────────────────

// BuildXrayWireGuardOutbound builds an Xray-compatible WireGuard outbound config.
func BuildXrayWireGuardOutbound(
	iface InterfaceConfig,
	peers []PeerConfig,
) map[string]any {
	children := make([]map[string]any, len(peers))
	for i, p := range peers {
		children[i] = map[string]any{
			"publicKey":  p.PublicKey,
			"allowedIPs": strings.Split(p.AllowedIPs, ","),
		}
		if p.PresharedKey != "" {
			children[i]["presharedKey"] = p.PresharedKey
		}
		if p.Endpoint != "" {
			children[i]["endpoint"] = p.Endpoint
		}
	}

	cfg := map[string]any{
		"secretKey": iface.PrivateKey,
		"address":   strings.Split(iface.Address, ","),
		"children":  children,
	}
	if iface.ListenPort > 0 {
		cfg["port"] = iface.ListenPort
	}
	if iface.DNS != "" {
		cfg["dns"] = iface.DNS
	}
	if iface.MTU > 0 {
		cfg["mtu"] = iface.MTU
	}
	if len(peers) > 0 && peers[0].Persistent {
		cfg["keepAlive"] = 25
	}

	return cfg
}

// ─── WireGuard Config Generator ─────────────────────────────────────────

// GenerateConfig generates a complete WireGuard interface configuration with peers.
func GenerateConfig(iface InterfaceConfig, peers []PeerConfig) string {
	var buf bytes.Buffer

	buf.WriteString("[Interface]\n")
	buf.WriteString(fmt.Sprintf("PrivateKey = %s\n", iface.PrivateKey))
	buf.WriteString(fmt.Sprintf("Address = %s\n", iface.Address))
	if iface.ListenPort > 0 {
		buf.WriteString(fmt.Sprintf("ListenPort = %d\n", iface.ListenPort))
	}
	if iface.DNS != "" {
		buf.WriteString(fmt.Sprintf("DNS = %s\n", iface.DNS))
	}
	if iface.MTU > 0 {
		buf.WriteString(fmt.Sprintf("MTU = %d\n", iface.MTU))
	}

	buf.WriteString("\n")
	for _, p := range peers {
		buf.WriteString("[Peer]\n")
		buf.WriteString(fmt.Sprintf("PublicKey = %s\n", p.PublicKey))
		if p.PresharedKey != "" {
			buf.WriteString(fmt.Sprintf("PresharedKey = %s\n", p.PresharedKey))
		}
		if p.AllowedIPs != "" {
			buf.WriteString(fmt.Sprintf("AllowedIPs = %s\n", p.AllowedIPs))
		}
		if p.Endpoint != "" {
			buf.WriteString(fmt.Sprintf("Endpoint = %s\n", p.Endpoint))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}
