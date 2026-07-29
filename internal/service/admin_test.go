package service

import (
	"testing"

	"vortexuipro/internal/auth"
	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

func TestAdminService_Login(t *testing.T) {
	// Initialize in-memory SQLite
	if err := database.InitDB(database.Config{
		Type:     "sqlite",
		DSN:      ":memory:",
		LogLevel: "silent",
	}); err != nil {
		t.Fatalf("init db: %v", err)
	}

	bus := events.Nop{}
	svc := NewAdminService("test-secret", bus)

	// Seed admin
	hash, _ := auth.HashPassword("admin123")
	database.CreateAdmin(&database.Admin{Username: "admin", PasswordHash: hash, Role: "super_admin"})

	t.Run("successful login", func(t *testing.T) {
		tokens, err := svc.Login("admin", "admin123")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if tokens.AccessToken == "" {
			t.Fatal("expected non-empty access token")
		}
		if tokens.RefreshToken == "" {
			t.Fatal("expected non-empty refresh token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login("admin", "wrong")
		if err == nil {
			t.Fatal("expected error for wrong password")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := svc.Login("nonexistent", "pass")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})
}
