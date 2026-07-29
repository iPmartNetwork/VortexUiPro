package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"sync"
	"time"

	"vortexuipro/internal/api/hub"
	"vortexuipro/internal/database"
)

// ─── Cluster Manager (Orchestrator) ──────────────────────────────────

// Config holds cluster configuration.
type Config struct {
	Enabled        bool
	NodeName       string
	NodeID         int64
	Addr           string   // this node's mesh address
	Peers          []string // list of peer addresses (ip:port)
	Region         string
	Priority       int
	HeartbeatInterval time.Duration
	HeartbeatTimeout   time.Duration
	PKIDir         string   // PKI certificates directory
	GRPCEnabled    bool     // enable gRPC streaming
	TLSEnabled     bool     // enable mTLS for mesh
	WebSocketHub   *hub.Hub // WebSocket hub for real-time streaming
}

// Manager is the top-level cluster orchestrator.
type Manager struct {
	mu sync.RWMutex

	config      Config
	nodeID      int64

	// Sub-components
	peerServer *PeerServer
	election   *LeaderElection
	syncSvc    *SyncService
	resolver   *ConflictResolver
	pki        *PKIManager
	grpcSvc    *GRPCService
	topology   *TopologyBroadcaster

	// Peer list (built from config + discovered peers)
	peers   []ClusterPeer
	peerClients map[int64]*PeerClient

	// Heartbeat
	heartbeatTicker *time.Ticker

	// Sync queue flush
	flushTicker *time.Ticker

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	started bool
	stopCh  chan struct{}
}

// NewManager creates a new cluster manager.
func NewManager(cfg Config) *Manager {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 5 * time.Second
	}
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = 15 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:      cfg,
		peers:       make([]ClusterPeer, 0),
		peerClients: make(map[int64]*PeerClient),
		ctx:         ctx,
		cancel:      cancel,
		stopCh:      make(chan struct{}),
	}

	// Register self as a peer
	selfPeer := ClusterPeer{
		ID:       cfg.NodeID,
		Name:     cfg.NodeName,
		Address:  cfg.Addr,
		Priority: cfg.Priority,
		Online:   true,
	}
	m.peers = append(m.peers, selfPeer)

	// Add configured peers
	for i, addr := range cfg.Peers {
		peerID := int64(i + 1000) // temporary ID, will be updated on first heartbeat
		m.peers = append(m.peers, ClusterPeer{
			ID:      peerID,
			Name:    fmt.Sprintf("peer-%d", peerID),
			Address: addr,
			Online:  false,
		})
		m.peerClients[peerID] = NewPeerClient(addr)
	}

	return m
}

// Start initializes and starts all cluster sub-systems.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	if !m.config.Enabled {
		log.Println("Cluster mode is disabled")
		return nil
	}

	// 1. Register or get node ID from database
	if err := m.registerNode(); err != nil {
		return fmt.Errorf("register cluster node: %w", err)
	}

	// 2. Initialize PKI / mTLS certificates
	if m.config.TLSEnabled || m.config.PKIDir != "" {
		pkiCfg := PKIConfig{
			CADir:   path.Join(m.config.PKIDir, "ca"),
			NodeDir: path.Join(m.config.PKIDir, "node"),
			Org:     "VortexUiPro",
			Validity:    365 * 24 * time.Hour,
			GracePeriod: 30 * 24 * time.Hour,
		}
		pki, err := NewPKIManager(pkiCfg)
		if err != nil {
			log.Printf("Warning: PKI init failed (mTLS disabled): %v", err)
		} else {
			m.pki = pki
			m.pki.StartAutoRenew()
			log.Println("Cluster PKI initialized (mTLS enabled)")
		}
	}

	// 3. Start peer server (mesh message listener with optional mTLS)
	if m.pki != nil && m.config.TLSEnabled {
		tlsCfg, err := TLSConfig(m.pki, true)
		if err != nil {
			log.Printf("Warning: mTLS config failed, using plaintext: %v", err)
			m.peerServer = NewPeerServer(m.config.Addr)
		} else {
			m.peerServer = NewPeerServerTLS(m.config.Addr, tlsCfg)
		}
	} else {
		m.peerServer = NewPeerServer(m.config.Addr)
	}
	m.peerServer.OnMessage(m.handleMeshMessage)
	if err := m.peerServer.Start(); err != nil {
		return fmt.Errorf("start peer server: %w", err)
	}

	// 4. Initialize gRPC mesh service (alongside HTTP mesh)
	if m.config.GRPCEnabled {
		m.grpcSvc = NewGRPCService(m.config, m.pki)
		m.grpcSvc.OnMessage(m.handleMeshMessage)
		if err := m.grpcSvc.Start(); err != nil {
			log.Printf("Warning: gRPC mesh start failed: %v", err)
		} else {
			log.Println("gRPC mesh service started")
		}
	}

	// 5. Initialize topology broadcaster for real-time WebSocket streaming
	m.topology = NewTopologyBroadcaster()
	if m.config.WebSocketHub != nil {
		m.topology.SetBroadcastFn(func(event TopologyEvent) {
			m.config.WebSocketHub.Broadcast(hub.Message{
				Type:    hub.MsgTopology,
				Payload: event,
				Time:    time.Now().UnixMilli(),
			})
		})
		m.topology.Start()
		log.Println("Topology broadcaster started (WebSocket streaming)")
	}

	// 6. Initialize conflict resolver (shared between election and sync)
	m.resolver = NewConflictResolver()
	for _, peer := range m.peers {
		m.resolver.SetNodePriority(peer.ID, peer.Priority)
	}

	// 7. Initialize election
	m.election = NewLeaderElection(m.nodeID, m.config.NodeName, m.config.Priority)
	m.election.SetPeers(m.peers)
	m.election.OnLeaderElected(m.onLeaderElected)
	m.election.OnVoteRequest(m.onVoteRequest)
	m.election.Start()

	// 8. Initialize sync service with shared resolver
	m.syncSvc = NewSyncServiceWithResolver(m.nodeID, m.config.NodeName, m.getPeers, m.isLeader, m.getLeaderAddr, m.resolver)

	// 9. Start heartbeat sender (leader broadcasts, followers listen)
	m.heartbeatTicker = time.NewTicker(m.config.HeartbeatInterval)
	go m.heartbeatLoop()

	// 10. Start sync service
	m.syncSvc.Start()

	// 11. Periodic cleanup of conflict resolver
	go func() {
		cleanupTicker := time.NewTicker(30 * time.Minute)
		defer cleanupTicker.Stop()
		for {
			select {
			case <-cleanupTicker.C:
				m.resolver.Cleanup()
			case <-m.stopCh:
				return
			}
		}
	}()

	// 12. Update node status in database
	m.updateNodeStatus("online")

	m.started = true
	log.Printf("Cluster manager started: node=%s addr=%s region=%s priority=%d peers=%d",
		m.config.NodeName, m.config.Addr, m.config.Region, m.config.Priority, len(m.config.Peers))

	return nil
}

// Stop gracefully shuts down all cluster sub-systems.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	log.Println("Shutting down cluster manager...")

	if m.heartbeatTicker != nil {
		m.heartbeatTicker.Stop()
	}
	if m.election != nil {
		m.election.Stop()
	}
	if m.syncSvc != nil {
		m.syncSvc.Stop()
	}
	if m.topology != nil {
		m.topology.Stop()
	}
	if m.pki != nil {
		m.pki.Stop()
	}
	if m.grpcSvc != nil {
		m.grpcSvc.Stop()
	}
	if m.peerServer != nil {
		m.peerServer.Stop()
	}

	m.cancel()
	m.updateNodeStatus("offline")

	close(m.stopCh)
	m.started = false
	log.Println("Cluster manager stopped")
}

// ─── Mesh Message Handler ────────────────────────────────────────────

func (m *Manager) handleMeshMessage(msg MeshMessage) (*MeshMessage, error) {
	switch msg.Type {
	case MsgHeartbeat:
		return m.handleHeartbeatMsg(msg), nil

	case MsgVoteRequest:
		if m.election != nil {
			return m.election.HandleVoteRequest(msg), nil
		}

	case MsgVoteResponse:
		// Handled internally by election

	case MsgLeaderAnnounce:
		if m.election != nil {
			m.election.HandleLeaderAnnounce(msg)
		}
		// Forward to sync service to request full sync if new leader
		go func() {
			var announce LeaderAnnouncePayload
			json.Unmarshal(msg.Payload, &announce)
			m.onLeaderElected(announce.LeaderID, announce.Term)
		}()

	case MsgSyncData:
		if m.syncSvc != nil {
			return m.syncSvc.HandleSyncData(msg), nil
		}

	case MsgSyncRequest:
		if m.syncSvc != nil && m.election != nil && m.election.IsLeader() {
			return m.syncSvc.HandleFullSyncRequest(), nil
		}

	case MsgSyncAck:
		// Handled internally by sync service
	}

	return nil, nil
}

func (m *Manager) handleHeartbeatMsg(msg MeshMessage) *MeshMessage {
	// Update peer state
	var hb HeartbeatPayload
	if err := json.Unmarshal(msg.Payload, &hb); err != nil {
		return nil
	}

	// Use heartbeat address if provided, otherwise fall back to HTTP RemoteAddr
	peerAddr := hb.Address
	if peerAddr == "" {
		peerAddr = m.extractAddr(msg)
	}

	m.mu.Lock()
	// Update or add peer
	found := false
	for i, peer := range m.peers {
		if peer.ID == hb.NodeID || peer.Address == peerAddr {
			m.peers[i].Online = true
			m.peers[i].Name = hb.Name
			m.peers[i].Address = peerAddr
			m.peers[i].Priority = int(hb.CPULoad * 100)
			m.peers[i].Term = hb.Term
			found = true
			break
		}
	}
	if !found {
		newID := hb.NodeID
		if newID == 0 {
			newID = int64(len(m.peers) + 1)
		}
		m.peers = append(m.peers, ClusterPeer{
			ID:      newID,
			Name:    hb.Name,
			Address: peerAddr,
			Online:  true,
		})
		m.peerClients[newID] = NewPeerClient(peerAddr)
	}
	m.mu.Unlock()

	// Update election heartbeat
	if m.election != nil {
		m.election.HandleHeartbeat()
	}

	// Update database node info
	m.updateNodeHeartbeat(hb)

	// Respond with our heartbeat
	return m.buildHeartbeatMsg()
}

func (m *Manager) extractAddr(msg MeshMessage) string {
	// The sender address is extracted from the HTTP request in peer.go handleMesh
	if msg.RemoteAddr != "" {
		// Strip port from RemoteAddr to get the IP
		for i := len(msg.RemoteAddr) - 1; i >= 0; i-- {
			if msg.RemoteAddr[i] == ':' {
				return msg.RemoteAddr[:i]
			}
		}
		return msg.RemoteAddr
	}
	return ""
}

// ─── Heartbeat ───────────────────────────────────────────────────────

func (m *Manager) heartbeatLoop() {
	for {
		select {
		case <-m.heartbeatTicker.C:
			m.sendHeartbeat()
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) sendHeartbeat() {
	if m.election == nil {
		return
	}

	isLeader := m.election.IsLeader()

	// Build our heartbeat payload
	hbPayload := HeartbeatPayload{
		NodeID:  m.nodeID,
		Name:    m.config.NodeName,
		Address: m.config.Addr,
		Role:    string(m.election.GetRole()),
		Status:  "online",
		Term:    m.election.GetTerm(),
		Version: "0.0.1",
		Region:  m.config.Region,
	}

	// Count users
	if database.DB != nil {
		var count int64
		database.DB.Model(&database.User{}).Count(&count)
		hbPayload.UserCount = int(count)
	}

	payloadBytes, _ := json.Marshal(hbPayload)
	msg := MeshMessage{
		Type:       MsgHeartbeat,
		SenderID:   m.nodeID,
		SenderName: m.config.NodeName,
		Term:       m.election.GetTerm(),
		Timestamp:  time.Now().UnixMilli(),
		Payload:    payloadBytes,
	}

	// Broadcast to all peers
	m.mu.RLock()
	peers := make([]ClusterPeer, len(m.peers))
	copy(peers, m.peers)
	m.mu.RUnlock()

	for _, peer := range peers {
		if peer.ID == m.nodeID || !peer.Online {
			continue
		}
		client := m.peerClients[peer.ID]
		if client == nil {
			client = NewPeerClient(peer.Address)
			m.mu.Lock()
			m.peerClients[peer.ID] = client
			m.mu.Unlock()
		}
		if _, err := client.Send(msg); err != nil {
			// Mark as offline on failure
			m.mu.Lock()
			for i, p := range m.peers {
				if p.ID == peer.ID {
					m.peers[i].Online = false
					break
				}
			}
			m.mu.Unlock()
			log.Printf("Heartbeat to %s (%s) failed: %v", peer.Name, peer.Address, err)
		}
	}

	// If leader, also update own heartbeat in election
	if isLeader {
		m.election.HandleHeartbeat()
	}
}

func (m *Manager) buildHeartbeatMsg() *MeshMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	role := RoleFollower
	if m.election != nil {
		role = m.election.GetRole()
	}

	hbPayload := HeartbeatPayload{
		NodeID:  m.nodeID,
		Name:    m.config.NodeName,
		Address: m.config.Addr,
		Role:    string(role),
		Status:  "online",
		Term:    0,
		Version: "0.0.1",
		Region:  m.config.Region,
	}

	if m.election != nil {
		hbPayload.Term = m.election.GetTerm()
	}

	payloadBytes, _ := json.Marshal(hbPayload)
	return &MeshMessage{
		Type:       MsgHeartbeat,
		SenderID:   m.nodeID,
		SenderName: m.config.NodeName,
		Timestamp:  time.Now().UnixMilli(),
		Payload:    payloadBytes,
	}
}

// ─── Election Callbacks ──────────────────────────────────────────────

func (m *Manager) onLeaderElected(leaderID int64, term int64) {
	log.Printf("Leader elected: id=%d term=%d", leaderID, term)

	m.mu.Lock()
	// Update node role in database
	database.DB.Model(&database.ClusterNode{}).Where("id = ?", m.nodeID).
		Update("role", "follower")

	if leaderID == m.nodeID {
		database.DB.Model(&database.ClusterNode{}).Where("id = ?", m.nodeID).
			Update("role", "leader")
	}
	m.mu.Unlock()

	// Request full sync if we're not the leader
	if leaderID != m.nodeID {
		leaderAddr := m.getPeerAddress(leaderID)
		if leaderAddr != "" && m.syncSvc != nil {
			go func() {
				log.Printf("Requesting full sync from leader %s", leaderAddr)
				if err := m.syncSvc.RequestFullSync(leaderAddr); err != nil {
					log.Printf("Full sync from leader failed: %v", err)
				} else {
					log.Println("Full sync from leader completed")
				}
			}()
		}
	}
}

func (m *Manager) onVoteRequest(candidateID int64, term int64) bool {
	return term > 0 // Simplified: grant vote if term is valid
}

// ─── Helpers ─────────────────────────────────────────────────────────

func (m *Manager) registerNode() error {
	// Check if node exists in database
	var existing database.ClusterNode
	result := database.DB.Where("name = ?", m.config.NodeName).First(&existing)

	if result.Error == nil {
		m.nodeID = existing.ID
		// Update address and other dynamic info
		database.DB.Model(&existing).Updates(map[string]any{
			"address":        m.config.Addr,
			"priority":       m.config.Priority,
			"region":         m.config.Region,
			"version":        "0.0.1",
			"advertise_addr": m.config.Addr,
		})
		return nil
	}

	// Create new node
	node := database.ClusterNode{
		Name:          m.config.NodeName,
		Address:       m.config.Addr,
		PeerPort:      1337,
		Role:          string(RoleFollower),
		Status:        "offline",
		Priority:      m.config.Priority,
		Region:        m.config.Region,
		Version:       "0.0.1",
		AdvertiseAddr: m.config.Addr,
		Enabled:       true,
	}

	if err := database.DB.Create(&node).Error; err != nil {
		return err
	}

	m.nodeID = node.ID
	return nil
}

func (m *Manager) updateNodeStatus(status string) {
	if database.DB == nil {
		return
	}
	database.DB.Model(&database.ClusterNode{}).Where("id = ?", m.nodeID).
		Update("status", status)
}

func (m *Manager) updateNodeHeartbeat(hb HeartbeatPayload) {
	if database.DB == nil {
		return
	}
	database.DB.Model(&database.ClusterNode{}).Where("id = ?", hb.NodeID).
		Updates(map[string]any{
			"status":          hb.Status,
			"cpu_load":        hb.CPULoad,
			"memory_used":     hb.MemUsed,
			"last_heartbeat":  time.Now().UnixMilli(),
			"role":            hb.Role,
		})
}

func (m *Manager) getPeerAddress(peerID int64) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, peer := range m.peers {
		if peer.ID == peerID {
			return peer.Address
		}
	}
	return ""
}

func (m *Manager) getPeers() []ClusterPeer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peers := make([]ClusterPeer, len(m.peers))
	copy(peers, m.peers)
	return peers
}

func (m *Manager) isLeader() bool {
	if m.election == nil {
		return false
	}
	return m.election.IsLeader()
}

func (m *Manager) getLeaderAddr() string {
	if m.election == nil {
		return ""
	}
	leaderID := m.election.GetLeaderID()
	return m.getPeerAddress(leaderID)
}

// ─── Stats & Status ──────────────────────────────────────────────────

// Status returns the current cluster status.
func (m *Manager) Status() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peerList := make([]map[string]any, 0)
	for _, peer := range m.peers {
		peerList = append(peerList, map[string]any{
			"id":       peer.ID,
			"name":     peer.Name,
			"address":  peer.Address,
			"priority": peer.Priority,
			"online":   peer.Online,
		})
	}

	electionStats := map[string]any{}
	if m.election != nil {
		electionStats = m.election.Stats()
	}

	resolverStats := map[string]any{}
	if m.resolver != nil {
		resolverStats = m.resolver.Stats()
	}

	return map[string]any{
		"enabled":         m.config.Enabled,
		"node_id":         m.nodeID,
		"node_name":       m.config.NodeName,
		"addr":            m.config.Addr,
		"region":          m.config.Region,
		"peers":           peerList,
		"started":         m.started,
		"election":        electionStats,
		"conflict_resolver": resolverStats,
	}
}

// GetNodeID returns this node's ID.
func (m *Manager) GetNodeID() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodeID
}

// GetNodeName returns this node's name.
func (m *Manager) GetNodeName() string {
	return m.config.NodeName
}

// IsEnabled returns whether cluster mode is enabled.
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}
