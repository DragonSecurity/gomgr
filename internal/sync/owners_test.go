package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// ownerServer answers the endpoints planOrgOwners consults: the admin listing,
// and the authenticated-user lookup used for self-demotion protection.
func ownerServer(t *testing.T, admins []string, self string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/members" && r.URL.Query().Get("role") == "admin":
			users := make([]map[string]any, 0, len(admins))
			for _, a := range admins {
				users = append(users, map[string]any{"login": a})
			}
			_ = json.NewEncoder(w).Encode(users)
		case r.URL.Path == "/user":
			if self == "" {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "app tokens have no user"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"login": self})
		default:
			http.NotFound(w, r)
		}
	}))
}

func ownerPlan(t *testing.T, cfg *config.Root, admins []string, self string) ([]util.Change, []string) {
	t.Helper()
	server := ownerServer(t, admins, self)
	defer server.Close()
	changes, warnings, err := planOrgOwners(context.Background(), newTestClient(t, server), cfg, &State{Org: "myorg"})
	if err != nil {
		t.Fatalf("planOrgOwners: %v", err)
	}
	return changes, warnings
}

func ownerCfg(owners []string, demote bool, warn bool) *config.Root {
	r := &config.Root{}
	r.App.Org = "myorg"
	r.App.DemoteUnconfiguredOwners = demote
	r.App.DryWarnings.WarnUnmanagedOwners = warn
	r.Org.Owners = owners
	return r
}

func TestPlanOrgOwnersPromotesMissing(t *testing.T) {
	changes, _ := ownerPlan(t, ownerCfg([]string{"alice", "Bob"}, false, false), []string{"alice"}, "")

	if len(changes) != 1 {
		t.Fatalf("expected one promotion, got %d: %+v", len(changes), changes)
	}
	ch := changes[0]
	if ch.Scope != scopeOrgOwner || ch.Action != util.ActionEnsure || ch.Target != "bob" {
		t.Errorf("expected org-owner:ensure of bob, got %s:%s %s", ch.Scope, ch.Action, ch.Target)
	}
}

// The empty list is the load-bearing case: it means "gomgr does not manage
// owners here", and must never be read as "this org should have none".
func TestPlanOrgOwnersDoesNothingWhenUnset(t *testing.T) {
	changes, warnings := ownerPlan(t, ownerCfg(nil, true, true), []string{"alice", "bob"}, "")

	if len(changes) != 0 {
		t.Fatalf("an empty owner list must not demote anyone, got %+v", changes)
	}
	if len(warnings) != 0 {
		t.Errorf("nothing to warn about when owners are unmanaged, got %v", warnings)
	}
}

func TestPlanOrgOwnersWarnsAboutExtrasWithoutDemoting(t *testing.T) {
	changes, warnings := ownerPlan(t, ownerCfg([]string{"alice"}, false, true), []string{"alice", "mallory"}, "")

	if len(changes) != 0 {
		t.Fatalf("demotion is opt-in, got %+v", changes)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "mallory") {
		t.Errorf("expected a warning naming mallory, got %v", warnings)
	}
}

func TestPlanOrgOwnersDemotesExtrasWhenAsked(t *testing.T) {
	changes, _ := ownerPlan(t, ownerCfg([]string{"alice"}, true, false), []string{"alice", "mallory"}, "")

	if len(changes) != 1 {
		t.Fatalf("expected one demotion, got %+v", changes)
	}
	ch := changes[0]
	if ch.Action != util.ActionRemove || ch.Target != "mallory" {
		t.Errorf("expected org-owner:remove of mallory, got %s %s", ch.Action, ch.Target)
	}
}

// An owner that demotes itself cannot put itself back through the API that did
// it, so the account this run authenticates as is exempt.
func TestPlanOrgOwnersNeverDemotesItself(t *testing.T) {
	changes, warnings := ownerPlan(t, ownerCfg([]string{"alice"}, true, false), []string{"alice", "Carol"}, "carol")

	if len(changes) != 0 {
		t.Fatalf("the authenticated user must not be demoted, got %+v", changes)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "carol") {
		t.Errorf("expected the skip to be explained, got %v", warnings)
	}
}

// Under App auth there is no user behind the token, so /user errors. That is
// the answer, not a failure, and the demotion should still go ahead.
func TestPlanOrgOwnersDemotesUnderAppAuth(t *testing.T) {
	changes, _ := ownerPlan(t, ownerCfg([]string{"alice"}, true, false), []string{"alice", "mallory"}, "")

	if len(changes) != 1 || changes[0].Target != "mallory" {
		t.Fatalf("expected mallory to be demoted, got %+v", changes)
	}
}

func TestPlanOrgOwnersRecordsStats(t *testing.T) {
	server := ownerServer(t, []string{"alice", "bob"}, "")
	defer server.Close()
	st := &State{Org: "myorg"}

	if _, _, err := planOrgOwners(context.Background(), newTestClient(t, server),
		ownerCfg([]string{"alice"}, false, false), st); err != nil {
		t.Fatalf("planOrgOwners: %v", err)
	}
	if st.CurrentOwners != 2 || st.DesiredOwners != 1 {
		t.Errorf("want current=2 desired=1, got current=%d desired=%d", st.CurrentOwners, st.DesiredOwners)
	}
}
