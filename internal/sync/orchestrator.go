package sync

import (
	"context"

	"github.com/DragonSecurity/github-org-manager-go/internal/config"
	"github.com/DragonSecurity/github-org-manager-go/internal/gh"
	"github.com/DragonSecurity/github-org-manager-go/util"
)

type State struct{ Org string }

func BuildPlan(ctx context.Context, c *gh.Client, cfg *config.Root) (util.Plan, error) {
	st := &State{Org: cfg.App.Org}
	var plan util.Plan

	// Owners (stub - optional)
	// ownerChanges, err := planOwners(ctx, c, cfg, st)
	// if err != nil { return plan, err }

	teamChanges, desiredBySlug, err := planTeams(ctx, c, cfg, st)
	if err != nil {
		return plan, err
	}

	memChanges, err := planTeamMembership(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, err
	}

	repoChanges, err := planRepoPerms(ctx, c, cfg, st)
	if err != nil {
		return plan, err
	}

	cleanupChanges, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, err
	}

	plan.Changes = append(plan.Changes, teamChanges...)
	plan.Changes = append(plan.Changes, memChanges...)
	plan.Changes = append(plan.Changes, repoChanges...)
	plan.Changes = append(plan.Changes, cleanupChanges...)
	return plan, nil
}

func Apply(ctx context.Context, c *gh.Client, plan util.Plan) error {
	return applyChanges(ctx, c, plan.Changes)
}
