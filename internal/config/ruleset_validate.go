package config

import (
	"fmt"
	"strconv"
	"strings"
)

var validRulesetTargets = map[string]bool{
	RulesetTargetBranch: true,
	RulesetTargetTag:    true,
	RulesetTargetPush:   true,
}

var validRulesetEnforcements = map[string]bool{
	RulesetEnforcementActive:   true,
	RulesetEnforcementEvaluate: true,
	RulesetEnforcementDisabled: true,
}

var validPatternOperators = map[string]bool{
	"starts_with": true,
	"ends_with":   true,
	"contains":    true,
	"regex":       true,
}

var validMergeMethods = map[string]bool{
	"merge":  true,
	"squash": true,
	"rebase": true,
}

var validGroupingStrategies = map[string]bool{
	"allgreen":  true,
	"headgreen": true,
}

var validAlertsThresholds = map[string]bool{
	"none":                true,
	"errors":              true,
	"errors_and_warnings": true,
	"all":                 true,
}

var validSecurityAlertsThresholds = map[string]bool{
	"none":             true,
	"critical":         true,
	"high_or_higher":   true,
	"medium_or_higher": true,
	"all":              true,
}

// RulesetScope distinguishes the two places a ruleset can be declared, because
// a few rules are only legal at one of them.
type RulesetScope string

const (
	// ScopeOrg is org.yaml `rulesets:` — applies across the organization.
	ScopeOrg RulesetScope = "org"
	// ScopeRepo is a `rulesets:` block on a repository in teams/*.yaml.
	ScopeRepo RulesetScope = "repo"
)

// ValidateRulesets checks a ruleset list for the mistakes GitHub would only
// report as an opaque 422 at apply time: unknown presets, invalid enumerations,
// duplicate names, and conditions used at the wrong scope.
//
// where labels the source ("org.yaml" or "team X, repo Y") in error messages.
func ValidateRulesets(scope RulesetScope, where string, rulesets []RulesetConfig) error {
	seen := map[string]bool{}
	for i, raw := range rulesets {
		if strings.TrimSpace(raw.Name) == "" {
			return fmt.Errorf("%s: rulesets[%d]: name must not be empty", where, i)
		}
		key := strings.ToLower(raw.Name)
		if seen[key] {
			return fmt.Errorf("%s: duplicate ruleset name %q", where, raw.Name)
		}
		seen[key] = true

		r, err := raw.Resolve()
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if err := validateResolvedRuleset(scope, where, r); err != nil {
			return err
		}
	}
	return nil
}

func validateResolvedRuleset(scope RulesetScope, where string, r RulesetConfig) error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%s: ruleset %q: "+format, append([]any{where, r.Name}, args...)...)
	}

	if !validRulesetTargets[r.Target] {
		return fail("invalid target %q (must be branch, tag or push)", r.Target)
	}
	if !validRulesetEnforcements[r.Enforcement] {
		return fail("invalid enforcement %q (must be active, evaluate or disabled)", r.Enforcement)
	}
	if r.Rules.IsEmpty() {
		return fail("no rules enabled; a ruleset with no rules enforces nothing")
	}

	if r.Conditions != nil {
		if scope == ScopeRepo && r.Conditions.RepositoryName != nil {
			return fail("repository_name conditions are only valid on organization rulesets")
		}
		if r.Target == RulesetTargetPush && r.Conditions.RefName != nil {
			return fail("ref_name conditions do not apply to a push ruleset")
		}
	}

	for i, actor := range r.BypassActors {
		if err := validateBypassActor(actor); err != nil {
			return fail("bypass_actors[%d]: %w", i, err)
		}
	}

	return validateRules(fail, r.Target, r.Rules)
}

func validateBypassActor(a BypassActorConfig) error {
	kind := a.NormalizedType()
	if kind == "" {
		return fmt.Errorf("invalid type %q (must be Integration, OrganizationAdmin, RepositoryRole, Team or DeployKey)", a.Type)
	}
	switch mode := a.BypassMode(); mode {
	case BypassModeAlways, BypassModePullRequest:
	default:
		return fmt.Errorf("invalid mode %q (must be always or pull_request)", mode)
	}

	switch kind {
	case BypassActorTypeTeam:
		if a.Team == "" && a.ActorID == 0 {
			return fmt.Errorf("a %s bypass actor needs a team slug or an actor_id", BypassActorTypeTeam)
		}
	case BypassActorTypeIntegration:
		if a.App == "" && a.ActorID == 0 {
			return fmt.Errorf("an %s bypass actor needs app: self, an app ID, or an actor_id", BypassActorTypeIntegration)
		}
		if a.App != "" && !strings.EqualFold(a.App, "self") {
			if _, err := strconv.ParseInt(a.App, 10, 64); err != nil {
				return fmt.Errorf("app must be %q or a numeric GitHub App ID, got %q", "self", a.App)
			}
		}
	case BypassActorTypeRepositoryRole:
		if a.ActorID == 0 {
			return fmt.Errorf("a %s bypass actor needs actor_id (the role ID reported by the repository-roles API)", BypassActorTypeRepositoryRole)
		}
	}
	return nil
}

// validateRules is a flat sweep over the rule list. Splitting it up would
// scatter the schema across helpers without making it easier to follow.
//
//nolint:gocyclo
func validateRules(fail func(string, ...any) error, target string, rules RulesetRules) error {
	branchOnly := map[string]bool{}
	if target != RulesetTargetBranch {
		branchOnly["pull_request"] = rules.PullRequest != nil
		branchOnly["required_status_checks"] = rules.RequiredStatusChecks != nil
		branchOnly["required_deployments"] = rules.RequiredDeployments != nil
		branchOnly["merge_queue"] = rules.MergeQueue != nil
		branchOnly["required_linear_history"] = isTrue(rules.RequiredLinearHistory)
	}
	for name, set := range branchOnly {
		if set {
			return fail("rule %q only applies to a branch ruleset, not target %q", name, target)
		}
	}

	pushRules := map[string]bool{
		"file_extension_restriction": rules.FileExtensionRestriction != nil,
		"file_path_restriction":      rules.FilePathRestriction != nil,
		"max_file_path_length":       rules.MaxFilePathLength != nil,
		"max_file_size":              rules.MaxFileSize != nil,
	}
	for name, set := range pushRules {
		if set && target != RulesetTargetPush {
			return fail("rule %q only applies to a push ruleset, not target %q", name, target)
		}
	}

	if pr := rules.PullRequest; pr != nil {
		if pr.RequiredApprovingReviewCount < 0 || pr.RequiredApprovingReviewCount > 10 {
			return fail("pull_request.required_approving_review_count must be 0-10, got %d", pr.RequiredApprovingReviewCount)
		}
		for _, m := range pr.AllowedMergeMethods {
			if !validMergeMethods[strings.ToLower(m)] {
				return fail("pull_request: invalid merge method %q (must be merge, squash or rebase)", m)
			}
		}
	}

	if sc := rules.RequiredStatusChecks; sc != nil {
		if len(sc.Checks) == 0 {
			return fail("required_status_checks: checks must not be empty")
		}
		for i, check := range sc.Checks {
			if strings.TrimSpace(check.Context) == "" {
				return fail("required_status_checks.checks[%d]: context must not be empty", i)
			}
		}
	}

	if rd := rules.RequiredDeployments; rd != nil && len(rd.Environments) == 0 {
		return fail("required_deployments: environments must not be empty")
	}

	if mq := rules.MergeQueue; mq != nil {
		if mq.MergeMethod != "" && !validMergeMethods[strings.ToLower(mq.MergeMethod)] {
			return fail("merge_queue: invalid merge_method %q (must be merge, squash or rebase)", mq.MergeMethod)
		}
		if mq.GroupingStrategy != "" && !validGroupingStrategies[strings.ToLower(mq.GroupingStrategy)] {
			return fail("merge_queue: invalid grouping_strategy %q (must be allgreen or headgreen)", mq.GroupingStrategy)
		}
	}

	patterns := map[string]*PatternRule{
		"commit_message_pattern":      rules.CommitMessagePattern,
		"commit_author_email_pattern": rules.CommitAuthorEmailPattern,
		"committer_email_pattern":     rules.CommitterEmailPattern,
		"branch_name_pattern":         rules.BranchNamePattern,
		"tag_name_pattern":            rules.TagNamePattern,
	}
	for name, p := range patterns {
		if p == nil {
			continue
		}
		if !validPatternOperators[strings.ToLower(p.Operator)] {
			return fail("%s: invalid operator %q (must be starts_with, ends_with, contains or regex)", name, p.Operator)
		}
		if p.Pattern == "" {
			return fail("%s: pattern must not be empty", name)
		}
	}
	if rules.BranchNamePattern != nil && target != RulesetTargetBranch {
		return fail("rule %q only applies to a branch ruleset, not target %q", "branch_name_pattern", target)
	}
	if rules.TagNamePattern != nil && target != RulesetTargetTag {
		return fail("rule %q only applies to a tag ruleset, not target %q", "tag_name_pattern", target)
	}

	if wf := rules.Workflows; wf != nil {
		if len(wf.Workflows) == 0 {
			return fail("workflows: workflows must not be empty")
		}
		for i, w := range wf.Workflows {
			if strings.TrimSpace(w.Path) == "" {
				return fail("workflows.workflows[%d]: path must not be empty", i)
			}
			if w.Repository == "" && w.RepositoryID == nil {
				return fail("workflows.workflows[%d]: repository or repository_id is required", i)
			}
		}
	}

	if cs := rules.CodeScanning; cs != nil {
		if len(cs.Tools) == 0 {
			return fail("code_scanning: tools must not be empty")
		}
		for i, t := range cs.Tools {
			if strings.TrimSpace(t.Tool) == "" {
				return fail("code_scanning.tools[%d]: tool must not be empty", i)
			}
			if t.AlertsThreshold != "" && !validAlertsThresholds[strings.ToLower(t.AlertsThreshold)] {
				return fail("code_scanning.tools[%d]: invalid alerts_threshold %q", i, t.AlertsThreshold)
			}
			if t.SecurityAlertsThreshold != "" && !validSecurityAlertsThresholds[strings.ToLower(t.SecurityAlertsThreshold)] {
				return fail("code_scanning.tools[%d]: invalid security_alerts_threshold %q", i, t.SecurityAlertsThreshold)
			}
		}
	}

	if fe := rules.FileExtensionRestriction; fe != nil && len(fe.RestrictedFileExtensions) == 0 {
		return fail("file_extension_restriction: restricted_file_extensions must not be empty")
	}
	if fp := rules.FilePathRestriction; fp != nil && len(fp.RestrictedFilePaths) == 0 {
		return fail("file_path_restriction: restricted_file_paths must not be empty")
	}
	if v := rules.MaxFilePathLength; v != nil && (*v < 1 || *v > 4096) {
		return fail("max_file_path_length must be 1-4096, got %d", *v)
	}
	if v := rules.MaxFileSize; v != nil && *v < 1 {
		return fail("max_file_size must be positive, got %d", *v)
	}

	return nil
}

func isTrue(b *bool) bool { return b != nil && *b }
