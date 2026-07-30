package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // pure-Go SQLite driver (CGO-free)
)

var DB *gorm.DB

// Config holds database connection parameters.
type Config struct {
	Type     string // sqlite or postgres
	DSN      string // file path or connection string
	LogLevel string
}

// InitDB opens the database, runs auto-migration, and seeds defaults.
func InitDB(cfg Config) error {
	if cfg.Type == "" {
		cfg.Type = "sqlite"
	}
	if cfg.DSN == "" {
		cfg.DSN = "/etc/vortex/data/vortex.db"
	}

	// Ensure directory exists for file-based databases
	if cfg.Type == "sqlite" {
		dir := filepath.Dir(cfg.DSN)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create db dir %s: %w", dir, err)
		}
	}

	logLevel := logger.Warn
	switch cfg.LogLevel {
	case "info":
		logLevel = logger.Info
	case "silent":
		logLevel = logger.Silent
	}

	var dialector gorm.Dialector
	switch cfg.Type {
	case "sqlite":
		// Use modernc pure-Go driver — registered as "sqlite" (CGO-free)
		dialector = gormsqlite.Open(cfg.DSN + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	// Connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate all models
	if err := DB.AutoMigrate(
		&Admin{},
		&AdminRole{},
		&ApiToken{},
		&User{},
		&Inbound{},
		&Outbound{},
		&Client{},
		&Node{},
		&Setting{},
		&RoutingRule{},
		&SubscriptionHost{},
		&SubscriptionProfile{},
		&NotificationChannel{},
		&SecurityEvent{},
		&Ticket{},
		&TicketReply{},
		&Plan{},
		&Order{},
		&Transaction{},
		&ClusterNode{},
		&SyncEvent{},
		&ConfigVersion{},
		&AuditEntry{},
		&ClientGroup{},
		&ClientGroupMember{},
		&RoutingPack{},
		&FederationProvider{},
		&TURNServer{},
		&P2PMeshConfig{},
		&HealthCheckConfig{},
		&HealthCheckResult{},
		&AutoRecoveryRule{},
		&AutoRecoveryAction{},
		&BackupEncryptionKey{},
		&RemoteStorageConfig{},
		&CDNDomain{},
		&DNSConfig{},
		&DNSRule{},
		&DockerContainer{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	log.Printf("Database initialized (%s): %s", cfg.Type, cfg.DSN)
	return nil
}

// CloseDB closes the database connection.
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
