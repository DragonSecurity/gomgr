package sync

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const (
	scopeRepoSettings   = "repo-settings"
	scopeRepoVisibility = "repo-visibility"
)

// settingField is one repository setting gomgr can reconcile: how to read it
// off a repository, and how to ask for it in an edit.
type settingField struct {
	name string
	want func(config.RepoSettingsConfig) *bool
	// current reads the setting off a repository. The second result is false
	// when GitHub did not report the field at all, which is not the same as
	// reporting it off — see newRepoDetailFetcher.
	current func(*github.Repository) (bool, bool)
	apply   func(*github.Repository, bool)
}

// known adapts a plain accessor into one that reports whether the field was
// present, given the pointer it was read from.
func known(ptr *bool) (bool, bool) {
	if ptr == nil {
		return false, false
	}
	return *ptr, true
}

// settingFields is the whole reconcilable set. Adding a setting means adding a
// line here and nothing else.
var settingFields = []settingField{
	{
		name:    "allow_auto_merge",
		want:    func(c config.RepoSettingsConfig) *bool { return c.AllowAutoMerge },
		current: func(r *github.Repository) (bool, bool) { return known(r.AllowAutoMerge) },
		apply:   func(r *github.Repository, v bool) { r.AllowAutoMerge = github.Ptr(v) },
	},
	{
		name:    "allow_squash_merge",
		want:    func(c config.RepoSettingsConfig) *bool { return c.AllowSquashMerge },
		current: func(r *github.Repository) (bool, bool) { return known(r.AllowSquashMerge) },
		apply:   func(r *github.Repository, v bool) { r.AllowSquashMerge = github.Ptr(v) },
	},
	{
		name:    "allow_merge_commit",
		want:    func(c config.RepoSettingsConfig) *bool { return c.AllowMergeCommit },
		current: func(r *github.Repository) (bool, bool) { return known(r.AllowMergeCommit) },
		apply:   func(r *github.Repository, v bool) { r.AllowMergeCommit = github.Ptr(v) },
	},
	{
		name:    "allow_rebase_merge",
		want:    func(c config.RepoSettingsConfig) *bool { return c.AllowRebaseMerge },
		current: func(r *github.Repository) (bool, bool) { return known(r.AllowRebaseMerge) },
		apply:   func(r *github.Repository, v bool) { r.AllowRebaseMerge = github.Ptr(v) },
	},
	{
		name:    "delete_branch_on_merge",
		want:    func(c config.RepoSettingsConfig) *bool { return c.DeleteBranchOnMerge },
		current: func(r *github.Repository) (bool, bool) { return known(r.DeleteBranchOnMerge) },
		apply:   func(r *github.Repository, v bool) { r.DeleteBranchOnMerge = github.Ptr(v) },
	},
	{
		name:    "allow_update_branch",
		want:    func(c config.RepoSettingsConfig) *bool { return c.AllowUpdateBranch },
		current: func(r *github.Repository) (bool, bool) { return known(r.AllowUpdateBranch) },
		apply:   func(r *github.Repository, v bool) { r.AllowUpdateBranch = github.Ptr(v) },
	},
}

// repoDetailFetcher reads a repository's full representation.
//
// The organization listing does not carry these settings — allow_auto_merge,
// allow_merge_commit and delete_branch_on_merge all come back null there, and
// only the single-repository endpoint fills them in. Comparing against the
// listing would read every unset pointer as false and report every repository
// as drifted, which is a plan that lies in the same way the file planner used
// to. So each repository is read once, and only when settings are declared.
type repoDetailFetcher func(ctx context.Context, org, repo string) (*github.Repository, error)

func newRepoDetailFetcher(c *gh.Client) repoDetailFetcher {
	cache := map[string]*github.Repository{}
	return func(ctx context.Context, org, repo string) (*github.Repository, error) {
		key := org + "/" + repo
		if hit, ok := cache[key]; ok {
			return hit, nil
		}
		full, _, err := c.REST.Repositories.Get(ctx, org, repo)
		if err != nil {
			return nil, fmt.Errorf("read repository %s: %w", key, err)
		}
		cache[key] = full
		return full, nil
	}
}

// planRepoSettings emits a change for each managed repository whose settings
// differ from what the configuration asks for.
//
// gomgr sets these same settings when it creates a repository. Reconciling them
// is what extends that to the repositories it did not create, which are exactly
// the ones quietly missing them.
func planRepoSettings(ctx context.Context, fetch repoDetailFetcher, cfg *config.Root, bySettings map[string]repoSettings, existingRepos map[string]*github.Repository) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string

	for _, repo := range sortedKeys(bySettings) {
		desired := bySettings[repo].settings.MergedWith(cfg.Org.RepoDefaults)
		if desired.IsEmpty() {
			continue
		}
		listed, ok := existingRepos[repo]
		if !ok {
			continue // being created this run; creation applies the settings
		}
		current := listed
		if fetch != nil {
			full, err := fetch(ctx, cfg.App.Org, listed.GetName())
			if err != nil {
				return nil, nil, err
			}
			current = full
		}
		if current.GetArchived() {
			warnings = append(warnings, fmt.Sprintf("Skipping settings for archived repository %s", repo))
			continue
		}

		changed := map[string]any{}
		var names []string
		for _, f := range settingFields {
			want := f.want(desired)
			if want == nil {
				continue
			}
			have, reported := f.current(current)
			if !reported {
				// Not knowing the current value is not the same as knowing it
				// is wrong. Guessing here is what produced a plan claiming
				// every repository had drifted.
				warnings = append(warnings, fmt.Sprintf(
					"GitHub did not report %s for %s; leaving it alone", f.name, repo))
				continue
			}
			if have == *want {
				continue
			}
			changed[f.name] = *want
			names = append(names, f.name)
		}
		if len(changed) == 0 {
			continue
		}
		sort.Strings(names)

		details := map[string]any{"org": cfg.App.Org, "repo": current.GetName()}
		for k, v := range changed {
			details[k] = v
		}
		out = append(out, util.Change{
			Scope:   scopeRepoSettings,
			Target:  current.GetName() + " (" + strings.Join(names, ", ") + ")",
			Action:  util.ActionEnsure,
			Details: details,
		})
	}
	return out, warnings, nil
}

// planRepoVisibility emits a change for a repository whose visibility differs
// from the one it explicitly declares.
//
// Two keys have to agree before gomgr will touch visibility: the repository
// must state one, and app.yaml must set reconcile_visibility. Neither an
// organization default nor a single edit can move a repository between public
// and private on its own. A deleted repository can be restored for ninety days;
// a disclosed one cannot be undisclosed, so this is the one setting that gets a
// second lock rather than a comment.
func planRepoVisibility(cfg *config.Root, bySettings map[string]repoSettings, existingRepos map[string]*github.Repository) ([]util.Change, []string) {
	var out []util.Change
	var warnings []string

	for _, repo := range sortedKeys(bySettings) {
		want := bySettings[repo].visibility
		if want == "" {
			continue
		}
		current, ok := existingRepos[repo]
		if !ok {
			continue // creation already honors the declared visibility
		}
		if current.GetVisibility() == want {
			continue
		}

		if !cfg.App.ReconcileVisibility {
			warnings = append(warnings, fmt.Sprintf(
				"Repository %s is %s but declares %s; set reconcile_visibility to let gomgr change it",
				repo, current.GetVisibility(), want))
			continue
		}
		if current.GetArchived() {
			warnings = append(warnings, fmt.Sprintf("Skipping visibility for archived repository %s", repo))
			continue
		}

		out = append(out, util.Change{
			Scope:  scopeRepoVisibility,
			Target: fmt.Sprintf("%s (%s -> %s)", current.GetName(), current.GetVisibility(), want),
			Action: util.ActionEnsure,
			Details: map[string]any{
				"org":        cfg.App.Org,
				"repo":       current.GetName(),
				"visibility": want,
				"from":       current.GetVisibility(),
			},
		})
	}
	return out, warnings
}

func applyRepoSettingsEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	edit := &github.Repository{}
	var applied []string
	for _, f := range settingFields {
		raw, ok := d[f.name]
		if !ok {
			continue
		}
		v, ok := raw.(bool)
		if !ok {
			return fmt.Errorf("setting %s for %s/%s must be a bool, got %T", f.name, org, repo, raw)
		}
		f.apply(edit, v)
		applied = append(applied, f.name)
	}
	if len(applied) == 0 {
		return nil
	}

	updated, _, err := c.REST.Repositories.Edit(ctx, org, repo, edit)
	if err != nil {
		return fmt.Errorf("update settings on %s/%s: %w", org, repo, err)
	}
	return verifyRepoSettingsApplied(updated, d, org, repo)
}

// verifyRepoSettingsApplied checks the repository GitHub returned against what
// was asked for, so a setting the API accepted but did not change is reported
// on the run that tried to change it rather than reappearing in every later
// plan.
func verifyRepoSettingsApplied(got *github.Repository, want map[string]any, org, repo string) error {
	if got == nil {
		return nil
	}
	for _, f := range settingFields {
		raw, ok := want[f.name]
		if !ok {
			continue
		}
		v, _ := raw.(bool)
		have, reported := f.current(got)
		if reported && have != v {
			return fmt.Errorf("%s/%s: GitHub accepted the request but %s is still %v, so the change did not take effect",
				org, repo, f.name, have)
		}
	}
	return nil
}

func applyRepoVisibilityEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	want := detailString(d, "visibility")

	updated, _, err := c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
		Visibility: github.Ptr(want),
	})
	if err != nil {
		return fmt.Errorf("change visibility of %s/%s to %s: %w", org, repo, want, err)
	}
	if updated != nil && updated.GetVisibility() != want {
		return fmt.Errorf("%s/%s: GitHub accepted the request but visibility is still %s",
			org, repo, updated.GetVisibility())
	}
	util.Warnf("%s/%s visibility is now %s", org, repo, want)
	return nil
}
