package domain

// Outbound represents a proxy outbound (egress) configuration.
type Outbound struct {
	ID             int64          `json:"id" db:"id"`
	NodeID         int64          `json:"node_id,omitempty" db:"node_id"`
	Tag            string         `json:"tag" db:"tag"`
	Protocol       Protocol       `json:"protocol" db:"protocol"`
	Settings       string         `json:"settings,omitempty" db:"settings"`
	StreamSettings string         `json:"stream_settings,omitempty" db:"stream_settings"`
	Remark         string         `json:"remark,omitempty" db:"remark"`
	Enable         bool           `json:"enable" db:"enable"`
	Hidden         bool           `json:"hidden,omitempty" db:"hidden"`
	CreatedAt      int64          `json:"created_at" db:"created_at"`
	UpdatedAt      int64          `json:"updated_at" db:"updated_at"`
}

// RoutingRule defines a single routing rule.
type RoutingRule struct {
	ID           int64    `json:"id" db:"id"`
	InboundTags  []string `json:"inbound_tags,omitempty" db:"inbound_tags"`
	OutboundTag  string   `json:"outbound_tag" db:"outbound_tag"`
	Domain       []string `json:"domain,omitempty" db:"domain"`
	IP           []string `json:"ip,omitempty" db:"ip"`
	Port         []string `json:"port,omitempty" db:"port"`
	Network      string   `json:"network,omitempty" db:"network"`
	Protocol     []string `json:"protocol,omitempty" db:"protocol"`
	GeoIP        []string `json:"geoip,omitempty" db:"geoip"`
	GeoSite      []string `json:"geosite,omitempty" db:"geosite"`
	SourceIP     []string `json:"source_ip,omitempty" db:"source_ip"`
	SourcePort   []string `json:"source_port,omitempty" db:"source_port"`
	UserEmail    []string `json:"user_email,omitempty" db:"user_email"`
	BalancerTag  string   `json:"balancer_tag,omitempty" db:"balancer_tag"`
	RuleType     string   `json:"rule_type,omitempty" db:"rule_type"`
	Enabled      bool     `json:"enable" db:"enable"`
	CreatedAt    int64    `json:"created_at" db:"created_at"`
}

// RoutingConfig wraps full routing configuration.
type RoutingConfig struct {
	DomainStrategy string         `json:"domain_strategy"`
	Rules          []RoutingRule  `json:"rules"`
	Balancers      []Balancer     `json:"balancers,omitempty"`
}

// Balancer defines load balancing between outbounds.
type Balancer struct {
	Tag      string   `json:"tag" db:"tag"`
	Selector []string `json:"selector" db:"selector"`
	Strategy string   `json:"strategy,omitempty" db:"strategy"` // random, roundrobin, leastping
}
