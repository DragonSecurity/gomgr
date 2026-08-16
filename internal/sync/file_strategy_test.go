package sync

import (
	"testing"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func deciderFor(t *testing.T, org []config.RulesetConfig, perRepo map[string][]config.RulesetConfig, strategy string) *routeDecider {
	t.Helper()
	cfg := &config.Root{App: config.AppConfig{Org: "myorg", AppID: 4242}}
	cfg.App.FileChanges.Strategy = strategy
	cfg.Org.Rulesets = org

	st := &State{Org: "myorg", ActualRepos: []*github.Repository{
		{Name: github.Ptr("infra"), DefaultBranch: github.Ptr("main")},
		{Name: github.Ptr("legacy"), DefaultBranch: github.Ptr("master")},
	}}
	settings := map[string]repoSettings{}
	for repo, rs := range perRepo {
		settings[repo] = repoSettings{rulesets: rs}
	}
	return newRouteDecider(cfg, st, settings)
}

func TestRouteRespectsExplicitStrategy(t *testing.T) {
	blocking := []config.RulesetConfig{{Name: "protect", Preset: config.PresetBranchProtection}}

	d := deciderFor(t, blocking, nil, config.FileStrategyDirect)
	got, err := d.route("infra", "main")
	if err != nil || got.UsePullRequest {
		t.Errorf("direct strategy must override the rulesets, got %+v (%v)", got, err)
	}

	d = deciderFor(t, nil, nil, config.FileStrategyPullRequest)
	got, err = d.route("infra", "main")
	if err != nil || !got.UsePullRequest {
		t.Errorf("pull_request strategy must apply even with no rulesets, got %+v (%v)", got, err)
	}
}

func TestRouteDerivesFromRulesets(t *testing.T) {
	tests := []struct {
		name    string
		org     []config.RulesetConfig
		repo    string
		branch  string
		wantPR  bool
		because string
	}{
		{
			name: "pull_request rule on the default branch",
			org:  []config.RulesetConfig{{Name: "protect", Preset: config.PresetBranchProtection}},
			repo: "infra", branch: "main", wantPR: true,
			because: "the preset requires a pull request on ~DEFAULT_BRANCH",
		},
		{
			name: "required_status_checks blocks a direct push",
			org: []config.RulesetConfig{{
				Name: "require-dco", Target: config.RulesetTargetBranch,
				Enforcement: config.RulesetEnforcementActive,
				Rules: config.RulesetRules{RequiredStatusChecks: &config.RequiredStatusChecksRule{
					Checks: []config.StatusCheck{{Context: "DCO"}},
				}},
			}},
			repo: "infra", branch: "main", wantPR: true,
			because: "a check cannot have run on a commit that does not exist",
		},
		{
			name: "deletion-only ruleset does not block",
			org: []config.RulesetConfig{{
				Name: "protect-default-branch", Target: config.RulesetTargetBranch,
				Enforcement: config.RulesetEnforcementActive,
				Rules:       config.RulesetRules{Deletion: ptrTo(true)},
			}},
			repo: "infra", branch: "main", wantPR: false,
			because: "blocking deletion says nothing about pushing",
		},
		{
			name: "signatures alone do not force a pull request",
			org: []config.RulesetConfig{{
				Name: "signed", Target: config.RulesetTargetBranch,
				Enforcement: config.RulesetEnforcementActive,
				Rules:       config.RulesetRules{RequiredSignatures: ptrTo(true)},
			}},
			repo: "infra", branch: "main", wantPR: false,
			because: "commits made through the API are signed by GitHub",
		},
		{
			name: "evaluate mode rejects nothing",
			org: []config.RulesetConfig{{
				Name: "trial", Preset: config.PresetBranchProtection,
				Enforcement: config.RulesetEnforcementEvaluate,
			}},
			repo: "infra", branch: "main", wantPR: false,
			because: "report-only rulesets do not block",
		},
		{
			name: "a ruleset scoped to other repositories does not apply",
			org: []config.RulesetConfig{{
				Name: "protect", Preset: config.PresetBranchProtection,
				Conditions: &config.RulesetConditions{
					RepositoryName: &config.RepositoryNameCondition{Include: []string{"svc-*"}},
				},
			}},
			repo: "infra", branch: "main", wantPR: false,
			because: "infra does not match svc-*",
		},
		{
			name: "a ruleset scoped to a different branch does not apply",
			org: []config.RulesetConfig{{
				Name: "protect", Preset: config.PresetBranchProtection,
			}},
			repo: "infra", branch: "docs-site", wantPR: false,
			because: "~DEFAULT_BRANCH is main, not docs-site",
		},
		{
			name: "~DEFAULT_BRANCH follows the repository, not a guess",
			org: []config.RulesetConfig{{
				Name: "protect", Preset: config.PresetBranchProtection,
			}},
			repo: "legacy", branch: "master", wantPR: true,
			because: "legacy's default branch is master",
		},
		{
			name: "an excluded repository is not covered",
			org: []config.RulesetConfig{{
				Name: "protect", Preset: config.PresetBranchProtection,
				Conditions: &config.RulesetConditions{
					RefName:        &config.RefNameCondition{Include: []string{"~DEFAULT_BRANCH"}},
					RepositoryName: &config.RepositoryNameCondition{Include: []string{"~ALL"}, Exclude: []string{"infra"}},
				},
			}},
			repo: "infra", branch: "main", wantPR: false,
			because: "infra is excluded by name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deciderFor(t, tt.org, nil, "")
			got, err := d.route(tt.repo, tt.branch)
			if err != nil {
				t.Fatalf("route: %v", err)
			}
			if got.UsePullRequest != tt.wantPR {
				t.Errorf("UsePullRequest = %v, want %v — %s\n  reason given: %s",
					got.UsePullRequest, tt.wantPR, tt.because, got.Reason)
			}
		})
	}
}

func TestRouteHonoursGomgrsOwnBypass(t *testing.T) {
	withBypass := func(actors ...config.BypassActorConfig) []config.RulesetConfig {
		return []config.RulesetConfig{{
			Name: "protect", Preset: config.PresetBranchProtection, BypassActors: actors,
		}}
	}

	t.Run("app: self bypassing always allows a direct push", func(t *testing.T) {
		d := deciderFor(t, withBypass(config.BypassActorConfig{Type: "Integration", App: "self"}), nil, "")
		got, _ := d.route("infra", "main")
		if got.UsePullRequest {
			t.Errorf("gomgr is exempt, so it can push directly: %s", got.Reason)
		}
	})

	t.Run("the app's numeric ID also counts", func(t *testing.T) {
		d := deciderFor(t, withBypass(config.BypassActorConfig{Type: "Integration", App: "4242"}), nil, "")
		got, _ := d.route("infra", "main")
		if got.UsePullRequest {
			t.Errorf("the configured app ID is gomgr's own: %s", got.Reason)
		}
	})

	t.Run("a pull_request-mode bypass still means use a pull request", func(t *testing.T) {
		d := deciderFor(t, withBypass(config.BypassActorConfig{
			Type: "Integration", App: "self", Mode: config.BypassModePullRequest,
		}), nil, "")
		got, _ := d.route("infra", "main")
		if !got.UsePullRequest {
			t.Error("an actor limited to bypassing via a pull request is being told to use one")
		}
	})

	t.Run("someone else's app does not exempt gomgr", func(t *testing.T) {
		d := deciderFor(t, withBypass(config.BypassActorConfig{Type: "Integration", App: "9999"}), nil, "")
		got, _ := d.route("infra", "main")
		if !got.UsePullRequest {
			t.Error("app 9999 is not gomgr")
		}
	})

	t.Run("an org admin bypass does not cover an app", func(t *testing.T) {
		d := deciderFor(t, withBypass(config.BypassActorConfig{Type: "OrganizationAdmin"}), nil, "")
		got, _ := d.route("infra", "main")
		if !got.UsePullRequest {
			t.Error("gomgr authenticates as an app, not as an org admin")
		}
	})
}

func TestRouteConsidersRepositoryRulesets(t *testing.T) {
	perRepo := map[string][]config.RulesetConfig{
		"infra": {{Name: "locked", Preset: config.PresetStrictBranchProtection}},
	}
	d := deciderFor(t, nil, perRepo, "")

	got, err := d.route("infra", "main")
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if !got.UsePullRequest {
		t.Errorf("a repository's own ruleset must count too: %s", got.Reason)
	}

	// A repository without that ruleset is unaffected.
	other, _ := d.route("legacy", "master")
	if other.UsePullRequest {
		t.Errorf("legacy declares no ruleset: %s", other.Reason)
	}
}
