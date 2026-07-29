package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"vortexuipro/internal/database"
)

type transportBits uint8

const (
	transportTCP transportBits = 1 << iota
	transportUDP
)

func inboundTransports(protocol database.Protocol, streamSettings, settings string) transportBits {
	switch protocol {
	case database.ProtoHysteria, database.ProtoHysteria2, database.ProtoWireGuard:
		return transportUDP
	case database.ProtoMTProto:
		return transportTCP
	}

	var bits transportBits
	network := ""
	if streamSettings != "" {
		var ss map[string]any
		if json.Unmarshal([]byte(streamSettings), &ss) == nil {
			if n, _ := ss["network"].(string); n != "" {
				network = n
			}
		}
	}
	switch network {
	case "kcp", "quic":
		bits |= transportUDP
	default:
		bits |= transportTCP
	}

	if settings != "" {
		var st map[string]any
		if json.Unmarshal([]byte(settings), &st) == nil {
			switch protocol {
			case database.ProtoShadowsocks, database.ProtoTunnel:
				key := "network"
				if protocol == database.ProtoTunnel {
					key = "allowedNetwork"
				}
				if n, ok := st[key].(string); ok && n != "" {
					bits = 0
					for _, part := range strings.Split(n, ",") {
						switch strings.TrimSpace(part) {
						case "tcp":
							bits |= transportTCP
						case "udp":
							bits |= transportUDP
						}
					}
				}
			case database.ProtoMixed:
				if udpOn, _ := st["udp"].(bool); udpOn {
					bits |= transportUDP
				}
			}
		}
	}

	if bits == 0 {
		bits = transportTCP
	}
	return bits
}

func listenOverlaps(a, b string) bool {
	if a == "" || a == "0.0.0.0" || a == "::" || a == "::0" {
		return true
	}
	if b == "" || b == "0.0.0.0" || b == "::" || b == "::0" {
		return true
	}
	return a == b
}

type portConflictDetail struct {
	InboundID  int
	Remark     string
	Tag        string
	Listen     string
	Port       int
	Transports transportBits
}

func (d *portConflictDetail) String() string {
	name := d.Remark
	if name == "" {
		name = d.Tag
	}
	if name == "" {
		name = fmt.Sprintf("#%d", d.InboundID)
	} else if d.InboundID > 0 {
		name = fmt.Sprintf("'%s' (#%d)", name, d.InboundID)
	} else {
		name = fmt.Sprintf("'%s'", name)
	}
	listen := d.Listen
	if listen == "" || listen == "0.0.0.0" {
		listen = "*"
	}
	return fmt.Sprintf("port %d (%s) already used by inbound %s on %s",
		d.Port, transportTagSuffix(d.Transports), name, listen)
}

func transportTagSuffix(b transportBits) string {
	switch b {
	case transportTCP:
		return "tcp"
	case transportUDP:
		return "udp"
	case transportTCP | transportUDP:
		return "tcpudp"
	}
	return "any"
}

func sameNode(a, b int64) bool {
	return a == b
}

// CheckPortConflict checks if a new inbound would conflict with existing ones.
func (s *InboundService) CheckPortConflict(inbound *database.Inbound, ignoreID int64) (*portConflictDetail, error) {
	newBits := inboundTransports(inbound.Protocol, inbound.StreamSettings, inbound.Settings)

	var candidates []database.Inbound
	q := database.DB.Model(&database.Inbound{}).Where("port = ?", inbound.Port)
	if ignoreID > 0 {
		q = q.Where("id != ?", ignoreID)
	}
	if err := q.Find(&candidates).Error; err != nil {
		return nil, err
	}

	for _, c := range candidates {
		if !sameNode(c.NodeID, inbound.NodeID) {
			continue
		}
		if !listenOverlaps(c.Listen, inbound.Listen) {
			continue
		}
		existingBits := inboundTransports(c.Protocol, c.StreamSettings, c.Settings)
		shared := existingBits & newBits
		if shared == 0 {
			continue
		}
		return &portConflictDetail{
			InboundID:  int(c.ID),
			Remark:     c.Remark,
			Tag:        c.Tag,
			Listen:     c.Listen,
			Port:       c.Port,
			Transports: shared,
		}, nil
	}
	return nil, nil
}

// GenerateInboundTag generates a unique tag for an inbound based on port and transport.
func GenerateInboundTag(port int, nodeID *int64, protocol database.Protocol, streamSettings, settings string) string {
	bits := inboundTransports(protocol, streamSettings, settings)
	prefix := ""
	if nodeID != nil && *nodeID > 0 {
		prefix = fmt.Sprintf("n%d-", *nodeID)
	}
	return prefix + fmt.Sprintf("in-%d-%s", port, transportTagSuffix(bits))
}
