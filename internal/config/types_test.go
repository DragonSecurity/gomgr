package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRepoConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantName string
		wantType string // "simple" or "advanced"
	}{
		{
			name: "simple string permission",
			yaml: `
name: Test Team
repositories:
  simple-repo: push
`,
			wantName: "simple-repo",
			wantType: "simple",
		},
		{
			name: "advanced config with topics",
			yaml: `
name: Test Team
repositories:
  advanced-repo:
    permission: maintain
    topics:
      - backend
      - project-test
`,
			wantName: "advanced-repo",
			wantType: "advanced",
		},
		{
			name: "pinned repository",
			yaml: `
name: Test Team
repositories:
  test-index:
    permission: admin
    topics:
      - documentation
    pinned: true
`,
			wantName: "test-index",
			wantType: "advanced",
		},
		{
			name: "mixed configuration",
			yaml: `
name: Test Team
repositories:
  simple: push
  advanced:
    permission: admin
    topics: [backend]
    pinned: true
`,
			wantName: "advanced",
			wantType: "advanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var team TeamConfig
			if err := yaml.Unmarshal([]byte(tt.yaml), &team); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if team.Name != "Test Team" {
				t.Errorf("Expected team name 'Test Team', got %q", team.Name)
			}

			if len(team.Repositories) == 0 {
				t.Fatal("No repositories found")
			}

			val, ok := team.Repositories[tt.wantName]
			if !ok {
				t.Fatalf("Repository %q not found", tt.wantName)
			}

			switch tt.wantType {
			case "simple":
				if _, ok := val.(string); !ok {
					t.Errorf("Expected string type, got %T", val)
				}
			case "advanced":
				if _, ok := val.(map[string]any); !ok {
					t.Errorf("Expected map[string]any type, got %T", val)
				}
			}
		})
	}
}

func TestAppConfigParsing(t *testing.T) {
	yamlData := `
org: TestOrg
dry_warnings:
  warn_unmanaged_teams: true
  warn_members_without_any_team: true
  warn_unmanaged_repos: true
remove_members_without_team: false
delete_unconfigured_teams: false
delete_unmanaged_repos: true
create_repo: true
`

	var app AppConfig
	if err := yaml.Unmarshal([]byte(yamlData), &app); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if app.Org != "TestOrg" {
		t.Errorf("Expected org 'TestOrg', got %q", app.Org)
	}

	if !app.DryWarnings.WarnUnmanagedRepos {
		t.Error("Expected WarnUnmanagedRepos to be true")
	}

	if !app.DeleteUnmanagedRepos {
		t.Error("Expected DeleteUnmanagedRepos to be true")
	}

	if !app.CreateRepo {
		t.Error("Expected CreateRepo to be true")
	}
}
