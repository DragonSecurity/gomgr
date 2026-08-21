package sync

import (
	"context"
	"fmt"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

type State struct {
	Org          string
	ManagedRepos map[string]bool

	// Cached API results to avoid duplicate calls
	ActualTeams []*github.Team
	ActualRepos []*github.Repository

	// Current state from GitHub
	CurrentTeams       int
	CurrentTeamMembers int
	CurrentRepos       int
	CurrentRepoPerms   int
	CurrentCustomRoles int
	CurrentRulesets    int
	CurrentOwners      int

	// Desired state from config
	DesiredTeams       int
	DesiredTeamMembers int
	DesiredRepos       int
	DesiredRepoPerms   int
	DesiredCustomRoles int
	DesiredRulesets    int
	DesiredOwners      int

	// RepoWarnings are advisories raised while planning repository settings.
	RepoWarnings []string

	// teamMembers caches every team's membership, keyed by slug then by
	// lowercased login. Three planners want it — team cleanups, member
	// cleanups and collaborator enforcement — and it costs a request per team,
	// so it is fetched at most once per run and only when something asks.
	teamMembers map[string]map[string]bool

	// teamRepos caches every team's repository permissions, keyed by slug then
	// by lowercased repository name. Fetched on the same terms as teamMembers.
	teamRepos map[string]map[string]string

	// defaultRepoPermission is the organization-wide baseline every member
	// holds on every repository: permNone, read, write or admin. Empty until
	// something asks for it.
	defaultRepoPermission string
}

func BuildPlan(ctx context.Context, c *gh.Client, cfg *config.Root) (util.Plan, error) {
	st := &State{Org: cfg.App.Org}
	var plan util.Plan

	// Prefetch teams and repos once to avoid duplicate API calls
	if err := prefetchState(ctx, c, st); err != nil {
		return plan, fmt.Errorf("prefetch state: %w", err)
	}

	// Owners come first: bringing an unmanaged org under management may mean
	// there is nobody yet who can approve the rest of the run.
	ownerChanges, ownerWarnings, err := planOrgOwners(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan org owners: %w", err)
	}

	// Custom roles must be created before teams/repos use them
	customRoleChanges, err := planCustomRoles(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan custom roles: %w", err)
	}

	teamChanges, desiredBySlug, err := planTeams(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan teams: %w", err)
	}

	memChanges, err := planTeamMembership(ctx, c, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan team membership: %w", err)
	}

	repoChanges, err := planRepoPerms(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan repo permissions: %w", err)
	}

	// Rulesets are planned after repo permissions so that ManagedRepos and the
	// resolved repo settings reflect what this run will create.
	orgRulesetChanges, orgRulesetWarnings, err := planOrgRulesets(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan org rulesets: %w", err)
	}

	repoRulesetChanges, repoRulesetWarnings, err := planRepoRulesets(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan repo rulesets: %w", err)
	}

	// Direct collaborator grants are compared last, because what someone is
	// entitled to depends on the repositories and teams the steps above have
	// resolved.
	collaboratorChanges, collaboratorWarnings, err := planCollaborators(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan collaborators: %w", err)
	}

	cleanupChanges, warnings, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan cleanups: %w", err)
	}

	customRoleCleanups, roleWarnings, err := planCustomRoleCleanups(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan custom role cleanups: %w", err)
	}

	plan.Changes = append(plan.Changes, ownerChanges...)
	plan.Changes = append(plan.Changes, customRoleChanges...)
	plan.Changes = append(plan.Changes, teamChanges...)
	plan.Changes = append(plan.Changes, memChanges...)
	plan.Changes = append(plan.Changes, repoChanges...)
	plan.Changes = append(plan.Changes, orgRulesetChanges...)
	plan.Changes = append(plan.Changes, repoRulesetChanges...)
	plan.Changes = append(plan.Changes, collaboratorChanges...)
	plan.Changes = append(plan.Changes, cleanupChanges...)
	plan.Changes = append(plan.Changes, customRoleCleanups...)
	plan.Warnings = append(warnings, roleWarnings...)
	plan.Warnings = append(plan.Warnings, ownerWarnings...)
	plan.Warnings = append(plan.Warnings, collaboratorWarnings...)
	plan.Warnings = append(plan.Warnings, orgRulesetWarnings...)
	plan.Warnings = append(plan.Warnings, repoRulesetWarnings...)
	plan.Warnings = append(plan.Warnings, st.RepoWarnings...)

	// Populate stats
	plan.Stats = &util.StateStats{
		Teams: util.StatePair{
			Current: st.CurrentTeams,
			Desired: st.DesiredTeams,
		},
		TeamMembers: util.StatePair{
			Current: st.CurrentTeamMembers,
			Desired: st.DesiredTeamMembers,
		},
		Repositories: util.StatePair{
			Current: st.CurrentRepos,
			Desired: st.DesiredRepos,
		},
		RepoPermissions: util.StatePair{
			Current: st.CurrentRepoPerms,
			Desired: st.DesiredRepoPerms,
		},
		CustomRoles: util.StatePair{
			Current: st.CurrentCustomRoles,
			Desired: st.DesiredCustomRoles,
		},
		Rulesets: util.StatePair{
			Current: st.CurrentRulesets,
			Desired: st.DesiredRulesets,
		},
		Owners: util.StatePair{
			Current: st.CurrentOwners,
			Desired: st.DesiredOwners,
		},
	}

	return plan, nil
}

// prefetchState fetches teams and repos from GitHub once, caching them in State
// so that both planning and cleanup phases can reuse the data.
func prefetchState(ctx context.Context, c *gh.Client, st *State) error {
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, st.Org, opts)
		if err != nil {
			return nil, err
		}
		st.ActualTeams = append(st.ActualTeams, ts...)
		return resp, nil
	}); err != nil {
		return fmt.Errorf("list teams: %w", err)
	}

	repoOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
		Type:        "all",
	}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		repoOpt.ListOptions = *opts
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, st.Org, repoOpt)
		if err != nil {
			return nil, err
		}
		st.ActualRepos = append(st.ActualRepos, repos...)
		return resp, nil
	}); err != nil {
		return fmt.Errorf("list repos: %w", err)
	}

	return nil
}

// ApplyOptions tunes how Apply processes the change set.
type ApplyOptions struct {
	// ContinueOnError keeps applying remaining changes after a handler fails,
	// then returns an aggregated error at the end. When false (the default),
	// the first handler error aborts the run.
	ContinueOnError bool
}

func Apply(ctx context.Context, c *gh.Client, plan util.Plan) error {
	return ApplyWithOptions(ctx, c, plan, ApplyOptions{})
}

// ApplyWithOptions applies the plan's changes using the given options.
func ApplyWithOptions(ctx context.Context, c *gh.Client, plan util.Plan, opts ApplyOptions) error {
	return applyChangesWith(ctx, c, plan.Changes, defaultRegistry, opts)
}
