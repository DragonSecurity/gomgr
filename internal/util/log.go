package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// AuditLog controls whether structured audit log entries are emitted to stderr.
var AuditLog bool

var (
	levelVar   = new(slog.LevelVar)
	showSource bool
	logger     = slog.New(newSimpleHandler(os.Stdout, levelVar))
)

// Logger returns the package-level structured logger.
func Logger() *slog.Logger { return logger }

// EnableDebug switches the package logger to debug level and prefixes records
// with file:line so operators can trace noisy runs.
func EnableDebug() {
	levelVar.Set(slog.LevelDebug)
	showSource = true
	slog.SetDefault(logger)
}

// simpleHandler renders slog records as plain text without the key=value
// prefix slog.TextHandler produces. Info records are emitted as bare messages
// so CLI output stays readable; warn/error/debug get a short label.
type simpleHandler struct {
	out   io.Writer
	level slog.Leveler
	mu    sync.Mutex
}

func newSimpleHandler(out io.Writer, level slog.Leveler) *simpleHandler {
	return &simpleHandler{out: out, level: level}
}

func (h *simpleHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *simpleHandler) Handle(_ context.Context, r slog.Record) error {
	var prefix string
	switch r.Level {
	case slog.LevelDebug:
		prefix = "DEBUG: "
	case slog.LevelWarn:
		prefix = "WARN: "
	case slog.LevelError:
		prefix = "ERROR: "
	}

	var suffix string
	if showSource && r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		if frame, _ := frames.Next(); frame.File != "" {
			suffix = fmt.Sprintf(" (%s:%d)", filepath.Base(frame.File), frame.Line)
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := fmt.Fprintln(h.out, prefix+r.Message+suffix)
	return err
}

func (h *simpleHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *simpleHandler) WithGroup(_ string) slog.Handler      { return h }

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
