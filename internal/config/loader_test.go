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

func TestValidateCodeOwner(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"bare username", "octocat", false},
		{"prefixed user", "@octocat", false},
		{"team ref", "@my-org/platform-team", false},
		{"empty", "", true},
		{"whitespace inside", "@octo cat", true},
		{"only at sign", "@", true},
		{"team ref empty slug", "@my-org/", true},
		{"team ref bad slug", "@my-org/bad slug", true},
		{"consecutive hyphens", "bad--user", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeOwner(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCodeOwner(%q) err=%v, wantErr=%v", tt.in, err, tt.wantErr)
			}
		})
	}
}

// writeSplitConfig writes a configuration whose repository definitions live in
// repos.yaml and whose team file carries only permissions.
func writeSplitConfig(t *testing.T, reposYAML, teamRepos string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "org: myorg\n")
	writeFile(t, filepath.Join(dir, "org.yaml"), "owners:\n  - alice\n")
	if reposYAML != "" {
		writeFile(t, filepath.Join(dir, "repos.yaml"), reposYAML)
	}
	teamsDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(teamsDir, "backend.yaml"),
		"name: Backend\nslug: backend\nrepositories:\n"+teamRepos)
	return dir
}

func TestLoad_ReposFileIsOptional(t *testing.T) {
	dir := writeSplitConfig(t, "", "  api: push\n")
	root, err := Load(dir)
	if err != nil {
		t.Fatalf("a config with no repos.yaml must still load: %v", err)
	}
	if len(root.Repos) != 0 {
		t.Errorf("expected no repo definitions, got %v", root.Repos)
	}
}

func TestLoad_ReposFileDefinitions(t *testing.T) {
	dir := writeSplitConfig(t, `repos:
  api:
    topics: [backend]
    visibility: private
`, "  api: push\n")

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := root.Repos["api"]; !ok {
		t.Fatalf("expected api to be defined in repos.yaml, got %v", root.Repos)
	}
}

// A permission in repos.yaml has no team to belong to. Accepting it would put
// the ambiguity the split removes straight back.
func TestLoad_ReposFileRejectsPermission(t *testing.T) {
	dir := writeSplitConfig(t, "repos:\n  api:\n    permission: admin\n", "  api: push\n")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("permission in repos.yaml must be refused")
	}
	if !strings.Contains(err.Error(), "belongs to a team") {
		t.Errorf("error should say where permission belongs, got: %v", err)
	}
}

func TestLoad_ReposFileRejectsUnknownKey(t *testing.T) {
	dir := writeSplitConfig(t, "repos:\n  api:\n    topic: [backend]\n", "  api: push\n")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("an unknown key in repos.yaml must be refused")
	}
	if !strings.Contains(err.Error(), "topics") {
		t.Errorf("error should suggest the intended key, got: %v", err)
	}
}

// `only:` on a repository's own file is either redundant or contradicts the
// repository it is written under.
func TestLoad_RepoFilesRejectOnly(t *testing.T) {
	dir := writeSplitConfig(t, `repos:
  api:
    files:
      - path: .github/renovate.json
        only: [other]
        content: "{}"
`, "  api: push\n")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("`only:` on a repository's own file must be refused")
	}
	if !strings.Contains(err.Error(), "only") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

func TestLoad_RepoFilesAreValidated(t *testing.T) {
	dir := writeSplitConfig(t, `repos:
  api:
    files:
      - path: .github/renovate.json
        content: "{}"
      - path: .github/renovate.json
        content: "{}"
`, "  api: push\n")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("duplicate paths in a repository's files must be refused")
	}
	if !strings.Contains(err.Error(), "duplicate path") {
		t.Errorf("unexpected error: %v", err)
	}
}
