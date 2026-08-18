package sync

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
)

// Change scopes for rulesets. Org rulesets are the org-wide guard rails;
// repo rulesets stack on top of them for a single repository.
const (
	scopeOrgRuleset  = "org-ruleset"
	scopeRepoRuleset = "repo-ruleset"
)

// refSelectorAll is GitHub's wildcard for ref and repository conditions: it is
// what a ruleset covers when the config does not narrow it.
const refSelectorAll = "~ALL"

// bypassActorSelf marks the app gomgr is authenticated as, so an imported
// ruleset does not pin the numeric ID of whichever app happened to run it.
const bypassActorSelf = "self"

// rulesetChange carries a planned ruleset mutation. Spec is the resolved
// configuration (preset expanded, references resolved as far as planning could
// see them) rather than a finished github.RepositoryRuleset, because a bypass
// actor may name a team that only exists once earlier changes in this same run
// have been applied. The handler rebuilds the ruleset at apply time, when those
// resources are real.
type rulesetChange struct {
	Org  string
	Repo string // empty for an organization-level ruleset
	ID   int64  // 0 when creating
	Name string
	Spec config.RulesetConfig
}

// refLookup turns names in the config — a team slug, a repository name, "self"
// — into the numeric IDs the rulesets API wants.
//
// With client nil it resolves only against state already fetched, which is what
// planning uses: a plan must not fire an API call per bypass actor, and a team
// it cannot see yet simply means "assume this ruleset needs applying". With a
// client set it falls back to the API, which is what apply uses, by which point
// teams and repositories created earlier in the run exist.
type refLookup struct {
	org    string
	appID  int64
	teams  map[string]int64 // team slug (lowercase) -> team ID
	repos  map[string]int64 // repo name (lowercase) -> repo ID
	client *gh.Client
}

// newPlanLookup builds a lookup backed purely by the prefetched state.
func newPlanLookup(cfg *config.Root, st *State) *refLookup {
	l := &refLookup{
		org:   st.Org,
		appID: cfg.App.ResolvedAppID(),
		teams: make(map[string]int64, len(st.ActualTeams)),
		repos: make(map[string]int64, len(st.ActualRepos)),
	}
	for _, t := range st.ActualTeams {
		l.teams[strings.ToLower(t.GetSlug())] = t.GetID()
	}
	for _, r := range st.ActualRepos {
		l.repos[strings.ToLower(r.GetName())] = r.GetID()
	}
	return l
}

// newApplyLookup builds a lookup that reaches the API for anything it is asked
// about, used once the plan is being applied.
func newApplyLookup(org string, c *gh.Client) *refLookup {
	return &refLookup{
		org:    org,
		teams:  map[string]int64{},
		repos:  map[string]int64{},
		client: c,
	}
}

func (l *refLookup) teamID(ctx context.Context, slug string) (int64, error) {
	key := strings.ToLower(slug)
	if id, ok := l.teams[key]; ok {
		return id, nil
	}
	if l.client == nil {
		return 0, fmt.Errorf("team %q not found in the organization", slug)
	}
	team, _, err := l.client.REST.Teams.GetTeamBySlug(ctx, l.org, slug)
	if err != nil {
		return 0, fmt.Errorf("look up team %q: %w", slug, err)
	}
	l.teams[key] = team.GetID()
	return team.GetID(), nil
}

func (l *refLookup) repoID(ctx context.Context, name string) (int64, error) {
	key := strings.ToLower(name)
	if id, ok := l.repos[key]; ok {
		return id, nil
	}
	if l.client == nil {
		return 0, fmt.Errorf("repository %q not found in the organization", name)
	}
	repo, _, err := l.client.REST.Repositories.Get(ctx, l.org, name)
	if err != nil {
		return 0, fmt.Errorf("look up repository %q: %w", name, err)
	}
	l.repos[key] = repo.GetID()
	return repo.GetID(), nil
}

// integrationID resolves the `app:` field of an Integration bypass actor.
// "self" means gomgr's own GitHub App, which is how a config keeps gomgr's
// file-sync pushes working under a ruleset that would otherwise require a pull
// request. It is unavailable under PAT auth, where there is no app to exempt.
func (l *refLookup) integrationID(app string) (int64, error) {
	if strings.EqualFold(app, "self") {
		if l.appID == 0 {
			return 0, fmt.Errorf("app: self needs app.app_id (or GITHUB_APP_ID) to be set; it has no meaning under PAT auth")
		}
		return l.appID, nil
	}
	id, err := strconv.ParseInt(app, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("app %q is neither %q nor a numeric GitHub App ID", app, "self")
	}
	return id, nil
}

// buildRuleset converts a resolved RulesetConfig into the API representation.
// orgLevel selects the conditions GitHub requires at each scope: an org ruleset
// must say which repositories it covers, a repository ruleset must not.
func buildRuleset(ctx context.Context, spec config.RulesetConfig, orgLevel bool, repo string, l *refLookup) (*github.RepositoryRuleset, error) {
	target := github.RulesetTarget(spec.Target)
	source := l.org
	if !orgLevel {
		source = l.org + "/" + repo
	}

	rs := &github.RepositoryRuleset{
		Name:        spec.Name,
		Target:      &target,
		Source:      source,
		Enforcement: github.RulesetEnforcement(spec.Enforcement),
		Conditions:  buildConditions(spec, orgLevel),
	}

	actors, err := buildBypassActors(ctx, spec.BypassActors, l)
	if err != nil {
		return nil, fmt.Errorf("ruleset %q: %w", spec.Name, err)
	}
	rs.BypassActors = actors

	rules, err := buildRules(ctx, spec.Rules, l)
	if err != nil {
		return nil, fmt.Errorf("ruleset %q: %w", spec.Name, err)
	}
	rs.Rules = rules

	return rs, nil
}

// buildConditions fills in the targeting GitHub requires but the config need
// not repeat: every ref unless narrowed, every repository unless narrowed.
// A push ruleset has no ref conditions at all.
func buildConditions(spec config.RulesetConfig, orgLevel bool) *github.RepositoryRulesetConditions {
	conds := &github.RepositoryRulesetConditions{}
	hasAny := false

	if spec.Target != config.RulesetTargetPush {
		include, exclude := []string{refSelectorAll}, []string{}
		if spec.Conditions != nil && spec.Conditions.RefName != nil {
			if len(spec.Conditions.RefName.Include) > 0 {
				include = spec.Conditions.RefName.Include
			}
			if spec.Conditions.RefName.Exclude != nil {
				exclude = spec.Conditions.RefName.Exclude
			}
		}
		conds.RefName = &github.RepositoryRulesetRefConditionParameters{Include: include, Exclude: exclude}
		hasAny = true
	}

	if orgLevel {
		include, exclude := []string{refSelectorAll}, []string{}
		var protected *bool
		if spec.Conditions != nil && spec.Conditions.RepositoryName != nil {
			rn := spec.Conditions.RepositoryName
			if len(rn.Include) > 0 {
				include = rn.Include
			}
			if rn.Exclude != nil {
				exclude = rn.Exclude
			}
			protected = rn.Protected
		}
		conds.RepositoryName = &github.RepositoryRulesetRepositoryNamesConditionParameters{
			Include:   include,
			Exclude:   exclude,
			Protected: protected,
		}
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return conds
}

func buildBypassActors(ctx context.Context, actors []config.BypassActorConfig, l *refLookup) ([]*github.BypassActor, error) {
	// Empty but not nil, deliberately. The field is tagged omitzero, so a nil
	// slice is dropped from the request body entirely — and GitHub reads an
	// absent bypass_actors as "leave them as they are". Removing the last
	// bypass actor from a configuration would then be planned as a change,
	// applied as a no-op, and planned again on the next run, for ever, while
	// the actor it was trying to remove stayed exempt. Sending [] says what is
	// meant.
	out := make([]*github.BypassActor, 0, len(actors))
	for _, a := range actors {
		kind := a.NormalizedType()
		if kind == "" {
			return nil, fmt.Errorf("bypass actor has unknown type %q", a.Type)
		}

		var id int64
		switch {
		// OrganizationAdmin, EnterpriseOwner and DeployKey are identified by
		// type alone. GitHub reports them with no actor_id, so sending one
		// would make every later comparison see a difference that is not there.
		case a.IdentifiedByTypeAlone():
		case a.ActorID != 0:
			id = a.ActorID
		case kind == config.BypassActorTypeTeam:
			resolved, err := l.teamID(ctx, a.Team)
			if err != nil {
				return nil, fmt.Errorf("bypass actor: %w", err)
			}
			id = resolved
		case kind == config.BypassActorTypeIntegration:
			resolved, err := l.integrationID(a.App)
			if err != nil {
				return nil, fmt.Errorf("bypass actor: %w", err)
			}
			id = resolved
		}

		actorType := github.BypassActorType(kind)
		mode := github.BypassMode(a.BypassMode())
		actor := &github.BypassActor{ActorType: &actorType, BypassMode: &mode}
		if id != 0 {
			actor.ActorID = &id
		}
		out = append(out, actor)
	}
	return out, nil
}

// add indirection without adding clarity.
//
//nolint:gocyclo // A flat translation of the rule schema; helpers per rule would
func buildRules(ctx context.Context, r config.RulesetRules, l *refLookup) (*github.RepositoryRulesetRules, error) {
	rules := &github.RepositoryRulesetRules{}
	empty := &github.EmptyRuleParameters{}

	if isEnabled(r.Creation) {
		rules.Creation = empty
	}
	if isEnabled(r.Deletion) {
		rules.Deletion = empty
	}
	if isEnabled(r.RequiredLinearHistory) {
		rules.RequiredLinearHistory = empty
	}
	if isEnabled(r.RequiredSignatures) {
		rules.RequiredSignatures = empty
	}
	if isEnabled(r.NonFastForward) {
		rules.NonFastForward = empty
	}
	if r.Update != nil {
		rules.Update = &github.UpdateRuleParameters{UpdateAllowsFetchAndMerge: r.Update.AllowsFetchAndMerge}
	}

	if pr := r.PullRequest; pr != nil {
		params := &github.PullRequestRuleParameters{
			DismissStaleReviewsOnPush:      pr.DismissStaleReviewsOnPush,
			RequireCodeOwnerReview:         pr.RequireCodeOwnerReview,
			RequireLastPushApproval:        pr.RequireLastPushApproval,
			RequiredApprovingReviewCount:   pr.RequiredApprovingReviewCount,
			RequiredReviewThreadResolution: pr.RequiredReviewThreadResolution,
		}
		for _, m := range pr.AllowedMergeMethods {
			params.AllowedMergeMethods = append(params.AllowedMergeMethods, github.PullRequestMergeMethod(strings.ToLower(m)))
		}
		rules.PullRequest = params
	}

	if sc := r.RequiredStatusChecks; sc != nil {
		params := &github.RequiredStatusChecksRuleParameters{
			StrictRequiredStatusChecksPolicy: sc.Strict,
			DoNotEnforceOnCreate:             sc.DoNotEnforceOnCreate,
		}
		for _, check := range sc.Checks {
			params.RequiredStatusChecks = append(params.RequiredStatusChecks, &github.RuleStatusCheck{
				Context:       check.Context,
				IntegrationID: check.IntegrationID,
			})
		}
		rules.RequiredStatusChecks = params
	}

	if rd := r.RequiredDeployments; rd != nil {
		rules.RequiredDeployments = &github.RequiredDeploymentsRuleParameters{
			RequiredDeploymentEnvironments: rd.Environments,
		}
	}

	if mq := r.MergeQueue; mq != nil {
		// Every merge_queue parameter is required by the API, so a config that
		// only cares about one of them still has to send the rest. These match
		// the defaults GitHub's own UI offers.
		rules.MergeQueue = &github.MergeQueueRuleParameters{
			CheckResponseTimeoutMinutes:  defaultInt(mq.CheckResponseTimeoutMinutes, 60),
			GroupingStrategy:             github.MergeGroupingStrategy(strings.ToUpper(defaultString(mq.GroupingStrategy, "allgreen"))),
			MaxEntriesToBuild:            defaultInt(mq.MaxEntriesToBuild, 5),
			MaxEntriesToMerge:            defaultInt(mq.MaxEntriesToMerge, 5),
			MergeMethod:                  github.MergeQueueMergeMethod(strings.ToUpper(defaultString(mq.MergeMethod, "merge"))),
			MinEntriesToMerge:            defaultInt(mq.MinEntriesToMerge, 1),
			MinEntriesToMergeWaitMinutes: defaultInt(mq.MinEntriesToMergeWaitMinutes, 5),
		}
	}

	rules.CommitMessagePattern = buildPattern(r.CommitMessagePattern)
	rules.CommitAuthorEmailPattern = buildPattern(r.CommitAuthorEmailPattern)
	rules.CommitterEmailPattern = buildPattern(r.CommitterEmailPattern)
	rules.BranchNamePattern = buildPattern(r.BranchNamePattern)
	rules.TagNamePattern = buildPattern(r.TagNamePattern)

	if wf := r.Workflows; wf != nil {
		params := &github.WorkflowsRuleParameters{DoNotEnforceOnCreate: wf.DoNotEnforceOnCreate}
		for _, w := range wf.Workflows {
			entry := &github.RuleWorkflow{Path: w.Path, RepositoryID: w.RepositoryID}
			if entry.RepositoryID == nil {
				id, err := l.repoID(ctx, w.Repository)
				if err != nil {
					return nil, fmt.Errorf("workflows rule: %w", err)
				}
				entry.RepositoryID = &id
			}
			if w.Ref != "" {
				entry.Ref = github.Ptr(w.Ref)
			}
			if w.SHA != "" {
				entry.SHA = github.Ptr(w.SHA)
			}
			params.Workflows = append(params.Workflows, entry)
		}
		rules.Workflows = params
	}

	if cs := r.CodeScanning; cs != nil {
		params := &github.CodeScanningRuleParameters{}
		for _, tool := range cs.Tools {
			params.CodeScanningTools = append(params.CodeScanningTools, &github.RuleCodeScanningTool{
				Tool:                    tool.Tool,
				AlertsThreshold:         github.CodeScanningAlertsThreshold(defaultString(tool.AlertsThreshold, "errors")),
				SecurityAlertsThreshold: github.CodeScanningSecurityAlertsThreshold(defaultString(tool.SecurityAlertsThreshold, "high_or_higher")),
			})
		}
		rules.CodeScanning = params
	}

	if fe := r.FileExtensionRestriction; fe != nil {
		rules.FileExtensionRestriction = &github.FileExtensionRestrictionRuleParameters{
			RestrictedFileExtensions: fe.RestrictedFileExtensions,
		}
	}
	if fp := r.FilePathRestriction; fp != nil {
		rules.FilePathRestriction = &github.FilePathRestrictionRuleParameters{
			RestrictedFilePaths: fp.RestrictedFilePaths,
		}
	}
	if v := r.MaxFilePathLength; v != nil {
		rules.MaxFilePathLength = &github.MaxFilePathLengthRuleParameters{MaxFilePathLength: *v}
	}
	if v := r.MaxFileSize; v != nil {
		rules.MaxFileSize = &github.MaxFileSizeRuleParameters{MaxFileSize: *v}
	}

	return rules, nil
}

func buildPattern(p *config.PatternRule) *github.PatternRuleParameters {
	if p == nil {
		return nil
	}
	params := &github.PatternRuleParameters{
		Operator: github.PatternRuleOperator(strings.ToLower(p.Operator)),
		Pattern:  p.Pattern,
		Negate:   p.Negate,
	}
	if p.Name != "" {
		params.Name = github.Ptr(p.Name)
	}
	return params
}

func isEnabled(b *bool) bool { return b != nil && *b }

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
