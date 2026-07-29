package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// MessageType identifies the kind of WebSocket message.
type MessageType string

const (
	MsgTraffic    MessageType = "traffic"
	MsgInbounds   MessageType = "inbounds"
	MsgNodes      MessageType = "nodes"
	MsgClients    MessageType = "clients"
	MsgNotify     MessageType = "notification"
	MsgCoreState  MessageType = "core_state"
	MsgOnline     MessageType = "online_count"
	MsgSystem      MessageType = "system"
	MsgInvalidate  MessageType = "invalidate"
	MsgTopology    MessageType = "topology"
	MsgCluster     MessageType = "cluster"
)

const (
	// WebSocket message topics for cluster topology
	WSTopoNodeJoined    = "node_joined"
	WSTopoNodeLeft      = "node_left"
	WSTopoNodeUpdated   = "node_updated"
	WSTopoLeaderChanged = "leader_changed"
	WSTopoElection      = "election"
)

// Message is the wire format for WebSocket messages.
type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
	Time    int64       `json:"time"`
}

// Client represents a single WebSocket connection.
type Client struct {
	ID   string
	Send chan []byte
	mu   sync.Mutex
}

// Hub fan-outs messages to all connected clients.
type Hub struct {
	clients  map[*Client]bool
	broadcast chan []byte
	register   chan *Client
	unregister chan *Client
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run starts the hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			n := len(h.clients)
			h.mu.Unlock()
			fmt.Printf("ws client connected (%d total)\n", n)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			n := len(h.clients)
			h.mu.Unlock()
			fmt.Printf("ws client disconnected (%d total)\n", n)
		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop gracefully shuts down the hub.
func (h *Hub) Stop() {
	h.cancel()
	h.mu.Lock()
	for client := range h.clients {
		close(client.Send)
	}
	h.clients = make(map[*Client]bool)
	h.mu.Unlock()
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg Message) {
	data, _ := json.Marshal(msg)
	select {
	case h.broadcast <- data:
	default:
	}
}

// BroadcastTraffic sends traffic stats update.
func (h *Hub) BroadcastTraffic(up, down int64, online int) {
	h.Broadcast(Message{
		Type: MsgTraffic,
		Payload: map[string]int64{
			"up": up, "down": down, "online": int64(online),
		},
		Time: time.Now().UnixMilli(),
	})
}

// BroadcastMetrics sends a metrics snapshot to all clients.
func (h *Hub) BroadcastMetrics(snapshot map[string]any) {
	if snapshot == nil {
		return
	}
	h.Broadcast(Message{
		Type:    MsgSystem,
		Payload: snapshot,
		Time:    time.Now().UnixMilli(),
	})
}

// BroadcastNotification sends a notification to all clients.
func (h *Hub) BroadcastNotification(level, title, message string) {
	h.Broadcast(Message{
		Type: MsgNotify,
		Payload: map[string]string{
			"level": level, "title": title, "message": message,
		},
		Time: time.Now().UnixMilli(),
	})
}

// HandleWebSocket handles a WebSocket upgrade request.
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ID:   fmt.Sprintf("client-%d", time.Now().UnixNano()),
		Send: make(chan []byte, 64),
	}

	h.register <- client

	// Write pump
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		defer func() {
			h.unregister <- client
			conn.Close()
		}()

		for {
			select {
			case msg, ok := <-client.Send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}

// HandleDashboard serves the WebSocket dashboard endpoint.
func (h *Hub) HandleDashboard(c *gin.Context) {
	h.HandleWebSocket(c)
}
