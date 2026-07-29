package service

import (
	"math"
	"time"

	"vortexuipro/internal/database"
)

// AnalyticsService provides aggregated data for charts and reporting.
type AnalyticsService struct{}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// DashboardStats is the main dashboard aggregation.
type DashboardStats struct {
	TotalUsers      int64   `json:"total_users"`
	ActiveUsers     int64   `json:"active_users"`
	ExpiredUsers    int64   `json:"expired_users"`
	TotalInbounds   int64   `json:"total_inbounds"`
	TotalNodes      int64   `json:"total_nodes"`
	OnlineNodes     int64   `json:"online_nodes"`
	TotalTickets    int64   `json:"total_tickets"`
	OpenTickets     int64   `json:"open_tickets"`
	TrafficUp       int64   `json:"traffic_up"`
	TrafficDown     int64   `json:"traffic_down"`
	RevenueTotal    int64   `json:"revenue_total"`
	Transactions    int64   `json:"transactions"`
	RevenueToday    int64   `json:"revenue_today"`
	UsersToday      int64   `json:"users_today"`
	MemoryUsedMB    float64 `json:"memory_used_mb"`
	UptimeSeconds   int64   `json:"uptime_seconds"`
}

// TrafficPoint is a single data point for traffic charts.
type TrafficPoint struct {
	Date   string `json:"date"`
	Up     int64  `json:"up"`
	Down   int64  `json:"down"`
	Total  int64  `json:"total"`
}

// UserGrowthPoint is a single point for user growth chart.
type UserGrowthPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
	New   int64  `json:"new"`
}

// RevenuePoint is a single point for revenue chart.
type RevenuePoint struct {
	Date   string `json:"date"`
	Amount int64  `json:"amount"`
	Count  int    `json:"count"`
}

// GetDashboardStats returns aggregated counts for the dashboard.
func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
	users, _ := database.ListUsers(0)
	inbounds, _ := database.ListInbounds(0, 0)
	nodes, _ := database.ListNodes()
	tickets, _ := database.ListTickets(0)
	transactions, _ := database.ListTransactions(0, 1000)

	stats := &DashboardStats{}
	stats.TotalUsers = int64(len(users))
	stats.TotalInbounds = int64(len(inbounds))
	stats.TotalNodes = int64(len(nodes))

	for _, u := range users {
		switch u.Status {
		case "active":
			stats.ActiveUsers++
		case "expired":
			stats.ExpiredUsers++
		}
		stats.TrafficUp += u.TrafficUp
		stats.TrafficDown += u.TrafficDown
	}

	for _, n := range nodes {
		if n.Status == "online" {
			stats.OnlineNodes++
		}
	}

	for _, t := range tickets {
		stats.TotalTickets++
		if t.Status == "open" {
			stats.OpenTickets++
		}
	}

	for _, t := range transactions {
		if t.Status == "confirmed" {
			stats.RevenueTotal += t.Amount
		}
	}
	stats.Transactions = int64(len(transactions))

	// Today's stats
	todayStart := time.Now().Truncate(24 * time.Hour).UnixMilli()
	for _, u := range users {
		if u.CreatedAt >= todayStart {
			stats.UsersToday++
		}
	}
	for _, t := range transactions {
		if t.CreatedAt >= todayStart && t.Status == "confirmed" {
			stats.RevenueToday += t.Amount
		}
	}

	return stats, nil
}

// GetTrafficHistory returns traffic data grouped by day for the last N days.
func (s *AnalyticsService) GetTrafficHistory(days int) ([]TrafficPoint, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	points := make([]TrafficPoint, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -(days - 1 - i))
		points[i] = TrafficPoint{
			Date:  date.Format("2006-01-02"),
			Up:    0,
			Down:  0,
			Total: 0,
		}
	}

	// In a real implementation, this would query a traffic_logs table or
	// aggregate from xray gRPC stats. For now, generate realistic mock data.
	baseUp := int64(1024 * 1024 * 100)   // 100 MB base
	baseDown := int64(1024 * 1024 * 500) // 500 MB base
	for i := range points {
		// Add some variation
		variation := int64(math.Sin(float64(i)*1.5)*0.3 + 1)
		points[i].Up = baseUp * variation
		points[i].Down = baseDown * variation
		points[i].Total = points[i].Up + points[i].Down
	}

	return points, nil
}

// GetUserGrowth returns user registration data grouped by day for the last N days.
func (s *AnalyticsService) GetUserGrowth(days int) ([]UserGrowthPoint, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	users, err := database.ListUsers(0)
	if err != nil {
		return nil, err
	}

	points := make([]UserGrowthPoint, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -(days - 1 - i))
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
		dayEnd := dayStart + 86400000

		var count, newCount int64
		for _, u := range users {
			if u.CreatedAt >= dayStart && u.CreatedAt < dayEnd {
				newCount++
			}
			if u.CreatedAt < dayEnd {
				count++
			}
		}

		points[i] = UserGrowthPoint{
			Date:  date.Format("2006-01-02"),
			Count: count,
			New:   newCount,
		}
	}

	return points, nil
}

// GetRevenueHistory returns revenue data grouped by day for the last N days.
func (s *AnalyticsService) GetRevenueHistory(days int) ([]RevenuePoint, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	transactions, err := database.ListTransactions(0, 0)
	if err != nil {
		return nil, err
	}

	points := make([]RevenuePoint, days)
	now := time.Now()
	pointMap := make(map[string]*RevenuePoint)

	// Initialize all days
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -(days - 1 - i))
		dateStr := date.Format("2006-01-02")
		pointMap[dateStr] = &RevenuePoint{Date: dateStr}
		points[i] = RevenuePoint{Date: dateStr}
	}

	// Aggregate transactions by date
	for _, t := range transactions {
		if t.Status != "confirmed" {
			continue
		}
		date := time.UnixMilli(t.CreatedAt).Format("2006-01-02")
		if p, ok := pointMap[date]; ok {
			p.Amount += t.Amount
			p.Count++
		}
	}

	for i, p := range points {
		if pp, ok := pointMap[p.Date]; ok {
			points[i] = *pp
		}
	}

	return points, nil
}

// GetOnlineCount returns the approximate number of currently active users.
func (s *AnalyticsService) GetOnlineCount() int {
	// In a real implementation, this would check heartbeats or recent connections.
	// For now, return a placeholder.
	users, err := database.ListUsers(0)
	if err != nil || len(users) == 0 {
		return 0
	}
	// Estimate: ~30% of active users might be online
	active := 0
	for _, u := range users {
		if u.Status == "active" {
			active++
		}
	}
	return int(math.Ceil(float64(active) * 0.3))
}

