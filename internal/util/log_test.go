package util

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestWarnf(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Warnf("something %s happened", "bad")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "WARNING:") {
		t.Errorf("expected WARNING: prefix, got %q", output)
	}
	if !strings.Contains(output, "something bad happened") {
		t.Errorf("expected formatted message, got %q", output)
	}
}

func TestAudit_Enabled(t *testing.T) {
	oldVal := AuditLog
	AuditLog = true
	defer func() { AuditLog = oldVal }()

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Audit("team", "backend", "create", "ok")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	var entry map[string]string
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, output)
	}
	if entry["scope"] != "team" {
		t.Errorf("expected scope=team, got %q", entry["scope"])
	}
	if entry["target"] != "backend" {
		t.Errorf("expected target=backend, got %q", entry["target"])
	}
	if entry["action"] != "create" {
		t.Errorf("expected action=create, got %q", entry["action"])
	}
	if entry["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", entry["status"])
	}
	if entry["ts"] == "" {
		t.Error("expected non-empty ts field")
	}
}

func TestAudit_Disabled(t *testing.T) {
	oldVal := AuditLog
	AuditLog = false
	defer func() { AuditLog = oldVal }()

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Audit("team", "backend", "create", "ok")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() != 0 {
		t.Errorf("expected no output when audit disabled, got %q", buf.String())
	}
}

func TestEnableDebug(t *testing.T) {
	oldOutput := log.Writer()
	oldFlags := log.Flags()
	defer func() {
		log.SetOutput(oldOutput)
		log.SetFlags(oldFlags)
	}()

	EnableDebug()

	if log.Flags()&log.Lshortfile == 0 {
		t.Error("expected Lshortfile flag after EnableDebug")
	}
}
