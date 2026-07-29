package cluster

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// ─── Mesh Message Types ──────────────────────────────────────────────

// MsgType defines the type of cluster mesh message.
type MsgType string

const (
	MsgHeartbeat   MsgType = "heartbeat"
	MsgVoteRequest MsgType = "vote_request"
	MsgVoteResponse MsgType = "vote_response"
	MsgLeaderAnnounce MsgType = "leader_announce"
	MsgSyncData    MsgType = "sync_data"
	MsgSyncAck     MsgType = "sync_ack"
	MsgSyncRequest MsgType = "sync_request"
)

// MeshMessage is the envelope for all cluster communication.
type MeshMessage struct {
	Type       MsgType         `json:"type"`
	SenderID   int64           `json:"sender_id"`
	SenderName string          `json:"sender_name"`
	Term       int64           `json:"term"`
	Timestamp  int64           `json:"timestamp"`
	RemoteAddr string          `json:"-"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

// HeartbeatPayload is sent periodically between nodes.
type HeartbeatPayload struct {
	NodeID    int64   `json:"node_id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"` // mesh address (ip:port) for auto-discovery
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	CPULoad   float64 `json:"cpu_load"`
	MemUsed   float64 `json:"mem_used"`
	Term      int64   `json:"term"`
	Version   string  `json:"version"`
	Region    string  `json:"region"`
	UserCount int     `json:"user_count"`
}

// VotePayload is sent during leader election.
type VotePayload struct {
	CandidateID   int64  `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	Priority      int    `json:"priority"`
	Term          int64  `json:"term"`
}

// VoteResultPayload is the response to a vote request.
type VoteResultPayload struct {
	VoterID  int64 `json:"voter_id"`
	Term     int64 `json:"term"`
	Granted  bool  `json:"granted"`
}

// LeaderAnnouncePayload announces a new leader.
type LeaderAnnouncePayload struct {
	LeaderID   int64  `json:"leader_id"`
	LeaderName string `json:"leader_name"`
	Term       int64  `json:"term"`
	Address    string `json:"address"`
}

// ─── Peer HTTP Server & Client ───────────────────────────────────────

// PeerServer handles incoming cluster mesh messages via HTTP.
type PeerServer struct {
	addr       string
	server     *http.Server
	mux        *http.ServeMux
	handler    func(msg MeshMessage) (*MeshMessage, error)
	mu         sync.RWMutex
	started    bool
	tlsConfig  interface{} // *tls.Config or nil
}

// NewPeerServer creates a new cluster mesh HTTP server.
func NewPeerServer(addr string) *PeerServer {
	mux := http.NewServeMux()
	ps := &PeerServer{
		addr: addr,
		mux:  mux,
	}
	mux.HandleFunc("/cluster/mesh", ps.handleMesh)
	mux.HandleFunc("/cluster/health", ps.handleHealth)
	return ps
}

// NewPeerServerTLS creates a new cluster mesh HTTP server with mTLS.
func NewPeerServerTLS(addr string, tlsCfg interface{}) *PeerServer {
	ps := NewPeerServer(addr)
	ps.tlsConfig = tlsCfg
	return ps
}

// OnMessage registers the message handler callback.
func (ps *PeerServer) OnMessage(fn func(msg MeshMessage) (*MeshMessage, error)) {
	ps.mu.Lock()
	ps.handler = fn
	ps.mu.Unlock()
}

// Start begins listening for peer connections.
func (ps *PeerServer) Start() error {
	ps.mu.Lock()
	ps.server = &http.Server{
		Addr:         ps.addr,
		Handler:      ps.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	ps.started = true
	ps.mu.Unlock()

	listener, err := net.Listen("tcp", ps.addr)
	if err != nil {
		return fmt.Errorf("cluster peer listen %s: %w", ps.addr, err)
	}

	// Wrap listener with TLS if mTLS is configured
	if ps.tlsConfig != nil {
		listener = tls.NewListener(listener, ps.tlsConfig.(*tls.Config))
		log.Printf("Cluster peer server listening on %s (mTLS enabled)", ps.addr)
	} else {
		log.Printf("Cluster peer server listening on %s (plaintext)", ps.addr)
	}

	go ps.server.Serve(listener)
	return nil
}

// Stop gracefully shuts down the peer server.
func (ps *PeerServer) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.server != nil {
		ctx, cancel := contextWithTimeout(5 * time.Second)
		defer cancel()
		ps.server.Shutdown(ctx)
	}
	ps.started = false
}

func (ps *PeerServer) handleMesh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var msg MeshMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Extract remote address from request and embed in message
	msg.RemoteAddr = r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		msg.RemoteAddr = forwarded
	}

	ps.mu.RLock()
	fn := ps.handler
	ps.mu.RUnlock()

	var resp *MeshMessage
	if fn != nil {
		resp, err = fn(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if resp != nil {
		json.NewEncoder(w).Encode(resp)
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (ps *PeerServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// ─── Peer Client ─────────────────────────────────────────────────────

// PeerClient sends mesh messages to a specific peer.
type PeerClient struct {
	baseURL string
	client  *http.Client
}

// NewPeerClient creates a new peer client targeting a specific address.
func NewPeerClient(addr string) *PeerClient {
	return &PeerClient{
		baseURL: fmt.Sprintf("http://%s/cluster", addr),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}
}

// NewPeerClientTLS creates a new peer client with mTLS.
func NewPeerClientTLS(addr string, tlsCfg *tls.Config) *PeerClient {
	return &PeerClient{
		baseURL: fmt.Sprintf("https://%s/cluster", addr),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     tlsCfg,
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}
}

// Send sends a mesh message to the peer and returns the response.
func (pc *PeerClient) Send(msg MeshMessage) (*MeshMessage, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal mesh msg: %w", err)
	}

	resp, err := pc.client.Post(pc.baseURL+"/mesh", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("send mesh msg: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read mesh response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("peer error %d: %s", resp.StatusCode, string(body))
	}

	var response MeshMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil // no response body is ok
	}
	return &response, nil
}

// Health checks if the peer is alive.
func (pc *PeerClient) Health() error {
	resp, err := pc.client.Get(pc.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
