package service

import (
	"context"
	"fmt"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// SubProfileService manages multi-profile subscription endpoints and format generation.
type SubProfileService struct {
	eventBus events.Publisher
}

// NewSubProfileService creates a new subscription profile service.
func NewSubProfileService(bus events.Publisher) *SubProfileService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &SubProfileService{eventBus: bus}
}

// ─── Profile CRUD ───────────────────────────────────────────────────

// CreateProfile adds a new subscription profile endpoint.
func (s *SubProfileService) CreateProfile(ctx context.Context, p *database.SubscriptionProfile) (*database.SubscriptionProfile, error) {
	p.CreatedAt = time.Now().UnixMilli()
	if err := database.DB.Create(p).Error; err != nil {
		return nil, fmt.Errorf("create sub profile: %w", err)
	}
	return p, nil
}

// GetProfile retrieves a profile by ID.
func (s *SubProfileService) GetProfile(id int64) (*database.SubscriptionProfile, error) {
	var p database.SubscriptionProfile
	if err := database.DB.First(&p, id).Error; err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}
	return &p, nil
}

// ListProfiles returns all profiles for a given inbound.
func (s *SubProfileService) ListProfiles(inboundID int64) ([]database.SubscriptionProfile, error) {
	var list []database.SubscriptionProfile
	q := database.DB.Model(&database.SubscriptionProfile{})
	if inboundID > 0 {
		q = q.Where("inbound_id = ?", inboundID)
	}
	return list, q.Order("id asc").Find(&list).Error
}

// UpdateProfile modifies a subscription profile.
func (s *SubProfileService) UpdateProfile(ctx context.Context, p *database.SubscriptionProfile) error {
	return database.DB.Save(p).Error
}

// DeleteProfile removes a profile.
func (s *SubProfileService) DeleteProfile(ctx context.Context, id int64) error {
	return database.DB.Delete(&database.SubscriptionProfile{}, id).Error
}

// ─── Subscription Hosts ─────────────────────────────────────────────

// CreateHost adds a custom subscription host.
func (s *SubProfileService) CreateHost(ctx context.Context, h *database.SubscriptionHost) (*database.SubscriptionHost, error) {
	if err := database.DB.Create(h).Error; err != nil {
		return nil, fmt.Errorf("create sub host: %w", err)
	}
	return h, nil
}

// ListHosts returns all subscription hosts.
func (s *SubProfileService) ListHosts() ([]database.SubscriptionHost, error) {
	var list []database.SubscriptionHost
	return list, database.DB.Order("domain asc").Find(&list).Error
}

// DeleteHost removes a subscription host.
func (s *SubProfileService) DeleteHost(ctx context.Context, id int64) error {
	return database.DB.Delete(&database.SubscriptionHost{}, id).Error
}

// ─── Format Information ─────────────────────────────────────────────

// FormatVariant describes a subscription output format.
type FormatVariant struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Mime        string `json:"mime"`
}

// ListFormats returns all supported subscription formats.
func (s *SubProfileService) ListFormats() []FormatVariant {
	return []FormatVariant{
		{Name: "xray", Description: "Xray JSON config", Mime: "application/json"},
		{Name: "clash", Description: "Clash / Mihomo YAML config", Mime: "application/yaml"},
		{Name: "singbox", Description: "Sing-box JSON config", Mime: "application/json"},
		{Name: "outline", Description: "Outline JSON config", Mime: "application/json"},
		{Name: "v2ray", Description: "V2Ray JSON compatible", Mime: "application/json"},
		{Name: "vless", Description: "VLESS share link", Mime: "text/plain"},
		{Name: "vmess", Description: "VMess share link", Mime: "text/plain"},
	}
}

// ─── Remark Variables ───────────────────────────────────────────────

// RemarkVar defines a variable that can be used in subscription remarks.
type RemarkVar struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// ListRemarkVars returns all available remark template variables.
func (s *SubProfileService) ListRemarkVars() []RemarkVar {
	return []RemarkVar{
		{Name: "{client_email}", Description: "Client email address", Example: "user@example.com"},
		{Name: "{client_id}", Description: "Client UUID / ID", Example: "a1b2c3d4-..."},
		{Name: "{inbound_remark}", Description: "Inbound remark name", Example: "🇺🇸 US-01"},
		{Name: "{inbound_tag}", Description: "Inbound tag", Example: "inbound-us-01"},
		{Name: "{inbound_protocol}", Description: "Inbound protocol", Example: "vless"},
		{Name: "{inbound_port}", Description: "Inbound port", Example: "443"},
		{Name: "{inbound_host}", Description: "Inbound listen address", Example: "1.2.3.4"},
		{Name: "{user_traffic_up}", Description: "User upload traffic", Example: "1.5 GB"},
		{Name: "{user_traffic_down}", Description: "User download traffic", Example: "3.2 GB"},
		{Name: "{user_data_limit}", Description: "User data limit", Example: "100 GB"},
		{Name: "{user_expiry}", Description: "User expiry date", Example: "2025-12-31"},
		{Name: "{user_status}", Description: "User account status", Example: "active"},
		{Name: "{node_name}", Description: "Node name", Example: "Frankfurt-1"},
		{Name: "{node_location}", Description: "Node location", Example: "Frankfurt, DE"},
		{Name: "{subscription_url}", Description: "Full subscription URL", Example: "https://..."},
		{Name: "{random_email}", Description: "Random email generator", Example: "abc123@x.com"},
		{Name: "{random_uuid}", Description: "Random UUID v4", Example: "550e8400-..."},
	}
}
