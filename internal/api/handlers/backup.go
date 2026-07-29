package handlers

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// BackupHandler handles backup and restore operations.
type BackupHandler struct {
	backupSvc *service.BackupService
}

// NewBackupHandler creates a new backup handler.
func NewBackupHandler(backupSvc *service.BackupService) *BackupHandler {
	return &BackupHandler{backupSvc: backupSvc}
}

// Create creates a new backup.
func (h *BackupHandler) Create(c *gin.Context) {
	var req struct {
		Name    string               `json:"name" binding:"required"`
		Scope   service.BackupScope  `json:"scope"`
		AdminID int64                `json:"admin_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	if req.Scope == "" {
		req.Scope = service.BackupScopeFull
	}

	entry, err := h.backupSvc.CreateBackup(req.Name, req.Scope, req.AdminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "entry": entry})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

// List returns all backups.
func (h *BackupHandler) List(c *gin.Context) {
	adminStr := c.Query("admin_id")
	if adminStr != "" {
		if adminID, err := strconv.ParseInt(adminStr, 10, 64); err == nil {
			c.JSON(http.StatusOK, gin.H{"backups": h.backupSvc.ListResellerBackups(adminID)})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"backups": h.backupSvc.ListBackups()})
}

// Download streams a backup file.
func (h *BackupHandler) Download(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup id required"})
		return
	}

	path, err := h.backupSvc.DownloadBackup(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	f, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer f.Close()

	c.Header("Content-Disposition", "attachment; filename=\""+path+"\"")
	c.Header("Content-Type", "application/zip")
	io.Copy(c.Writer, f)
}

// Delete removes a backup.
func (h *BackupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup id required"})
		return
	}
	if err := h.backupSvc.DeleteBackup(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "backup deleted"})
}

// Restore restores from a backup.
func (h *BackupHandler) Restore(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup id required"})
		return
	}

	var req struct {
		Scope   service.BackupScope `json:"scope"`
		AdminID int64               `json:"admin_id,omitempty"`
	}
	c.ShouldBindJSON(&req)
	if req.Scope == "" {
		req.Scope = service.BackupScopeFull
	}

	if err := h.backupSvc.RestoreBackup(id, req.Scope, req.AdminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restore initiated"})
}

// ─── Phase 14: Encryption ───────────────────────────────────────────

// GenerateEncryptionKey generates a new AES-256 key.
func (h *BackupHandler) GenerateEncryptionKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	cryptoSvc := service.NewBackupCryptoService()
	keyData, err := cryptoSvc.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	key, err := cryptoSvc.SaveKey(req.Name, keyData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": key, "key_preview": keyData[:16] + "..."})
}

// ListEncryptionKeys returns all encryption keys.
func (h *BackupHandler) ListEncryptionKeys(c *gin.Context) {
	cryptoSvc := service.NewBackupCryptoService()
	keys, err := cryptoSvc.ListKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": keys})
}

// DeleteEncryptionKey deletes an encryption key.
func (h *BackupHandler) DeleteEncryptionKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	cryptoSvc := service.NewBackupCryptoService()
	if err := cryptoSvc.DeleteKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Key deleted"})
}

// ─── Phase 14: Remote Storage Config ────────────────────────────────

// ListRemoteStorageConfigs returns all remote storage configs.
func (h *BackupHandler) ListRemoteStorageConfigs(c *gin.Context) {
	storageSvc := service.NewRemoteStorageService()
	configs, err := storageSvc.ListStorageConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

// SaveRemoteStorageConfig creates or updates a remote storage config.
func (h *BackupHandler) SaveRemoteStorageConfig(c *gin.Context) {
	var cfg database.RemoteStorageConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	storageSvc := service.NewRemoteStorageService()
	if err := storageSvc.SaveStorageConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// DeleteRemoteStorageConfig deletes a remote storage config.
func (h *BackupHandler) DeleteRemoteStorageConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	storageSvc := service.NewRemoteStorageService()
	if err := storageSvc.DeleteStorageConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Config deleted"})
}

// ─── Phase 14: Sync to Remote ───────────────────────────────────────

// SyncBackupToRemote uploads a backup to remote storage.
func (h *BackupHandler) SyncBackupToRemote(c *gin.Context) {
	backupID := c.Param("id")
	var req struct {
		StorageID int64 `json:"storage_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage_id required"})
		return
	}
	if err := h.backupSvc.SyncToRemoteStorage(backupID, req.StorageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Backup synced to remote storage"})
}

// ─── Phase 14: Telegram ─────────────────────────────────────────────

// SendBackupToTelegram sends backup info via Telegram.
func (h *BackupHandler) SendBackupToTelegram(c *gin.Context) {
	backupID := c.Param("id")
	var req struct {
		ChatID string `json:"chat_id"`
	}
	c.ShouldBindJSON(&req)
	if err := h.backupSvc.SendToTelegram(backupID, req.ChatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Backup notification sent via Telegram"})
}

// ─── AutoBackupConfig ───────────────────────────────────────────────

func (h *BackupHandler) AutoBackupConfig(c *gin.Context) {
	var req struct {
		Enabled  bool `json:"enabled"`
		IntervalHours int `json:"interval_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config"})
		return
	}
	if req.Enabled {
		h.backupSvc.StartAutoBackup(time.Duration(req.IntervalHours) * time.Hour)
	} else {
		h.backupSvc.StopAutoBackup()
	}
	c.JSON(http.StatusOK, gin.H{"message": "auto-backup configured"})
}


