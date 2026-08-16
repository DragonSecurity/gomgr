package sync

import (
	"context"
	"fmt"

	"github.com/google/go-github/v90/github"

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
		created, _, err := c.REST.Organizations.CreateRepositoryRuleset(ctx, d.Org, *rs)
		if err != nil {
			return fmt.Errorf("create org ruleset %q in %s: %w", d.Name, d.Org, err)
		}
		return verifyRulesetApplied(created, rs, d.Name, d.Org)
	}
	updated, _, err := c.REST.Organizations.UpdateRepositoryRuleset(ctx, d.Org, d.ID, *rs)
	if err != nil {
		return fmt.Errorf("update org ruleset %q (ID %d) in %s: %w", d.Name, d.ID, d.Org, err)
	}
	return verifyRulesetApplied(updated, rs, d.Name, d.Org)
}

// verifyRulesetApplied checks that the ruleset GitHub returned is the one that
// was asked for.
//
// The create and update endpoints answer with the resulting ruleset, so this
// costs no extra call — and it closes a gap that is otherwise invisible. An
// accepted request is not the same as an applied one: a field the request did
// not manage to express is silently kept as it was, gomgr reports success, and
// the next plan plans the same change again, for ever.
//
// That matters most in a pipeline that treats a successful apply as proof.
// Reporting success for a change that did not happen is worse than failing,
// because the failure is what a reviewer would have acted on.
func verifyRulesetApplied(got, want *github.RepositoryRuleset, name, where string) error {
	if got == nil {
		return nil // nothing came back to check against
	}
	same, err := rulesetMatches(got, want)
	if err != nil {
		return fmt.Errorf("verify ruleset %q on %s: %w", name, where, err)
	}
	if !same {
		return fmt.Errorf("ruleset %q on %s: GitHub accepted the request but the ruleset it returned "+
			"is not what was asked for, so the change did not take effect and the next plan will show "+
			"it again", name, where)
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
	where := d.Org + "/" + d.Repo
	if d.ID == 0 {
		created, _, err := c.REST.Repositories.CreateRuleset(ctx, d.Org, d.Repo, *rs)
		if err != nil {
			return fmt.Errorf("create ruleset %q on %s: %w", d.Name, where, err)
		}
		return verifyRulesetApplied(created, rs, d.Name, where)
	}
	updated, _, err := c.REST.Repositories.UpdateRuleset(ctx, d.Org, d.Repo, d.ID, *rs)
	if err != nil {
		return fmt.Errorf("update ruleset %q (ID %d) on %s: %w", d.Name, d.ID, where, err)
	}
	return verifyRulesetApplied(updated, rs, d.Name, where)
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
