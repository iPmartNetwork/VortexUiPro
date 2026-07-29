package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// RoutingService manages routing rules and configuration.
type RoutingService struct {
	eventBus events.Publisher
}

// NewRoutingService creates a new routing service.
func NewRoutingService(bus events.Publisher) *RoutingService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &RoutingService{eventBus: bus}
}

// ─── Routing Rules ─────────────────────────────────────────────────

// CreateRule adds a new routing rule.
func (s *RoutingService) CreateRule(ctx context.Context, rule *database.RoutingRule) (*database.RoutingRule, error) {
	rule.CreatedAt = time.Now().UnixMilli()
	if err := database.DB.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("create routing rule: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "routing.rule.created",
		Data: map[string]any{"id": rule.ID, "outbound_tag": rule.OutboundTag},
	})
	return rule, nil
}

// GetRule retrieves a routing rule by ID.
func (s *RoutingService) GetRule(id int64) (*database.RoutingRule, error) {
	var rule database.RoutingRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("routing rule not found: %w", err)
	}
	return &rule, nil
}

// ListRules returns all routing rules.
func (s *RoutingService) ListRules() ([]database.RoutingRule, error) {
	var list []database.RoutingRule
	return list, database.DB.Order("id asc").Find(&list).Error
}

// UpdateRule modifies a routing rule.
func (s *RoutingService) UpdateRule(ctx context.Context, rule *database.RoutingRule) error {
	if err := database.DB.Save(rule).Error; err != nil {
		return fmt.Errorf("update routing rule: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "routing.rule.updated",
		Data: map[string]any{"id": rule.ID},
	})
	return nil
}

// DeleteRule removes a routing rule.
func (s *RoutingService) DeleteRule(ctx context.Context, id int64) error {
	if err := database.DB.Delete(&database.RoutingRule{}, id).Error; err != nil {
		return fmt.Errorf("delete routing rule: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "routing.rule.deleted",
		Data: map[string]any{"id": id},
	})
	return nil
}

// ToggleRule enables or disables a routing rule.
func (s *RoutingService) ToggleRule(ctx context.Context, id int64, enable bool) error {
	return database.DB.Model(&database.RoutingRule{}).Where("id = ?", id).
		Update("enable", enable).Error
}

// ─── Routing Packs (groups of rules for quick switching) ───────────

// ListPacks returns all routing packs.
func (s *RoutingService) ListPacks() ([]database.RoutingPack, error) {
	var packs []database.RoutingPack
	return packs, database.DB.Order("name asc").Find(&packs).Error
}

// GetPack retrieves a routing pack by ID.
func (s *RoutingService) GetPack(id int64) (*database.RoutingPack, error) {
	var pack database.RoutingPack
	if err := database.DB.First(&pack, id).Error; err != nil {
		return nil, fmt.Errorf("routing pack not found: %w", err)
	}
	return &pack, nil
}

// CreatePack creates a new routing pack.
func (s *RoutingService) CreatePack(ctx context.Context, pack *database.RoutingPack) (*database.RoutingPack, error) {
	pack.CreatedAt = time.Now().UnixMilli()
	if err := database.DB.Create(pack).Error; err != nil {
		return nil, fmt.Errorf("create routing pack: %w", err)
	}
	return pack, nil
}

// UpdatePack modifies a routing pack.
func (s *RoutingService) UpdatePack(ctx context.Context, pack *database.RoutingPack) error {
	return database.DB.Save(pack).Error
}

// DeletePack removes a routing pack.
func (s *RoutingService) DeletePack(ctx context.Context, id int64) error {
	return database.DB.Delete(&database.RoutingPack{}, id).Error
}

// ─── Xray Routing Config Generation ─────────────────────────────────

// GenerateXrayConfig produces the full Xray routing configuration from all enabled rules.
func (s *RoutingService) GenerateXrayConfig() (string, error) {
	rules, err := s.ListRules()
	if err != nil {
		return "", err
	}

	var enabledRules []map[string]any
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		rule := map[string]any{
			"outboundTag": r.OutboundTag,
		}
		if r.Domain != "" {
			var domains []string
			json.Unmarshal([]byte(r.Domain), &domains)
			rule["domain"] = domains
		}
		if r.IP != "" {
			var ips []string
			json.Unmarshal([]byte(r.IP), &ips)
			rule["ip"] = ips
		}
		if r.Port != "" {
			rule["port"] = r.Port
		}
		if r.Network != "" {
			rule["network"] = r.Network
		}
		if r.Protocol != "" {
			var protos []string
			json.Unmarshal([]byte(r.Protocol), &protos)
			rule["protocol"] = protos
		}
		if r.InboundTags != "" {
			var tags []string
			json.Unmarshal([]byte(r.InboundTags), &tags)
			rule["inboundTag"] = tags
		}
		if r.SourceIP != "" {
			var srcIPs []string
			json.Unmarshal([]byte(r.SourceIP), &srcIPs)
			rule["source"] = srcIPs
		}
		if r.BalancerTag != "" {
			rule["balancerTag"] = r.BalancerTag
		}
		enabledRules = append(enabledRules, rule)
	}

	config := map[string]any{
		"domainStrategy": "AsIs",
		"rules":         enabledRules,
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data), nil
}
