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

func TestPermRankOrdersPermissions(t *testing.T) {
	ordered := []string{"", "pull", "triage", "push", "maintain", "admin"}
	for i := 1; i < len(ordered); i++ {
		if permRank(ordered[i]) <= permRank(ordered[i-1]) {
			t.Errorf("%q should outrank %q", ordered[i], ordered[i-1])
		}
	}
	// GitHub says "read"/"write" in some responses and "pull"/"push" in others.
	if permRank("read") != permRank("pull") || permRank("write") != permRank("push") {
		t.Error("the two spellings of the same permission must rank equally")
	}
	// An unrecognized value must never be read as covering a real grant.
	if permRank("some-custom-role") != 0 {
		t.Error("an unknown permission must rank below every real one")
	}
}

func TestHigherPerm(t *testing.T) {
	if got := higherPerm("pull", "admin"); got != "admin" {
		t.Errorf("want admin, got %q", got)
	}
	if got := higherPerm("maintain", "push"); got != "maintain" {
		t.Errorf("want maintain, got %q", got)
	}
	if got := higherPerm("", "pull"); got != "pull" {
		t.Errorf("want pull, got %q", got)
	}
}

// collabWorld is the slice of an organization the collaborator planner reads.
type collabWorld struct {
	teams       []string            // slugs that exist on GitHub
	teamMembers map[string][]string // slug -> logins
	teamRepos   map[string][]string // slug -> "repo:permission"
	orgMembers  []string
	admins      []string
	direct      map[string][]string // repo -> "login:permission"
	baseline    string
}

func (w collabWorld) server(t *testing.T) *httptest.Server {
	t.Helper()
	enc := func(w http.ResponseWriter, v any) { _ = json.NewEncoder(w).Encode(v) }
	logins := func(names []string) []map[string]any {
		out := make([]map[string]any, 0, len(names))
		for _, n := range names {
			out = append(out, map[string]any{"login": n})
		}
		return out
	}
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/orgs/myorg":
			base := w.baseline
			if base == "" {
				base = "none"
			}
			enc(rw, map[string]any{"login": "myorg", "default_repository_permission": base})
		case path == "/orgs/myorg/members":
			if r.URL.Query().Get("role") == "admin" {
				enc(rw, logins(w.admins))
				return
			}
			enc(rw, logins(w.orgMembers))
		case strings.HasSuffix(path, "/members") && strings.Contains(path, "/teams/"):
			slug := strings.Split(strings.TrimPrefix(path, "/orgs/myorg/teams/"), "/")[0]
			enc(rw, logins(w.teamMembers[slug]))
		case strings.HasSuffix(path, "/repos") && strings.Contains(path, "/teams/"):
			slug := strings.Split(strings.TrimPrefix(path, "/orgs/myorg/teams/"), "/")[0]
			out := []map[string]any{}
			for _, spec := range w.teamRepos[slug] {
				name, perm, _ := strings.Cut(spec, ":")
				out = append(out, map[string]any{"name": name, "permissions": permBlock(perm)})
			}
			enc(rw, out)
		case strings.HasSuffix(path, "/collaborators"):
			repo := strings.TrimSuffix(strings.TrimPrefix(path, "/repos/myorg/"), "/collaborators")
			if r.URL.Query().Get("affiliation") != "direct" {
				t.Errorf("collaborators must be fetched with affiliation=direct, got %q", r.URL.Query().Get("affiliation"))
			}
			out := []map[string]any{}
			for _, spec := range w.direct[repo] {
				login, perm, _ := strings.Cut(spec, ":")
				out = append(out, map[string]any{"login": login, "role_name": perm})
			}
			enc(rw, out)
		default:
			http.NotFound(rw, r)
		}
	}))
}

func permBlock(perm string) map[string]bool {
	return map[string]bool{
		"pull":     permRank(perm) >= permRank("pull"),
		"triage":   permRank(perm) >= permRank("triage"),
		"push":     permRank(perm) >= permRank("push"),
		"maintain": permRank(perm) >= permRank("maintain"),
		"admin":    permRank(perm) >= permRank("admin"),
	}
}

func (w collabWorld) state() *State {
	st := &State{Org: "myorg", ManagedRepos: map[string]bool{}}
	for repo := range w.direct {
		st.ManagedRepos[repo] = true
	}
	for _, slug := range w.teams {
		st.ActualTeams = append(st.ActualTeams, &github.Team{Slug: github.Ptr(slug)})
	}
	return st
}

func collabCfg(remove bool) *config.Root {
	r := &config.Root{}
	r.App.Org = "myorg"
	r.App.RemoveExcessCollaborators = remove
	r.App.DryWarnings.WarnExcessCollaborators = true
	return r
}

func runCollab(t *testing.T, w collabWorld, cfg *config.Root, desired map[string]config.TeamConfig) ([]util.Change, []string) {
	t.Helper()
	server := w.server(t)
	defer server.Close()
	changes, warnings, err := planCollaborators(context.Background(), newTestClient(t, server), cfg, w.state(), desired)
	if err != nil {
		t.Fatalf("planCollaborators: %v", err)
	}
	return changes, warnings
}

// Nothing is fetched or planned unless one of the two flags asks for it.
func TestPlanCollaboratorsIsInertWhenUnconfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("no request should be made when both flags are off, got %s", r.URL.Path)
	}))
	defer server.Close()

	cfg := &config.Root{}
	changes, warnings, err := planCollaborators(context.Background(), newTestClient(t, server), cfg,
		&State{Org: "myorg", ManagedRepos: map[string]bool{"api": true}}, nil)
	if err != nil || len(changes) != 0 || len(warnings) != 0 {
		t.Fatalf("expected a no-op, got changes=%v warnings=%v err=%v", changes, warnings, err)
	}
}

func TestPlanCollaboratorsRemovesGrantBeyondTeamAccess(t *testing.T) {
	w := collabWorld{
		teams:       []string{"backend"},
		teamMembers: map[string][]string{"backend": {"alice"}},
		teamRepos:   map[string][]string{"backend": {"api:push"}},
		orgMembers:  []string{"alice"},
		direct:      map[string][]string{"api": {"alice:admin"}},
	}

	changes, warnings := runCollab(t, w, collabCfg(true), nil)

	if len(changes) != 1 {
		t.Fatalf("expected one revocation, got %+v", changes)
	}
	ch := changes[0]
	if ch.Scope != scopeRepoCollaborator || ch.Action != util.ActionRemove || ch.Target != "api/alice" {
		t.Errorf("unexpected change %s:%s %s", ch.Scope, ch.Action, ch.Target)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "alice") {
		t.Errorf("expected the revocation to be reported, got %v", warnings)
	}
}

// A direct grant no higher than what the team already confers changes nothing,
// so revoking it would be churn.
func TestPlanCollaboratorsLeavesRedundantGrantAlone(t *testing.T) {
	w := collabWorld{
		teams:       []string{"backend"},
		teamMembers: map[string][]string{"backend": {"alice"}},
		teamRepos:   map[string][]string{"backend": {"api:admin"}},
		orgMembers:  []string{"alice"},
		direct:      map[string][]string{"api": {"alice:push"}},
	}

	if changes, _ := runCollab(t, w, collabCfg(true), nil); len(changes) != 0 {
		t.Fatalf("a grant covered by team access should be left alone, got %+v", changes)
	}
}

// The org-wide default is access everyone already has; a direct grant at or
// below it confers nothing.
func TestPlanCollaboratorsAccountsForTheOrgDefaultPermission(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"alice"},
		direct:     map[string][]string{"api": {"alice:push"}},
		baseline:   "write",
	}

	if changes, _ := runCollab(t, w, collabCfg(true), nil); len(changes) != 0 {
		t.Fatalf("push is the org default here, so the grant adds nothing: %+v", changes)
	}

	w.baseline = "read"
	if changes, _ := runCollab(t, w, collabCfg(true), nil); len(changes) != 1 {
		t.Fatalf("push above a read default is excess, got %+v", changes)
	}
}

// An outside collaborator is in no team, so the org default does not reach them
// either: every direct grant they hold is excess.
func TestPlanCollaboratorsRevokesOutsideCollaborator(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"alice"},
		direct:     map[string][]string{"api": {"contractor:push"}},
		baseline:   "write",
	}

	changes, warnings := runCollab(t, w, collabCfg(true), nil)

	if len(changes) != 1 || changes[0].Target != "api/contractor" {
		t.Fatalf("expected the outside collaborator's grant to go, got %+v", changes)
	}
	if !strings.Contains(warnings[0], "no team access") {
		t.Errorf("the report should say they have no team access, got %v", warnings)
	}
}

// Owners hold admin everywhere already, so a direct grant on top is redundant
// rather than excessive and revoking it would change nothing.
func TestPlanCollaboratorsExemptsOwners(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"root"},
		admins:     []string{"root"},
		direct:     map[string][]string{"api": {"root:admin"}},
	}

	if changes, _ := runCollab(t, w, collabCfg(true), nil); len(changes) != 0 {
		t.Fatalf("owners should be exempt, got %+v", changes)
	}
}

// Warning without removing is the whole point of the split: report, change
// nothing.
func TestPlanCollaboratorsWarnsWithoutRemoving(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"alice"},
		direct:     map[string][]string{"api": {"alice:admin"}},
	}
	cfg := collabCfg(false)

	changes, warnings := runCollab(t, w, cfg, nil)

	if len(changes) != 0 {
		t.Fatalf("removal is opt-in, got %+v", changes)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "found") {
		t.Errorf("expected a report of what was found, got %v", warnings)
	}
}

// A team gomgr does not manage still grants what it grants. Enabling this in a
// half-adopted org must not revoke access that a real team confers.
func TestPlanCollaboratorsCountsUnmanagedTeams(t *testing.T) {
	w := collabWorld{
		teams:       []string{"legacy-admins"},
		teamMembers: map[string][]string{"legacy-admins": {"alice"}},
		teamRepos:   map[string][]string{"legacy-admins": {"api:admin"}},
		orgMembers:  []string{"alice"},
		direct:      map[string][]string{"api": {"alice:admin"}},
	}

	// No desired config at all — the entitlement comes entirely from GitHub.
	if changes, _ := runCollab(t, w, collabCfg(true), nil); len(changes) != 0 {
		t.Fatalf("an unmanaged team still entitles its members, got %+v", changes)
	}
}

// A team this run is about to create entitles its members too, or the first
// sync of a new config would revoke grants it is in the middle of replacing.
func TestPlanCollaboratorsCountsTeamsThisRunWillCreate(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"alice"},
		direct:     map[string][]string{"api": {"alice:admin"}},
	}
	desired := map[string]config.TeamConfig{
		"backend": {
			Name:         "Backend",
			Slug:         "backend",
			Members:      []string{"Alice"},
			Repositories: map[string]any{"api": "admin"},
		},
	}

	if changes, _ := runCollab(t, w, collabCfg(true), desired); len(changes) != 0 {
		t.Fatalf("a team about to be created entitles its members, got %+v", changes)
	}
}

// Nesting grants the parent's repository access to the child, so a child-team
// member is entitled to whatever the parent holds.
func TestPlanCollaboratorsFollowsParentInheritance(t *testing.T) {
	w := collabWorld{
		orgMembers: []string{"dave"},
		direct:     map[string][]string{"api": {"dave:maintain"}},
	}
	desired := map[string]config.TeamConfig{
		"platform": {
			Name: "Platform", Slug: "platform",
			Repositories: map[string]any{"api": "maintain"},
		},
		"oncall": {
			Name: "Oncall", Slug: "oncall",
			Parents: []string{"platform"},
			Members: []string{"dave"},
		},
	}

	if changes, _ := runCollab(t, w, collabCfg(true), desired); len(changes) != 0 {
		t.Fatalf("the child inherits the parent's maintain on api, got %+v", changes)
	}

	// Without the nesting, the same grant is excess.
	flat := map[string]config.TeamConfig{
		"platform": desired["platform"],
		"oncall":   {Name: "Oncall", Slug: "oncall", Members: []string{"dave"}},
	}
	if changes, _ := runCollab(t, w, collabCfg(true), flat); len(changes) != 1 {
		t.Fatalf("without the parent, dave is entitled to nothing on api: %+v", changes)
	}
}

func TestCollaboratorPermissionPrefersRoleName(t *testing.T) {
	u := &github.User{
		RoleName:    github.Ptr("write"),
		Permissions: &github.RepositoryPermissions{Pull: github.Ptr(true)},
	}
	if got := collaboratorPermission(u); got != permPush {
		t.Errorf("role_name should win and normalize to push, got %q", got)
	}
}

func TestCollaboratorPermissionFallsBackToPermissionBlock(t *testing.T) {
	u := &github.User{Permissions: &github.RepositoryPermissions{
		Pull: github.Ptr(true), Push: github.Ptr(true), Maintain: github.Ptr(true),
	}}
	if got := collaboratorPermission(u); got != permMaintain {
		t.Errorf("want maintain, got %q", got)
	}
	if got := collaboratorPermission(&github.User{}); got != "" {
		t.Errorf("no permission information means no permission, got %q", got)
	}
}
