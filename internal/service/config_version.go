package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// ConfigVersion is an alias to database.ConfigVersion for convenience.
type ConfigVersion = database.ConfigVersion

// ConfigVersionService manages versioned configuration snapshots.
type ConfigVersionService struct {
	eventBus events.Publisher
	mu       sync.Mutex
}

// NewConfigVersionService creates a new config version service.
func NewConfigVersionService(bus events.Publisher) *ConfigVersionService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &ConfigVersionService{eventBus: bus}
}

// Snapshot creates a new version of a configuration.
func (s *ConfigVersionService) Snapshot(ctx context.Context, resource string, resourceID int64, data interface{}, description, createdBy string) (*ConfigVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal config data: %w", err)
	}

	// Get next version number
	var lastVersion int
	var last ConfigVersion
	if err := database.DB.Where("resource = ? AND resource_id = ?", resource, resourceID).
		Order("version desc").First(&last).Error; err == nil {
		lastVersion = last.Version
	}

	ver := &ConfigVersion{
		Resource:    resource,
		ResourceID:  resourceID,
		Version:     lastVersion + 1,
		Data:        string(raw),
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().UnixMilli(),
	}

	if err := database.DB.Create(ver).Error; err != nil {
		return nil, fmt.Errorf("create config version: %w", err)
	}

	// Cleanup old versions (keep last 20)
	var count int64
	database.DB.Model(&ConfigVersion{}).
		Where("resource = ? AND resource_id = ?", resource, resourceID).
		Count(&count)
	if count > 20 {
		var oldVersions []ConfigVersion
		database.DB.Where("resource = ? AND resource_id = ?", resource, resourceID).
			Order("version asc").Limit(int(count) - 20).Find(&oldVersions)
		for _, v := range oldVersions {
			database.DB.Delete(&v)
		}
	}

	s.eventBus.Publish(events.Event{
		Type:    "config.version.created",
		Message: fmt.Sprintf("%s/%d version %d", resource, resourceID, ver.Version),
	})

	return ver, nil
}

// ListVersions returns all versions for a resource.
func (s *ConfigVersionService) ListVersions(resource string, resourceID int64) ([]ConfigVersion, error) {
	var list []ConfigVersion
	if err := database.DB.Where("resource = ? AND resource_id = ?", resource, resourceID).
		Order("version desc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list config versions: %w", err)
	}
	return list, nil
}

// GetVersion retrieves a specific version.
func (s *ConfigVersionService) GetVersion(id int64) (*ConfigVersion, error) {
	var v ConfigVersion
	if err := database.DB.First(&v, id).Error; err != nil {
		return nil, fmt.Errorf("config version not found: %w", err)
	}
	return &v, nil
}

// Rollback restores a previous version of a configuration.
func (s *ConfigVersionService) Rollback(ctx context.Context, id int64, createdBy string) (*ConfigVersion, error) {
	ver, err := s.GetVersion(id)
	if err != nil {
		return nil, err
	}

	// Create a new version with the rolled-back data
	return s.Snapshot(ctx, ver.Resource, ver.ResourceID, ver.Data,
		fmt.Sprintf("Rollback to version %d", ver.Version), createdBy)
}

// DeleteVersion removes a specific version.
func (s *ConfigVersionService) DeleteVersion(id int64) error {
	return database.DB.Delete(&ConfigVersion{}, id).Error
}
