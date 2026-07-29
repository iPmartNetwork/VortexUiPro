package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// ClientGroupHandler manages client group endpoints.
type ClientGroupHandler struct {
	svc *service.ClientGroupService
}

// NewClientGroupHandler creates a new handler.
func NewClientGroupHandler(svc *service.ClientGroupService) *ClientGroupHandler {
	return &ClientGroupHandler{svc: svc}
}

// ListGroups returns all client groups.
func (h *ClientGroupHandler) ListGroups(c *gin.Context) {
	adminID, _ := strconv.ParseInt(c.Query("admin_id"), 10, 64)
	groups, err := h.svc.ListGroups(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups, "total": len(groups)})
}

// GetGroup returns a group with its members.
func (h *ClientGroupHandler) GetGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	group, err := h.svc.GetGroup(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	clients, _ := h.svc.GetGroupClients(id)
	c.JSON(http.StatusOK, gin.H{"group": group, "clients": clients})
}

// CreateGroup creates a new client group.
func (h *ClientGroupHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	adminID, _ := c.Get("admin_id")
	adminIDInt, _ := adminID.(int64)
	group, err := h.svc.CreateGroup(c.Request.Context(), req.Name, req.Description, adminIDInt)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

// UpdateGroup updates a group.
func (h *ClientGroupHandler) UpdateGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	if err := h.svc.UpdateGroup(c.Request.Context(), id, req.Name, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group updated"})
}

// DeleteGroup deletes a group.
func (h *ClientGroupHandler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteGroup(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

// AddClientToGroup adds a client to a group.
func (h *ClientGroupHandler) AddClientToGroup(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		ClientID string `json:"client_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id required"})
		return
	}
	if err := h.svc.AddClientToGroup(c.Request.Context(), groupID, req.ClientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client added to group"})
}

// RemoveClientFromGroup removes a client from a group.
func (h *ClientGroupHandler) RemoveClientFromGroup(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id query param required"})
		return
	}
	if err := h.svc.RemoveClientFromGroup(c.Request.Context(), groupID, clientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client removed from group"})
}

// GetGroupClients returns all clients in a group.
func (h *ClientGroupHandler) GetGroupClients(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	clients, err := h.svc.GetGroupClients(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": len(clients)})
}

// BulkAddClients adds multiple clients to a group.
func (h *ClientGroupHandler) BulkAddClients(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		ClientIDs []string `json:"client_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_ids required"})
		return
	}
	if err := h.svc.BulkAddClientsToGroup(c.Request.Context(), groupID, req.ClientIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "clients added to group"})
}

// BulkRemoveClients removes multiple clients from a group.
func (h *ClientGroupHandler) BulkRemoveClients(c *gin.Context) {
	groupID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		ClientIDs []string `json:"client_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_ids required"})
		return
	}
	if err := h.svc.BulkRemoveClientsFromGroup(c.Request.Context(), groupID, req.ClientIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "clients removed from group"})
}

// GetClientsWithGroups returns all clients with their group memberships.
func (h *ClientGroupHandler) GetClientsWithGroups(c *gin.Context) {
	adminID, _ := strconv.ParseInt(c.Query("admin_id"), 10, 64)
	clients, err := h.svc.GetClientsWithGroups(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": len(clients)})
}
