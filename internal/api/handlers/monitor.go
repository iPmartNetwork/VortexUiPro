package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/rbac"
	"vortexuipro/internal/service"
)

// MonitorHandler handles online users, traffic, and system monitoring.
type MonitorHandler struct {
	onlineTracker *service.OnlineTracker
	activitySvc   *service.ClientActivityService
	xraySvc       *service.XrayService
}

// NewMonitorHandler creates a new monitor handler.
func NewMonitorHandler(ot *service.OnlineTracker, as *service.ClientActivityService, xs *service.XrayService) *MonitorHandler {
	return &MonitorHandler{
		onlineTracker: ot,
		activitySvc:   as,
		xraySvc:       xs,
	}
}

// GetOnlineUsers returns the list of currently online users.
func (h *MonitorHandler) GetOnlineUsers(c *gin.Context) {
	users := h.onlineTracker.GetOnline()
	if users == nil {
		users = []service.OnlineUser{}
	}
	byInbound := h.onlineTracker.GetOnlineByInbound()
	c.JSON(http.StatusOK, gin.H{
		"online":      users,
		"total":       len(users),
		"by_inbound":  byInbound,
	})
}

// GetOnlineCount returns just the count.
func (h *MonitorHandler) GetOnlineCount(c *gin.Context) {
	count := h.onlineTracker.GetOnlineCount()
	c.JSON(http.StatusOK, gin.H{"online": count})
}

// GetRecentActivity returns recent client activity.
func (h *MonitorHandler) GetRecentActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	activities := h.activitySvc.GetRecentActivity(limit)
	if activities == nil {
		activities = []service.ActivityRecord{}
	}
	c.JSON(http.StatusOK, gin.H{"activities": activities, "total": len(activities)})
}

// ─── Traffic Reset ─────────────────────────────────────────────────

// ResetUserTraffic resets traffic for a user and their clients.
func (h *MonitorHandler) ResetUserTraffic(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	adminID := getAdminID(c)
	role, roleErr := database.GetAdminRole(adminID)
	if roleErr != nil || (!role.OwnerRole && !rbac.CheckPermission(role, "users", "resetUsage")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reset traffic permission denied"})
		return
	}

	user, err := database.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Reset user traffic
	user.TrafficUp = 0
	user.TrafficDown = 0
	if user.Status == "limited" || user.Status == "expired" {
		user.Status = "active"
	}
	database.UpdateUser(user)

	// Reset all clients for this user
	clients, _ := database.ListClients(id, 0)
	for _, cl := range clients {
		if !cl.Enable {
			cl.Enable = true
			database.UpdateClient(&cl)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "traffic reset successful",
		"user_id":       id,
		"clients_reactivated": len(clients),
	})
}

// SyncUserTraffic syncs traffic from xray gRPC for a specific user.
func (h *MonitorHandler) SyncUserTraffic(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query param required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.xraySvc.CollectTraffic(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "sync attempted (no gRPC available)", "synced": false})
		return
	}

	// Find stats for this email
	var up, down int64
	for _, stat := range stats {
		if stat.Tag == email {
			up = stat.Up
			down = stat.Down
			break
		}
	}

	client, err := database.GetClientByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	user, err := database.GetUserByID(client.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.TrafficUp = up
	user.TrafficDown = down
	database.UpdateUser(user)

	c.JSON(http.StatusOK, gin.H{
		"message": "traffic synced from xray",
		"email":   email,
		"up":      up,
		"down":    down,
		"synced":  true,
	})
}

// ─── Reseller Management ──────────────────────────────────────────

// ResellerStats returns stats for reseller accounts.
func (h *MonitorHandler) ResellerStats(c *gin.Context) {
	admins, err := database.ListAdmins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type ResellerInfo struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		Role        string `json:"role"`
		TrafficLimit int64 `json:"traffic_limit"`
		UserLimit   int    `json:"user_limit"`
		ClientCount int64  `json:"client_count"`
		TotalTraffic int64 `json:"total_traffic"`
	}

	infos := make([]ResellerInfo, 0)
	for _, a := range admins {
		var clientCount int64
		database.DB.Model(&database.Client{}).Where("owner_admin_id = ?", a.ID).Count(&clientCount)

		var totalTraffic int64
		database.DB.Raw(`
			SELECT COALESCE(SUM(u.traffic_up + u.traffic_down), 0)
			FROM clients c
			JOIN users u ON u.id = c.user_id
			WHERE c.owner_admin_id = ?
		`, a.ID).Scan(&totalTraffic)

		infos = append(infos, ResellerInfo{
			ID:           a.ID,
			Username:     a.Username,
			Role:         a.Role,
			TrafficLimit: a.TrafficLimit,
			UserLimit:    a.UserLimit,
			ClientCount:  clientCount,
			TotalTraffic: totalTraffic,
		})
	}
	if infos == nil {
		infos = []ResellerInfo{}
	}
	c.JSON(http.StatusOK, gin.H{"resellers": infos, "total": len(infos)})
}

// ─── Telegram Bot Settings ─────────────────────────────────────────

// TelegramSettingsHandler manages Telegram bot integration for clients.
type TelegramSettingsHandler struct {
	telegramBot *service.TelegramBot
}

// NewTelegramSettingsHandler creates a new telegram settings handler.
func NewTelegramSettingsHandler(tb *service.TelegramBot) *TelegramSettingsHandler {
	return &TelegramSettingsHandler{telegramBot: tb}
}

// SetClientTelegram links a Telegram chat ID to a client.
func (h *TelegramSettingsHandler) SetClientTelegram(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		ChatID   string `json:"chat_id" binding:"required"`
		Remove bool   `json:"remove,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and chat_id required"})
		return
	}

	client, err := database.GetClientByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	adminID := getAdminID(c)
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "update")
	if !scope.ClientAllowed(client) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if req.Remove {
		client.TgID = 0
	} else {
		chatID, _ := strconv.ParseInt(req.ChatID, 10, 64)
		client.TgID = chatID
	}
	database.UpdateClient(client)

	// Send test message
	if !req.Remove && req.ChatID != "" {
		h.telegramBot.SendMessage(req.ChatID, "✅ <b>VortexUiPro Bot</b>\nTelegram notifications enabled for <code>"+client.Email+"</code>")
	}

	c.JSON(http.StatusOK, gin.H{"message": "telegram chat updated", "email": req.Email, "chat_id": req.ChatID})
}

// SendTestNotification sends a test notification to a client.
func (h *TelegramSettingsHandler) SendTestNotification(c *gin.Context) {
	var req struct {
		ChatID string `json:"chat_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id required"})
		return
	}

	h.telegramBot.SendMessage(req.ChatID, "🔔 <b>Test Notification</b>\n\nIf you see this, Telegram notifications are working correctly!\n\n<i>VortexUiPro v0.0.1</i>")
	c.JSON(http.StatusOK, gin.H{"message": "test notification sent"})
}

// NotifyClientUsage sends a usage notification to a specific client.
func (h *TelegramSettingsHandler) NotifyClientUsage(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email required"})
		return
	}

	client, err := database.GetClientByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	if client.TgID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client has no Telegram chat linked"})
		return
	}

	user, err := database.GetUserByID(client.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	usage := user.TrafficUp + user.TrafficDown
	usagePercent := 0.0
	if user.DataLimit > 0 {
		usagePercent = float64(usage) / float64(user.DataLimit) * 100
	}

	msg := fmt.Sprintf(
		"📊 <b>Usage Report</b>\n\n"+
			"<b>Client:</b> <code>%s</code>\n"+
			"<b>Traffic:</b> %.2f GB / %.2f GB\n"+
			"<b>Usage:</b> %.1f%%\n"+
			"<b>Expires:</b> %s\n",
		client.Email,
		float64(usage)/(1024*1024*1024),
		float64(user.DataLimit)/(1024*1024*1024),
		usagePercent,
		time.UnixMilli(user.ExpiryTime).Format("2006-01-02"),
	)

	h.telegramBot.SendMessage(strconv.FormatInt(client.TgID, 10), msg)
	c.JSON(http.StatusOK, gin.H{"message": "usage notification sent"})
}
