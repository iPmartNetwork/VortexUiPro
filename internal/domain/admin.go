package domain

// AdminRole defines the role of an administrative user.
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin"
	RoleAdmin      AdminRole = "admin"
	RoleReseller   AdminRole = "reseller"
	RoleOperator   AdminRole = "operator"
	RoleViewer     AdminRole = "viewer"
)

// Admin represents an administrative user of the panel.
type Admin struct {
	ID            int64     `json:"id" db:"id"`
	Username      string    `json:"username" db:"username"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	Email         string    `json:"email,omitempty" db:"email"`
	Role          AdminRole `json:"role" db:"role"`
	TOTPSecret    string    `json:"-" db:"totp_secret"`
	TOTPEnabled   bool      `json:"totp_enabled" db:"totp_enabled"`
	LoginAttempts int       `json:"-" db:"login_attempts"`
	LockedUntil   int64     `json:"-" db:"locked_until"`
	APIKey        string    `json:"-" db:"api_key"`
	APITokenHash  string    `json:"-" db:"api_token_hash"`

	// Reseller-specific
	TrafficLimit    int64 `json:"traffic_limit,omitempty" db:"traffic_limit"`
	UserLimit       int   `json:"user_limit,omitempty" db:"user_limit"`
	InboundLimit    int   `json:"inbound_limit,omitempty" db:"inbound_limit"`
	CommissionRate  int   `json:"commission_rate,omitempty" db:"commission_rate"`

	CreatedAt int64 `json:"created_at" db:"created_at"`
	UpdatedAt int64 `json:"updated_at" db:"updated_at"`
}
