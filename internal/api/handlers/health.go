package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// ─── HealthHandler ───────────────────────────────────────────────────

type HealthHandler struct {
	svc *service.HealthCheckService
}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler(svc *service.HealthCheckService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// ─── Check Configs ─────────────────────────────────────────────────

// ListCheckConfigs returns all health check configurations.
func (h *HealthHandler) ListCheckConfigs(c *gin.Context) {
	configs, err := h.svc.ListCheckConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if configs == nil {
		configs = []database.HealthCheckConfig{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

// CreateCheckConfig creates a new health check configuration.
func (h *HealthHandler) CreateCheckConfig(c *gin.Context) {
	var cfg database.HealthCheckConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 30
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 3
	}

	if err := h.svc.CreateCheckConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cfg})
}

// UpdateCheckConfig updates a health check configuration.
func (h *HealthHandler) UpdateCheckConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var cfg database.HealthCheckConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateCheckConfig(id, &cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Config updated"})
}

// DeleteCheckConfig deletes a health check configuration.
func (h *HealthHandler) DeleteCheckConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteCheckConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Config deleted"})
}

// ─── Recovery Rules ────────────────────────────────────────────────

// ListRecoveryRules returns all auto-recovery rules.
func (h *HealthHandler) ListRecoveryRules(c *gin.Context) {
	rules, err := h.svc.ListRecoveryRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rules == nil {
		rules = []database.AutoRecoveryRule{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rules})
}

// CreateRecoveryRule creates a new auto-recovery rule.
func (h *HealthHandler) CreateRecoveryRule(c *gin.Context) {
	var rule database.AutoRecoveryRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateRecoveryRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": rule})
}

// UpdateRecoveryRule updates an auto-recovery rule.
func (h *HealthHandler) UpdateRecoveryRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var rule database.AutoRecoveryRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateRecoveryRule(id, &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Rule updated"})
}

// DeleteRecoveryRule deletes an auto-recovery rule.
func (h *HealthHandler) DeleteRecoveryRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteRecoveryRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Rule deleted"})
}

// ─── Status & History ──────────────────────────────────────────────

// GetStatuses returns current health status of all checks.
func (h *HealthHandler) GetStatuses(c *gin.Context) {
	statuses := h.svc.GetStatuses()
	if statuses == nil {
		statuses = []*service.HealthStatus{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": statuses})
}

// GetHistory returns health check result history.
func (h *HealthHandler) GetHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	results, err := h.svc.GetHistory(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if results == nil {
		results = []database.HealthCheckResult{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// GetRecoveryHistory returns recovery action history.
func (h *HealthHandler) GetRecoveryHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	actions, err := h.svc.GetRecoveryHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if actions == nil {
		actions = []database.AutoRecoveryAction{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": actions})
}

// RunManualCheck triggers a manual health check.
func (h *HealthHandler) RunManualCheck(c *gin.Context) {
	var req struct {
		ConfigID int64 `json:"config_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.RunManualCheck(req.ConfigID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
