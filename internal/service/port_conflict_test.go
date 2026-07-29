package service

import (
	"testing"

	"vortexuipro/internal/database"
)

func TestInboundTransports(t *testing.T) {
	tests := []struct {
		name     string
		protocol database.Protocol
		expected transportBits
	}{
		{"VMess TCP", database.ProtoVMess, transportTCP},
		{"VLESS TCP", database.ProtoVLESS, transportTCP},
		{"Trojan TCP", database.ProtoTrojan, transportTCP},
		{"Shadowsocks TCP", database.ProtoShadowsocks, transportTCP},
		{"Hysteria UDP", database.ProtoHysteria, transportUDP},
		{"Hysteria2 UDP", database.ProtoHysteria2, transportUDP},
		{"WireGuard UDP", database.ProtoWireGuard, transportUDP},
		{"MTProto TCP", database.ProtoMTProto, transportTCP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inboundTransports(tt.protocol, "", "")
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestListenOverlaps(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"0.0.0.0", "192.168.1.1", true},
		{"192.168.1.1", "0.0.0.0", true},
		{"192.168.1.1", "192.168.1.1", true},
		{"192.168.1.1", "10.0.0.1", false},
		{"", "192.168.1.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := listenOverlaps(tt.a, tt.b); got != tt.want {
				t.Errorf("listenOverlaps(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestTransportTagSuffix(t *testing.T) {
	tests := []struct {
		bits transportBits
		want string
	}{
		{transportTCP, "tcp"},
		{transportUDP, "udp"},
		{transportTCP | transportUDP, "tcpudp"},
		{0, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := transportTagSuffix(tt.bits); got != tt.want {
				t.Errorf("transportTagSuffix(%d) = %q, want %q", tt.bits, got, tt.want)
			}
		})
	}
}

func TestSameNode(t *testing.T) {
	tests := []struct {
		a, b int64
		want bool
	}{
		{1, 1, true},
		{1, 2, false},
		{0, 0, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := sameNode(tt.a, tt.b); got != tt.want {
				t.Errorf("sameNode(%d, %d) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestClientAccessScope(t *testing.T) {
	scope := DefaultClientAccessScope()
	if scope.Mode != ClientAccessAll {
		t.Errorf("expected all access mode, got %s", scope.Mode)
	}
	if !scope.AllowAllGroups {
		t.Error("expected all groups allowed")
	}
	if !scope.AllowAllInbounds {
		t.Error("expected all inbounds allowed")
	}
}

func TestNormalizeAllowedGroups(t *testing.T) {
	groups := normalizeAllowedGroups([]string{"Group1", "group1", "Group2", ""})
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d: %v", len(groups), groups)
	}
}

func TestCanClientAccessInbound(t *testing.T) {
	scope := DefaultClientAccessScope()
	inbound := &database.Inbound{ID: 1}

	// All access mode should always work
	if !CanClientAccessInbound(scope, inbound, nil) {
		t.Error("expected all access mode to allow")
	}

	// None access mode should never work
	scopeNone := ClientAccessScope{Mode: ClientAccessNone}
	if CanClientAccessInbound(scopeNone, inbound, nil) {
		t.Error("expected none access mode to deny")
	}
}

func TestGenerateInboundTag(t *testing.T) {
	var nodeID int64 = 1
	tag := GenerateInboundTag(8080, &nodeID, database.ProtoVMess, "", "")
	if tag != "n1-in-8080-tcp" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag2 := GenerateInboundTag(8080, nil, database.ProtoVMess, "", "")
	if tag2 != "in-8080-tcp" {
		t.Errorf("unexpected tag: %s", tag2)
	}
}
