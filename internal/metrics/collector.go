package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/mem"
	"vortexuipro/internal/database"
	"vortexuipro/internal/service"
)

// ─── Metric Point (for history / time-series) ───────────────────────

// MetricPoint is a single snapshot stored in the ring buffer.
type MetricPoint struct {
	Time           int64   `json:"t"`
	CPUPercent     float64 `json:"cpu"`
	MemoryUsedMB   float64 `json:"mem"`
	DiskUsedPercent float64 `json:"disk"`
	NetBytesSent   uint64  `json:"net_sent"`
	NetBytesRecv   uint64  `json:"net_recv"`
	OnlineNow      int64   `json:"online"`
	GoRoutines     int     `json:"goroutines"`
}

// ─── Collector ──────────────────────────────────────────────────────

// Collector gathers system and application metrics.
type Collector struct {
	mu            sync.RWMutex
	onlineTracker *service.OnlineTracker

	// Gauges — application
	UsersTotal     int64 `json:"users_total"`
	InboundsTotal  int64 `json:"inbounds_total"`
	ClientsTotal   int64 `json:"clients_total"`
	NodesTotal     int64 `json:"nodes_total"`
	OnlineNow      int64 `json:"online_now"`
	TrafficUpTotal   int64 `json:"traffic_up_total"`
	TrafficDownTotal int64 `json:"traffic_down_total"`

	// Gauges — runtime
	GoRoutines    int     `json:"go_routines"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	MemoryUsedMB  float64 `json:"memory_used_mb"`
	MemoryPct     float64 `json:"memory_pct"`
	CPUThreads    int     `json:"cpu_threads"`
	CPUPercent    float64 `json:"cpu_percent"`
	UptimeSec     int64   `json:"uptime_seconds"`

	// Gauges — disk
	DiskTotalMB     float64 `json:"disk_total_mb"`
	DiskUsedMB      float64 `json:"disk_used_mb"`
	DiskFreeMB      float64 `json:"disk_free_mb"`
	DiskUsedPercent float64 `json:"disk_used_percent"`

	// Gauges — network
	NetBytesSent   uint64 `json:"net_bytes_sent"`
	NetBytesRecv   uint64 `json:"net_bytes_recv"`
	NetPacketsSent uint64 `json:"net_packets_sent"`
	NetPacketsRecv uint64 `json:"net_packets_recv"`

	// Gauges — load
	LoadAvg1  float64 `json:"load_avg_1"`
	LoadAvg5  float64 `json:"load_avg_5"`
	LoadAvg15 float64 `json:"load_avg_15"`

	// System info (set once)
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`

	// History ring buffer (last 60 snapshots @ 15s = 15 min)
	history      []MetricPoint
	historyCap   int
	historyIndex int
	historyFull  bool

	startedAt time.Time
	stopCh    chan struct{}


}

// NewCollector creates a new collector and starts its refresh loop.
func NewCollector(ot *service.OnlineTracker) *Collector {
	c := &Collector{
		onlineTracker: ot,
		startedAt:     time.Now(),
		stopCh:        make(chan struct{}),
		historyCap:    60,
		history:       make([]MetricPoint, 60),
	}

	// Populate system info once
	hostInfo, _ := host.Info()
	if hostInfo != nil {
		c.Hostname = hostInfo.Hostname
		c.OS = hostInfo.OS
	}
	c.Arch = runtime.GOARCH

	go c.refreshLoop()
	return c
}

// Stop shuts down the refresh loop.
func (c *Collector) Stop() {
	close(c.stopCh)
}

func (c *Collector) refreshLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Do an immediate first refresh
	c.refresh()

	for {
		select {
		case <-ticker.C:
			c.refresh()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Collector) refresh() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// ─── CPU ────────────────────────────────────────────────
	percent, _ := cpu.Percent(0, false)
	if len(percent) > 0 {
		c.CPUPercent = percent[0]
	}
	c.CPUThreads = runtime.NumCPU()

	// ─── Memory ─────────────────────────────────────────────
	vm, _ := mem.VirtualMemory()
	if vm != nil {
		c.MemoryTotalMB = float64(vm.Total) / 1024 / 1024
		c.MemoryUsedMB = float64(vm.Used) / 1024 / 1024
		c.MemoryPct = vm.UsedPercent
	}

	// ─── Disk ───────────────────────────────────────────────
	// Use root partition
	parts, _ := disk.Partitions(false)
	for _, p := range parts {
		if p.Mountpoint == "/" || p.Mountpoint == "C:\\" || p.Mountpoint == "C:" {
			usage, _ := disk.Usage(p.Mountpoint)
			if usage != nil {
				c.DiskTotalMB = float64(usage.Total) / 1024 / 1024
				c.DiskUsedMB = float64(usage.Used) / 1024 / 1024
				c.DiskFreeMB = float64(usage.Free) / 1024 / 1024
				c.DiskUsedPercent = usage.UsedPercent
			}
			break
		}
	}

	// ─── Network I/O ────────────────────────────────────────
	ioCounters, _ := net.IOCounters(false)
	if len(ioCounters) > 0 {
		io := ioCounters[0]
		c.NetBytesSent = io.BytesSent
		c.NetBytesRecv = io.BytesRecv
		c.NetPacketsSent = io.PacketsSent
		c.NetPacketsRecv = io.PacketsRecv
	}

	// ─── Load Average ───────────────────────────────────────
	avg, _ := load.Avg()
	if avg != nil {
		c.LoadAvg1 = avg.Load1
		c.LoadAvg5 = avg.Load5
		c.LoadAvg15 = avg.Load15
	}

	// ─── Runtime ────────────────────────────────────────────
	c.GoRoutines = runtime.NumGoroutine()
	c.UptimeSec = int64(time.Since(c.startedAt).Seconds())

	// ─── Application metrics (guard against nil DB) ─────────
	if database.DB != nil {
		database.DB.Model(&database.User{}).Count(&c.UsersTotal)
		database.DB.Model(&database.Inbound{}).Count(&c.InboundsTotal)
		database.DB.Model(&database.Client{}).Count(&c.ClientsTotal)
		database.DB.Model(&database.Node{}).Count(&c.NodesTotal)

		database.DB.Raw("SELECT COALESCE(SUM(traffic_up), 0) FROM users").Scan(&c.TrafficUpTotal)
		database.DB.Raw("SELECT COALESCE(SUM(traffic_down), 0) FROM users").Scan(&c.TrafficDownTotal)
	}

	if c.onlineTracker != nil {
		c.OnlineNow = int64(c.onlineTracker.GetOnlineCount())
	}

	// ─── Store history point ────────────────────────────────
	pt := MetricPoint{
		Time:           now.UnixMilli(),
		CPUPercent:     c.CPUPercent,
		MemoryUsedMB:   c.MemoryUsedMB,
		DiskUsedPercent: c.DiskUsedPercent,
		NetBytesSent:   c.NetBytesSent,
		NetBytesRecv:   c.NetBytesRecv,
		OnlineNow:      c.OnlineNow,
		GoRoutines:     c.GoRoutines,
	}
	c.history[c.historyIndex] = pt
	c.historyIndex = (c.historyIndex + 1) % c.historyCap
	if c.historyIndex == 0 {
		c.historyFull = true
	}
}

// Snapshot returns a full metrics snapshot.
func (c *Collector) Snapshot() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]any{
		// App
		"users_total":      c.UsersTotal,
		"inbounds_total":   c.InboundsTotal,
		"clients_total":    c.ClientsTotal,
		"nodes_total":      c.NodesTotal,
		"online_now":       c.OnlineNow,
		"traffic_up_gb":    float64(c.TrafficUpTotal) / 1e9,
		"traffic_down_gb":  float64(c.TrafficDownTotal) / 1e9,

		// CPU
		"cpu_percent":      c.CPUPercent,
		"cpu_threads":      c.CPUThreads,
		"load_avg_1":       c.LoadAvg1,
		"load_avg_5":       c.LoadAvg5,
		"load_avg_15":      c.LoadAvg15,

		// Memory
		"memory_total_mb":  c.MemoryTotalMB,
		"memory_used_mb":   c.MemoryUsedMB,
		"memory_pct":       c.MemoryPct,

		// Disk
		"disk_total_mb":    c.DiskTotalMB,
		"disk_used_mb":     c.DiskUsedMB,
		"disk_free_mb":     c.DiskFreeMB,
		"disk_used_percent": c.DiskUsedPercent,

		// Network
		"net_bytes_sent":   c.NetBytesSent,
		"net_bytes_recv":   c.NetBytesRecv,
		"net_packets_sent": c.NetPacketsSent,
		"net_packets_recv": c.NetPacketsRecv,

		// Runtime
		"go_routines":      c.GoRoutines,
		"uptime_seconds":   c.UptimeSec,
		"uptime_human":     time.Since(c.startedAt).Round(time.Second).String(),

		// System info
		"hostname":         c.Hostname,
		"os":               c.OS,
		"arch":             c.Arch,
		"started_at":       c.startedAt.UnixMilli(),
	}
}

// History returns the time-series data for charts.
func (c *Collector) History() []MetricPoint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.historyFull {
		// Not full yet — only return what we have
		return c.history[:c.historyIndex]
	}

	// Ring buffer is full — return oldest-first
	out := make([]MetricPoint, c.historyCap)
	start := c.historyIndex
	for i := 0; i < c.historyCap; i++ {
		out[i] = c.history[(start+i)%c.historyCap]
	}
	return out
}

// PrometheusText returns metrics in Prometheus text format.
func (c *Collector) PrometheusText() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s := "# HELP vortex_metrics VortexUiPro application metrics\n"
	s += "# TYPE vortex_metrics gauge\n"
	s += fmtLine("vortex_users_total", float64(c.UsersTotal))
	s += fmtLine("vortex_inbounds_total", float64(c.InboundsTotal))
	s += fmtLine("vortex_clients_total", float64(c.ClientsTotal))
	s += fmtLine("vortex_nodes_total", float64(c.NodesTotal))
	s += fmtLine("vortex_online_now", float64(c.OnlineNow))
	s += fmtLine("vortex_traffic_up_bytes", float64(c.TrafficUpTotal))
	s += fmtLine("vortex_traffic_down_bytes", float64(c.TrafficDownTotal))
	s += fmtLine("vortex_go_routines", float64(c.GoRoutines))
	s += fmtLine("vortex_memory_used_bytes", c.MemoryUsedMB*1024*1024)
	s += fmtLine("vortex_memory_total_bytes", c.MemoryTotalMB*1024*1024)
	s += fmtLine("vortex_memory_percent", c.MemoryPct)
	s += fmtLine("vortex_cpu_percent", c.CPUPercent)
	s += fmtLine("vortex_cpu_threads", float64(c.CPUThreads))
	s += fmtLine("vortex_disk_used_bytes", c.DiskUsedMB*1024*1024)
	s += fmtLine("vortex_disk_total_bytes", c.DiskTotalMB*1024*1024)
	s += fmtLine("vortex_disk_percent", c.DiskUsedPercent)
	s += fmtLine("vortex_net_bytes_sent", float64(c.NetBytesSent))
	s += fmtLine("vortex_net_bytes_recv", float64(c.NetBytesRecv))
	s += fmtLine("vortex_load_avg_1", c.LoadAvg1)
	s += fmtLine("vortex_load_avg_5", c.LoadAvg5)
	s += fmtLine("vortex_load_avg_15", c.LoadAvg15)
	s += fmtLine("vortex_uptime_seconds", float64(c.UptimeSec))

	// Per-inbound metrics
	var inbounds []database.Inbound
	if database.DB != nil {
		database.DB.Find(&inbounds)
		for _, ib := range inbounds {
			label := sanitizeLabel(ib.Tag)
			s += fmt.Sprintf("vortex_inbound_traffic_up{tag=%q,protocol=%q} %d\n", label, ib.Protocol, ib.Up)
			s += fmt.Sprintf("vortex_inbound_traffic_down{tag=%q,protocol=%q} %d\n", label, ib.Protocol, ib.Down)
		}
	}

	return s
}

func fmtLine(name string, value float64) string {
	return fmt.Sprintf("vortex_%s %f\n", name, value)
}

func sanitizeLabel(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == ':' || c == '-' {
			b = append(b, c)
		} else {
			b = append(b, '_')
		}
	}
	return string(b)
}
