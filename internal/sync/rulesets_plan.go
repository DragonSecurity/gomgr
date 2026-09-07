package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// planOrgRulesets reconciles org.yaml's `rulesets:` against the organization's
// rulesets, and reports (or removes) any that the config does not declare.
func planOrgRulesets(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	wantCleanup := cfg.App.DeleteUnmanagedRulesets || cfg.App.DryWarnings.WarnUnmanagedRulesets
	if len(cfg.Org.Rulesets) == 0 && !wantCleanup {
		return nil, nil, nil
	}

	existing, err := fetchOrgRulesets(ctx, c, org)
	if err != nil {
		// Enabling warn_unmanaged_rulesets should not be able to break a sync
		// that declares no rulesets. When the config does declare them, the
		// failure is the user's problem to see.
		if len(cfg.Org.Rulesets) == 0 && (isNotFound(err) || isRulesetsUnavailable(err)) {
			return nil, []string{fmt.Sprintf("Cannot read rulesets for %s (%v); skipping ruleset checks", org, err)}, nil
		}
		return nil, nil, err
	}
	st.CurrentRulesets += len(existing)
	st.DesiredRulesets += len(cfg.Org.Rulesets)

	lookup := newPlanLookup(c, st)
	changes, setWarnings, err := planRulesetSet(ctx, rulesetScopeArgs{
		scope:    scopeOrgRuleset,
		org:      org,
		orgLevel: true,
	}, cfg.Org.Rulesets, existing, lookup)
	if err != nil {
		return nil, nil, err
	}
	out = append(out, changes...)
	warnings = append(warnings, setWarnings...)

	deletes, unmanaged := planRulesetCleanup(rulesetScopeArgs{
		scope:    scopeOrgRuleset,
		org:      org,
		orgLevel: true,
	}, cfg.Org.Rulesets, existing, cfg.App.DeleteUnmanagedRulesets)
	out = append(out, deletes...)
	if cfg.App.DryWarnings.WarnUnmanagedRulesets && len(unmanaged) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d unmanaged organization rulesets: %v", len(unmanaged), unmanaged))
	}

	warnings = append(warnings, warnSelfLockout(cfg, cfg.Org.Rulesets, "organization")...)

	return out, warnings, nil
}

// planRepoRulesets reconciles the `rulesets:` block on each managed repository.
// Repositories that this run is about to create have nothing to compare
// against, so their rulesets are planned as creates.
func planRepoRulesets(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	allSettings, _, _, err := collectRepoSettings(cfg, org)
	if err != nil {
		return nil, nil, err
	}
	bySettings, err := resolveAllTemplates(allSettings, org)
	if err != nil {
		return nil, nil, err
	}

	existingRepos := map[string]bool{}
	for _, r := range st.ActualRepos {
		existingRepos[strings.ToLower(r.GetName())] = true
	}

	lookup := newPlanLookup(c, st)

	for _, repo := range sortedKeys(bySettings) {
		settings := bySettings[repo]
		wantCleanup := cfg.App.DeleteUnmanagedRulesets || cfg.App.DryWarnings.WarnUnmanagedRulesets
		if len(settings.rulesets) == 0 && !wantCleanup {
			continue
		}
		st.DesiredRulesets += len(settings.rulesets)

		// A repository this run is creating cannot be queried yet; everything
		// declared for it is a create.
		var existing []*github.RepositoryRuleset
		if existingRepos[repo] {
			fetched, err := fetchRepoRulesets(ctx, c, org, repo)
			if err != nil {
				return nil, nil, err
			}
			existing = fetched
			st.CurrentRulesets += len(existing)
		}

		args := rulesetScopeArgs{scope: scopeRepoRuleset, org: org, repo: repo}
		changes, setWarnings, err := planRulesetSet(ctx, args, settings.rulesets, existing, lookup)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
		warnings = append(warnings, setWarnings...)

		deletes, unmanaged := planRulesetCleanup(args, settings.rulesets, existing, cfg.App.DeleteUnmanagedRulesets)
		out = append(out, deletes...)
		if cfg.App.DryWarnings.WarnUnmanagedRulesets && len(unmanaged) > 0 {
			warnings = append(warnings, fmt.Sprintf("Found %d unmanaged rulesets on %s/%s: %v", len(unmanaged), org, repo, unmanaged))
		}
		warnings = append(warnings, warnSelfLockout(cfg, settings.rulesets, org+"/"+repo)...)
	}

	return out, warnings, nil
}

// rulesetScopeArgs bundles the details that distinguish an org ruleset from a
// repository one, so the planning below can be written once.
type rulesetScopeArgs struct {
	scope    string
	org      string
	repo     string
	orgLevel bool
}

func (a rulesetScopeArgs) target(name string) string {
	if a.repo == "" {
		return name
	}
	return a.repo + "/" + name
}

// planRulesetSet emits the create and update changes needed to bring existing
// in line with desired. Rulesets are matched by name, case-insensitively.
func planRulesetSet(ctx context.Context, args rulesetScopeArgs, desired []config.RulesetConfig, existing []*github.RepositoryRuleset, lookup *refLookup) ([]util.Change, []string, error) {
	byName := map[string]*github.RepositoryRuleset{}
	for _, rs := range existing {
		byName[strings.ToLower(rs.Name)] = rs
	}

	var warnings []string
	var out []util.Change
	for _, raw := range desired {
		spec, err := raw.Resolve()
		if err != nil {
			return nil, nil, err
		}

		current, exists := byName[strings.ToLower(spec.Name)]

		// A declared ruleset whose live counterpart carries a rule gomgr cannot
		// express is the dangerous case: apply replaces the whole ruleset from
		// configuration, so that rule is deleted on this run. The import refuses
		// to adopt such a ruleset, but a hand-written declaration, or one that
		// grew a rule in the web interface after being adopted, arrives here.
		if exists {
			if unmodeled := unmodeledRuleTypes(current.Rules); len(unmodeled) > 0 {
				warnings = append(warnings, fmt.Sprintf(
					"Ruleset %q on %s carries %s that gomgr cannot express; applying this configuration would DELETE %s. "+
						"Remove the ruleset from your configuration to leave it alone.",
					spec.Name, args.target(spec.Name), plural("rule type", unmodeled), pronounFor(unmodeled)))
			}
		}

		action := util.ActionCreate
		var id int64
		if exists {
			id = current.GetID()
			action = util.ActionUpdate

			// A reference the plan cannot resolve yet — most often a team this
			// run is about to create — means the comparison cannot be trusted,
			// so the change is planned and apply resolves it for real.
			built, buildErr := buildRuleset(ctx, spec, args.orgLevel, args.repo, lookup)
			if buildErr == nil {
				same, err := rulesetMatches(current, built)
				if err != nil {
					return nil, nil, fmt.Errorf("compare ruleset %q: %w", spec.Name, err)
				}
				if same {
					continue
				}
			}
		}

		out = append(out, util.Change{
			Scope:  args.scope,
			Target: args.target(spec.Name),
			Action: action,
			Details: rulesetChange{
				Org:  args.org,
				Repo: args.repo,
				ID:   id,
				Name: spec.Name,
				Spec: spec,
			},
		})
	}
	return out, warnings, nil
}

// planRulesetCleanup finds rulesets on GitHub that the config does not declare.
// Rulesets inherited from the organization or an enterprise are skipped: they
// are not this scope's to delete, and the API refuses anyway.
func planRulesetCleanup(args rulesetScopeArgs, desired []config.RulesetConfig, existing []*github.RepositoryRuleset, deleteThem bool) ([]util.Change, []string) {
	declared := map[string]bool{}
	for _, r := range desired {
		declared[strings.ToLower(r.Name)] = true
	}

	var out []util.Change
	var unmanaged []string
	for _, rs := range existing {
		if declared[strings.ToLower(rs.Name)] {
			continue
		}
		if !ownedAtScope(rs, args.orgLevel) {
			continue
		}
		unmanaged = append(unmanaged, rs.Name)
		if !deleteThem {
			continue
		}
		out = append(out, util.Change{
			Scope:  args.scope,
			Target: args.target(rs.Name),
			Action: util.ActionDelete,
			Details: rulesetChange{
				Org:  args.org,
				Repo: args.repo,
				ID:   rs.GetID(),
				Name: rs.Name,
			},
		})
	}
	return out, unmanaged
}

// ownedAtScope reports whether a ruleset was defined at the scope being synced,
// rather than inherited from above it.
func ownedAtScope(rs *github.RepositoryRuleset, orgLevel bool) bool {
	want := github.RulesetSourceTypeRepository
	if orgLevel {
		want = github.RulesetSourceTypeOrganization
	}
	// The list endpoints omit source_type on rulesets defined at the scope being
	// listed in some responses; treat "unset" as owned so a stale ruleset is
	// still reported rather than silently skipped.
	if rs.SourceType == nil {
		return true
	}
	return *rs.SourceType == want
}

// warnSelfLockout flags the configuration that locks gomgr out of its own file
// sync: gomgr commits templated files straight to the default branch, so a
// pull-request or signature rule covering that branch rejects its pushes unless
// the app is listed as a bypass actor.
func warnSelfLockout(cfg *config.Root, rulesets []config.RulesetConfig, where string) []string {
	if len(materializeFileSpecs(cfg.App)) == 0 {
		return nil
	}

	var warnings []string
	for _, raw := range rulesets {
		spec, err := raw.Resolve()
		if err != nil || spec.Enforcement != config.RulesetEnforcementActive {
			continue
		}
		if spec.Target != config.RulesetTargetBranch {
			continue
		}
		blocking := spec.Rules.PullRequest != nil || isEnabled(spec.Rules.RequiredSignatures)
		if !blocking || hasAppBypass(spec.BypassActors) {
			continue
		}
		warnings = append(warnings, fmt.Sprintf(
			"Ruleset %q on %s requires a pull request or signed commits on the branches gomgr writes files to; "+
				"add a bypass actor (type: Integration, app: self) or gomgr's file sync will be rejected",
			spec.Name, where))
	}
	return warnings
}

// hasAppBypass reports whether an actor that can carry gomgr's own pushes is
// exempt from the ruleset.
func hasAppBypass(actors []config.BypassActorConfig) bool {
	for _, a := range actors {
		switch a.NormalizedType() {
		case config.BypassActorTypeIntegration, config.BypassActorTypeOrganizationAdmin:
			if a.BypassMode() == config.BypassModeAlways {
				return true
			}
		}
	}
	return false
}

// fetchOrgRulesets lists the organization's rulesets and expands each one. The
// list endpoint returns only a summary — no rules, no conditions — so every
// ruleset has to be fetched individually before it can be compared.
func fetchOrgRulesets(ctx context.Context, c *gh.Client, org string) ([]*github.RepositoryRuleset, error) {
	var summaries []*github.RepositoryRuleset
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		page, resp, err := c.REST.Organizations.ListAllRepositoryRulesets(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, page...)
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list rulesets for org %s: %w", org, err)
	}

	out := make([]*github.RepositoryRuleset, 0, len(summaries))
	for _, s := range summaries {
		full, _, err := c.REST.Organizations.GetRepositoryRuleset(ctx, org, s.GetID())
		if err != nil {
			return nil, fmt.Errorf("get org ruleset %q (ID %d): %w", s.Name, s.GetID(), err)
		}
		out = append(out, full)
	}
	return out, nil
}

// fetchRepoRulesets lists a repository's own rulesets, excluding those
// inherited from the organization, and expands each one.
func fetchRepoRulesets(ctx context.Context, c *gh.Client, org, repo string) ([]*github.RepositoryRuleset, error) {
	var summaries []*github.RepositoryRuleset
	listOpts := &github.RepositoryListRulesetsOptions{IncludesParents: github.Ptr(false)}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		listOpts.ListOptions = *opts
		page, resp, err := c.REST.Repositories.GetAllRulesets(ctx, org, repo, listOpts)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, page...)
		return resp, nil
	}); err != nil {
		if isNotFound(err) || isRulesetsUnavailable(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list rulesets for %s/%s: %w", org, repo, err)
	}

	out := make([]*github.RepositoryRuleset, 0, len(summaries))
	for _, s := range summaries {
		full, _, err := c.REST.Repositories.GetRuleset(ctx, org, repo, s.GetID(), false)
		if err != nil {
			// A ruleset that lists at the repository but 404s when fetched
			// without parents was inherited from the organization or the
			// enterprise. It is not this scope's to read, let alone manage.
			if isNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("get ruleset %q on %s/%s (ID %d): %w", s.Name, org, repo, s.GetID(), err)
		}
		out = append(out, full)
	}
	return out, nil
}

// isRulesetsUnavailable reports whether the rulesets API is simply not offered
// for this repository — an empty or archived repository answers 403 or 409
// rather than returning an empty list.
func isRulesetsUnavailable(err error) bool {
	var ghErr *github.ErrorResponse
	if !errors.As(err, &ghErr) || ghErr.Response == nil {
		return false
	}
	switch ghErr.Response.StatusCode {
	case http.StatusForbidden, http.StatusConflict:
		return true
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
