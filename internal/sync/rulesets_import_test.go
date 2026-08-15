package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-github/v88/github"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func testImportLookup() *importLookup {
	return &importLookup{
		teamSlugByID: map[int64]string{77: "platform"},
		repoNameByID: map[int64]string{900: "infra"},
		appID:        4242,
	}
}

// TestImportRoundTripsEveryPreset is the property that matters: a ruleset gomgr
// built from a preset, once read back off GitHub, must be recognized as that
// same preset rather than transcribed as forty lines of expanded rules.
func TestImportRoundTripsEveryPreset(t *testing.T) {
	for _, name := range config.PresetNames() {
		t.Run(name, func(t *testing.T) {
			spec, err := config.RulesetConfig{Name: "adopted", Preset: name}.Resolve()
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			built, err := buildRuleset(context.Background(), spec, true, "", testLookup())
			if err != nil {
				t.Fatalf("build: %v", err)
			}

			// Go through the wire, including the parameters GitHub fills in.
			live := roundTripThroughAPI(t, built, func(g map[string]any) {
				rules, _ := g["rules"].([]any)
				for _, raw := range rules {
					rule, _ := raw.(map[string]any)
					if rule["type"] != "pull_request" {
						continue
					}
					params, _ := rule["parameters"].(map[string]any)
					params["allowed_merge_methods"] = []any{"merge", "squash", "rebase"}
				}
			})

			got := rulesetToConfig(live, true, testImportLookup())
			if got.Preset != name {
				t.Errorf("preset = %q, want %q (rules came back as %+v)", got.Preset, name, got.Rules)
			}
			if !got.Rules.IsEmpty() {
				t.Errorf("rules should collapse into the preset, got %+v", got.Rules)
			}
			if got.Name != "adopted" {
				t.Errorf("name = %q, want adopted", got.Name)
			}
		})
	}
}

// TestImportedConfigReproducesTheRuleset closes the loop the other way: feeding
// the imported configuration back through the planner must rebuild the ruleset
// that was there, so adopting a ruleset never changes what it enforces.
func TestImportedConfigReproducesTheRuleset(t *testing.T) {
	original := &github.RepositoryRuleset{
		Name:        "Protect Main",
		Target:      ptrTo(github.RulesetTargetBranch),
		Enforcement: github.RulesetEnforcementActive,
		BypassActors: []*github.BypassActor{
			{
				ActorID:    github.Ptr(int64(77)),
				ActorType:  ptrTo(github.BypassActorTypeTeam),
				BypassMode: ptrTo(github.BypassModePullRequest),
			},
			{
				ActorID:    github.Ptr(int64(4242)),
				ActorType:  ptrTo(github.BypassActorTypeIntegration),
				BypassMode: ptrTo(github.BypassModeAlways),
			},
		},
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: []string{"~DEFAULT_BRANCH"},
				Exclude: []string{},
			},
			RepositoryName: &github.RepositoryRulesetRepositoryNamesConditionParameters{
				Include: []string{"svc-*"},
				Exclude: []string{"svc-sandbox"},
			},
		},
		Rules: &github.RepositoryRulesetRules{
			Deletion:       &github.EmptyRuleParameters{},
			NonFastForward: &github.EmptyRuleParameters{},
			PullRequest: &github.PullRequestRuleParameters{
				RequiredApprovingReviewCount: 2,
				RequireCodeOwnerReview:       true,
			},
			RequiredStatusChecks: &github.RequiredStatusChecksRuleParameters{
				StrictRequiredStatusChecksPolicy: true,
				RequiredStatusChecks: []*github.RuleStatusCheck{
					{Context: "build"},
					{Context: "test"},
				},
			},
		},
	}

	imported := rulesetToConfig(original, true, testImportLookup())

	// The imported form must be legal configuration.
	if err := config.ValidateRulesets(config.ScopeOrg, "imported", []config.RulesetConfig{imported}); err != nil {
		t.Fatalf("imported ruleset does not validate: %v", err)
	}

	// The named references must have survived as names, not raw IDs.
	if len(imported.BypassActors) != 2 {
		t.Fatalf("bypass actors = %+v", imported.BypassActors)
	}
	if imported.BypassActors[0].Team != "platform" || imported.BypassActors[0].Mode != "pull_request" {
		t.Errorf("team actor = %+v, want the slug platform bypassing via pull request", imported.BypassActors[0])
	}
	if imported.BypassActors[1].App != "self" {
		t.Errorf("integration actor = %+v, want app: self for gomgr's own app", imported.BypassActors[1])
	}

	resolved, err := imported.Resolve()
	if err != nil {
		t.Fatalf("resolve imported: %v", err)
	}
	rebuilt, err := buildRuleset(context.Background(), resolved, true, "", testLookup())
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	same, err := rulesetMatches(original, rebuilt)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if !same {
		o, _ := json.Marshal(original)
		r, _ := json.Marshal(rebuilt)
		t.Errorf("adopting the ruleset changed it\n original: %s\n rebuilt:  %s", o, r)
	}
}

func TestImportDropsDefaultConditions(t *testing.T) {
	rs := &github.RepositoryRuleset{
		Name:        "everything",
		Target:      ptrTo(github.RulesetTargetBranch),
		Enforcement: github.RulesetEnforcementActive,
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: []string{"~ALL"}, Exclude: []string{},
			},
			RepositoryName: &github.RepositoryRulesetRepositoryNamesConditionParameters{
				Include: []string{"~ALL"}, Exclude: []string{},
			},
		},
		Rules: &github.RepositoryRulesetRules{NonFastForward: &github.EmptyRuleParameters{}},
	}

	got := rulesetToConfig(rs, true, testImportLookup())
	if got.Conditions != nil {
		t.Errorf("conditions = %+v; a ~ALL/~ALL selector is what gomgr emits by default and should not be restated", got.Conditions)
	}
}

func TestImportKeepsUnknownActorIDs(t *testing.T) {
	rs := &github.RepositoryRuleset{
		Name:        "guarded",
		Target:      ptrTo(github.RulesetTargetBranch),
		Enforcement: github.RulesetEnforcementActive,
		BypassActors: []*github.BypassActor{
			{ActorID: github.Ptr(int64(999)), ActorType: ptrTo(github.BypassActorTypeTeam), BypassMode: ptrTo(github.BypassModeAlways)},
			{ActorID: github.Ptr(int64(5)), ActorType: ptrTo(github.BypassActorTypeRepositoryRole), BypassMode: ptrTo(github.BypassModeAlways)},
			{ActorID: github.Ptr(int64(1)), ActorType: ptrTo(github.BypassActorTypeOrganizationAdmin), BypassMode: ptrTo(github.BypassModeAlways)},
		},
		Rules: &github.RepositoryRulesetRules{NonFastForward: &github.EmptyRuleParameters{}},
	}

	got := rulesetToConfig(rs, true, testImportLookup())
	if got.BypassActors[0].Team != "" || got.BypassActors[0].ActorID != 999 {
		t.Errorf("a team whose slug is unknown should keep its ID, got %+v", got.BypassActors[0])
	}
	if got.BypassActors[1].ActorID != 5 {
		t.Errorf("repository role should keep its ID, got %+v", got.BypassActors[1])
	}
	if got.BypassActors[2].ActorID != 0 {
		t.Errorf("OrganizationAdmin has a fixed ID and should not restate it, got %+v", got.BypassActors[2])
	}
}

func TestImportDropsGitHubSuppliedFalses(t *testing.T) {
	rs := &github.RepositoryRuleset{
		Name:        "patterned",
		Target:      ptrTo(github.RulesetTargetBranch),
		Enforcement: github.RulesetEnforcementActive,
		Rules: &github.RepositoryRulesetRules{
			CommitMessagePattern: &github.PatternRuleParameters{
				Operator: github.PatternRuleOperatorContains,
				Pattern:  "Signed-off-by:",
				Negate:   github.Ptr(false),
			},
			RequiredStatusChecks: &github.RequiredStatusChecksRuleParameters{
				RequiredStatusChecks: []*github.RuleStatusCheck{{Context: "build"}},
				DoNotEnforceOnCreate: github.Ptr(false),
			},
		},
	}

	got := rulesetToConfig(rs, true, testImportLookup())
	if got.Rules.CommitMessagePattern.Negate != nil {
		t.Error("negate: false is GitHub's default, not a choice worth writing down")
	}
	if got.Rules.RequiredStatusChecks.DoNotEnforceOnCreate != nil {
		t.Error("do_not_enforce_on_create: false should be dropped")
	}
}

// importServer answers the endpoints ImportRulesets calls.
func importServer(t *testing.T, orgRulesets map[string]any, repoRulesets map[string]map[string]any) *httptest.Server {
	t.Helper()

	encodeSummaries := func(w http.ResponseWriter, set map[string]any) {
		var out []map[string]any
		for name, body := range set {
			id := body.(map[string]any)["id"]
			out = append(out, map[string]any{"id": id, "name": name})
		}
		_ = json.NewEncoder(w).Encode(out)
	}
	findByID := func(set map[string]any, id string) (map[string]any, bool) {
		for _, body := range set {
			b := body.(map[string]any)
			if fmtID(b["id"]) == id {
				return b, true
			}
		}
		return nil, false
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/orgs/myorg/teams":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 77, "slug": "platform"}})

		case path == "/orgs/myorg/repos":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 900, "name": "infra"},
				{"id": 901, "name": "orphan"},
			})

		case path == "/orgs/myorg/rulesets":
			encodeSummaries(w, orgRulesets)

		case strings.HasPrefix(path, "/orgs/myorg/rulesets/"):
			id := strings.TrimPrefix(path, "/orgs/myorg/rulesets/")
			if body, ok := findByID(orgRulesets, id); ok {
				_ = json.NewEncoder(w).Encode(body)
				return
			}
			http.NotFound(w, r)

		case strings.HasSuffix(path, "/rulesets"):
			repo := strings.TrimSuffix(strings.TrimPrefix(path, "/repos/myorg/"), "/rulesets")
			encodeSummaries(w, repoRulesets[repo])

		case strings.Contains(path, "/rulesets/"):
			parts := strings.SplitN(strings.TrimPrefix(path, "/repos/myorg/"), "/rulesets/", 2)
			if len(parts) == 2 {
				if body, ok := findByID(repoRulesets[parts[0]], parts[1]); ok {
					_ = json.NewEncoder(w).Encode(body)
					return
				}
			}
			http.NotFound(w, r)

		default:
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
}

func fmtID(v any) string {
	switch n := v.(type) {
	case int:
		return strconv.Itoa(n)
	case float64:
		return strconv.FormatInt(int64(n), 10)
	}
	return ""
}

func liveRuleset(id int, name string, sourceType string) map[string]any {
	return map[string]any{
		"id":          id,
		"name":        name,
		"target":      "branch",
		"enforcement": "active",
		"source_type": sourceType,
		"rules":       []any{map[string]any{"type": "non_fast_forward"}},
	}
}

func TestImportRulesets(t *testing.T) {
	server := importServer(t,
		map[string]any{
			"hand-made-org-rule": liveRuleset(1, "hand-made-org-rule", "Organization"),
			"already-declared":   liveRuleset(2, "already-declared", "Organization"),
			"from-enterprise":    liveRuleset(3, "from-enterprise", "Enterprise"),
		},
		map[string]map[string]any{
			"infra":  {"hand-made-repo-rule": liveRuleset(10, "hand-made-repo-rule", "Repository")},
			"orphan": {"orphan-rule": liveRuleset(11, "orphan-rule", "Repository")},
		},
	)
	defer server.Close()

	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Org: config.OrgConfig{
			Rulesets: []config.RulesetConfig{
				{Name: "Already-Declared", Preset: config.PresetNoForcePush},
			},
		},
		Team: []config.TeamConfig{{
			Name:         "Platform",
			Slug:         "platform",
			Repositories: map[string]any{"infra": "maintain"},
		}},
	}

	result, err := ImportRulesets(context.Background(), newTestClient(t, server), cfg, ImportOptions{})
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if len(result.Org) != 1 || result.Org[0].Spec.Name != "hand-made-org-rule" {
		t.Errorf("org rulesets = %+v, want just the hand-made one", specNames(result.Org))
	}
	if result.AlreadyDeclared != 1 {
		t.Errorf("AlreadyDeclared = %d, want 1 (matched case-insensitively)", result.AlreadyDeclared)
	}
	if got := result.Repos["infra"]; len(got) != 1 || got[0].Spec.Name != "hand-made-repo-rule" {
		t.Errorf("infra rulesets = %+v", specNames(got))
	}
	if len(result.Unmanaged) != 1 || result.Unmanaged[0] != "orphan" {
		t.Errorf("Unmanaged = %v, want [orphan] — it has a ruleset but no team file to write to", result.Unmanaged)
	}
	if _, wrote := result.Repos["orphan"]; wrote {
		t.Error("an unmanaged repo must not be queued for writing; there is no home for it")
	}
	if result.Scanned != 2 {
		t.Errorf("Scanned = %d, want 2", result.Scanned)
	}
	if result.Total() != 2 {
		t.Errorf("Total() = %d, want 2", result.Total())
	}
}

func TestImportRulesetsHonoursOnlyGlob(t *testing.T) {
	server := importServer(t, map[string]any{},
		map[string]map[string]any{
			"infra": {"repo-rule": liveRuleset(10, "repo-rule", "Repository")},
		},
	)
	defer server.Close()

	cfg := &config.Root{
		App:  config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{{Name: "Platform", Repositories: map[string]any{"infra": "maintain"}}},
	}

	result, err := ImportRulesets(context.Background(), newTestClient(t, server), cfg,
		ImportOptions{Only: []string{"nothing-*"}})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Scanned != 0 || result.Total() != 0 {
		t.Errorf("scanned %d / found %d, want the glob to exclude everything", result.Scanned, result.Total())
	}
}

func specNames(imported []ImportedRuleset) []string {
	out := make([]string, 0, len(imported))
	for _, i := range imported {
		out = append(out, i.Spec.Name)
	}
	return out
}
