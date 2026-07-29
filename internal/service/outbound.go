package service

import (
	"context"
	"fmt"

	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// OutboundService manages proxy outbound (egress) configurations.
type OutboundService struct {
	eventBus events.Publisher
}

// NewOutboundService creates a new outbound service.
func NewOutboundService(bus events.Publisher) *OutboundService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &OutboundService{
		eventBus: bus,
	}
}

// Create adds a new outbound.
func (s *OutboundService) Create(ctx context.Context, ob *database.Outbound) (*database.Outbound, error) {
	if err := database.CreateOutbound(ob); err != nil {
		return nil, fmt.Errorf("create outbound: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "outbound.created",
		Data: map[string]any{"id": ob.ID, "tag": ob.Tag},
	})
	return ob, nil
}

// GetByID retrieves an outbound by ID.
func (s *OutboundService) GetByID(id int64) (*database.Outbound, error) {
	return database.GetOutboundByID(id)
}

// GetByTag retrieves an outbound by tag.
func (s *OutboundService) GetByTag(tag string) (*database.Outbound, error) {
	return database.GetOutboundByTag(tag)
}

// List returns all outbounds with optional node filter.
func (s *OutboundService) List(nodeID int64) ([]database.Outbound, error) {
	return database.ListOutbounds(nodeID)
}

// ListVisible returns only non-hidden outbounds (for subscription).
func (s *OutboundService) ListVisible() ([]database.Outbound, error) {
	var list []database.Outbound
	return list, database.DB.Where("hidden = ? AND enable = ?", false, true).
		Order("id asc").Find(&list).Error
}

// Update modifies an existing outbound.
func (s *OutboundService) Update(ctx context.Context, ob *database.Outbound) error {
	if err := database.UpdateOutbound(ob); err != nil {
		return fmt.Errorf("update outbound: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "outbound.updated",
		Data: map[string]any{"id": ob.ID, "tag": ob.Tag},
	})
	return nil
}

// Delete removes an outbound.
func (s *OutboundService) Delete(ctx context.Context, id int64) error {
	if err := database.DeleteOutbound(id); err != nil {
		return fmt.Errorf("delete outbound: %w", err)
	}
	s.eventBus.Publish(events.Event{
		Type: "outbound.deleted",
		Data: map[string]any{"id": id},
	})
	return nil
}

// ToggleEnable enables or disables an outbound.
func (s *OutboundService) ToggleEnable(ctx context.Context, id int64, enable bool) error {
	return database.DB.Model(&database.Outbound{}).Where("id = ?", id).
		Update("enable", enable).Error
}

// ToggleHide shows or hides an outbound in subscription exports.
func (s *OutboundService) ToggleHide(ctx context.Context, id int64, hidden bool) error {
	return database.DB.Model(&database.Outbound{}).Where("id = ?", id).
		Update("hidden", hidden).Error
}
