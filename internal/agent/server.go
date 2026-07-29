package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── Proto-style message types (lightweight, no protoc needed) ──────

// Heartbeat is a node status update sent to the panel.
type Heartbeat struct {
	NodeID      int64   `json:"node_id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	Status      string  `json:"status"`
	CPULoad     float64 `json:"cpu_load"`
	MemoryUsed  float64 `json:"memory_used"`
	Uplink      int64   `json:"uplink"`
	Downlink    int64   `json:"downlink"`
	TrafficUp   int64   `json:"traffic_up"`
	TrafficDown int64   `json:"traffic_down"`
	LoadAvg     float64 `json:"load_avg"`
	Uptime      int64   `json:"uptime"`
	CoreVersion string  `json:"core_version"`
	Timestamp   int64   `json:"timestamp"`
}

// ConfigSyncRequest is sent from panel to node to update its configuration.
type ConfigSyncRequest struct {
	NodeID  int64  `json:"node_id"`
	Config  []byte `json:"config"`
	Version string `json:"version"`
}

// ConfigSyncResponse is the node's reply after applying config.
type ConfigSyncResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Version string `json:"version"`
}

// ─── Node Agent Server ──────────────────────────────────────────────

// NodeAgentServer handles incoming gRPC connections from remote nodes.
type NodeAgentServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	addr       string

	mu        sync.RWMutex
	nodes     map[int64]*NodeState
	onHeartbeat func(hb Heartbeat)
}

// NodeState tracks the last known state of a connected node.
type NodeState struct {
	Heartbeat   Heartbeat
	LastSeen    time.Time
	Connected   bool
	ConfigHash  string
}

// NewNodeAgentServer creates a new gRPC node agent server.
func NewNodeAgentServer(addr string) *NodeAgentServer {
	return &NodeAgentServer{
		addr:  addr,
		nodes: make(map[int64]*NodeState),
	}
}

// OnHeartbeat registers a callback for incoming heartbeats.
func (s *NodeAgentServer) OnHeartbeat(fn func(hb Heartbeat)) {
	s.onHeartbeat = fn
}

// Start launches the gRPC server.
func (s *NodeAgentServer) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("node agent listen %s: %w", s.addr, err)
	}
	s.listener = lis

	s.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	// Register handlers for the raw gRPC API
	// Using the raw codec approach for simplicity
	go s.grpcServer.Serve(lis)

	log.Printf("Node Agent gRPC server listening on %s", s.addr)
	return nil
}

// Stop gracefully stops the gRPC server.
func (s *NodeAgentServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
}

// RegisterHeartbeat stores a heartbeat from a node.
func (s *NodeAgentServer) RegisterHeartbeat(hb Heartbeat) {
	s.mu.Lock()
	state, exists := s.nodes[hb.NodeID]
	if !exists {
		state = &NodeState{Connected: true}
		s.nodes[hb.NodeID] = state
	}
	state.Heartbeat = hb
	state.LastSeen = time.Now()
	state.Connected = true
	s.mu.Unlock()

	if s.onHeartbeat != nil {
		s.onHeartbeat(hb)
	}
}

// GetNodeState returns the current state of a node.
func (s *NodeAgentServer) GetNodeState(nodeID int64) (*NodeState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.nodes[nodeID]
	return state, ok
}

// ListNodes returns all known nodes and their states.
func (s *NodeAgentServer) ListNodes() []NodeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]NodeState, 0, len(s.nodes))
	for _, state := range s.nodes {
		result = append(result, *state)
	}
	return result
}

// CheckStaleNodes marks nodes as disconnected if they haven't sent a heartbeat.
func (s *NodeAgentServer) CheckStaleNodes(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, state := range s.nodes {
		if state.Connected && now.Sub(state.LastSeen) > timeout {
			state.Connected = false
			log.Printf("Node %d (%s) marked as disconnected", id, state.Heartbeat.Name)
		}
	}
}

// ─── Node Agent Client ──────────────────────────────────────────────

// NodeAgentClient is used by nodes to connect to the panel.
type NodeAgentClient struct {
	conn *grpc.ClientConn
	addr string
}

// NewNodeAgentClient creates a new client connection to the panel.
func NewNodeAgentClient(addr string) *NodeAgentClient {
	return &NodeAgentClient{addr: addr}
}

// Connect establishes the gRPC connection.
func (c *NodeAgentClient) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("node agent connect %s: %w", c.addr, err)
	}
	c.conn = conn
	return nil
}

// Close closes the connection.
func (c *NodeAgentClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
