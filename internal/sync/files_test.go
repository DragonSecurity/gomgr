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
