package service

import (
	"fmt"
	"net"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// CleanIPScanner scans IPs for clean (unblocked) status.
type CleanIPScanner struct {
	mu       sync.Mutex
	running  bool
	results  []CleanIPScanResult
	stopCh   chan struct{}
	interval time.Duration
}

// CleanIPScanResult stores a single scan result.
type CleanIPScanResult struct {
	ID        int64  `json:"id"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol,omitempty"`
	LatencyMs int64  `json:"latency_ms"`
	IsClean   bool   `json:"is_clean"`
	CheckedAt int64  `json:"checked_at"`
}

// NewCleanIPScanner creates a new clean IP scanner.
func NewCleanIPScanner() *CleanIPScanner {
	return &CleanIPScanner{
		interval: 6 * time.Hour,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the periodic scanning loop.
func (s *CleanIPScanner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run initial scan
		s.scan()

		for {
			select {
			case <-ticker.C:
				s.scan()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop stops the scanning loop.
func (s *CleanIPScanner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

// ScanNow triggers an immediate scan.
func (s *CleanIPScanner) ScanNow() {
	go s.scan()
}

// GetResults returns the latest scan results.
func (s *CleanIPScanner) GetResults() []CleanIPScanResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	results := make([]CleanIPScanResult, len(s.results))
	copy(results, s.results)
	return results
}

// scan performs IP scanning logic.
func (s *CleanIPScanner) scan() {
	s.mu.Lock()
	if s.running == false {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	var nodes []database.Node
	if err := database.DB.Find(&nodes).Error; err != nil {
		return
	}

	var newResults []CleanIPScanResult
	for _, node := range nodes {
		if node.Address == "" {
			continue
		}
		result := s.checkIP(node.Address, node.Port)
		newResults = append(newResults, result)
	}

	s.mu.Lock()
	s.results = newResults
	s.mu.Unlock()
}

// checkIP tests connectivity to an IP:port.
func (s *CleanIPScanner) checkIP(ip string, port int) CleanIPScanResult {
	result := CleanIPScanResult{
		IP:        ip,
		Port:      port,
		CheckedAt: time.Now().UnixMilli(),
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 5*time.Second)
	if err != nil {
		result.IsClean = false
		result.LatencyMs = 0
		return result
	}
	defer conn.Close()

	result.IsClean = true
	result.LatencyMs = time.Since(start).Milliseconds()
	return result
}

// SetInterval sets the scan interval.
func (s *CleanIPScanner) SetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = d
}
