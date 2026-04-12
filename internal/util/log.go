package util

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// AuditLog controls whether structured audit log entries are emitted to stderr.
var AuditLog bool

var (
	levelVar = new(slog.LevelVar)
	logger   = newLogger(false)
)

func newLogger(addSource bool) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: addSource,
	}))
}

// Logger returns the package-level structured logger.
func Logger() *slog.Logger { return logger }

// EnableDebug switches the package logger to debug level with source info.
func EnableDebug() {
	levelVar.Set(slog.LevelDebug)
	logger = newLogger(true)
	slog.SetDefault(logger)
}

// Infof emits a formatted info-level log line via slog.
func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Debugf emits a formatted debug-level log line via slog.
func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Warnf prints a warning message to stderr.
func Warnf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", v...)
}

// Audit emits a structured JSON audit log entry to stderr when AuditLog is enabled.
func Audit(scope, target, action, status string) {
	if !AuditLog {
		return
	}
	entry := map[string]string{
		"ts":     time.Now().UTC().Format(time.RFC3339),
		"scope":  scope,
		"target": target,
		"action": action,
		"status": status,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit marshal error: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}
