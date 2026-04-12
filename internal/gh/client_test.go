package gh

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeReadPEM_InlineKey(t *testing.T) {
	inline := "-----BEGIN RSA PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END RSA PRIVATE KEY-----\n"
	b, err := maybeReadPEM(inline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != inline {
		t.Errorf("got %q, want inline key back", string(b))
	}
}

func TestMaybeReadPEM_InlineGarbage(t *testing.T) {
	_, err := maybeReadPEM("BEGIN but not a real pem")
	if err == nil {
		t.Fatal("expected error for malformed inline key")
	}
	if !strings.Contains(err.Error(), "invalid PEM") {
		t.Errorf("expected 'invalid PEM' in error, got: %v", err)
	}
}

func TestMaybeReadPEM_WrongBlockType(t *testing.T) {
	cert := "-----BEGIN CERTIFICATE-----\nMIIBogIBAAJBALRiMLAH\n-----END CERTIFICATE-----\n"
	_, err := maybeReadPEM(cert)
	if err == nil {
		t.Fatal("expected error for non-private-key block")
	}
	if !strings.Contains(err.Error(), "expected a private key block") {
		t.Errorf("expected block-type error, got: %v", err)
	}
}

func TestMaybeReadPEM_PKCS8(t *testing.T) {
	pkcs8 := "-----BEGIN PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END PRIVATE KEY-----\n"
	if _, err := maybeReadPEM(pkcs8); err != nil {
		t.Fatalf("unexpected error for PKCS#8 block: %v", err)
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

func newGraphQLClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		httpClient: srv.Client(),
		GraphQLURL: srv.URL,
	}, srv
}

func TestDoGraphQL_Success(t *testing.T) {
	var gotBody map[string]any
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type=application/json, got %q", ct)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"viewer":{"login":"octocat"}}}`))
	})

	var out struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}
	err := c.DoGraphQL(context.Background(), "query { viewer { login } }", map[string]any{"x": 1}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Viewer.Login != "octocat" {
		t.Errorf("expected login=octocat, got %q", out.Viewer.Login)
	}
	if gotBody["query"] == nil {
		t.Error("expected request body to include query")
	}
	if gotBody["variables"] == nil {
		t.Error("expected request body to include variables")
	}
}

func TestDoGraphQL_OmitsEmptyVariables(t *testing.T) {
	var gotBody map[string]any
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})

	if err := c.DoGraphQL(context.Background(), "query { viewer { login } }", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := gotBody["variables"]; present {
		t.Error("expected variables to be omitted when empty")
	}
}

func TestDoGraphQL_HTTPError(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "status 401") {
		t.Errorf("expected 'status 401' in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Bad credentials") {
		t.Errorf("expected upstream message to be surfaced, got %v", err)
	}
}

func TestDoGraphQL_GraphQLErrors(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errors":[{"message":"field missing"},{"message":"another issue"}]}`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for graphql errors")
	}
	if !strings.Contains(err.Error(), "field missing") || !strings.Contains(err.Error(), "another issue") {
		t.Errorf("expected both error messages to be joined, got %v", err)
	}
}

func TestDoGraphQL_MalformedJSON(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": not-json`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "decode graphql response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestDoGraphQL_NilResultSkipsDataUnmarshal(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"anything":123}}`))
	})

	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err != nil {
		t.Errorf("expected nil error when result is nil, got %v", err)
	}
}

func TestDoGraphQL_EmptyQueryRejected(t *testing.T) {
	c := &Client{httpClient: &http.Client{}}
	if err := c.DoGraphQL(context.Background(), "   ", nil, nil); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestDoGraphQL_NilClientRejected(t *testing.T) {
	var c *Client
	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestDoGraphQL_NilHTTPClientRejected(t *testing.T) {
	c := &Client{}
	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err == nil {
		t.Fatal("expected error for nil httpClient")
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
