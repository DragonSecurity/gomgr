package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
)

func TestParseRepoConfig(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		wantPerm   string
		wantTopics []string
		wantPinned bool
		wantError  bool
	}{
		{
			name:       "simple string permission",
			input:      "push",
			wantPerm:   "push",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "advanced config with permission only",
			input: map[string]any{
				"permission": "maintain",
			},
			wantPerm:   "maintain",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "advanced config with topics",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{"backend", "api"},
			},
			wantPerm:   "push",
			wantTopics: []string{"backend", "api"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "advanced config with pinning",
			input: map[string]any{
				"permission": "admin",
				"topics":     []any{"documentation"},
				"pinned":     true,
			},
			wantPerm:   "admin",
			wantTopics: []string{"documentation"},
			wantPinned: true,
			wantError:  false,
		},
		{
			name: "map[any]any format (YAML unmarshal variant)",
			input: map[any]any{
				"permission": "pull",
				"topics":     []any{"frontend", "web"},
				"pinned":     false,
			},
			wantPerm:   "pull",
			wantTopics: []string{"frontend", "web"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "missing permission field (topics only)",
			input: map[string]any{
				"topics": []any{"backend"},
			},
			wantPerm:   "",
			wantTopics: []string{"backend"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "empty topics array",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{},
			},
			wantPerm:   "push",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "non-string values in topics array (should be ignored)",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{"valid", 123, "another"},
			},
			wantPerm:   "push",
			wantTopics: []string{"valid", "another"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name:       "empty string permission",
			input:      "",
			wantPerm:   "",
			wantTopics: nil,
			wantPinned: false,
			wantError:  true,
		},
		{
			name: "permission as non-string type",
			input: map[string]any{
				"permission": 123,
			},
			wantPerm:   "",
			wantTopics: nil,
			wantPinned: false,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return // Skip validation if error was expected
			}

			if settings.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", settings.permission, tt.wantPerm)
			}

			if len(settings.topics) != len(tt.wantTopics) {
				t.Errorf("topics length = %d, want %d", len(settings.topics), len(tt.wantTopics))
			} else {
				for i, topic := range settings.topics {
					if topic != tt.wantTopics[i] {
						t.Errorf("topics[%d] = %q, want %q", i, topic, tt.wantTopics[i])
					}
				}
			}

			if settings.pinned != tt.wantPinned {
				t.Errorf("pinned = %v, want %v", settings.pinned, tt.wantPinned)
			}
		})
	}
}

func TestParseRepoConfigWithTemplate(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		wantPerm     string
		wantTopics   []string
		wantPinned   bool
		wantTemplate bool
		wantFrom     string
		wantError    bool
	}{
		{
			name: "template repository",
			input: map[string]any{
				"permission": "push",
				"template":   true,
				"topics":     []any{"backend", "template"},
			},
			wantPerm:     "push",
			wantTopics:   []string{"backend", "template"},
			wantPinned:   false,
			wantTemplate: true,
			wantFrom:     "",
			wantError:    false,
		},
		{
			name: "repository using template (same org)",
			input: map[string]any{
				"permission": "push",
				"from":       "template-go-api",
				"topics":     []any{"my-project"},
			},
			wantPerm:     "push",
			wantTopics:   []string{"my-project"},
			wantPinned:   false,
			wantTemplate: false,
			wantFrom:     "template-go-api",
			wantError:    false,
		},
		{
			name: "repository using template (cross-org)",
			input: map[string]any{
				"from":   "some-org/template-repo",
				"topics": []any{"backend"},
			},
			wantPerm:     "",
			wantTopics:   []string{"backend"},
			wantPinned:   false,
			wantTemplate: false,
			wantFrom:     "some-org/template-repo",
			wantError:    false,
		},
		{
			name: "template with from (both should work)",
			input: map[string]any{
				"permission": "admin",
				"template":   true,
				"from":       "another-template",
			},
			wantPerm:     "admin",
			wantTopics:   nil,
			wantPinned:   false,
			wantTemplate: true,
			wantFrom:     "another-template",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return
			}

			if settings.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", settings.permission, tt.wantPerm)
			}

			if len(settings.topics) != len(tt.wantTopics) {
				t.Errorf("topics length = %d, want %d", len(settings.topics), len(tt.wantTopics))
			} else {
				for i, topic := range settings.topics {
					if topic != tt.wantTopics[i] {
						t.Errorf("topics[%d] = %q, want %q", i, topic, tt.wantTopics[i])
					}
				}
			}

			if settings.pinned != tt.wantPinned {
				t.Errorf("pinned = %v, want %v", settings.pinned, tt.wantPinned)
			}

			if settings.template != tt.wantTemplate {
				t.Errorf("template = %v, want %v", settings.template, tt.wantTemplate)
			}

			if settings.from != tt.wantFrom {
				t.Errorf("from = %q, want %q", settings.from, tt.wantFrom)
			}
		})
	}
}

func TestResolveTemplate(t *testing.T) {
	tests := []struct {
		name          string
		repoName      string
		settings      repoSettings
		allRepos      map[string]repoSettings
		defaultOrg    string
		wantPerm      string
		wantTopics    []string
		wantError     bool
		errorContains string
	}{
		{
			name:     "no template reference",
			repoName: "api",
			settings: repoSettings{
				permission: "push",
				topics:     []string{"backend"},
			},
			allRepos:   map[string]repoSettings{},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend"},
			wantError:  false,
		},
		{
			name:     "inherit from template",
			repoName: "my-api",
			settings: repoSettings{
				from:   "template-go-api",
				topics: []string{"my-project"},
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend", "api"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend", "api", "my-project"},
			wantError:  false,
		},
		{
			name:     "override permission from template",
			repoName: "my-api",
			settings: repoSettings{
				permission: "admin",
				from:       "template-go-api",
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "admin",
			wantTopics: []string{"backend"},
			wantError:  false,
		},
		{
			name:     "template not found",
			repoName: "my-api",
			settings: repoSettings{
				from: "nonexistent-template",
			},
			allRepos:      map[string]repoSettings{},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "not found",
		},
		{
			name:     "referenced repo not marked as template",
			repoName: "my-api",
			settings: repoSettings{
				from: "regular-repo",
			},
			allRepos: map[string]repoSettings{
				"regular-repo": {
					permission: "push",
					template:   false,
				},
			},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "not marked with template: true",
		},
		{
			name:     "cross-org template reference",
			repoName: "my-api",
			settings: repoSettings{
				from: "other-org/template-repo",
			},
			allRepos:      map[string]repoSettings{},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "cross-organization template references not yet supported",
		},
		{
			name:     "deduplicate topics",
			repoName: "my-api",
			settings: repoSettings{
				from:   "template-go-api",
				topics: []string{"backend", "my-service"},
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend", "api"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend", "api", "my-service"},
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveTemplate(tt.repoName, tt.settings, tt.allRepos, tt.defaultOrg)

			if (err != nil) != tt.wantError {
				t.Errorf("resolveTemplate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if result.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", result.permission, tt.wantPerm)
			}

			if len(result.topics) != len(tt.wantTopics) {
				t.Errorf("topics = %v, want %v", result.topics, tt.wantTopics)
			} else {
				for i, topic := range result.topics {
					if topic != tt.wantTopics[i] {
						t.Errorf("topics[%d] = %q, want %q", i, topic, tt.wantTopics[i])
					}
				}
			}
		})
	}
}

func TestValidateTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "valid topic",
			topic:     "backend",
			wantError: false,
		},
		{
			name:      "valid topic with hyphens",
			topic:     "my-project-backend",
			wantError: false,
		},
		{
			name:      "valid topic with numbers",
			topic:     "project123",
			wantError: false,
		},
		{
			name:      "empty topic",
			topic:     "",
			wantError: true,
		},
		{
			name:      "topic too long (>50 chars)",
			topic:     "this-is-a-very-long-topic-name-that-exceeds-fifty-characters-limit",
			wantError: true,
		},
		{
			name:      "topic starting with hyphen",
			topic:     "-invalid",
			wantError: true,
		},
		{
			name:      "topic with uppercase",
			topic:     "Backend",
			wantError: true,
		},
		{
			name:      "topic with underscore",
			topic:     "my_project",
			wantError: true,
		},
		{
			name:      "topic with space",
			topic:     "my project",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTopic(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("validateTopic() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNormalizePermission(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Standard permissions
		{
			name:  "pull permission",
			input: "pull",
			want:  "pull",
		},
		{
			name:  "read permission (normalized to pull)",
			input: "read",
			want:  "pull",
		},
		{
			name:  "triage permission",
			input: "triage",
			want:  "triage",
		},
		{
			name:  "push permission",
			input: "push",
			want:  "push",
		},
		{
			name:  "write permission (normalized to push)",
			input: "write",
			want:  "push",
		},
		{
			name:  "maintain permission",
			input: "maintain",
			want:  "maintain",
		},
		{
			name:  "admin permission",
			input: "admin",
			want:  "admin",
		},
		// Case insensitive
		{
			name:  "uppercase PUSH",
			input: "PUSH",
			want:  "push",
		},
		{
			name:  "mixed case Admin",
			input: "Admin",
			want:  "admin",
		},
		// Custom repository roles (GitHub Enterprise Cloud)
		{
			name:  "custom role: actions-manager",
			input: "actions-manager",
			want:  "actions-manager",
		},
		{
			name:  "custom role: release-manager",
			input: "release-manager",
			want:  "release-manager",
		},
		{
			name:  "custom role: runner-admin",
			input: "runner-admin",
			want:  "runner-admin",
		},
		{
			name:  "custom role: security-scanner",
			input: "security-scanner",
			want:  "security-scanner",
		},
		{
			name:  "custom role with mixed case (preserved)",
			input: "Custom-Role-Name",
			want:  "Custom-Role-Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePermission(tt.input)
			if got != tt.want {
				t.Errorf("normalizePermission(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRepoConfigWithCustomRoles(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantPerm  string
		wantError bool
	}{
		{
			name:      "custom role as simple string",
			input:     "actions-manager",
			wantPerm:  "actions-manager",
			wantError: false,
		},
		{
			name: "custom role in advanced config",
			input: map[string]any{
				"permission": "release-manager",
				"topics":     []any{"cicd", "releases"},
			},
			wantPerm:  "release-manager",
			wantError: false,
		},
		{
			name: "custom role with hyphens",
			input: map[string]any{
				"permission": "github-actions-admin",
				"topics":     []any{"cicd"},
			},
			wantPerm:  "github-actions-admin",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return
			}

			if settings.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", settings.permission, tt.wantPerm)
			}
		})
	}
}

func TestContainsErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		errResp     *github.ErrorResponse
		searchTerms []string
		want        bool
	}{
		{
			name: "message in main Message field",
			errResp: &github.ErrorResponse{
				Message: `"sha" wasn't supplied`,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in main Message field without quotes",
			errResp: &github.ErrorResponse{
				Message: `sha wasn't supplied`,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in Errors array",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: `"sha" wasn't supplied`},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in Errors array among multiple errors",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: "some other error"},
					{Message: `"sha" wasn't supplied`},
					{Message: "yet another error"},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "reference already exists in main Message",
			errResp: &github.ErrorResponse{
				Message: "reference already exists",
			},
			searchTerms: []string{"reference already exists"},
			want:        true,
		},
		{
			name: "reference already exists in Errors array",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: "reference already exists"},
				},
			},
			searchTerms: []string{"reference already exists"},
			want:        true,
		},
		{
			name: "partial match should fail",
			errResp: &github.ErrorResponse{
				Message: "sha is required",
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
		{
			name: "no match in Message or Errors",
			errResp: &github.ErrorResponse{
				Message: "some other error",
				Errors: []github.Error{
					{Message: "different error"},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
		{
			name: "empty ErrorResponse",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors:  nil,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsErrorMessage(tt.errResp, tt.searchTerms...)
			if got != tt.want {
				t.Errorf("containsErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---- Planning Function Tests ----

func TestPlanTeams(t *testing.T) {
	// Server returns 2 existing teams: "backend" and "frontend"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/teams" && r.Method == "GET" {
			teams := []map[string]any{
				{"id": 1, "slug": "backend", "name": "Backend", "description": "Backend team", "privacy": "closed"},
				{"id": 2, "slug": "frontend", "name": "Frontend", "description": "Old desc", "privacy": "closed"},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(teams)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "Backend", Slug: "backend", Description: "Backend team", Privacy: "closed"},
			{Name: "Frontend", Slug: "frontend", Description: "New desc", Privacy: "closed"},
			{Name: "Infra", Slug: "infra"},
		},
	}
	st := &State{Org: "myorg"}

	changes, desired, err := planTeams(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(desired) != 3 {
		t.Errorf("expected 3 desired teams, got %d", len(desired))
	}

	var creates, updates int
	for _, ch := range changes {
		switch ch.Action {
		case "create":
			creates++
			if ch.Target != "infra" {
				t.Errorf("expected create for infra, got %s", ch.Target)
			}
		case "update":
			updates++
			if ch.Target != "frontend" {
				t.Errorf("expected update for frontend, got %s", ch.Target)
			}
		}
	}
	if creates != 1 {
		t.Errorf("expected 1 create, got %d", creates)
	}
	if updates != 1 {
		t.Errorf("expected 1 update, got %d", updates)
	}
}

func TestPlanTeamMembership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/teams/backend/members" && r.URL.Query().Get("role") == "maintainer":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"login": "alice"},
			})
		case r.URL.Path == "/orgs/myorg/teams/backend/members" && r.URL.Query().Get("role") == "member":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"login": "bob"},
			})
		case r.URL.Path == "/users/alice" || r.URL.Path == "/users/bob" || r.URL.Path == "/users/charlie":
			_ = json.NewEncoder(w).Encode(map[string]any{"login": strings.TrimPrefix(r.URL.Path, "/users/")})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	st := &State{Org: "myorg"}
	desiredBySlug := map[string]config.TeamConfig{
		"backend": {
			Name:        "Backend",
			Slug:        "backend",
			Maintainers: []string{"alice"},
			Members:     []string{"charlie"}, // bob removed, charlie added
		},
	}

	changes, err := planTeamMembership(context.Background(), c, st, desiredBySlug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have a change for charlie (new member)
	found := false
	for _, ch := range changes {
		if ch.Scope == "team-member" && ch.Action == "ensure" {
			d := ch.Details.(teamMemberChange)
			if d.User == "charlie" && d.Role == "member" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected team-member:ensure change for charlie")
	}
}

func TestPlanRepoPerms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/repos" && r.Method == "GET":
			// Return one existing repo
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"name": "api", "topics": []string{"backend"}},
			})
		case strings.HasPrefix(r.URL.Path, "/orgs/myorg/teams/") && strings.HasSuffix(r.URL.Path, "/repos"):
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg", CreateRepo: true},
		Team: []config.TeamConfig{
			{
				Name: "Backend",
				Slug: "backend",
				Repositories: map[string]any{
					"api": map[string]any{
						"permission": "push",
						"topics":     []any{"backend", "go"},
					},
					"new-service": "maintain",
				},
			},
		},
	}
	st := &State{Org: "myorg"}

	changes, err := planRepoPerms(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var repoEnsures, teamRepoGrants, topicChanges int
	for _, ch := range changes {
		switch ch.Scope + ":" + ch.Action {
		case "repo:ensure":
			repoEnsures++
		case "team-repo:grant":
			teamRepoGrants++
		case "repo-topics:ensure":
			topicChanges++
		}
	}
	if repoEnsures != 1 {
		t.Errorf("expected 1 repo:ensure (new-service), got %d", repoEnsures)
	}
	if teamRepoGrants != 2 {
		t.Errorf("expected 2 team-repo:grant, got %d", teamRepoGrants)
	}
	if topicChanges != 1 {
		t.Errorf("expected 1 repo-topics:ensure, got %d", topicChanges)
	}
}

func TestPlanCleanups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/teams" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 1, "slug": "backend"},
				{"id": 2, "slug": "old-team"},
			})
		case r.URL.Path == "/orgs/myorg/repos" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"name": "api"},
				{"name": "legacy-app"},
			})
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{
			Org:                     "myorg",
			DeleteUnconfiguredTeams: true,
			DeleteUnmanagedRepos:    true,
		},
	}
	desired := map[string]config.TeamConfig{
		"backend": {Name: "Backend", Slug: "backend"},
	}
	st := &State{Org: "myorg", ManagedRepos: map[string]bool{"api": true}}

	changes, _, err := planCleanups(context.Background(), c, cfg, st, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var teamDeletes, repoDeletes int
	for _, ch := range changes {
		switch ch.Scope + ":" + ch.Action {
		case "team:delete":
			teamDeletes++
			if ch.Target != "old-team" {
				t.Errorf("expected delete for old-team, got %s", ch.Target)
			}
		case "repo:delete":
			repoDeletes++
			if ch.Target != "legacy-app" {
				t.Errorf("expected delete for legacy-app, got %s", ch.Target)
			}
		}
	}
	if teamDeletes != 1 {
		t.Errorf("expected 1 team:delete, got %d", teamDeletes)
	}
	if repoDeletes != 1 {
		t.Errorf("expected 1 repo:delete, got %d", repoDeletes)
	}
}

func TestApplyChanges_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	changes := []util.Change{
		{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg", "name": "Backend"}},
	}

	err := applyChanges(ctx, nil, changes)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestPlanCustomRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/custom-repository-roles" && r.Method == "GET" {
			resp := map[string]any{
				"total_count": 1,
				"custom_roles": []map[string]any{
					{
						"id":          1,
						"name":        "deployer",
						"description": "Old desc",
						"base_role":   "read",
						"permissions": []string{"manage_actions"},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Org: config.OrgConfig{
			CustomRoles: []config.CustomRoleConfig{
				{Name: "deployer", Description: "Updated desc", BaseRole: "read", Permissions: []string{"manage_actions"}},
				{Name: "release-manager", Description: "New role", BaseRole: "write", Permissions: []string{"create_releases"}},
			},
		},
	}
	st := &State{Org: "myorg"}

	changes, err := planCustomRoles(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var creates, updates int
	for _, ch := range changes {
		switch ch.Action {
		case "create":
			creates++
		case "update":
			updates++
		}
	}
	if creates != 1 {
		t.Errorf("expected 1 custom-role:create, got %d", creates)
	}
	if updates != 1 {
		t.Errorf("expected 1 custom-role:update, got %d", updates)
	}
}
