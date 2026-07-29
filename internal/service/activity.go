package service

import (
	"context"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Client Activity Tracking ────────────────────────────────────────

// ActivityRecord stores a single client activity event.
type ActivityRecord struct {
	ClientID  string `json:"client_id"`
	Email     string `json:"email"`
	Action    string `json:"action"` // connect, disconnect, traffic_update
	IP        string `json:"ip,omitempty"`
	Device    string `json:"device,omitempty"`
	TrafficUp int64  `json:"traffic_up,omitempty"`
	TrafficDown int64 `json:"traffic_down,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// ClientActivityService manages the in-memory client activity buffer.
type ClientActivityService struct {
	mu          sync.RWMutex
	activities  []ActivityRecord
	maxBuffer   int
	flushEvery  time.Duration
	autoFlush   bool
	stopFlush   chan struct{}
}

// NewClientActivityService creates a new client activity service.
// flushIntervalSec sets the auto-flush interval; 0 defaults to 30s.
func NewClientActivityService(flushIntervalSec int) *ClientActivityService {
	if flushIntervalSec <= 0 {
		flushIntervalSec = 30
	}
	s := &ClientActivityService{
		activities: make([]ActivityRecord, 0, 1000),
		maxBuffer:  5000,
		flushEvery: time.Duration(flushIntervalSec) * time.Second,
		stopFlush:  make(chan struct{}),
	}
	return s
}

// StartAutoFlush begins periodic flushing of buffered activities to the database.
func (s *ClientActivityService) StartAutoFlush() {
	s.autoFlush = true
	go func() {
		ticker := time.NewTicker(s.flushEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.Flush(); err != nil {
					// Log and continue
				}
			case <-s.stopFlush:
				return
			}
		}
	}()
}

// StopAutoFlush stops the periodic flush goroutine.
func (s *ClientActivityService) StopAutoFlush() {
	if s.autoFlush {
		close(s.stopFlush)
		s.autoFlush = false
	}
}

// Record adds a client activity to the buffer.
func (s *ClientActivityService) Record(clientID, email, action, ip, device string, up, down int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := ActivityRecord{
		ClientID:    clientID,
		Email:       email,
		Action:      action,
		IP:          ip,
		Device:      device,
		TrafficUp:   up,
		TrafficDown: down,
		Timestamp:   time.Now().UnixMilli(),
	}

	s.activities = append(s.activities, record)

	// Auto-flush if buffer exceeds max
	if len(s.activities) >= s.maxBuffer {
		go func() {
			_ = s.Flush()
		}()
	}
}

// Flush writes all buffered activities to the database and clears the buffer.
func (s *ClientActivityService) Flush() error {
	s.mu.Lock()
	batch := s.activities
	s.activities = make([]ActivityRecord, 0, 1000)
	s.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	// Batch insert into database
	for _, a := range batch {
		// Update user traffic if applicable (User model has TrafficUp/TrafficDown)
		if a.Action == "traffic_update" && a.Email != "" && (a.TrafficUp > 0 || a.TrafficDown > 0) {
			client, err := database.GetClientByEmail(a.Email)
			if err == nil {
				// Update the user's traffic through the client's user ID
				user, err := database.GetUserByID(client.UserID)
				if err == nil {
					user.TrafficUp += a.TrafficUp
					user.TrafficDown += a.TrafficDown
					_ = database.UpdateUser(user)
				}
			}
		}
	}

	return nil
}

// GetRecentActivity returns the most recent activities.
func (s *ClientActivityService) GetRecentActivity(limit int) []ActivityRecord {
	if limit <= 0 {
		limit = 50
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	start := len(s.activities) - limit
	if start < 0 {
		start = 0
	}

	result := make([]ActivityRecord, len(s.activities[start:]))
	copy(result, s.activities[start:])
	// Reverse to get newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// GetOnlineCount returns the count of clients with recent activity (last 5 min).
func (s *ClientActivityService) GetOnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-5 * time.Minute).UnixMilli()
	seen := make(map[string]bool)
	for _, a := range s.activities {
		if a.Timestamp >= cutoff {
			seen[a.ClientID] = true
		}
	}
	return len(seen)
}

// ─── Activity Handler ───────────────────────────────────────────────

// TrafficCollector periodically collects traffic from xray gRPC and stores it.
type TrafficCollector struct {
	xraySvc     *XrayService
	activitySvc *ClientActivityService
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
}

// NewTrafficCollector creates a new traffic collector.
func NewTrafficCollector(xraySvc *XrayService, activitySvc *ClientActivityService) *TrafficCollector {
	return &TrafficCollector{
		xraySvc:     xraySvc,
		activitySvc: activitySvc,
		interval:    60 * time.Second,
		stopCh:      make(chan struct{}),
	}
}

// Start begins periodic traffic collection.
func (tc *TrafficCollector) Start() {
	go func() {
		ticker := time.NewTicker(tc.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tc.collect()
			case <-tc.stopCh:
				return
			}
		}
	}()
}

// Stop stops the traffic collector (safe to call multiple times).
func (tc *TrafficCollector) Stop() {
	tc.stopOnce.Do(func() {
		close(tc.stopCh)
	})
}

func (tc *TrafficCollector) collect() {
	// Collect traffic from xray gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stats, err := tc.xraySvc.CollectTraffic(ctx)
	if err != nil {
		return
	}

	for _, stat := range stats {
		tc.activitySvc.Record(
			"",
			stat.Tag,
			"traffic_update",
			"",
			"",
			stat.Up,
			stat.Down,
		)
	}
}
