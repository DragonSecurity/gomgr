package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// AuditLog controls whether structured audit log entries are emitted to stderr.
var AuditLog bool

func EnableDebug() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Warnf prints a warning message to stderr
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
