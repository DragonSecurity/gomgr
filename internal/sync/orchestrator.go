package sync

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"

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

	// Desired state from config
	DesiredTeams       int
	DesiredTeamMembers int
	DesiredRepos       int
	DesiredRepoPerms   int
	DesiredCustomRoles int
}

func BuildPlan(ctx context.Context, c *gh.Client, cfg *config.Root) (util.Plan, error) {
	st := &State{Org: cfg.App.Org}
	var plan util.Plan

	// Prefetch teams and repos once to avoid duplicate API calls
	if err := prefetchState(ctx, c, st); err != nil {
		return plan, fmt.Errorf("prefetch state: %w", err)
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

	cleanupChanges, warnings, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan cleanups: %w", err)
	}

	customRoleCleanups, roleWarnings, err := planCustomRoleCleanups(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan custom role cleanups: %w", err)
	}

	plan.Changes = append(plan.Changes, customRoleChanges...)
	plan.Changes = append(plan.Changes, teamChanges...)
	plan.Changes = append(plan.Changes, memChanges...)
	plan.Changes = append(plan.Changes, repoChanges...)
	plan.Changes = append(plan.Changes, cleanupChanges...)
	plan.Changes = append(plan.Changes, customRoleCleanups...)
	plan.Warnings = append(warnings, roleWarnings...)

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

func Apply(ctx context.Context, c *gh.Client, plan util.Plan) error {
	return applyChanges(ctx, c, plan.Changes)
}
