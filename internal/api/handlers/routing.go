package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// RoutingHandler manages routing rules and packs.
type RoutingHandler struct {
	svc *service.RoutingService
}

// NewRoutingHandler creates a new routing handler.
func NewRoutingHandler(svc *service.RoutingService) *RoutingHandler {
	return &RoutingHandler{svc: svc}
}

// ─── Rules ──────────────────────────────────────────────────────────

// ListRules returns all routing rules.
func (h *RoutingHandler) ListRules(c *gin.Context) {
	rules, err := h.svc.ListRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rules": rules, "total": len(rules)})
}

// GetRule returns a single routing rule.
func (h *RoutingHandler) GetRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rule, err := h.svc.GetRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// CreateRuleRequest is the rule creation payload.
type CreateRuleRequest struct {
	OutboundTag string   `json:"outbound_tag" binding:"required"`
	InboundTags []string `json:"inbound_tags,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	GeoIP       []string `json:"geoip,omitempty"`
	GeoSite     []string `json:"geosite,omitempty"`
	SourceIP    []string `json:"source_ip,omitempty"`
	BalancerTag string   `json:"balancer_tag,omitempty"`
	RuleType    string   `json:"rule_type,omitempty"`
	Enabled     bool     `json:"enable"`
}

// CreateRule adds a new routing rule.
func (h *RoutingHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outbound_tag required"})
		return
	}

	rule := &database.RoutingRule{
		OutboundTag: req.OutboundTag,
		Network:     req.Network,
		Port:        req.Port,
		BalancerTag: req.BalancerTag,
		RuleType:    req.RuleType,
		Enabled:     req.Enabled,
	}

	if req.Domain != nil { data, _ := json.Marshal(req.Domain); rule.Domain = string(data) }
	if req.IP != nil { data, _ := json.Marshal(req.IP); rule.IP = string(data) }
	if req.Protocol != nil { data, _ := json.Marshal(req.Protocol); rule.Protocol = string(data) }
	if req.GeoIP != nil { data, _ := json.Marshal(req.GeoIP); rule.GeoIP = string(data) }
	if req.GeoSite != nil { data, _ := json.Marshal(req.GeoSite); rule.GeoSite = string(data) }
	if req.SourceIP != nil { data, _ := json.Marshal(req.SourceIP); rule.SourceIP = string(data) }
	if req.InboundTags != nil { data, _ := json.Marshal(req.InboundTags); rule.InboundTags = string(data) }

	created, err := h.svc.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// UpdateRule modifies a routing rule.
func (h *RoutingHandler) UpdateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.svc.GetRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	existing.OutboundTag = req.OutboundTag
	existing.Network = req.Network
	existing.Port = req.Port
	existing.BalancerTag = req.BalancerTag
	existing.RuleType = req.RuleType
	existing.Enabled = req.Enabled
	if req.Domain != nil { data, _ := json.Marshal(req.Domain); existing.Domain = string(data) }
	if req.IP != nil { data, _ := json.Marshal(req.IP); existing.IP = string(data) }
	if err := h.svc.UpdateRule(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule updated", "rule": existing})
}

// DeleteRule removes a routing rule.
func (h *RoutingHandler) DeleteRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
}

// ToggleRule enables/disables a routing rule.
func (h *RoutingHandler) ToggleRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct { Enabled bool `json:"enable"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.ToggleRule(c.Request.Context(), id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule updated"})
}

// ─── Packs ──────────────────────────────────────────────────────────

// ListPacks returns all routing packs.
func (h *RoutingHandler) ListPacks(c *gin.Context) {
	packs, err := h.svc.ListPacks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"packs": packs})
}

// CreatePack adds a new routing pack.
func (h *RoutingHandler) CreatePack(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description,omitempty"`
		RuleIDs     []int64 `json:"rule_ids,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	ruleIDsStr, _ := json.Marshal(req.RuleIDs)
	pack := &database.RoutingPack{
		Name:        req.Name,
		Description: req.Description,
		RuleIDs:     string(ruleIDsStr),
		Enabled:     true,
	}
	created, err := h.svc.CreatePack(c.Request.Context(), pack)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// DeletePack removes a routing pack.
func (h *RoutingHandler) DeletePack(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeletePack(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pack deleted"})
}

// GenerateConfig generates the Xray routing config from enabled rules.
func (h *RoutingHandler) GenerateConfig(c *gin.Context) {
	config, err := h.svc.GenerateXrayConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": config})
}
