package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"vortexuipro/internal/database"
)

// ClientAccessMode defines how a client's access is restricted.
type ClientAccessMode string

const (
	ClientAccessAll  ClientAccessMode = "all"
	ClientAccessOwn  ClientAccessMode = "own"
	ClientAccessNone ClientAccessMode = "none"
)

// ClientAccessScope defines the access scope for a client.
type ClientAccessScope struct {
	AdminID          int64             `json:"-"`
	Mode             ClientAccessMode  `json:"mode"`
	RestrictGroups   bool              `json:"restrictGroups"`
	AllowAllGroups   bool              `json:"allowAllGroups"`
	AllowedGroups    []string          `json:"allowedGroups,omitempty"`
	RestrictInbounds bool              `json:"restrictInbounds"`
	AllowAllInbounds bool              `json:"allowAllInbounds"`
	AllowedInboundIDs []int64          `json:"allowedInboundIDs,omitempty"`
}

// DefaultClientAccessScope returns a scope with full access.
func DefaultClientAccessScope() ClientAccessScope {
	return ClientAccessScope{
		Mode:             ClientAccessAll,
		AllowAllGroups:   true,
		AllowAllInbounds: true,
	}
}

// ParseClientAccessScope parses a JSON string into ClientAccessScope.
func ParseClientAccessScope(data string) (ClientAccessScope, error) {
	var scope ClientAccessScope
	if data == "" {
		return DefaultClientAccessScope(), nil
	}
	if err := json.Unmarshal([]byte(data), &scope); err != nil {
		return DefaultClientAccessScope(), fmt.Errorf("parse access scope: %w", err)
	}
	return normalizeClientAccessScope(scope), nil
}

func normalizeClientAccessScope(scope ClientAccessScope) ClientAccessScope {
	if scope.Mode == "" {
		scope.Mode = ClientAccessAll
	}
	scope.AllowedGroups = normalizeAllowedGroups(scope.AllowedGroups)
	scope.AllowedInboundIDs = normalizeAllowedInboundIDs(scope.AllowedInboundIDs)
	return scope
}

func normalizeAllowedGroups(groups []string) []string {
	if len(groups) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		key := strings.ToLower(strings.TrimSpace(group))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func normalizeAllowedInboundIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// CanClientAccessInbound checks if a client can access a specific inbound.
func CanClientAccessInbound(scope ClientAccessScope, inbound *database.Inbound, clientGroups []string) bool {
	switch scope.Mode {
	case ClientAccessNone:
		return false
	case ClientAccessOwn:
		// Client can only access inbounds they own
		return true // caller must check ownership separately
	case ClientAccessAll:
		// Check group restriction
		if scope.RestrictGroups && !scope.AllowAllGroups && len(clientGroups) > 0 {
			allowed := false
			for _, cg := range clientGroups {
				for _, ag := range scope.AllowedGroups {
					if strings.EqualFold(cg, ag) {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			}
			if !allowed {
				return false
			}
		}

		// Check inbound restriction
		if scope.RestrictInbounds && !scope.AllowAllInbounds {
			allowed := false
			for _, aid := range scope.AllowedInboundIDs {
				if inbound.ID == aid {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}

		return true
	}
	return true
}

// FilterClientsByScope filters a list of clients by the access scope.
func FilterClientsByScope(clients []database.Client, scope ClientAccessScope, inboundMap map[int64]*database.Inbound, clientGroupMap map[string][]string) []database.Client {
	if scope.Mode == ClientAccessAll && scope.AllowAllInbounds && scope.AllowAllGroups {
		return clients
	}

	filtered := make([]database.Client, 0, len(clients))
	for _, c := range clients {
		inbound, ok := inboundMap[c.InboundID]
		if !ok {
			continue
		}
		groups := clientGroupMap[c.ID]
		if CanClientAccessInbound(scope, inbound, groups) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
