package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

var (
	log       zerolog.Logger
	mu        atomic.Uint32 // 0=uninitialized, 1=initialized
	logFile   *os.File
	writer    io.Writer = os.Stderr
	onceGuard           = make(chan struct{}, 1)
)

func init() {
	onceGuard <- struct{}{}
	Init(Config{})
	<-onceGuard
}

// Config holds logger configuration.
type Config struct {
	Level     string // debug, info, warn, error
	JSON      bool   // JSON output format
	File      string // log file path (empty = stderr)
	AddSource bool   // include caller source info
}

// Init initializes the logger with the given config. Can be called multiple
// times to reconfigure. Safe for concurrent use.
func Init(cfg Config) {
	<-onceGuard
	defer func() { onceGuard <- struct{}{} }()

	// Close previous log file if any
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	// Parse level
	level := zerolog.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	// Output writer
	writer = os.Stderr
	if cfg.File != "" {
		f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			logFile = f
			writer = f
		}
	}

	// Format
	var output io.Writer = writer
	if !cfg.JSON {
		output = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
	}

	l := zerolog.New(output).Level(level)
	if cfg.AddSource {
		l = l.With().Timestamp().Caller().Logger()
	} else {
		l = l.With().Timestamp().Logger()
	}
	log = l
	mu.Store(1)
}

// Close cleans up resources.
func Close() {
	<-onceGuard
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	onceGuard <- struct{}{}
}

// ─── Public API ─────────────────────────────────────────────────────────

// Debug logs a debug message.
func Debug(msg string) {
	if mu.Load() == 0 {
		return
	}
	log.Debug().Msg(msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	if mu.Load() == 0 {
		return
	}
	log.Debug().Msg(fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string) {
	if mu.Load() == 0 {
		return
	}
	log.Info().Msg(msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	if mu.Load() == 0 {
		return
	}
	log.Info().Msg(fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	if mu.Load() == 0 {
		return
	}
	log.Warn().Msg(msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	if mu.Load() == 0 {
		return
	}
	log.Warn().Msg(fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msg string) {
	if mu.Load() == 0 {
		return
	}
	log.Error().Msg(msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	if mu.Load() == 0 {
		return
	}
	log.Error().Msg(fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits.
func Fatal(msg string) {
	if mu.Load() == 0 {
		os.Exit(1)
		return
	}
	log.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits.
func Fatalf(format string, args ...any) {
	if mu.Load() == 0 {
		os.Exit(1)
		return
	}
	log.Fatal().Msg(fmt.Sprintf(format, args...))
}

// ─── With Fields ────────────────────────────────────────────────────────

// WithField returns a logger with a key-value field attached.
func WithField(key string, val any) zerolog.Logger {
	if mu.Load() == 0 {
		return zerolog.New(io.Discard)
	}
	return log.With().Interface(key, val).Logger()
}

// WithError returns a logger with an error field attached.
func WithError(err error) zerolog.Logger {
	if mu.Load() == 0 {
		return zerolog.New(io.Discard)
	}
	return log.With().Err(err).Logger()
}

// ─── Log Adaptor ─────────────────────────────────────────────────────────

// Printf implements a log.Printf-compatible interface using zerolog Infof.
func Printf(format string, args ...any) {
	Infof(format, args...)
}

// Println implements a log.Println-compatible interface using zerolog Info.
func Println(args ...any) {
	Info(strings.TrimRight(fmt.Sprint(args...), "\n"))
}
