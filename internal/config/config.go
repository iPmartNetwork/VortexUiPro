package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment.
type Config struct {
	// Listener
	HTTPAddr  string
	HTTPSAddr string
	GRPCAddr  string

	// Database
	DatabaseURL string
	DBType      string // sqlite or postgres

	// Redis (for caching and rate limiting)
	RedisURL string

	// Authentication
	JWTSecret      string
	JWTTTL         time.Duration
	SessionTimeout time.Duration

	// mTLS for node hub
	TLSCert string
	TLSKey  string
	TLSCA   string

	// Core engines
	Core        string   // xray or singbox
	Cores       []string // enabled cores
	CoreBin     string
	CoreConfig  string
	CoreAPIPort  int
	SingboxBin  string

	// Notifications
	WebhookURL        string
	WebhookSecret     string
	TelegramToken     string
	TelegramChat      string
	TelegramBotToken  string

	// Client Activity
	ActivityEnabled  bool
	ActivityFlushSec int

	// Cloudflare
	CFToken string
	CFZone  string

	// Subscription
	SubPort    int
	SubHost    string
	SubPath    string
	SubSSL     bool
	SubCert    string
	SubKey     string

	// Auto-recovery
	AutoRecover     bool
	RecoverAfter    time.Duration
	RecoverCooldown time.Duration

	// GeoIP
	GeoIPDB string

	// Tunnel Monitor
	TunnelMonitorEnabled bool
	TunnelMonitorURL     string
	TunnelMonitorProxy   string

	// Payments
	ZarinPalMerchant     string
	NowPaymentsAPIKey    string
	NowPaymentsIPNSecret string

	// Logging
	LogLevel   string
	LogFile    string
	LogMaxSize int // megabytes
	LogMaxAge  int // days

	// Cluster / Multi-Node Mesh
	ClusterEnabled     bool
	ClusterNodeName    string
	ClusterAddr        string   // this node's mesh address (ip:port)
	ClusterPeers       []string // list of peer addresses
	ClusterRegion      string
	ClusterPriority    int
	ClusterTLSCertPath string   // path to mTLS cert
	ClusterTLSKeyPath  string   // path to mTLS key
	ClusterTLSCAPath   string   // path to mTLS CA cert
	ClusterGRPCEnabled bool     // enable gRPC streaming (requires VORTEX_CLUSTER_ENABLED)

	// Dev mode
	Dev bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	core := env("VORTEX_CORE", "xray")
	cores := parseList(env("VORTEX_ENABLED_CORES", ""))
	if len(cores) == 0 {
		cores = []string{core}
	}

	cfg := &Config{
		HTTPAddr:  env("VORTEX_HTTP_ADDR", ":8080"),
		HTTPSAddr: env("VORTEX_HTTPS_ADDR", ""),
		GRPCAddr:  env("VORTEX_GRPC_ADDR", ":50051"),

		DatabaseURL: os.Getenv("VORTEX_DATABASE_URL"),
		DBType:      env("VORTEX_DB_TYPE", "sqlite"),
		RedisURL:    env("VORTEX_REDIS_URL", ""),

		JWTSecret:      os.Getenv("VORTEX_JWT_SECRET"),
		JWTTTL:         envDur("VORTEX_JWT_TTL", 24*time.Hour),
		SessionTimeout: envDur("VORTEX_SESSION_TIMEOUT", 30*time.Minute),

		TLSCert: os.Getenv("VORTEX_TLS_CERT"),
		TLSKey:  os.Getenv("VORTEX_TLS_KEY"),
		TLSCA:   os.Getenv("VORTEX_TLS_CA"),

		Core:        core,
		Cores:       cores,
		CoreBin:     os.Getenv("VORTEX_CORE_BIN"),
		CoreConfig:  env("VORTEX_CORE_CONFIG", "/etc/vortex/core.json"),
		CoreAPIPort: envInt("VORTEX_CORE_API_PORT", 10085),
		SingboxBin:  env("VORTEX_SINGBOX_BIN", "/usr/local/bin/sing-box"),

		WebhookURL:        os.Getenv("VORTEX_WEBHOOK_URL"),
		WebhookSecret:     os.Getenv("VORTEX_WEBHOOK_SECRET"),
		TelegramToken:     os.Getenv("VORTEX_TELEGRAM_TOKEN"),
		TelegramChat:      os.Getenv("VORTEX_TELEGRAM_CHAT_ID"),
		TelegramBotToken:  os.Getenv("VORTEX_TELEGRAM_BOT_TOKEN"),

		ActivityEnabled:  envBool("VORTEX_ACTIVITY_ENABLED", true),
		ActivityFlushSec: envInt("VORTEX_ACTIVITY_FLUSH_SEC", 30),

		CFToken: os.Getenv("VORTEX_CF_API_TOKEN"),
		CFZone:  os.Getenv("VORTEX_CF_ZONE_ID"),

		SubPort:    envInt("VORTEX_SUB_PORT", 2087),
		SubHost:    env("VORTEX_SUB_HOST", ""),
		SubPath:    env("VORTEX_SUB_PATH", "/sub"),
		SubSSL:     envBool("VORTEX_SUB_SSL", false),
		SubCert:    os.Getenv("VORTEX_SUB_CERT"),
		SubKey:     os.Getenv("VORTEX_SUB_KEY"),

		AutoRecover:     envBool("VORTEX_AUTO_RECOVER", true),
		RecoverAfter:    envDur("VORTEX_RECOVER_AFTER", 2*time.Minute),
		RecoverCooldown: envDur("VORTEX_RECOVER_COOLDOWN", 5*time.Minute),

		GeoIPDB: env("VORTEX_GEOIP_DB", ""),

		TunnelMonitorEnabled: envBool("VORTEX_TUNNEL_MONITOR", false),
		TunnelMonitorURL:     env("VORTEX_TUNNEL_MONITOR_URL", "https://www.cloudflare.com/cdn-cgi/trace"),
		TunnelMonitorProxy:   os.Getenv("VORTEX_TUNNEL_MONITOR_PROXY"),

		ZarinPalMerchant:     os.Getenv("VORTEX_ZARINPAL_MERCHANT"),
		NowPaymentsAPIKey:    os.Getenv("VORTEX_NOWPAYMENTS_KEY"),
		NowPaymentsIPNSecret: os.Getenv("VORTEX_NOWPAYMENTS_IPN_SECRET"),

		LogLevel:   env("VORTEX_LOG_LEVEL", "info"),
		LogFile:    env("VORTEX_LOG_FILE", ""),
		LogMaxSize: envInt("VORTEX_LOG_MAX_SIZE", 100),
		LogMaxAge:  envInt("VORTEX_LOG_MAX_AGE", 30),

		ClusterEnabled:     envBool("VORTEX_CLUSTER_ENABLED", false),
		ClusterNodeName:    env("VORTEX_CLUSTER_NODE_NAME", "node-1"),
		ClusterAddr:        env("VORTEX_CLUSTER_ADDR", ":1337"),
		ClusterPeers:       parseList(env("VORTEX_CLUSTER_PEERS", "")),
		ClusterRegion:      env("VORTEX_CLUSTER_REGION", "default"),
		ClusterPriority:    envInt("VORTEX_CLUSTER_PRIORITY", 100),
		ClusterTLSCertPath: env("VORTEX_CLUSTER_TLS_CERT", "/etc/vortex/pki/node/node.pem"),
		ClusterTLSKeyPath:  env("VORTEX_CLUSTER_TLS_KEY", "/etc/vortex/pki/node/node-key.pem"),
		ClusterTLSCAPath:   env("VORTEX_CLUSTER_TLS_CA", "/etc/vortex/pki/ca/ca.pem"),
		ClusterGRPCEnabled: envBool("VORTEX_CLUSTER_GRPC", false),

		Dev: envBool("VORTEX_DEV", false),
	}

	if cfg.CoreBin == "" {
		cfg.CoreBin = defaultCoreBin(cfg.Core)
	}

	var errs []error
	if cfg.DatabaseURL == "" && cfg.DBType == "sqlite" {
		cfg.DatabaseURL = "/etc/vortex/vortex.db"
	}
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("VORTEX_JWT_SECRET must be at least 32 characters"))
	}
	if cfg.DBType != "sqlite" && cfg.DBType != "postgres" {
		errs = append(errs, errors.New("VORTEX_DB_TYPE must be sqlite or postgres"))
	}

	return cfg, errors.Join(errs...)
}

// String returns a redacted string representation.
func (c *Config) String() string {
	return fmt.Sprintf("Config{http=%s grpc=%s db=%s jwt_ttl=%s cores=%v cluster=%v}",
		c.HTTPAddr, c.GRPCAddr, redact(c.DatabaseURL), c.JWTTTL, c.Cores, c.ClusterEnabled)
}

func defaultCoreBin(core string) string {
	switch core {
	case "xray":
		return "/usr/local/bin/xray"
	case "singbox":
		return "/usr/local/bin/sing-box"
	default:
		return core
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDur(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func parseList(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	seen := make(map[string]struct{})
	for _, part := range split(raw, ",") {
		part = trimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}

func split(s, sep string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func redact(url string) string {
	if url == "" {
		return "<unset>"
	}
	return "<set>"
}
