package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg
create_repo: true
`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners:
  - alice
`)
	teamsDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(teamsDir, "backend.yaml"), `name: Backend
slug: backend
members:
  - alice
repositories:
  api: push
`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.App.Org != "myorg" {
		t.Errorf("expected org=myorg, got %q", root.App.Org)
	}
	if !root.App.CreateRepo {
		t.Error("expected CreateRepo=true")
	}
	if len(root.Org.Owners) != 1 || root.Org.Owners[0] != "alice" {
		t.Errorf("expected owners=[alice], got %v", root.Org.Owners)
	}
	if len(root.Team) != 1 {
		t.Fatalf("expected 1 team, got %d", len(root.Team))
	}
	if root.Team[0].Name != "Backend" {
		t.Errorf("expected team name=Backend, got %q", root.Team[0].Name)
	}
}

func TestLoad_MissingAppYaml(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing app.yaml")
	}
	if !strings.Contains(err.Error(), "app.yaml") {
		t.Errorf("expected error about app.yaml, got: %v", err)
	}
}

func TestLoad_MissingOrg(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `create_repo: true`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty org")
	}
	if !strings.Contains(err.Error(), "app.org is required") {
		t.Errorf("expected 'app.org is required' error, got: %v", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `{{{invalid yaml`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse YAML") {
		t.Errorf("expected parse YAML error, got: %v", err)
	}
}

func TestLoad_NoTeamsDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Team) != 0 {
		t.Errorf("expected 0 teams, got %d", len(root.Team))
	}
}

func TestLoad_IgnoresNonYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)
	teamsDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(teamsDir, ".DS_Store"), "binary junk")
	writeFile(t, filepath.Join(teamsDir, "README.md"), "# Teams")
	writeFile(t, filepath.Join(teamsDir, "backend.yaml"), `name: Backend
members:
  - alice
`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Team) != 1 {
		t.Errorf("expected 1 team (ignoring non-YAML), got %d", len(root.Team))
	}
}

func TestBootstrapTeamYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "teams", "new-team.yaml")

	if err := BootstrapTeamYAML(path, "New Team"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "name: New Team") {
		t.Errorf("expected 'name: New Team' in output, got:\n%s", content)
	}
}

func TestResolvedSlug(t *testing.T) {
	tests := []struct {
		name string
		tc   TeamConfig
		want string
	}{
		{
			name: "explicit slug",
			tc:   TeamConfig{Name: "Backend", Slug: "be-team"},
			want: "be-team",
		},
		{
			name: "derived from name",
			tc:   TeamConfig{Name: "Backend Team"},
			want: "backend-team",
		},
		{
			name: "empty both",
			tc:   TeamConfig{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tc.ResolvedSlug()
			if got != tt.want {
				t.Errorf("ResolvedSlug() = %q, want %q", got, tt.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
