package monitor

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"vortexuipro/internal/logger"
)

// ClientActivityCollector collects client activity events via Unix domain socket.
type ClientActivityCollector struct {
	socketPath    string
	flushInterval time.Duration
	queue         chan clientActivityEvent

	mu      sync.Mutex
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	conn    *net.UnixConn
	wg      sync.WaitGroup

	lastMaintenance atomic.Int64
	lastFlush       time.Time
	eventsFlushed   atomic.Int64
}

type clientActivityEvent struct {
	Version       int    `json:"version"`
	ClientID      int64  `json:"clientId"`
	Email         string `json:"email"`
	Generation    int64  `json:"generation"`
	DataEpoch     int64  `json:"dataEpoch"`
	SourceIP      string `json:"sourceIp"`
	Destination   string `json:"destination"`
	UploadBytes   int64  `json:"uploadBytes"`
	DownloadBytes int64  `json:"downloadBytes"`
	ObservedAt    int64  `json:"observedAt"`
}

const (
	defaultSocketPath        = "/run/vortexuipro/client-activity.sock"
	queueSize               = 4096
	defaultFlushInterval    = time.Second
	datagramSize            = 4096
	batchSize              = 512
	retentionPeriod         = 7 * 24 * time.Hour
	maintenanceInterval     = 10 * time.Minute
)

// NewClientActivityCollector creates a new activity collector.
func NewClientActivityCollector() *ClientActivityCollector {
	socketPath := os.Getenv("VORTEX_ACTIVITY_SOCKET")
	if socketPath == "" {
		socketPath = defaultSocketPath
	}
	return &ClientActivityCollector{
		socketPath:    filepath.Clean(socketPath),
		flushInterval: defaultFlushInterval,
		queue:         make(chan clientActivityEvent, queueSize),
	}
}

// Start starts the collector's read and aggregate loops.
func (c *ClientActivityCollector) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(c.socketPath), 0755); err != nil {
		return err
	}

	// Remove stale socket
	os.Remove(c.socketPath)

	addr := &net.UnixAddr{Name: c.socketPath, Net: "unixgram"}
	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		return err
	}
	os.Chmod(c.socketPath, 0600)
	conn.SetReadBuffer(1024 * 1024)

	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	c.conn = conn
	c.started = true

	c.wg.Add(2)
	go c.readLoop(ctx, conn)
	go c.aggregateLoop(ctx)

	return nil
}

// Stop stops the collector.
func (c *ClientActivityCollector) Stop() {
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return
	}
	cancel := c.cancel
	conn := c.conn
	socketPath := c.socketPath
	c.started = false
	c.mu.Unlock()

	cancel()
	conn.Close()
	c.wg.Wait()
	os.Remove(socketPath)
}

func (c *ClientActivityCollector) readLoop(ctx context.Context, conn *net.UnixConn) {
	defer c.wg.Done()
	buf := make([]byte, datagramSize)

	for {
		n, _, err := conn.ReadFromUnix(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		if n <= 0 || n > datagramSize {
			continue
		}

		var event clientActivityEvent
		if err := json.Unmarshal(buf[:n], &event); err != nil {
			continue
		}

		select {
		case c.queue <- event:
		default:
		}
	}
}

func (c *ClientActivityCollector) aggregateLoop(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	pending := make(map[string]clientActivityEvent)
	count := 0

	flush := func() {
		if len(pending) == 0 {
			return
		}
		events := make([]clientActivityEvent, 0, len(pending))
		for _, e := range pending {
			events = append(events, e)
		}
		pending = make(map[string]clientActivityEvent)
		count = 0

		if err := c.persist(events); err != nil {
			log.Printf("client activity flush: %v", err)
		}
	}

	add := func(event clientActivityEvent) {
		key := eventKey(event)
		existing, ok := pending[key]
		if !ok {
			pending[key] = event
		} else {
			existing.UploadBytes += event.UploadBytes
			existing.DownloadBytes += event.DownloadBytes
			if event.ObservedAt > existing.ObservedAt {
				existing.ObservedAt = event.ObservedAt
			}
			pending[key] = existing
		}
		count++
	}

	for {
		select {
		case event := <-c.queue:
			add(event)
			if count >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			for {
				select {
				case event := <-c.queue:
					add(event)
				default:
					flush()
					return
				}
			}
		}
	}
}

func eventKey(e clientActivityEvent) string {
	return strings.Join([]string{
		int64Str(e.ClientID), e.Email, int64Str(e.Generation),
		int64Str(e.DataEpoch), e.SourceIP, e.Destination,
	}, "|")
}

func int64Str(v int64) string {
	return strconv.FormatInt(v, 10)
}

func (c *ClientActivityCollector) persist(events []clientActivityEvent) error {
	c.eventsFlushed.Add(int64(len(events)))
	c.lastFlush = time.Now()

	// Log activity for now (DB persistence can be added when needed)
	for _, e := range events {
		logger.Printf("client activity: id=%d email=%s up=%d down=%d src=%s",
			e.ClientID, e.Email, e.UploadBytes, e.DownloadBytes, e.SourceIP)
	}
	return nil
}

// Stats returns collector statistics.
func (c *ClientActivityCollector) Stats() map[string]any {
	return map[string]any{
		"started":        c.started,
		"events_flushed": c.eventsFlushed.Load(),
		"queue_size":     len(c.queue),
		"socket_path":    c.socketPath,
		"last_flush":     c.lastFlush.Format(time.RFC3339),
	}
}
