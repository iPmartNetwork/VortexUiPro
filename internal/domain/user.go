package domain

// UserStatus represents the current state of a user account.
type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserDisabled UserStatus = "disabled"
	UserLimited  UserStatus = "limited"
	UserExpired  UserStatus = "expired"
	UserOnHold   UserStatus = "on_hold"
)

// User represents a panel user (subscriber).
type User struct {
	ID            int64      `json:"id" db:"id"`
	AdminID       int64      `json:"admin_id,omitempty" db:"admin_id"`
	Username      string     `json:"username" db:"username"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Email         string     `json:"email,omitempty" db:"email"`
	Status        UserStatus `json:"status" db:"status"`
	TrafficUp     int64      `json:"traffic_up" db:"traffic_up"`
	TrafficDown   int64      `json:"traffic_down" db:"traffic_down"`
	TrafficTotal  int64      `json:"traffic_total" db:"traffic_total"`
	DataLimit     int64      `json:"data_limit,omitempty" db:"data_limit"`
	ExpiryTime    int64      `json:"expiry_time,omitempty" db:"expiry_time"`
	DeviceLimit   int        `json:"device_limit,omitempty" db:"device_limit"`
	SpeedLimitUp  int        `json:"speed_limit_up,omitempty" db:"speed_limit_up"`
	SpeedLimitDown int       `json:"speed_limit_down,omitempty" db:"speed_limit_down"`
	Note          string     `json:"note,omitempty" db:"note"`
	CreatedAt     int64      `json:"created_at" db:"created_at"`
	UpdatedAt     int64      `json:"updated_at" db:"updated_at"`
}

// Client represents a proxy client linked to a user.
type Client struct {
	ID          string   `json:"id" db:"id"`             // UUID
	UserID      int64    `json:"user_id" db:"user_id"`
	InboundID   int64    `json:"inbound_id" db:"inbound_id"`
	Email       string   `json:"email" db:"email"`
	Enable      bool     `json:"enable" db:"enable"`
	Flow        string   `json:"flow,omitempty" db:"flow"`
	Password    string   `json:"password,omitempty" db:"password"`
	Security    string   `json:"security,omitempty" db:"security"`
	TotalGB     int64    `json:"total_gb,omitempty" db:"total_gb"`
	ExpiryTime  int64    `json:"expiry_time,omitempty" db:"expiry_time"`
	SubID       string   `json:"sub_id,omitempty" db:"sub_id"`
	UpMbps      int      `json:"up_mbps,omitempty" db:"up_mbps"`
	DownMbps    int      `json:"down_mbps,omitempty" db:"down_mbps"`
	PrivateKey  string   `json:"private_key,omitempty" db:"private_key"`
	PublicKey   string   `json:"public_key,omitempty" db:"public_key"`
	PreSharedKey string  `json:"pre_shared_key,omitempty" db:"pre_shared_key"`
	AllowedIPs  []string `json:"allowed_ips,omitempty" db:"allowed_ips"`
	KeepAlive   int      `json:"keep_alive,omitempty" db:"keep_alive"`
	// MTProto specific
	Secret      string   `json:"secret,omitempty" db:"secret"`
	AdTag       string   `json:"ad_tag,omitempty" db:"ad_tag"`
}
