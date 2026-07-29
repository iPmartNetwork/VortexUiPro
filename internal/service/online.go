package service

import (
	"sync"
	"time"
)

// OnlineUser represents a currently connected/active user.
type OnlineUser struct {
	ClientID   string `json:"client_id"`
	Email      string `json:"email"`
	InboundID  int64  `json:"inbound_id"`
	IP         string `json:"ip"`
	Device     string `json:"device,omitempty"`
	ConnectedAt int64 `json:"connected_at"`
	LastActive  int64 `json:"last_active"`
	TrafficUp   int64 `json:"traffic_up"`
	TrafficDown int64 `json:"traffic_down"`
}

// OnlineTracker maintains a real-time list of online users.
type OnlineTracker struct {
	mu           sync.RWMutex
	users        map[string]*OnlineUser // key: email
	cleanupEvery time.Duration
	timeout      time.Duration
	stopCh       chan struct{}
}

// NewOnlineTracker creates a new online user tracker.
func NewOnlineTracker() *OnlineTracker {
	ot := &OnlineTracker{
		users:        make(map[string]*OnlineUser),
		cleanupEvery: 30 * time.Second,
		timeout:      5 * time.Minute, // user considered offline after 5 min of inactivity
		stopCh:       make(chan struct{}),
	}
	go ot.cleanupLoop()
	return ot
}

// Stop stops the cleanup loop.
func (ot *OnlineTracker) Stop() {
	close(ot.stopCh)
}

// Connect registers a user connection.
func (ot *OnlineTracker) Connect(clientID, email string, inboundID int64, ip, device string) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	now := time.Now().UnixMilli()
	ot.users[email] = &OnlineUser{
		ClientID:    clientID,
		Email:       email,
		InboundID:   inboundID,
		IP:          ip,
		Device:      device,
		ConnectedAt: now,
		LastActive:  now,
	}
}

// Disconnect removes a user.
func (ot *OnlineTracker) Disconnect(email string) {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	delete(ot.users, email)
}

// UpdateActivity updates a user's last active timestamp and traffic.
func (ot *OnlineTracker) UpdateActivity(email string, up, down int64) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	if user, ok := ot.users[email]; ok {
		user.LastActive = time.Now().UnixMilli()
		user.TrafficUp += up
		user.TrafficDown += down
	}
}

// GetOnline returns the current list of online users.
func (ot *OnlineTracker) GetOnline() []OnlineUser {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	result := make([]OnlineUser, 0, len(ot.users))
	for _, u := range ot.users {
		result = append(result, *u)
	}
	return result
}

// GetOnlineCount returns the number of online users.
func (ot *OnlineTracker) GetOnlineCount() int {
	ot.mu.RLock()
	defer ot.mu.RUnlock()
	return len(ot.users)
}

// GetOnlineByInbound groups online users by inbound ID.
func (ot *OnlineTracker) GetOnlineByInbound() map[int64]int {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	counts := make(map[int64]int)
	for _, u := range ot.users {
		counts[u.InboundID]++
	}
	return counts
}

func (ot *OnlineTracker) cleanupLoop() {
	ticker := time.NewTicker(ot.cleanupEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ot.cleanup()
		case <-ot.stopCh:
			return
		}
	}
}

func (ot *OnlineTracker) cleanup() {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	cutoff := time.Now().Add(-ot.timeout).UnixMilli()
	for email, user := range ot.users {
		if user.LastActive < cutoff {
			delete(ot.users, email)
		}
	}
}
