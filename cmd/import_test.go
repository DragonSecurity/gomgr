package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
)

// newImportConfigDir builds a small but realistic config tree: an org.yaml with
// no rulesets block yet, and two team files, one declaring its repo as a bare
// permission string.
func newImportConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	teams := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teams, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	write(filepath.Join(dir, "app.yaml"), "org: myorg\n")
	write(filepath.Join(dir, "org.yaml"), "# Owners of record.\nowners:\n  - alice\n")
	write(filepath.Join(teams, "platform.yaml"), `name: Platform Team
slug: platform-team

repositories:
  # Bare permission string, the older shape.
  infra: maintain
`)
	write(filepath.Join(teams, "security.yaml"), `name: Security Team
slug: security-team

repositories:
  vulnerability-reports:
    permission: admin
    rulesets:
      - name: locked-down-main
        preset: strict-branch-protection
`)
	return dir
}

func adoptedSpec(name string) config.RulesetConfig {
	return config.RulesetConfig{
		Name:        name,
		Preset:      config.PresetBranchProtection,
		Enforcement: config.RulesetEnforcementActive,
	}
}

func TestWriteImportSplicesEveryScope(t *testing.T) {
	dir := newImportConfigDir(t)

	result := &insync.ImportResult{
		Scanned: 2,
		Org:     []insync.ImportedRuleset{{Spec: adoptedSpec("Protect Main")}},
		Repos: map[string][]insync.ImportedRuleset{
			"infra":                 {{Repo: "infra", Spec: adoptedSpec("infra-guard")}},
			"vulnerability-reports": {{Repo: "vulnerability-reports", Spec: adoptedSpec("extra-guard")}},
		},
	}

	if err := writeImport(dir, result); err != nil {
		t.Fatalf("write: %v", err)
	}

	// The whole directory must still load — that is what the next sync will do.
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("config no longer loads: %v", err)
	}

	if len(cfg.Org.Rulesets) != 1 || cfg.Org.Rulesets[0].Name != "Protect Main" {
		t.Errorf("org rulesets = %+v", cfg.Org.Rulesets)
	}

	byTeam := map[string]config.TeamConfig{}
	for _, team := range cfg.Team {
		byTeam[team.ResolvedSlug()] = team
	}

	infra, ok := byTeam["platform-team"].Repositories["infra"].(map[string]any)
	if !ok {
		t.Fatalf("infra is %T, want the bare string grown into a settings map", byTeam["platform-team"].Repositories["infra"])
	}
	if infra["permission"] != "maintain" {
		t.Errorf("infra permission = %v, want the original maintain", infra["permission"])
	}
	infraRulesets, err := config.ParseRulesets(infra["rulesets"])
	if err != nil || len(infraRulesets) != 1 || infraRulesets[0].Name != "infra-guard" {
		t.Errorf("infra rulesets = %+v (err %v)", infraRulesets, err)
	}

	vuln := byTeam["security-team"].Repositories["vulnerability-reports"].(map[string]any)
	vulnRulesets, err := config.ParseRulesets(vuln["rulesets"])
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(vulnRulesets) != 2 {
		t.Fatalf("got %d rulesets, want the existing one plus the adopted one: %+v", len(vulnRulesets), vulnRulesets)
	}
	if vulnRulesets[0].Name != "locked-down-main" || vulnRulesets[1].Name != "extra-guard" {
		t.Errorf("rulesets = %+v", vulnRulesets)
	}

	// Comments in the files that were edited must survive.
	orgText := readTestFile(t, filepath.Join(dir, "org.yaml"))
	if !strings.Contains(orgText, "# Owners of record.") {
		t.Errorf("org.yaml comment lost:\n%s", orgText)
	}
	platformText := readTestFile(t, filepath.Join(dir, "teams", "platform.yaml"))
	if !strings.Contains(platformText, "# Bare permission string, the older shape.") {
		t.Errorf("team file comment lost:\n%s", platformText)
	}
}

func TestWriteImportIsIdempotentAgainstItsOwnOutput(t *testing.T) {
	dir := newImportConfigDir(t)
	result := &insync.ImportResult{
		Scanned: 1,
		Org:     []insync.ImportedRuleset{{Spec: adoptedSpec("Protect Main")}},
	}
	if err := writeImport(dir, result); err != nil {
		t.Fatalf("first write: %v", err)
	}
	first := readTestFile(t, filepath.Join(dir, "org.yaml"))

	// A second import of the same live state finds it already declared and
	// writes nothing, so the file must be byte-identical.
	if err := writeImport(dir, &insync.ImportResult{Scanned: 1, AlreadyDeclared: 1}); err != nil {
		t.Fatalf("second write: %v", err)
	}
	if second := readTestFile(t, filepath.Join(dir, "org.yaml")); second != first {
		t.Errorf("re-import changed the file:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestWriteImportNothingToDo(t *testing.T) {
	dir := newImportConfigDir(t)
	before := readTestFile(t, filepath.Join(dir, "org.yaml"))
	if err := writeImport(dir, &insync.ImportResult{Scanned: 5, Repos: map[string][]insync.ImportedRuleset{}}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if after := readTestFile(t, filepath.Join(dir, "org.yaml")); after != before {
		t.Error("an empty import must not touch the configuration")
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func TestNewClientPrefersFlagsOverConfig(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_APP_ID", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "")

	origID, origKey := appIDFlag, privateKeyFlag
	t.Cleanup(func() { appIDFlag, privateKeyFlag = origID, origKey })

	cfg := &config.Root{App: config.AppConfig{Org: "myorg", AppID: 1, PrivateKey: "from-config.pem"}}
	appIDFlag = 99
	privateKeyFlag = "from-flag.pem"

	// The client build will fail (the key does not exist), but by then the
	// overrides have been applied, which is what this asserts.
	_, _ = newClient(context.Background(), cfg)

	if cfg.App.AppID != 99 {
		t.Errorf("AppID = %d, want the flag's 99", cfg.App.AppID)
	}
	if cfg.App.PrivateKey != "from-flag.pem" {
		t.Errorf("PrivateKey = %q, want the flag's value", cfg.App.PrivateKey)
	}
}

func TestNewClientLeavesConfigAloneWithoutFlags(t *testing.T) {
	origID, origKey := appIDFlag, privateKeyFlag
	t.Cleanup(func() { appIDFlag, privateKeyFlag = origID, origKey })
	appIDFlag, privateKeyFlag = 0, ""

	cfg := &config.Root{App: config.AppConfig{Org: "myorg", AppID: 7, PrivateKey: "from-config.pem"}}
	_, _ = newClient(context.Background(), cfg)

	if cfg.App.AppID != 7 || cfg.App.PrivateKey != "from-config.pem" {
		t.Errorf("app config = %+v, want it untouched when no flags are set", cfg.App)
	}
}
