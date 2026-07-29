package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// ─── WebRTC Handler ──────────────────────────────────────────────────

type WebRTCHandler struct {
	svc *service.WebRTCService
}

// NewWebRTCHandler creates a new WebRTC handler.
func NewWebRTCHandler(svc *service.WebRTCService) *WebRTCHandler {
	return &WebRTCHandler{svc: svc}
}

// ─── ICE / STUN / TURN ──────────────────────────────────────────────

// GetICEConfig returns the ICE configuration (STUN + TURN servers).
func (h *WebRTCHandler) GetICEConfig(c *gin.Context) {
	config := h.svc.GetICEConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// ListTURNServers returns all configured TURN servers.
func (h *WebRTCHandler) ListTURNServers(c *gin.Context) {
	servers, err := h.svc.GetTURNServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if servers == nil {
		servers = []database.TURNServer{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    servers,
	})
}

// CreateTURNServer adds a new TURN server.
func (h *WebRTCHandler) CreateTURNServer(c *gin.Context) {
	var req struct {
		Address  string `json:"address" binding:"required"`
		Username string `json:"username"`
		Password string `json:"password"`
		Realm    string `json:"realm"`
		Protocol string `json:"protocol"`
		Region   string `json:"region"`
		Bandwidth int   `json:"bandwidth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	protocol := req.Protocol
	if protocol == "" {
		protocol = "udp"
	}
	bandwidth := req.Bandwidth
	if bandwidth <= 0 {
		bandwidth = 100
	}

	server := &database.TURNServer{
		Address:   req.Address,
		Username:  req.Username,
		Password:  req.Password,
		Realm:     req.Realm,
		Protocol:  protocol,
		Region:    req.Region,
		Bandwidth: bandwidth,
		Status:    "offline",
		Enabled:   true,
	}

	if err := h.svc.AddTURNServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "TURN server added",
		"data":    server,
	})
}

// DeleteTURNServer removes a TURN server.
func (h *WebRTCHandler) DeleteTURNServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteTURNServer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "TURN server deleted",
	})
}

// TestTURNServer tests connectivity to a TURN server.
func (h *WebRTCHandler) TestTURNServer(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reachable, latency, err := h.svc.TestTURNServer(req.Address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"reachable": false,
			"error":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"reachable": reachable,
		"latency_ms": latency,
	})
}

// ─── P2P Mesh ───────────────────────────────────────────────────────

// GetMeshConfig returns the P2P mesh configuration.
func (h *WebRTCHandler) GetMeshConfig(c *gin.Context) {
	cfg := h.svc.GetMeshConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cfg,
	})
}

// UpdateMeshConfig updates the P2P mesh configuration.
func (h *WebRTCHandler) UpdateMeshConfig(c *gin.Context) {
	var req struct {
		Enabled       bool   `json:"enabled"`
		MeshName      string `json:"mesh_name"`
		Role          string `json:"role"`
		ListenPort    int    `json:"listen_port"`
		MaxPeers      int    `json:"max_peers"`
		AutoReconnect bool   `json:"auto_reconnect"`
		Discovery     string `json:"discovery"`
		Encryption    bool   `json:"encryption"`
		HeartbeatSec  int    `json:"heartbeat_sec"`
		DataChannel   bool   `json:"data_channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := &service.P2PMeshConfig{
		Enabled:       req.Enabled,
		MeshName:      req.MeshName,
		Role:          req.Role,
		ListenPort:    req.ListenPort,
		MaxPeers:      req.MaxPeers,
		AutoReconnect: req.AutoReconnect,
		Discovery:     req.Discovery,
		Encryption:    req.Encryption,
		HeartbeatSec:  req.HeartbeatSec,
		DataChannel:   req.DataChannel,
	}

	if err := h.svc.UpdateMeshConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mesh config updated",
		"data":    cfg,
	})
}

// ─── Peers ──────────────────────────────────────────────────────────

// ListPeers returns all WebRTC peers.
func (h *WebRTCHandler) ListPeers(c *gin.Context) {
	peers := h.svc.ListPeers()
	if peers == nil {
		peers = []*service.WebRTCPeer{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    peers,
	})
}

// GetPeer returns a specific WebRTC peer.
func (h *WebRTCHandler) GetPeer(c *gin.Context) {
	id := c.Param("id")
	peer := h.svc.GetPeer(id)
	if peer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "peer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    peer,
	})
}

// DisconnectPeer disconnects a WebRTC peer.
func (h *WebRTCHandler) DisconnectPeer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DisconnectPeer(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Peer disconnected",
	})
}

// GetPeerStats returns WebRTC peer statistics.
func (h *WebRTCHandler) GetPeerStats(c *gin.Context) {
	stats := h.svc.GetPeerStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ─── Signaling ──────────────────────────────────────────────────────

// HandleSignalingWS handles WebRTC signaling over WebSocket.
// Uses the WebSocket hub from the router to manage signaling connections.
func (h *WebRTCHandler) HandleSignalingWS(c *gin.Context) {
	h.svc.SendSignalingMessage(service.SignalingMessage{
		Type:      "ws_connected",
		FromID:    "client",
		Timestamp: time.Now().UnixMilli(),
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Signaling WebSocket upgraded",
	})
}

// PostSignalingMessage accepts a signaling message via REST POST.
func (h *WebRTCHandler) PostSignalingMessage(c *gin.Context) {
	var msg service.SignalingMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg.Timestamp = time.Now().UnixMilli()
	h.svc.SendSignalingMessage(msg)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Signaling message sent",
	})
}

// ─── Discovery & NAT ───────────────────────────────────────────────

// DiscoverPeers discovers WebRTC peers via the configured method.
func (h *WebRTCHandler) DiscoverPeers(c *gin.Context) {
	peers, err := h.svc.DiscoverPeers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    peers,
	})
}

// DetectNATType detects the NAT type of the current server.
func (h *WebRTCHandler) DetectNATType(c *gin.Context) {
	natType, err := h.svc.DetectNATType()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": natType,
	})
}

// ─── Serialization ──────────────────────────────────────────────────

func (h *WebRTCHandler) SendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
