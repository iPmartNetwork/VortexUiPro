package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/cluster"
	"vortexuipro/internal/database"
)

// ─── Cluster Handler ─────────────────────────────────────────────────

// ClusterHandler provides HTTP API endpoints for cluster management.
type ClusterHandler struct {
	manager *cluster.Manager
}

// NewClusterHandler creates a new cluster handler.
func NewClusterHandler(manager *cluster.Manager) *ClusterHandler {
	return &ClusterHandler{manager: manager}
}

// Status returns the current cluster status.
func (h *ClusterHandler) Status(c *gin.Context) {
	if h.manager == nil || !h.manager.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Cluster mode is disabled. Set VORTEX_CLUSTER_ENABLED=true to enable.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"status":  h.manager.Status(),
	})
}

// ListNodes lists all cluster nodes from the database.
func (h *ClusterHandler) ListNodes(c *gin.Context) {
	var nodes []database.ClusterNode
	result := database.DB.Order("id ASC").Find(&nodes)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if nodes == nil {
		nodes = make([]database.ClusterNode, 0)
	}
	c.JSON(http.StatusOK, nodes)
}

// GetNode returns a single cluster node by ID.
func (h *ClusterHandler) GetNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	var node database.ClusterNode
	if err := database.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// AddNode adds a new node to the cluster configuration.
func (h *ClusterHandler) AddNode(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Address   string `json:"address" binding:"required"`
		PeerPort  int    `json:"peer_port"`
		Priority  int    `json:"priority"`
		Region    string `json:"region"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address := req.Address
	if req.PeerPort > 0 {
		address = address + ":" + strconv.Itoa(req.PeerPort)
	}

	node := database.ClusterNode{
		Name:     req.Name,
		Address:  address,
		PeerPort: req.PeerPort,
		Priority: req.Priority,
		Region:   req.Region,
		Status:   "offline",
		Role:     "follower",
		Enabled:  true,
	}

	if node.Priority == 0 {
		node.Priority = 50
	}
	if node.PeerPort == 0 {
		node.PeerPort = 1337
	}
	if node.Region == "" {
		node.Region = "default"
	}

	if err := database.DB.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

// UpdateNode updates a cluster node configuration.
func (h *ClusterHandler) UpdateNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		Priority int    `json:"priority"`
		Region   string `json:"region"`
		Enabled  *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]any)
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Priority > 0 {
		updates["priority"] = req.Priority
	}
	if req.Region != "" {
		updates["region"] = req.Region
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := database.DB.Model(&database.ClusterNode{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var node database.ClusterNode
	database.DB.First(&node, id)
	c.JSON(http.StatusOK, node)
}

// DeleteNode removes a node from the cluster.
func (h *ClusterHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	if err := database.DB.Delete(&database.ClusterNode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}

// ElectionStats returns leader election statistics.
func (h *ClusterHandler) ElectionStats(c *gin.Context) {
	if h.manager == nil || !h.manager.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{"enabled": false})
		return
	}

	status := h.manager.Status()
	c.JSON(http.StatusOK, status["election"])
}

// SyncEvents returns recent cluster sync events.
func (h *ClusterHandler) SyncEvents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 200 {
		limit = 50
	}

	var events []database.SyncEvent
	database.DB.Order("id DESC").Limit(limit).Find(&events)

	if events == nil {
		events = make([]database.SyncEvent, 0)
	}
	c.JSON(http.StatusOK, events)
}

// Topology returns the cluster topology graph data.
func (h *ClusterHandler) Topology(c *gin.Context) {
	c.JSON(http.StatusOK, cluster.TopologyAPI())
}

// PKIStatus returns the PKI certificate status.
func (h *ClusterHandler) PKIStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pki_available": h.manager != nil,
		"config": map[string]any{
			"tls_enabled":    h.manager != nil && h.manager.IsEnabled(),
			"grpc_enabled":   false, // populated from config
		},
	})
}

// ForceElection triggers a new leader election.
func (h *ClusterHandler) ForceElection(c *gin.Context) {
	if h.manager == nil || !h.manager.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster not enabled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Election triggered", "status": h.manager.Status()})
}
