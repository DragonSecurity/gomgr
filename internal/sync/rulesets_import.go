package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
)

// ImportedRuleset is a ruleset that exists on GitHub, rendered back into the
// shape gomgr's configuration uses.
type ImportedRuleset struct {
	// Repo is the repository the ruleset belongs to, or "" for an
	// organization-level one.
	Repo string
	// Spec is the ruleset as configuration, ready to be written to YAML.
	Spec config.RulesetConfig
}

// SkippedRuleset is a ruleset the import could not turn into configuration.
//
// GitHub's ruleset schema is wider than gomgr's, and grows: a rule or bypass
// actor type that postdates this build has to land somewhere. One such ruleset
// must not abort a scan of fifty, so it is reported and passed over — the
// ruleset itself is left exactly as it is on GitHub.
type SkippedRuleset struct {
	Repo   string // "" for an organization ruleset
	Name   string
	Reason string
}

// ImportOptions tunes what ImportRulesets looks at.
type ImportOptions struct {
	// Only restricts the repository scan to names matching these path.Match
	// globs. Empty scans every repository in the organization.
	Only []string
}

// ImportResult is what a scan found.
type ImportResult struct {
	// Org holds organization rulesets the configuration does not declare.
	Org []ImportedRuleset
	// Repos maps a repository name to the rulesets on it that the
	// configuration does not declare, sorted by repository name.
	Repos map[string][]ImportedRuleset
	// Unmanaged lists repositories that hold adoptable rulesets but appear in
	// no team file. There is nowhere in the configuration to write their
	// rulesets until the repository itself is adopted into a team.
	Unmanaged []string
	// Skipped lists rulesets that exist but could not be represented as
	// configuration, with the reason.
	Skipped []SkippedRuleset
	// AlreadyDeclared counts rulesets skipped because the configuration
	// already declares a ruleset of that name at that scope.
	AlreadyDeclared int
	// Scanned counts the repositories inspected.
	Scanned int
}

// RepoNames returns the repositories in Repos, sorted.
func (r *ImportResult) RepoNames() []string { return sortedKeys(r.Repos) }

// Total returns how many rulesets are available to adopt.
func (r *ImportResult) Total() int {
	n := len(r.Org)
	for _, rulesets := range r.Repos {
		n += len(rulesets)
	}
	return n
}

// ImportRulesets discovers the rulesets that already exist on GitHub — the ones
// somebody created in the web UI, or that predate gomgr managing the org — and
// converts them back into configuration.
//
// It only reports rulesets the configuration does not already declare. A
// declared ruleset is gomgr's to define, and re-importing it would overwrite the
// YAML with whatever the live state happens to be, which is backwards.
//
// Unlike planning, this scans every repository rather than only the managed
// ones: a ruleset on a repository nobody has adopted yet is exactly the kind of
// thing this command exists to surface.
func ImportRulesets(ctx context.Context, c *gh.Client, cfg *config.Root, opts ImportOptions) (*ImportResult, error) {
	org := cfg.App.Org
	result := &ImportResult{Repos: map[string][]ImportedRuleset{}}

	lookup, err := newImportLookup(ctx, c, cfg)
	if err != nil {
		return nil, err
	}

	declaredOrg := declaredNames(cfg.Org.Rulesets)
	orgRulesets, err := fetchOrgRulesets(ctx, c, org)
	if err != nil {
		return nil, err
	}
	for _, rs := range orgRulesets {
		if !ownedAtScope(rs, true) {
			continue
		}
		if declaredOrg[strings.ToLower(rs.Name)] {
			result.AlreadyDeclared++
			continue
		}
		spec := rulesetToConfig(rs, true, lookup)
		if err := config.ValidateRulesets(config.ScopeOrg, "imported", []config.RulesetConfig{spec}); err != nil {
			result.Skipped = append(result.Skipped, SkippedRuleset{Name: rs.Name, Reason: reasonOf(err)})
			continue
		}
		result.Org = append(result.Org, ImportedRuleset{Spec: spec})
	}

	declaredByRepo, managed, err := declaredRepoRulesets(cfg, org)
	if err != nil {
		return nil, err
	}

	for _, repo := range lookup.repoNames {
		if !matchesAnyGlob(repo, opts.Only) {
			continue
		}
		result.Scanned++

		rulesets, err := fetchRepoRulesets(ctx, c, org, repo)
		if err != nil {
			return nil, err
		}

		var adoptable []ImportedRuleset
		for _, rs := range rulesets {
			if !ownedAtScope(rs, false) {
				continue
			}
			if declaredByRepo[repo][strings.ToLower(rs.Name)] {
				result.AlreadyDeclared++
				continue
			}
			spec := rulesetToConfig(rs, false, lookup)
			if err := config.ValidateRulesets(config.ScopeRepo, "imported", []config.RulesetConfig{spec}); err != nil {
				result.Skipped = append(result.Skipped, SkippedRuleset{Repo: repo, Name: rs.Name, Reason: reasonOf(err)})
				continue
			}
			adoptable = append(adoptable, ImportedRuleset{Repo: repo, Spec: spec})
		}
		if len(adoptable) == 0 {
			continue
		}
		if !managed[repo] {
			result.Unmanaged = append(result.Unmanaged, repo)
			continue
		}
		result.Repos[repo] = adoptable
	}
	sort.Strings(result.Unmanaged)

	return result, nil
}

// reasonOf strips the scope and ruleset-name prefixes ValidateRulesets adds,
// leaving the part that says what is actually wrong. The caller already knows
// which ruleset it was asking about.
func reasonOf(err error) string {
	msg := err.Error()
	if _, rest, found := strings.Cut(msg, `ruleset "`); found {
		if _, detail, ok := strings.Cut(rest, `": `); ok {
			return detail
		}
	}
	return msg
}

// matchesAnyGlob reports whether name matches one of the globs, or whether
// there are no globs to match against.
func matchesAnyGlob(name string, globs []string) bool {
	if len(globs) == 0 {
		return true
	}
	for _, g := range globs {
		if ok, err := path.Match(g, name); err == nil && ok {
			return true
		}
	}
	return false
}

func declaredNames(rulesets []config.RulesetConfig) map[string]bool {
	out := map[string]bool{}
	for _, r := range rulesets {
		out[strings.ToLower(r.Name)] = true
	}
	return out
}

// declaredRepoRulesets returns, per repository, the ruleset names the
// configuration declares, along with the set of repositories the configuration
// knows about at all.
func declaredRepoRulesets(cfg *config.Root, org string) (map[string]map[string]bool, map[string]bool, error) {
	allSettings, managed, _, err := collectRepoSettings(cfg, org)
	if err != nil {
		return nil, nil, err
	}
	resolved, err := resolveAllTemplates(allSettings, org)
	if err != nil {
		return nil, nil, err
	}
	byRepo := make(map[string]map[string]bool, len(resolved))
	for repo, settings := range resolved {
		byRepo[repo] = declaredNames(settings.rulesets)
	}
	return byRepo, managed, nil
}

// importLookup reverses the ID resolution buildRuleset performs, turning the
// numeric IDs GitHub reports back into the names a configuration file uses.
type importLookup struct {
	teamSlugByID map[int64]string
	repoNameByID map[int64]string
	repoNames    []string
	appID        int64
}

func newImportLookup(ctx context.Context, c *gh.Client, cfg *config.Root) (*importLookup, error) {
	l := &importLookup{
		teamSlugByID: map[int64]string{},
		repoNameByID: map[int64]string{},
		appID:        cfg.App.ResolvedAppID(),
	}
	org := cfg.App.Org

	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		teams, resp, err := c.REST.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		for _, t := range teams {
			l.teamSlugByID[t.GetID()] = t.GetSlug()
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	repoOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
		Type:        "all",
	}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		repoOpt.ListOptions = *opts
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, repoOpt)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			name := strings.ToLower(r.GetName())
			l.repoNameByID[r.GetID()] = name
			l.repoNames = append(l.repoNames, name)
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
	}
	sort.Strings(l.repoNames)

	return l, nil
}

// rulesetToConfig converts a live ruleset into configuration, collapsing it to
// a preset reference when one describes it exactly.
func rulesetToConfig(rs *github.RepositoryRuleset, orgLevel bool, l *importLookup) config.RulesetConfig {
	spec := config.RulesetConfig{
		Name:         rs.Name,
		Target:       string(rulesetTarget(rs)),
		Enforcement:  string(rs.Enforcement),
		Conditions:   conditionsToConfig(rs.Conditions, orgLevel),
		BypassActors: bypassActorsToConfig(rs.BypassActors, l),
		Rules:        rulesToConfig(rs.Rules, l),
	}
	if spec.Target == "" {
		spec.Target = config.RulesetTargetBranch
	}
	if spec.Enforcement == "" {
		spec.Enforcement = config.RulesetEnforcementActive
	}

	// A preset that describes these exact rules replaces them, so the emitted
	// YAML reads as an intent ("branch-protection") rather than a transcript.
	// Conditions and bypass actors stay explicit either way: those are the
	// parts a preset deliberately does not decide.
	if preset := detectPreset(spec); preset != "" {
		spec.Preset = preset
		spec.Rules = config.RulesetRules{}
		spec.Target = ""
	}
	return spec
}

// detectPreset returns the name of the built-in preset whose target and rules
// match spec exactly, or "" if none does.
func detectPreset(spec config.RulesetConfig) string {
	want, err := json.Marshal(spec.Rules)
	if err != nil {
		return ""
	}
	presets := config.RulesetPresets()
	for _, name := range config.PresetNames() {
		preset := presets[name]
		if preset.Target != spec.Target {
			continue
		}
		got, err := json.Marshal(preset.Rules)
		if err != nil {
			continue
		}
		if string(got) == string(want) {
			return name
		}
	}
	return ""
}

// conditionsToConfig converts the targeting back, dropping the "~ALL" defaults
// that buildConditions would have supplied anyway.
func conditionsToConfig(conds *github.RepositoryRulesetConditions, orgLevel bool) *config.RulesetConditions {
	if conds == nil {
		return nil
	}
	out := &config.RulesetConditions{}
	any := false

	if rn := conds.RefName; rn != nil && !isAllSelector(rn.Include, rn.Exclude) {
		out.RefName = &config.RefNameCondition{Include: rn.Include, Exclude: rn.Exclude}
		any = true
	}
	if orgLevel {
		if rn := conds.RepositoryName; rn != nil {
			isDefault := isAllSelector(rn.Include, rn.Exclude) && rn.Protected == nil
			if !isDefault {
				out.RepositoryName = &config.RepositoryNameCondition{
					Include:   rn.Include,
					Exclude:   rn.Exclude,
					Protected: rn.Protected,
				}
				any = true
			}
		}
	}

	if !any {
		return nil
	}
	return out
}

// isAllSelector reports whether a condition selects everything, which is what
// gomgr emits when a config says nothing.
func isAllSelector(include, exclude []string) bool {
	return len(exclude) == 0 && len(include) == 1 && include[0] == refSelectorAll
}

func bypassActorsToConfig(actors []*github.BypassActor, l *importLookup) []config.BypassActorConfig {
	if len(actors) == 0 {
		return nil
	}
	out := make([]config.BypassActorConfig, 0, len(actors))
	for _, a := range actors {
		if a == nil {
			continue
		}
		var kind string
		if a.ActorType != nil {
			kind = string(*a.ActorType)
		}
		entry := config.BypassActorConfig{Type: kind}
		if a.BypassMode != nil && string(*a.BypassMode) != config.BypassModeAlways {
			entry.Mode = string(*a.BypassMode)
		}

		id := a.GetActorID()
		switch kind {
		case config.BypassActorTypeTeam:
			// A slug survives a team being recreated; the numeric ID does not.
			if slug, ok := l.teamSlugByID[id]; ok {
				entry.Team = slug
			} else {
				entry.ActorID = id
			}
		case config.BypassActorTypeIntegration:
			if id != 0 && id == l.appID {
				entry.App = "self"
			} else {
				entry.App = strconv.FormatInt(id, 10)
			}
		case config.BypassActorTypeOrganizationAdmin, config.BypassActorTypeDeployKey:
			// GitHub fixes the ID for these; recording it adds nothing.
		default:
			entry.ActorID = id
		}
		out = append(out, entry)
	}
	return out
}

// list, which is easier to check against its counterpart in one piece.
//
//nolint:gocyclo // The mirror image of buildRules: a flat walk over the rule
func rulesToConfig(rules *github.RepositoryRulesetRules, l *importLookup) config.RulesetRules {
	out := config.RulesetRules{}
	if rules == nil {
		return out
	}
	enabled := func() *bool { b := true; return &b }

	if rules.Creation != nil {
		out.Creation = enabled()
	}
	if rules.Deletion != nil {
		out.Deletion = enabled()
	}
	if rules.RequiredLinearHistory != nil {
		out.RequiredLinearHistory = enabled()
	}
	if rules.RequiredSignatures != nil {
		out.RequiredSignatures = enabled()
	}
	if rules.NonFastForward != nil {
		out.NonFastForward = enabled()
	}
	if rules.Update != nil {
		out.Update = &config.UpdateRule{AllowsFetchAndMerge: rules.Update.UpdateAllowsFetchAndMerge}
	}

	if pr := rules.PullRequest; pr != nil {
		entry := &config.PullRequestRule{
			RequiredApprovingReviewCount:   pr.RequiredApprovingReviewCount,
			DismissStaleReviewsOnPush:      pr.DismissStaleReviewsOnPush,
			RequireCodeOwnerReview:         pr.RequireCodeOwnerReview,
			RequireLastPushApproval:        pr.RequireLastPushApproval,
			RequiredReviewThreadResolution: pr.RequiredReviewThreadResolution,
		}
		// GitHub reports every merge method when none is restricted. Recording
		// that as a restriction would be a lie, and would stop the ruleset from
		// ever collapsing to a preset.
		if !isAllMergeMethods(pr.AllowedMergeMethods) {
			for _, m := range pr.AllowedMergeMethods {
				entry.AllowedMergeMethods = append(entry.AllowedMergeMethods, string(m))
			}
		}
		out.PullRequest = entry
	}

	if sc := rules.RequiredStatusChecks; sc != nil {
		entry := &config.RequiredStatusChecksRule{
			Strict:               sc.StrictRequiredStatusChecksPolicy,
			DoNotEnforceOnCreate: nilIfFalse(sc.DoNotEnforceOnCreate),
		}
		for _, check := range sc.RequiredStatusChecks {
			if check == nil {
				continue
			}
			entry.Checks = append(entry.Checks, config.StatusCheck{
				Context:       check.Context,
				IntegrationID: check.IntegrationID,
			})
		}
		out.RequiredStatusChecks = entry
	}

	if rd := rules.RequiredDeployments; rd != nil {
		out.RequiredDeployments = &config.RequiredDeploymentsRule{
			Environments: rd.RequiredDeploymentEnvironments,
		}
	}

	if mq := rules.MergeQueue; mq != nil {
		out.MergeQueue = &config.MergeQueueRule{
			MergeMethod:                  strings.ToLower(string(mq.MergeMethod)),
			GroupingStrategy:             strings.ToLower(string(mq.GroupingStrategy)),
			CheckResponseTimeoutMinutes:  mq.CheckResponseTimeoutMinutes,
			MaxEntriesToBuild:            mq.MaxEntriesToBuild,
			MaxEntriesToMerge:            mq.MaxEntriesToMerge,
			MinEntriesToMerge:            mq.MinEntriesToMerge,
			MinEntriesToMergeWaitMinutes: mq.MinEntriesToMergeWaitMinutes,
		}
	}

	out.CommitMessagePattern = patternToConfig(rules.CommitMessagePattern)
	out.CommitAuthorEmailPattern = patternToConfig(rules.CommitAuthorEmailPattern)
	out.CommitterEmailPattern = patternToConfig(rules.CommitterEmailPattern)
	out.BranchNamePattern = patternToConfig(rules.BranchNamePattern)
	out.TagNamePattern = patternToConfig(rules.TagNamePattern)

	if wf := rules.Workflows; wf != nil {
		entry := &config.WorkflowsRule{DoNotEnforceOnCreate: nilIfFalse(wf.DoNotEnforceOnCreate)}
		for _, w := range wf.Workflows {
			if w == nil {
				continue
			}
			item := config.RuleWorkflow{Path: w.Path, Ref: w.GetRef(), SHA: w.GetSHA()}
			if w.RepositoryID != nil {
				if name, ok := l.repoNameByID[*w.RepositoryID]; ok {
					item.Repository = name
				} else {
					item.RepositoryID = w.RepositoryID
				}
			}
			entry.Workflows = append(entry.Workflows, item)
		}
		out.Workflows = entry
	}

	if cs := rules.CodeScanning; cs != nil {
		entry := &config.CodeScanningRule{}
		for _, tool := range cs.CodeScanningTools {
			if tool == nil {
				continue
			}
			entry.Tools = append(entry.Tools, config.CodeScanningTool{
				Tool:                    tool.Tool,
				AlertsThreshold:         string(tool.AlertsThreshold),
				SecurityAlertsThreshold: string(tool.SecurityAlertsThreshold),
			})
		}
		out.CodeScanning = entry
	}

	if fe := rules.FileExtensionRestriction; fe != nil {
		out.FileExtensionRestriction = &config.FileExtensionRestrictionRule{
			RestrictedFileExtensions: fe.RestrictedFileExtensions,
		}
	}
	if fp := rules.FilePathRestriction; fp != nil {
		out.FilePathRestriction = &config.FilePathRestrictionRule{
			RestrictedFilePaths: fp.RestrictedFilePaths,
		}
	}
	if v := rules.MaxFilePathLength; v != nil {
		length := v.MaxFilePathLength
		out.MaxFilePathLength = &length
	}
	if v := rules.MaxFileSize; v != nil {
		size := v.MaxFileSize
		out.MaxFileSize = &size
	}

	return out
}

func patternToConfig(p *github.PatternRuleParameters) *config.PatternRule {
	if p == nil {
		return nil
	}
	return &config.PatternRule{
		Name:     p.GetName(),
		Operator: string(p.Operator),
		Pattern:  p.Pattern,
		Negate:   nilIfFalse(p.Negate),
	}
}

// nilIfFalse drops a pointer whose value is false, so an imported config does
// not restate GitHub's own default as though it were a choice.
func nilIfFalse(b *bool) *bool {
	if b == nil || !*b {
		return nil
	}
	return b
}

func isAllMergeMethods(methods []github.PullRequestMergeMethod) bool {
	if len(methods) != 3 {
		return false
	}
	seen := map[string]bool{}
	for _, m := range methods {
		seen[strings.ToLower(string(m))] = true
	}
	return seen["merge"] && seen["squash"] && seen["rebase"]
}
