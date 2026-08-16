package sync

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
)

// refAll and refDefaultBranch are GitHub's magic ref selectors.
const (
	refSelectorDefaultBranch = "~DEFAULT_BRANCH"
)

// pushBlockingRule names a rule that stops gomgr committing straight to a
// branch, with the reason to report.
type pushBlockingRule struct {
	name   string
	reason string
}

// blockingRules returns the rules in a ruleset that would reject a direct push
// to a protected branch.
//
// The list is deliberately short. Most rules constrain *what* a commit may
// contain — a message pattern, a signature, linear history — and a direct push
// can satisfy them. These three constrain *how* the branch may be advanced at
// all:
//
//   - pull_request: the ref only moves through a merged pull request.
//   - required_status_checks: a check cannot have passed on a commit that does
//     not exist yet, so a push carrying that commit is refused.
//   - update: updates to the ref are restricted outright.
//
// required_signatures is absent on purpose: commits gomgr makes through the
// Contents API are signed by GitHub, so that rule is satisfied either way.
func blockingRules(rules config.RulesetRules) []pushBlockingRule {
	var out []pushBlockingRule
	if rules.PullRequest != nil {
		out = append(out, pushBlockingRule{
			name:   "pull_request",
			reason: "the branch only advances through a merged pull request",
		})
	}
	if rules.RequiredStatusChecks != nil {
		out = append(out, pushBlockingRule{
			name:   "required_status_checks",
			reason: "a required check cannot have passed on a commit that does not exist yet",
		})
	}
	if rules.Update != nil {
		out = append(out, pushBlockingRule{
			name:   "update",
			reason: "updates to the branch are restricted",
		})
	}
	return out
}

// fileRoute is how a file change should reach its branch.
type fileRoute struct {
	// UsePullRequest is true when a direct commit would be rejected.
	UsePullRequest bool
	// Reason explains the decision, for the plan and the audit log.
	Reason string
	// HeadBranch is the branch the pull request is opened from.
	HeadBranch string
}

// routeDecider works out, per repository and branch, whether gomgr has to go
// through a pull request.
type routeDecider struct {
	strategy   string
	appID      int64
	headBranch string
	// orgRulesets are the organization-wide rulesets from org.yaml.
	orgRulesets []config.RulesetConfig
	// repoRulesets maps a lowercase repository name to its own rulesets.
	repoRulesets map[string][]config.RulesetConfig
	// defaultBranch maps a lowercase repository name to its default branch, so
	// ~DEFAULT_BRANCH can be resolved.
	defaultBranch map[string]string
}

// newRouteDecider builds a decider from the configuration and prefetched state.
func newRouteDecider(cfg *config.Root, st *State, bySettings map[string]repoSettings) *routeDecider {
	d := &routeDecider{
		strategy:      cfg.App.FileChanges.ResolvedStrategy(),
		appID:         cfg.App.ResolvedAppID(),
		headBranch:    cfg.App.FileChanges.ResolvedBranch(),
		orgRulesets:   cfg.Org.Rulesets,
		repoRulesets:  map[string][]config.RulesetConfig{},
		defaultBranch: map[string]string{},
	}
	for repo, settings := range bySettings {
		if len(settings.rulesets) > 0 {
			d.repoRulesets[repo] = settings.rulesets
		}
	}
	for _, r := range st.ActualRepos {
		d.defaultBranch[strings.ToLower(r.GetName())] = r.GetDefaultBranch()
	}
	return d
}

// route decides how a file change for repo/branch should be applied.
func (d *routeDecider) route(repo, branch string) (fileRoute, error) {
	switch d.strategy {
	case config.FileStrategyDirect:
		return fileRoute{Reason: "file_changes.strategy is direct"}, nil
	case config.FileStrategyPullRequest:
		return d.pullRequestRoute("file_changes.strategy is pull_request"), nil
	}

	repoKey := strings.ToLower(repo)
	candidates := append([]config.RulesetConfig{}, d.orgRulesets...)
	candidates = append(candidates, d.repoRulesets[repoKey]...)

	for _, raw := range candidates {
		spec, err := raw.Resolve()
		if err != nil {
			return fileRoute{}, err
		}
		if spec.Enforcement != config.RulesetEnforcementActive {
			continue // evaluate and disabled do not reject anything
		}
		if spec.Target != config.RulesetTargetBranch {
			continue
		}
		if !d.rulesetCovers(spec, repoKey, branch) {
			continue
		}
		blocking := blockingRules(spec.Rules)
		if len(blocking) == 0 {
			continue
		}
		if bypassesDirectPush(spec.BypassActors, d.appID) {
			continue // gomgr is exempt from this one
		}
		return d.pullRequestRoute(fmt.Sprintf("ruleset %q rule %q applies to %s: %s",
			spec.Name, blocking[0].name, branch, blocking[0].reason)), nil
	}

	return fileRoute{Reason: "no declared ruleset blocks a direct push"}, nil
}

// pullRequestRoute builds a pull-request route carrying the configured mechanics.
func (d *routeDecider) pullRequestRoute(reason string) fileRoute {
	return fileRoute{
		UsePullRequest: true,
		Reason:         reason,
		HeadBranch:     d.headBranch,
	}
}

// rulesetCovers reports whether a ruleset applies to this repository and branch.
func (d *routeDecider) rulesetCovers(spec config.RulesetConfig, repoKey, branch string) bool {
	include, exclude := []string{refSelectorAll}, []string(nil)
	var repoInclude, repoExclude []string
	repoInclude = []string{refSelectorAll}

	if spec.Conditions != nil {
		if rn := spec.Conditions.RefName; rn != nil {
			if len(rn.Include) > 0 {
				include = rn.Include
			}
			exclude = rn.Exclude
		}
		if rn := spec.Conditions.RepositoryName; rn != nil {
			if len(rn.Include) > 0 {
				repoInclude = rn.Include
			}
			repoExclude = rn.Exclude
		}
	}

	if !matchesSelector(repoKey, repoInclude, repoExclude, "") {
		return false
	}
	return matchesSelector(branch, include, exclude, d.defaultBranch[repoKey])
}

// matchesSelector applies GitHub's include/exclude semantics: a value is
// covered when it matches an include and no exclude. defaultBranch, when
// non-empty, is what ~DEFAULT_BRANCH resolves to.
func matchesSelector(value string, include, exclude []string, defaultBranch string) bool {
	for _, pattern := range exclude {
		if selectorMatches(pattern, value, defaultBranch) {
			return false
		}
	}
	for _, pattern := range include {
		if selectorMatches(pattern, value, defaultBranch) {
			return true
		}
	}
	return false
}

func selectorMatches(pattern, value, defaultBranch string) bool {
	switch pattern {
	case refSelectorAll:
		return true
	case refSelectorDefaultBranch:
		// Only meaningful for refs; an empty defaultBranch means the caller is
		// matching repository names, where this selector does not apply.
		return defaultBranch != "" && strings.EqualFold(defaultBranch, value)
	}
	// GitHub writes ref conditions as full refs; accept either spelling.
	candidate := value
	if strings.HasPrefix(pattern, "refs/heads/") {
		candidate = "refs/heads/" + value
	}
	if ok, err := path.Match(pattern, candidate); err == nil && ok {
		return true
	}
	return strings.EqualFold(pattern, candidate)
}

// bypassesDirectPush reports whether gomgr itself can push past the ruleset.
//
// Only an "always" bypass helps. An actor limited to bypass_mode
// "pull_request" is, by definition, being told to use one.
func bypassesDirectPush(actors []config.BypassActorConfig, appID int64) bool {
	for _, a := range actors {
		if a.BypassMode() != config.BypassModeAlways {
			continue
		}
		switch a.NormalizedType() {
		case config.BypassActorTypeIntegration:
			if strings.EqualFold(a.App, "self") {
				return true
			}
			if appID != 0 && a.App == fmt.Sprint(appID) {
				return true
			}
			if appID != 0 && a.ActorID == appID {
				return true
			}
		case config.BypassActorTypeOrganizationAdmin, config.BypassActorTypeEnterpriseOwner:
			// gomgr authenticates as an app, not as a member, so an admin
			// bypass does not cover it.
			continue
		}
	}
	return false
}

// ExplainFileRoutes reports the route gomgr would take for each managed
// repository's file writes, and why. It exists so the decision can be inspected
// without applying anything.
func ExplainFileRoutes(ctx context.Context, c *gh.Client, cfg *config.Root) []string {
	st := &State{Org: cfg.App.Org}
	if err := prefetchState(ctx, c, st); err != nil {
		return []string{fmt.Sprintf("prefetch: %v", err)}
	}
	all, _, err := collectRepoSettings(cfg, cfg.App.Org)
	if err != nil {
		return []string{fmt.Sprintf("collect: %v", err)}
	}
	resolved, err := resolveAllTemplates(all, cfg.App.Org)
	if err != nil {
		return []string{fmt.Sprintf("resolve: %v", err)}
	}
	d := newRouteDecider(cfg, st, resolved)

	specs := materializeFileSpecs(cfg.App)
	var out []string
	for _, repo := range sortedKeys(resolved) {
		for _, spec := range specs {
			branch := spec.Branch
			if branch == "" {
				branch = defaultFileBranch
			}
			r, err := d.route(repo, branch)
			if err != nil {
				out = append(out, fmt.Sprintf("%-26s ERROR %v", repo, err))
				continue
			}
			route := "direct"
			if r.UsePullRequest {
				route = "pull request"
			}
			out = append(out, fmt.Sprintf("%-26s %-12s %s", repo+"@"+branch, route, r.Reason))
		}
	}
	return out
}

// fileRouter decides the route for one repository and branch.
type fileRouter func(repo, branch string) (fileRoute, error)

// decorate adds the fields a pull-request-routed change needs at apply time.
func (r fileRoute) decorate(details map[string]any) {
	details["route_reason"] = r.Reason
	details["head_branch"] = r.HeadBranch
}
