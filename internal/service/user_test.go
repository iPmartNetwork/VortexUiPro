package service

import (
	"testing"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

func TestUserService_CRUD(t *testing.T) {
	if err := database.InitDB(database.Config{
		Type: "sqlite", DSN: ":memory:", LogLevel: "silent",
	}); err != nil {
		t.Fatalf("init db: %v", err)
	}

	bus := events.Nop{}
	svc := NewUserService(bus)

	t.Run("create user", func(t *testing.T) {
		u, err := svc.CreateUser("testuser", "test@example.com", 0, 0)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		if u.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
		if u.Username != "testuser" {
			t.Fatalf("expected testuser, got %s", u.Username)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		_, err := svc.CreateUser("testuser", "test@example.com", 0, 0)
		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
	})

	t.Run("list users", func(t *testing.T) {
		users, err := svc.ListUsers(0)
		if err != nil {
			t.Fatalf("list users: %v", err)
		}
		if len(users) < 1 {
			t.Fatal("expected at least 1 user")
		}
	})

	t.Run("delete user", func(t *testing.T) {
		users, _ := svc.ListUsers(0)
		if len(users) > 0 {
			err := svc.DeleteUser(users[0].ID)
			if err != nil {
				t.Fatalf("delete user: %v", err)
			}
		}
	})
}
