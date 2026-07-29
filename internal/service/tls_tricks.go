package service

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ─── TLSTrickType ──────────────────────────────────────────────────────
type TLSTrickType string

const (
	TrickFragment    TLSTrickType = "fragment"
	TrickPadding     TLSTrickType = "padding"
	TrickMixedCase   TLSTrickType = "mixed_case"
	TrickHello       TLSTrickType = "tls_hello"
	TrickTLSOverTLS  TLSTrickType = "tls_over_tls"
	TrickRandomSNI   TLSTrickType = "random_sni"
)

// ─── TLSTrickConfig ────────────────────────────────────────────────────
type TLSTrickConfig struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Type          TLSTrickType `json:"type"`
	Enabled       bool         `json:"enabled"`
	Description   string       `json:"description,omitempty"`

	// Fragment settings
	FragmentPackets string `json:"fragment_packets,omitempty"` // "tlshello" or "all"
	FragmentLength  string `json:"fragment_length,omitempty"`  // e.g., "10-50"
	FragmentSleep   string `json:"fragment_sleep,omitempty"`   // e.g., "5-15"

	// Padding settings
	PaddingType string `json:"padding_type,omitempty"` // "comment", "packet", "random"
	PaddingSize string `json:"padding_size,omitempty"` // e.g., "100-500"

	// TLS Hello settings
	HelloFingerprint string `json:"hello_fingerprint,omitempty"`
	HelloALPN        string `json:"hello_alpn,omitempty"`

	// TLS over TLS settings
	TOTLayerCount    int  `json:"tot_layer_count,omitempty"`
	TOTRandomDelay   bool `json:"tot_random_delay,omitempty"`

	// Random SNI
	SNIDomains string `json:"sni_domains,omitempty"` // JSON array
}

// ─── TLSProfile ────────────────────────────────────────────────────────
type TLSProfile struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Enabled     bool             `json:"enabled"`
	Tricks      []TLSTrickConfig `json:"tricks"`
	Description string           `json:"description,omitempty"`
	CreatedAt   int64            `json:"created_at"`
}

// ─── TLSTricksService ─────────────────────────────────────────────────
type TLSTricksService struct {
	profiles map[string]*TLSProfile
	mu       sync.RWMutex
}

// NewTLSTricksService creates a new TLS Tricks service.
func NewTLSTricksService() *TLSTricksService {
	svc := &TLSTricksService{
		profiles: make(map[string]*TLSProfile),
	}
	svc.seedDefaults()
	return svc
}

func (s *TLSTricksService) seedDefaults() {
	s.profiles["fragment_basic"] = &TLSProfile{
		ID:      "fragment_basic",
		Name:    "Basic Fragment",
		Enabled: false,
		Description: "Split TLS handshake packets to evade DPI",
		CreatedAt: 0,
		Tricks: []TLSTrickConfig{{
			ID:   "frag_1",
			Name: "TLS Hello Fragment",
			Type: TrickFragment, Enabled: true,
			FragmentPackets: "tlshello",
			FragmentLength:  "50-100",
			FragmentSleep:   "10-20",
		}},
	}
	s.profiles["padding_def"] = &TLSProfile{
		ID:      "padding_def",
		Name:    "Traffic Padding",
		Enabled: false,
		Description: "Add random padding to traffic to confuse packet inspection",
		CreatedAt: 0,
		Tricks: []TLSTrickConfig{{
			ID: "pad_1", Name: "Random Padding",
			Type: TrickPadding, Enabled: true,
			PaddingType: "packet",
			PaddingSize: "100-300",
		}},
	}
	s.profiles["full_stealth"] = &TLSProfile{
		ID:      "full_stealth",
		Name:    "Full Stealth Mode",
		Enabled: false,
		Description: "Combined fragment + padding + TLS fingerprint spoofing",
		CreatedAt: 0,
		Tricks: []TLSTrickConfig{
			{ID: "fs_1", Name: "Fragment", Type: TrickFragment, Enabled: true,
				FragmentPackets: "tlshello", FragmentLength: "40-80", FragmentSleep: "10-30"},
			{ID: "fs_2", Name: "Padding", Type: TrickPadding, Enabled: true,
				PaddingType: "random", PaddingSize: "50-200"},
			{ID: "fs_3", Name: "TLS Fingerprint", Type: TrickHello, Enabled: true,
				HelloFingerprint: "chrome", HelloALPN: "h2,http/1.1"},
		},
	}
}

// ListProfiles returns all TLS trick profiles.
func (s *TLSTricksService) ListProfiles() []*TLSProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*TLSProfile, 0, len(s.profiles))
	for _, p := range s.profiles {
		list = append(list, p)
	}
	return list
}

// GetProfile returns a specific profile.
func (s *TLSTricksService) GetProfile(id string) (*TLSProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profiles[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("profile not found: %s", id)
}

// SaveProfile creates or updates a profile.
func (s *TLSTricksService) SaveProfile(profile *TLSProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles[profile.ID] = profile
	return nil
}

// DeleteProfile removes a profile.
func (s *TLSTricksService) DeleteProfile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profiles, id)
	return nil
}

// EnableProfile enables/disables a profile.
func (s *TLSTricksService) EnableProfile(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.profiles[id]; ok {
		p.Enabled = enabled
		return nil
	}
	return fmt.Errorf("profile not found: %s", id)
}

// GenerateXrayConfig generates Xray stream config with TLS tricks.
func (s *TLSTricksService) GenerateXrayConfig(profileID string) (map[string]any, error) {
	profile, err := s.GetProfile(profileID)
	if err != nil {
		return nil, err
	}

	config := map[string]any{
		"streamSettings": map[string]any{
			"network": "tcp",
			"tcpSettings": map[string]any{},
			"security": "tls",
			"tlsSettings": map[string]any{},
		},
	}

	for _, trick := range profile.Tricks {
		if !trick.Enabled {
			continue
		}
		switch trick.Type {
		case TrickFragment:
			config["streamSettings"].(map[string]any)["tcpSettings"] = map[string]any{
				"header": map[string]any{
					"type": "http",
					"request": map[string]any{
						"version": "1.1",
						"method":  "GET",
						"path":    []string{"/"},
					},
				},
			}
		case TrickPadding:
			config["streamSettings"].(map[string]any)["tcpSettings"] = map[string]any{
				"header": map[string]any{
					"type": "http",
				},
			}
		case TrickHello:
			tlsSettings := config["streamSettings"].(map[string]any)["tlsSettings"].(map[string]any)
			tlsSettings["fingerprint"] = trick.HelloFingerprint
			if trick.HelloALPN != "" {
				tlsSettings["alpn"] = []string{"h2", "http/1.1"}
			}
			config["streamSettings"].(map[string]any)["tlsSettings"] = tlsSettings
		}
	}

	return config, nil
}

// ToJSON returns the service state as JSON.
func (s *TLSTricksService) ToJSON() (string, error) {
	data, err := json.MarshalIndent(s.ListProfiles(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
