package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/domain"
	"vortexuipro/internal/rbac"
	"vortexuipro/internal/service"
)

type InboundHandler struct {
	inboundSvc *service.InboundService
	xraySvc    *service.XrayService
}

func NewInboundHandler(inboundSvc *service.InboundService, xraySvc *service.XrayService) *InboundHandler {
	return &InboundHandler{
		inboundSvc: inboundSvc,
		xraySvc:    xraySvc,
	}
}

func (h *InboundHandler) List(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	nodeID, _ := strconv.ParseInt(c.Query("node_id"), 10, 64)

	inbounds, err := h.inboundSvc.List(userID, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// RBAC: Filter inbounds by role access scope
	adminID := getAdminID(c)
	if restricted, allowedIDs := rbac.RestrictedInboundIDs(adminID); restricted {
		allowedSet := make(map[int64]struct{}, len(allowedIDs))
		for _, id := range allowedIDs {
			allowedSet[int64(id)] = struct{}{}
		}
		filtered := make([]database.Inbound, 0, len(inbounds))
		for _, ib := range inbounds {
			if _, ok := allowedSet[ib.ID]; ok {
				filtered = append(filtered, ib)
			}
		}
		inbounds = filtered
	}

	c.JSON(http.StatusOK, gin.H{"inbounds": inbounds, "total": len(inbounds)})
}

func (h *InboundHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// RBAC: Check inbound access
	adminID := getAdminID(c)
	if restricted, allowedIDs := rbac.RestrictedInboundIDs(adminID); restricted {
		allowed := false
		for _, allowedID := range allowedIDs {
			if int64(allowedID) == id {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "inbound access denied"})
			return
		}
	}

	ib, err := h.inboundSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "inbound not found"})
		return
	}
	c.JSON(http.StatusOK, ib)
}

func (h *InboundHandler) Create(c *gin.Context) {
	var ib database.Inbound
	if err := c.ShouldBindJSON(&ib); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	created, err := h.inboundSvc.Create(c.Request.Context(), &ib)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *InboundHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ib database.Inbound
	if err := c.ShouldBindJSON(&ib); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	ib.ID = id
	if err := h.inboundSvc.Update(c.Request.Context(), &ib); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "inbound updated"})
}

func (h *InboundHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.inboundSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "inbound deleted"})
}

func (h *InboundHandler) GetXrayConfig(c *gin.Context) {
	inbounds, err := h.inboundSvc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ibList := make([]domain.Inbound, 0, len(inbounds))
	for _, ib := range inbounds {
		ibList = append(ibList, domain.Inbound{
			ID:             ib.ID,
			Tag:            ib.Tag,
			Protocol:       domain.Protocol(ib.Protocol),
			Listen:         ib.Listen,
			Port:           ib.Port,
			Settings:       ib.Settings,
			StreamSettings: ib.StreamSettings,
			Sniffing:       ib.Sniffing,
			Status:         domain.InboundStatus(ib.Status),
			Enable:         ib.Enable,
			Remark:         ib.Remark,
		})
	}

	config, err := domain.GenerateXrayConfig(ibList, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config, "inbounds": len(ibList)})
}
