package xray

// Traffic represents network traffic statistics for Xray connections.
type Traffic struct {
	IsInbound  bool   `json:"is_inbound"`
	IsOutbound bool   `json:"is_outbound"`
	Tag        string `json:"tag"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
}

// ClientTraffic represents per-client traffic statistics.
type ClientTraffic struct {
	Email string `json:"email"`
	Up    int64  `json:"up"`
	Down  int64  `json:"down"`
}

// OnlineIP represents one source address of a live connection.
type OnlineIP struct {
	IP       string `json:"ip"`
	LastSeen int64  `json:"lastSeen"`
}

// OnlineUser represents a client email with live connections.
type OnlineUser struct {
	Email string     `json:"email"`
	IPs   []OnlineIP `json:"ips"`
}

// BalancerInfo is the live state of one balancer inside the running core.
type BalancerInfo struct {
	Tag      string   `json:"tag"`
	Override string   `json:"override"`
	Selected []string `json:"selected"`
}

// RouteTestRequest describes a synthetic connection to test routing.
type RouteTestRequest struct {
	InboundTag string `json:"inbound_tag,omitempty"`
	Domain     string `json:"domain,omitempty"`
	IP         string `json:"ip,omitempty"`
	Port       int    `json:"port"`
	Network    string `json:"network,omitempty"` // tcp, udp
	Protocol   string `json:"protocol,omitempty"`
	Email      string `json:"email,omitempty"`
}

// RouteTestResult is the routing decision from the core.
type RouteTestResult struct {
	Matched     bool     `json:"matched"`
	OutboundTag string   `json:"outboundTag"`
	GroupTags   []string `json:"groupTags,omitempty"`
}
