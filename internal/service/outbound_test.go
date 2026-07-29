package service

import (
	"context"
	"testing"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

func TestOutboundService_CRUD(t *testing.T) {
	if err := database.InitDB(database.Config{
		Type: "sqlite", DSN: ":memory:", LogLevel: "silent",
	}); err != nil {
		t.Fatalf("init db: %v", err)
	}

	bus := events.Nop{}
	svc := NewOutboundService(bus)
	ctx := context.Background()

	t.Run("create outbound", func(t *testing.T) {
		ob, err := svc.Create(ctx, &database.Outbound{
			Tag:      "proxy-out",
			Protocol: "freedom",
			Remark:   "Direct outbound",
			Enable:   true,
		})
		if err != nil {
			t.Fatalf("create outbound: %v", err)
		}
		if ob.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
	})

	t.Run("duplicate tag", func(t *testing.T) {
		_, err := svc.Create(ctx, &database.Outbound{
			Tag:      "proxy-out",
			Protocol: "freedom",
		})
		if err == nil {
			t.Fatal("expected error for duplicate tag")
		}
	})

	t.Run("list outbounds", func(t *testing.T) {
		list, err := svc.List(0)
		if err != nil {
			t.Fatalf("list outbounds: %v", err)
		}
		if len(list) < 1 {
			t.Fatal("expected at least 1 outbound")
		}
	})

	t.Run("get by tag", func(t *testing.T) {
		ob, err := svc.GetByTag("proxy-out")
		if err != nil {
			t.Fatalf("get by tag: %v", err)
		}
		if ob.Tag != "proxy-out" {
			t.Fatalf("expected proxy-out, got %s", ob.Tag)
		}
	})

	t.Run("delete outbound", func(t *testing.T) {
		list, _ := svc.List(0)
		if len(list) > 0 {
			err := svc.Delete(ctx, list[0].ID)
			if err != nil {
				t.Fatalf("delete outbound: %v", err)
			}
		}
	})
}
