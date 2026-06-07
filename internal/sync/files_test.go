package sync

import (
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func TestMaterializeFileSpecs_LegacyFlags(t *testing.T) {
	app := config.AppConfig{
		AddDefaultReadme:  true,
		AddRenovateConfig: true,
		RenovateConfig:    `{"extends":["x"]}`,
	}
	files := materializeFileSpecs(app)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "README.md" {
		t.Errorf("expected README.md first, got %q", files[0].Path)
	}
	if files[1].Path != ".github/renovate.json" {
		t.Errorf("expected renovate.json second, got %q", files[1].Path)
	}
}

func TestMaterializeFileSpecs_SkipsRenovateWhenContentEmpty(t *testing.T) {
	app := config.AppConfig{AddRenovateConfig: true, RenovateConfig: "   "}
	files := materializeFileSpecs(app)
	if len(files) != 0 {
		t.Errorf("expected empty list when renovate content is blank, got %d entries", len(files))
	}
}

func TestMaterializeFileSpecs_UserOverridesLegacy(t *testing.T) {
	app := config.AppConfig{
		AddDefaultReadme: true,
		Files: []config.FileSpec{
			{Path: "README.md", Content: "# custom readme for {{.Repo}}", Message: "chore: README", Branch: "main"},
		},
	}
	files := materializeFileSpecs(app)
	if len(files) != 1 {
		t.Fatalf("expected 1 file after dedup, got %d", len(files))
	}
	if !strings.Contains(files[0].Content, "custom readme") {
		t.Errorf("expected user override to win, got %q", files[0].Content)
	}
}

func TestPlanRepoFiles_RendersAndDedupes(t *testing.T) {
	specs := []config.FileSpec{
		{Path: "README.md", Content: "# {{.Repo}} in {{.Org}}", Message: "chore: readme", Branch: "main"},
		{Path: "LICENSE", Content: "MIT\n"},
	}
	emitted := map[string]bool{}

	changes, err := planRepoFiles("Acme", "widgets", "widgets", specs, emitted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	readmeDetails := changes[0].Details.(map[string]any)
	if readmeDetails["content"] != "# widgets in Acme" {
		t.Errorf("expected rendered README content, got %q", readmeDetails["content"])
	}
	if readmeDetails["message"] != "chore: readme" {
		t.Errorf("expected custom commit message, got %q", readmeDetails["message"])
	}
	if readmeDetails["branch"] != "main" {
		t.Errorf("expected branch main, got %q", readmeDetails["branch"])
	}

	licenseDetails := changes[1].Details.(map[string]any)
	if licenseDetails["message"] != "chore: add LICENSE" {
		t.Errorf("expected default commit message, got %q", licenseDetails["message"])
	}
	if licenseDetails["branch"] != "main" {
		t.Errorf("expected default branch main, got %q", licenseDetails["branch"])
	}

	// Calling again should be a no-op because emitted tracks both paths now.
	more, err := planRepoFiles("Acme", "widgets", "widgets", specs, emitted)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if len(more) != 0 {
		t.Errorf("expected no new changes on second call, got %d", len(more))
	}
}

func TestPlanRepoFiles_OnlyGlobSkipsNonMatch(t *testing.T) {
	specs := []config.FileSpec{
		{Path: "LICENSE", Content: "MIT", Only: []string{"public-*"}},
	}
	changes, err := planRepoFiles("Acme", "internal-api", "internal-api", specs, map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes for non-matching repo, got %d", len(changes))
	}

	changes, err = planRepoFiles("Acme", "public-docs", "public-docs", specs, map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change for matching repo, got %d", len(changes))
	}
}

func TestPlanRepoFiles_BadTemplatePropagates(t *testing.T) {
	specs := []config.FileSpec{{Path: "bad.md", Content: "{{.Missing}}"}}
	_, err := planRepoFiles("Acme", "widgets", "widgets", specs, map[string]bool{})
	if err == nil {
		t.Fatal("expected template error")
	}
}

func TestMaterializeFileSpecs_PreservesUserOrder(t *testing.T) {
	app := config.AppConfig{
		Files: []config.FileSpec{
			{Path: "LICENSE", Content: "MIT"},
			{Path: "CODEOWNERS", Content: "* @team"},
		},
	}
	files := materializeFileSpecs(app)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "LICENSE" || files[1].Path != "CODEOWNERS" {
		t.Errorf("expected LICENSE, CODEOWNERS order, got %q, %q", files[0].Path, files[1].Path)
	}
}

func TestRenderCodeowners(t *testing.T) {
	tests := []struct {
		name   string
		owners []string
		want   string
	}{
		{"empty", nil, ""},
		{"bare username", []string{"octocat"}, "* @octocat\n"},
		{"already prefixed", []string{"@octocat"}, "* @octocat\n"},
		{"team ref", []string{"@my-org/team"}, "* @my-org/team\n"},
		{"dedup", []string{"octocat", "@octocat"}, "* @octocat\n"},
		{"multiple", []string{"a", "@b", "@org/t"}, "* @a @b @org/t\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderCodeowners(tt.owners)
			if got != tt.want {
				t.Errorf("renderCodeowners(%v) = %q, want %q", tt.owners, got, tt.want)
			}
		})
	}
}

func TestPlanCodeowners_EmitsPerRepo(t *testing.T) {
	owners := map[string][]string{
		"api": {"allanice001"},
		"web": {"@org/frontend"},
	}
	names := map[string]string{"api": "api", "web": "web"}
	emitted := map[string]bool{}

	changes := planCodeowners("acme", owners, names, map[string]bool{}, emitted)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	// Deterministic order (sorted): api, web
	if changes[0].Target != "api:.github/CODEOWNERS" {
		t.Errorf("expected api first, got %q", changes[0].Target)
	}
	d := changes[0].Details.(map[string]any)
	if d["path"] != ".github/CODEOWNERS" {
		t.Errorf("expected .github/CODEOWNERS path, got %q", d["path"])
	}
	if d["content"] != "* @allanice001\n" {
		t.Errorf("unexpected api content: %q", d["content"])
	}
	if d["branch"] != "main" {
		t.Errorf("expected branch main, got %q", d["branch"])
	}
	if d["reconcile"] != true {
		t.Errorf("expected synthesized CODEOWNERS to be reconciled, got %v", d["reconcile"])
	}
	if !emitted["api:.github/CODEOWNERS"] {
		t.Error("expected emitted set to be updated")
	}
}

func TestPlanCodeowners_SkipsWhenUserDeclared(t *testing.T) {
	owners := map[string][]string{"api": {"octocat"}}
	names := map[string]string{"api": "api"}
	userFiles := map[string]bool{".github/CODEOWNERS": true}

	changes := planCodeowners("acme", owners, names, userFiles, map[string]bool{})
	if len(changes) != 0 {
		t.Errorf("expected user-declared CODEOWNERS to win, got %d synthesized changes", len(changes))
	}
}

func TestPlanCodeowners_SkipsRepoWithoutOwners(t *testing.T) {
	owners := map[string][]string{
		"api":   {"octocat"},
		"empty": nil,
	}
	names := map[string]string{"api": "api", "empty": "empty"}
	changes := planCodeowners("acme", owners, names, map[string]bool{}, map[string]bool{})
	if len(changes) != 1 {
		t.Fatalf("expected 1 change (api only), got %d", len(changes))
	}
	if changes[0].Target != "api:.github/CODEOWNERS" {
		t.Errorf("expected api change, got %q", changes[0].Target)
	}
}

func TestPlanCodeowners_RespectsEmittedSet(t *testing.T) {
	owners := map[string][]string{"api": {"octocat"}}
	names := map[string]string{"api": "api"}
	emitted := map[string]bool{"api:.github/CODEOWNERS": true}

	changes := planCodeowners("acme", owners, names, map[string]bool{}, emitted)
	if len(changes) != 0 {
		t.Errorf("expected no changes when already emitted, got %d", len(changes))
	}
}

func TestPlanCodeownersDeletions_OnlyForReposWithoutOwners(t *testing.T) {
	managed := map[string]bool{"api": true, "web": true, "infra": true}
	names := map[string]string{"api": "api", "web": "web", "infra": "infra"}
	owners := map[string][]string{"api": {"octocat"}}

	changes := planCodeownersDeletions("acme", managed, names, owners, map[string]bool{}, map[string]bool{})
	if len(changes) != 2 {
		t.Fatalf("expected 2 delete changes (web, infra), got %d", len(changes))
	}
	// Sorted: infra, web
	if changes[0].Target != "infra:.github/CODEOWNERS" {
		t.Errorf("expected infra first, got %q", changes[0].Target)
	}
	if changes[0].Action != "delete" {
		t.Errorf("expected delete action, got %q", changes[0].Action)
	}
	d := changes[0].Details.(map[string]any)
	if d["path"] != ".github/CODEOWNERS" {
		t.Errorf("expected .github/CODEOWNERS path, got %q", d["path"])
	}
	if d["message"] == "" {
		t.Error("expected non-empty commit message")
	}
}

func TestPlanCodeownersDeletions_SkipsWhenUserDeclared(t *testing.T) {
	managed := map[string]bool{"api": true}
	names := map[string]string{"api": "api"}
	owners := map[string][]string{} // no owners -> would normally delete
	userFiles := map[string]bool{".github/CODEOWNERS": true}

	changes := planCodeownersDeletions("acme", managed, names, owners, userFiles, map[string]bool{})
	if len(changes) != 0 {
		t.Errorf("expected user-declared CODEOWNERS to suppress deletion, got %d", len(changes))
	}
}

func TestPlanCodeownersDeletions_RespectsEmittedSet(t *testing.T) {
	managed := map[string]bool{"api": true}
	names := map[string]string{"api": "api"}
	emitted := map[string]bool{"api:.github/CODEOWNERS": true}

	changes := planCodeownersDeletions("acme", managed, names, map[string][]string{}, map[string]bool{}, emitted)
	if len(changes) != 0 {
		t.Errorf("expected no changes when already emitted (write wins), got %d", len(changes))
	}
}
