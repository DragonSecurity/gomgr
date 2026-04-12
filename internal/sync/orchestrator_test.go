package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestBuildPlan_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return empty lists for all endpoints
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
	}

	plan, err := BuildPlan(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(plan.Changes))
	}
	if plan.Stats == nil {
		t.Fatal("expected stats to be populated")
	}
	if plan.Stats.Teams.Current != 0 || plan.Stats.Teams.Desired != 0 {
		t.Errorf("expected 0/0 teams, got %d/%d", plan.Stats.Teams.Current, plan.Stats.Teams.Desired)
	}
}

func TestBuildPlan_WithTeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/teams" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case r.URL.Path == "/orgs/myorg/repos" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "Backend", Slug: "backend"},
		},
	}

	plan, err := BuildPlan(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Stats.Teams.Desired != 1 {
		t.Errorf("expected 1 desired team, got %d", plan.Stats.Teams.Desired)
	}
	// Should have at least a team:create change
	found := false
	for _, ch := range plan.Changes {
		if ch.Scope == "team" && ch.Action == "create" {
			found = true
		}
	}
	if !found {
		t.Error("expected a team:create change")
	}
}

func TestApply_Empty(t *testing.T) {
	plan := util.Plan{Changes: []util.Change{}}
	err := Apply(context.Background(), nil, plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	plan := util.Plan{
		Changes: []util.Change{
			{Scope: "team", Target: "test", Action: "create", Details: map[string]any{"org": "myorg", "name": "test"}},
		},
	}
	err := Apply(ctx, nil, plan)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
