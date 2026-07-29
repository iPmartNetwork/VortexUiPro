package events

import (
	"sync"
	"time"
)

// Type identifies the category of an event.
type Type string

const (
	UserCreated     Type = "user.created"
	UserDeleted     Type = "user.deleted"
	UserLimited     Type = "user.limited"
	UserExpired     Type = "user.expired"
	UserReset       Type = "user.reset"
	UserIPLimit     Type = "user.ip_limit"
	UserExpiryWarn  Type = "user.expiry_warning"

	NodeDown         Type = "node.down"
	NodeUp           Type = "node.up"
	NodeDisconnect   Type = "node.disconnect"
	NodeAutoRecover  Type = "node.auto_recover"
	NodeAlert        Type = "node.alert"

	AdminQuotaWarning Type = "admin.quota_warning"
	SecurityProbe     Type = "security.probe"
	CertExpiring      Type = "cert.expiring"
	ProtocolSwitch    Type = "protocol.switch"

	SystemCPUHigh    Type = "system.cpu_high"
	SystemMemoryHigh Type = "system.memory_high"
	SystemCrash      Type = "system.crash"

	OutboundDown Type = "outbound.down"
	OutboundUp   Type = "outbound.up"

	TrafficUpdate  Type = "traffic.update"
	CoreRestart    Type = "core.restart"
	CoreCrash      Type = "core.crash"
	ConfigApplied  Type = "config.applied"
)

// Event is a single notification with structured data.
type Event struct {
	Type     Type              `json:"type"`
	Time     time.Time         `json:"time"`
	Source   string            `json:"source,omitempty"`
	UserID   string            `json:"user_id,omitempty"`
	Username string            `json:"username,omitempty"`
	NodeID   string            `json:"node_id,omitempty"`
	NodeName string            `json:"node_name,omitempty"`
	Message  string            `json:"message,omitempty"`
	Data     map[string]any    `json:"data,omitempty"`
}

// Publisher is the interface producers call.
type Publisher interface {
	Publish(e Event)
}

// Bus is an in-process pub/sub event bus with channel-based subscribers.
type Bus struct {
	mu   sync.RWMutex
	subs []chan Event
	log  func(string, ...any)
}

// New creates an event bus. The log function is optional; defaults to no-op.
func New(logFn func(string, ...any)) *Bus {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}
	return &Bus{log: logFn}
}

// Subscribe returns a channel that receives all published events. The buffer
// size must be >= 0; if <= 0, a default of 64 is used. Call Unsubscribe when
// done to avoid leaking the channel.
func (b *Bus) Subscribe(buffer int) <-chan Event {
	if buffer <= 0 {
		buffer = 64
	}
	ch := make(chan Event, buffer)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes a channel previously returned by Subscribe.
func (b *Bus) Unsubscribe(ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, c := range b.subs {
		if c == ch {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			close(c)
			return
		}
	}
}

// Publish fans out an event to all subscribers. Non-blocking per subscriber:
// a full buffer drops the event for that subscriber.
func (b *Bus) Publish(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
			b.log("event dropped: subscriber queue full", "type", string(e.Type))
		}
	}
}

// Nop is a no-op publisher that discards all events. Safe as a default.
type Nop struct{}

func (Nop) Publish(Event) {}
