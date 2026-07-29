package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// AdvancedSecurityHandler handles advanced security endpoints.
type AdvancedSecurityHandler struct {
	svc *service.AdvancedSecurityService
}

// NewAdvancedSecurityHandler creates a new advanced security handler.
func NewAdvancedSecurityHandler(svc *service.AdvancedSecurityService) *AdvancedSecurityHandler {
	return &AdvancedSecurityHandler{svc: svc}
}

// ListAuditLogs returns audit log entries.
func (h *AdvancedSecurityHandler) ListAuditLogs(c *gin.Context) {
	actor := c.Query("actor")
	action := c.Query("action")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, total, err := h.svc.ListAuditLogs(c.Request.Context(), actor, action, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries, "total": total})
}

// GetThreatSummary returns threat detection summary.
func (h *AdvancedSecurityHandler) GetThreatSummary(c *gin.Context) {
	summary := h.svc.GetThreatSummary()
	c.JSON(http.StatusOK, summary)
}

// RunComplianceCheck runs all compliance checks.
func (h *AdvancedSecurityHandler) RunComplianceCheck(c *gin.Context) {
	report, err := h.svc.RunComplianceCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
