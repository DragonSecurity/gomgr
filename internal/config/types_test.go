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

func TestOrgConfigParsingWithCustomRoles(t *testing.T) {
	yamlData := `
owners:
  - alice
  - bob
custom_roles:
  - name: actions-manager
    description: Manage GitHub Actions without code access
    base_role: read
    permissions:
      - manage_actions
      - manage_runners
  - name: release-manager
    description: Manage releases and deployments
    base_role: write
    permissions:
      - create_releases
      - manage_environments
`

	var org OrgConfig
	if err := yaml.Unmarshal([]byte(yamlData), &org); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(org.Owners) != 2 {
		t.Errorf("Expected 2 owners, got %d", len(org.Owners))
	}

	if len(org.CustomRoles) != 2 {
		t.Fatalf("Expected 2 custom roles, got %d", len(org.CustomRoles))
	}

	// Check first role
	role1 := org.CustomRoles[0]
	if role1.Name != "actions-manager" {
		t.Errorf("Expected name 'actions-manager', got %q", role1.Name)
	}
	if role1.BaseRole != "read" {
		t.Errorf("Expected base_role 'read', got %q", role1.BaseRole)
	}
	if len(role1.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(role1.Permissions))
	}

	// Check second role
	role2 := org.CustomRoles[1]
	if role2.Name != "release-manager" {
		t.Errorf("Expected name 'release-manager', got %q", role2.Name)
	}
	if role2.BaseRole != "write" {
		t.Errorf("Expected base_role 'write', got %q", role2.BaseRole)
	}
}

