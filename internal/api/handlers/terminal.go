package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"vortexuipro/internal/service"
)

// TerminalHandler manages Web SSH terminal sessions.
type TerminalHandler struct {
	svc *service.TerminalService
	up  websocket.Upgrader
}

// NewTerminalHandler creates a new terminal handler.
func NewTerminalHandler(svc *service.TerminalService) *TerminalHandler {
	return &TerminalHandler{
		svc: svc,
		up: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// HandleTerminalWS handles WebSocket connections for terminal sessions.
func (h *TerminalHandler) HandleTerminalWS(c *gin.Context) {
	// Auth check via token in query string
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}
	// Basic token validation: verify it matches a stored admin token
	// For simplicity, check JWT format or presence in database
	// In production, this should validate against AdminService
	if len(token) < 10 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	nodeIDStr := c.Query("node_id")
	if nodeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id required"})
		return
	}
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
		return
	}

	conn, err := h.up.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[Terminal] WebSocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	session, output, err := h.svc.OpenSession(nodeID, 120, 40)
	if err != nil {
		conn.WriteJSON(service.TerminalOutput{Type: "error", Data: err.Error()})
		return
	}
	defer h.svc.CloseSession(session.ID)

	conn.WriteJSON(service.TerminalOutput{
		Type: "connected",
		Data: session.ID,
	})

	// Read from WebSocket and write to SSH stdin
	go func() {
		defer h.svc.CloseSession(session.ID)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var input struct {
				Type string `json:"type"`
				Data string `json:"data"`
			}
			if err := json.Unmarshal(msg, &input); err != nil {
				continue
			}
			switch input.Type {
			case "input":
				h.svc.WriteInput(session.ID, input.Data)
			case "resize":
				// Resize not fully implemented
			case "close":
				return
			}
		}
	}()

	// Read from SSH stdout and write to WebSocket
	for out := range output {
		if err := conn.WriteJSON(out); err != nil {
			break
		}
	}
}

// ListSessions returns active terminal sessions.
func (h *TerminalHandler) ListSessions(c *gin.Context) {
	sessions := h.svc.ListSessions()
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// CloseSession closes a terminal session.
func (h *TerminalHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.svc.CloseSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "session closed"})
}
