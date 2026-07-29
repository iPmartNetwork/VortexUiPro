package domain

// CoreType identifies the proxy engine.
type CoreType string

const (
	CoreXray    CoreType = "xray"
	CoreSingbox CoreType = "singbox"
)

// Protocol is an inbound proxy protocol.
type Protocol string

const (
	ProtoVMess       Protocol = "vmess"
	ProtoVLESS       Protocol = "vless"
	ProtoTrojan      Protocol = "trojan"
	ProtoShadowsocks Protocol = "shadowsocks"
	ProtoSocks       Protocol = "socks"
	ProtoHTTP        Protocol = "http"
	ProtoDokodemo    Protocol = "dokodemo-door"
	ProtoHysteria2   Protocol = "hysteria2"
	ProtoHysteria    Protocol = "hysteria"
	ProtoTUIC        Protocol = "tuic"
	ProtoWireGuard   Protocol = "wireguard"
	ProtoShadowTLS   Protocol = "shadowtls"
	ProtoAnyTLS      Protocol = "anytls"
	ProtoNaive       Protocol = "naive"
	ProtoMTProto     Protocol = "mtproto"
)

// Security modes for inbound TLS configuration.
type Security string

const (
	SecurityNone    Security = "none"
	SecurityTLS     Security = "tls"
	SecurityReality Security = "reality"
)

// InboundStatus represents the operational state of an inbound.
type InboundStatus string

const (
	InboundActive   InboundStatus = "active"
	InboundInactive InboundStatus = "inactive"
	InboundError    InboundStatus = "error"
)

// Inbound represents a proxy inbound configuration.
type Inbound struct {
	ID              int64          `json:"id" db:"id"`
	UserID          int64          `json:"user_id" db:"user_id"`
	NodeID          int64          `json:"node_id,omitempty" db:"node_id"`
	Tag             string         `json:"tag" db:"tag"`
	Protocol        Protocol       `json:"protocol" db:"protocol"`
	Listen          string         `json:"listen,omitempty" db:"listen_ip"`
	Port            int            `json:"port" db:"port"`
	Status          InboundStatus  `json:"status" db:"status"`
	StreamSettings  string         `json:"stream_settings,omitempty" db:"stream_settings"`
	Settings        string         `json:"settings,omitempty" db:"settings"`
	Sniffing        string         `json:"sniffing,omitempty" db:"sniffing"`
	Allocate        string         `json:"allocate,omitempty" db:"allocate"`
	Remark          string         `json:"remark" db:"remark"`
	UpMbps          int            `json:"up_mbps,omitempty" db:"up_mbps"`
	DownMbps        int            `json:"down_mbps,omitempty" db:"down_mbps"`
	TotalGB         int64          `json:"total_gb,omitempty" db:"total_gb"`
	ExpiryTime      int64          `json:"expiry_time,omitempty" db:"expiry_time"`
	Enable          bool           `json:"enable" db:"enable"`
	SubscriptionID  string         `json:"subscription_id,omitempty" db:"subscription_id"`
	CreatedAt       int64          `json:"created_at" db:"created_at"`
	UpdatedAt       int64          `json:"updated_at" db:"updated_at"`

	// Multi-profile subscription endpoints (Heimdall feature)
	ExternalProxy   []SubscriptionProfile `json:"external_proxy,omitempty"`
}

// SubscriptionProfile represents one endpoint in a multi-profile inbound.
type SubscriptionProfile struct {
	Dest              string         `json:"dest"`
	Port              int            `json:"port"`
	Remark            string         `json:"remark,omitempty"`
	Enabled           bool           `json:"enabled"`
	Network           string         `json:"network,omitempty"`
	Security          string         `json:"security,omitempty"`
	TLSSettings       map[string]any `json:"tlsSettings,omitempty"`
	RealitySettings   map[string]any `json:"realitySettings,omitempty"`
	Sockopt           map[string]any `json:"sockopt,omitempty"`
	Mux               map[string]any `json:"mux,omitempty"`
	FinalMask         map[string]any `json:"finalmask,omitempty"`
	ForceTLS          string         `json:"forceTls,omitempty"`
	SNI               string         `json:"sni,omitempty"`
	ALPN              []string       `json:"alpn,omitempty"`
	Fingerprint       string         `json:"fingerprint,omitempty"`
	OverrideSNIFromAddr bool        `json:"overrideSniFromAddress,omitempty"`
	KeepSNIBlank      bool           `json:"keepSniBlank,omitempty"`
	VerifyPeerCert    string         `json:"verifyPeerCertByName,omitempty"`
	AllowInsecure     bool           `json:"allowInsecure,omitempty"`
	// Stream settings overrides
	TCPSettings       map[string]any `json:"tcpSettings,omitempty"`
	WSSettings        map[string]any `json:"wsSettings,omitempty"`
	KCPOptions        map[string]any `json:"kcpSettings,omitempty"`
	GRPCSettings      map[string]any `json:"grpcSettings,omitempty"`
	HTTPUpgradeSettings map[string]any `json:"httpupgradeSettings,omitempty"`
	XHTTPSettings     map[string]any `json:"xhttpSettings,omitempty"`
	HysteriaSettings  map[string]any `json:"hysteriaSettings,omitempty"`
}
