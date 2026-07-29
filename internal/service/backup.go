package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"gorm.io/gorm/clause"
)

// ─── Backup Service ─────────────────────────────────────────────────

// BackupScope defines what data to include in a backup.
type BackupScope string

const (
	BackupScopeFull     BackupScope = "full"     // everything
	BackupScopeSystem   BackupScope = "system"   // admins, roles, settings
	BackupScopeUsers    BackupScope = "users"    // users, clients, inbounds
	BackupScopeReseller BackupScope = "reseller" // specific reseller's data
)

// BackupEntry represents a single backup record.
type BackupEntry struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Scope             BackupScope `json:"scope"`
	Size              int64       `json:"size"`
	Path              string      `json:"path"`
	AdminID           int64       `json:"admin_id,omitempty"`
	AdminName         string      `json:"admin_name,omitempty"`
	Encrypted         bool        `json:"encrypted"`
	EncryptionKeyID   int64       `json:"encryption_key_id,omitempty"`
	RemoteStorageID   int64       `json:"remote_storage_id,omitempty"`
	RemoteStorageName string      `json:"remote_storage_name,omitempty"`
	TelegramSent      bool        `json:"telegram_sent,omitempty"`
	CreatedAt         int64       `json:"created_at"`
	Status            string      `json:"status"` // completed, failed, sent
}

// BackupService manages system backups.
type BackupService struct {
	mu            sync.Mutex
	backupDir     string
	backups       []BackupEntry
	autoEnabled   bool
	autoInterval  time.Duration
	stopAuto      chan struct{}
	cryptoSvc     *BackupCryptoService
	storageSvc    *RemoteStorageService
	telegramBot   *TelegramBot
}

// NewBackupService creates a new backup service.
func NewBackupService(backupDir string, cryptoSvc *BackupCryptoService, storageSvc *RemoteStorageService, telegramBot *TelegramBot) *BackupService {
	if backupDir == "" {
		backupDir = "/etc/vortex/backups"
	}
	os.MkdirAll(backupDir, 0755)
	if cryptoSvc == nil {
		cryptoSvc = NewBackupCryptoService()
	}
	if storageSvc == nil {
		storageSvc = NewRemoteStorageService()
	}
	return &BackupService{
		backupDir:  backupDir,
		backups:    make([]BackupEntry, 0),
		stopAuto:   make(chan struct{}),
		cryptoSvc:  cryptoSvc,
		storageSvc: storageSvc,
		telegramBot: telegramBot,
	}
}

// StartAutoBackup begins automatic backups on an interval.
func (s *BackupService) StartAutoBackup(interval time.Duration) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	s.autoInterval = interval
	s.autoEnabled = true
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// Initial backup after 1 minute
		time.Sleep(time.Minute)
		s.CreateBackup("auto-full", BackupScopeFull, 0)
		for {
			select {
			case <-ticker.C:
				s.CreateBackup(fmt.Sprintf("auto-%s", time.Now().Format("2006-01-02")), BackupScopeFull, 0)
			case <-s.stopAuto:
				return
			}
		}
	}()
}

// StopAutoBackup stops the auto-backup loop.
func (s *BackupService) StopAutoBackup() {
	if s.autoEnabled {
		close(s.stopAuto)
		s.autoEnabled = false
	}
}

// CreateBackup creates a new backup archive.
func (s *BackupService) CreateBackup(name string, scope BackupScope, adminID int64) (*BackupEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().UnixMilli()
	filename := fmt.Sprintf("%s-%d.zip", sanitizeFilename(name), timestamp)
	filePath := filepath.Join(s.backupDir, filename)

	// Determine admin name for reseller backups
	adminName := ""
	if adminID > 0 {
		if admin, err := database.GetAdminByID(adminID); err == nil {
			adminName = admin.Username
		}
	}

	entry := &BackupEntry{
		ID:        fmt.Sprintf("bkp-%d", timestamp),
		Name:      name,
		Scope:     scope,
		Path:      filePath,
		AdminID:   adminID,
		AdminName: adminName,
		CreatedAt: timestamp,
		Status:    "completed",
	}

	if err := s.writeBackupArchive(filePath, scope, adminID); err != nil {
		entry.Status = "failed"
		s.backups = append(s.backups, *entry)
		return entry, fmt.Errorf("backup failed: %w", err)
	}

	// Get file size
	if fi, err := os.Stat(filePath); err == nil {
		entry.Size = fi.Size()
	}

	// Auto-encrypt if an encryption key exists
	if key, err := s.cryptoSvc.GetActiveKey(); err == nil && key != nil {
		encPath := filePath + ".enc"
		if err := s.cryptoSvc.EncryptFile(filePath, encPath, int64(key.ID)); err != nil {
			log.Printf("[Phase 14] Backup encryption failed: %v", err)
		} else {
			// Replace original with encrypted
			os.Remove(filePath)
			os.Rename(encPath, filePath)
			entry.Encrypted = true
			entry.EncryptionKeyID = int64(key.ID)
			log.Printf("[Phase 14] Backup encrypted with key: %s", key.Name)
		}
	}

	s.backups = append(s.backups, *entry)
	return entry, nil
}

func (s *BackupService) writeBackupArchive(filePath string, scope BackupScope, adminID int64) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// ─── Metadata ─────────────────────────────────────────────
	meta := map[string]any{
		"version":   "1.0",
		"createdAt": time.Now().UnixMilli(),
		"scope":     scope,
		"app":       "VortexUiPro",
	}
	if adminID > 0 {
		meta["adminID"] = adminID
	}
	if err := writeJSONEntry(w, "metadata.json", meta); err != nil {
		return err
	}

	// ─── Full backup: everything ──────────────────────────────
	if scope == BackupScopeFull || scope == BackupScopeSystem {
		// Admins
		admins, err := database.ListAdmins()
		if err != nil {
			log.Printf("backup: list admins: %v", err)
		}
		writeJSONEntry(w, "admins.json", admins)
		// Admin roles
		var roles []database.AdminRole
		database.DB.Find(&roles)
		writeJSONEntry(w, "admin_roles.json", roles)
		// API tokens
		var tokens []database.ApiToken
		database.DB.Find(&tokens)
		writeJSONEntry(w, "api_tokens.json", tokens)
		// Settings
		var settings []database.Setting
		database.DB.Find(&settings)
		writeJSONEntry(w, "settings.json", settings)
	}

	if scope == BackupScopeFull || scope == BackupScopeUsers {
		users, err := database.ListUsers(0)
		if err != nil {
			log.Printf("backup: list users: %v", err)
		}
		writeJSONEntry(w, "users.json", users)
		var clients []database.Client
		database.DB.Find(&clients)
		writeJSONEntry(w, "clients.json", clients)
		inbounds, err := database.ListInbounds(0, 0)
		if err != nil {
			log.Printf("backup: list inbounds: %v", err)
		}
		writeJSONEntry(w, "inbounds.json", inbounds)
		var outbounds []database.Outbound
		database.DB.Find(&outbounds)
		writeJSONEntry(w, "outbounds.json", outbounds)
		nodes, err := database.ListNodes()
		if err != nil {
			log.Printf("backup: list nodes: %v", err)
		}
		writeJSONEntry(w, "nodes.json", nodes)
		var rules []database.RoutingRule
		database.DB.Find(&rules)
		writeJSONEntry(w, "routing_rules.json", rules)
		var plans []database.Plan
		database.DB.Find(&plans)
		writeJSONEntry(w, "plans.json", plans)
		var orders []database.Order
		database.DB.Find(&orders)
		writeJSONEntry(w, "orders.json", orders)
		var tickets []database.Ticket
		database.DB.Find(&tickets)
		writeJSONEntry(w, "tickets.json", tickets)
		var transactions []database.Transaction
		database.DB.Find(&transactions)
		writeJSONEntry(w, "transactions.json", transactions)
	}

	if scope == BackupScopeReseller && adminID > 0 {
		users, err := database.ListUsers(adminID)
		if err != nil {
			log.Printf("backup: list reseller users: %v", err)
		}
		writeJSONEntry(w, "users.json", users)
		var clients []database.Client
		database.DB.Where("owner_admin_id = ?", adminID).Find(&clients)
		writeJSONEntry(w, "clients.json", clients)
		admin, err := database.GetAdminByID(adminID)
		if err != nil {
			log.Printf("backup: get reseller admin: %v", err)
		}
		if admin != nil {
			writeJSONEntry(w, "admin.json", admin)
		}
	}

	return nil
}

func writeJSONEntry(w *zip.Writer, name string, data any) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// ListBackups returns all backup entries.
func (s *BackupService) ListBackups() []BackupEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.backups == nil {
		return []BackupEntry{}
	}
	return s.backups
}

// ListResellerBackups returns backups for a specific admin.
func (s *BackupService) ListResellerBackups(adminID int64) []BackupEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []BackupEntry
	for _, b := range s.backups {
		if b.AdminID == adminID || b.Scope == BackupScopeFull {
			result = append(result, b)
		}
	}
	if result == nil {
		return []BackupEntry{}
	}
	return result
}

// DownloadBackup returns the file path for a backup by ID.
func (s *BackupService) DownloadBackup(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.backups {
		if b.ID == id {
			if _, err := os.Stat(b.Path); err == nil {
				return b.Path, nil
			}
			return "", fmt.Errorf("backup file not found")
		}
	}
	return "", fmt.Errorf("backup not found")
}

// DeleteBackup removes a backup file and entry.
func (s *BackupService) DeleteBackup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, b := range s.backups {
		if b.ID == id {
			os.Remove(b.Path)
			s.backups = append(s.backups[:i], s.backups[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("backup not found")
}

// RestoreBackup restores data from a backup archive.
func (s *BackupService) RestoreBackup(id string, scope BackupScope, adminID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var entry *BackupEntry
	for _, b := range s.backups {
		if b.ID == id {
			entry = &b
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("backup not found")
	}

	r, err := zip.OpenReader(entry.Path)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer r.Close()

	restoreMap := make(map[string][]byte)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(rc)
		restoreMap[f.Name] = data
		rc.Close()
	}

	return s.restoreFromMap(restoreMap, scope, adminID)
}

func (s *BackupService) restoreFromMap(data map[string][]byte, scope BackupScope, adminID int64) error {
	tx := database.DB.Begin()

	if scope == BackupScopeFull || scope == BackupScopeSystem {
		if raw, ok := data["admin_roles.json"]; ok {
			var roles []database.AdminRole
			json.Unmarshal(raw, &roles)
			for _, r := range roles {
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&r)
			}
		}
		if raw, ok := data["settings.json"]; ok {
			var settings []database.Setting
			json.Unmarshal(raw, &settings)
			for _, s := range settings {
				tx.Where("\"key\" = ?", s.Key).Assign(s).FirstOrCreate(&s)
			}
		}
	}

	if scope == BackupScopeFull || scope == BackupScopeUsers {
		if raw, ok := data["inbounds.json"]; ok {
			var inbounds []database.Inbound
			json.Unmarshal(raw, &inbounds)
			for _, ib := range inbounds {
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ib)
			}
		}
		if raw, ok := data["users.json"]; ok {
			var users []database.User
			json.Unmarshal(raw, &users)
			for _, u := range users {
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&u)
			}
		}
		if raw, ok := data["clients.json"]; ok {
			var clients []database.Client
			json.Unmarshal(raw, &clients)
			for _, c := range clients {
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&c)
			}
		}
	}

	if scope == BackupScopeReseller && adminID > 0 {
		if raw, ok := data["clients.json"]; ok {
			var clients []database.Client
			json.Unmarshal(raw, &clients)
			for _, c := range clients {
				c.OwnerAdminID = adminID
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&c)
			}
		}
	}

	return tx.Commit().Error
}

// SyncToRemoteStorage uploads a backup to remote storage (S3 or GDrive).
func (s *BackupService) SyncToRemoteStorage(backupID string, storageID int64) error {
	s.mu.Lock()
	var entry *BackupEntry
	for i := range s.backups {
		if s.backups[i].ID == backupID {
			entry = &s.backups[i]
			break
		}
	}
	s.mu.Unlock()

	if entry == nil {
		return fmt.Errorf("backup not found")
	}

	cfg, err := s.storageSvc.GetStorageConfig(storageID)
	if err != nil {
		return fmt.Errorf("storage config: %w", err)
	}

	fileName := filepath.Base(entry.Path)

	switch cfg.Type {
	case "s3":
		objectKey := GetS3ObjectKey(fileName)
		if err := s.storageSvc.UploadToS3(cfg, entry.Path, objectKey); err != nil {
			return err
		}
	case "gdrive":
		if err := s.storageSvc.UploadToGDrive(cfg, entry.Path, fileName); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}

	s.mu.Lock()
	entry.RemoteStorageID = storageID
	entry.RemoteStorageName = cfg.Name
	s.mu.Unlock()

	log.Printf("[Phase 14] Backup synced to remote: %s -> %s (%s)", backupID, cfg.Name, cfg.Type)
	return nil
}

// SendToTelegram sends a backup file notification via Telegram bot.
func (s *BackupService) SendToTelegram(backupID, chatID string) error {
	if s.telegramBot == nil {
		return fmt.Errorf("Telegram bot not initialized")
	}

	s.mu.Lock()
	var entry *BackupEntry
	for i := range s.backups {
		if s.backups[i].ID == backupID {
			entry = &s.backups[i]
			break
		}
	}
	s.mu.Unlock()

	if entry == nil {
		return fmt.Errorf("backup not found")
	}

	msg := formatBackupTelegramMessage(entry)
	if chatID != "" {
		s.telegramBot.SendMessage(chatID, msg)
	} else {
		s.telegramBot.SendToAll(msg)
	}

	s.mu.Lock()
	entry.TelegramSent = true
	s.mu.Unlock()

	log.Printf("[Phase 14] Backup notification sent via Telegram: %s", backupID)
	return nil
}

func formatBackupTelegramMessage(entry *BackupEntry) string {
	sizeStr := ""
	if entry.Size < 1024 {
		sizeStr = fmt.Sprintf("%d B", entry.Size)
	} else if entry.Size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(entry.Size)/1024)
	} else {
		sizeStr = fmt.Sprintf("%.1f MB", float64(entry.Size)/(1024*1024))
	}

	return fmt.Sprintf(
		"📦 <b>Backup Ready</b>\n\n"+
			"Name: <code>%s</code>\n"+
			"Scope: <b>%s</b>\n"+
			"Size: <b>%s</b>\n"+
			"Encrypted: <b>%s</b>\n"+
			"Status: <b>%s</b>\n"+
			"Date: <code>%s</code>\n\n"+
			"🔗 Download: <code>/api/v1/backups/%s/download</code>",
		entry.Name, string(entry.Scope), sizeStr,
		map[bool]string{true: "✅ Yes", false: "❌ No"}[entry.Encrypted],
		entry.Status,
		time.UnixMilli(entry.CreatedAt).Format("2006-01-02 15:04:05"),
		entry.ID,
	)
}

// CleanupOldBackups removes backups older than maxAge.
func (s *BackupService) CleanupOldBackups(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-maxAge).UnixMilli()
	var kept []BackupEntry
	for _, b := range s.backups {
		if b.CreatedAt < cutoff && strings.HasPrefix(b.Name, "auto-") {
			os.Remove(b.Path)
			continue
		}
		kept = append(kept, b)
	}
	s.backups = kept
}

func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
