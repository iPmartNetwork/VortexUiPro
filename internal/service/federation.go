package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// FederationService manages cross-panel federation connections.
// Uses database.FederationProvider model (moved to database package to avoid circular import).
type FederationService struct {
	eventBus events.Publisher
	mu       sync.RWMutex
	client   *http.Client
	running  bool
	stopCh   chan struct{}
}

// NewFederationService creates a new federation service.
func NewFederationService(bus events.Publisher) *FederationService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &FederationService{
		eventBus: bus,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 60 * time.Second,
			},
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins the periodic sync loop.
func (s *FederationService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		// Initial sync
		s.syncAll()
		for {
			select {
			case <-ticker.C:
				s.syncAll()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop stops the periodic sync loop.
func (s *FederationService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

// ─── CRUD ────────────────────────────────────────────────────────────

// ListProviders returns all federation providers.
func (s *FederationService) ListProviders() ([]database.FederationProvider, error) {
	var list []database.FederationProvider
	return list, database.DB.Order("name asc").Find(&list).Error
}

// GetProvider returns a provider by ID.
func (s *FederationService) GetProvider(id int64) (*database.FederationProvider, error) {
	var p database.FederationProvider
	if err := database.DB.First(&p, id).Error; err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return &p, nil
}

// CreateProvider adds a new federation provider.
func (s *FederationService) CreateProvider(name, apiURL, apiKey string, syncUsers, syncPlans, syncTraffic bool) (*database.FederationProvider, error) {
	p := &database.FederationProvider{
		Name:        name,
		APIURL:      apiURL,
		APIKey:      apiKey,
		SyncUsers:   syncUsers,
		SyncPlans:   syncPlans,
		SyncTraffic: syncTraffic,
		Status:      "offline",
		CreatedAt:   time.Now().UnixMilli(),
	}
	if err := database.DB.Create(p).Error; err != nil {
		return nil, fmt.Errorf("create federation provider: %w", err)
	}

	// Test connection immediately
	go s.testConnection(p.ID)

	s.eventBus.Publish(events.Event{
		Type:    "federation.provider.created",
		Message: fmt.Sprintf("Federation provider %s (%s) created", name, apiURL),
	})
	return p, nil
}

// UpdateProvider updates a federation provider.
func (s *FederationService) UpdateProvider(ctx context.Context, id int64, name, apiURL, apiKey string, syncUsers, syncPlans, syncTraffic bool) error {
	updates := map[string]interface{}{
		"name":         name,
		"api_url":      apiURL,
		"sync_users":   syncUsers,
		"sync_plans":   syncPlans,
		"sync_traffic": syncTraffic,
	}
	if apiKey != "" {
		updates["api_key"] = apiKey
	}
	if err := database.DB.Model(&database.FederationProvider{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("update federation provider: %w", err)
	}
	return nil
}

// DeleteProvider removes a federation provider.
func (s *FederationService) DeleteProvider(id int64) error {
	return database.DB.Delete(&database.FederationProvider{}, id).Error
}

// ─── Sync Operations ────────────────────────────────────────────────

// SyncAll triggers sync for all enabled providers (exported for handler access).
func (s *FederationService) SyncAll() {
	s.syncAll()
}

// syncAll triggers sync for all enabled providers.
func (s *FederationService) syncAll() {
	providers, err := s.ListProviders()
	if err != nil {
		return
	}
	for _, p := range providers {
		if p.APIURL == "" {
			continue
		}
		s.syncWithProvider(&p)
	}
}

// SyncWithProvider performs sync with a single provider (exported for handler access).
func (s *FederationService) SyncWithProvider(p *database.FederationProvider) {
	s.syncWithProvider(p)
}

// syncWithProvider performs sync with a single provider.
func (s *FederationService) syncWithProvider(p *database.FederationProvider) {
	start := time.Now()

	if p.SyncUsers {
		if err := s.pushUsers(p); err != nil {
			database.DB.Model(p).Update("status", "error")
			return
		}
		if err := s.pullUsers(p); err != nil {
			database.DB.Model(p).Update("status", "error")
			return
		}
	}

	if p.SyncPlans {
		if err := s.pushPlans(p); err != nil {
			database.DB.Model(p).Update("status", "error")
			return
		}
		if err := s.pullPlans(p); err != nil {
			database.DB.Model(p).Update("status", "error")
			return
		}
	}

	if p.SyncTraffic {
		s.syncTraffic(p)
	}

	// Update status
	database.DB.Model(p).Updates(map[string]interface{}{
		"status":       "online",
		"last_sync_at": time.Now().UnixMilli(),
	})

	s.eventBus.Publish(events.Event{
		Type:    "federation.sync.completed",
		Message: fmt.Sprintf("Synced with %s (%s) in %v", p.Name, p.APIURL, time.Since(start)),
	})
}

// testConnection tests connectivity to a provider.
func (s *FederationService) testConnection(providerID int64) {
	p, err := s.GetProvider(providerID)
	if err != nil {
		return
	}

	status := "offline"
	// Simple health check
	if resp, err := s.fedGet(p, "/api/v1/health"); err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			status = "online"
		}
	}

	database.DB.Model(p).Update("status", status)
}

// ─── Sync Users ─────────────────────────────────────────────────────

func (s *FederationService) pushUsers(p *database.FederationProvider) error {
	var users []database.User
	if err := database.DB.Find(&users).Error; err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]interface{}{
		"users": users,
		"source": "vortexuipro",
	})

	resp, err := s.fedPost(p, "/api/v1/federation/users", body)
	if err != nil {
		return fmt.Errorf("push users: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (s *FederationService) pullUsers(p *database.FederationProvider) error {
	resp, err := s.fedGet(p, "/api/v1/federation/users")
	if err != nil {
		return fmt.Errorf("pull users: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Users []database.User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, u := range result.Users {
		dbUser := u
		dbUser.ID = 0 // Let DB assign ID
		database.DB.Where("username = ?", u.Username).Assign(dbUser).FirstOrCreate(&dbUser)
	}

	return nil
}

// ─── Sync Plans ─────────────────────────────────────────────────────

func (s *FederationService) pushPlans(p *database.FederationProvider) error {
	var plans []database.Plan
	if err := database.DB.Find(&plans).Error; err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]interface{}{
		"plans":  plans,
		"source": "vortexuipro",
	})

	resp, err := s.fedPost(p, "/api/v1/federation/plans", body)
	if err != nil {
		return fmt.Errorf("push plans: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (s *FederationService) pullPlans(p *database.FederationProvider) error {
	resp, err := s.fedGet(p, "/api/v1/federation/plans")
	if err != nil {
		return fmt.Errorf("pull plans: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Plans []database.Plan `json:"plans"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, pl := range result.Plans {
		dbPlan := pl
		dbPlan.ID = 0
		database.DB.Where("name = ?", pl.Name).Assign(dbPlan).FirstOrCreate(&dbPlan)
	}

	return nil
}

// ─── Sync Traffic ────────────────────────────────────────────────────

func (s *FederationService) syncTraffic(p *database.FederationProvider) {
	// Send local traffic summary
	var users []database.User
	database.DB.Select("id, username, traffic_up, traffic_down").Find(&users)

	trafficData := make([]map[string]interface{}, len(users))
	for i, u := range users {
		trafficData[i] = map[string]interface{}{
			"username":     u.Username,
			"traffic_up":   u.TrafficUp,
			"traffic_down": u.TrafficDown,
		}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"traffic": trafficData,
		"source":  "vortexuipro",
	})

	if resp, err := s.fedPost(p, "/api/v1/federation/traffic", body); err == nil {
		resp.Body.Close()
	}
}

// ─── Federation API Handlers (called by remote panels) ─────────────

// HandleFederationUsers handles incoming user sync requests.
func (s *FederationService) HandleFederationUsers(body []byte) (interface{}, error) {
	// Check if this is a push (has "users" field) or pull (no body)
	var req struct {
		Users  []database.User `json:"users,omitempty"`
		Source string          `json:"source,omitempty"`
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err == nil && len(req.Users) > 0 {
			// Push: store received users
			for _, u := range req.Users {
				dbUser := u
				dbUser.ID = 0
				database.DB.Where("username = ?", u.Username).Assign(dbUser).FirstOrCreate(&dbUser)
			}
			return map[string]interface{}{"synced": len(req.Users), "status": "ok"}, nil
		}
	}

	// Pull: return local users
	var users []database.User
	database.DB.Find(&users)
	return map[string]interface{}{"users": users, "total": len(users)}, nil
}

// HandleFederationPlans handles incoming plan sync requests.
func (s *FederationService) HandleFederationPlans(body []byte) (interface{}, error) {
	var req struct {
		Plans  []database.Plan `json:"plans,omitempty"`
		Source string          `json:"source,omitempty"`
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err == nil && len(req.Plans) > 0 {
			for _, pl := range req.Plans {
				dbPlan := pl
				dbPlan.ID = 0
				database.DB.Where("name = ?", pl.Name).Assign(dbPlan).FirstOrCreate(&dbPlan)
			}
			return map[string]interface{}{"synced": len(req.Plans), "status": "ok"}, nil
		}
	}

	var plans []database.Plan
	database.DB.Find(&plans)
	return map[string]interface{}{"plans": plans, "total": len(plans)}, nil
}

// HandleFederationTraffic handles incoming traffic sync.
func (s *FederationService) HandleFederationTraffic(body []byte) (interface{}, error) {
	var req struct {
		Traffic []struct {
			Username     string `json:"username"`
			TrafficUp    int64  `json:"traffic_up"`
			TrafficDown  int64  `json:"traffic_down"`
		} `json:"traffic"`
		Source string `json:"source"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid traffic data: %w", err)
	}

	updated := 0
	for _, t := range req.Traffic {
		result := database.DB.Model(&database.User{}).
			Where("username = ?", t.Username).
			Updates(map[string]interface{}{
				"traffic_up":   t.TrafficUp,
				"traffic_down": t.TrafficDown,
			})
		if result.Error == nil && result.RowsAffected > 0 {
			updated++
		}
	}

	return map[string]interface{}{
		"synced":  updated,
		"status":  "ok",
	}, nil
}

// ─── Internal HTTP Helpers ──────────────────────────────────────────

func (s *FederationService) fedGet(p *database.FederationProvider, path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", p.APIURL+path, nil)
	if err != nil {
		return nil, err
	}
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
		req.Header.Set("X-Federation-Key", p.APIKey)
	}
	return s.client.Do(req)
}

func (s *FederationService) fedPost(p *database.FederationProvider, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", p.APIURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
		req.Header.Set("X-Federation-Key", p.APIKey)
	}
	return s.client.Do(req)
}
