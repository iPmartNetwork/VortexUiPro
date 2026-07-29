package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
)

// NodeHandler manages proxy server node endpoints.
type NodeHandler struct{}

// NewNodeHandler creates a new node handler.
func NewNodeHandler() *NodeHandler {
	return &NodeHandler{}
}

// List returns all nodes.
func (h *NodeHandler) List(c *gin.Context) {
	nodes, err := database.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": len(nodes)})
}

// Get returns a single node by ID.
func (h *NodeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	node, err := database.GetNodeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, node)
}

// CreateRequest is the node creation payload.
type CreateNodeRequest struct {
	Name     string `json:"name" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Port     int    `json:"port"`
	APIPort  int    `json:"api_port"`
	CoreType string `json:"core_type"`
	Location string `json:"location,omitempty"`
	Country  string `json:"country,omitempty"`
	Remark   string `json:"remark,omitempty"`
}

// Create adds a new node.
func (h *NodeHandler) Create(c *gin.Context) {
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and address required"})
		return
	}

	node := &database.Node{
		Name:     req.Name,
		Address:  req.Address,
		Port:     req.Port,
		APIPort:  req.APIPort,
		CoreType: req.CoreType,
		Location: req.Location,
		Country:  req.Country,
		Remark:   req.Remark,
		Status:   "offline",
		Enable:   true,
	}

	if node.Port == 0 {
		node.Port = 2053
	}
	if node.APIPort == 0 {
		node.APIPort = 10085
	}
	if node.CoreType == "" {
		node.CoreType = "xray"
	}

	if err := database.CreateNode(node); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, node)
}

// Update modifies an existing node.
func (h *NodeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := database.GetNodeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Apply updates
	if v, ok := updates["name"]; ok { existing.Name = v.(string) }
	if v, ok := updates["address"]; ok { existing.Address = v.(string) }
	if v, ok := updates["port"]; ok { existing.Port = int(v.(float64)) }
	if v, ok := updates["api_port"]; ok { existing.APIPort = int(v.(float64)) }
	if v, ok := updates["status"]; ok { existing.Status = v.(string) }
	if v, ok := updates["location"]; ok { existing.Location = v.(string) }
	if v, ok := updates["country"]; ok { existing.Country = v.(string) }
	if v, ok := updates["remark"]; ok { existing.Remark = v.(string) }
	if v, ok := updates["enable"]; ok { existing.Enable = v.(bool) }

	if err := database.UpdateNode(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "node updated", "node": existing})
}

// Delete removes a node.
func (h *NodeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := database.DeleteNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}
