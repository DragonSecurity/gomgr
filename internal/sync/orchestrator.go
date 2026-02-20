package sync

import (
	"context"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/plan"
)

type State struct{ Org string }

func BuildPlan(ctx context.Context, c *gh.Client, cfg *config.Root) (plan.Plan, error) {
	st := &State{Org: cfg.App.Org}
	var p plan.Plan

	// Owners (stub - optional)
	// ownerChanges, err := planOwners(ctx, c, cfg, st)
	// if err != nil { return plan, err }

	teamChanges, desiredBySlug, err := planTeams(ctx, c, cfg, st)
	if err != nil {
		return p, err
	}

	memChanges, err := planTeamMembership(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return p, err
	}

	repoChanges, err := planRepoPerms(ctx, c, cfg, st)
	if err != nil {
		return p, err
	}

	cleanupChanges, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return p, err
	}

	p.Changes = append(p.Changes, teamChanges...)
	p.Changes = append(p.Changes, memChanges...)
	p.Changes = append(p.Changes, repoChanges...)
	p.Changes = append(p.Changes, cleanupChanges...)
	return p, nil
}

func Apply(ctx context.Context, c *gh.Client, p plan.Plan) error {
	return applyChanges(ctx, c, p.Changes)
}
