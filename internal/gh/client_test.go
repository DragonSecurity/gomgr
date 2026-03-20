package gh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeReadPEM_InlineKey(t *testing.T) {
	inline := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	b, err := maybeReadPEM(inline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != inline {
		t.Errorf("got %q, want inline key back", string(b))
	}
}

func TestMaybeReadPEM_FromFile(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END RSA PRIVATE KEY-----\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(path, []byte(pem), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := maybeReadPEM(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != pem {
		t.Errorf("got %q, want file contents", string(b))
	}
}

func TestMaybeReadPEM_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(path, []byte("not a pem"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := maybeReadPEM(path)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
	if !strings.Contains(err.Error(), "invalid PEM") {
		t.Errorf("expected 'invalid PEM' in error, got: %v", err)
	}
}

func TestMaybeReadPEM_MissingFile(t *testing.T) {
	_, err := maybeReadPEM("/nonexistent/path/key.pem")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"hello", "world", "hello"},
		{"", "world", "world"},
		{"", "", ""},
		{"a", "", "a"},
	}
	for _, tt := range tests {
		got := firstNonEmpty(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("firstNonEmpty(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}
