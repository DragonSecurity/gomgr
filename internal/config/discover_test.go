package config

import (
	"os"
	"path/filepath"
	"testing"
)

func mkConfigDir(t *testing.T, root, rel, body string) string {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDiscoverOrgDirs(t *testing.T) {
	root := t.TempDir()
	mkConfigDir(t, root, "alpha", "org: AlphaOrg\ncreate_repo: true\n")
	mkConfigDir(t, root, "nested/beta", "org: beta-org\n")

	dirs, err := DiscoverOrgDirs(root)
	if err != nil {
		t.Fatalf("DiscoverOrgDirs: %v", err)
	}
	if len(dirs) != 2 {
		t.Fatalf("expected 2 config directories, got %+v", dirs)
	}
	// Orgs are lowercased, because every other org comparison in gomgr is.
	got := map[string]string{}
	for _, d := range dirs {
		got[filepath.Base(d.Dir)] = d.Org
	}
	if got["alpha"] != "alphaorg" {
		t.Errorf("alpha org = %q, want alphaorg", got["alpha"])
	}
	if got["beta"] != "beta-org" {
		t.Errorf("beta org = %q, want beta-org", got["beta"])
	}
}

// One config directory does not contain another, and teams/ is full of YAML
// that would otherwise be probed for nothing.
func TestDiscoverOrgDirsDoesNotDescendIntoAConfigDir(t *testing.T) {
	root := t.TempDir()
	outer := mkConfigDir(t, root, "alpha", "org: alpha\n")
	if err := os.MkdirAll(filepath.Join(outer, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A stray app.yaml below a config directory must not be reported.
	mkConfigDir(t, root, filepath.Join("alpha", "teams", "oops"), "org: should-not-appear\n")

	dirs, err := DiscoverOrgDirs(root)
	if err != nil {
		t.Fatalf("DiscoverOrgDirs: %v", err)
	}
	if len(dirs) != 1 || dirs[0].Org != "alpha" {
		t.Fatalf("expected only the outer directory, got %+v", dirs)
	}
}

func TestDiscoverOrgDirsSkipsHiddenDirectories(t *testing.T) {
	root := t.TempDir()
	mkConfigDir(t, root, "alpha", "org: alpha\n")
	mkConfigDir(t, root, ".git/weird", "org: not-real\n")

	dirs, err := DiscoverOrgDirs(root)
	if err != nil {
		t.Fatalf("DiscoverOrgDirs: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("hidden directories must be skipped, got %+v", dirs)
	}
}

// A directory with an unrelated problem — a bad team file, a broken ruleset —
// is still a configured organization. Reporting it as missing would be worse
// than reporting it as present.
func TestDiscoverOrgDirsIgnoresProblemsOutsideAppYAML(t *testing.T) {
	root := t.TempDir()
	dir := mkConfigDir(t, root, "alpha", "org: alpha\n")
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "teams", "broken.yaml"), []byte("{{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	dirs, err := DiscoverOrgDirs(root)
	if err != nil {
		t.Fatalf("a broken team file must not hide the org: %v", err)
	}
	if len(dirs) != 1 || dirs[0].Org != "alpha" {
		t.Fatalf("expected alpha, got %+v", dirs)
	}
}

func TestDiscoverOrgDirsReportsANamelessAppYAML(t *testing.T) {
	root := t.TempDir()
	mkConfigDir(t, root, "alpha", "create_repo: true\n")

	dirs, err := DiscoverOrgDirs(root)
	if err != nil {
		t.Fatalf("DiscoverOrgDirs: %v", err)
	}
	if len(dirs) != 1 || dirs[0].Org != "" {
		t.Fatalf("an app.yaml naming no org is found with an empty Org, got %+v", dirs)
	}
}

func TestDiscoverOrgDirsRejectsAMissingRoot(t *testing.T) {
	if _, err := DiscoverOrgDirs(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("a missing config root should be an error, not an empty result")
	}
}
