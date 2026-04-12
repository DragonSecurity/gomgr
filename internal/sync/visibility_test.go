package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestParseRepoConfig_Visibility(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{"public", map[string]any{"permission": "push", "visibility": "public"}, "public", false},
		{"private", map[string]any{"permission": "push", "visibility": "private"}, "private", false},
		{"internal", map[string]any{"permission": "push", "visibility": "internal"}, "internal", false},
		{"absent stays empty", map[string]any{"permission": "push"}, "", false},
		{"invalid string", map[string]any{"permission": "push", "visibility": "secret"}, "", true},
		{"wrong type", map[string]any{"permission": "push", "visibility": 42}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := parseRepoConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRepoConfig() err=%v, wantErr=%v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if s.visibility != tt.want {
				t.Errorf("visibility = %q, want %q", s.visibility, tt.want)
			}
		})
	}
}

func TestApplyRepoEnsure_VisibilityPublic(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/orgs/myorg/repos" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "public-docs"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo",
		Target: "public-docs",
		Action: "ensure",
		Details: map[string]any{
			"org":        "myorg",
			"name":       "public-docs",
			"visibility": "public",
		},
	}
	if err := applyRepoEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["visibility"] != "public" {
		t.Errorf("expected visibility=public in request, got %v", gotBody["visibility"])
	}
	// The GitHub REST API also accepts `private`; for a public repo we send false.
	if gotBody["private"] != false {
		t.Errorf("expected private=false, got %v", gotBody["private"])
	}
}

func TestApplyRepoEnsure_VisibilityInternalMakesItNonPrivate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 2})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:   "repo",
		Target:  "internal-api",
		Action:  "ensure",
		Details: map[string]any{"org": "myorg", "name": "internal-api", "visibility": "internal"},
	}
	if err := applyRepoEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["visibility"] != "internal" {
		t.Errorf("expected visibility=internal, got %v", gotBody["visibility"])
	}
	if gotBody["private"] != false {
		t.Errorf("expected private=false for internal repo, got %v", gotBody["private"])
	}
}

func TestApplyRepoEnsure_DefaultsToPrivate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 3})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:   "repo",
		Target:  "legacy",
		Action:  "ensure",
		Details: map[string]any{"org": "myorg", "name": "legacy", "private": true},
	}
	if err := applyRepoEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["private"] != true {
		t.Errorf("expected private=true, got %v", gotBody["private"])
	}
	if _, ok := gotBody["visibility"]; ok {
		t.Errorf("expected no visibility in legacy payload, got %v", gotBody["visibility"])
	}
}
