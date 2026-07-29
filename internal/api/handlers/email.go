package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// EmailHandler handles email/SMTP endpoints.
type EmailHandler struct {
	svc *service.EmailService
}

// NewEmailHandler creates a new email handler.
func NewEmailHandler(svc *service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

// GetConfig returns the current SMTP configuration.
func (h *EmailHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.LoadConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"configured": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configured": true,
		"host":       cfg.Host,
		"port":       cfg.Port,
		"username":   cfg.Username,
		"from":       cfg.From,
		"use_tls":    cfg.UseTLS,
		"password":   cfg.Password != "",
	})
}

// SaveConfig saves SMTP configuration.
func (h *EmailHandler) SaveConfig(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" binding:"required"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
		UseTLS   bool   `json:"use_tls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host and port required"})
		return
	}

	cfg := &service.SMTPConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		From:     req.From,
		UseTLS:   req.UseTLS,
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}

	if err := h.svc.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SMTP configuration saved"})
}

// SendTest sends a test email.
func (h *EmailHandler) SendTest(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipient email required"})
		return
	}

	if err := h.svc.Send(req.To, "Test Email from VortexUiPro", "<h1>Test</h1><p>If you receive this, SMTP is working correctly.</p>"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}
