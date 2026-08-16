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

func ptrTo[T any](v T) *T { return &v }

func testLookup() *refLookup {
	return &refLookup{
		org:   "myorg",
		appID: 4242,
		teams: map[string]int64{"platform": 77},
		repos: map[string]int64{"infra": 900},
	}
}

func TestBuildRulesetOrgDefaults(t *testing.T) {
	spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	rs, err := buildRuleset(context.Background(), spec, true, "", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if rs.Source != "myorg" {
		t.Errorf("source = %q, want myorg", rs.Source)
	}
	if rulesetTarget(rs) != github.RulesetTargetBranch {
		t.Errorf("target = %q, want branch", rulesetTarget(rs))
	}
	if rs.Conditions == nil || rs.Conditions.RepositoryName == nil {
		t.Fatal("an org ruleset must say which repositories it covers")
	}
	if got := rs.Conditions.RepositoryName.Include; len(got) != 1 || got[0] != "~ALL" {
		t.Errorf("repository_name include = %v, want [~ALL]", got)
	}
	if rs.Conditions.RefName == nil || rs.Conditions.RefName.Include[0] != "~DEFAULT_BRANCH" {
		t.Errorf("ref_name = %+v, want the preset's default-branch condition", rs.Conditions.RefName)
	}
	// Empty rather than nil, so the diff sees the same shape GitHub returns.
	if rs.Conditions.RefName.Exclude == nil {
		t.Error("exclude should be an empty list, not null")
	}
}

func TestBuildRulesetRepoScopeOmitsRepositoryCondition(t *testing.T) {
	spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	rs, err := buildRuleset(context.Background(), spec, false, "infra", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if rs.Source != "myorg/infra" {
		t.Errorf("source = %q, want myorg/infra", rs.Source)
	}
	if rs.Conditions.RepositoryName != nil {
		t.Error("a repository ruleset must not carry repository_name conditions")
	}
}

func TestBuildRulesetPushTargetHasNoRefConditions(t *testing.T) {
	spec, err := config.RulesetConfig{Name: "no-keys", Preset: config.PresetNoCommittedKeys}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	rs, err := buildRuleset(context.Background(), spec, false, "infra", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if rs.Conditions != nil {
		t.Errorf("conditions = %+v, want none on a repository push ruleset", rs.Conditions)
	}
	if rs.Rules.FileExtensionRestriction == nil {
		t.Fatal("the preset's file extension restriction is missing")
	}
}

func TestBuildBypassActors(t *testing.T) {
	spec := config.RulesetConfig{
		Name:        "guarded",
		Target:      config.RulesetTargetBranch,
		Enforcement: config.RulesetEnforcementActive,
		Rules:       config.RulesetRules{NonFastForward: ptrTo(true)},
		BypassActors: []config.BypassActorConfig{
			{Type: "Team", Team: "platform", Mode: config.BypassModePullRequest},
			{Type: "Integration", App: "self"},
			{Type: "OrganizationAdmin"},
			{Type: "RepositoryRole", ActorID: 5},
			{Type: "EnterpriseOwner"},
		},
	}

	rs, err := buildRuleset(context.Background(), spec, true, "", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rs.BypassActors) != 5 {
		t.Fatalf("got %d bypass actors, want 5", len(rs.BypassActors))
	}

	want := []struct {
		kind github.BypassActorType
		id   int64
		mode github.BypassMode
	}{
		{github.BypassActorTypeTeam, 77, github.BypassModePullRequest},
		{github.BypassActorTypeIntegration, 4242, github.BypassModeAlways},
		// OrganizationAdmin and EnterpriseOwner carry no ID: GitHub identifies
		// them by type and reports them back without one.
		{github.BypassActorTypeOrganizationAdmin, 0, github.BypassModeAlways},
		{github.BypassActorTypeRepositoryRole, 5, github.BypassModeAlways},
		{github.BypassActorType(config.BypassActorTypeEnterpriseOwner), 0, github.BypassModeAlways},
	}
	for i, w := range want {
		got := rs.BypassActors[i]
		if *got.ActorType != w.kind || got.GetActorID() != w.id || *got.BypassMode != w.mode {
			t.Errorf("actor[%d] = %s/%d/%s, want %s/%d/%s",
				i, *got.ActorType, got.GetActorID(), *got.BypassMode, w.kind, w.id, w.mode)
		}
	}
}

func TestBuildBypassActorSelfNeedsAppID(t *testing.T) {
	spec := config.RulesetConfig{
		Name:         "guarded",
		Target:       config.RulesetTargetBranch,
		Enforcement:  config.RulesetEnforcementActive,
		Rules:        config.RulesetRules{NonFastForward: ptrTo(true)},
		BypassActors: []config.BypassActorConfig{{Type: "Integration", App: "self"}},
	}

	l := testLookup()
	l.appID = 0
	_, err := buildRuleset(context.Background(), spec, true, "", l)
	if err == nil || !strings.Contains(err.Error(), "app_id") {
		t.Fatalf("err = %v, want a complaint about the missing app ID", err)
	}
}

func TestBuildRulesetUnresolvedTeamIsAnErrorAtPlanTime(t *testing.T) {
	spec := config.RulesetConfig{
		Name:         "guarded",
		Target:       config.RulesetTargetBranch,
		Enforcement:  config.RulesetEnforcementActive,
		Rules:        config.RulesetRules{NonFastForward: ptrTo(true)},
		BypassActors: []config.BypassActorConfig{{Type: "Team", Team: "brand-new"}},
	}

	_, err := buildRuleset(context.Background(), spec, true, "", testLookup())
	if err == nil || !strings.Contains(err.Error(), "brand-new") {
		t.Fatalf("err = %v, want the unresolvable team named", err)
	}
}

// roundTripThroughAPI mimics what GitHub hands back: the ruleset is serialized,
// given the identity fields and default parameters the API fills in, and read
// back as a fresh object.
func roundTripThroughAPI(t *testing.T, rs *github.RepositoryRuleset, mutate func(map[string]any)) *github.RepositoryRuleset {
	t.Helper()
	b, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(b, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	generic["id"] = 12345
	generic["source_type"] = "Organization"
	if mutate != nil {
		mutate(generic)
	}
	b, err = json.Marshal(generic)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var out github.RepositoryRuleset
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return &out
}

func TestRulesetMatchesIsIdempotent(t *testing.T) {
	spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	desired, err := buildRuleset(context.Background(), spec, true, "", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	actual := roundTripThroughAPI(t, desired, func(g map[string]any) {
		// GitHub answers with parameters nobody asked for: every merge method
		// on a pull_request rule, and a protected flag on the repo condition.
		rules, _ := g["rules"].([]any)
		for _, raw := range rules {
			rule, _ := raw.(map[string]any)
			if rule["type"] != "pull_request" {
				continue
			}
			params, _ := rule["parameters"].(map[string]any)
			params["allowed_merge_methods"] = []any{"merge", "squash", "rebase"}
		}
		conds, _ := g["conditions"].(map[string]any)
		repoName, _ := conds["repository_name"].(map[string]any)
		repoName["protected"] = false
	})

	same, err := rulesetMatches(actual, desired)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if !same {
		t.Error("a ruleset that only differs by GitHub's own defaults should compare equal")
	}
}

func TestRulesetMatchesDetectsRealDifferences(t *testing.T) {
	spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	desired, err := buildRuleset(context.Background(), spec, true, "", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name:   "enforcement relaxed",
			mutate: func(g map[string]any) { g["enforcement"] = "evaluate" },
		},
		{
			name: "approval count lowered",
			mutate: func(g map[string]any) {
				rules, _ := g["rules"].([]any)
				for _, raw := range rules {
					rule, _ := raw.(map[string]any)
					if rule["type"] == "pull_request" {
						params, _ := rule["parameters"].(map[string]any)
						params["required_approving_review_count"] = 0
					}
				}
			},
		},
		{
			name: "a rule removed",
			mutate: func(g map[string]any) {
				rules, _ := g["rules"].([]any)
				g["rules"] = rules[:len(rules)-1]
			},
		},
		{
			name: "an extra rule added",
			mutate: func(g map[string]any) {
				rules, _ := g["rules"].([]any)
				g["rules"] = append(rules, map[string]any{"type": "required_signatures"})
			},
		},
		{
			name: "condition widened to every branch",
			mutate: func(g map[string]any) {
				conds, _ := g["conditions"].(map[string]any)
				refName, _ := conds["ref_name"].(map[string]any)
				refName["include"] = []any{"~ALL"}
			},
		},
		{
			name: "an unconfigured bypass actor appeared",
			mutate: func(g map[string]any) {
				g["bypass_actors"] = []any{
					map[string]any{"actor_id": 1, "actor_type": "OrganizationAdmin", "bypass_mode": "always"},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := roundTripThroughAPI(t, desired, tt.mutate)
			same, err := rulesetMatches(actual, desired)
			if err != nil {
				t.Fatalf("compare: %v", err)
			}
			if same {
				t.Error("expected a difference to be detected")
			}
		})
	}
}

func TestPlanRulesetSetCreatesUpdatesAndSkips(t *testing.T) {
	baseline := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}
	resolved, err := baseline.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	inSync, err := buildRuleset(context.Background(), resolved, true, "", testLookup())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	tags := config.RulesetConfig{Name: "tags", Preset: config.PresetTagProtection}
	drifted := roundTripThroughAPI(t, inSync, func(g map[string]any) { g["enforcement"] = "disabled" })

	args := rulesetScopeArgs{scope: scopeOrgRuleset, org: "myorg", orgLevel: true}

	t.Run("creates what is missing", func(t *testing.T) {
		changes, err := planRulesetSet(context.Background(), args, []config.RulesetConfig{baseline, tags}, nil, testLookup())
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(changes) != 2 {
			t.Fatalf("got %d changes, want 2", len(changes))
		}
		for _, ch := range changes {
			if ch.Action != "create" {
				t.Errorf("change %s has action %q, want create", ch.Target, ch.Action)
			}
		}
	})

	t.Run("skips what already matches", func(t *testing.T) {
		existing := roundTripThroughAPI(t, inSync, nil)
		changes, err := planRulesetSet(context.Background(), args, []config.RulesetConfig{baseline},
			[]*github.RepositoryRuleset{existing}, testLookup())
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(changes) != 0 {
			t.Errorf("got %d changes, want none for an in-sync ruleset: %+v", len(changes), changes)
		}
	})

	t.Run("updates what has drifted", func(t *testing.T) {
		changes, err := planRulesetSet(context.Background(), args, []config.RulesetConfig{baseline},
			[]*github.RepositoryRuleset{drifted}, testLookup())
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(changes) != 1 || changes[0].Action != "update" {
			t.Fatalf("got %+v, want a single update", changes)
		}
		d, ok := changes[0].Details.(rulesetChange)
		if !ok {
			t.Fatalf("details are %T, want rulesetChange", changes[0].Details)
		}
		if d.ID != 12345 {
			t.Errorf("update targets ID %d, want the existing ruleset's 12345", d.ID)
		}
	})

	t.Run("plans a change when a reference cannot be resolved yet", func(t *testing.T) {
		withNewTeam := config.RulesetConfig{
			Name:         "baseline",
			Preset:       config.PresetBranchProtection,
			BypassActors: []config.BypassActorConfig{{Type: "Team", Team: "not-created-yet"}},
		}
		existing := roundTripThroughAPI(t, inSync, nil)
		changes, err := planRulesetSet(context.Background(), args, []config.RulesetConfig{withNewTeam},
			[]*github.RepositoryRuleset{existing}, testLookup())
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(changes) != 1 || changes[0].Action != "update" {
			t.Fatalf("got %+v, want an update so apply can resolve the team", changes)
		}
	})
}

func TestPlanRulesetCleanup(t *testing.T) {
	orgOwned := &github.RepositoryRuleset{
		ID:         github.Ptr(int64(7)),
		Name:       "legacy",
		SourceType: ptrTo(github.RulesetSourceTypeOrganization),
	}
	inherited := &github.RepositoryRuleset{
		ID:         github.Ptr(int64(8)),
		Name:       "enterprise-wide",
		SourceType: ptrTo(github.RulesetSourceTypeEnterprise),
	}
	args := rulesetScopeArgs{scope: scopeOrgRuleset, org: "myorg", orgLevel: true}
	existing := []*github.RepositoryRuleset{orgOwned, inherited}

	t.Run("reports without deleting", func(t *testing.T) {
		changes, unmanaged := planRulesetCleanup(args, nil, existing, false)
		if len(changes) != 0 {
			t.Errorf("got %d deletes, want none", len(changes))
		}
		if len(unmanaged) != 1 || unmanaged[0] != "legacy" {
			t.Errorf("unmanaged = %v, want just the org-owned ruleset", unmanaged)
		}
	})

	t.Run("deletes only what this scope owns", func(t *testing.T) {
		changes, _ := planRulesetCleanup(args, nil, existing, true)
		if len(changes) != 1 {
			t.Fatalf("got %d deletes, want 1: %+v", len(changes), changes)
		}
		if changes[0].Target != "legacy" || changes[0].Action != "delete" {
			t.Errorf("change = %+v, want a delete of legacy", changes[0])
		}
	})

	t.Run("leaves declared rulesets alone", func(t *testing.T) {
		declared := []config.RulesetConfig{{Name: "Legacy", Preset: config.PresetNoForcePush}}
		changes, unmanaged := planRulesetCleanup(args, declared, existing, true)
		if len(changes) != 0 || len(unmanaged) != 0 {
			t.Errorf("changes = %+v, unmanaged = %v; a declared name matches case-insensitively", changes, unmanaged)
		}
	})
}

func TestWarnSelfLockout(t *testing.T) {
	cfgWithFiles := func() *config.Root {
		cfg := &config.Root{}
		cfg.App.Org = "myorg"
		cfg.App.Files = []config.FileSpec{{Path: "README.md", Content: "hi"}}
		return cfg
	}

	t.Run("warns about a pull request rule with no bypass", func(t *testing.T) {
		got := warnSelfLockout(cfgWithFiles(), []config.RulesetConfig{
			{Name: "baseline", Preset: config.PresetBranchProtection},
		}, "organization")
		if len(got) != 1 || !strings.Contains(got[0], "Integration") {
			t.Errorf("warnings = %v, want one suggesting an Integration bypass actor", got)
		}
	})

	t.Run("stays quiet when the app can bypass", func(t *testing.T) {
		got := warnSelfLockout(cfgWithFiles(), []config.RulesetConfig{{
			Name:         "baseline",
			Preset:       config.PresetBranchProtection,
			BypassActors: []config.BypassActorConfig{{Type: "Integration", App: "self"}},
		}}, "organization")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none", got)
		}
	})

	t.Run("stays quiet when gomgr writes no files", func(t *testing.T) {
		cfg := &config.Root{}
		cfg.App.Org = "myorg"
		got := warnSelfLockout(cfg, []config.RulesetConfig{
			{Name: "baseline", Preset: config.PresetBranchProtection},
		}, "organization")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none", got)
		}
	})

	t.Run("stays quiet for a ruleset only being evaluated", func(t *testing.T) {
		got := warnSelfLockout(cfgWithFiles(), []config.RulesetConfig{{
			Name:        "baseline",
			Preset:      config.PresetBranchProtection,
			Enforcement: config.RulesetEnforcementEvaluate,
		}}, "organization")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none", got)
		}
	})
}

func TestApplyOrgRulesetUpsert(t *testing.T) {
	t.Run("creates", func(t *testing.T) {
		var gotBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/orgs/myorg/rulesets" {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "baseline"})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		ch := util.Change{
			Scope:   scopeOrgRuleset,
			Target:  "baseline",
			Action:  "create",
			Details: rulesetChange{Org: "myorg", Name: "baseline", Spec: spec},
		}
		if err := applyOrgRulesetUpsert(context.Background(), newTestClient(t, server), ch); err != nil {
			t.Fatalf("apply: %v", err)
		}
		if gotBody["name"] != "baseline" || gotBody["enforcement"] != "active" {
			t.Errorf("request body = %v", gotBody)
		}
	})

	t.Run("updates in place", func(t *testing.T) {
		var hit string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hit = r.Method + " " + r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99, "name": "baseline"})
		}))
		defer server.Close()

		spec, err := config.RulesetConfig{Name: "baseline", Preset: config.PresetBranchProtection}.Resolve()
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		ch := util.Change{
			Scope:   scopeOrgRuleset,
			Target:  "baseline",
			Action:  "update",
			Details: rulesetChange{Org: "myorg", ID: 99, Name: "baseline", Spec: spec},
		}
		if err := applyOrgRulesetUpsert(context.Background(), newTestClient(t, server), ch); err != nil {
			t.Fatalf("apply: %v", err)
		}
		if hit != "PUT /orgs/myorg/rulesets/99" {
			t.Errorf("hit %q, want PUT /orgs/myorg/rulesets/99", hit)
		}
	})
}

func TestApplyRepoRulesetDelete(t *testing.T) {
	var hit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = r.Method + " " + r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ch := util.Change{
		Scope:   scopeRepoRuleset,
		Target:  "infra/legacy",
		Action:  "delete",
		Details: rulesetChange{Org: "myorg", Repo: "infra", ID: 7, Name: "legacy"},
	}
	if err := applyRepoRulesetDelete(context.Background(), newTestClient(t, server), ch); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if hit != "DELETE /repos/myorg/infra/rulesets/7" {
		t.Errorf("hit %q, want DELETE /repos/myorg/infra/rulesets/7", hit)
	}
}

func TestApplyRulesetRejectsWrongDetails(t *testing.T) {
	ch := util.Change{Scope: scopeOrgRuleset, Action: "create", Details: map[string]any{"org": "myorg"}}
	if err := applyOrgRulesetUpsert(context.Background(), nil, ch); err == nil {
		t.Fatal("expected an error for details of the wrong type")
	}
}

func TestParseRepoConfigReadsRulesets(t *testing.T) {
	val := map[string]any{
		"permission": "push",
		"rulesets": []any{
			map[string]any{"name": "protect-main", "preset": config.PresetBranchProtection},
		},
	}
	settings, err := parseRepoConfig(val)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(settings.rulesets) != 1 || settings.rulesets[0].Name != "protect-main" {
		t.Errorf("rulesets = %+v", settings.rulesets)
	}
}

func TestResolveTemplateMergesRulesets(t *testing.T) {
	all := map[string]repoSettings{
		"base": {
			template: true,
			rulesets: []config.RulesetConfig{
				{Name: "protect-main", Preset: config.PresetBranchProtection},
				{Name: "protect-tags", Preset: config.PresetTagProtection},
			},
		},
	}
	child := repoSettings{
		from: "base",
		rulesets: []config.RulesetConfig{
			{Name: "protect-main", Preset: config.PresetStrictBranchProtection},
		},
	}

	got, err := resolveTemplate("child", child, all, "myorg")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(got.rulesets) != 2 {
		t.Fatalf("rulesets = %+v, want the child's plus the template's unique one", got.rulesets)
	}
	if got.rulesets[0].Name != "protect-main" || got.rulesets[0].Preset != config.PresetStrictBranchProtection {
		t.Errorf("the child's ruleset should win for a shared name, got %+v", got.rulesets[0])
	}
	if got.rulesets[1].Name != "protect-tags" {
		t.Errorf("the template's other ruleset should be inherited, got %+v", got.rulesets[1])
	}
}

func TestRulesetHandlersAreRegistered(t *testing.T) {
	for _, scope := range []string{scopeOrgRuleset, scopeRepoRuleset} {
		for _, action := range []string{"create", "update", "delete"} {
			if _, ok := defaultRegistry.Lookup(scope, action); !ok {
				t.Errorf("no handler registered for %s:%s", scope, action)
			}
		}
	}
	// Guard rails must go on after the file writes that push to the branches
	// they protect, and come off before the rest of the cleanup phase.
	if defaultRegistry.Precedence(scopeOrgRuleset, "create") <= defaultRegistry.Precedence("repo-file", "ensure") {
		t.Error("org rulesets should be created after repo files are written")
	}
	if defaultRegistry.Precedence(scopeOrgRuleset, "delete") <= defaultRegistry.Precedence(scopeOrgRuleset, "create") {
		t.Error("ruleset deletion belongs in the cleanup phase")
	}
}

func TestBuildPlanIncludesRulesets(t *testing.T) {
	var listedOrgRulesets, listedRepoRulesets bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orgs/myorg/rulesets":
			listedOrgRulesets = true
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case "/repos/myorg/infra/rulesets":
			listedRepoRulesets = true
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case "/orgs/myorg/repos":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 900, "name": "infra"},
			})
		default:
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
	defer server.Close()

	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Org: config.OrgConfig{
			Rulesets: []config.RulesetConfig{
				{Name: "org-baseline", Preset: config.PresetBranchProtection},
			},
		},
		Team: []config.TeamConfig{{
			Name: "Platform",
			Slug: "platform",
			Repositories: map[string]any{
				"infra": map[string]any{
					"permission": "push",
					"rulesets": []any{
						map[string]any{"name": "repo-tags", "preset": config.PresetTagProtection},
					},
				},
			},
		}},
	}

	plan, err := BuildPlan(context.Background(), newTestClient(t, server), cfg)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if !listedOrgRulesets || !listedRepoRulesets {
		t.Errorf("planning did not read rulesets (org=%v repo=%v)", listedOrgRulesets, listedRepoRulesets)
	}

	var org, repo int
	for _, ch := range plan.Changes {
		switch ch.Scope {
		case scopeOrgRuleset:
			org++
			if ch.Target != "org-baseline" || ch.Action != "create" {
				t.Errorf("org ruleset change = %+v", ch)
			}
		case scopeRepoRuleset:
			repo++
			if ch.Target != "infra/repo-tags" || ch.Action != "create" {
				t.Errorf("repo ruleset change = %+v", ch)
			}
		}
	}
	if org != 1 || repo != 1 {
		t.Errorf("got %d org and %d repo ruleset changes, want 1 each", org, repo)
	}
	if plan.Stats.Rulesets.Desired != 2 {
		t.Errorf("desired rulesets = %d, want 2", plan.Stats.Rulesets.Desired)
	}
}

// TestIdentityFreeBypassActorsAreIdempotent is a regression test for a bug that
// only a live organization exposed. gomgr used to send actor_id 1 for an
// OrganizationAdmin bypass actor. GitHub accepts that and then reports the
// actor back with no actor_id at all, so the next comparison saw a difference
// that did not exist and rewrote the ruleset — forever, on every run.
func TestIdentityFreeBypassActorsAreIdempotent(t *testing.T) {
	for _, kind := range []string{"OrganizationAdmin", "EnterpriseOwner", "DeployKey"} {
		t.Run(kind, func(t *testing.T) {
			spec, err := config.RulesetConfig{
				Name:         "guarded",
				Preset:       config.PresetBranchProtection,
				BypassActors: []config.BypassActorConfig{{Type: kind}},
			}.Resolve()
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			desired, err := buildRuleset(context.Background(), spec, true, "", testLookup())
			if err != nil {
				t.Fatalf("build: %v", err)
			}

			// This is what GitHub sends back: the type, the mode, no ID.
			actual := roundTripThroughAPI(t, desired, func(g map[string]any) {
				g["bypass_actors"] = []any{
					map[string]any{"actor_type": kind, "bypass_mode": "always"},
				}
			})

			same, err := rulesetMatches(actual, desired)
			if err != nil {
				t.Fatalf("compare: %v", err)
			}
			if !same {
				t.Errorf("a %s bypass actor reports drift against GitHub's own representation of it", kind)
			}
		})
	}
}
