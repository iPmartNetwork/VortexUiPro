package domain

// SubscriptionHost represents a custom host configuration for subscription delivery.
type SubscriptionHost struct {
	ID           int64  `json:"id" db:"id"`
	Remark       string `json:"remark" db:"remark"`
	Domain       string `json:"domain" db:"domain"`
	CertFile     string `json:"cert_file,omitempty" db:"cert_file"`
	KeyFile      string `json:"key_file,omitempty" db:"key_file"`
	Enable       bool   `json:"enable" db:"enable"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

// SubscriptionLink is a shareable subscription link for a user.
type SubscriptionLink struct {
	ID           int64  `json:"id" db:"id"`
	UserID       int64  `json:"user_id" db:"user_id"`
	InboundID    int64  `json:"inbound_id" db:"inbound_id"`
	URL          string `json:"url" db:"url"`
	Token        string `json:"token" db:"token"`
	Enabled      bool   `json:"enable" db:"enable"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

// SubscriptionProfileDB is a multi-profile endpoint stored in the database (Heimdall feature).
type SubscriptionProfileDB struct {
	ID             int64  `json:"id" db:"id"`
	InboundID      int64  `json:"inbound_id" db:"inbound_id"`
	Dest           string `json:"dest" db:"dest"`
	Port           int    `json:"port" db:"port"`
	Remark         string `json:"remark,omitempty" db:"remark"`
	Enabled        bool   `json:"enabled" db:"enabled"`
	Network        string `json:"network,omitempty" db:"network"`
	Security       string `json:"security,omitempty" db:"security"`
	TLSSettings    string `json:"tls_settings,omitempty" db:"tls_settings"`
	RealitySettings string `json:"reality_settings,omitempty" db:"reality_settings"`
	Sockopt        string `json:"sockopt,omitempty" db:"sockopt"`
	MuxConfig      string `json:"mux_config,omitempty" db:"mux_config"`
	SNI            string `json:"sni,omitempty" db:"sni"`
	ALPN           string `json:"alpn,omitempty" db:"alpn"`
	Fingerprint    string `json:"fingerprint,omitempty" db:"fingerprint"`
	CreatedAt      int64  `json:"created_at" db:"created_at"`
}
