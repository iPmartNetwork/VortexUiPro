package service

import (
	"context"
	"fmt"

	"vortexuipro/internal/database"
	"vortexuipro/internal/domain"
	"vortexuipro/internal/events"
)

// InboundService manages proxy inbound configurations via database.
type InboundService struct {
	eventBus events.Publisher
	xray     *XrayService
}

// NewInboundService creates a new inbound service.
func NewInboundService(bus events.Publisher, xray *XrayService) *InboundService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &InboundService{
		eventBus: bus,
		xray:     xray,
	}
}

// Create adds a new inbound to the database.
func (s *InboundService) Create(ctx context.Context, ib *database.Inbound) (*database.Inbound, error) {
	if err := database.CreateInbound(ib); err != nil {
		return nil, fmt.Errorf("create inbound: %w", err)
	}
	s.syncToCore(ctx)
	return ib, nil
}

// GetByID retrieves an inbound by ID.
func (s *InboundService) GetByID(id int64) (*database.Inbound, error) {
	return database.GetInboundByID(id)
}

// List returns all inbounds with optional filtering.
func (s *InboundService) List(userID, nodeID int64) ([]database.Inbound, error) {
	return database.ListInbounds(userID, nodeID)
}

// Update modifies an existing inbound.
func (s *InboundService) Update(ctx context.Context, ib *database.Inbound) error {
	if err := database.UpdateInbound(ib); err != nil {
		return err
	}
	s.syncToCore(ctx)
	return nil
}

// Delete removes an inbound by ID.
func (s *InboundService) Delete(ctx context.Context, id int64) error {
	if err := database.DeleteInbound(id); err != nil {
		return err
	}
	s.syncToCore(ctx)
	return nil
}

// GetAll returns all inbounds for core sync.
func (s *InboundService) GetAll() ([]database.Inbound, error) {
	return database.ListInbounds(0, 0)
}

// syncToCore regenerates the core config with current inbounds.
func (s *InboundService) syncToCore(ctx context.Context) {
	inbounds, err := database.ListInbounds(0, 0)
	if err != nil || s.xray == nil {
		return
	}
	go func() {
		ibList := make([]domain.Inbound, 0, len(inbounds))
		for _, ib := range inbounds {
			ibList = append(ibList, domain.Inbound{
				ID:              ib.ID,
				Tag:             ib.Tag,
				Protocol:        domain.Protocol(ib.Protocol),
				Listen:          ib.Listen,
				Port:            ib.Port,
				Settings:        ib.Settings,
				StreamSettings:  ib.StreamSettings,
				Sniffing:        ib.Sniffing,
				Status:          domain.InboundStatus(ib.Status),
				Enable:          ib.Enable,
				Remark:          ib.Remark,
			})
		}
		_ = s.xray.ApplyInbounds(ctx, ibList)
	}()
}
