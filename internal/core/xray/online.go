package xray

import (
	"sync"
	"time"
)

// OnlineTracker tracks online users and their source IPs.
type OnlineTracker struct {
	mu          sync.RWMutex
	users       map[string]*onlineUserEntry
	gracePeriod time.Duration
}

type onlineUserEntry struct {
	Email   string
	IPs     map[string]int64 // IP -> last seen unix timestamp
	Updated int64
}

// NewOnlineTracker creates a new online user tracker.
func NewOnlineTracker(gracePeriod time.Duration) *OnlineTracker {
	if gracePeriod == 0 {
		gracePeriod = 2 * time.Minute
	}
	return &OnlineTracker{
		users:       make(map[string]*onlineUserEntry),
		gracePeriod: gracePeriod,
	}
}

// RecordActivity records traffic activity for a user from an IP.
func (t *OnlineTracker) RecordActivity(email, ip string) {
	now := time.Now().UnixMilli()
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.users[email]
	if !ok {
		entry = &onlineUserEntry{
			Email: email,
			IPs:   make(map[string]int64),
		}
		t.users[email] = entry
	}
	entry.IPs[ip] = now
	entry.Updated = now
}

// GetOnlineUsers returns all users seen within the grace period with their IPs.
func (t *OnlineTracker) GetOnlineUsers() []OnlineUser {
	now := time.Now().UnixMilli()
	threshold := now - t.gracePeriod.Milliseconds()

	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []OnlineUser
	for email, entry := range t.users {
		if entry.Updated < threshold {
			continue
		}
		user := OnlineUser{Email: email}
		for ip, lastSeen := range entry.IPs {
			if lastSeen >= threshold {
				user.IPs = append(user.IPs, OnlineIP{
					IP:       ip,
					LastSeen: lastSeen,
				})
			}
		}
		if len(user.IPs) > 0 {
			result = append(result, user)
		}
	}
	return result
}

// GetOnlineCount returns the number of unique online users.
func (t *OnlineTracker) GetOnlineCount() int {
	now := time.Now().UnixMilli()
	threshold := now - t.gracePeriod.Milliseconds()

	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, entry := range t.users {
		if entry.Updated >= threshold {
			count++
		}
	}
	return count
}

// Cleanup removes stale entries that are past the grace period.
func (t *OnlineTracker) Cleanup() {
	now := time.Now().UnixMilli()
	threshold := now - t.gracePeriod.Milliseconds()

	t.mu.Lock()
	defer t.mu.Unlock()

	for email, entry := range t.users {
		if entry.Updated < threshold {
			delete(t.users, email)
			continue
		}
		for ip, lastSeen := range entry.IPs {
			if lastSeen < threshold {
				delete(entry.IPs, ip)
			}
		}
		if len(entry.IPs) == 0 {
			delete(t.users, email)
		}
	}
}
