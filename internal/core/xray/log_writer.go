package xray

import (
	"bytes"
	"fmt"
	"sync"
)

// LogWriter captures the last N lines of xray-core output.
type LogWriter struct {
	mu       sync.RWMutex
	buf      *bytes.Buffer
	lastLine string
	maxLines int
	lines    []string
}

// NewLogWriter creates a new log writer that keeps the last maxLines lines.
func NewLogWriter() *LogWriter {
	return &LogWriter{
		buf:      new(bytes.Buffer),
		maxLines: 100,
		lines:    make([]string, 0, 100),
	}
}

// Write implements io.Writer.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	n, err = w.buf.Write(p)
	if err != nil {
		return n, err
	}

	// Extract complete lines
	for {
		line, err := w.buf.ReadBytes('\n')
		if err != nil {
			// Put remaining bytes back
			w.buf.Write(line)
			break
		}
		// Trim newline
		lineStr := string(bytes.TrimRight(line, "\n\r"))

		w.mu.Lock()
		w.lastLine = lineStr
		w.lines = append(w.lines, lineStr)
		if len(w.lines) > w.maxLines {
			w.lines = w.lines[len(w.lines)-w.maxLines:]
		}
		w.mu.Unlock()
	}

	return n, nil
}

// LastLine returns the last log line.
func (w *LogWriter) LastLine() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastLine
}

// Lines returns all buffered log lines.
func (w *LogWriter) Lines() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]string, len(w.lines))
	copy(out, w.lines)
	return out
}

// LastLines returns the last n log lines.
func (w *LogWriter) LastLines(n int) []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if n > len(w.lines) {
		n = len(w.lines)
	}
	out := make([]string, n)
	copy(out, w.lines[len(w.lines)-n:])
	return out
}

// GetCrashReport returns the recent log as a string for crash reporting.
func (w *LogWriter) GetCrashReport() string {
	lines := w.LastLines(50)
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line + "\n")
	}
	if buf.Len() == 0 {
		return "no crash data available"
	}
	return fmt.Sprintf("last %d lines:\n%s", len(lines), buf.String())
}
