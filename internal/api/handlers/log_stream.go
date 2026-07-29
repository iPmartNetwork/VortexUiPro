package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"vortexuipro/internal/service"
)

// LogStreamHandler handles WebSocket log streaming.
type LogStreamHandler struct {
	svc *service.LogStreamService
	up  websocket.Upgrader
}

// NewLogStreamHandler creates a new log stream handler.
func NewLogStreamHandler(svc *service.LogStreamService) *LogStreamHandler {
	return &LogStreamHandler{
		svc: svc,
		up: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// HandleLogStream handles WebSocket connections for log streaming.
func (h *LogStreamHandler) HandleLogStream(c *gin.Context) {
	// Auth check via token in query string
	token := c.Query("token")
	if token == "" || len(token) < 10 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "valid token required"})
		return
	}

	levelStr := c.DefaultQuery("level", "info")
	source := c.DefaultQuery("source", "")
	filter := c.DefaultQuery("filter", "")

	minLevel := service.LogLevelInfo
	switch levelStr {
	case "debug":
		minLevel = service.LogLevelDebug
	case "warn":
		minLevel = service.LogLevelWarn
	case "error":
		minLevel = service.LogLevelError
	}

	conn, err := h.up.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[LogStream] WebSocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	subID := "ws_" + time.Now().Format("150405.000")
	sub := h.svc.Subscribe(subID, minLevel, filter, source)
	defer h.svc.Unsubscribe(subID)

	// Send initial recent logs
	recent := h.svc.GetRecentLogs(100)
	for _, entry := range recent {
		data, _ := json.Marshal(entry)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}

	// Stream new logs
	for {
		select {
		case <-sub.WaitForLogs():
			logs := sub.DrainLogs()
			for _, entry := range logs {
				data, _ := json.Marshal(entry)
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			}
		case <-time.After(30 * time.Second):
			// Send keepalive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetRecentLogs returns recent log entries via REST.
func (h *LogStreamHandler) GetRecentLogs(c *gin.Context) {
	count := 200
	logs := h.svc.GetRecentLogs(count)
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": len(logs)})
}
