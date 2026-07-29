package service

import (
	"encoding/json"
	"fmt"

	"vortexuipro/internal/database"
	"vortexuipro/internal/logger"
)

// MigrationResult holds the result of a migration operation.
type MigrationResult struct {
	Migrated int `json:"migrated"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// MigrateInboundsToNode migrates all inbounds from one node to another.
func (s *InboundService) MigrateInboundsToNode(fromNodeID, toNodeID int64) (*MigrationResult, error) {
	var result MigrationResult

	// Load inbounds from source node
	var inbounds []database.Inbound
	if err := database.DB.Where("node_id = ?", fromNodeID).Find(&inbounds).Error; err != nil {
		return nil, fmt.Errorf("load source inbounds: %w", err)
	}

	for _, ib := range inbounds {
		// Check for port conflicts on the target node
		conflict, err := s.CheckPortConflict(&ib, ib.ID)
		if err != nil {
			logger.Warnf("migration: port conflict check failed for inbound %d: %v", ib.ID, err)
			result.Failed++
			continue
		}
		if conflict != nil {
			logger.Infof("migration: skipping inbound %d due to port conflict: %s", ib.ID, conflict.String())
			result.Skipped++
			continue
		}

		// Migrate the inbound
		if err := database.DB.Model(&ib).Update("node_id", toNodeID).Error; err != nil {
			logger.Warnf("migration: failed to migrate inbound %d: %v", ib.ID, err)
			result.Failed++
			continue
		}
		result.Migrated++
	}

	return &result, nil
}

// DuplicateInbound creates a copy of an inbound on a different node.
func (s *InboundService) DuplicateInbound(inboundID int64, targetNodeID int64) (*database.Inbound, error) {
	// Load source inbound
	ib, err := s.GetByID(inboundID)
	if err != nil {
		return nil, fmt.Errorf("load source inbound: %w", err)
	}

	// Create a new inbound based on the source
	newIB := database.Inbound{
		UserID:         ib.UserID,
		NodeID:         targetNodeID,
		Tag:            ib.Tag + "-copy",
		Protocol:       ib.Protocol,
		Listen:         ib.Listen,
		Port:           ib.Port,
		Status:         ib.Status,
		Remark:         ib.Remark + " (copy)",
		Enable:         ib.Enable,
		Settings:       ib.Settings,
		StreamSettings: ib.StreamSettings,
		Sniffing:       ib.Sniffing,
		UpMbps:         ib.UpMbps,
		DownMbps:       ib.DownMbps,
		TotalGB:        ib.TotalGB,
	}

	// Check for tag uniqueness
	var count int64
	database.DB.Model(&database.Inbound{}).Where("tag = ?", newIB.Tag).Count(&count)
	if count > 0 {
		newIB.Tag = fmt.Sprintf("%s-copy-%d", ib.Tag, targetNodeID)
	}

	// Check port conflict
	conflict, err := s.CheckPortConflict(&newIB, 0)
	if err != nil {
		return nil, fmt.Errorf("port conflict check: %w", err)
	}
	if conflict != nil {
		return nil, fmt.Errorf("port conflict: %s", conflict.String())
	}

	if err := database.DB.Create(&newIB).Error; err != nil {
		return nil, fmt.Errorf("create duplicate inbound: %w", err)
	}

	return &newIB, nil
}

// MigrateProtocol migrates an inbound from one protocol to another,
// preserving client configurations where possible.
func (s *InboundService) MigrateProtocol(inboundID int64, targetProtocol database.Protocol) (*database.Inbound, error) {
	ib, err := s.GetByID(inboundID)
	if err != nil {
		return nil, fmt.Errorf("load inbound: %w", err)
	}

	// Validate protocol migration compatibility
	sourceProtocol := ib.Protocol
	if sourceProtocol == targetProtocol {
		return ib, nil // nothing to do
	}

	// Migrate protocol and adjust settings
	oldSettings := ib.Settings
	var settings map[string]any
	if oldSettings != "" {
		json.Unmarshal([]byte(oldSettings), &settings)
	}

	newSettings := migrateSettings(sourceProtocol, targetProtocol, settings)
	settingsJSON, _ := json.Marshal(newSettings)

	ib.Protocol = targetProtocol
	ib.Settings = string(settingsJSON)

	if err := database.DB.Save(ib).Error; err != nil {
		return nil, fmt.Errorf("update inbound protocol: %w", err)
	}

	logger.Infof("inbound %d migrated from %s to %s", inboundID, sourceProtocol, targetProtocol)
	return ib, nil
}

// migrateSettings attempts to convert settings from one protocol to another.
func migrateSettings(from, to database.Protocol, settings map[string]any) map[string]any {
	if settings == nil {
		settings = make(map[string]any)
	}

	// VMess/VLESS -> Trojan: preserve client IDs as passwords
	if (from == database.ProtoVMess || from == database.ProtoVLESS) &&
		to == database.ProtoTrojan {
		if clients, ok := settings["clients"].([]any); ok {
			for _, c := range clients {
				if client, ok := c.(map[string]any); ok {
					client["password"] = client["id"]
					delete(client, "id")
				}
			}
			settings["clients"] = clients
		}
	}

	// Trojan -> VMess: generate UUIDs for each password
	if from == database.ProtoTrojan &&
		(to == database.ProtoVMess || to == database.ProtoVLESS) {
		if clients, ok := settings["clients"].([]any); ok {
			for _, c := range clients {
				if client, ok := c.(map[string]any); ok {
					if pass, ok := client["password"].(string); ok {
						client["id"] = pass // use password as UUID placeholder
						delete(client, "password")
					}
				}
			}
			settings["clients"] = clients
		}
	}

	return settings
}

// BatchMigrateInbounds migrates multiple inbounds in a single operation.
func (s *InboundService) BatchMigrateInbounds(ids []int64, targetNodeID int64) (*MigrationResult, error) {
	result := &MigrationResult{}

	for _, id := range ids {
		if _, err := s.GetByID(id); err != nil {
			result.Failed++
			continue
		}
		if err := database.DB.Model(&database.Inbound{}).Where("id = ?", id).Update("node_id", targetNodeID).Error; err != nil {
			result.Failed++
			continue
		}
		result.Migrated++
	}

	return result, nil
}

// ExportInboundJSON exports an inbound configuration as JSON.
func (s *InboundService) ExportInboundJSON(inboundID int64) (string, error) {
	ib, err := s.GetByID(inboundID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(ib, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal inbound: %w", err)
	}

	return string(data), nil
}

// ImportInboundJSON imports an inbound from JSON configuration.
func (s *InboundService) ImportInboundJSON(data string) (*database.Inbound, error) {
	var ib database.Inbound
	if err := json.Unmarshal([]byte(data), &ib); err != nil {
		return nil, fmt.Errorf("unmarshal inbound: %w", err)
	}

	// Reset ID to create new
	ib.ID = 0

	// Check port conflict
	conflict, err := s.CheckPortConflict(&ib, 0)
	if err != nil {
		return nil, fmt.Errorf("port conflict check: %w", err)
	}
	if conflict != nil {
		return nil, fmt.Errorf("port conflict: %s", conflict.String())
	}

	if err := database.DB.Create(&ib).Error; err != nil {
		return nil, fmt.Errorf("create inbound: %w", err)
	}

	return &ib, nil
}
