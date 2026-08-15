package config

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Ruleset targets. A ruleset protects refs (branch/tag) or the push itself.
const (
	RulesetTargetBranch = "branch"
	RulesetTargetTag    = "tag"
	RulesetTargetPush   = "push"
)

// Ruleset enforcement levels. "evaluate" reports what would have been blocked
// without blocking it — the safe way to roll a new guardrail out.
const (
	RulesetEnforcementActive   = "active"
	RulesetEnforcementEvaluate = "evaluate"
	RulesetEnforcementDisabled = "disabled"
)

// Bypass actor types accepted in YAML. These mirror the GitHub API values and
// are case-sensitive there, so they are matched case-insensitively here and
// normalized on the way out.
const (
	BypassActorTypeIntegration       = "Integration"
	BypassActorTypeOrganizationAdmin = "OrganizationAdmin"
	BypassActorTypeRepositoryRole    = "RepositoryRole"
	BypassActorTypeTeam              = "Team"
	BypassActorTypeDeployKey         = "DeployKey"
)

// Bypass modes. "pull_request" lets the actor bypass only via a pull request.
const (
	BypassModeAlways      = "always"
	BypassModePullRequest = "pull_request"
)

// orgAdminActorID is the fixed actor ID GitHub assigns to the OrganizationAdmin
// bypass actor type; it carries no per-org identity.
const orgAdminActorID = 1

// RulesetConfig declares a GitHub ruleset — the successor to branch protection,
// and the mechanism behind org-wide guard rails.
//
// The same shape is used at both scopes:
//   - org.yaml `rulesets:` applies across the organization, narrowed by
//     Conditions.RepositoryName.
//   - a repository entry in teams/*.yaml may carry its own `rulesets:`, which
//     apply to that repository alone and stack on top of the org's.
//
// Preset names a built-in guard rail (see RulesetPresets) that supplies the
// target, conditions and rules. Anything set explicitly alongside it wins, at
// whole-rule granularity: naming a rule key replaces the preset's version of
// that rule outright rather than merging field by field.
type RulesetConfig struct {
	Name         string              `yaml:"name"`
	Preset       string              `yaml:"preset,omitempty"`
	Target       string              `yaml:"target,omitempty"`      // branch|tag|push (default branch)
	Enforcement  string              `yaml:"enforcement,omitempty"` // active|evaluate|disabled (default active)
	Conditions   *RulesetConditions  `yaml:"conditions,omitempty"`
	BypassActors []BypassActorConfig `yaml:"bypass_actors,omitempty"`
	Rules        RulesetRules        `yaml:"rules,omitempty"`
}

// RulesetConditions narrows which refs and repositories a ruleset applies to.
// RepositoryName is only meaningful for organization-level rulesets; a
// repository-level ruleset already knows its repository.
type RulesetConditions struct {
	RefName        *RefNameCondition        `yaml:"ref_name,omitempty"`
	RepositoryName *RepositoryNameCondition `yaml:"repository_name,omitempty"`
}

// RefNameCondition selects refs by fnmatch pattern. GitHub also accepts the
// magic values "~ALL" and "~DEFAULT_BRANCH".
type RefNameCondition struct {
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// RepositoryNameCondition selects repositories by fnmatch pattern, or "~ALL".
// Protected, when set, additionally restricts to repositories GitHub considers
// protected.
type RepositoryNameCondition struct {
	Include   []string `yaml:"include,omitempty"`
	Exclude   []string `yaml:"exclude,omitempty"`
	Protected *bool    `yaml:"protected,omitempty"`
}

// BypassActorConfig grants an actor the right to bypass the ruleset.
//
// Exactly one identifier applies per Type:
//   - Team: Team names a team slug, resolved to its ID against the org.
//   - Integration: App is either "self" (gomgr's own app_id, so its file-sync
//     pushes survive a pull-request rule) or a numeric GitHub App ID.
//   - RepositoryRole: ActorID is the role ID, which the roles API reports.
//   - OrganizationAdmin, DeployKey: no identifier; GitHub fixes the ID.
type BypassActorConfig struct {
	Type    string `yaml:"type"`
	Team    string `yaml:"team,omitempty"`
	App     string `yaml:"app,omitempty"`
	ActorID int64  `yaml:"actor_id,omitempty"`
	Mode    string `yaml:"mode,omitempty"` // always|pull_request (default always)
}

// RulesetRules is the YAML face of GitHub's rule list. Every field is a pointer
// so that "unset" is distinguishable from "explicitly off" — a preset can turn
// a rule on and the config can turn it back off with `deletion: false`.
//
// Field order here is also the merge order used by mergeRules; adding a field
// requires no change there as long as it stays nil-able.
type RulesetRules struct {
	// Branch and tag rules.
	Creation                 *bool                     `yaml:"creation,omitempty"`
	Update                   *UpdateRule               `yaml:"update,omitempty"`
	Deletion                 *bool                     `yaml:"deletion,omitempty"`
	RequiredLinearHistory    *bool                     `yaml:"required_linear_history,omitempty"`
	RequiredSignatures       *bool                     `yaml:"required_signatures,omitempty"`
	NonFastForward           *bool                     `yaml:"non_fast_forward,omitempty"`
	PullRequest              *PullRequestRule          `yaml:"pull_request,omitempty"`
	RequiredStatusChecks     *RequiredStatusChecksRule `yaml:"required_status_checks,omitempty"`
	RequiredDeployments      *RequiredDeploymentsRule  `yaml:"required_deployments,omitempty"`
	MergeQueue               *MergeQueueRule           `yaml:"merge_queue,omitempty"`
	CommitMessagePattern     *PatternRule              `yaml:"commit_message_pattern,omitempty"`
	CommitAuthorEmailPattern *PatternRule              `yaml:"commit_author_email_pattern,omitempty"`
	CommitterEmailPattern    *PatternRule              `yaml:"committer_email_pattern,omitempty"`
	BranchNamePattern        *PatternRule              `yaml:"branch_name_pattern,omitempty"`
	TagNamePattern           *PatternRule              `yaml:"tag_name_pattern,omitempty"`
	Workflows                *WorkflowsRule            `yaml:"workflows,omitempty"`
	CodeScanning             *CodeScanningRule         `yaml:"code_scanning,omitempty"`

	// Push rules. These require GitHub Enterprise Cloud on private repositories.
	FileExtensionRestriction *FileExtensionRestrictionRule `yaml:"file_extension_restriction,omitempty"`
	FilePathRestriction      *FilePathRestrictionRule      `yaml:"file_path_restriction,omitempty"`
	MaxFilePathLength        *int                          `yaml:"max_file_path_length,omitempty"`
	MaxFileSize              *int64                        `yaml:"max_file_size,omitempty"`
}

// UpdateRule restricts updates to a matching ref.
type UpdateRule struct {
	AllowsFetchAndMerge bool `yaml:"allows_fetch_and_merge,omitempty"`
}

// PullRequestRule requires changes to arrive through a pull request.
type PullRequestRule struct {
	RequiredApprovingReviewCount   int      `yaml:"required_approving_review_count,omitempty"`
	DismissStaleReviewsOnPush      bool     `yaml:"dismiss_stale_reviews_on_push,omitempty"`
	RequireCodeOwnerReview         bool     `yaml:"require_code_owner_review,omitempty"`
	RequireLastPushApproval        bool     `yaml:"require_last_push_approval,omitempty"`
	RequiredReviewThreadResolution bool     `yaml:"required_review_thread_resolution,omitempty"`
	AllowedMergeMethods            []string `yaml:"allowed_merge_methods,omitempty"` // merge|squash|rebase
}

// RequiredStatusChecksRule requires named checks to pass before merge.
// Strict additionally requires the branch to be up to date with its base.
type RequiredStatusChecksRule struct {
	Checks               []StatusCheck `yaml:"checks"`
	Strict               bool          `yaml:"strict,omitempty"`
	DoNotEnforceOnCreate *bool         `yaml:"do_not_enforce_on_create,omitempty"`
}

// StatusCheck names a required check. IntegrationID pins the check to a
// specific GitHub App, so an unrelated app cannot report the same context.
type StatusCheck struct {
	Context       string `yaml:"context"`
	IntegrationID *int64 `yaml:"integration_id,omitempty"`
}

// RequiredDeploymentsRule requires deployments to the named environments to
// succeed before merge.
type RequiredDeploymentsRule struct {
	Environments []string `yaml:"environments"`
}

// MergeQueueRule routes merges through GitHub's merge queue.
type MergeQueueRule struct {
	MergeMethod                  string `yaml:"merge_method,omitempty"`      // merge|squash|rebase
	GroupingStrategy             string `yaml:"grouping_strategy,omitempty"` // allgreen|headgreen
	CheckResponseTimeoutMinutes  int    `yaml:"check_response_timeout_minutes,omitempty"`
	MaxEntriesToBuild            int    `yaml:"max_entries_to_build,omitempty"`
	MaxEntriesToMerge            int    `yaml:"max_entries_to_merge,omitempty"`
	MinEntriesToMerge            int    `yaml:"min_entries_to_merge,omitempty"`
	MinEntriesToMergeWaitMinutes int    `yaml:"min_entries_to_merge_wait_minutes,omitempty"`
}

// PatternRule matches a commit message, email, branch name or tag name.
// Negate inverts the match, so the rule fails when the pattern *does* match.
type PatternRule struct {
	Name     string `yaml:"name,omitempty"`
	Operator string `yaml:"operator"` // starts_with|ends_with|contains|regex
	Pattern  string `yaml:"pattern"`
	Negate   *bool  `yaml:"negate,omitempty"`
}

// WorkflowsRule requires the named workflows to have run.
type WorkflowsRule struct {
	Workflows            []RuleWorkflow `yaml:"workflows"`
	DoNotEnforceOnCreate *bool          `yaml:"do_not_enforce_on_create,omitempty"`
}

// RuleWorkflow points at a workflow file, optionally pinned to a ref or SHA in
// a specific repository.
type RuleWorkflow struct {
	Path         string `yaml:"path"`
	Repository   string `yaml:"repository,omitempty"` // repo name in this org; resolved to an ID
	RepositoryID *int64 `yaml:"repository_id,omitempty"`
	Ref          string `yaml:"ref,omitempty"`
	SHA          string `yaml:"sha,omitempty"`
}

// CodeScanningRule blocks merges when a scanning tool reports alerts at or
// above the configured thresholds.
type CodeScanningRule struct {
	Tools []CodeScanningTool `yaml:"tools"`
}

// CodeScanningTool sets the thresholds for one scanning tool, e.g. CodeQL.
type CodeScanningTool struct {
	Tool                    string `yaml:"tool"`
	AlertsThreshold         string `yaml:"alerts_threshold,omitempty"`          // none|errors|errors_and_warnings|all
	SecurityAlertsThreshold string `yaml:"security_alerts_threshold,omitempty"` // none|critical|high_or_higher|medium_or_higher|all
}

// FileExtensionRestrictionRule blocks pushes touching the listed extensions.
type FileExtensionRestrictionRule struct {
	RestrictedFileExtensions []string `yaml:"restricted_file_extensions"`
}

// FilePathRestrictionRule blocks pushes touching the listed paths.
type FilePathRestrictionRule struct {
	RestrictedFilePaths []string `yaml:"restricted_file_paths"`
}

// ParseRulesets decodes a `rulesets:` value that arrived as an untyped YAML
// node — the shape repository entries in teams/*.yaml are held in. Round-tripping
// through YAML keeps RulesetConfig the single definition of the schema instead
// of duplicating the field walk here.
func ParseRulesets(v any) ([]RulesetConfig, error) {
	if v == nil {
		return nil, nil
	}
	if _, ok := v.([]any); !ok {
		return nil, fmt.Errorf("rulesets must be a list, got %T", v)
	}
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("re-encode rulesets: %w", err)
	}
	var out []RulesetConfig
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse rulesets: %w", err)
	}
	return out, nil
}

// Resolve expands the ruleset's preset and fills in defaults, returning a
// configuration with Target and Enforcement always set. The receiver is not
// modified.
func (r RulesetConfig) Resolve() (RulesetConfig, error) {
	if r.Preset != "" {
		preset, ok := RulesetPresets()[r.Preset]
		if !ok {
			return r, fmt.Errorf("ruleset %q: unknown preset %q (known: %s)", r.Name, r.Preset, strings.Join(PresetNames(), ", "))
		}
		if r.Target == "" {
			r.Target = preset.Target
		}
		if r.Enforcement == "" {
			r.Enforcement = preset.Enforcement
		}
		if r.Conditions == nil {
			r.Conditions = preset.Conditions
		}
		if len(r.BypassActors) == 0 {
			r.BypassActors = preset.BypassActors
		}
		r.Rules = mergeRules(r.Rules, preset.Rules)
	}
	if r.Target == "" {
		r.Target = RulesetTargetBranch
	}
	if r.Enforcement == "" {
		r.Enforcement = RulesetEnforcementActive
	}
	return r, nil
}

// mergeRules fills every nil field in dst from src. All RulesetRules fields are
// pointers precisely so this stays a uniform nil check rather than a per-field
// switch that new rules would have to be threaded through.
func mergeRules(dst, src RulesetRules) RulesetRules {
	dv := reflect.ValueOf(&dst).Elem()
	sv := reflect.ValueOf(src)
	for i := range dv.NumField() {
		if dv.Field(i).IsNil() && !sv.Field(i).IsNil() {
			dv.Field(i).Set(sv.Field(i))
		}
	}
	return dst
}

// IsEmpty reports whether no rule at all is enabled. A ruleset with no rules is
// accepted by GitHub but enforces nothing, so it is worth catching in validation.
func (rr RulesetRules) IsEmpty() bool {
	v := reflect.ValueOf(rr)
	for i := range v.NumField() {
		f := v.Field(i)
		if f.IsNil() {
			continue
		}
		// A *bool set to false is "explicitly off", not an enabled rule.
		if b, ok := f.Interface().(*bool); ok && !*b {
			continue
		}
		return false
	}
	return true
}

// BypassMode returns the actor's bypass mode, defaulting to "always".
func (b BypassActorConfig) BypassMode() string {
	if b.Mode == "" {
		return BypassModeAlways
	}
	return b.Mode
}

// NormalizedType returns the canonical, correctly-cased actor type for the
// GitHub API, or "" if the configured type is not recognized.
func (b BypassActorConfig) NormalizedType() string {
	for _, known := range []string{
		BypassActorTypeIntegration,
		BypassActorTypeOrganizationAdmin,
		BypassActorTypeRepositoryRole,
		BypassActorTypeTeam,
		BypassActorTypeDeployKey,
	} {
		if strings.EqualFold(b.Type, known) {
			return known
		}
	}
	return ""
}

// FixedActorID returns the actor ID GitHub mandates for actor types that have
// no per-instance identity, and false for types that need one resolved.
func (b BypassActorConfig) FixedActorID() (int64, bool) {
	if b.NormalizedType() == BypassActorTypeOrganizationAdmin {
		return orgAdminActorID, true
	}
	return 0, false
}
