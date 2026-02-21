package sync

import (
	"strings"
	"testing"

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
