package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Constants ───────────────────────────────────────────────────────

// CDNProvider represents a known CDN provider.
type CDNProvider string

const (
	CDNCloudflare CDNProvider = "cloudflare"
	CDNFastly     CDNProvider = "fastly"
	CDNAkamai     CDNProvider = "akamai"
	CDNCloudFront CDNProvider = "cloudfront"
)

// KnownCDNFrontable holds known frontable domains per CDN provider.
type KnownCDNFrontable struct {
	Provider CDNProvider `json:"provider"`
	Domain   string      `json:"domain"`
	IP       string      `json:"ip,omitempty"`
}

// FrontableDomainResult holds the scan result for a potential frontable domain.
type FrontableDomainResult struct {
	Domain      string `json:"domain"`
	Provider    string `json:"provider"`
	Frontable   bool   `json:"frontable"`
	Reachable   bool   `json:"reachable"`
	LatencyMS   int    `json:"latency_ms"`
	TLSVersion  string `json:"tls_version"`
	ServerName  string `json:"server_name"`
	Error       string `json:"error,omitempty"`
}

// CDNProxyConfig holds generated proxy config for a CDN-fronted setup.
type CDNProxyConfig struct {
	Provider      string `json:"provider"`
	FrontDomain   string `json:"front_domain"`
	HiddenDomain  string `json:"hidden_domain"`
	Port          int    `json:"port"`
	TLS           bool   `json:"tls"`
	SNI           string `json:"sni"`
	HostHeader    string `json:"host_header"`
	XrayOutbound  any    `json:"xray_outbound,omitempty"`
	SingboxOutbound any `json:"singbox_outbound,omitempty"`
}

// ─── DomainFrontingService ───────────────────────────────────────────

// DomainFrontingService provides CDN fronting discovery and config generation.
type DomainFrontingService struct {
	mu        sync.RWMutex
	knownFrontables []KnownCDNFrontable
}

// NewDomainFrontingService creates a new domain fronting service.
func NewDomainFrontingService() *DomainFrontingService {
	return &DomainFrontingService{
		knownFrontables: defaultFrontableDomains(),
	}
}

func defaultFrontableDomains() []KnownCDNFrontable {
	return []KnownCDNFrontable{
		// Cloudflare (most common for fronting)
		{Provider: CDNCloudflare, Domain: "cloudflare.com"},
		{Provider: CDNCloudflare, Domain: "www.cloudflare.com"},
		{Provider: CDNCloudflare, Domain: "api.cloudflare.com"},
		{Provider: CDNCloudflare, Domain: "cdnjs.cloudflare.com"},
		{Provider: CDNCloudflare, Domain: "developers.cloudflare.com"},
		{Provider: CDNCloudflare, Domain: "cloudflare.net"},
		// Fastly
		{Provider: CDNFastly, Domain: "fastly.com"},
		{Provider: CDNFastly, Domain: "www.fastly.com"},
		{Provider: CDNFastly, Domain: "app.fastly.com"},
		// Akamai
		{Provider: CDNAkamai, Domain: "akamai.com"},
		{Provider: CDNAkamai, Domain: "www.akamai.com"},
		{Provider: CDNAkamai, Domain: "akamaiedge.net"},
		// CloudFront
		{Provider: CDNCloudFront, Domain: "cloudfront.net"},
		{Provider: CDNCloudFront, Domain: "aws.amazon.com"},
	}
}

// ─── CDN Domain Scanner ─────────────────────────────────────────────

// ScanDomain checks if a domain supports fronting by analyzing TLS properties.
func (s *DomainFrontingService) ScanDomain(domain string, cdnProvider string) *FrontableDomainResult {
	result := &FrontableDomainResult{
		Domain:   domain,
		Provider: cdnProvider,
	}

	addr := net.JoinHostPort(domain, "443")
	start := time.Now()

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         domain,
	})
	if err != nil {
		result.Error = fmt.Sprintf("connection failed: %v", err)
		s.saveResult(result)
		return result
	}
	defer conn.Close()

	result.Reachable = true
	result.LatencyMS = int(time.Since(start).Milliseconds())
	state := conn.ConnectionState()
	result.TLSVersion = tlsVersionString(state.Version)
	result.ServerName = state.ServerName

	// A domain is frontable if it:
	// 1. Is reachable on port 443 with TLS
	// 2. Has a CDN edge server name
	// 3. Doesn't have strict SNI matching
	// We detect this by checking if the server name is a CDN edge
	edgeDetected := false
	for _, known := range s.knownFrontables {
		if (known.Provider == CDNProvider(cdnProvider) || cdnProvider == "") && domain == known.Domain {
			edgeDetected = true
			break
		}
	}
	// Also check if server name contains CDN keywords
	if !edgeDetected && state.ServerName != "" {
		cdnKeywords := []string{"cloudflare", "fastly", "akamai", "cloudfront", "edge"}
		for _, kw := range cdnKeywords {
			if containsSubstr(state.ServerName, kw) || containsSubstr(domain, kw) {
				edgeDetected = true
				break
			}
		}
	}

	result.Frontable = result.Reachable && edgeDetected
	s.saveResult(result)
	return result
}

// ScanAllKnown scans all known frontable domains from the list.
func (s *DomainFrontingService) ScanAllKnown() []*FrontableDomainResult {
	var results []*FrontableDomainResult
	for _, known := range s.knownFrontables {
		result := s.ScanDomain(known.Domain, string(known.Provider))
		results = append(results, result)
		time.Sleep(100 * time.Millisecond) // be nice to CDNs
	}
	return results
}

// ScanCustomDomain scans a user-provided domain for fronting suitability.
func (s *DomainFrontingService) ScanCustomDomain(domain string) *FrontableDomainResult {
	// Auto-detect CDN provider
	provider := detectCDNProvider(domain)
	return s.ScanDomain(domain, provider)
}

// ─── CDN Provider Detection ─────────────────────────────────────────

func detectCDNProvider(domain string) string {
	if containsSubstr(domain, "cloudflare") || containsSubstr(domain, "cdnjs") {
		return string(CDNCloudflare)
	}
	if containsSubstr(domain, "fastly") {
		return string(CDNFastly)
	}
	if containsSubstr(domain, "akamai") || containsSubstr(domain, "akamaiedge") {
		return string(CDNAkamai)
	}
	if containsSubstr(domain, "cloudfront") || containsSubstr(domain, "aws") {
		return string(CDNCloudFront)
	}
	return "unknown"
}

// ─── Proxy Config Generation ─────────────────────────────────────────

// GenerateProxyConfig creates a CDN-fronted proxy configuration for xray/sing-box.
func (s *DomainFrontingService) GenerateProxyConfig(frontDomain, hiddenDomain, provider string) *CDNProxyConfig {
	cfg := &CDNProxyConfig{
		Provider:     provider,
		FrontDomain:  frontDomain,
		HiddenDomain: hiddenDomain,
		Port:         443,
		TLS:          true,
		SNI:          frontDomain,
		HostHeader:   frontDomain,
	}

	// Xray outbound config
	cfg.XrayOutbound = map[string]any{
		"tag":       "cdn-fronting",
		"protocol":  "vless",
		"settings":  map[string]any{
			"vnext": []any{map[string]any{
				"address": frontDomain,
				"port":    443,
				"users": []any{map[string]any{
					"id":         "REPLACE_WITH_UUID",
					"encryption": "none",
				}},
			}},
		},
		"streamSettings": map[string]any{
			"network":  "tcp",
			"security": "tls",
			"tlsSettings": map[string]any{
				"serverName":    frontDomain,
				"allowInsecure": false,
				"fingerprint":   "chrome",
			},
			"sockopt": map[string]any{
				"dialerProxy": "freedom",
			},
		},
	}

	// Sing-box outbound config
	cfg.SingboxOutbound = map[string]any{
		"tag":     "cdn-fronting",
		"type":    "vless",
		"server":  frontDomain,
		"server_port": 443,
		"tls": map[string]any{
			"enabled":    true,
			"server_name": frontDomain,
			"utls": map[string]any{
				"enabled":     true,
				"fingerprint": "chrome",
			},
		},
		"transport": map[string]any{
			"type": "ws",
			"headers": map[string]any{
				"Host": hiddenDomain,
			},
		},
	}

	return cfg
}

// ─── Database Helpers ───────────────────────────────────────────────

func (s *DomainFrontingService) saveResult(result *FrontableDomainResult) {
	domain := &database.CDNDomain{
		Domain:      result.Domain,
		CDNProvider: result.Provider,
		Status:      map[bool]string{true: "active", false: "blocked"}[result.Reachable],
		Reachable:   result.Reachable,
		LatencyMS:   result.LatencyMS,
		TLSVersion:  result.TLSVersion,
		ServerName:  result.ServerName,
		Frontable:   result.Frontable,
		LastChecked: time.Now().UnixMilli(),
	}

	// Upsert
	var existing database.CDNDomain
	if err := database.DB.Where("domain = ?", result.Domain).First(&existing).Error; err == nil {
		domain.ID = existing.ID
		database.DB.Model(domain).Updates(map[string]any{
			"cdn_provider": domain.CDNProvider,
			"status":       domain.Status,
			"reachable":    domain.Reachable,
			"latency_ms":   domain.LatencyMS,
			"tls_version":  domain.TLSVersion,
			"server_name":  domain.ServerName,
			"frontable":    domain.Frontable,
			"last_checked": domain.LastChecked,
		})
	} else {
		database.DB.Create(domain)
	}
}

// ListFrontableDomains returns all scanned CDN domains from the database.
func (s *DomainFrontingService) ListFrontableDomains() ([]database.CDNDomain, error) {
	var domains []database.CDNDomain
	if err := database.DB.Order("frontable desc, latency_ms asc").Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

// ListProviders returns the list of known frontable domains (default list).
func (s *DomainFrontingService) ListProviders() []KnownCDNFrontable {
	return s.knownFrontables
}

// ─── Helpers ─────────────────────────────────────────────────────────

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// tlsVersionString is defined in anticensor.go; using alias here
var _ = log.Printf
