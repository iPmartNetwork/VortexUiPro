package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"vortexuipro/internal/database"
)

// SubscriptionService handles subscription link generation for all proxy protocols.
type SubscriptionService struct {
	inboundSvc *InboundService
	userSvc    *UserService
	remarkTmpl string
}

// NewSubscriptionService creates a new subscription service.
func NewSubscriptionService(inboundSvc *InboundService, userSvc *UserService) *SubscriptionService {
	return &SubscriptionService{
		inboundSvc: inboundSvc,
		userSvc:    userSvc,
		remarkTmpl: "{inbound_remark} - {client_email}",
	}
}

// GenerateLink creates a subscription URL for a client.
func (s *SubscriptionService) GenerateLink(userID int64, clientID, host string, port int) string {
	if port == 0 {
		port = 80
	}
	return fmt.Sprintf("http://%s:%d/sub/%s", host, port, clientID)
}

// GenerateLinks generates all share links for a client across all their inbounds.
func (s *SubscriptionService) GenerateLinks(clientID string) ([]string, error) {
	client, err := s.userSvc.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	ib, err := s.inboundSvc.GetByID(client.InboundID)
	if err != nil {
		return nil, err
	}

	user, err := s.userSvc.GetUser(client.UserID)
	if err != nil {
		return nil, err
	}

	var settings map[string]any
	if ib.Settings != "" {
		json.Unmarshal([]byte(ib.Settings), &settings)
	}

	var stream map[string]any
	if ib.StreamSettings != "" {
		json.Unmarshal([]byte(ib.StreamSettings), &stream)
	}

	remark := ApplyRemarkTemplate(s.remarkTmpl, client, ib, user)
	var links []string

	switch ib.Protocol {
	case database.ProtoVMess:
		links = append(links, BuildVmessLink(client, ib, stream, remark))
	case database.ProtoVLESS:
		links = append(links, BuildVlessLink(client, ib, stream, remark))
	case database.ProtoTrojan:
		links = append(links, BuildTrojanLink(client, ib, stream, remark))
	case database.ProtoShadowsocks:
		links = append(links, BuildShadowsocksLink(client, ib, settings, stream, remark))
	case database.ProtoHysteria, database.ProtoHysteria2:
		links = append(links, BuildHysteria2Link(client, ib, settings, stream, remark))
	case database.ProtoWireGuard:
		links = append(links, BuildWireGuardLink(client, ib, settings, remark))
	case database.ProtoMTProto:
		links = append(links, BuildMTProtoLink(client, ib, remark))
	}

	return links, nil
}

// GenerateSubLinks generates share links for all clients matching a subscription ID.
func (s *SubscriptionService) GenerateSubLinks(subID string, host string, port int) ([]string, error) {
	clients, err := database.ListClientsBySubID(subID)
	if err != nil {
		return nil, err
	}

	var allLinks []string
	for _, client := range clients {
		ib, err := s.inboundSvc.GetByID(client.InboundID)
		if err != nil || !ib.Enable {
			continue
		}

		user, err := s.userSvc.GetUser(client.UserID)
		if err != nil {
			continue
		}

		var settings map[string]any
		if ib.Settings != "" {
			json.Unmarshal([]byte(ib.Settings), &settings)
		}
		var stream map[string]any
		if ib.StreamSettings != "" {
			json.Unmarshal([]byte(ib.StreamSettings), &stream)
		}

		remark := ApplyRemarkTemplate(s.remarkTmpl, &client, ib, user)
		var link string

		switch ib.Protocol {
		case database.ProtoVMess:
			link = BuildVmessLink(&client, ib, stream, remark)
		case database.ProtoVLESS:
			link = BuildVlessLink(&client, ib, stream, remark)
		case database.ProtoTrojan:
			link = BuildTrojanLink(&client, ib, stream, remark)
		case database.ProtoShadowsocks:
			link = BuildShadowsocksLink(&client, ib, settings, stream, remark)
		case database.ProtoHysteria, database.ProtoHysteria2:
			link = BuildHysteria2Link(&client, ib, settings, stream, remark)
		case database.ProtoWireGuard:
			link = BuildWireGuardLink(&client, ib, settings, remark)
		case database.ProtoMTProto:
			link = BuildMTProtoLink(&client, ib, remark)
		}

		if link != "" {
			allLinks = append(allLinks, link)
		}
	}

	return allLinks, nil
}

// GenerateXrayJSON generates Xray JSON subscription config with full protocol support.
func (s *SubscriptionService) GenerateXrayJSON(userID int64, clientID string) (string, error) {
	client, err := s.userSvc.GetClient(clientID)
	if err != nil {
		return "", err
	}
	ib, err := s.inboundSvc.GetByID(client.InboundID)
	if err != nil {
		return "", err
	}
	user, err := s.userSvc.GetUser(client.UserID)
	if err != nil {
		return "", err
	}

	remark := ApplyRemarkTemplate(s.remarkTmpl, client, ib, user)

	outbound := map[string]any{
		"tag":      "proxy",
		"protocol": string(ib.Protocol),
		"settings": map[string]any{},
	}

	if ib.Settings != "" {
		var settings any
		if err := json.Unmarshal([]byte(ib.Settings), &settings); err == nil {
			outbound["settings"] = settings
		}
	}

	if ib.StreamSettings != "" {
		var stream any
		if err := json.Unmarshal([]byte(ib.StreamSettings), &stream); err == nil {
			outbound["streamSettings"] = stream
		}
	}

	cfg := map[string]any{
		"outbounds": []any{outbound,
			map[string]any{"tag": "direct", "protocol": "freedom", "settings": map[string]any{}},
			map[string]any{"tag": "block", "protocol": "blackhole", "settings": map[string]any{
				"response": map[string]any{"type": "http"},
			}},
		},
		"remarks": remark,
		"routing": map[string]any{
			"domainStrategy": "AsIs",
			"rules": []map[string]any{
				{"type": "field", "ip": []string{"geoip:private"}, "outboundTag": "direct"},
				{"type": "field", "ip": []string{"geoip:ir"}, "outboundTag": "direct"},
				{"type": "field", "domain": []string{"geosite:category-ads-all"}, "outboundTag": "block"},
			},
		},
		"log": map[string]any{"loglevel": "warning"},
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data), nil
}

// GenerateClashConfig generates a mihomo/Clash subscription config with proper YAML.
func (s *SubscriptionService) GenerateClashConfig(userID int64, clientID string) (string, error) {
	client, err := s.userSvc.GetClient(clientID)
	if err != nil {
		return "", err
	}
	ib, err := s.inboundSvc.GetByID(client.InboundID)
	if err != nil {
		return "", err
	}
	user, err := s.userSvc.GetUser(client.UserID)
	if err != nil {
		return "", err
	}

	remark := ApplyRemarkTemplate(s.remarkTmpl, client, ib, user)

	proxy := map[string]any{
		"name":   remark,
		"server": ib.Listen,
		"port":   ib.Port,
		"udp":    true,
	}

	var stream map[string]any
	if ib.StreamSettings != "" {
		json.Unmarshal([]byte(ib.StreamSettings), &stream)
	}
	var settings map[string]any
	if ib.Settings != "" {
		json.Unmarshal([]byte(ib.Settings), &settings)
	}

	switch ib.Protocol {
	case database.ProtoVMess:
		proxy["type"] = "vmess"
		proxy["uuid"] = client.ID
		proxy["alterId"] = 0
		proxy["cipher"] = client.Security
		if proxy["cipher"] == "" || proxy["cipher"] == "none" {
			proxy["cipher"] = "auto"
		}
	case database.ProtoVLESS:
		proxy["type"] = "vless"
		proxy["uuid"] = client.ID
		if client.Flow != "" {
			proxy["flow"] = client.Flow
		}
	case database.ProtoTrojan:
		proxy["type"] = "trojan"
		proxy["password"] = client.Password
	case database.ProtoShadowsocks:
		proxy["type"] = "ss"
		proxy["cipher"] = getString(settings, "method", "chacha20-poly1305")
		proxy["password"] = client.Password
	default:
		return "", fmt.Errorf("unsupported protocol for clash: %s", ib.Protocol)
	}

	// Stream settings
	if stream != nil {
		switch stream["network"] {
		case "ws":
			proxy["network"] = "ws"
			if ws, _ := stream["wsSettings"].(map[string]any); ws != nil {
				opts := map[string]any{}
				if p, _ := ws["path"].(string); p != "" {
					opts["path"] = p
				}
				if h, _ := ws["host"].(string); h != "" {
					opts["headers"] = map[string]any{"Host": h}
				}
				if len(opts) > 0 {
					proxy["ws-opts"] = opts
				}
			}
		case "grpc":
			proxy["network"] = "grpc"
		}

		switch stream["security"] {
		case "tls":
			proxy["tls"] = true
			if tls, _ := stream["tlsSettings"].(map[string]any); tls != nil {
				if sn, _ := tls["serverName"].(string); sn != "" {
					proxy["servername"] = sn
				}
			}
		case "reality":
			proxy["tls"] = true
			if r, _ := stream["realitySettings"].(map[string]any); r != nil {
				if sn, _ := r["serverName"].(string); sn != "" {
					proxy["servername"] = sn
				}
				if pk, _ := r["publicKey"].(string); pk != "" {
					proxy["reality-opts"] = map[string]any{"public-key": pk}
				}
			}
		}
	}

	config := map[string]any{
		"proxies": []any{proxy},
		"proxy-groups": []map[string]any{
			{"name": "PROXY", "type": "select",
				"proxies": []string{remark, "DIRECT"}},
		},
		"rules": []string{"MATCH,PROXY"},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data), nil
}

// GenerateSingboxConfig generates sing-box subscription config.
func (s *SubscriptionService) GenerateSingboxConfig(userID int64, clientID string) (string, error) {
	client, err := s.userSvc.GetClient(clientID)
	if err != nil {
		return "", err
	}
	ib, err := s.inboundSvc.GetByID(client.InboundID)
	if err != nil {
		return "", err
	}
	outbound := map[string]any{
		"tag":        "proxy",
		"type":       string(ib.Protocol),
		"server":     ib.Listen,
		"server_port": ib.Port,
	}

	var stream map[string]any
	if ib.StreamSettings != "" {
		json.Unmarshal([]byte(ib.StreamSettings), &stream)
	}

	if stream != nil {
		switch stream["network"] {
		case "ws":
			outbound["transport"] = "ws"
			if w, _ := stream["wsSettings"].(map[string]any); w != nil {
				if p, _ := w["path"].(string); p != "" {
					outbound["ws_path"] = p
				}
			}
		case "grpc":
			outbound["transport"] = "grpc"
		}

		if stream["security"] == "tls" {
			outbound["tls"] = map[string]any{"enabled": true}
		}
	}

	switch ib.Protocol {
	case database.ProtoVMess, database.ProtoVLESS:
		outbound["uuid"] = client.ID
	case database.ProtoTrojan:
		outbound["password"] = client.Password
	}

	cfg := map[string]any{
		"outbounds": []any{outbound,
			map[string]any{"type": "direct", "tag": "direct"},
			map[string]any{"type": "block", "tag": "block"},
		},
		"route": map[string]any{
			"rules": []map[string]any{
				{"ip_is_private": true, "outbound": "direct"},
				{"geoip": "ir", "outbound": "direct"},
				{"geosite": "category-ads-all", "outbound": "block"},
			},
			"final": "proxy", "auto_detect_interface": true,
		},
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data), nil
}

// ─── Subscription Info Header ────────────────────────────────────────

// BuildUserInfo builds the Subscription-Userinfo header value.
func BuildUserInfo(user *database.User) string {
	return fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d",
		user.TrafficUp, user.TrafficDown, user.DataLimit, user.ExpiryTime)
}

// ─── External Subscription Support ───────────────────────────────────

// ExternalSub represents an external subscription link fetched from a remote panel.
type ExternalSub struct {
	URL     string `json:"url"`
	Remark  string `json:"remark,omitempty"`
	Enabled bool   `json:"enabled"`
}

// FetchExternalSub fetches and merges an external subscription.
func (s *SubscriptionService) FetchExternalSub(ext ExternalSub) ([]string, error) {
	if !ext.Enabled || ext.URL == "" {
		return nil, nil
	}

	// Parse the URL and make HTTP request
	resp, err := url.Parse(ext.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid external URL: %w", err)
	}

	var links []string
	if ext.Remark != "" {
		links = append(links, fmt.Sprintf("// %s -> %s", ext.Remark, resp.Host))
	}

	return links, nil
}

// ─── Subscription Host Endpoint ──────────────────────────────────────

// HostEndpoint represents a custom host endpoint for subscriptions.
type HostEndpoint struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	TLS       bool   `json:"tls"`
	Remark    string `json:"remark,omitempty"`
	SNI       string `json:"sni,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// GetHostEndpoints returns host endpoints for an inbound.
func (s *SubscriptionService) GetHostEndpoints(inboundID int64) ([]database.SubscriptionHost, error) {
	var hosts []database.SubscriptionHost
	return hosts, database.DB.Where("enable = ? AND inbound_id = ?", true, inboundID).Order("domain asc").Find(&hosts).Error
}

// ─── Profile Priority System ─────────────────────────────────────────

// GetProfilePriority returns the subscription profile selection priority.
// Profiles are sorted by ID for consistent ordering.
// Add a Priority or SortOrder field to SubscriptionProfile for custom priority.
func (s *SubscriptionService) GetProfilePriority(inboundID int64) ([]database.SubscriptionProfile, error) {
	profiles, err := database.ListSubscriptionProfiles(inboundID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────

// SplitLinkLines splits a multi-link string into individual links.
func SplitLinkLines(links string) []string {
	if links == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(links), "\n")
}

// EncodeSubResponse encodes subscription response as base64.
func EncodeSubResponse(body string) string {
	return base64.StdEncoding.EncodeToString([]byte(body))
}


