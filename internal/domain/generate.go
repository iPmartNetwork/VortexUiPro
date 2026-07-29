package domain

import "encoding/json"

// GenerateXrayConfig builds a full xray JSON configuration string from domain models.
func GenerateXrayConfig(inbounds []Inbound, outbounds []Outbound, routing *RoutingConfig) (string, error) {
	obList := outbounds
	if obList == nil {
		obList = []Outbound{}
	}

	// Build outbounds list
	outboundMaps := make([]map[string]any, 0, len(obList)+2)
	for _, ob := range obList {
		if !ob.Enable {
			continue
		}
		obMap := map[string]any{
			"tag":      ob.Tag,
			"protocol": string(ob.Protocol),
			"settings": ob.Settings,
		}
		if ob.StreamSettings != "" {
			obMap["streamSettings"] = ob.StreamSettings
		}
		outboundMaps = append(outboundMaps, obMap)
	}

	// Default outbounds
	outboundMaps = append(outboundMaps,
		map[string]any{"tag": "direct", "protocol": "freedom", "settings": map[string]any{}},
		map[string]any{"tag": "block", "protocol": "blackhole", "settings": map[string]any{
			"response": map[string]any{"type": "http"},
		}},
	)

	// Build routing
	routingMap := map[string]any{
		"domainStrategy": "AsIs",
		"rules": []map[string]any{
			{"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api"},
		},
	}

	if routing != nil {
		rules := make([]map[string]any, 0)
		rules = append(rules, map[string]any{
			"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api",
		})
		for _, rule := range routing.Rules {
			if !rule.Enabled {
				continue
			}
			r := map[string]any{"type": "field"}
			if len(rule.Domain) > 0 {
				r["domain"] = rule.Domain
			}
			if len(rule.IP) > 0 {
				r["ip"] = rule.IP
			}
			if rule.OutboundTag != "" {
				r["outboundTag"] = rule.OutboundTag
			}
			rules = append(rules, r)
		}
		routingMap = map[string]any{
			"domainStrategy": routing.DomainStrategy,
			"rules":          rules,
		}
	}

	// Stats config
	stats := map[string]any{
		"enabled":  true,
		"interval": "5s",
	}

	// Build full config
	config := map[string]any{
		"log":       map[string]any{"loglevel": "warning"},
		"inbounds":  buildConfigInbounds(inbounds),
		"outbounds": outboundMaps,
		"routing":   routingMap,
		"dns": map[string]any{
			"servers": []string{"https://1.1.1.1/dns-query", "https://8.8.8.8/dns-query"},
		},
		"stats": stats,
		"api": map[string]any{
			"tag":      "api",
			"services": []string{"HandlerService", "StatsService"},
		},
		"policy": map[string]any{
			"levels": map[string]any{
				"8": map[string]any{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]any{
				"statsInboundUplink":   true,
				"statsInboundDownlink": true,
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildConfigInbounds(inbounds []Inbound) []map[string]any {
	out := make([]map[string]any, 0, len(inbounds))
	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}
		inbound := map[string]any{
			"tag":      ib.Tag,
			"port":     ib.Port,
			"listen":   ib.Listen,
			"protocol": string(ib.Protocol),
			"settings": ib.Settings,
		}
		if ib.StreamSettings != "" {
			inbound["streamSettings"] = ib.StreamSettings
		}
		if ib.Sniffing != "" {
			inbound["sniffing"] = ib.Sniffing
		}
		out = append(out, inbound)
	}
	return out
}
