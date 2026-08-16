package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// teamFixture describes one team as the fake API should report it.
type teamFixture struct {
	slug        string
	name        string
	description string
	privacy     string
	parent      string
	maintainers []string
	members     []string
	repos       map[string]string // repo name -> permission
}

// teamImportServer answers the endpoints ImportTeams calls. orgRepos is every
// repository in the organization, whether or not a team reaches it.
func teamImportServer(t *testing.T, teams []teamFixture, orgRepos []string) *httptest.Server {
	t.Helper()

	bySlug := map[string]teamFixture{}
	for _, tf := range teams {
		bySlug[tf.slug] = tf
	}

	permBlock := func(perm string) map[string]any {
		levels := map[string]bool{"pull": false, "triage": false, "push": false, "maintain": false, "admin": false}
		switch perm {
		case "admin":
			levels["admin"] = true
		case "maintain":
			levels["maintain"] = true
		case "push":
			levels["push"] = true
		case "triage":
			levels["triage"] = true
		default:
			levels["pull"] = true
		}
		out := map[string]any{}
		for k, v := range levels {
			out[k] = v
		}
		return out
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case path == "/orgs/myorg/teams":
			var out []map[string]any
			for _, tf := range teams {
				entry := map[string]any{
					"slug":        tf.slug,
					"name":        tf.name,
					"description": tf.description,
					"privacy":     tf.privacy,
				}
				if tf.parent != "" {
					entry["parent"] = map[string]any{"slug": tf.parent}
				}
				out = append(out, entry)
			}
			_ = json.NewEncoder(w).Encode(out)

		case path == "/orgs/myorg/repos":
			var out []map[string]any
			for i, name := range orgRepos {
				out = append(out, map[string]any{"id": 100 + i, "name": name})
			}
			_ = json.NewEncoder(w).Encode(out)

		case strings.HasSuffix(path, "/members"):
			slug := strings.TrimSuffix(strings.TrimPrefix(path, "/orgs/myorg/teams/"), "/members")
			tf := bySlug[slug]
			logins := tf.members
			if r.URL.Query().Get("role") == roleMaintainer {
				logins = tf.maintainers
			}
			var out []map[string]any
			for _, login := range logins {
				out = append(out, map[string]any{"login": login})
			}
			_ = json.NewEncoder(w).Encode(out)

		case strings.HasSuffix(path, "/repos"):
			slug := strings.TrimSuffix(strings.TrimPrefix(path, "/orgs/myorg/teams/"), "/repos")
			var out []map[string]any
			for name, perm := range bySlug[slug].repos {
				out = append(out, map[string]any{"name": name, "permissions": permBlock(perm)})
			}
			_ = json.NewEncoder(w).Encode(out)

		default:
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
}

func TestImportTeams(t *testing.T) {
	server := teamImportServer(t, []teamFixture{
		{
			slug: "platform", name: "Platform", description: "Runs the platform", privacy: "closed",
			maintainers: []string{"zoe", "alice"},
			members:     []string{"bob"},
			repos:       map[string]string{"infra": "admin", "docs": "pull"},
		},
		{
			slug: "already-here", name: "Already Here", privacy: "closed",
			members: []string{"carol"},
			repos:   map[string]string{"legacy": "push"},
		},
	}, []string{"infra", "docs", "legacy", "orphan"})
	defer server.Close()

	cfg := &config.Root{
		App:  config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{Name: "Already Here", Slug: "already-here"}},
	}

	result, err := ImportTeams(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if result.AlreadyDeclared != 1 {
		t.Errorf("AlreadyDeclared = %d, want 1", result.AlreadyDeclared)
	}
	if len(result.Teams) != 1 {
		t.Fatalf("adopted %d teams, want 1: %+v", len(result.Teams), result.Teams)
	}

	got := result.Teams[0].Config
	if got.Name != "Platform" || got.Privacy != "closed" || got.Description != "Runs the platform" {
		t.Errorf("team = %+v", got)
	}
	// The slug is derivable from the name, so it should not be restated.
	if got.Slug != "" {
		t.Errorf("slug = %q, want it omitted since the name derives it", got.Slug)
	}
	// Sorted, so re-importing produces the same file.
	if len(got.Maintainers) != 2 || got.Maintainers[0] != "alice" || got.Maintainers[1] != "zoe" {
		t.Errorf("maintainers = %v, want them sorted", got.Maintainers)
	}
	if len(got.Members) != 1 || got.Members[0] != "bob" {
		t.Errorf("members = %v", got.Members)
	}
	if got.Repositories["infra"] != "admin" || got.Repositories["docs"] != permPull {
		t.Errorf("repositories = %+v", got.Repositories)
	}

	// "orphan" is in the org but no team reaches it.
	if len(result.Ungranted) != 1 || result.Ungranted[0] != "orphan" {
		t.Errorf("Ungranted = %v, want [orphan]", result.Ungranted)
	}
	if result.DeletionRisk {
		t.Error("DeletionRisk should be false when delete_unmanaged_repos is off")
	}
}

func TestImportTeamsFlagsDeletionRisk(t *testing.T) {
	server := teamImportServer(t, []teamFixture{
		{slug: "platform", name: "Platform", privacy: "closed", repos: map[string]string{"infra": "admin"}},
	}, []string{"infra", "forgotten-one", "forgotten-two"})
	defer server.Close()

	cfg := &config.Root{App: config.AppConfig{Org: "myorg"}}
	cfg.App.DeleteUnmanagedRepos = true

	result, err := ImportTeams(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !result.DeletionRisk {
		t.Error("repos reached by no team plus delete_unmanaged_repos is exactly the case worth shouting about")
	}
	if len(result.Ungranted) != 2 {
		t.Errorf("Ungranted = %v, want both forgotten repos", result.Ungranted)
	}
}

func TestImportTeamsRecordsSlugWhenNotDerivable(t *testing.T) {
	server := teamImportServer(t, []teamFixture{
		{slug: "plat", name: "Platform Team", privacy: "closed"},
	}, nil)
	defer server.Close()

	result, err := ImportTeams(context.Background(), newTestClient(t, server),
		&config.Root{App: config.AppConfig{Org: "myorg"}})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	got := result.Teams[0].Config
	if got.Slug != "plat" {
		t.Errorf("slug = %q; a slug the name does not derive must be recorded", got.Slug)
	}
	if got.ResolvedSlug() != "plat" {
		t.Errorf("ResolvedSlug() = %q, want plat", got.ResolvedSlug())
	}
}

func TestImportTeamsReportsNesting(t *testing.T) {
	server := teamImportServer(t, []teamFixture{
		{slug: "child", name: "Child", privacy: "closed", parent: "parent-team"},
	}, nil)
	defer server.Close()

	result, err := ImportTeams(context.Background(), newTestClient(t, server),
		&config.Root{App: config.AppConfig{Org: "myorg"}})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Teams[0].Parent != "parent-team" {
		t.Errorf("Parent = %q, want parent-team so the caller can warn about it", result.Teams[0].Parent)
	}
	// gomgr does not manage hierarchy, so the written config must not imply it does.
	if len(result.Teams[0].Config.Parents) != 0 {
		t.Error("parents must not be written into a file gomgr will not act on")
	}
}

func TestImportTeamsSkipsUnrepresentable(t *testing.T) {
	server := teamImportServer(t, []teamFixture{
		{slug: "good", name: "Good", privacy: "closed"},
		// An empty name cannot be expressed: gomgr's schema requires one.
		{slug: "nameless", name: "", privacy: "closed"},
	}, nil)
	defer server.Close()

	result, err := ImportTeams(context.Background(), newTestClient(t, server),
		&config.Root{App: config.AppConfig{Org: "myorg"}})
	if err != nil {
		t.Fatalf("one unrepresentable team must not fail the scan: %v", err)
	}
	if len(result.Teams) != 1 || result.Teams[0].Config.Name != "Good" {
		t.Errorf("adopted = %+v, want just the representable team", result.Teams)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Slug != "nameless" {
		t.Errorf("Skipped = %+v", result.Skipped)
	}
}

func TestImportTeamsIsIdempotent(t *testing.T) {
	fixtures := []teamFixture{
		{slug: "platform", name: "Platform", privacy: "closed", members: []string{"bob"}},
	}
	server := teamImportServer(t, fixtures, []string{"infra"})
	defer server.Close()

	cfg := &config.Root{App: config.AppConfig{Org: "myorg"}}
	first, err := ImportTeams(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(first.Teams) != 1 {
		t.Fatalf("first import adopted %d teams, want 1", len(first.Teams))
	}

	// Declaring what the first import produced must leave nothing to adopt.
	cfg.Team = []config.TeamConfig{first.Teams[0].Config}
	second, err := ImportTeams(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("re-import: %v", err)
	}
	if len(second.Teams) != 0 {
		t.Errorf("re-import adopted %+v, want nothing", second.Teams)
	}
	if second.AlreadyDeclared != 1 {
		t.Errorf("AlreadyDeclared = %d, want 1", second.AlreadyDeclared)
	}
}
