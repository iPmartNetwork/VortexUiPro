package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// ─── Constants ───────────────────────────────────────────────────────

const (
	WebRTCPeerStatusConnecting = "connecting"
	WebRTCPeerStatusConnected  = "connected"
	WebRTCPeerStatusDisconnected = "disconnected"
	WebRTCPeerStatusFailed     = "failed"

	TURNServerStatusOnline  = "online"
	TURNServerStatusOffline = "offline"

	P2PMeshRoleRelay   = "relay"
	P2PMeshRoleDirect  = "direct"
	P2PMeshRoleHybrid  = "hybrid"

	DefaultSTUNServer = "stun:stun.l.google.com:19302"
)

// ─── Data Types ──────────────────────────────────────────────────────

// ICEConfig holds STUN/TURN server configuration for WebRTC peers.
type ICEConfig struct {
	STUNServers []string `json:"stun_servers"`
	TURNServers []TURNConfig `json:"turn_servers"`
	ICEServers  []ICEServer `json:"ice_servers"`
}

// ICEServer represents a single ICE server (STUN or TURN).
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// TURNConfig holds TURN server configuration.
type TURNConfig struct {
	ID         int64  `json:"id"`
	Address    string `json:"address"`     // e.g., turn:turn.example.com:3478
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	Realm      string `json:"realm,omitempty"`
	Protocol   string `json:"protocol"`    // udp, tcp, tls
	Status     string `json:"status"`
	Region     string `json:"region,omitempty"`
	Bandwidth  int    `json:"bandwidth"`  // Mbps
	CreatedAt  int64  `json:"created_at"`
}

// WebRTCPeer represents a WebRTC peer connection.
type WebRTCPeer struct {
	ID         string    `json:"id"`
	NodeID     string    `json:"node_id"`
	Name       string    `json:"name"`
	Protocol   string    `json:"protocol"`    // webrtc, datachannel
	Mode       string    `json:"mode"`        // relay, direct, hybrid
	Status     string    `json:"status"`
	LocalAddr  string    `json:"local_addr,omitempty"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
	Latency    float64   `json:"latency,omitempty"`  // ms
	Bandwidth  float64   `json:"bandwidth,omitempty"` // Mbps
	ConnectedAt int64    `json:"connected_at,omitempty"`
	LastSeen   int64     `json:"last_seen"`
	CreatedAt  int64     `json:"created_at"`
}

// P2PMeshConfig holds P2P mesh network configuration.
type P2PMeshConfig struct {
	Enabled        bool   `json:"enabled"`
	MeshName       string `json:"mesh_name"`
	Role           string `json:"role"`            // relay, direct, hybrid
	ListenPort     int    `json:"listen_port"`
	MaxPeers       int    `json:"max_peers"`
	AutoReconnect  bool   `json:"auto_reconnect"`
	Discovery      string `json:"discovery"`       // dns, manual, signaling
	Encryption     bool   `json:"encryption"`
	HeartbeatSec   int    `json:"heartbeat_sec"`
	DataChannel    bool   `json:"data_channel"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// SignalingMessage represents a WebRTC signaling message over WebSocket.
type SignalingMessage struct {
	Type      string          `json:"type"`      // offer, answer, ice_candidate, hangup, join, leave
	FromID    string          `json:"from_id"`
	ToID      string          `json:"to_id,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate string          `json:"candidate,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// ─── WebRTCService ───────────────────────────────────────────────────

type WebRTCService struct {
	mu         sync.RWMutex
	eventBus   *events.Bus
	peers      map[string]*WebRTCPeer
	turnServers []database.TURNServer
	meshConfig *P2PMeshConfig
	signaling  chan SignalingMessage

	// ICE config
	iceConfig *ICEConfig

	started bool
	stopCh  chan struct{}
}

// NewWebRTCService creates a new WebRTC service.
func NewWebRTCService(eventBus *events.Bus) *WebRTCService {
	svc := &WebRTCService{
		eventBus:   eventBus,
		peers:      make(map[string]*WebRTCPeer),
		turnServers: make([]database.TURNServer, 0),
		signaling:  make(chan SignalingMessage, 256),
		stopCh:     make(chan struct{}),
		iceConfig: &ICEConfig{
			STUNServers: []string{DefaultSTUNServer},
			TURNServers: make([]TURNConfig, 0),
		},
		meshConfig: &P2PMeshConfig{
			Enabled:       false,
			MeshName:      "vortexuipro-mesh",
			Role:          P2PMeshRoleHybrid,
			ListenPort:    0,
			MaxPeers:      10,
			AutoReconnect: true,
			Discovery:     "signaling",
			Encryption:    true,
			HeartbeatSec:  30,
			DataChannel:   true,
		},
	}
	return svc
}

// Start initializes the WebRTC service and loads config from database.
func (s *WebRTCService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	// Load TURN servers from database
	var dbServers []database.TURNServer
	if err := database.DB.Find(&dbServers).Error; err != nil {
		log.Printf("Warning: failed to load TURN servers: %v", err)
	}
	s.turnServers = dbServers

	// Load mesh config from database
	var meshCfg database.P2PMeshConfig
	if err := database.DB.First(&meshCfg).Error; err == nil {
		s.meshConfig.Enabled = meshCfg.Enabled
		s.meshConfig.MeshName = meshCfg.MeshName
		s.meshConfig.Role = meshCfg.Role
		s.meshConfig.ListenPort = meshCfg.ListenPort
		s.meshConfig.MaxPeers = meshCfg.MaxPeers
		s.meshConfig.AutoReconnect = meshCfg.AutoReconnect
		s.meshConfig.HeartbeatSec = meshCfg.HeartbeatSec
	}

	// Rebuild ICE config
	s.rebuildICEConfig()

	// Start signaling processor
	go s.processSignaling()

	// Start heartbeat if mesh enabled
	if s.meshConfig.Enabled {
		go s.heartbeatLoop()
	}

	s.started = true
	log.Println("[Phase 10] WebRTC service started")
	return nil
}

// Stop shuts down the WebRTC service.
func (s *WebRTCService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}
	close(s.stopCh)
	s.started = false
	log.Println("[Phase 10] WebRTC service stopped")
}

// ─── ICE / STUN / TURN ──────────────────────────────────────────────

// GetICEConfig returns the ICE configuration for WebRTC peers.
func (s *WebRTCService) GetICEConfig() *ICEConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.iceConfig
}

// GetTURNServers returns the list of configured TURN servers.
func (s *WebRTCService) GetTURNServers() ([]database.TURNServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var servers []database.TURNServer
	if err := database.DB.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

// AddTURNServer adds a new TURN server to the database.
func (s *WebRTCService) AddTURNServer(server *database.TURNServer) error {
	server.CreatedAt = time.Now().UnixMilli()
	server.UpdatedAt = server.CreatedAt
	if err := database.DB.Create(server).Error; err != nil {
		return err
	}

	s.mu.Lock()
	s.rebuildICEConfig()
	s.mu.Unlock()

	s.eventBus.Publish(events.Event{
		Type:    "webrtc.turn_added",
		Message: fmt.Sprintf("TURN server added: %s", server.Address),
	})
	return nil
}

// DeleteTURNServer removes a TURN server from the database.
func (s *WebRTCService) DeleteTURNServer(id int64) error {
	if err := database.DB.Delete(&database.TURNServer{}, id).Error; err != nil {
		return err
	}

	s.mu.Lock()
	s.rebuildICEConfig()
	s.mu.Unlock()

	return nil
}

// TestTURNServer tests connectivity to a TURN server.
func (s *WebRTCService) TestTURNServer(address string) (bool, float64, error) {
	start := time.Now()
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		return false, 0, fmt.Errorf("cannot reach TURN server: %w", err)
	}
	defer conn.Close()
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	return true, latency, nil
}

// rebuildICEConfig rebuilds the ICE server list from STUN + TURN config.
func (s *WebRTCService) rebuildICEConfig() {
	iceServers := make([]ICEServer, 0)

	// Add STUN servers
	for _, stun := range s.iceConfig.STUNServers {
		iceServers = append(iceServers, ICEServer{
			URLs: []string{stun},
		})
	}

	// Add TURN servers from database
	var dbServers []database.TURNServer
	if err := database.DB.Find(&dbServers).Error; err == nil {
		for _, ts := range dbServers {
			if ts.Enabled {
				iceServers = append(iceServers, ICEServer{
					URLs:       []string{ts.Address},
					Username:   ts.Username,
					Credential: ts.Password,
				})
			}
		}
	}

	s.iceConfig.ICEServers = iceServers
}

// ─── Signaling ───────────────────────────────────────────────────────

// GetSignalingChannel returns the signaling message channel.
func (s *WebRTCService) GetSignalingChannel() chan SignalingMessage {
	return s.signaling
}

// SendSignalingMessage sends a signaling message to the channel.
func (s *WebRTCService) SendSignalingMessage(msg SignalingMessage) {
	msg.Timestamp = time.Now().UnixMilli()
	select {
	case s.signaling <- msg:
	default:
		log.Printf("[WebRTC] signaling channel full, dropping message: %s", msg.Type)
	}
}

// processSignaling processes incoming signaling messages.
func (s *WebRTCService) processSignaling() {
	for {
		select {
		case msg := <-s.signaling:
			s.handleSignalingMessage(msg)
		case <-s.stopCh:
			return
		}
	}
}

func (s *WebRTCService) handleSignalingMessage(msg SignalingMessage) {
	switch msg.Type {
	case "offer":
		log.Printf("[WebRTC] Offer from %s to %s", msg.FromID, msg.ToID)
		// Broadcast to WebSocket hub (handled by handler)
	case "answer":
		log.Printf("[WebRTC] Answer from %s to %s", msg.FromID, msg.ToID)
	case "ice_candidate":
		log.Printf("[WebRTC] ICE candidate from %s", msg.FromID)
	case "join":
		s.handlePeerJoin(msg)
	case "leave":
		s.handlePeerLeave(msg)
	case "hangup":
		s.handlePeerDisconnect(msg)
	}
}

func (s *WebRTCService) handlePeerJoin(msg SignalingMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer := &WebRTCPeer{
		ID:        msg.FromID,
		Name:      msg.FromID,
		Protocol:  "datachannel",
		Mode:      s.meshConfig.Role,
		Status:    WebRTCPeerStatusConnecting,
		LastSeen:  time.Now().UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
	}
	s.peers[msg.FromID] = peer

	s.eventBus.Publish(events.Event{
		Type:    "webrtc.peer_joined",
		Message: fmt.Sprintf("WebRTC peer joined: %s", msg.FromID),
	})
}

func (s *WebRTCService) handlePeerLeave(msg SignalingMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if peer, ok := s.peers[msg.FromID]; ok {
		peer.Status = WebRTCPeerStatusDisconnected
		peer.LastSeen = time.Now().UnixMilli()
	}

	s.eventBus.Publish(events.Event{
		Type:    "webrtc.peer_left",
		Message: fmt.Sprintf("WebRTC peer left: %s", msg.FromID),
	})
}

func (s *WebRTCService) handlePeerDisconnect(msg SignalingMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if peer, ok := s.peers[msg.FromID]; ok {
		peer.Status = WebRTCPeerStatusDisconnected
		peer.LastSeen = time.Now().UnixMilli()
	}
}

// ─── Peer Management ────────────────────────────────────────────────

// ListPeers returns all active WebRTC peers.
func (s *WebRTCService) ListPeers() []*WebRTCPeer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peers := make([]*WebRTCPeer, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, p)
	}
	return peers
}

// GetPeer returns a specific WebRTC peer by ID.
func (s *WebRTCService) GetPeer(id string) *WebRTCPeer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.peers[id]
}

// GetPeerStats returns statistics about peer connections.
func (s *WebRTCService) GetPeerStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.peers)
	connected := 0
	connecting := 0
	disconnected := 0

	for _, p := range s.peers {
		switch p.Status {
		case WebRTCPeerStatusConnected:
			connected++
		case WebRTCPeerStatusConnecting:
			connecting++
		default:
			disconnected++
		}
	}

	return map[string]interface{}{
		"total":        total,
		"connected":    connected,
		"connecting":   connecting,
		"disconnected": disconnected,
		"turn_servers": len(s.turnServers),
		"mesh_enabled": s.meshConfig.Enabled,
	}
}

// DisconnectPeer disconnects a WebRTC peer.
func (s *WebRTCService) DisconnectPeer(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if peer, ok := s.peers[id]; ok {
		peer.Status = WebRTCPeerStatusDisconnected
		peer.LastSeen = time.Now().UnixMilli()
		return nil
	}
	return fmt.Errorf("peer not found: %s", id)
}

// ─── P2P Mesh ───────────────────────────────────────────────────────

// GetMeshConfig returns the P2P mesh configuration.
func (s *WebRTCService) GetMeshConfig() *P2PMeshConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := *s.meshConfig
	return &cfg
}

// UpdateMeshConfig updates the P2P mesh configuration.
func (s *WebRTCService) UpdateMeshConfig(cfg *P2PMeshConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.meshConfig = cfg

	// Save to database
	var dbCfg database.P2PMeshConfig
	if err := database.DB.First(&dbCfg).Error; err != nil {
		// Create new
		dbCfg = database.P2PMeshConfig{
			Enabled:       cfg.Enabled,
			MeshName:      cfg.MeshName,
			Role:          cfg.Role,
			ListenPort:    cfg.ListenPort,
			MaxPeers:      cfg.MaxPeers,
			AutoReconnect: cfg.AutoReconnect,
			HeartbeatSec:  cfg.HeartbeatSec,
		}
		dbCfg.CreatedAt = time.Now().UnixMilli()
		dbCfg.UpdatedAt = dbCfg.CreatedAt
		return database.DB.Create(&dbCfg).Error
	}

	dbCfg.Enabled = cfg.Enabled
	dbCfg.MeshName = cfg.MeshName
	dbCfg.Role = cfg.Role
	dbCfg.ListenPort = cfg.ListenPort
	dbCfg.MaxPeers = cfg.MaxPeers
	dbCfg.AutoReconnect = cfg.AutoReconnect
	dbCfg.HeartbeatSec = cfg.HeartbeatSec
	dbCfg.UpdatedAt = time.Now().UnixMilli()
	return database.DB.Save(&dbCfg).Error
}

// heartbeatLoop sends periodic heartbeats to connected peers.
func (s *WebRTCService) heartbeatLoop() {
	ticker := time.NewTicker(time.Duration(s.meshConfig.HeartbeatSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			for id, peer := range s.peers {
				if peer.Status == WebRTCPeerStatusConnected {
					peer.LastSeen = time.Now().UnixMilli()
					// Send heartbeat via signaling
					msg := SignalingMessage{
						Type:      "heartbeat",
						FromID:    "server",
						ToID:      id,
						Timestamp: time.Now().UnixMilli(),
					}
					select {
					case s.signaling <- msg:
					default:
					}
				}
			}
			s.mu.RUnlock()
		case <-s.stopCh:
			return
		}
	}
}

// ─── Discovery ───────────────────────────────────────────────────────

// DiscoverPeers finds other WebRTC peers via the configured discovery method.
func (s *WebRTCService) DiscoverPeers() ([]string, error) {
	s.mu.RLock()
	method := s.meshConfig.Discovery
	s.mu.RUnlock()

	switch method {
	case "manual":
		return s.discoverPeersManual()
	case "dns":
		return s.discoverPeersDNS()
	default:
		return s.discoverPeersSignaling()
	}
}

func (s *WebRTCService) discoverPeersManual() ([]string, error) {
	s.mu.RLock()
	peers := make([]string, 0, len(s.peers))
	for id, p := range s.peers {
		if p.Status == WebRTCPeerStatusConnected {
			peers = append(peers, id)
		}
	}
	s.mu.RUnlock()
	return peers, nil
}

func (s *WebRTCService) discoverPeersDNS() ([]string, error) {
	// DNS-based discovery using SRV records
	_, err := net.LookupHost(fmt.Sprintf("_vortexuipro-mesh._udp.%s", s.meshConfig.MeshName))
	if err != nil {
		return nil, fmt.Errorf("DNS discovery failed: %w", err)
	}
	return []string{}, nil
}

func (s *WebRTCService) discoverPeersSignaling() ([]string, error) {
	s.mu.RLock()
	peers := make([]string, 0, len(s.peers))
	for id := range s.peers {
		peers = append(peers, id)
	}
	s.mu.RUnlock()
	return peers, nil
}

// ─── NAT Type Detection ─────────────────────────────────────────────

// NATType represents the detected NAT type.
type NATType struct {
	Type        string `json:"type"`        // full_cone, restricted_cone, port_restricted, symmetric, unknown
	PublicIP    string `json:"public_ip"`
	PublicPort  int    `json:"public_port"`
	BehindNAT   bool   `json:"behind_nat"`
	Description string `json:"description"`
}

// DetectNATType attempts to detect the NAT type.
func (s *WebRTCService) DetectNATType() (*NATType, error) {
	// Simple NAT detection by connecting to STUN server
	conn, err := net.DialTimeout("udp", "8.8.8.8:53", 5*time.Second)
	if err != nil {
		return &NATType{
			Type:      "unknown",
			BehindNAT: true,
			Description: "Cannot determine NAT type",
		}, nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	remoteAddr := conn.RemoteAddr().(*net.UDPAddr)

	natType := "unknown"
	if localAddr.IP.IsPrivate() {
		natType = "symmetric"
	} else {
		natType = "full_cone"
	}

	return &NATType{
		Type:        natType,
		PublicIP:    localAddr.IP.String(),
		PublicPort:  localAddr.Port,
		BehindNAT:   localAddr.IP.IsPrivate(),
		Description: fmt.Sprintf("NAT type: %s (local: %s, remote: %s)", natType, localAddr.String(), remoteAddr.String()),
	}, nil
}

// ─── Utility ─────────────────────────────────────────────────────────

// GeneratePeerID generates a unique peer ID.
func GeneratePeerID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("peer-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("peer-%x", b)
}
