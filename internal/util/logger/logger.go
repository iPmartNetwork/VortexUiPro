// Package logger provides structured logging with dual-backend (console + file),
// log rotation, and in-memory buffer for web UI retrieval.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	DEBUG   LogLevel = 0
	INFO    LogLevel = 1
	NOTICE  LogLevel = 2
	WARNING LogLevel = 3
	ERROR   LogLevel = 4
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case NOTICE:
		return "NOTICE"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

const (
	maxLogBufferSize = 10240
	timeFormat       = "2006/01/02 15:04:05"
	maxLogFileMB     = 10
	maxLogBackups    = 5
	maxLogAgeDays    = 7
)

// logEntry holds a single buffered log entry.
type logEntry struct {
	time  string
	level LogLevel
	log   string
}

var (
	mu             sync.Mutex
	minLevel       LogLevel = INFO
	writers        []io.Writer
	logBuffer      []logEntry
	logBufferMu    sync.Mutex
	fileWriter     io.WriteCloser
)

// Init initialises the logger with the given minimum level.
// Call once at startup; safe to call multiple times (closes previous file writer).
func Init(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	minLevel = level

	// Close previous file writer before replacing
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}

	writers = append(writers[:0], os.Stdout)

	// File writer with rotation
	logDir := "."
	if runtime.GOOS != "windows" {
		logDir = "/var/log/vortexuipro"
	}
	_ = os.MkdirAll(logDir, 0o750)
	logPath := filepath.Join(logDir, "vortexuipro.log")

	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640); err == nil {
		fileWriter = &rotateWriter{file: f, path: logPath, maxSize: maxLogFileMB * 1024 * 1024}
		writers = append(writers, fileWriter)
	}

	log.SetFlags(0)
	log.SetOutput(io.MultiWriter(writers...))
}

// Close flushes and closes the file writer.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
}

// Debug logs a debug message.
func Debug(args ...any) {
	write(DEBUG, fmt.Sprint(args...))
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	write(DEBUG, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(args ...any) {
	write(INFO, fmt.Sprint(args...))
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	write(INFO, fmt.Sprintf(format, args...))
}

// Notice logs a notice message.
func Notice(args ...any) {
	write(NOTICE, fmt.Sprint(args...))
}

// Noticef logs a formatted notice message.
func Noticef(format string, args ...any) {
	write(NOTICE, fmt.Sprintf(format, args...))
}

// Warning logs a warning message.
func Warning(args ...any) {
	write(WARNING, fmt.Sprint(args...))
}

// Warningf logs a formatted warning message.
func Warningf(format string, args ...any) {
	write(WARNING, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(args ...any) {
	write(ERROR, fmt.Sprint(args...))
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	write(ERROR, fmt.Sprintf(format, args...))
}

// write logs a message at the given level.
func write(level LogLevel, msg string) {
	if level < minLevel {
		return
	}
	now := time.Now().Format(timeFormat)
	line := fmt.Sprintf("%s %s - %s", now, level, msg)
	log.Println(line)

	// Buffer for web UI retrieval
	logBufferMu.Lock()
	if len(logBuffer) >= maxLogBufferSize {
		logBuffer = logBuffer[1:]
	}
	logBuffer = append(logBuffer, logEntry{time: now, level: level, log: msg})
	logBufferMu.Unlock()
}

// GetLogs retrieves up to n log entries at or below the given level.
func GetLogs(n int, level LogLevel) []string {
	logBufferMu.Lock()
	snapshot := make([]logEntry, len(logBuffer))
	copy(snapshot, logBuffer)
	logBufferMu.Unlock()

	var out []string
	for i := len(snapshot) - 1; i >= 0 && len(out) < n; i-- {
		if snapshot[i].level <= level {
			out = append(out, fmt.Sprintf("%s %s - %s", snapshot[i].time, snapshot[i].level, snapshot[i].log))
		}
	}
	return out
}

// ─── Rotating File Writer ────────────────────────────────────────────

type rotateWriter struct {
	mu      sync.Mutex
	file    *os.File
	path    string
	maxSize int64
	size    int64
}

func (w *rotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.size+int64(len(p)) >= w.maxSize {
		_ = w.rotate()
	}
	w.size += int64(len(p))
	return w.file.Write(p)
}

func (w *rotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func (w *rotateWriter) rotate() error {
	_ = w.file.Close()
	for i := maxLogBackups - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", w.path, i)
		newer := fmt.Sprintf("%s.%d", w.path, i+1)
		_ = os.Rename(old, newer)
	}
	_ = os.Rename(w.path, w.path+".1")

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return err
	}
	w.file = f
	w.size = 0
	return nil
}

