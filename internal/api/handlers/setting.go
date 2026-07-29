package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
)

// SettingHandler manages panel settings endpoints.
type SettingHandler struct{}

// NewSettingHandler creates a new settings handler.
func NewSettingHandler() *SettingHandler {
	return &SettingHandler{}
}

// List returns all settings as a key-value map.
func (h *SettingHandler) List(c *gin.Context) {
	settings := make(map[string]string)

	// Predefined setting keys
	keys := []string{
		"panel_port", "panel_language", "panel_brand",
		"sub_port", "sub_path", "sub_enable",
		"telegram_enabled", "telegram_token", "telegram_chat_id",
		"webhook_enabled", "webhook_url",
		"totp_enabled",
		"tunnel_monitor_enabled", "tunnel_monitor_url",
		"auto_restart_core", "log_level",
		"default_core",
	}

	for _, key := range keys {
		val, err := database.GetSetting(key)
		if err == nil {
			settings[key] = val
		}
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// UpdateRequest is the settings update payload.
type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// Update saves multiple settings at once.
func (h *SettingHandler) Update(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "settings map required"})
		return
	}

	var errors []string
	for key, value := range req.Settings {
		if err := database.SetSetting(key, value); err != nil {
			errors = append(errors, key+": "+err.Error())
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "settings saved with some errors",
			"errors":  errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings saved successfully"})
}

// Get returns a single setting by key.
func (h *SettingHandler) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
		return
	}

	val, err := database.GetSetting(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": val})
}
