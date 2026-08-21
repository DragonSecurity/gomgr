package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/util"
)

// teamServer records the body of the team create/update call and resolves one
// parent slug to an ID.
func teamServer(t *testing.T, parentSlug string, parentID int64) (*httptest.Server, *map[string]any) {
	t.Helper()
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/orgs/myorg/teams/"+parentSlug:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": parentID, "slug": parentSlug})
		case r.Method == http.MethodPost && r.URL.Path == "/orgs/myorg/teams":
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99, "slug": "child"})
		case r.Method == http.MethodPatch && r.URL.Path == "/orgs/myorg/teams/child":
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99, "slug": "child"})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, &body
}

func TestApplyTeamCreateResolvesParentSlugToID(t *testing.T) {
	server, body := teamServer(t, "platform", 4242)
	defer server.Close()

	err := applyTeamCreate(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeTeam, Target: "child", Action: util.ActionCreate,
		Details: map[string]any{"org": "myorg", "name": "Child", "parent": "platform"},
	})
	if err != nil {
		t.Fatalf("applyTeamCreate: %v", err)
	}
	if got := (*body)["parent_team_id"]; got != float64(4242) {
		t.Errorf("expected parent_team_id 4242, got %v", got)
	}
}

func TestApplyTeamCreateFailsLoudlyOnAnUnresolvableParent(t *testing.T) {
	server, _ := teamServer(t, "somewhere-else", 1)
	defer server.Close()

	err := applyTeamCreate(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeTeam, Target: "child", Action: util.ActionCreate,
		Details: map[string]any{"org": "myorg", "name": "Child", "parent": "platform"},
	})
	if err == nil {
		t.Fatal("a parent that cannot be resolved must not silently create an unnested team")
	}
}

// parent_team_id carries omitempty, so an update that is not about nesting must
// leave the field out entirely rather than sending null.
func TestApplyTeamUpdateLeavesNestingAloneWhenNotChanging(t *testing.T) {
	server, body := teamServer(t, "platform", 4242)
	defer server.Close()

	err := applyTeamUpdate(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeTeam, Target: "child", Action: util.ActionUpdate,
		Details: map[string]any{"org": "myorg", "slug": "child", "name": "Child", "privacy": "closed"},
	})
	if err != nil {
		t.Fatalf("applyTeamUpdate: %v", err)
	}
	if _, present := (*body)["parent_team_id"]; present {
		t.Errorf("parent_team_id must be omitted when nesting is unchanged, got %v", *body)
	}
}

func TestApplyTeamUpdateSetsANewParent(t *testing.T) {
	server, body := teamServer(t, "platform", 4242)
	defer server.Close()

	err := applyTeamUpdate(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeTeam, Target: "child", Action: util.ActionUpdate,
		Details: map[string]any{"org": "myorg", "slug": "child", "name": "Child", "parent": "platform"},
	})
	if err != nil {
		t.Fatalf("applyTeamUpdate: %v", err)
	}
	if got := (*body)["parent_team_id"]; got != float64(4242) {
		t.Errorf("expected parent_team_id 4242, got %v", got)
	}
}

// Clearing a parent is a different request from leaving it alone, which is why
// the planner distinguishes the two.
func TestApplyTeamUpdateClearsAParent(t *testing.T) {
	server, body := teamServer(t, "platform", 4242)
	defer server.Close()

	err := applyTeamUpdate(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeTeam, Target: "child", Action: util.ActionUpdate,
		Details: map[string]any{"org": "myorg", "slug": "child", "name": "Child", "remove_parent": true},
	})
	if err != nil {
		t.Fatalf("applyTeamUpdate: %v", err)
	}
	got, present := (*body)["parent_team_id"]
	if !present || got != nil {
		t.Errorf("removing a parent must send an explicit null, got present=%v value=%v", present, got)
	}
}

func TestApplyOrgOwnerEnsureAndRemove(t *testing.T) {
	var gotRole string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/orgs/myorg/memberships/mallory" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotRole, _ = body["role"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{"role": gotRole})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	c := newTestClient(t, server)
	ch := util.Change{Scope: scopeOrgOwner, Target: "mallory", Action: util.ActionEnsure,
		Details: map[string]any{"org": "myorg", "user": "mallory"}}

	if err := applyOrgOwnerEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("applyOrgOwnerEnsure: %v", err)
	}
	if gotRole != orgRoleAdmin {
		t.Errorf("promotion should send role=admin, got %q", gotRole)
	}

	ch.Action = util.ActionRemove
	if err := applyOrgOwnerRemove(context.Background(), c, ch); err != nil {
		t.Fatalf("applyOrgOwnerRemove: %v", err)
	}
	// Demotion drops the role; it does not remove the account from the org.
	if gotRole != orgRoleMember {
		t.Errorf("demotion should send role=member, got %q", gotRole)
	}
}

func TestApplyRepoCollaboratorRemove(t *testing.T) {
	removed := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/repos/myorg/api/collaborators/alice" {
			removed = "alice"
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	err := applyRepoCollaboratorRemove(context.Background(), newTestClient(t, server), util.Change{
		Scope: scopeRepoCollaborator, Target: "api/alice", Action: util.ActionRemove,
		Details: map[string]any{"org": "myorg", "repo": "api", "user": "alice"},
	})
	if err != nil {
		t.Fatalf("applyRepoCollaboratorRemove: %v", err)
	}
	if removed != "alice" {
		t.Error("expected the direct grant to be revoked")
	}
}
