package monitor

import (
	"context"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/logger"
)

const (
	nodeTrafficSyncConcurrency    = 8
	nodeTrafficSyncTimeout        = 4 * time.Second
	nodeTrafficSyncInterval       = 5 * time.Second
	nodeClientIpSyncInterval      = 10 * time.Second
)

// NodeTrafficSyncJob syncs traffic data from remote nodes.
type NodeTrafficSyncJob struct {
	running      sync.Mutex
	lastIpSync   int64
	lastSync     int64
	ipSyncMu     sync.Mutex
	activityMu   sync.Mutex
	prevTotals   map[string]inboundSample
}

type inboundSample struct {
	up, down, at int64
}

// NewNodeTrafficSyncJob creates a new traffic sync job.
func NewNodeTrafficSyncJob() *NodeTrafficSyncJob {
	return &NodeTrafficSyncJob{
		prevTotals: make(map[string]inboundSample),
	}
}

// Run executes one sync cycle — fetches traffic from all online nodes.
func (j *NodeTrafficSyncJob) Run() {
	if !j.running.TryLock() {
		return
	}
	defer j.running.Unlock()

	// Load enabled nodes from database
	var nodes []database.Node
	if err := database.DB.Where("enable = ? AND status = ?", true, "online").Find(&nodes).Error; err != nil {
		logger.Warnf("node traffic sync: load nodes failed: %v", err)
		return
	}
	if len(nodes) == 0 {
		return
	}

	// Check if we should sync client IPs this tick
	doIpSync := false
	j.ipSyncMu.Lock()
	now := time.Now().Unix()
	if now-j.lastIpSync >= int64(nodeClientIpSyncInterval/time.Second) {
		doIpSync = true
		j.lastIpSync = now
	}
	j.ipSyncMu.Unlock()

	// Sync each node concurrently
	sem := make(chan struct{}, nodeTrafficSyncConcurrency)
	var wg sync.WaitGroup

	for _, n := range nodes {
		if !n.Enable || n.Status != "online" {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		n := n
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			j.syncNode(n, doIpSync)
		}()
	}
	wg.Wait()

	j.lastSync = time.Now().Unix()
}

func (j *NodeTrafficSyncJob) syncNode(n database.Node, doIpSync bool) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), nodeTrafficSyncTimeout)
	defer cancel()

	// Fetch traffic data from node via its API
	// For now, just log the sync attempt
	logger.Infof("node traffic sync: pulling from %s (%s:%d)", n.Name, n.Address, n.Port)
	_ = ctx

	if doIpSync {
		logger.Infof("node traffic sync: client IP sync for %s", n.Name)
	}
}

// Stats returns sync job statistics.
func (j *NodeTrafficSyncJob) Stats() map[string]any {
	return map[string]any{
		"running":    true,
		"last_sync":  j.lastSync,
		"last_ip_sync": j.lastIpSync,
	}
}
