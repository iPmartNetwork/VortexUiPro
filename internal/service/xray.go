package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"vortexuipro/internal/core"
	"vortexuipro/internal/core/xray"
	"vortexuipro/internal/domain"
	"vortexuipro/internal/events"
)

// XrayService manages xray-core lifecycle, configuration, and gRPC API integration.
type XrayService struct {
	driver    *xray.Driver
	eventBus  events.Publisher
	mu        struct{}
	cfg       xray.Config

	inbounds  []domain.Inbound
	outbounds []domain.Outbound
	routing   *domain.RoutingConfig
}

// NewXrayService creates a new xray service with full gRPC capabilities.
func NewXrayService(cfg xray.Config, bus events.Publisher) *XrayService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &XrayService{
		driver:   xray.New(cfg),
		eventBus: bus,
		cfg:      cfg,
	}
}

// Start launches xray-core with the current configuration.
func (s *XrayService) Start(ctx context.Context) error {
	return s.driver.Start(ctx)
}

// Stop terminates xray-core gracefully.
func (s *XrayService) Stop(ctx context.Context) error {
	return s.driver.Stop(ctx)
}

// Restart restarts xray-core.
func (s *XrayService) Restart(ctx context.Context) error {
	return s.driver.Restart(ctx)
}

// Status returns the current core status.
func (s *XrayService) Status(ctx context.Context) core.CoreStatus {
	return s.driver.Status(ctx)
}

// ApplyInbounds rebuilds the config with given inbounds and applies it.
func (s *XrayService) ApplyInbounds(ctx context.Context, inbounds []domain.Inbound) error {
	cfg, err := xray.XrayJSONConfig(inbounds, s.outbounds, s.routing)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	return s.driver.ApplyConfig(ctx, cfg)
}

// ApplyOutbounds rebuilds the config with given outbounds.
func (s *XrayService) ApplyOutbounds(ctx context.Context, outbounds []domain.Outbound) error {
	cfg, err := xray.XrayJSONConfig(s.inbounds, outbounds, s.routing)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	return s.driver.ApplyConfig(ctx, cfg)
}

// ApplyRouting rebuilds the config with given routing rules.
func (s *XrayService) ApplyRouting(ctx context.Context, routing *domain.RoutingConfig) error {
	cfg, err := xray.XrayJSONConfig(s.inbounds, s.outbounds, routing)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	return s.driver.ApplyConfig(ctx, cfg)
}

// ApplyRoutingConfig applies a raw routing JSON to the running core via gRPC.
func (s *XrayService) ApplyRoutingConfig(ctx context.Context, routingJSON []byte) error {
	return s.driver.ApplyRoutingConfig(ctx, routingJSON)
}

// CollectTraffic gathers traffic stats from the running core.
func (s *XrayService) CollectTraffic(ctx context.Context) ([]core.TrafficStats, error) {
	return s.driver.CollectTraffic(ctx)
}

// CollectClientTraffic gathers per-client traffic stats.
func (s *XrayService) CollectClientTraffic(ctx context.Context) ([]xray.ClientTraffic, error) {
	return s.driver.CollectClientTraffic(ctx)
}

// GetOnlineUsers returns users with live connections via xray's gRPC API.
func (s *XrayService) GetOnlineUsers(ctx context.Context) ([]xray.OnlineUser, error) {
	return s.driver.GetOnlineUsers(ctx)
}

// GetBalancerInfo queries a balancer's live state.
func (s *XrayService) GetBalancerInfo(ctx context.Context, tag string) (*xray.BalancerInfo, error) {
	return s.driver.GetBalancerInfo(ctx, tag)
}

// SetBalancerTarget forces a balancer to a specific outbound.
func (s *XrayService) SetBalancerTarget(ctx context.Context, balancerTag, target string) error {
	return s.driver.SetBalancerTarget(ctx, balancerTag, target)
}

// TestRoute tests a route through the running core's router.
func (s *XrayService) TestRoute(ctx context.Context, req *xray.RouteTestRequest) (*xray.RouteTestResult, error) {
	return s.driver.TestRoute(ctx, req)
}

// AddUser adds a user to the running core in real-time via gRPC.
func (s *XrayService) AddUser(ctx context.Context, tag string, user domain.Client) error {
	return s.driver.AddUser(ctx, tag, user)
}

// RemoveUser removes a user from the running core.
func (s *XrayService) RemoveUser(ctx context.Context, tag string, email string) error {
	return s.driver.RemoveUser(ctx, tag, email)
}

// AddInbound adds an inbound to the running core via gRPC (hot-reload).
func (s *XrayService) AddInbound(ctx context.Context, inboundJSON []byte) error {
	return s.driver.AddInbound(ctx, inboundJSON)
}

// RemoveInbound removes an inbound from the running core via gRPC.
func (s *XrayService) RemoveInbound(ctx context.Context, tag string) error {
	return s.driver.RemoveInbound(ctx, tag)
}

// ValidateConfig validates an xray JSON config.
func (s *XrayService) ValidateConfig(config []byte) error {
	return xray.ValidateConfig(config)
}

// GetProcessInfo returns runtime info about the xray process.
func (s *XrayService) GetProcessInfo() map[string]any {
	return s.driver.GetProcessInfo()
}

// GetLogs returns recent log lines from the xray process.
func (s *XrayService) GetLogs(n int) []string {
	return s.driver.GetLogs(n)
}

// StartTrafficPolling starts a background goroutine that periodically collects
// traffic and publishes it to the event bus.
func (s *XrayService) StartTrafficPolling(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := s.CollectTraffic(ctx)
				if err != nil {
					log.Printf("[xray] Traffic poll error: %v", err)
					continue
				}
				for _, stat := range stats {
					s.eventBus.Publish(events.Event{
						Type:   events.TrafficUpdate,
						Data: map[string]any{
							"tag":  stat.Tag,
							"up":   stat.Up,
							"down": stat.Down,
						},
					})
				}
			}
		}
	}()
}
