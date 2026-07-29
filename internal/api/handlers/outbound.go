package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// OutboundHandler manages proxy outbound endpoints.
type OutboundHandler struct {
	svc *service.OutboundService
}

// NewOutboundHandler creates a new outbound handler.
func NewOutboundHandler(svc *service.OutboundService) *OutboundHandler {
	return &OutboundHandler{svc: svc}
}

// List returns all outbounds.
func (h *OutboundHandler) List(c *gin.Context) {
	nodeID, _ := strconv.ParseInt(c.Query("node_id"), 10, 64)
	outbounds, err := h.svc.List(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outbounds": outbounds, "total": len(outbounds)})
}

// Get returns a single outbound by ID.
func (h *OutboundHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ob, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "outbound not found"})
		return
	}
	c.JSON(http.StatusOK, ob)
}

// CreateRequest is the outbound creation payload.
type CreateOutboundRequest struct {
	Tag            string `json:"tag" binding:"required"`
	Protocol       string `json:"protocol" binding:"required"`
	NodeID         int64  `json:"node_id,omitempty"`
	Settings       string `json:"settings,omitempty"`
	StreamSettings string `json:"stream_settings,omitempty"`
	Remark         string `json:"remark,omitempty"`
	Enable         bool   `json:"enable"`
	Hidden         bool   `json:"hidden,omitempty"`
}

// Create adds a new outbound.
func (h *OutboundHandler) Create(c *gin.Context) {
	var req CreateOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag and protocol required"})
		return
	}

	ob := &database.Outbound{
		Tag:            req.Tag,
		Protocol:       req.Protocol,
		NodeID:         req.NodeID,
		Settings:       req.Settings,
		StreamSettings: req.StreamSettings,
		Remark:         req.Remark,
		Enable:         req.Enable,
		Hidden:         req.Hidden,
	}

	created, err := h.svc.Create(c.Request.Context(), ob)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// UpdateRequest is the outbound update payload.
type UpdateOutboundRequest struct {
	Tag            *string `json:"tag,omitempty"`
	Protocol       *string `json:"protocol,omitempty"`
	NodeID         *int64  `json:"node_id,omitempty"`
	Settings       *string `json:"settings,omitempty"`
	StreamSettings *string `json:"stream_settings,omitempty"`
	Remark         *string `json:"remark,omitempty"`
	Enable         *bool   `json:"enable,omitempty"`
	Hidden         *bool   `json:"hidden,omitempty"`
}

// Update modifies an existing outbound.
func (h *OutboundHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "outbound not found"})
		return
	}

	var req UpdateOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Tag != nil { existing.Tag = *req.Tag }
	if req.Protocol != nil { existing.Protocol = *req.Protocol }
	if req.NodeID != nil { existing.NodeID = *req.NodeID }
	if req.Settings != nil { existing.Settings = *req.Settings }
	if req.StreamSettings != nil { existing.StreamSettings = *req.StreamSettings }
	if req.Remark != nil { existing.Remark = *req.Remark }
	if req.Enable != nil { existing.Enable = *req.Enable }
	if req.Hidden != nil { existing.Hidden = *req.Hidden }

	if err := h.svc.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "outbound updated", "outbound": existing})
}

// Delete removes an outbound.
func (h *OutboundHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "outbound deleted"})
}

// ToggleHide shows or hides an outbound from subscription exports.
func (h *OutboundHandler) ToggleHide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Hidden bool `json:"hidden"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.ToggleHide(c.Request.Context(), id, req.Hidden); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "outbound visibility updated"})
}
