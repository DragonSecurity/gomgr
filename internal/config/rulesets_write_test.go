package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	return string(b)
}

// adopted is a small ruleset standing in for something the importer produced.
func adopted(name string) RulesetConfig {
	return RulesetConfig{
		Name:        name,
		Preset:      PresetBranchProtection,
		Enforcement: RulesetEnforcementActive,
	}
}

// assertStillValid reparses the edited file to prove the splice produced YAML
// that still loads, and that nothing was lost on the way.
func assertStillValid(t *testing.T, path string, out any) {
	t.Helper()
	if err := readYAML(path, out); err != nil {
		t.Fatalf("edited file no longer parses: %v\n---\n%s", err, readFile(t, path))
	}
}

func TestInsertOrgRulesetsCreatesBlock(t *testing.T) {
	path := writeTemp(t, "org.yaml", `# Organization owners.
owners:
  - alice
  - bob

# Custom roles need Enterprise Cloud.
custom_roles:
  - name: release-manager
    base_role: write
`)

	if err := InsertOrgRulesets(path, []RulesetConfig{adopted("Protect Main")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got := readFile(t, path)
	for _, want := range []string{
		"# Organization owners.",
		"# Custom roles need Enterprise Cloud.",
		adoptedComment,
		"rulesets:",
		"- name: Protect Main",
		"preset: branch-protection",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}

	var org OrgConfig
	assertStillValid(t, path, &org)
	if len(org.Rulesets) != 1 || org.Rulesets[0].Name != "Protect Main" {
		t.Errorf("reparsed rulesets = %+v", org.Rulesets)
	}
	if len(org.Owners) != 2 || len(org.CustomRoles) != 1 {
		t.Errorf("existing content was disturbed: owners=%v roles=%+v", org.Owners, org.CustomRoles)
	}
}

func TestInsertOrgRulesetsAppendsToExistingBlock(t *testing.T) {
	path := writeTemp(t, "org.yaml", `owners:
  - alice

rulesets:
  # The baseline everyone inherits.
  - name: existing-guard-rail
    preset: no-force-push

# A trailing comment that must survive.
custom_roles: []
`)

	if err := InsertOrgRulesets(path, []RulesetConfig{adopted("Protect Main")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "# The baseline everyone inherits.") {
		t.Errorf("in-block comment lost:\n%s", got)
	}
	if !strings.Contains(got, "# A trailing comment that must survive.") {
		t.Errorf("trailing comment lost:\n%s", got)
	}

	var org OrgConfig
	assertStillValid(t, path, &org)
	if len(org.Rulesets) != 2 {
		t.Fatalf("got %d rulesets, want 2: %+v", len(org.Rulesets), org.Rulesets)
	}
	if org.Rulesets[0].Name != "existing-guard-rail" || org.Rulesets[1].Name != "Protect Main" {
		t.Errorf("wrong order or content: %+v", org.Rulesets)
	}

	// The new entry must line up with the one already there.
	lines := strings.Split(got, "\n")
	var indents []int
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimLeft(l, " "), "- name:") {
			indents = append(indents, lineIndent(l))
		}
	}
	if len(indents) != 2 || indents[0] != indents[1] {
		t.Errorf("item indents = %v, want two matching values", indents)
	}
}

func TestInsertOrgRulesetsHandlesDashAtKeyIndent(t *testing.T) {
	// Sequence items at the same column as their key are legal YAML and appear
	// in plenty of hand-written configs.
	path := writeTemp(t, "org.yaml", `owners:
- alice

rulesets:
- name: existing
  preset: no-force-push
`)

	if err := InsertOrgRulesets(path, []RulesetConfig{adopted("Protect Main")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var org OrgConfig
	assertStillValid(t, path, &org)
	if len(org.Rulesets) != 2 {
		t.Fatalf("got %d rulesets, want 2:\n%s", len(org.Rulesets), readFile(t, path))
	}
}

func TestInsertRepoRulesetsIntoSettingsMap(t *testing.T) {
	path := writeTemp(t, "team.yaml", `name: Platform Team
slug: platform-team

repositories:
  # Core infrastructure.
  infra:
    permission: maintain
    topics:
      - infrastructure

  docs: pull
`)

	if err := InsertRepoRulesets(path, "repositories", "infra", []RulesetConfig{adopted("Protect Main")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "# Core infrastructure.") {
		t.Errorf("comment lost:\n%s", got)
	}

	var team TeamConfig
	assertStillValid(t, path, &team)
	settings, ok := team.Repositories["infra"].(map[string]any)
	if !ok {
		t.Fatalf("infra is %T, want a settings map:\n%s", team.Repositories["infra"], got)
	}
	if settings["permission"] != "maintain" {
		t.Errorf("permission = %v, want maintain", settings["permission"])
	}
	if topics, ok := settings["topics"].([]any); !ok || len(topics) != 1 {
		t.Errorf("topics were disturbed: %v", settings["topics"])
	}
	rulesets, err := ParseRulesets(settings["rulesets"])
	if err != nil || len(rulesets) != 1 || rulesets[0].Name != "Protect Main" {
		t.Errorf("rulesets = %+v (err %v)\n%s", rulesets, err, got)
	}
	// The sibling entry must be untouched.
	if team.Repositories["docs"] != "pull" {
		t.Errorf("docs = %v, want the untouched string pull", team.Repositories["docs"])
	}
}

func TestInsertRepoRulesetsGrowsBarePermission(t *testing.T) {
	path := writeTemp(t, "team.yaml", `name: Platform Team

repositories:
  infra: push
  docs: pull
`)

	if err := InsertRepoRulesets(path, "repositories", "infra", []RulesetConfig{adopted("Protect Main")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var team TeamConfig
	assertStillValid(t, path, &team)
	settings, ok := team.Repositories["infra"].(map[string]any)
	if !ok {
		t.Fatalf("infra is %T, want the bare string grown into a map:\n%s",
			team.Repositories["infra"], readFile(t, path))
	}
	if settings["permission"] != "push" {
		t.Errorf("permission = %v, want the original push", settings["permission"])
	}
	rulesets, err := ParseRulesets(settings["rulesets"])
	if err != nil || len(rulesets) != 1 {
		t.Errorf("rulesets = %+v (err %v)", rulesets, err)
	}
	if team.Repositories["docs"] != "pull" {
		t.Errorf("docs = %v, want untouched", team.Repositories["docs"])
	}
}

func TestInsertRepoRulesetsAppendsToExistingBlock(t *testing.T) {
	path := writeTemp(t, "team.yaml", `name: Security Team

repositories:
  vulnerability-reports:
    permission: admin
    rulesets:
      - name: locked-down-main
        preset: strict-branch-protection

  other: pull
`)

	if err := InsertRepoRulesets(path, "repositories", "vulnerability-reports", []RulesetConfig{adopted("Protect Tags")}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var team TeamConfig
	assertStillValid(t, path, &team)
	settings := team.Repositories["vulnerability-reports"].(map[string]any)
	rulesets, err := ParseRulesets(settings["rulesets"])
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rulesets) != 2 {
		t.Fatalf("got %d rulesets, want 2:\n%s", len(rulesets), readFile(t, path))
	}
	if rulesets[0].Name != "locked-down-main" || rulesets[1].Name != "Protect Tags" {
		t.Errorf("rulesets = %+v", rulesets)
	}
	if team.Repositories["other"] != "pull" {
		t.Errorf("sibling entry disturbed: %v", team.Repositories["other"])
	}
}

func TestInsertRepoRulesetsUnknownRepo(t *testing.T) {
	path := writeTemp(t, "team.yaml", "name: T\nrepositories:\n  infra: push\n")
	err := InsertRepoRulesets(path, "repositories", "nope", []RulesetConfig{adopted("x")})
	if err == nil || !strings.Contains(err.Error(), "not declared here") {
		t.Fatalf("err = %v, want a complaint that the repo is not declared", err)
	}
}

func TestInsertRulesetsIsANoopForNothing(t *testing.T) {
	const src = "owners:\n  - alice\n"
	path := writeTemp(t, "org.yaml", src)
	if err := InsertOrgRulesets(path, nil); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if got := readFile(t, path); got != src {
		t.Errorf("file changed for an empty insert:\n%s", got)
	}
}

func TestFindTeamFileForRepo(t *testing.T) {
	dir := t.TempDir()
	teams := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teams, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(teams, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	write("platform.yaml", "name: Platform\nrepositories:\n  infra: maintain\n")
	write("security.yaml", "name: Security\nrepositories:\n  Vuln-Reports: admin\n")
	write("notes.txt", "ignored")

	t.Run("finds the declaring file", func(t *testing.T) {
		got, err := FindTeamFileForRepo(dir, "infra")
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		if filepath.Base(got) != "platform.yaml" {
			t.Errorf("got %q, want platform.yaml", got)
		}
	})

	t.Run("matches case-insensitively", func(t *testing.T) {
		got, err := FindTeamFileForRepo(dir, "vuln-reports")
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		if filepath.Base(got) != "security.yaml" {
			t.Errorf("got %q, want security.yaml", got)
		}
	})

	t.Run("reports nothing for an undeclared repo", func(t *testing.T) {
		got, err := FindTeamFileForRepo(dir, "ghost")
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want no file", got)
		}
	})
}

func TestInsertedRulesetsPassValidation(t *testing.T) {
	path := writeTemp(t, "org.yaml", "owners:\n  - alice\n")
	if err := InsertOrgRulesets(path, []RulesetConfig{
		adopted("Protect Main"),
		{
			Name:        "custom",
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementEvaluate,
			Rules: RulesetRules{
				RequiredStatusChecks: &RequiredStatusChecksRule{
					Strict: true,
					Checks: []StatusCheck{{Context: "build"}},
				},
			},
		},
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var org OrgConfig
	assertStillValid(t, path, &org)
	if err := ValidateRulesets(ScopeOrg, "org.yaml", org.Rulesets); err != nil {
		t.Errorf("written rulesets do not validate: %v\n%s", err, readFile(t, path))
	}
}
