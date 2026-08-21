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

func hierarchyCfg(teams ...config.TeamConfig) *config.Root {
	return &config.Root{App: config.AppConfig{Org: "myorg"}, Team: teams}
}

func TestPlanTeamsCreatesChildrenAfterParents(t *testing.T) {
	cfg := hierarchyCfg(
		config.TeamConfig{Name: "Oncall", Slug: "oncall", Parents: []string{"platform"}},
		config.TeamConfig{Name: "Platform", Slug: "platform"},
	)

	changes, _, err := planTeams(context.Background(), nil, cfg, &State{Org: "myorg"})
	if err != nil {
		t.Fatalf("planTeams: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected two creates, got %+v", changes)
	}
	if changes[0].Target != "platform" || changes[1].Target != "oncall" {
		t.Fatalf("the parent must be created first, got %s then %s", changes[0].Target, changes[1].Target)
	}
	details, ok := changes[1].Details.(map[string]any)
	if !ok || details["parent"] != "platform" {
		t.Errorf("the child should name its parent, got %+v", changes[1].Details)
	}
	if d := changes[0].Details.(map[string]any); d["parent"] != nil {
		t.Errorf("a root team should carry no parent, got %+v", d)
	}
}

func TestPlanTeamsPlansAReParent(t *testing.T) {
	cfg := hierarchyCfg(
		config.TeamConfig{Name: "Platform", Slug: "platform"},
		config.TeamConfig{Name: "Oncall", Slug: "oncall", Parents: []string{"platform"}},
	)
	st := &State{Org: "myorg", ActualTeams: []*github.Team{
		{Slug: github.Ptr("platform")},
		{Slug: github.Ptr("oncall")},
	}}

	changes, _, err := planTeams(context.Background(), nil, cfg, st)
	if err != nil {
		t.Fatalf("planTeams: %v", err)
	}
	if len(changes) != 1 || changes[0].Action != util.ActionUpdate || changes[0].Target != "oncall" {
		t.Fatalf("expected one update of oncall, got %+v", changes)
	}
	if d := changes[0].Details.(map[string]any); d["parent"] != "platform" {
		t.Errorf("expected parent=platform, got %+v", d)
	}
}

// Un-nesting has to be stated explicitly: GitHub's default for an omitted
// parent_team_id is to leave the nesting in place.
func TestPlanTeamsPlansAnUnNesting(t *testing.T) {
	cfg := hierarchyCfg(config.TeamConfig{Name: "Oncall", Slug: "oncall"})
	st := &State{Org: "myorg", ActualTeams: []*github.Team{
		{Slug: github.Ptr("oncall"), Parent: &github.Team{Slug: github.Ptr("platform")}},
	}}

	changes, _, err := planTeams(context.Background(), nil, cfg, st)
	if err != nil {
		t.Fatalf("planTeams: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected one update, got %+v", changes)
	}
	d := changes[0].Details.(map[string]any)
	if d["remove_parent"] != true {
		t.Errorf("expected remove_parent, got %+v", d)
	}
	if _, present := d["parent"]; present {
		t.Errorf("a removal must not also name a parent, got %+v", d)
	}
}

// Nesting that already matches is not a change.
func TestPlanTeamsLeavesMatchingNestingAlone(t *testing.T) {
	cfg := hierarchyCfg(config.TeamConfig{Name: "Oncall", Slug: "oncall", Parents: []string{"Platform"}})
	st := &State{Org: "myorg", ActualTeams: []*github.Team{
		{Slug: github.Ptr("oncall"), Parent: &github.Team{Slug: github.Ptr("platform")}},
	}}

	changes, _, err := planTeams(context.Background(), nil, cfg, st)
	if err != nil {
		t.Fatalf("planTeams: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("nothing has changed, so nothing should be planned: %+v", changes)
	}
}

// The plan used to come out in map order, so two dry runs of one unchanged
// config produced two differently ordered plans.
func TestPlanTeamsIsDeterministic(t *testing.T) {
	cfg := hierarchyCfg(
		config.TeamConfig{Name: "Zebra", Slug: "zebra"},
		config.TeamConfig{Name: "Alpha", Slug: "alpha"},
		config.TeamConfig{Name: "Middle", Slug: "middle"},
	)

	var first []string
	for i := 0; i < 15; i++ {
		changes, _, err := planTeams(context.Background(), nil, cfg, &State{Org: "myorg"})
		if err != nil {
			t.Fatalf("planTeams: %v", err)
		}
		var order []string
		for _, ch := range changes {
			order = append(order, ch.Target)
		}
		if first == nil {
			first = order
			continue
		}
		if strings.Join(order, ",") != strings.Join(first, ",") {
			t.Fatalf("plan order is not stable: %v then %v", first, order)
		}
	}
	if strings.Join(first, ",") != "alpha,middle,zebra" {
		t.Errorf("expected alphabetical roots, got %v", first)
	}
}

// GitHub refuses a permission change on an archived repository, so under
// ignore_archived the grant is skipped — and said out loud, because a
// repository archived out from under the config is worth hearing about.
func TestIgnoreArchivedSkipsGrantsAndSaysSo(t *testing.T) {
	repos := []map[string]any{{"name": "attic", "archived": true}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/repos" {
			_ = json.NewEncoder(w).Encode(repos)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer server.Close()

	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{
			Name: "Platform", Slug: "platform",
			Repositories: map[string]any{"attic": "push"},
		}},
	}

	granted := func(plan util.Plan) bool {
		for _, ch := range plan.Changes {
			if ch.Scope == scopeTeamRepo {
				return true
			}
		}
		return false
	}

	plan, err := BuildPlan(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if !granted(plan) {
		t.Fatal("without the flag the grant is planned, and fails against GitHub — that is the status quo being changed")
	}

	cfg.App.IgnoreArchived = true
	plan, err = BuildPlan(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if granted(plan) {
		t.Error("ignore_archived should skip the grant")
	}
	var said bool
	for _, w := range plan.Warnings {
		if strings.Contains(w, "archived") && strings.Contains(w, "attic") {
			said = true
		}
	}
	if !said {
		t.Errorf("skipping a repository silently is the failure mode; warnings were %v", plan.Warnings)
	}
}
