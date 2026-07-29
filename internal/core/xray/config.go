package xray

// XrayConfig represents the complete Xray configuration structure.
type XrayConfig struct {
	LogConfig        string           `json:"log"`
	RouterConfig     string           `json:"routing"`
	DNSConfig        string           `json:"dns,omitempty"`
	InboundConfigs   []InboundConfig  `json:"inbounds"`
	OutboundConfigs  string           `json:"outbounds"`
	Transport        string           `json:"transport,omitempty"`
	Policy           string           `json:"policy"`
	API              string           `json:"api"`
	Stats            string           `json:"stats"`
	Reverse          string           `json:"reverse,omitempty"`
	FakeDNS          string           `json:"fakedns,omitempty"`
	Observatory      string           `json:"observatory,omitempty"`
	BurstObservatory string           `json:"burstObservatory,omitempty"`
	Metrics          string           `json:"metrics"`
	Geodata          string           `json:"geodata,omitempty"`
	Env              string           `json:"env,omitempty"`
}
