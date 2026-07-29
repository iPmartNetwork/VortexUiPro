package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Constants ───────────────────────────────────────────────────────

const (
	DefaultDoHServer = "https://cloudflare-dns.com/dns-query"
	DefaultDotServer = "tls://1.1.1.1:853"
	DefaultUDPServer = "1.1.1.1:53"

	DNSActionBlock    = "block"
	DNSActionRedirect = "redirect"
	DNSActionProxy    = "proxy"
	DNSActionCustomIP = "custom_ip"
)

// AdBlockLists are commonly used block lists for DNS-level ad blocking.
var AdBlockLists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://someonewhocares.org/hosts/zero/hosts",
	"https://raw.githubusercontent.com/AdguardTeam/AdguardFilters/master/SpyFilter/sections/tracking_servers.txt",
}

// ─── DNS Query Types ─────────────────────────────────────────────────

// DNSQuery represents a DNS query to be resolved.
type DNSQuery struct {
	Domain   string `json:"domain"`
	Type     string `json:"type"` // A, AAAA, CNAME, MX, TXT, etc.
	Protocol string `json:"protocol"` // doh, dot, udp
}

// DNSAnswer represents a DNS resolution result.
type DNSAnswer struct {
	Domain  string `json:"domain"`
	Type    string `json:"type"`
	TTL     int    `json:"ttl"`
	Data    string `json:"data"`
}

// DNSResult holds the complete DNS query result.
type DNSResult struct {
	Query    DNSQuery    `json:"query"`
	Answers  []DNSAnswer `json:"answers"`
	LatencyMS int        `json:"latency_ms"`
	Error    string      `json:"error,omitempty"`
	Cached   bool        `json:"cached"`
}

// ─── SmartDNSService ─────────────────────────────────────────────────

// SmartDNSService provides DNS-over-HTTPS resolution, ad-blocking, DNSSEC, and routing rules.
type SmartDNSService struct {
	mu         sync.RWMutex
	client     *http.Client
	cache      map[string]*DNSResult
	cacheMu    sync.RWMutex
	blocked    map[string]bool
	blockedMu  sync.RWMutex
}

// NewSmartDNSService creates a new smart DNS service.
func NewSmartDNSService() *SmartDNSService {
	return &SmartDNSService{
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  make(map[string]*DNSResult),
		blocked: make(map[string]bool),
	}
}

// ─── DoH Resolution ─────────────────────────────────────────────────

// Resolve resolves a domain using DoH or falls back to UDP.
func (s *SmartDNSService) Resolve(query DNSQuery) *DNSResult {
	start := time.Now()

	// Check cache first
	cacheKey := query.Domain + ":" + query.Type
	s.cacheMu.RLock()
	if cached, ok := s.cache[cacheKey]; ok && cached != nil {
		s.cacheMu.RUnlock()
		result := *cached
		result.Cached = true
		result.LatencyMS = int(time.Since(start).Microseconds()) / 1000
		return &result
	}
	s.cacheMu.RUnlock()

	if query.Protocol == "" {
		query.Protocol = "doh"
	}
	if query.Type == "" {
		query.Type = "A"
	}

	var result *DNSResult
	switch query.Protocol {
	case "doh":
		result = s.resolveDoH(query)
	case "dot":
		result = s.resolveDoT(query)
	default:
		result = s.resolveUDP(query)
	}

	if result != nil {
		result.LatencyMS = int(time.Since(start).Microseconds()) / 1000
		// Cache result
		s.cacheMu.Lock()
		s.cache[cacheKey] = result
		s.cacheMu.Unlock()
	}

	return result
}

// resolveDoH resolves via DNS-over-HTTPS.
func (s *SmartDNSService) resolveDoH(query DNSQuery) *DNSResult {
	// Load active DNS config
	var cfg database.DNSConfig
	if err := database.DB.Where("enabled = ? AND type = 'doh'", true).First(&cfg).Error; err != nil {
		// Use default
		cfg.Upstream = DefaultDoHServer
	}

	url := strings.TrimRight(cfg.Upstream, "/") + "?name=" + query.Domain + "&type=" + query.Type
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return &DNSResult{Query: query, Error: fmt.Sprintf("create request: %v", err)}
	}
	req.Header.Set("Accept", "application/dns-json")

	resp, err := s.client.Do(req)
	if err != nil {
		return &DNSResult{Query: query, Error: fmt.Sprintf("doh request: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &DNSResult{Query: query, Error: fmt.Sprintf("read response: %v", err)}
	}

	// Parse DNS JSON response (RFC 8426 format)
	var dnsResp struct {
		Status   int `json:"Status"`
		Answer   []struct {
			Name string `json:"name"`
			Type int    `json:"type"`
			TTL  int    `json:"TTL"`
			Data string `json:"data"`
		} `json:"Answer"`
	}

	if err := json.Unmarshal(body, &dnsResp); err != nil {
		return &DNSResult{Query: query, Error: fmt.Sprintf("parse response: %v", err)}
	}

	if dnsResp.Status != 0 {
		return &DNSResult{Query: query, Error: fmt.Sprintf("DNS status: %d", dnsResp.Status)}
	}

	result := &DNSResult{Query: query}
	for _, ans := range dnsResp.Answer {
		result.Answers = append(result.Answers, DNSAnswer{
			Domain: ans.Name,
			Type:   dnsTypeString(ans.Type),
			TTL:    ans.TTL,
			Data:   ans.Data,
		})
	}

	// Check against block/redirect rules
	s.applyRules(result)

	return result
}

// resolveDoT resolves via DNS-over-TLS (mock for now, uses UDP with port 853).
func (s *SmartDNSService) resolveDoT(query DNSQuery) *DNSResult {
	// For now, fallback to UDP since DoT requires TLS library
	return s.resolveUDP(query)
}

// resolveUDP resolves via standard UDP DNS.
func (s *SmartDNSService) resolveUDP(query DNSQuery) *DNSResult {
	var cfg database.DNSConfig
	if err := database.DB.Where("enabled = ?", true).First(&cfg).Error; err == nil {
		// Use config's upstream
	}

	// Use system resolver via net.LookupHost
	ips, err := net.LookupHost(query.Domain)
	if err != nil {
		return &DNSResult{Query: query, Error: fmt.Sprintf("lookup failed: %v", err)}
	}

	result := &DNSResult{Query: query}
	for _, ip := range ips {
		result.Answers = append(result.Answers, DNSAnswer{
			Domain: query.Domain,
			Type:   query.Type,
			TTL:    300,
			Data:   ip,
		})
	}

	s.applyRules(result)
	return result
}

// ─── DNS Rules (Blocking, Redirect, Proxy) ──────────────────────────

// applyRules checks and applies DNS rules to a result (blocking, redirecting).
func (s *SmartDNSService) applyRules(result *DNSResult) {
	var rules []database.DNSRule
	if err := database.DB.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return
	}

	for _, rule := range rules {
		if strings.Contains(result.Query.Domain, rule.Domain) || matchesWildcard(result.Query.Domain, rule.Domain) {
			switch rule.Type {
			case DNSActionBlock:
				result.Answers = nil
				result.Error = "blocked by DNS rule"
				return
			case DNSActionRedirect:
				result.Answers = []DNSAnswer{{
					Domain: result.Query.Domain,
					Type:   "A",
					TTL:    60,
					Data:   rule.Action,
				}}
				return
			case DNSActionCustomIP:
				result.Answers = append(result.Answers, DNSAnswer{
					Domain: result.Query.Domain,
					Type:   "A",
					TTL:    300,
					Data:   rule.Action,
				})
			}
		}
	}
}

// ─── Ad Blocking ─────────────────────────────────────────────────────

// LoadAdBlockList loads domains from a block list URL and adds them as block rules.
func (s *SmartDNSService) LoadAdBlockList(url string) (int, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("fetch block list: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read block list: %w", err)
	}

	lines := strings.Split(string(body), "\n")
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		// Extract domain from hosts file format: "127.0.0.1 domain.com"
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] != "" {
			domain := strings.TrimSpace(parts[1])
			if domain != "" && !strings.Contains(domain, "/") {
				rule := &database.DNSRule{
					Domain:  domain,
					Type:    DNSActionBlock,
					Action:  "0.0.0.0",
					Enabled: true,
					Note:    "auto-blocked (ad/tracker)",
				}
				// Upsert
				var existing database.DNSRule
				if err := database.DB.Where("domain = ? AND type = 'block'", domain).First(&existing).Error; err != nil {
					database.DB.Create(rule)
					count++
				}
			}
		}
	}

	return count, nil
}

// LoadDefaultAdBlockLists loads all default ad-blocking block lists.
func (s *SmartDNSService) LoadDefaultAdBlockLists() (int, error) {
	total := 0
	for _, url := range AdBlockLists {
		count, err := s.LoadAdBlockList(url)
		if err != nil {
			log.Printf("[Phase15] Ad block list load error (%s): %v", url, err)
			continue
		}
		total += count
		log.Printf("[Phase15] Loaded %d rules from %s", count, url)
	}
	return total, nil
}

// ─── DNS Config Management ──────────────────────────────────────────

// ListDNSConfigs returns all DNS configurations.
func (s *SmartDNSService) ListDNSConfigs() ([]database.DNSConfig, error) {
	var configs []database.DNSConfig
	if err := database.DB.Order("name asc").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// SaveDNSConfig creates or updates a DNS config.
func (s *SmartDNSService) SaveDNSConfig(cfg *database.DNSConfig) error {
	if cfg.ID > 0 {
		return database.DB.Model(cfg).Updates(cfg).Error
	}
	return database.DB.Create(cfg).Error
}

// DeleteDNSConfig deletes a DNS config.
func (s *SmartDNSService) DeleteDNSConfig(id int64) error {
	return database.DB.Delete(&database.DNSConfig{}, id).Error
}

// ─── DNS Rule Management ────────────────────────────────────────────

// ListDNSRules returns all DNS rules.
func (s *SmartDNSService) ListDNSRules() ([]database.DNSRule, error) {
	var rules []database.DNSRule
	if err := database.DB.Order("domain asc").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// SaveDNSRule creates or updates a DNS rule.
func (s *SmartDNSService) SaveDNSRule(rule *database.DNSRule) error {
	if rule.ID > 0 {
		return database.DB.Model(rule).Updates(rule).Error
	}
	return database.DB.Create(rule).Error
}

// DeleteDNSRule deletes a DNS rule.
func (s *SmartDNSService) DeleteDNSRule(id int64) error {
	return database.DB.Delete(&database.DNSRule{}, id).Error
}

// ─── Helpers ─────────────────────────────────────────────────────────

func dnsTypeString(t int) string {
	switch t {
	case 1:
		return "A"
	case 28:
		return "AAAA"
	case 5:
		return "CNAME"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 33:
		return "SRV"
	case 257:
		return "CAA"
	default:
		return fmt.Sprintf("TYPE%d", t)
	}
}

func matchesWildcard(domain, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return domain == pattern
	}
	parts := strings.SplitN(pattern, "*", 2)
	if len(parts) != 2 {
		return domain == pattern
	}
	return strings.HasPrefix(domain, parts[0]) && strings.HasSuffix(domain, parts[1])
}
