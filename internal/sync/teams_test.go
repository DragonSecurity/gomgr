package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
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
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "Backend", Slug: "backend", Description: "Backend team", Privacy: "closed"},
			{Name: "Frontend", Slug: "frontend", Description: "New desc", Privacy: "closed"},
			{Name: "Infra", Slug: "infra"},
		},
	}
	st := &State{
		Org: "myorg",
		ActualTeams: []*github.Team{
			{ID: github.Ptr(int64(1)), Slug: github.Ptr("backend"), Name: github.Ptr("Backend"), Description: github.Ptr("Backend team"), Privacy: github.Ptr("closed")},
			{ID: github.Ptr(int64(2)), Slug: github.Ptr("frontend"), Name: github.Ptr("Frontend"), Description: github.Ptr("Old desc"), Privacy: github.Ptr("closed")},
		},
	}

	changes, desired, err := planTeams(context.Background(), nil, cfg, st)
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
	st := &State{
		Org: "myorg",
		ActualRepos: []*github.Repository{
			{Name: github.Ptr("api"), Topics: []string{"backend"}},
		},
	}

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
	st := &State{
		Org:          "myorg",
		ManagedRepos: map[string]bool{"api": true},
		ActualTeams: []*github.Team{
			{ID: github.Ptr(int64(1)), Slug: github.Ptr("backend")},
			{ID: github.Ptr(int64(2)), Slug: github.Ptr("old-team")},
		},
		ActualRepos: []*github.Repository{
			{Name: github.Ptr("api")},
			{Name: github.Ptr("legacy-app")},
		},
	}

	changes, _, err := planCleanups(context.Background(), nil, cfg, st, desired)
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

func TestApplyChanges_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	changes := []util.Change{
		{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg", "name": "Backend"}},
	}

	err := applyChanges(ctx, nil, changes)
	if err == nil {
		t.Fatal("expected error for canceled context")
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

func TestParseRepoConfig_Codeowners(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []string
		wantErr bool
	}{
		{
			name: "list of usernames",
			input: map[string]any{
				"permission": "push",
				"codeowners": []any{"allanice001", "@octocat"},
			},
			want: []string{"allanice001", "@octocat"},
		},
		{
			name: "team ref",
			input: map[string]any{
				"codeowners": []any{"@dragon/platform"},
			},
			want: []string{"@dragon/platform"},
		},
		{
			name: "not a list",
			input: map[string]any{
				"codeowners": "octocat",
			},
			wantErr: true,
		},
		{
			name: "non-string entry",
			input: map[string]any{
				"codeowners": []any{"octocat", 123},
			},
			wantErr: true,
		},
		{
			name: "invalid entry",
			input: map[string]any{
				"codeowners": []any{"bad user"},
			},
			wantErr: true,
		},
		{
			name: "empty entry",
			input: map[string]any{
				"codeowners": []any{""},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRepoConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRepoConfig err=%v wantErr=%v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(got.codeowners) != len(tt.want) {
				t.Fatalf("codeowners=%v want %v", got.codeowners, tt.want)
			}
			for i, co := range got.codeowners {
				if co != tt.want[i] {
					t.Errorf("codeowners[%d]=%q want %q", i, co, tt.want[i])
				}
			}
		})
	}
}

func TestResolveTemplate_CodeownersUnion(t *testing.T) {
	all := map[string]repoSettings{
		"template-go-api": {
			permission: "push",
			template:   true,
			codeowners: []string{"allanice001"},
		},
		"my-api": {
			permission: "push",
			from:       "template-go-api",
			codeowners: []string{"octocat"},
		},
	}
	resolved, err := resolveTemplate("my-api", all["my-api"], all, "myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]bool{"allanice001": true, "octocat": true}
	if len(resolved.codeowners) != len(want) {
		t.Fatalf("codeowners=%v want union of both", resolved.codeowners)
	}
	for _, co := range resolved.codeowners {
		if !want[co] {
			t.Errorf("unexpected codeowner %q", co)
		}
	}
}

// TestParseRepoConfigRejectsUnknownKeys covers a misspelled setting.
//
// A live configuration wrote "permissions: admin" for a repository. The key was
// ignored, so no permission was declared, gomgr granted the empty string,
// GitHub read that as its default of read — and the admins team sat on read
// access to a repository the configuration said was admin, while gomgr
// re-granted it on every run and reported success.
func TestParseRepoConfigRejectsUnknownKeys(t *testing.T) {
	_, err := parseRepoConfig(map[string]any{"permissions": "admin"})
	if err == nil {
		t.Fatal("a key gomgr does not understand must not be ignored")
	}
	if !strings.Contains(err.Error(), `"permissions"`) {
		t.Errorf("error should name the offending key: %v", err)
	}
	if !strings.Contains(err.Error(), `did you mean "permission"`) {
		t.Errorf("error should suggest the intended key: %v", err)
	}
}

func TestParseRepoConfigAcceptsEveryKnownKey(t *testing.T) {
	settings, err := parseRepoConfig(map[string]any{
		"permission": "push",
		"topics":     []any{"backend"},
		"pinned":     true,
		"template":   false,
		"from":       "base",
		"visibility": "private",
		"codeowners": []any{"@octocat"},
		"rulesets":   []any{},
		"settings":   map[string]any{"allow_auto_merge": true},
	})
	if err != nil {
		t.Fatalf("every documented key must be accepted: %v", err)
	}
	if settings.permission != "push" || settings.visibility != "private" {
		t.Errorf("settings = %+v", settings)
	}
}

func TestNearestRepoKeySuggestsTheLikelySlip(t *testing.T) {
	for typo, want := range map[string]string{
		"permissions": "permission",
		"permision":   "permission",
		"topic":       "topics",
		"visibilty":   "visibility",
		"ruleset":     "rulesets",
	} {
		if got := config.NearestRepoKey(typo); got != want {
			t.Errorf("config.NearestRepoKey(%q) = %q, want %q", typo, got, want)
		}
	}
	// Something genuinely unrelated gets no misleading suggestion.
	if got := config.NearestRepoKey("elephant"); got != "" {
		t.Errorf("nearestRepoKey(elephant) = %q, want no suggestion", got)
	}
}

// TestNoGrantWithoutADeclaredPermission covers a repository entry that declares
// only topics. Granting the empty string is not "leave it alone": GitHub reads
// it as read, so the grant never converges.
func TestNoGrantWithoutADeclaredPermission(t *testing.T) {
	settings, err := parseRepoConfig(map[string]any{"topics": []any{"backend"}})
	if err != nil {
		t.Fatalf("topics without a permission is a legitimate entry: %v", err)
	}
	if settings.permission != "" {
		t.Fatalf("permission = %q, want empty", settings.permission)
	}
	// planRepoPerms skips the grant for this case; see the guard there.
}

// twoTeams builds a config where two teams name one repository. The team files
// are loaded in directory order, so "aaa" is folded in before "zzz".
func twoTeams(aaa, zzz any) *config.Root {
	return &config.Root{Team: []config.TeamConfig{
		{Name: "aaa", Repositories: map[string]any{"apikit": aaa}},
		{Name: "zzz", Repositories: map[string]any{"apikit": zzz}},
	}}
}

// TestPermissionIsPerTeamNotPerRepo covers a privilege escalation: repository
// settings are keyed by repository, so reading a grant's permission out of them
// gave every team naming a repository whatever the last team file to name it
// asked for. A team declaring pull was handed admin because an unrelated file
// sorted later.
func TestPermissionIsPerTeamNotPerRepo(t *testing.T) {
	cfg := twoTeams("pull", "admin")

	all, _, perTeam, err := collectRepoSettings(cfg, "acme")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	perms, err := resolveTeamPerms(perTeam, all, "acme")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got := perms["aaa/apikit"]; got != "pull" {
		t.Errorf("team aaa asked for pull, would be granted %q", got)
	}
	if got := perms["zzz/apikit"]; got != "admin" {
		t.Errorf("team zzz asked for admin, would be granted %q", got)
	}
}

// TestRepoDefinitionMergesAcrossTeams covers the other half: the repository-level
// map was overwritten rather than merged, so a repository defined by an earlier
// team silently lost its topics and visibility to a later team that only wanted
// a grant.
func TestRepoDefinitionMergesAcrossTeams(t *testing.T) {
	cfg := twoTeams(
		map[string]any{"permission": "pull", "topics": []any{"backend"}, "visibility": "private"},
		"admin",
	)

	all, _, _, err := collectRepoSettings(cfg, "acme")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	got := all["apikit"]
	if len(got.topics) != 1 || got.topics[0] != "backend" {
		t.Errorf("topics = %v, want [backend] to survive the later team", got.topics)
	}
	if got.visibility != "private" {
		t.Errorf("visibility = %q, want private to survive the later team", got.visibility)
	}
}

// TestRepoDefinitionUnionsSetFields checks the documented merge rule for the
// set-like fields: topics and codeowners union across teams.
func TestRepoDefinitionUnionsSetFields(t *testing.T) {
	cfg := twoTeams(
		map[string]any{"topics": []any{"backend"}, "codeowners": []any{"@a"}},
		map[string]any{"topics": []any{"api", "backend"}, "codeowners": []any{"@b", "@a"}},
	)

	all, _, _, err := collectRepoSettings(cfg, "acme")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	got := all["apikit"]
	if strings.Join(got.topics, ",") != "backend,api" {
		t.Errorf("topics = %v, want the union in first-seen order", got.topics)
	}
	if strings.Join(got.codeowners, ",") != "@a,@b" {
		t.Errorf("codeowners = %v, want the union in first-seen order", got.codeowners)
	}
}

// TestRepoDefinitionRejectsContradiction checks the scalar half of the merge
// rule. Only one of the two values can win and neither file says which, so the
// config is refused rather than resolved by filename order.
func TestRepoDefinitionRejectsContradiction(t *testing.T) {
	cases := map[string]*config.Root{
		"visibility": twoTeams(
			map[string]any{"visibility": "private"},
			map[string]any{"visibility": "public"},
		),
		"settings.allow_auto_merge": twoTeams(
			map[string]any{"settings": map[string]any{"allow_auto_merge": true}},
			map[string]any{"settings": map[string]any{"allow_auto_merge": false}},
		),
	}
	for field, cfg := range cases {
		t.Run(field, func(t *testing.T) {
			_, _, _, err := collectRepoSettings(cfg, "acme")
			if err == nil {
				t.Fatalf("two teams contradicting each other on %s must be refused", field)
			}
			if !strings.Contains(err.Error(), field) {
				t.Errorf("error should name the field in conflict, got: %v", err)
			}
			if !strings.Contains(err.Error(), "aaa") || !strings.Contains(err.Error(), "zzz") {
				t.Errorf("error should name both teams, got: %v", err)
			}
		})
	}
}

// TestDifferentPermissionsAreNotAConflict guards the merge from overreaching:
// two teams holding different access to one repository is the entire point of
// teams, and must not be mistaken for a contradiction.
func TestDifferentPermissionsAreNotAConflict(t *testing.T) {
	if _, _, _, err := collectRepoSettings(twoTeams("pull", "admin"), "acme"); err != nil {
		t.Fatalf("differing permissions are legitimate, got: %v", err)
	}
}

// TestReposFileDefinitionApplies covers the split: a definition in repos.yaml
// reaches the repository, while the team file keeps only the grant.
func TestReposFileDefinitionApplies(t *testing.T) {
	cfg := &config.Root{
		Repos: map[string]any{
			"apikit": map[string]any{
				"topics":     []any{"backend"},
				"visibility": "private",
			},
		},
		Team: []config.TeamConfig{
			{Name: "admins", Repositories: map[string]any{"apikit": "admin"}},
		},
	}

	all, managed, perTeam, err := collectRepoSettings(cfg, "acme")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	got := all["apikit"]
	if len(got.topics) != 1 || got.topics[0] != "backend" {
		t.Errorf("topics = %v, want [backend] from repos.yaml", got.topics)
	}
	if got.visibility != "private" {
		t.Errorf("visibility = %q, want private from repos.yaml", got.visibility)
	}
	if !managed["apikit"] {
		t.Error("a repo the team names should be managed")
	}
	perms, err := resolveTeamPerms(perTeam, all, "acme")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if perms["admins/apikit"] != "admin" {
		t.Errorf("grant = %q, want admin from the team file", perms["admins/apikit"])
	}
}

// TestReposFileOverlapIsRefused covers the migration rule. Both places are
// honored so a config can move one repository at a time, but defining one in
// both is refused rather than resolved by precedence — a precedence rule being
// the thing the split exists to remove.
func TestReposFileOverlapIsRefused(t *testing.T) {
	cfg := &config.Root{
		Repos: map[string]any{
			"apikit": map[string]any{"topics": []any{"backend"}},
		},
		Team: []config.TeamConfig{
			{Name: "admins", Repositories: map[string]any{
				"apikit": map[string]any{"permission": "admin", "visibility": "private"},
			}},
		},
	}
	_, _, _, err := collectRepoSettings(cfg, "acme")
	if err == nil {
		t.Fatal("a repo defined in repos.yaml and a team file must be refused")
	}
	if !strings.Contains(err.Error(), "repos.yaml") || !strings.Contains(err.Error(), "admins") {
		t.Errorf("error should name both places, got: %v", err)
	}
}

// A team entry that says only which permission the team holds is a grant, not a
// definition, so it never collides with repos.yaml.
func TestReposFilePermissionOnlyEntryIsNotAnOverlap(t *testing.T) {
	cfg := &config.Root{
		Repos: map[string]any{"apikit": map[string]any{"topics": []any{"backend"}}},
		Team: []config.TeamConfig{
			{Name: "a", Repositories: map[string]any{"apikit": "pull"}},
			{Name: "b", Repositories: map[string]any{"apikit": map[string]any{"permission": "admin"}}},
		},
	}
	if _, _, _, err := collectRepoSettings(cfg, "acme"); err != nil {
		t.Fatalf("permission-only entries are grants, not definitions: %v", err)
	}
}

// TestSpecsForOverridesByPath covers the whole point of the exercise: a
// repository states its own version of an org-wide file by naming itself, with
// no ordering rule and no `only:` filter involved.
func TestSpecsForOverridesByPath(t *testing.T) {
	p := &repoPlanner{
		fileSpecs: []config.FileSpec{
			{Path: ".github/renovate.json", Content: "generic", Reconcile: true},
			{Path: "LICENSE", Content: "MIT"},
		},
		settings: map[string]repoSettings{
			"apikit": {files: []config.FileSpec{
				{Path: ".github/renovate.json", Content: "special", Reconcile: true},
				{Path: "CODEOWNERS", Content: "@team"},
			}},
			"other": {},
		},
	}

	got := map[string]string{}
	for _, s := range p.specsFor("apikit") {
		got[s.Path] = s.Content
	}
	if got[".github/renovate.json"] != "special" {
		t.Errorf("renovate.json = %q, want the repository's own version", got[".github/renovate.json"])
	}
	if got["LICENSE"] != "MIT" {
		t.Errorf("LICENSE = %q, want the org-wide entry to survive", got["LICENSE"])
	}
	if got["CODEOWNERS"] != "@team" {
		t.Errorf("CODEOWNERS = %q, want a repo-only file to be added", got["CODEOWNERS"])
	}

	// Order is the org-wide order, so the override does not move anything.
	paths := []string{}
	for _, s := range p.specsFor("apikit") {
		paths = append(paths, s.Path)
	}
	if strings.Join(paths, ",") != ".github/renovate.json,LICENSE,CODEOWNERS" {
		t.Errorf("order = %v, want the org-wide order with repo-only files appended", paths)
	}

	// A repository that overrides nothing is untouched.
	if len(p.specsFor("other")) != 2 {
		t.Errorf("a repo with no files of its own should get the org-wide list unchanged")
	}
}

// A team's notification setting has to reach the create, and has to stop
// re-planning once GitHub agrees — a comparison that always differs would
// re-send the same update on every run forever.
func TestPlanTeams_NotificationSetting(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "New", Slug: "new", NotificationSetting: config.NotificationsDisabled},
			{Name: "Agrees", Slug: "agrees", NotificationSetting: config.NotificationsDisabled},
			{Name: "Differs", Slug: "differs", NotificationSetting: config.NotificationsDisabled},
			{Name: "Silent", Slug: "silent"},
		},
	}
	st := &State{
		Org: "myorg",
		ActualTeams: []*github.Team{
			{Slug: github.Ptr("agrees"), Name: github.Ptr("Agrees"), NotificationSetting: github.Ptr(config.NotificationsDisabled)},
			{Slug: github.Ptr("differs"), Name: github.Ptr("Differs"), NotificationSetting: github.Ptr(config.NotificationsEnabled)},
			{Slug: github.Ptr("silent"), Name: github.Ptr("Silent"), NotificationSetting: github.Ptr(config.NotificationsEnabled)},
		},
	}

	changes, _, err := planTeams(context.Background(), nil, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	byTarget := map[string]map[string]any{}
	for _, ch := range changes {
		d, ok := ch.Details.(map[string]any)
		if !ok {
			t.Fatalf("change for %s has details of type %T", ch.Target, ch.Details)
		}
		byTarget[ch.Target] = d
	}

	if got := byTarget["new"]["notification_setting"]; got != config.NotificationsDisabled {
		t.Errorf("create did not carry the notification setting: %v", got)
	}
	if _, replanned := byTarget["agrees"]; replanned {
		t.Error("planned an update for a team GitHub already agrees with")
	}
	if got := byTarget["differs"]["notification_setting"]; got != config.NotificationsDisabled {
		t.Errorf("update did not carry the notification setting: %v", got)
	}
	// A config that says nothing must not fight whatever the org has set.
	if _, touched := byTarget["silent"]; touched {
		t.Error("an unset notification_setting still planned a change")
	}
}

// A pending invitation means the person has already been asked. Planning the
// same add again re-sends the invitation on every run, which is how a config
// naming someone who never accepts turns into a weekly reminder from GitHub.
func TestPlanTeamMembershipSkipsPendingInvitations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		switch {
		case r.URL.Path == "/orgs/myorg/teams/backend/members":
			_ = enc.Encode([]map[string]any{})
		case r.URL.Path == "/orgs/myorg/teams/backend/invitations":
			_ = enc.Encode([]map[string]any{{"login": "Slowpoke"}})
		case strings.HasPrefix(r.URL.Path, "/users/"):
			_ = enc.Encode(map[string]any{"login": strings.TrimPrefix(r.URL.Path, "/users/")})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	st := &State{Org: "myorg"}
	desired := map[string]config.TeamConfig{
		"backend": {Name: "Backend", Slug: "backend", Members: []string{"slowpoke", "eager"}},
	}

	changes, err := planTeamMembership(context.Background(), newTestClient(t, server), st, desired)
	if err != nil {
		t.Fatalf("planTeamMembership: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected only the uninvited member to be planned, got %+v", changes)
	}
	if d, ok := changes[0].Details.(teamMemberChange); !ok || d.User != "eager" {
		t.Errorf("expected eager to be added, got %+v", changes[0].Details)
	}
	if len(st.RepoWarnings) != 1 || !strings.Contains(st.RepoWarnings[0], "slowpoke") {
		t.Errorf("the pending invitation should be reported, got %v", st.RepoWarnings)
	}
}

// Someone who is already a member at the wrong role is corrected even if an
// invitation is somehow also on file — the membership is the stronger fact.
func TestPlanTeamMembershipStillFixesRoleOfExistingMember(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		switch {
		case r.URL.Path == "/orgs/myorg/teams/backend/members" && r.URL.Query().Get("role") == "member":
			_ = enc.Encode([]map[string]any{{"login": "alice"}})
		case r.URL.Path == "/orgs/myorg/teams/backend/members":
			_ = enc.Encode([]map[string]any{})
		case r.URL.Path == "/orgs/myorg/teams/backend/invitations":
			_ = enc.Encode([]map[string]any{{"login": "alice"}})
		case strings.HasPrefix(r.URL.Path, "/users/"):
			_ = enc.Encode(map[string]any{"login": strings.TrimPrefix(r.URL.Path, "/users/")})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	desired := map[string]config.TeamConfig{
		"backend": {Name: "Backend", Slug: "backend", Maintainers: []string{"alice"}},
	}

	changes, err := planTeamMembership(context.Background(), newTestClient(t, server), &State{Org: "myorg"}, desired)
	if err != nil {
		t.Fatalf("planTeamMembership: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected the role correction to survive, got %+v", changes)
	}
	if d := changes[0].Details.(teamMemberChange); d.Role != roleMaintainer {
		t.Errorf("expected a promotion to maintainer, got %q", d.Role)
	}
}
