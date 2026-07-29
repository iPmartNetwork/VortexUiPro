package database

// ─── Protocol ────────────────────────────────────────────────────────
type Protocol string

const (
	ProtoVMess       Protocol = "vmess"
	ProtoVLESS       Protocol = "vless"
	ProtoTrojan      Protocol = "trojan"
	ProtoShadowsocks Protocol = "shadowsocks"
	ProtoHTTP        Protocol = "http"
	ProtoSocks       Protocol = "socks"
	ProtoWireGuard   Protocol = "wireguard"
	ProtoHysteria    Protocol = "hysteria"
	ProtoHysteria2   Protocol = "hysteria2"
	ProtoTUIC        Protocol = "tuic"
	ProtoMTProto     Protocol = "mtproto"
	ProtoDokodemo    Protocol = "dokodemo-door"
	ProtoMixed       Protocol = "mixed"
	ProtoTunnel      Protocol = "tunnel"
	ProtoNaive       Protocol = "naive"
	ProtoShadowTLS   Protocol = "shadowtls"
	ProtoAnyTLS      Protocol = "anytls"
)

// ─── Admin (panel operators) ─────────────────────────────────────────
type Admin struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string         `gorm:"uniqueIndex;not null;size:64" json:"username"`
	PasswordHash  string         `gorm:"not null" json:"-"`
	Email         string         `gorm:"size:128" json:"email,omitempty"`
	Role          string         `gorm:"default:admin;index;size:32" json:"role"`
	RoleID        int            `gorm:"default:0;index" json:"roleId"`
	Status        string         `gorm:"default:'active';index;size:16" json:"status"`
	TOTPSecret    string         `json:"-"`
	TOTPEnabled   bool           `gorm:"default:false" json:"totp_enabled"`
	LoginAttempts int            `gorm:"default:0" json:"-"`
	LockedUntil   int64          `gorm:"default:0" json:"-"`
	LoginEpoch    int64          `gorm:"default:0" json:"-"`
	APIKey        string         `json:"-"`
	APITokenHash  string         `json:"-"`

	TrafficLimit int64 `gorm:"default:0" json:"traffic_limit,omitempty"`
	UserLimit    int   `gorm:"default:0" json:"user_limit,omitempty"`
	InboundLimit int   `gorm:"default:0" json:"inbound_limit,omitempty"`

	DataLimit int64 `gorm:"default:0" json:"data_limit,omitempty"`
	UsedBytes int64 `gorm:"default:0" json:"used_bytes,omitempty"`

	TelegramID string `gorm:"size:64" json:"telegram_id,omitempty"`
	SupportURL string `gorm:"size:255" json:"support_url,omitempty"`
	Note       string `gorm:"type:text" json:"note,omitempty"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

// GetRoleID returns the role ID for this admin.
func (a *Admin) GetRoleID() int {
	return a.RoleID
}

func (Admin) TableName() string { return "admins" }

// ─── User (subscribers) ──────────────────────────────────────────────
type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID      int64  `gorm:"index;default:0" json:"admin_id,omitempty"`
	Username     string `gorm:"uniqueIndex;not null;size:64" json:"username"`
	PasswordHash string `gorm:"not null" json:"-"`
	Email        string `gorm:"size:128" json:"email,omitempty"`
	Status       string `gorm:"default:active;index;size:16" json:"status"`

	TrafficUp   int64 `gorm:"default:0" json:"traffic_up"`
	TrafficDown int64 `gorm:"default:0" json:"traffic_down"`
	DataLimit   int64 `gorm:"default:0" json:"data_limit"`
	ExpiryTime  int64 `gorm:"default:0" json:"expiry_time"`

	DeviceLimit   int `gorm:"default:0" json:"device_limit"`
	SpeedLimitUp   int `gorm:"default:0" json:"speed_limit_up"`
	SpeedLimitDown int `gorm:"default:0" json:"speed_limit_down"`

	Note      string `gorm:"type:text" json:"note,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (User) TableName() string { return "users" }

// ─── Inbound ─────────────────────────────────────────────────────────
type Inbound struct {
	ID        int64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64    `gorm:"index;default:0" json:"user_id,omitempty"`
	NodeID    int64    `gorm:"index;default:0" json:"node_id,omitempty"`
	Tag       string   `gorm:"uniqueIndex;not null;size:64" json:"tag"`
	Protocol  Protocol `gorm:"not null;index;size:32" json:"protocol"`
	Listen    string   `gorm:"default:'0.0.0.0';size:64" json:"listen"`
	Port      int      `gorm:"not null" json:"port"`
	Status    string   `gorm:"default:'active';size:16" json:"status"`
	Remark    string   `gorm:"size:128" json:"remark"`
	Enable    bool     `gorm:"default:true;index" json:"enable"`

	Settings       string `gorm:"type:text" json:"settings"`
	StreamSettings string `gorm:"type:text" json:"stream_settings"`
	Sniffing       string `gorm:"type:text" json:"sniffing,omitempty"`

	UpMbps   int   `gorm:"default:0" json:"up_mbps"`
	DownMbps int   `gorm:"default:0" json:"down_mbps"`
	TotalGB  int64 `gorm:"default:0" json:"total_gb"`
	ExpiryTime int64 `gorm:"default:0" json:"expiry_time"`

	Up   int64 `gorm:"default:0" json:"up"`
	Down int64 `gorm:"default:0" json:"down"`
	Total int64 `gorm:"default:0" json:"total"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Inbound) TableName() string { return "inbounds" }

// ─── Outbound ────────────────────────────────────────────────────────
type Outbound struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	NodeID    int64  `gorm:"index;default:0" json:"node_id,omitempty"`
	Tag       string `gorm:"uniqueIndex;not null;size:64" json:"tag"`
	Protocol  string `gorm:"not null;size:32" json:"protocol"`
	Settings  string `gorm:"type:text" json:"settings,omitempty"`
	StreamSettings string `gorm:"type:text" json:"stream_settings,omitempty"`
	Remark    string `gorm:"size:128" json:"remark,omitempty"`
	Enable    bool   `gorm:"default:true" json:"enable"`
	Hidden    bool   `gorm:"default:false" json:"hidden"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Outbound) TableName() string { return "outbounds" }

// ─── Client ──────────────────────────────────────────────────────────
type Client struct {
	ID        string `gorm:"primaryKey;size:64" json:"id"`
	UserID    int64  `gorm:"index;not null" json:"user_id"`
	InboundID int64  `gorm:"index;default:0" json:"inbound_id"`
	Email     string `gorm:"uniqueIndex;not null;size:128" json:"email"`
	Enable    bool   `gorm:"default:true;index" json:"enable"`

	Flow     string `gorm:"size:32" json:"flow,omitempty"`
	Password string `gorm:"type:text" json:"-"`
	Security string `gorm:"size:32" json:"security,omitempty"`
	Auth     string `gorm:"type:text" json:"auth,omitempty"`

	TotalGB    int64 `gorm:"default:0" json:"total_gb"`
	ExpiryTime int64 `gorm:"default:0" json:"expiry_time"`
	SubID      string `gorm:"index;size:64" json:"sub_id,omitempty"`

	UpMbps   int `gorm:"default:0" json:"up_mbps"`
	DownMbps int `gorm:"default:0" json:"down_mbps"`

	// WireGuard
	PrivateKey   string   `gorm:"type:text" json:"-" yaml:"-"`
	PublicKey    string   `gorm:"type:text" json:"-"`
	PreSharedKey string   `gorm:"type:text" json:"-"`
	AllowedIPs   string   `gorm:"type:text" json:"allowed_ips,omitempty"`
	KeepAlive    int      `gorm:"default:0" json:"keep_alive"`

	// MTProto
	Secret string `gorm:"size:128" json:"secret,omitempty"`
	AdTag  string `gorm:"size:64" json:"ad_tag,omitempty"`

	TgID    int64  `gorm:"default:0" json:"tg_id,omitempty"`
	Group   string `gorm:"size:64;default:'';index" json:"group,omitempty"`
	Comment string `gorm:"type:text" json:"comment,omitempty"`

	// RBAC: ownership tracking
	OwnerAdminID int64 `gorm:"default:0;index" json:"owner_admin_id,omitempty"`
	AdminID      int64 `gorm:"default:0;index" json:"admin_id,omitempty"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Client) TableName() string { return "clients" }

// ─── Node ────────────────────────────────────────────────────────────
type Node struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"uniqueIndex;not null;size:128" json:"name"`
	Address   string `gorm:"not null;size:255" json:"address"`
	Port      int    `gorm:"default:2053" json:"port"`
	APIPort   int    `gorm:"default:10085" json:"api_port"`
	Status    string `gorm:"default:'offline';index;size:16" json:"status"`
	CoreType  string `gorm:"default:'xray';size:16" json:"core_type"`
	Enable    bool   `gorm:"default:true" json:"enable"`

	Country  string `gorm:"size:64" json:"country,omitempty"`
	Location string `gorm:"size:128" json:"location,omitempty"`

	CPULoad    float64 `gorm:"default:0" json:"cpu_load"`
	MemoryUsed float64 `gorm:"default:0" json:"memory_used"`
	Uplink     int64   `gorm:"default:0" json:"uplink"`
	Downlink   int64   `gorm:"default:0" json:"downlink"`
	TrafficUp   int64   `gorm:"default:0" json:"traffic_up"`
	TrafficDown int64   `gorm:"default:0" json:"traffic_down"`

	LastHeartbeat int64  `gorm:"default:0" json:"last_heartbeat"`
	LastError     string `gorm:"type:text" json:"last_error,omitempty"`
	Remark        string `gorm:"size:255" json:"remark,omitempty"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Node) TableName() string { return "nodes" }

// ─── Setting (key-value store) ───────────────────────────────────────
type Setting struct {
	ID    int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Key   string `gorm:"uniqueIndex;not null;size:128" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (Setting) TableName() string { return "settings" }

// ─── RoutingRule ─────────────────────────────────────────────────────
type RoutingRule struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	InboundTags  string `gorm:"type:text" json:"inbound_tags,omitempty"`
	OutboundTag  string `gorm:"size:64;not null" json:"outbound_tag"`
	Domain       string `gorm:"type:text" json:"domain,omitempty"`
	IP           string `gorm:"type:text" json:"ip,omitempty"`
	Port         string `gorm:"size:255" json:"port,omitempty"`
	Network      string `gorm:"size:16" json:"network,omitempty"`
	Protocol     string `gorm:"type:text" json:"protocol,omitempty"`
	GeoIP        string `gorm:"type:text" json:"geoip,omitempty"`
	GeoSite      string `gorm:"type:text" json:"geosite,omitempty"`
	SourceIP     string `gorm:"type:text" json:"source_ip,omitempty"`
	BalancerTag  string `gorm:"size:64" json:"balancer_tag,omitempty"`
	RuleType     string `gorm:"size:32" json:"rule_type,omitempty"`
	Enabled      bool   `gorm:"default:true" json:"enable"`
	CreatedAt    int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (RoutingRule) TableName() string { return "routing_rules" }

// ─── SubscriptionHost (custom host for subscription delivery) ──────
type SubscriptionHost struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Remark    string `gorm:"size:128" json:"remark"`
	Domain    string `gorm:"uniqueIndex;not null;size:255" json:"domain"`
	CertFile  string `gorm:"size:255" json:"cert_file,omitempty"`
	KeyFile   string `gorm:"size:255" json:"key_file,omitempty"`
	Enable    bool   `gorm:"default:true" json:"enable"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (SubscriptionHost) TableName() string { return "subscription_hosts" }

// ─── SubscriptionProfile (multi-profile, Heimdall feature) ───────────
type SubscriptionProfile struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	InboundID int64  `gorm:"index;not null" json:"inbound_id"`
	Dest      string `gorm:"not null;size:255" json:"dest"`
	Port      int    `gorm:"not null" json:"port"`
	Remark    string `gorm:"size:128" json:"remark,omitempty"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	Network   string `gorm:"size:32" json:"network,omitempty"`
	Security  string `gorm:"size:32" json:"security,omitempty"`
	SNI       string `gorm:"size:255" json:"sni,omitempty"`
	ALPN      string `gorm:"size:255" json:"alpn,omitempty"`
	Fingerprint string `gorm:"size:64" json:"fingerprint,omitempty"`
	CreatedAt  int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (SubscriptionProfile) TableName() string { return "subscription_profiles" }

// ─── NotificationChannel ─────────────────────────────────────────────
type NotificationChannel struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string `gorm:"not null;size:32" json:"type"` // telegram, webhook
	Name      string `gorm:"not null;size:128" json:"name"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	Token     string `gorm:"type:text" json:"-"`
	ChatID    string `gorm:"size:64" json:"chat_id,omitempty"`
	WebhookURL string `gorm:"type:text" json:"webhook_url,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (NotificationChannel) TableName() string { return "notification_channels" }

// ─── SecurityEvent ───────────────────────────────────────────────────
type SecurityEvent struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string `gorm:"not null;index;size:32" json:"type"`
	SourceIP  string `gorm:"size:64" json:"source_ip,omitempty"`
	Username  string `gorm:"size:64" json:"username,omitempty"`
	Detail    string `gorm:"type:text" json:"detail,omitempty"`
	Level     string `gorm:"default:'info';size:16" json:"level"`
	NodeID    int64  `gorm:"index;default:0" json:"node_id,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index" json:"created_at"`
}

func (SecurityEvent) TableName() string { return "security_events" }

// ─── Ticket ──────────────────────────────────────────────────────────
type Ticket struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64  `gorm:"index;not null" json:"user_id"`
	Subject   string `gorm:"not null;size:255" json:"subject"`
	Message   string `gorm:"type:text;not null" json:"message"`
	Status    string `gorm:"default:'open';index;size:16" json:"status"` // open, answered, closed
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`

	Replies []TicketReply `gorm:"foreignKey:TicketID" json:"replies,omitempty"`
}

func (Ticket) TableName() string { return "tickets" }

// ─── TicketReply ─────────────────────────────────────────────────────
type TicketReply struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TicketID  int64  `gorm:"index;not null" json:"ticket_id"`
	UserID    int64  `gorm:"not null" json:"user_id"`
	IsAdmin   bool   `gorm:"default:false" json:"is_admin"`
	Message   string `gorm:"type:text;not null" json:"message"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (TicketReply) TableName() string { return "ticket_replies" }

// ─── Plan ────────────────────────────────────────────────────────────
type Plan struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"not null;size:128" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Price       int64  `gorm:"not null" json:"price"`
	DataLimit   int64  `gorm:"default:0" json:"data_limit"`
	SpeedLimit  int    `gorm:"default:0" json:"speed_limit"`
	DeviceLimit int    `gorm:"default:0" json:"device_limit"`
	Duration    int    `gorm:"default:0" json:"duration"` // days
	Protocol    string `gorm:"size:32" json:"protocol,omitempty"`
	NodeGroup   string `gorm:"size:64" json:"node_group,omitempty"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (Plan) TableName() string { return "plans" }

// ─── Order ───────────────────────────────────────────────────────────
type Order struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64  `gorm:"index;not null" json:"user_id"`
	PlanID    int64  `gorm:"not null" json:"plan_id"`
	Amount    int64  `gorm:"not null" json:"amount"`
	Status    string `gorm:"default:'pending';index;size:16" json:"status"`
	ProofFile string `gorm:"type:text" json:"proof_file,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	PaidAt    int64  `gorm:"default:0" json:"paid_at,omitempty"`
}

func (Order) TableName() string { return "orders" }

// ─── Transaction (wallet) ────────────────────────────────────────────
type Transaction struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64  `gorm:"index;not null" json:"user_id"`
	Amount      int64  `gorm:"not null" json:"amount"`
	Type        string `gorm:"not null;size:16" json:"type"` // deposit, withdrawal, payment
	Description string `gorm:"type:text" json:"description,omitempty"`
	ReferenceID string `gorm:"size:64" json:"reference_id,omitempty"`
	Status      string `gorm:"default:'pending';size:16" json:"status"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (Transaction) TableName() string { return "transactions" }

// ─── AdminRole (RBAC) ────────────────────────────────────────────────
type AdminRole struct {
	ID              int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"uniqueIndex;not null;size:64" json:"name"`
	Slug            string `gorm:"uniqueIndex;not null;size:64" json:"slug"`
	BuiltIn         bool   `gorm:"default:false" json:"builtIn"`
	OwnerRole       bool   `gorm:"default:false" json:"ownerRole"`
	PermissionsJSON string `gorm:"column:permissions;type:text" json:"-"`
	LimitsJSON      string `gorm:"column:limits;type:text" json:"-"`
	FeaturesJSON    string `gorm:"column:features;type:text" json:"-"`
	AccessJSON      string `gorm:"column:access;type:text" json:"-"`
	CreatedAt       int64  `gorm:"autoCreateTime:milli" json:"createdAt"`
	UpdatedAt       int64  `gorm:"autoUpdateTime:milli" json:"updatedAt"`
}

func (AdminRole) TableName() string { return "admin_roles" }

const (
	AdminRoleSlugOwner         = "owner"
	AdminRoleSlugAdministrator = "administrator"
	AdminRoleSlugOperator      = "operator"

	ApiTokenKindService   = "service"
	ApiTokenKindDelegated = "delegated"
	ApiTokenMillisecondsThreshold int64 = 1e12
)

// ─── ApiToken ────────────────────────────────────────────────────────
type ApiToken struct {
	ID             int     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string  `gorm:"uniqueIndex;not null;size:64" json:"name"`
	Token          string  `gorm:"uniqueIndex;not null;size:128" json:"-"`
	Kind           string  `gorm:"default:'service';size:16" json:"kind"`
	SubjectAdminID *int    `gorm:"index" json:"subjectAdminId,omitempty"`
	CreatedByAdminID *int  `gorm:"index" json:"createdByAdminId,omitempty"`
	ScopesJSON     string  `gorm:"type:text" json:"-"`
	ExpiresAt      int64   `gorm:"default:0" json:"expiresAt"`
	Enabled        bool    `gorm:"default:true" json:"enabled"`
	CreatedAt      int64   `gorm:"autoCreateTime:milli" json:"createdAt"`
}

func (ApiToken) TableName() string { return "api_tokens" }

// ─── ClusterNode ─────────────────────────────────────────────────────
type ClusterNode struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"uniqueIndex;not null;size:128" json:"name"`
	Address   string `gorm:"not null;size:255" json:"address"` // ip:port
	APIPort   int    `gorm:"default:8080" json:"api_port"`
	PeerPort  int    `gorm:"default:1337" json:"peer_port"` // cluster mesh port
	Role      string `gorm:"default:'follower';size:16" json:"role"` // leader, follower, candidate
	Status    string `gorm:"default:'offline';size:16" json:"status"` // online, offline, syncing
	Priority  int    `gorm:"default:100" json:"priority"`  // higher = more likely to be leader
	Term      int64  `gorm:"default:0" json:"term"`        // election term
	VotedFor  int64  `gorm:"default:0" json:"voted_for"`  // node ID voted for in current term
	Version   string `gorm:"size:32" json:"version,omitempty"`

	CPULoad    float64 `gorm:"default:0" json:"cpu_load"`
	MemoryUsed float64 `gorm:"default:0" json:"memory_used"`

	LastHeartbeat int64 `gorm:"default:0" json:"last_heartbeat"`
	LastSyncedAt  int64 `gorm:"default:0" json:"last_synced_at"`

	AdvertiseAddr string `gorm:"size:255" json:"advertise_addr,omitempty"`
	Region        string `gorm:"size:64" json:"region,omitempty"`
	Enabled       bool   `gorm:"default:true" json:"enabled"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (ClusterNode) TableName() string { return "cluster_nodes" }

// ─── SyncEvent (audit log for cluster sync operations) ───────────────
type SyncEvent struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string `gorm:"not null;index;size:32" json:"type"` // user_sync, traffic_sync, inbound_sync, ban_sync, election
	SourceID  int64  `gorm:"index;default:0" json:"source_id"`    // source cluster node ID
	TargetID  int64  `gorm:"default:0" json:"target_id"`
	EntityID  string `gorm:"size:64" json:"entity_id,omitempty"` // affected entity ID
	Status    string `gorm:"default:'completed';size:16" json:"status"`
	Detail    string `gorm:"type:text" json:"detail,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index" json:"created_at"`
}

func (SyncEvent) TableName() string { return "sync_events" }

// ─── ConfigVersion (from service package, registered in AutoMigrate) ─────
type ConfigVersion struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Resource    string `gorm:"size:64;not null;index" json:"resource"`
	ResourceID  int64  `gorm:"default:0;index" json:"resource_id"`
	Version     int    `gorm:"not null" json:"version"`
	Data        string `gorm:"type:text;not null" json:"data"`
	Description string `gorm:"size:255" json:"description,omitempty"`
	CreatedBy   string `gorm:"size:64" json:"created_by,omitempty"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (ConfigVersion) TableName() string { return "config_versions" }

// ─── AuditEntry ───────────────────────────────────────────────────────
type AuditEntry struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Actor     string `gorm:"size:128;not null;index" json:"actor"`
	Action    string `gorm:"size:64;not null;index" json:"action"`
	Resource  string `gorm:"size:128" json:"resource"`
	Detail    string `gorm:"type:text" json:"detail,omitempty"`
	SourceIP  string `gorm:"size:64" json:"source_ip,omitempty"`
	Outcome   string `gorm:"size:16" json:"outcome"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (AuditEntry) TableName() string { return "audit_logs" }

// ─── ClientGroup ─────────────────────────────────────────────────────
type ClientGroup struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"uniqueIndex;not null;size:128" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	AdminID     int64  `gorm:"index;default:0" json:"admin_id,omitempty"`
	MemberCount int    `gorm:"default:0" json:"member_count"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (ClientGroup) TableName() string { return "client_groups" }

// ─── ClientGroupMember ───────────────────────────────────────────────
type ClientGroupMember struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID   int64  `gorm:"uniqueIndex:idx_group_client;not null" json:"group_id"`
	ClientID  string `gorm:"uniqueIndex:idx_group_client;size:64;not null" json:"client_id"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (ClientGroupMember) TableName() string { return "client_group_members" }

// ─── RoutingPack (group of routing rules for quick switching) ─────
type RoutingPack struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"not null;size:128" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	RuleIDs     string `gorm:"type:text" json:"rule_ids,omitempty"` // JSON array of int64
	Enabled     bool   `gorm:"default:true" json:"enable"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (RoutingPack) TableName() string { return "routing_packs" }

// ─── HealthCheckConfig ──────────────────────────────────────────────
type HealthCheckConfig struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string `gorm:"uniqueIndex;not null;size:128" json:"name"`
	Target        string `gorm:"not null;size:255" json:"target"` // host:port or URL
	ProbeType     string `gorm:"default:'tcp';size:16" json:"probe_type"` // tcp, http, ping, grpc
	Interval      int    `gorm:"default:30" json:"interval"` // seconds
	Timeout       int    `gorm:"default:5" json:"timeout"` // seconds
	Threshold     int    `gorm:"default:3" json:"threshold"` // consecutive failures before alert
	ExpectedCode  int    `gorm:"default:200" json:"expected_code,omitempty"` // for HTTP probes
	ExpectedBody  string `gorm:"size:255" json:"expected_body,omitempty"`
	Enabled       bool   `gorm:"default:true" json:"enabled"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (HealthCheckConfig) TableName() string { return "health_check_configs" }

// ─── HealthCheckResult ──────────────────────────────────────────────
type HealthCheckResult struct {
	ID         int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigID   int64   `gorm:"index;not null" json:"config_id"`
	Target     string  `gorm:"size:255" json:"target"`
	ProbeType  string  `gorm:"size:16" json:"probe_type"`
	Success    bool    `json:"success"`
	LatencyMs  float64 `json:"latency_ms"`
	StatusCode int     `json:"status_code,omitempty"`
	ErrorMsg   string  `gorm:"type:text" json:"error,omitempty"`
	CreatedAt  int64   `gorm:"autoCreateTime:milli;index" json:"created_at"`
}

func (HealthCheckResult) TableName() string { return "health_check_results" }

// ─── AutoRecoveryRule ───────────────────────────────────────────────
type AutoRecoveryRule struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"not null;size:128" json:"name"`
	MatchLabel  string `gorm:"size:64" json:"match_label"` // match by check name, target, etc.
	ActionType  string `gorm:"not null;size:32" json:"action_type"` // restart_core, restart_node, reboot, webhook, script
	ActionParams string `gorm:"type:text" json:"action_params,omitempty"` // JSON params for the action
	Cooldown    int    `gorm:"default:300" json:"cooldown"` // seconds between recoveries
	MaxRetries  int    `gorm:"default:3" json:"max_retries"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (AutoRecoveryRule) TableName() string { return "auto_recovery_rules" }

// ─── AutoRecoveryAction (action history) ────────────────────────────
type AutoRecoveryAction struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID      int64  `gorm:"index" json:"rule_id"`
	CheckID     int64  `gorm:"index" json:"check_id"`
	ActionType  string `gorm:"size:32" json:"action_type"`
	Target      string `gorm:"size:255" json:"target"`
	Status      string `gorm:"default:'pending';size:16" json:"status"` // pending, success, failed
	Result      string `gorm:"type:text" json:"result,omitempty"`
	LatencyMs   float64 `json:"latency_ms"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (AutoRecoveryAction) TableName() string { return "auto_recovery_actions" }

// ─── TURNServer (WebRTC TURN server configuration) ─────────────────
type TURNServer struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Address   string `gorm:"uniqueIndex;not null;size:255" json:"address"` // e.g., turn:turn.example.com:3478
	Username  string `gorm:"size:128" json:"username,omitempty"`
	Password  string `gorm:"size:255" json:"password,omitempty"`
	Realm     string `gorm:"size:128" json:"realm,omitempty"`
	Protocol  string `gorm:"default:'udp';size:16" json:"protocol"` // udp, tcp, tls
	Status    string `gorm:"default:'offline';size:16" json:"status"`
	Region    string `gorm:"size:64" json:"region,omitempty"`
	Bandwidth int    `gorm:"default:100" json:"bandwidth"` // Mbps
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (TURNServer) TableName() string { return "turn_servers" }

// ─── P2PMeshConfig (WebRTC P2P mesh network configuration) ─────────
type P2PMeshConfig struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Enabled       bool   `gorm:"default:false" json:"enabled"`
	MeshName      string `gorm:"size:128;default:'vortexuipro-mesh'" json:"mesh_name"`
	Role          string `gorm:"size:16;default:'hybrid'" json:"role"` // relay, direct, hybrid
	ListenPort    int    `gorm:"default:0" json:"listen_port"`
	MaxPeers      int    `gorm:"default:10" json:"max_peers"`
	AutoReconnect bool   `gorm:"default:true" json:"auto_reconnect"`
	Discovery     string `gorm:"size:16;default:'signaling'" json:"discovery"` // dns, manual, signaling
	Encryption    bool   `gorm:"default:true" json:"encryption"`
	HeartbeatSec  int    `gorm:"default:30" json:"heartbeat_sec"`
	DataChannel   bool   `gorm:"default:true" json:"data_channel"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (P2PMeshConfig) TableName() string { return "p2p_mesh_configs" }

// ─── BackupEncryptionKey ───────────────────────────────────────────
type BackupEncryptionKey struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"size:128" json:"name"`
	KeyData   string `gorm:"type:text;not null" json:"-"` // AES-256 key, base64 encoded
	Active    bool   `gorm:"default:true" json:"active"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (BackupEncryptionKey) TableName() string { return "backup_encryption_keys" }

// ─── RemoteStorageConfig ────────────────────────────────────────────
type RemoteStorageConfig struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"size:128;not null" json:"name"`
	Type      string `gorm:"size:32;not null" json:"type"` // s3, gdrive
	Enabled   bool   `gorm:"default:true" json:"enabled"`

	// S3 fields
	S3Endpoint  string `gorm:"size:255" json:"s3_endpoint,omitempty"`
	S3Region    string `gorm:"size:64" json:"s3_region,omitempty"`
	S3Bucket    string `gorm:"size:128" json:"s3_bucket,omitempty"`
	S3AccessKey string `gorm:"size:255" json:"-"`
	S3SecretKey string `gorm:"size:255" json:"-"`
	S3Prefix    string `gorm:"size:255" json:"s3_prefix,omitempty"` // path prefix in bucket

	// Google Drive fields
	GDriveCredentials string `gorm:"type:text" json:"-"` // JSON credentials
	GDriveFolderID    string `gorm:"size:255" json:"gdrive_folder_id,omitempty"`

	CreatedAt int64 `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (RemoteStorageConfig) TableName() string { return "remote_storage_configs" }

// ─── CDNDomain (Domain Fronting / CDN Proxy) ───────────────────────
type CDNDomain struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Domain     string `gorm:"uniqueIndex;not null;size:255" json:"domain"`
	CDNProvider string `gorm:"size:32;index" json:"cdn_provider"` // cloudflare, fastly, akamai, cloudfront
	Status     string `gorm:"default:'unknown';size:16" json:"status"` // active, blocked, unknown
	Reachable  bool   `gorm:"default:false" json:"reachable"`
	LatencyMS  int    `gorm:"default:0" json:"latency_ms"`
	TLSVersion string `gorm:"size:16" json:"tls_version,omitempty"`
	ServerName string `gorm:"size:255" json:"server_name,omitempty"`
	Frontable  bool   `gorm:"default:false" json:"frontable"`
	LastChecked int64 `gorm:"default:0" json:"last_checked"`
	CreatedAt  int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (CDNDomain) TableName() string { return "cdn_domains" }

// ─── DNSConfig (Smart DNS System) ────────────────────────────────────
type DNSConfig struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"size:128;not null" json:"name"`
	Upstream     string `gorm:"size:255;not null" json:"upstream"` // e.g. https://dns.cloudflare.com/dns-query
	Type         string `gorm:"default:'doh';size:16" json:"type"` // doh, dot, udp
	Enabled      bool   `gorm:"default:true" json:"enabled"`
	Timeout      int    `gorm:"default:5" json:"timeout"` // seconds
	AdBlock      bool   `gorm:"default:false" json:"ad_block"`
	DNSSEC       bool   `gorm:"default:false" json:"dnssec"`
	CacheSize    int    `gorm:"default:1000" json:"cache_size"`
	RateLimit    int    `gorm:"default:100" json:"rate_limit"` // queries per second
	CreatedAt    int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (DNSConfig) TableName() string { return "dns_configs" }

// ─── DNSRule (DNS Routing / Rewrite Rules) ──────────────────────────
type DNSRule struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Domain    string `gorm:"not null;size:255;index" json:"domain"`
	Type      string `gorm:"default:'block';size:16" json:"type"` // block, redirect, custom_ip, proxy
	Action    string `gorm:"size:255" json:"action,omitempty"` // IP address, redirect URL, or proxy tag
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	Note      string `gorm:"size:255" json:"note,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (DNSRule) TableName() string { return "dns_rules" }

// ─── DockerContainer (Docker Native Mode) ───────────────────────────
type DockerContainer struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ContainerID string `gorm:"uniqueIndex;size:128" json:"container_id"`
	Name        string `gorm:"size:128;not null" json:"name"`
	Image       string `gorm:"size:255;not null" json:"image"`
	Status      string `gorm:"default:'created';size:16" json:"status"` // running, stopped, paused, exited
	Port        int    `gorm:"default:0" json:"port"`
	NodeID      int64  `gorm:"index;default:0" json:"node_id,omitempty"`
	CoreType    string `gorm:"size:16" json:"core_type,omitempty"` // xray, singbox
	CPU         float64 `gorm:"default:0" json:"cpu"`
	Memory      float64 `gorm:"default:0" json:"memory"`
	AutoRestart bool   `gorm:"default:true" json:"auto_restart"`
	HealthCheck string `gorm:"size:255" json:"health_check,omitempty"`
	EnvVars     string `gorm:"type:text" json:"env_vars,omitempty"` // JSON
	Labels      string `gorm:"type:text" json:"labels,omitempty"` // JSON
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (DockerContainer) TableName() string { return "docker_containers" }

// ─── FederationProvider (cross-panel sync) ──────────────────────────
type FederationProvider struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"size:128;not null"`
	APIURL      string `json:"api_url" gorm:"size:255;not null"`
	APIKey      string `json:"api_key,omitempty" gorm:"size:255"`
	SyncUsers   bool   `json:"sync_users" gorm:"default:true"`
	SyncPlans   bool   `json:"sync_plans" gorm:"default:true"`
	SyncTraffic bool   `json:"sync_traffic" gorm:"default:false"`
	Status      string `json:"status" gorm:"default:'offline';size:16"`
	LastSyncAt  int64  `json:"last_sync_at" gorm:"default:0"`
	CreatedAt   int64  `json:"created_at" gorm:"autoCreateTime:milli"`
}

func (FederationProvider) TableName() string { return "federation_providers" }
