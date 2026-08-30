package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func boolPtr(b bool) *bool { return &b }

func repoFixture(name string, archived bool) *github.Repository {
	return &github.Repository{Name: github.Ptr(name), Archived: github.Ptr(archived)}
}

// Omitting `archived:` and setting it to false are different instructions. An
// absent key must leave a hand-archived repository alone.
func TestPlanRepoArchiveIgnoresRepositoriesThatSayNothing(t *testing.T) {
	changes := planRepoArchive("myorg",
		map[string]repoSettings{"attic": {permission: permAdmin}},
		map[string]*github.Repository{"attic": repoFixture("attic", true)})

	if len(changes) != 0 {
		t.Fatalf("an absent archived: key must not un-archive anything: %+v", changes)
	}
}

func TestPlanRepoArchiveArchivesWhenDeclared(t *testing.T) {
	changes := planRepoArchive("myorg",
		map[string]repoSettings{"attic": {archived: boolPtr(true)}},
		map[string]*github.Repository{"attic": repoFixture("Attic", false)})

	if len(changes) != 1 {
		t.Fatalf("expected one archive, got %+v", changes)
	}
	ch := changes[0]
	if ch.Scope != scopeRepoArchive || ch.Action != util.ActionEnsure {
		t.Errorf("unexpected %s:%s", ch.Scope, ch.Action)
	}
	d := ch.Details.(map[string]any)
	// The API call must carry GitHub's casing, not the lowercased map key.
	if d["repo"] != "Attic" || d[detailArchived] != true || d["org"] != "myorg" {
		t.Errorf("details lost something: %+v", d)
	}
}

// Un-archiving is a different scope because it has to run before the writes
// that an archived repository would reject.
func TestPlanRepoArchiveUnarchivesUnderItsOwnScope(t *testing.T) {
	changes := planRepoArchive("myorg",
		map[string]repoSettings{"attic": {archived: boolPtr(false)}},
		map[string]*github.Repository{"attic": repoFixture("attic", true)})

	if len(changes) != 1 || changes[0].Scope != scopeRepoUnarchive {
		t.Fatalf("expected a repo-unarchive change, got %+v", changes)
	}
	if defaultRegistry.Precedence(scopeRepoUnarchive, util.ActionEnsure) >=
		defaultRegistry.Precedence(scopeRepoFile, util.ActionEnsure) {
		t.Error("un-archiving must run before file writes, or those writes fail")
	}
	if defaultRegistry.Precedence(scopeRepoArchive, util.ActionEnsure) <=
		defaultRegistry.Precedence(scopeRepoFile, util.ActionEnsure) {
		t.Error("archiving must run after file writes, or those writes fail")
	}
}

func TestPlanRepoArchiveIsANoOpWhenAlreadyInTheDeclaredState(t *testing.T) {
	changes := planRepoArchive("myorg",
		map[string]repoSettings{
			"a": {archived: boolPtr(true)},
			"b": {archived: boolPtr(false)},
		},
		map[string]*github.Repository{
			"a": repoFixture("a", true),
			"b": repoFixture("b", false),
		})

	if len(changes) != 0 {
		t.Fatalf("nothing to do, so nothing planned: %+v", changes)
	}
}

// GitHub creates repositories unarchived, so only "archive it" is a change for
// a repository this run is about to create.
func TestPlanRepoArchiveHandlesARepositoryBeingCreated(t *testing.T) {
	changes := planRepoArchive("myorg",
		map[string]repoSettings{"new": {archived: boolPtr(true)}},
		map[string]*github.Repository{})
	if len(changes) != 1 {
		t.Fatalf("expected the archive to be planned blind, got %+v", changes)
	}

	changes = planRepoArchive("myorg",
		map[string]repoSettings{"new": {archived: boolPtr(false)}},
		map[string]*github.Repository{})
	if len(changes) != 0 {
		t.Fatalf("a repository being created is already unarchived: %+v", changes)
	}
}

func archiveCfg(archive, del bool) *config.Root {
	r := &config.Root{}
	r.App.Org = "myorg"
	r.App.ArchiveUnmanagedRepos = archive
	r.App.DeleteUnmanagedRepos = del
	return r
}

func unmanagedState() *State {
	return &State{
		Org:          "myorg",
		ManagedRepos: map[string]bool{"kept": true},
		ActualRepos: []*github.Repository{
			repoFixture("kept", false),
			repoFixture("Stray", false),
			repoFixture("already-parked", true),
		},
	}
}

func TestPlanUnmanagedArchive(t *testing.T) {
	changes, warnings := planUnmanagedArchive(archiveCfg(true, false), unmanagedState())

	if len(changes) != 1 {
		t.Fatalf("expected only the unmanaged, unarchived repo: %+v", changes)
	}
	if d := changes[0].Details.(map[string]any); d["repo"] != "Stray" {
		t.Errorf("expected Stray, got %+v", d)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "reversible") {
		t.Errorf("the report should say it can be undone: %v", warnings)
	}
}

// Being wrong in the direction somebody can undo is the whole point, so archive
// wins and the deletion is reported rather than silently dropped.
func TestArchiveWinsOverDelete(t *testing.T) {
	cfg := archiveCfg(true, true)
	st := unmanagedState()

	archives, archiveWarnings := planUnmanagedArchive(cfg, st)
	deletes, _, err := planRepoCleanups(cfg, st)
	if err != nil {
		t.Fatalf("planRepoCleanups: %v", err)
	}

	if len(archives) != 1 {
		t.Fatalf("expected the archive, got %+v", archives)
	}
	for _, ch := range deletes {
		if ch.Action == util.ActionDelete {
			t.Fatalf("nothing should be deleted when archiving is on: %+v", ch)
		}
	}
	var said bool
	for _, w := range archiveWarnings {
		if strings.Contains(w, "archiving wins") {
			said = true
		}
	}
	if !said {
		t.Errorf("the ignored flag must be reported, not silently dropped: %v", archiveWarnings)
	}
}

// delete_unmanaged_repos stays a real choice. It just says out loud what it is
// about to do, and only when it is actually about to do it.
func TestDeleteWarnsAboutTheReversibleAlternative(t *testing.T) {
	cfg := archiveCfg(false, true)
	cfg.App.DryWarnings.WarnUnmanagedRepos = true

	_, warnings, err := planRepoCleanups(cfg, unmanagedState())
	if err != nil {
		t.Fatalf("planRepoCleanups: %v", err)
	}

	var said bool
	for _, w := range warnings {
		if strings.Contains(w, "will DELETE") && strings.Contains(w, "archive_unmanaged_repos") {
			said = true
		}
	}
	if !said {
		t.Errorf("expected a warning naming the reversible option: %v", warnings)
	}
}

func TestDeleteSaysNothingWhenThereIsNothingToDelete(t *testing.T) {
	st := &State{Org: "myorg", ManagedRepos: map[string]bool{"kept": true},
		ActualRepos: []*github.Repository{repoFixture("kept", false)}}

	_, warnings, err := planRepoCleanups(archiveCfg(false, true), st)
	if err != nil {
		t.Fatalf("planRepoCleanups: %v", err)
	}
	for _, w := range warnings {
		if strings.Contains(w, "will DELETE") {
			t.Errorf("no unmanaged repos, so no scary warning: %v", warnings)
		}
	}
}

func TestApplyRepoArchiveEnsure(t *testing.T) {
	for _, tc := range []struct {
		name string
		want bool
	}{
		{"archive", true}, {"un-archive", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPatch && r.URL.Path == "/repos/myorg/attic" {
					_ = json.NewDecoder(r.Body).Decode(&got)
					_ = json.NewEncoder(w).Encode(map[string]any{"name": "attic"})
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			err := applyRepoArchiveEnsure(context.Background(), newTestClient(t, server), util.Change{
				Scope: scopeRepoArchive, Target: "attic", Action: util.ActionEnsure,
				Details: map[string]any{"org": "myorg", "repo": "attic", detailArchived: tc.want},
			})
			if err != nil {
				t.Fatalf("apply: %v", err)
			}
			if got["archived"] != tc.want {
				t.Errorf("expected archived=%v, got %+v", tc.want, got)
			}
		})
	}
}

// A repository being un-archived this run must not be skipped by the planners
// that refuse to touch archived repositories — the un-archive runs first.
func TestUnarchivingThisRun(t *testing.T) {
	got := unarchivingThisRun(map[string]repoSettings{
		"coming-back": {archived: boolPtr(false)},
		"going-away":  {archived: boolPtr(true)},
		"silent":      {},
	})
	if !got["coming-back"] || got["going-away"] || got["silent"] {
		t.Errorf("only an explicit false counts: %+v", got)
	}
}

// An archived repository is a parked one, and parking is what
// archive_unmanaged_repos does. Deleting it on a later run would turn a
// deliberate switch from the reversible option to the destructive one into a
// sweep of everything the reversible option had already parked.
func TestDeleteSparesAnAlreadyArchivedRepo(t *testing.T) {
	changes, warnings, err := planRepoCleanups(archiveCfg(false, true), unmanagedState())
	if err != nil {
		t.Fatalf("planRepoCleanups: %v", err)
	}

	var deleted []string
	for _, ch := range changes {
		if ch.Action == util.ActionDelete {
			deleted = append(deleted, ch.Target)
		}
	}
	if len(deleted) != 1 || deleted[0] != "stray" {
		t.Errorf("deleted = %v, want only the unarchived stray", deleted)
	}

	var said bool
	for _, w := range warnings {
		if strings.Contains(w, "already-parked") && strings.Contains(w, "Leaving") {
			said = true
		}
		if strings.Contains(w, "will DELETE") && strings.Contains(w, "already-parked") {
			t.Errorf("the deletion warning must count only what is actually deleted: %v", w)
		}
	}
	if !said {
		t.Errorf("skipping a repository must be reported, not silent: %v", warnings)
	}
}

// A repository repos.yaml names is written down. Nothing is applied to it until
// a team names it too — repo_plan warns about that — but "no team names it" is
// not a reason to delete or archive a repository the configuration declares.
func TestCleanupsSpareARepoReposYamlDeclares(t *testing.T) {
	for _, tc := range []struct {
		name             string
		archive, del     bool
		wantChangeCounts int
	}{
		{name: "delete", del: true},
		{name: "archive", archive: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := archiveCfg(tc.archive, tc.del)
			cfg.Repos = map[string]any{"Stray": map[string]any{"visibility": "private"}}
			st := unmanagedState()

			deletes, _, err := planRepoCleanups(cfg, st)
			if err != nil {
				t.Fatalf("planRepoCleanups: %v", err)
			}
			archives, _ := planUnmanagedArchive(cfg, st)

			for _, ch := range append(deletes, archives...) {
				if ch.Target == "stray" {
					t.Errorf("planned %s:%s on a repository repos.yaml declares", ch.Scope, ch.Action)
				}
			}
		})
	}
}
