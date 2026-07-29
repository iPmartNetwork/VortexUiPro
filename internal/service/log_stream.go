package service

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ─── Log Levels ────────────────────────────────────────────────────────
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ─── LogEntry ──────────────────────────────────────────────────────────
type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"` // "core", "panel", "agent"
	Message   string `json:"message"`
}

// ─── LogSubscriber ─────────────────────────────────────────────────────
type LogSubscriber struct {
	ID        string
	MinLevel  LogLevel
	Filter    string // keyword filter (empty = all)
	Source    string // "core", "panel", "agent" (empty = all)
	Buffer    []LogEntry
	BufferMu  sync.Mutex
	Notify    chan struct{}
	Closed    bool
	CloseMu   sync.Mutex
}

// ─── LogStreamService ──────────────────────────────────────────────────
type LogStreamService struct {
	subs     map[string]*LogSubscriber
	subMu    sync.RWMutex
	mu       sync.Mutex
	maxLines int

	// File tailing
	coreLogPath string
	panelLogPath string
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewLogStreamService creates a new log streaming service.
func NewLogStreamService(coreLogPath, panelLogPath string) *LogStreamService {
	return &LogStreamService{
		subs:         make(map[string]*LogSubscriber),
		maxLines:     10000,
		coreLogPath:  coreLogPath,
		panelLogPath: panelLogPath,
		stopCh:       make(chan struct{}),
	}
}

// Start begins tailing log files.
func (s *LogStreamService) Start() {
	if s.coreLogPath != "" {
		s.wg.Add(1)
		go s.tailFile(s.coreLogPath, "core", s.stopCh)
	}
	if s.panelLogPath != "" {
		s.wg.Add(1)
		go s.tailFile(s.panelLogPath, "panel", s.stopCh)
	}
	log.Println("[LogStream] started")
}

// Stop stops all log tailing.
func (s *LogStreamService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Println("[LogStream] stopped")
}

// Subscribe creates a new log subscriber.
func (s *LogStreamService) Subscribe(id string, minLevel LogLevel, filter, source string) *LogSubscriber {
	sub := &LogSubscriber{
		ID:       id,
		MinLevel: minLevel,
		Filter:   strings.ToLower(filter),
		Source:   source,
		Buffer:   make([]LogEntry, 0, 256),
		Notify:   make(chan struct{}, 1),
	}

	s.subMu.Lock()
	s.subs[id] = sub
	s.subMu.Unlock()

	return sub
}

// Unsubscribe removes a log subscriber.
func (s *LogStreamService) Unsubscribe(id string) {
	s.subMu.Lock()
	if sub, ok := s.subs[id]; ok {
		sub.CloseMu.Lock()
		sub.Closed = true
		sub.CloseMu.Unlock()
		delete(s.subs, id)
	}
	s.subMu.Unlock()
}

// GetRecentLogs returns recent logs from memory buffer.
func (s *LogStreamService) GetRecentLogs(count int) []LogEntry {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	var all []LogEntry
	for _, sub := range s.subs {
		sub.BufferMu.Lock()
		start := 0
		if len(sub.Buffer) > count {
			start = len(sub.Buffer) - count
		}
		all = append(all, sub.Buffer[start:]...)
		sub.BufferMu.Unlock()
	}

	if len(all) > count {
		all = all[len(all)-count:]
	}
	return all
}

// PublishLog injects a log entry programmatically.
func (s *LogStreamService) PublishLog(entry LogEntry) {
	s.broadcast(entry)
}

func (s *LogStreamService) tailFile(path, source string, stop chan struct{}) {
	defer s.wg.Done()

	file, err := os.Open(path)
	if err != nil {
		log.Printf("[LogStream] cannot open %s: %v", path, err)
		return
	}
	defer file.Close()

	// Seek to end for tail -f behavior
	file.Seek(0, io.SeekEnd)

	reader := bufio.NewReader(file)
	for {
		select {
		case <-stop:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}

			line = strings.TrimRight(line, "\n\r")
			if line == "" {
				continue
			}

			entry := s.parseLine(line, source)
			s.broadcast(entry)
		}
	}
}

func (s *LogStreamService) parseLine(line string, source string) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Source:    source,
		Message:   line,
		Level:     "info",
	}

	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal"):
		entry.Level = "error"
	case strings.Contains(lower, "warn"):
		entry.Level = "warn"
	case strings.Contains(lower, "debug"):
		entry.Level = "debug"
	}

	return entry
}

func (s *LogStreamService) broadcast(entry LogEntry) {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	for _, sub := range s.subs {
		if sub.Closed {
			continue
		}
		if sub.Source != "" && sub.Source != entry.Source {
			continue
		}
		if sub.Filter != "" && !strings.Contains(strings.ToLower(entry.Message), sub.Filter) {
			continue
		}

		sub.BufferMu.Lock()
		sub.Buffer = append(sub.Buffer, entry)
		if len(sub.Buffer) > s.maxLines {
			sub.Buffer = sub.Buffer[len(sub.Buffer)-s.maxLines:]
		}
		sub.BufferMu.Unlock()

		// Non-blocking notify
		select {
		case sub.Notify <- struct{}{}:
		default:
		}
	}
}

// DrainLogs reads and clears buffered logs for a subscriber.
func (sub *LogSubscriber) DrainLogs() []LogEntry {
	sub.BufferMu.Lock()
	defer sub.BufferMu.Unlock()

	logs := make([]LogEntry, len(sub.Buffer))
	copy(logs, sub.Buffer)
	sub.Buffer = sub.Buffer[:0]
	return logs
}

// WaitForLogs returns a channel that fires when new logs are available.
func (sub *LogSubscriber) WaitForLogs() <-chan struct{} {
	return sub.Notify
}
