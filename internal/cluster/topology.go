package cluster

import (
	"log"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Topology Event Types ────────────────────────────────────────────

type TopologyEventType string

const (
	TopoNodeJoined   TopologyEventType = "node_joined"
	TopoNodeLeft     TopologyEventType = "node_left"
	TopoNodeUpdated  TopologyEventType = "node_updated"
	TopoLeaderChanged TopologyEventType = "leader_changed"
	TopoSyncStarted  TopologyEventType = "sync_started"
	TopoSyncComplete TopologyEventType = "sync_complete"
	TopoElection     TopologyEventType = "election"
)

// TopologyEvent is broadcast via WebSocket.
type TopologyEvent struct {
	Type      TopologyEventType `json:"type"`
	NodeID    int64             `json:"node_id,omitempty"`
	NodeName  string            `json:"node_name,omitempty"`
	LeaderID  int64             `json:"leader_id,omitempty"`
	Term      int64             `json:"term,omitempty"`
	Peers     []TopologyPeer    `json:"peers,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

// TopologyPeer represents a node in the topology graph.
type TopologyPeer struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Role         string  `json:"role"`
	Status       string  `json:"status"`
	CPULoad      float64 `json:"cpu_load"`
	MemoryUsed   float64 `json:"memory_used"`
	Latency      int64   `json:"latency_ms"`
	Region       string  `json:"region"`
	Priority     int     `json:"priority"`
	UserCount    int     `json:"user_count"`
	LastHeartbeat int64  `json:"last_heartbeat"`
	Online       bool    `json:"online"`
	IsLeader     bool    `json:"is_leader"`
}

// TopologyBroadcaster sends real-time topology updates via WebSocket.
type TopologyBroadcaster struct {
	mu          sync.RWMutex
	peers       map[int64]*TopologyPeer
	broadcastFn func(event TopologyEvent)
	stopCh      chan struct{}
}

// NewTopologyBroadcaster creates a new topology broadcaster.
func NewTopologyBroadcaster() *TopologyBroadcaster {
	return &TopologyBroadcaster{
		peers:  make(map[int64]*TopologyPeer),
		stopCh: make(chan struct{}),
	}
}

// SetBroadcastFn sets the callback for broadcasting topology events.
func (tb *TopologyBroadcaster) SetBroadcastFn(fn func(event TopologyEvent)) {
	tb.mu.Lock()
	tb.broadcastFn = fn
	tb.mu.Unlock()
}

// Start begins periodic topology snapshots.
func (tb *TopologyBroadcaster) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tb.broadcastSnapshot()
			case <-tb.stopCh:
				return
			}
		}
	}()
	log.Println("Topology broadcaster started (5s interval)")
}

// Stop stops the broadcaster.
func (tb *TopologyBroadcaster) Stop() {
	close(tb.stopCh)
}

// UpdateNode updates or adds a peer in the topology.
func (tb *TopologyBroadcaster) UpdateNode(hb HeartbeatPayload, nodeID int64, isLeader bool) {
	tb.mu.Lock()
	peer, exists := tb.peers[nodeID]
	if !exists {
		peer = &TopologyPeer{}
		tb.peers[nodeID] = peer
	}
	peer.ID = nodeID
	peer.Name = hb.Name
	peer.Role = hb.Role
	peer.Status = "online"
	peer.CPULoad = hb.CPULoad
	peer.MemoryUsed = hb.MemUsed
	peer.Region = hb.Region
	peer.LastHeartbeat = time.Now().UnixMilli()
	peer.Online = true
	peer.IsLeader = isLeader
	peer.UserCount = hb.UserCount
	tb.mu.Unlock()

	if !exists {
		tb.broadcast(TopoNodeJoined)
	}
	tb.broadcast(TopoNodeUpdated)
}

// RemoveNode marks a node as offline.
func (tb *TopologyBroadcaster) RemoveNode(nodeID int64) {
	tb.mu.Lock()
	if peer, ok := tb.peers[nodeID]; ok {
		peer.Online = false
		peer.Status = "offline"
	}
	tb.mu.Unlock()
	tb.broadcast(TopoNodeLeft)
}

// SetLeader marks a node as the leader.
func (tb *TopologyBroadcaster) SetLeader(nodeID int64, term int64) {
	tb.mu.Lock()
	for _, peer := range tb.peers {
		peer.IsLeader = peer.ID == nodeID
	}
	tb.mu.Unlock()
	tb.broadcast(TopoLeaderChanged)
}

// Snapshot returns the current topology state.
func (tb *TopologyBroadcaster) Snapshot() []TopologyPeer {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	peers := make([]TopologyPeer, 0, len(tb.peers))
	for _, p := range tb.peers {
		peers = append(peers, *p)
	}
	return peers
}

func (tb *TopologyBroadcaster) broadcast(eventType TopologyEventType) {
	tb.mu.RLock()
	peers := make([]TopologyPeer, 0, len(tb.peers))
	for _, p := range tb.peers {
		peers = append(peers, *p)
	}
	fn := tb.broadcastFn
	tb.mu.RUnlock()

	if fn != nil {
		fn(TopologyEvent{
			Type:      eventType,
			Peers:     peers,
			Timestamp: time.Now().UnixMilli(),
		})
	}
}

func (tb *TopologyBroadcaster) broadcastSnapshot() {
	tb.mu.RLock()
	peers := make([]TopologyPeer, 0, len(tb.peers))
	for _, p := range tb.peers {
		peers = append(peers, *p)
	}
	fn := tb.broadcastFn
	tb.mu.RUnlock()

	if fn != nil {
		fn(TopologyEvent{
			Type:      TopoNodeUpdated,
			Peers:     peers,
			Timestamp: time.Now().UnixMilli(),
		})
	}
}

// ─── Topology API ────────────────────────────────────────────────────

// TopologyAPI generates a topology response for the REST API.
func TopologyAPI() map[string]any {
	peers := make([]map[string]any, 0)

	var nodes []database.ClusterNode
	database.DB.Find(&nodes)
	for _, n := range nodes {
		isOnline := time.Now().UnixMilli()-n.LastHeartbeat < 30000 // 30s
		peers = append(peers, map[string]any{
			"id":             n.ID,
			"name":           n.Name,
			"address":        n.Address,
			"role":           n.Role,
			"status":         n.Status,
			"priority":       n.Priority,
			"region":         n.Region,
			"term":           n.Term,
			"cpu_load":       n.CPULoad,
			"memory_used":    n.MemoryUsed,
			"last_heartbeat": n.LastHeartbeat,
			"online":         isOnline,
			"enabled":        n.Enabled,
		})
	}

	return map[string]any{
		"peers":   peers,
		"total":   len(peers),
		"online":  countOnline(peers),
	}
}

func countOnline(peers []map[string]any) int {
	count := 0
	for _, p := range peers {
		if online, ok := p["online"].(bool); ok && online {
			count++
		}
	}
	return count
}
