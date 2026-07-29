package domain

// NotificationChannel represents a configured notification destination.
type NotificationChannel struct {
	ID          int64  `json:"id" db:"id"`
	Type        string `json:"type" db:"type"` // telegram, webhook, email
	Name        string `json:"name" db:"name"`
	Token       string `json:"-" db:"token"`
	ChatID      string `json:"chat_id,omitempty" db:"chat_id"`
	WebhookURL  string `json:"webhook_url,omitempty" db:"webhook_url"`
	WebhookSecret string `json:"-" db:"webhook_secret"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
}

// NotificationEvent types are the events that trigger notifications.
type NotificationEvent string

const (
	NotifyUserCreated       NotificationEvent = "user.created"
	NotifyUserDeleted       NotificationEvent = "user.deleted"
	NotifyUserLimited       NotificationEvent = "user.limited"
	NotifyUserExpired       NotificationEvent = "user.expired"
	NotifyUserReset         NotificationEvent = "user.reset"
	NotifyUserExpiryWarn    NotificationEvent = "user.expiry_warning"
	NotifyUserIPLimit       NotificationEvent = "user.ip_limit"
	NotifyNodeDown          NotificationEvent = "node.down"
	NotifyNodeUp            NotificationEvent = "node.up"
	NotifyNodeDisconnect    NotificationEvent = "node.disconnect_alert"
	NotifyNodeAutoRecover   NotificationEvent = "node.auto_recover"
	NotifySystemAlert       NotificationEvent = "system.alert"
	NotifyCertExpiring      NotificationEvent = "cert.expiring"
	NotifySecurityProbe     NotificationEvent = "security.probe"
	NotifyCPUHigh           NotificationEvent = "cpu.high"
	NotifyMemoryHigh        NotificationEvent = "memory.high"
)

// EventSubscription links a notification channel to event types.
type EventSubscription struct {
	ID              int64              `json:"id" db:"id"`
	ChannelID       int64              `json:"channel_id" db:"channel_id"`
	Events          []NotificationEvent `json:"events" db:"events"`
	Enabled         bool               `json:"enabled" db:"enabled"`
	CreatedAt       int64              `json:"created_at" db:"created_at"`
}
