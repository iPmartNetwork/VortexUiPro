package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// SubscriptionHandler handles subscription endpoints for client config generation.
type SubscriptionHandler struct {
	subSvc  *service.SubscriptionService
	userSvc *service.UserService
}

// NewSubscriptionHandler creates a new subscription handler.
func NewSubscriptionHandler(subSvc *service.SubscriptionService, userSvc *service.UserService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subSvc:  subSvc,
		userSvc: userSvc,
	}
}

// GetConfig handles the main subscription request with format negotiation.
func (h *SubscriptionHandler) GetConfig(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		clientID = c.Query("token")
	}
	if clientID == "" {
		c.String(http.StatusBadRequest, "missing client identifier")
		return
	}

	client, err := h.userSvc.GetClient(clientID)
	if err != nil {
		c.String(http.StatusNotFound, "invalid subscription")
		return
	}

	user, err := h.userSvc.GetUser(client.UserID)
	if err != nil {
		c.String(http.StatusNotFound, "user not found")
		return
	}

	accept := c.GetHeader("Accept")
	format := c.DefaultQuery("format", "xray")

	var config string
	switch {
	case strings.Contains(accept, "clash") || format == "clash":
		config, err = h.subSvc.GenerateClashConfig(client.UserID, clientID)
	case strings.Contains(accept, "sing-box") || format == "singbox":
		config, err = h.subSvc.GenerateSingboxConfig(client.UserID, clientID)
	default:
		config, err = h.subSvc.GenerateXrayJSON(client.UserID, clientID)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "failed to generate config")
		return
	}

	// Set content type based on format
	switch {
	case strings.Contains(accept, "clash") || format == "clash":
		c.Header("Content-Type", "application/yaml; charset=utf-8")
	default:
		c.Header("Content-Type", "application/json; charset=utf-8")
	}

	// Set Subscription-Userinfo header
	c.Header("Subscription-Userinfo", service.BuildUserInfo(user))
	c.String(http.StatusOK, config)
}

// GetLink generates a subscription link for the user.
func (h *SubscriptionHandler) GetLink(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client id required"})
		return
	}

	host := c.Request.Host
	url := h.subSvc.GenerateLink(0, clientID, host, 0)

	c.JSON(http.StatusOK, gin.H{
		"url":       url,
		"client_id": clientID,
		"format":    c.DefaultQuery("format", "xray"),
	})
}

// GetInfo returns subscription info (traffic/expiry) for a client.
func (h *SubscriptionHandler) GetInfo(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client id required"})
		return
	}

	client, err := h.userSvc.GetClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	user, err := h.userSvc.GetUser(client.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_email": client.Email,
		"traffic_up":   user.TrafficUp,
		"traffic_down": user.TrafficDown,
		"traffic_total": user.DataLimit,
		"expiry_time":  user.ExpiryTime,
		"status":       user.Status,
	})
}

// GetShareLinks returns share links for a client in all supported formats.
func (h *SubscriptionHandler) GetShareLinks(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		clientID = c.Query("id")
	}
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client id required"})
		return
	}

	links, err := h.subSvc.GenerateLinks(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": links, "count": len(links)})
}

// GetSubLinks returns subscription links for all clients matching a sub ID.
func (h *SubscriptionHandler) GetSubLinks(c *gin.Context) {
	subID := c.Param("subId")
	if subID == "" {
		subID = c.Query("token")
	}
	if subID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sub id required"})
		return
	}

	host := c.Request.Host
	links, err := h.subSvc.GenerateSubLinks(subID, host, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(links) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no links found"})
		return
	}

	// Return as base64-encoded body for subscription clients
	body := strings.Join(links, "\n")
	encoded := service.EncodeSubResponse(body)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Subscription-Userinfo", fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", 0, 0, 0, 0))
	c.String(http.StatusOK, encoded)
}
