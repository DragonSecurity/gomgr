package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestResolveExpandsPreset(t *testing.T) {
	spec := RulesetConfig{Name: "default", Preset: PresetBranchProtection}

	got, err := spec.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got.Target != RulesetTargetBranch {
		t.Errorf("target = %q, want %q", got.Target, RulesetTargetBranch)
	}
	if got.Enforcement != RulesetEnforcementActive {
		t.Errorf("enforcement = %q, want %q", got.Enforcement, RulesetEnforcementActive)
	}
	if !isTrue(got.Rules.NonFastForward) {
		t.Error("non_fast_forward should come from the preset")
	}
	if got.Rules.PullRequest == nil || got.Rules.PullRequest.RequiredApprovingReviewCount != 1 {
		t.Errorf("pull_request = %+v, want 1 required approval from the preset", got.Rules.PullRequest)
	}
	if got.Conditions == nil || got.Conditions.RefName == nil ||
		len(got.Conditions.RefName.Include) != 1 || got.Conditions.RefName.Include[0] != refDefaultBranch {
		t.Errorf("conditions = %+v, want the preset's default-branch condition", got.Conditions)
	}
}

func TestResolveConfigOverridesPreset(t *testing.T) {
	spec := RulesetConfig{
		Name:        "default",
		Preset:      PresetBranchProtection,
		Enforcement: RulesetEnforcementEvaluate,
		Rules: RulesetRules{
			// A rule named in the config replaces the preset's version wholesale.
			PullRequest: &PullRequestRule{RequiredApprovingReviewCount: 3},
			// An explicit false switches a preset rule back off.
			Deletion: boolPtr(false),
		},
	}

	got, err := spec.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got.Enforcement != RulesetEnforcementEvaluate {
		t.Errorf("enforcement = %q, want the config's %q", got.Enforcement, RulesetEnforcementEvaluate)
	}
	if got.Rules.PullRequest.RequiredApprovingReviewCount != 3 {
		t.Errorf("required approvals = %d, want the config's 3", got.Rules.PullRequest.RequiredApprovingReviewCount)
	}
	if got.Rules.PullRequest.DismissStaleReviewsOnPush {
		t.Error("the preset's pull_request fields should not survive alongside the config's")
	}
	if isTrue(got.Rules.Deletion) {
		t.Error("deletion: false should switch the preset's rule off")
	}
	if !isTrue(got.Rules.NonFastForward) {
		t.Error("a rule the config does not mention should still come from the preset")
	}
}

func TestResolveUnknownPreset(t *testing.T) {
	_, err := RulesetConfig{Name: "x", Preset: "nope"}.Resolve()
	if err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
	if !strings.Contains(err.Error(), PresetBranchProtection) {
		t.Errorf("error %q should list the known presets", err)
	}
}

func TestPresetsAreIndependentCopies(t *testing.T) {
	first, err := RulesetConfig{Name: "a", Preset: PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	first.Rules.PullRequest.RequiredApprovingReviewCount = 99

	second, err := RulesetConfig{Name: "b", Preset: PresetBranchProtection}.Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if second.Rules.PullRequest.RequiredApprovingReviewCount != 1 {
		t.Errorf("second ruleset saw %d approvals; presets are aliasing each other",
			second.Rules.PullRequest.RequiredApprovingReviewCount)
	}
}

func TestEveryPresetValidates(t *testing.T) {
	for name := range RulesetPresets() {
		t.Run(name, func(t *testing.T) {
			spec := RulesetConfig{Name: "test", Preset: name}
			if err := ValidateRulesets(ScopeOrg, "org.yaml", []RulesetConfig{spec}); err != nil {
				t.Errorf("preset %q does not validate: %v", name, err)
			}
		})
	}
}

func TestParseRulesets(t *testing.T) {
	var raw any
	src := `
- name: protect-main
  preset: branch-protection
  bypass_actors:
    - type: Team
      team: platform
      mode: pull_request
  rules:
    required_status_checks:
      strict: true
      checks:
        - context: build
`
	if err := yaml.Unmarshal([]byte(src), &raw); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	got, err := ParseRulesets(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d rulesets, want 1", len(got))
	}
	rs := got[0]
	if rs.Name != "protect-main" || rs.Preset != PresetBranchProtection {
		t.Errorf("got %+v, want name protect-main with the branch-protection preset", rs)
	}
	if len(rs.BypassActors) != 1 || rs.BypassActors[0].Team != "platform" {
		t.Errorf("bypass actors = %+v", rs.BypassActors)
	}
	if rs.Rules.RequiredStatusChecks == nil || len(rs.Rules.RequiredStatusChecks.Checks) != 1 {
		t.Fatalf("required_status_checks = %+v", rs.Rules.RequiredStatusChecks)
	}
	if rs.Rules.RequiredStatusChecks.Checks[0].Context != "build" {
		t.Errorf("check context = %q, want build", rs.Rules.RequiredStatusChecks.Checks[0].Context)
	}
}

func TestParseRulesetsRejectsNonList(t *testing.T) {
	if _, err := ParseRulesets(map[string]any{"name": "x"}); err == nil {
		t.Fatal("expected an error for a non-list rulesets value")
	}
}

func TestValidateRulesets(t *testing.T) {
	tests := []struct {
		name    string
		scope   RulesetScope
		ruleset RulesetConfig
		wantErr string
	}{
		{
			name:    "empty name",
			scope:   ScopeOrg,
			ruleset: RulesetConfig{Preset: PresetNoForcePush},
			wantErr: "name must not be empty",
		},
		{
			name:    "no rules",
			scope:   ScopeOrg,
			ruleset: RulesetConfig{Name: "hollow"},
			wantErr: "no rules enabled",
		},
		{
			name:    "invalid target",
			scope:   ScopeOrg,
			ruleset: RulesetConfig{Name: "x", Target: "commit", Rules: RulesetRules{Deletion: boolPtr(true)}},
			wantErr: "invalid target",
		},
		{
			name:    "invalid enforcement",
			scope:   ScopeOrg,
			ruleset: RulesetConfig{Name: "x", Enforcement: "on", Rules: RulesetRules{Deletion: boolPtr(true)}},
			wantErr: "invalid enforcement",
		},
		{
			name:  "repository_name at repo scope",
			scope: ScopeRepo,
			ruleset: RulesetConfig{
				Name:       "x",
				Rules:      RulesetRules{Deletion: boolPtr(true)},
				Conditions: &RulesetConditions{RepositoryName: &RepositoryNameCondition{Include: []string{"~ALL"}}},
			},
			wantErr: "only valid on organization rulesets",
		},
		{
			name:  "pull request rule on a tag ruleset",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:   "x",
				Target: RulesetTargetTag,
				Rules:  RulesetRules{PullRequest: &PullRequestRule{}},
			},
			wantErr: "only applies to a branch ruleset",
		},
		{
			name:  "push rule on a branch ruleset",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:  "x",
				Rules: RulesetRules{MaxFileSize: int64Ptr(100)},
			},
			wantErr: "only applies to a push ruleset",
		},
		{
			name:  "unknown bypass actor type",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:         "x",
				Rules:        RulesetRules{Deletion: boolPtr(true)},
				BypassActors: []BypassActorConfig{{Type: "Wizard"}},
			},
			wantErr: "invalid type",
		},
		{
			name:  "team bypass actor without a team",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:         "x",
				Rules:        RulesetRules{Deletion: boolPtr(true)},
				BypassActors: []BypassActorConfig{{Type: "Team"}},
			},
			wantErr: "needs a team slug",
		},
		{
			name:  "integration bypass actor with a bad app",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:         "x",
				Rules:        RulesetRules{Deletion: boolPtr(true)},
				BypassActors: []BypassActorConfig{{Type: "Integration", App: "gomgr"}},
			},
			wantErr: "numeric GitHub App ID",
		},
		{
			name:  "pattern rule without an operator",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:  "x",
				Rules: RulesetRules{CommitMessagePattern: &PatternRule{Pattern: "x"}},
			},
			wantErr: "invalid operator",
		},
		{
			name:  "too many approvals",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:  "x",
				Rules: RulesetRules{PullRequest: &PullRequestRule{RequiredApprovingReviewCount: 11}},
			},
			wantErr: "must be 0-10",
		},
		{
			name:  "status checks without any check",
			scope: ScopeOrg,
			ruleset: RulesetConfig{
				Name:  "x",
				Rules: RulesetRules{RequiredStatusChecks: &RequiredStatusChecksRule{}},
			},
			wantErr: "checks must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRulesets(tt.scope, "test", []RulesetConfig{tt.ruleset})
			if err == nil {
				t.Fatalf("expected an error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRulesetsRejectsDuplicateNames(t *testing.T) {
	rulesets := []RulesetConfig{
		{Name: "protect", Preset: PresetNoForcePush},
		{Name: "Protect", Preset: PresetTagProtection},
	}
	err := ValidateRulesets(ScopeOrg, "org.yaml", rulesets)
	if err == nil || !strings.Contains(err.Error(), "duplicate ruleset name") {
		t.Fatalf("err = %v, want a duplicate-name error", err)
	}
}

func TestRulesEmptyIgnoresExplicitFalse(t *testing.T) {
	if !(RulesetRules{Deletion: boolPtr(false)}).IsEmpty() {
		t.Error("a rule set holding only `false` enforces nothing and should count as empty")
	}
	if (RulesetRules{Deletion: boolPtr(true)}).IsEmpty() {
		t.Error("an enabled rule should not count as empty")
	}
}

func TestBypassActorDefaults(t *testing.T) {
	a := BypassActorConfig{Type: "organizationadmin"}
	if got := a.NormalizedType(); got != BypassActorTypeOrganizationAdmin {
		t.Errorf("NormalizedType() = %q, want %q", got, BypassActorTypeOrganizationAdmin)
	}
	if got := a.BypassMode(); got != BypassModeAlways {
		t.Errorf("BypassMode() = %q, want %q", got, BypassModeAlways)
	}
	if !a.IdentifiedByTypeAlone() {
		t.Error("OrganizationAdmin is identified by its type; GitHub reports it with no actor_id")
	}
}

func TestBypassActorEnterpriseOwner(t *testing.T) {
	// Not in go-github's enumeration, but what an enterprise-owned org returns.
	a := BypassActorConfig{Type: "EnterpriseOwner"}
	if got := a.NormalizedType(); got != BypassActorTypeEnterpriseOwner {
		t.Errorf("NormalizedType() = %q, want %q", got, BypassActorTypeEnterpriseOwner)
	}
	if !a.IdentifiedByTypeAlone() {
		t.Error("EnterpriseOwner carries no actor_id")
	}
	ruleset := RulesetConfig{
		Name:         "main",
		Rules:        RulesetRules{Deletion: boolPtr(true)},
		BypassActors: []BypassActorConfig{a},
	}
	if err := ValidateRulesets(ScopeRepo, "test", []RulesetConfig{ruleset}); err != nil {
		t.Errorf("EnterpriseOwner should validate: %v", err)
	}
}

func TestBypassActorTeamIsNotIdentifiedByTypeAlone(t *testing.T) {
	a := BypassActorConfig{Type: "Team", Team: "platform"}
	if a.IdentifiedByTypeAlone() {
		t.Error("a Team actor needs its ID")
	}
}

func int64Ptr(v int64) *int64 { return &v }
