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

// detailArchived is the details key carrying the desired archived state. It is
// read back by applyRepoArchiveEnsure, across files, so it is a constant rather
// than a literal repeated at each end.
const detailArchived = "archived"

// planRepoArchive reconciles each managed repository's `archived:` declaration
// against GitHub.
//
// Only an explicit value does anything. A repository whose entry says nothing
// is left as GitHub has it, so one somebody archived by hand stays archived
// until a configuration says otherwise in as many words — un-archiving is never
// something a config does by omission.
//
// The two directions are emitted as different scopes because they run at
// opposite ends of the plan: an archived repository rejects writes, so
// un-archiving precedes the file, settings and topic changes for that
// repository, and archiving follows them.
func planRepoArchive(org string, bySettings map[string]repoSettings, existingRepos map[string]*github.Repository) []util.Change {
	var out []util.Change
	for _, repo := range sortedKeys(bySettings) {
		want := bySettings[repo].archived
		if want == nil {
			continue
		}
		current, ok := existingRepos[repo]
		if !ok {
			// Being created this run. GitHub creates repositories unarchived,
			// so only "archive it" is a change, and it can be planned blind.
			if *want {
				out = append(out, archiveChange(scopeRepoArchive, org, current, repo, true))
			}
			continue
		}
		if current.GetArchived() == *want {
			continue
		}
		scope := scopeRepoArchive
		if !*want {
			scope = scopeRepoUnarchive
		}
		out = append(out, archiveChange(scope, org, current, repo, *want))
	}
	return out
}

// archiveChange builds one archive or un-archive change. name is taken from
// GitHub when it is known, so the API call carries the repository's real
// casing rather than the lowercased key.
func archiveChange(scope, org string, current *github.Repository, repo string, archived bool) util.Change {
	name := repo
	if current != nil && current.GetName() != "" {
		name = current.GetName()
	}
	return util.Change{
		Scope:  scope,
		Target: repo,
		Action: util.ActionEnsure,
		Details: map[string]any{
			"org":          org,
			"repo":         name,
			detailArchived: archived,
		},
	}
}

// unarchivingThisRun returns the repositories a plan will un-archive, so the
// planners that skip archived repositories do not skip one that is about to
// stop being archived.
func unarchivingThisRun(bySettings map[string]repoSettings) map[string]bool {
	out := map[string]bool{}
	for repo, s := range bySettings {
		if s.archived != nil && !*s.archived {
			out[repo] = true
		}
	}
	return out
}

// planUnmanagedArchive archives the repositories no team names.
//
// This is the reversible sibling of deleting them. When both flags are set this
// one wins: being wrong in the direction somebody can undo is the whole reason
// to have it, and the deletion is reported rather than silently dropped.
//
// A repository repos.yaml names is left alone even though no team manages it,
// for the reason declaredInReposFile gives.
func planUnmanagedArchive(cfg *config.Root, st *State) ([]util.Change, []string) {
	if !cfg.App.ArchiveUnmanagedRepos {
		return nil, nil
	}
	var out []util.Change
	var archived []string
	declared := declaredInReposFile(cfg)
	for _, repo := range st.ActualRepos {
		name := repo.GetName()
		key := strings.ToLower(name)
		if st.ManagedRepos[key] || declared[key] || repo.GetArchived() {
			continue
		}
		archived = append(archived, name)
		out = append(out, util.Change{
			Scope:  scopeRepoArchive,
			Target: strings.ToLower(name),
			Action: util.ActionEnsure,
			Details: map[string]any{
				"org":          st.Org,
				"repo":         name,
				detailArchived: true,
			},
		})
	}
	sort.Strings(archived)

	var warnings []string
	if len(archived) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"Archiving %d unmanaged repositories: %v (reversible — un-archive from the repository's settings)",
			len(archived), archived))
	}
	if cfg.App.DeleteUnmanagedRepos {
		warnings = append(warnings, "Both archive_unmanaged_repos and delete_unmanaged_repos are set; "+
			"archiving wins and nothing is deleted. Remove one so the configuration says which you meant.")
	}
	return out, warnings
}

// applyRepoArchiveEnsure sets or clears a repository's archived flag.
func applyRepoArchiveEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	archived := detailBool(d, detailArchived)

	_, _, err = c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
		Archived: github.Ptr(archived),
	})
	if err != nil {
		verb := "archive"
		if !archived {
			verb = "un-archive"
		}
		return fmt.Errorf("%s %s/%s: %w", verb, org, repo, err)
	}
	return nil
}
