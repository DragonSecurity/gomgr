package sync

import (
	"context"
	"fmt"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func rulesetDetails(ch util.Change) (rulesetChange, error) {
	d, ok := ch.Details.(rulesetChange)
	if !ok {
		return rulesetChange{}, fmt.Errorf("invalid details for %s:%s: expected rulesetChange, got %T", ch.Scope, ch.Action, ch.Details)
	}
	return d, nil
}

func applyOrgRulesetUpsert(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := rulesetDetails(ch)
	if err != nil {
		return err
	}
	rs, err := buildRuleset(ctx, d.Spec, true, "", newApplyLookup(d.Org, c))
	if err != nil {
		return err
	}
	if d.ID == 0 {
		if _, _, err := c.REST.Organizations.CreateRepositoryRuleset(ctx, d.Org, *rs); err != nil {
			return fmt.Errorf("create org ruleset %q in %s: %w", d.Name, d.Org, err)
		}
		return nil
	}
	if _, _, err := c.REST.Organizations.UpdateRepositoryRuleset(ctx, d.Org, d.ID, *rs); err != nil {
		return fmt.Errorf("update org ruleset %q (ID %d) in %s: %w", d.Name, d.ID, d.Org, err)
	}
	return nil
}

func applyOrgRulesetDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := rulesetDetails(ch)
	if err != nil {
		return err
	}
	if _, err := c.REST.Organizations.DeleteRepositoryRuleset(ctx, d.Org, d.ID); err != nil {
		return fmt.Errorf("delete org ruleset %q (ID %d) in %s: %w", d.Name, d.ID, d.Org, err)
	}
	return nil
}

func applyRepoRulesetUpsert(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := rulesetDetails(ch)
	if err != nil {
		return err
	}
	rs, err := buildRuleset(ctx, d.Spec, false, d.Repo, newApplyLookup(d.Org, c))
	if err != nil {
		return err
	}
	if d.ID == 0 {
		if _, _, err := c.REST.Repositories.CreateRuleset(ctx, d.Org, d.Repo, *rs); err != nil {
			return fmt.Errorf("create ruleset %q on %s/%s: %w", d.Name, d.Org, d.Repo, err)
		}
		return nil
	}
	if _, _, err := c.REST.Repositories.UpdateRuleset(ctx, d.Org, d.Repo, d.ID, *rs); err != nil {
		return fmt.Errorf("update ruleset %q (ID %d) on %s/%s: %w", d.Name, d.ID, d.Org, d.Repo, err)
	}
	return nil
}

func applyRepoRulesetDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := rulesetDetails(ch)
	if err != nil {
		return err
	}
	if _, err := c.REST.Repositories.DeleteRuleset(ctx, d.Org, d.Repo, d.ID); err != nil {
		return fmt.Errorf("delete ruleset %q (ID %d) on %s/%s: %w", d.Name, d.ID, d.Org, d.Repo, err)
	}
	return nil
}
