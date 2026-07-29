package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// ApiTokenHandler handles API token management endpoints.
type ApiTokenHandler struct {
	tokenSvc *service.ApiTokenService
}

// NewApiTokenHandler creates a new API token handler.
func NewApiTokenHandler(tokenSvc *service.ApiTokenService) *ApiTokenHandler {
	return &ApiTokenHandler{tokenSvc: tokenSvc}
}

// List returns all API tokens.
func (h *ApiTokenHandler) List(c *gin.Context) {
	tokens, err := h.tokenSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tokens == nil {
		tokens = []*service.ApiTokenView{}
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens, "total": len(tokens)})
}

// ListDelegatedSubjects returns admins eligible for delegated tokens.
func (h *ApiTokenHandler) ListDelegatedSubjects(c *gin.Context) {
	subjects, err := h.tokenSvc.ListDelegatedSubjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if subjects == nil {
		subjects = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"subjects": subjects})
}

// Create creates a new API token.
func (h *ApiTokenHandler) Create(c *gin.Context) {
	var req struct {
		Name           string   `json:"name" binding:"required"`
		Kind           string   `json:"kind,omitempty"`
		SubjectAdminID int      `json:"subjectAdminId,omitempty"`
		Scopes         []string `json:"scopes,omitempty"`
		ExpiresAt      int64    `json:"expiresAt,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Get creator ID
	var createdBy int
	if adminID, exists := c.Get("admin_id"); exists {
		switch v := adminID.(type) {
		case int64:
			createdBy = int(v)
		case float64:
			createdBy = int(v)
		}
	}

	token, err := h.tokenSvc.CreateWithOptions(service.ApiTokenCreateOptions{
		Name:             req.Name,
		Kind:             req.Kind,
		SubjectAdminID:   req.SubjectAdminID,
		CreatedByAdminID: createdBy,
		Scopes:           req.Scopes,
		ExpiresAt:        req.ExpiresAt,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, token)
}

// Delete deletes an API token.
func (h *ApiTokenHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}
	if err := h.tokenSvc.Delete(id); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token deleted"})
}

// SetEnabled enables or disables a token.
func (h *ApiTokenHandler) SetEnabled(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.tokenSvc.SetEnabled(id, req.Enabled); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token updated"})
}
