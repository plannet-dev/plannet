package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	// Log levels
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents the logger instance
type Logger struct {
	out       io.Writer
	level     Level
	fields    map[string]interface{}
	mu        sync.Mutex
	useColors bool
}

// New creates a new logger instance
func New(out io.Writer, level Level, useColors bool) *Logger {
	return &Logger{
		out:       out,
		level:     level,
		fields:    make(map[string]interface{}),
		useColors: useColors,
	}
}

// DefaultLogger is the default logger instance
var DefaultLogger = New(os.Stderr, InfoLevel, true)

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		out:       l.out,
		level:     l.level,
		fields:    make(map[string]interface{}),
		useColors: l.useColors,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		out:       l.out,
		level:     l.level,
		fields:    make(map[string]interface{}),
		useColors: l.useColors,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithContext adds context values to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract relevant context values and add them as fields
	newLogger := l.WithFields(extractContextFields(ctx))
	return newLogger
}

// log performs the actual logging
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	// Create the log entry
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level.String(),
		"message":   fmt.Sprintf(msg, args...),
		"caller":    fmt.Sprintf("%s:%d", file, line),
	}

	// Add fields
	for k, v := range l.fields {
		entry[k] = v
	}

	// Marshal to JSON
	output, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.out, "Error marshaling log entry: %v\n", err)
		return
	}

	// Write the log entry
	if l.useColors {
		color := getColorForLevel(level)
		fmt.Fprintf(l.out, "%s%s\n%s", color, output, "\033[0m")
	} else {
		fmt.Fprintf(l.out, "%s\n", output)
	}

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DebugLevel, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(InfoLevel, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WarnLevel, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ErrorLevel, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FatalLevel, msg, args...)
}

// Helper functions for the default logger
func Debug(msg string, args ...interface{}) {
	DefaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	DefaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	DefaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	DefaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	DefaultLogger.Fatal(msg, args...)
}

// WithField adds a field to the default logger
func WithField(key string, value interface{}) *Logger {
	return DefaultLogger.WithField(key, value)
}

// WithFields adds multiple fields to the default logger
func WithFields(fields map[string]interface{}) *Logger {
	return DefaultLogger.WithFields(fields)
}

// WithContext adds context values to the default logger
func WithContext(ctx context.Context) *Logger {
	return DefaultLogger.WithContext(ctx)
}

// SetLevel sets the logging level for the default logger
func SetLevel(level Level) {
	DefaultLogger.level = level
}

// Helper functions

// extractContextFields extracts relevant fields from context
func extractContextFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})

	// Add trace ID if present
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}

	// Add request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	return fields
}

// getColorForLevel returns the ANSI color code for a log level
func getColorForLevel(level Level) string {
	switch level {
	case DebugLevel:
		return "\033[36m" // Cyan
	case InfoLevel:
		return "\033[32m" // Green
	case WarnLevel:
		return "\033[33m" // Yellow
	case ErrorLevel:
		return "\033[31m" // Red
	case FatalLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}
