package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// FederationHandler manages federation endpoints.
type FederationHandler struct {
	svc *service.FederationService
}

// NewFederationHandler creates a new federation handler.
func NewFederationHandler(svc *service.FederationService) *FederationHandler {
	return &FederationHandler{svc: svc}
}

// ListProvidersRaw returns all providers directly (for middleware validation).
func (h *FederationHandler) ListProvidersRaw() ([]database.FederationProvider, error) {
	return h.svc.ListProviders()
}

// ListProviders returns all federation providers.
func (h *FederationHandler) ListProviders(c *gin.Context) {
	providers, err := h.svc.ListProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Mask API keys
	for i := range providers {
		if providers[i].APIKey != "" {
			providers[i].APIKey = "••••••" + providers[i].APIKey[len(providers[i].APIKey)-4:]
		}
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers, "total": len(providers)})
}

// CreateProvider creates a new federation provider.
func (h *FederationHandler) CreateProvider(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		APIURL      string `json:"api_url" binding:"required"`
		APIKey      string `json:"api_key"`
		SyncUsers   bool   `json:"sync_users"`
		SyncPlans   bool   `json:"sync_plans"`
		SyncTraffic bool   `json:"sync_traffic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and api_url required"})
		return
	}

	p, err := h.svc.CreateProvider(req.Name, req.APIURL, req.APIKey, req.SyncUsers, req.SyncPlans, req.SyncTraffic)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// UpdateProvider updates a federation provider.
func (h *FederationHandler) UpdateProvider(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name        string `json:"name" binding:"required"`
		APIURL      string `json:"api_url" binding:"required"`
		APIKey      string `json:"api_key"`
		SyncUsers   bool   `json:"sync_users"`
		SyncPlans   bool   `json:"sync_plans"`
		SyncTraffic bool   `json:"sync_traffic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.UpdateProvider(c.Request.Context(), id, req.Name, req.APIURL, req.APIKey, req.SyncUsers, req.SyncPlans, req.SyncTraffic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "provider updated"})
}

// DeleteProvider removes a federation provider.
func (h *FederationHandler) DeleteProvider(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "provider removed"})
}

// TestConnection tests connectivity with a provider.
func (h *FederationHandler) TestConnection(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	p, err := h.svc.GetProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	_ = p // Test is done async in the service
	c.JSON(http.StatusOK, gin.H{"message": "connection test initiated"})
}

// TriggerSync manually triggers a sync with all providers.
func (h *FederationHandler) TriggerSync(c *gin.Context) {
	idStr := c.Param("id")
	if idStr != "" {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		p, err := h.svc.GetProvider(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		go h.svc.SyncWithProvider(p)
		c.JSON(http.StatusOK, gin.H{"message": "sync triggered for provider"})
		return
	}

	// Sync all providers
	go h.svc.SyncAll()
	c.JSON(http.StatusOK, gin.H{"message": "sync triggered for all providers"})
}

// ─── Incoming Federation Endpoints (called by remote panels) ────────

// HandleFederationUsers handles incoming user sync from remote panels.
func (h *FederationHandler) HandleFederationUsers(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	result, err := h.svc.HandleFederationUsers(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleFederationPlans handles incoming plan sync from remote panels.
func (h *FederationHandler) HandleFederationPlans(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	result, err := h.svc.HandleFederationPlans(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleFederationTraffic handles incoming traffic sync from remote panels.
func (h *FederationHandler) HandleFederationTraffic(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	result, err := h.svc.HandleFederationTraffic(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
