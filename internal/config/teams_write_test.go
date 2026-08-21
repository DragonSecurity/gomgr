package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func sampleTeam() TeamConfig {
	return TeamConfig{
		Name:        "Platform Team",
		Description: "Runs the platform",
		Privacy:     "closed",
		Maintainers: []string{"alice"},
		Members:     []string{"bob", "carol"},
		Repositories: map[string]any{
			"infra": "admin",
			"docs":  "pull",
		},
	}
}

func TestWriteTeamFile(t *testing.T) {
	dir := t.TempDir()

	path, err := WriteTeamFile(dir, sampleTeam())
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := filepath.Base(path); got != "platform-team.yaml" {
		t.Errorf("wrote %q, want the file named after the resolved slug", got)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !strings.HasPrefix(string(body), adoptedTeamComment) {
		t.Errorf("file should open with the marker comment:\n%s", body)
	}

	var reparsed TeamConfig
	if err := readYAML(path, &reparsed); err != nil {
		t.Fatalf("written file does not parse: %v\n%s", err, body)
	}
	if reparsed.Name != "Platform Team" || reparsed.Privacy != "closed" {
		t.Errorf("reparsed = %+v", reparsed)
	}
	if reparsed.Repositories["infra"] != "admin" || reparsed.Repositories["docs"] != "pull" {
		t.Errorf("repositories = %+v", reparsed.Repositories)
	}
	if len(reparsed.Members) != 2 {
		t.Errorf("members = %v", reparsed.Members)
	}
}

func TestWriteTeamFileRefusesToOverwrite(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteTeamFile(dir, sampleTeam()); err != nil {
		t.Fatalf("first write: %v", err)
	}

	// A team file somebody edited must never be silently replaced.
	path := filepath.Join(dir, "teams", "platform-team.yaml")
	if err := os.WriteFile(path, []byte("name: Hand Edited\n"), 0o644); err != nil {
		t.Fatalf("simulate hand edit: %v", err)
	}

	_, err := WriteTeamFile(dir, sampleTeam())
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v, want a refusal to overwrite", err)
	}
	body, _ := os.ReadFile(path)
	if string(body) != "name: Hand Edited\n" {
		t.Errorf("the existing file was modified:\n%s", body)
	}
}

func TestWrittenTeamFilesLoadAsAConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("org: myorg\n"), 0o644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "org.yaml"), []byte("owners:\n  - alice\n"), 0o644); err != nil {
		t.Fatalf("write org.yaml: %v", err)
	}

	teams := []TeamConfig{
		sampleTeam(),
		{
			Name:    "Security",
			Slug:    "sec",
			Privacy: "secret",
			Members: []string{"dave"},
		},
	}
	for _, team := range teams {
		if _, err := WriteTeamFile(dir, team); err != nil {
			t.Fatalf("write %s: %v", team.Name, err)
		}
	}

	// The whole point: what the importer writes has to be a config gomgr reads.
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("generated configuration does not load: %v", err)
	}
	if len(cfg.Team) != 2 {
		t.Fatalf("loaded %d teams, want 2", len(cfg.Team))
	}

	bySlug := map[string]TeamConfig{}
	for _, team := range cfg.Team {
		bySlug[team.ResolvedSlug()] = team
	}
	if _, ok := bySlug["platform-team"]; !ok {
		t.Errorf("platform-team missing from %v", bySlug)
	}
	if sec, ok := bySlug["sec"]; !ok || sec.Privacy != "secret" {
		t.Errorf("sec = %+v", sec)
	}
}

func TestWriteTeamFileCreatesTeamsDirectory(t *testing.T) {
	dir := t.TempDir()
	if _, err := os.Stat(filepath.Join(dir, "teams")); !os.IsNotExist(err) {
		t.Fatalf("fixture should start without a teams directory")
	}
	if _, err := WriteTeamFile(dir, sampleTeam()); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "teams")); err != nil {
		t.Errorf("teams directory was not created: %v", err)
	}
}

// A nested team's file has to carry the nesting, or importing an organization
// and syncing it straight back would flatten the hierarchy.
func TestWriteTeamFileRecordsParents(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteTeamFile(dir, TeamConfig{
		Name:    "Oncall",
		Privacy: "closed",
		Parents: []string{"platform"},
	}); err != nil {
		t.Fatalf("WriteTeamFile: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "teams", "oncall.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "parents:") || !strings.Contains(string(b), "platform") {
		t.Errorf("the nesting must survive the round trip, got:\n%s", b)
	}

	var back TeamConfig
	if err := yaml.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.ParentSlug() != "platform" {
		t.Errorf("re-read parent = %q, want platform", back.ParentSlug())
	}
}

// An un-nested team must not gain an empty key it never asked for.
func TestWriteTeamFileOmitsParentsWhenThereIsNone(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteTeamFile(dir, TeamConfig{Name: "Solo", Privacy: "closed"}); err != nil {
		t.Fatalf("WriteTeamFile: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "teams", "solo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "parents:") {
		t.Errorf("expected no parents key, got:\n%s", b)
	}
}
