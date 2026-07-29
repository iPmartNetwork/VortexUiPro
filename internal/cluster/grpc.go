package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── gRPC Mesh Service ───────────────────────────────────────────────

// GRPCMessage is the runtime equivalent of the proto message.
type GRPCMessage struct {
	Type       string `json:"type" protobuf:"bytes,1,opt,name=type"`
	SenderID   int64  `json:"sender_id" protobuf:"varint,2,opt,name=sender_id"`
	SenderName string `json:"sender_name" protobuf:"bytes,3,opt,name=sender_name"`
	Term       int64  `json:"term" protobuf:"varint,4,opt,name=term"`
	Timestamp  int64  `json:"timestamp" protobuf:"varint,5,opt,name=timestamp"`
	Payload    []byte `json:"payload,omitempty" protobuf:"bytes,6,opt,name=payload"`
}

// Reset implements proto.Message.
func (m *GRPCMessage) Reset() { *m = GRPCMessage{} }

// String implements proto.Message.
func (m *GRPCMessage) String() string { data, _ := json.Marshal(m); return string(data) }

// ProtoMessage implements proto.Message.
func (m *GRPCMessage) ProtoMessage() {}

// Marshal implements proto.Marshaler.
func (m *GRPCMessage) Marshal() ([]byte, error) { return json.Marshal(m) }

// Unmarshal implements proto.Unmarshaler.
func (m *GRPCMessage) Unmarshal(data []byte) error { return json.Unmarshal(data, m) }

// GRPCStream is the interface for bidirectional streaming.
type GRPCStream interface {
	Send(msg *GRPCMessage) error
	Recv() (*GRPCMessage, error)
	CloseSend() error
}

// GRPCService wraps the gRPC server and client functionality.
type GRPCService struct {
	server   *grpc.Server
	clients  map[int64]*gRPCClientConn
	cfg      Config
	pki      *PKIManager
	onMsg    func(msg MeshMessage) (*MeshMessage, error)
	mu       sync.RWMutex
	stopCh   chan struct{}
}

type gRPCClientConn struct {
	conn   *grpc.ClientConn
	addr   string
	stream GRPCStream
}

// NewGRPCService creates a new gRPC mesh service.
func NewGRPCService(cfg Config, pki *PKIManager) *GRPCService {
	return &GRPCService{
		cfg:     cfg,
		pki:     pki,
		clients: make(map[int64]*gRPCClientConn),
		stopCh:  make(chan struct{}),
	}
}

// OnMessage registers the message handler callback.
func (g *GRPCService) OnMessage(fn func(msg MeshMessage) (*MeshMessage, error)) {
	g.onMsg = fn
}

// Start begins the gRPC server with optional mTLS.
func (g *GRPCService) Start() error {
	lis, err := net.Listen("tcp", g.cfg.Addr)
	if err != nil {
		return fmt.Errorf("gRPC listen %s: %w", g.cfg.Addr, err)
	}

	var opts []grpc.ServerOption
	if g.pki != nil && g.cfg.TLSEnabled {
		tlsCfg, err := TLSConfig(g.pki, true)
		if err == nil {
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
			log.Println("gRPC: mTLS enabled")
		}
	}

	g.server = grpc.NewServer(opts...)

	sd := &grpc.ServiceDesc{
		ServiceName: "vortexcluster.ClusterMesh",
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Handshake",
				Handler:    grpcHandshakeHandler,
			},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "HeartbeatStream",
				Handler:       grpcHeartbeatStreamHandler,
				ServerStreams: true,
				ClientStreams: true,
			},
			{
				StreamName:    "SyncStream",
				Handler:       grpcSyncStreamHandler,
				ServerStreams: true,
				ClientStreams: true,
			},
		},
		Metadata: "cluster.proto",
	}
	g.server.RegisterService(sd, g)

	go g.server.Serve(lis)
	log.Printf("gRPC mesh server listening on %s", g.cfg.Addr)
	return nil
}

// Stop gracefully stops the gRPC server.
func (g *GRPCService) Stop() {
	if g.server != nil {
		g.server.GracefulStop()
	}
	close(g.stopCh)
}

// HandleStream is called for each incoming bidirectional stream.
func (g *GRPCService) HandleStream(stream GRPCStream) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return
		}

		meshMsg := MeshMessage{
			Type:       MsgType(msg.Type),
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
			Term:       msg.Term,
			Timestamp:  msg.Timestamp,
			Payload:    msg.Payload,
		}

		if g.onMsg != nil {
			resp, err := g.onMsg(meshMsg)
			if err == nil && resp != nil {
				grpcResp := &GRPCMessage{
					Type:       string(resp.Type),
					SenderID:   resp.SenderID,
					SenderName: resp.SenderName,
					Term:       resp.Term,
					Timestamp:  resp.Timestamp,
					Payload:    resp.Payload,
				}
				stream.Send(grpcResp)
			}
		}
	}
}

// ConnectToPeer establishes a gRPC connection to a peer.
func (g *GRPCService) ConnectToPeer(peerID int64, addr string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.clients[peerID]; ok {
		return nil
	}

	var opts []grpc.DialOption
	if g.pki != nil && g.cfg.TLSEnabled {
		tlsCfg, err := TLSDialConfig(g.pki, "cluster.local")
		if err == nil {
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return fmt.Errorf("gRPC dial %s: %w", addr, err)
	}

	stream := &simpleClientStream{conn: conn}
	if err := stream.Open(); err != nil {
		conn.Close()
		return fmt.Errorf("open stream: %w", err)
	}

	g.clients[peerID] = &gRPCClientConn{
		conn:   conn,
		addr:   addr,
		stream: stream,
	}

	log.Printf("gRPC: Connected to peer %d at %s", peerID, addr)
	return nil
}

// SendToPeer sends a message via gRPC stream.
func (g *GRPCService) SendToPeer(peerID int64, msg MeshMessage) error {
	g.mu.RLock()
	conn, ok := g.clients[peerID]
	g.mu.RUnlock()

	if !ok || conn.stream == nil {
		return fmt.Errorf("no gRPC connection to peer %d", peerID)
	}

	grpcMsg := &GRPCMessage{
		Type:       string(msg.Type),
		SenderID:   msg.SenderID,
		SenderName: msg.SenderName,
		Term:       msg.Term,
		Timestamp:  msg.Timestamp,
		Payload:    msg.Payload,
	}

	return conn.stream.Send(grpcMsg)
}

// ─── Simple Client Stream ────────────────────────────────────────────

type simpleClientStream struct {
	conn   *grpc.ClientConn
	stream grpc.ClientStream
	closed bool
	mu     sync.Mutex
}

func (s *simpleClientStream) Open() error {
	desc := &grpc.StreamDesc{
		StreamName:    "HeartbeatStream",
		ServerStreams: true,
		ClientStreams: true,
	}
	stream, err := s.conn.NewStream(context.Background(), desc, "/vortexcluster.ClusterMesh/HeartbeatStream")
	if err != nil {
		return err
	}
	s.stream = stream
	return nil
}

func (s *simpleClientStream) Send(msg *GRPCMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.stream == nil {
		return fmt.Errorf("stream closed")
	}
	return s.stream.SendMsg(msg)
}

func (s *simpleClientStream) Recv() (*GRPCMessage, error) {
	if s.stream == nil {
		return nil, fmt.Errorf("stream not open")
	}
	var msg GRPCMessage
	if err := s.stream.RecvMsg(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *simpleClientStream) CloseSend() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if s.stream != nil {
		return s.stream.CloseSend()
	}
	return nil
}

// ─── Server Stream Handlers ──────────────────────────────────────────

func grpcHandshakeHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return &map[string]interface{}{"accepted": true}, nil
}

func grpcHeartbeatStreamHandler(srv interface{}, stream grpc.ServerStream) error {
	svc := srv.(*GRPCService)
	adapter := &serverStreamAdapter{ServerStream: stream}
	svc.HandleStream(adapter)
	return nil
}

func grpcSyncStreamHandler(srv interface{}, stream grpc.ServerStream) error {
	svc := srv.(*GRPCService)
	adapter := &serverStreamAdapter{ServerStream: stream}
	svc.HandleStream(adapter)
	return nil
}

// serverStreamAdapter adapts grpc.ServerStream to GRPCStream interface.
type serverStreamAdapter struct {
	grpc.ServerStream
}

func (a *serverStreamAdapter) Send(msg *GRPCMessage) error {
	return a.ServerStream.SendMsg(msg)
}

func (a *serverStreamAdapter) Recv() (*GRPCMessage, error) {
	var msg GRPCMessage
	if err := a.ServerStream.RecvMsg(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (a *serverStreamAdapter) CloseSend() error {
	return nil
}
