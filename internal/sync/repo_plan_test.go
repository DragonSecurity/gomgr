package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func planScopes(t *testing.T, cfg *config.Root, repos []map[string]any) []string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/repos" {
			_ = json.NewEncoder(w).Encode(repos)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer server.Close()

	plan, err := BuildPlan(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	var scopes []string
	for _, ch := range plan.Changes {
		scopes = append(scopes, ch.Scope)
	}
	return scopes
}

func contains(scopes []string, want string) bool {
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

// TestEntryWithoutPermissionStillGetsEverythingElse is the regression this
// refactor was prompted by.
//
// A guard added to stop gomgr granting the empty permission was written as a
// `continue`, which skipped the rest of the repository — its topics, pins,
// codeowners and files — rather than only the grant. Three hundred lines of one
// function is what let that read as correct.
func TestEntryWithoutPermissionStillGetsEverythingElse(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{
			Name: "Platform", Slug: "platform",
			Repositories: map[string]any{
				"infra": map[string]any{"topics": []any{"backend"}, "pinned": true},
			},
		}},
	}
	scopes := planScopes(t, cfg, []map[string]any{{"id": 1, "name": "infra", "topics": []string{}}})

	if !contains(scopes, "repo-topics") {
		t.Errorf("topics declared without a permission must still apply; got %v", scopes)
	}
	if !contains(scopes, "repo-pin") {
		t.Errorf("pinning declared without a permission must still apply; got %v", scopes)
	}
	if contains(scopes, "team-repo") {
		t.Errorf("no permission was declared, so no grant should be planned; got %v", scopes)
	}
}

func TestEntryWithPermissionGetsBoth(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{
			Name: "Platform", Slug: "platform",
			Repositories: map[string]any{
				"infra": map[string]any{"permission": "push", "topics": []any{"backend"}},
			},
		}},
	}
	scopes := planScopes(t, cfg, []map[string]any{{"id": 1, "name": "infra", "topics": []string{}}})
	if !contains(scopes, "team-repo") || !contains(scopes, "repo-topics") {
		t.Errorf("both the grant and the topics should be planned; got %v", scopes)
	}
}

func TestTopicsMatchingSuppressesTheChange(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{
			Name: "Platform", Slug: "platform",
			Repositories: map[string]any{
				"infra": map[string]any{"permission": "push", "topics": []any{"backend"}},
			},
		}},
	}
	// The repository already carries the wanted topic.
	scopes := planScopes(t, cfg, []map[string]any{{"id": 1, "name": "infra", "topics": []string{"backend"}}})
	if contains(scopes, "repo-topics") {
		t.Errorf("topics that already match must not be planned; got %v", scopes)
	}
}

func TestTopicsMatch(t *testing.T) {
	cases := []struct {
		current, want []string
		match         bool
	}{
		{nil, nil, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"b", "a"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a"}, []string{"a", "b"}, false},
	}
	for _, c := range cases {
		if got := topicsMatch(c.current, c.want); got != c.match {
			t.Errorf("topicsMatch(%v, %v) = %v, want %v", c.current, c.want, got, c.match)
		}
	}
}
