package domain

// SecurityEvent represents a security-related event logged by the system.
type SecurityEvent struct {
	ID        int64  `json:"id" db:"id"`
	Type      string `json:"type" db:"type"` // probe, brute_force, suspicious_ip, etc.
	SourceIP  string `json:"source_ip,omitempty" db:"source_ip"`
	Username  string `json:"username,omitempty" db:"username"`
	Detail    string `json:"detail,omitempty" db:"detail"`
	Level     string `json:"level" db:"level"` // info, warning, critical
	NodeID    int64  `json:"node_id,omitempty" db:"node_id"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

// TLSConfig represents TLS configuration for an inbound or host.
type TLSConfig struct {
	ServerName    string   `json:"server_name,omitempty"`
	ALPN          []string `json:"alpn,omitempty"`
	CertFile      string   `json:"cert_file,omitempty"`
	KeyFile       string   `json:"key_file,omitempty"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	AllowInsecure bool     `json:"allow_insecure,omitempty"`
	MinVersion    string   `json:"min_version,omitempty"`
	MaxVersion    string   `json:"max_version,omitempty"`
	CipherSuites  []string `json:"cipher_suites,omitempty"`
	EchConfigList string   `json:"ech_config_list,omitempty"`
	PinnedPeerSHA []string `json:"pinned_peer_cert_sha256,omitempty"`
}

// RealityConfig represents REALITY protocol configuration.
type RealityConfig struct {
	ServerNames  []string `json:"server_names,omitempty"`
	Dest         string   `json:"dest,omitempty"`
	PrivateKey   string   `json:"private_key,omitempty"`
	PublicKey    string   `json:"public_key,omitempty"`
	ShortIDs     []string `json:"short_ids,omitempty"`
	Fingerprint  string   `json:"fingerprint,omitempty"`
	SpiderX      string   `json:"spider_x,omitempty"`
	MldsaKey     string   `json:"mldsa_65_key,omitempty"`
	Show         bool     `json:"show,omitempty"`
}

// PerformanceMetrics holds system performance data.
type PerformanceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsed  float64 `json:"memory_used"`
	MemoryTotal float64 `json:"memory_total"`
	DiskUsed    float64 `json:"disk_used"`
	DiskTotal   float64 `json:"disk_total"`
	LoadAvg     float64 `json:"load_avg"`
	Uptime      int64   `json:"uptime"`
	ProcessCount int    `json:"process_count"`
	NetworkIn   int64   `json:"network_in"`
	NetworkOut  int64   `json:"network_out"`
}

// CleanIPResult stores the result of a clean IP scan.
type CleanIPResult struct {
	ID          int64  `json:"id" db:"id"`
	NodeID      int64  `json:"node_id" db:"node_id"`
	IP          string `json:"ip" db:"ip"`
	Port        int    `json:"port" db:"port"`
	Protocol    string `json:"protocol,omitempty" db:"protocol"`
	LatencyMs   int    `json:"latency_ms,omitempty" db:"latency_ms"`
	IsClean     bool   `json:"is_clean" db:"is_clean"`
	CheckedAt   int64  `json:"checked_at" db:"checked_at"`
}
