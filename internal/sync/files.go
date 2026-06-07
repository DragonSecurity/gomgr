package sync

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const defaultFileBranch = "main"

// codeownersPath is the canonical CODEOWNERS location gomgr writes to.
// GitHub also reads CODEOWNERS and docs/CODEOWNERS; .github/ is preferred so
// the file stays out of the repo root.
const codeownersPath = ".github/CODEOWNERS"

// materializeFileSpecs merges user-declared app.files with any legacy
// convenience flags (AddDefaultReadme, AddRenovateConfig) so downstream code
// only has to iterate a single list. Legacy entries are prepended; if a user
// entry targets the same path, the user entry wins (dedup keeps the last
// occurrence of each path).
func materializeFileSpecs(app config.AppConfig) []config.FileSpec {
	merged := make([]config.FileSpec, 0, len(app.Files)+2)

	if app.AddDefaultReadme {
		merged = append(merged, templates.DefaultReadmeSpec())
	}
	if app.AddRenovateConfig && strings.TrimSpace(app.RenovateConfig) != "" {
		merged = append(merged, config.FileSpec{
			Path:    ".github/renovate.json",
			Content: app.RenovateConfig,
			Message: "chore: add Renovate config",
			Branch:  defaultFileBranch,
		})
	}

	merged = append(merged, app.Files...)

	latest := map[string]int{}
	for i, f := range merged {
		latest[f.Path] = i
	}
	final := make([]config.FileSpec, 0, len(latest))
	for i, f := range merged {
		if latest[f.Path] == i {
			final = append(final, f)
		}
	}
	return final
}

// planRepoFiles renders each FileSpec for the given repo (when it matches the
// Only filter) and returns a list of repo-file:ensure changes. emittedFiles is
// updated in place so the same path is only emitted once per repo, even when
// multiple teams reference the same repository.
func planRepoFiles(org, repo, repoKey string, specs []config.FileSpec, emittedFiles map[string]bool) ([]util.Change, error) {
	var out []util.Change
	for _, spec := range specs {
		if !templates.MatchesRepo(spec, repo) {
			continue
		}
		dedupeKey := repoKey + ":" + spec.Path
		if emittedFiles[dedupeKey] {
			continue
		}

		content, err := templates.RenderFile(spec, templates.FileData{Org: org, Repo: repo})
		if err != nil {
			return nil, fmt.Errorf("render %s for %s/%s: %w", spec.Path, org, repo, err)
		}

		message := spec.Message
		if message == "" {
			message = fmt.Sprintf("chore: add %s", spec.Path)
		}
		branch := spec.Branch
		if branch == "" {
			branch = defaultFileBranch
		}

		out = append(out, util.Change{
			Scope:  "repo-file",
			Target: repoKey + ":" + spec.Path,
			Action: "ensure",
			Details: map[string]any{
				"org":       org,
				"repo":      repo,
				"path":      spec.Path,
				"content":   content,
				"message":   message,
				"branch":    branch,
				"reconcile": spec.Reconcile,
			},
		})
		emittedFiles[dedupeKey] = true
	}
	return out, nil
}

// normalizeOwnerRef prefixes bare usernames with @ so CODEOWNERS reads
// correctly. Entries that already start with @ (including @org/team refs)
// pass through unchanged.
func normalizeOwnerRef(o string) string {
	if strings.HasPrefix(o, "@") {
		return o
	}
	return "@" + o
}

// renderCodeowners returns the canonical CODEOWNERS body for a single repo:
// a single catch-all rule listing every owner. Duplicate refs are collapsed.
func renderCodeowners(owners []string) string {
	refs := make([]string, 0, len(owners))
	seen := map[string]bool{}
	for _, o := range owners {
		ref := normalizeOwnerRef(o)
		if seen[ref] {
			continue
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	if len(refs) == 0 {
		return ""
	}
	return fmt.Sprintf("* %s\n", strings.Join(refs, " "))
}

// planCodeowners emits a repo-file:ensure change writing .github/CODEOWNERS
// for every repo with declared owners. If the user has already declared a
// CODEOWNERS file via app.files, synthesis is skipped entirely so the
// hand-authored content wins.
//
// ownersByRepo is keyed by lower-cased repo name; repoNames maps that key
// back to the canonical name for the apply payload.
func planCodeowners(org string, ownersByRepo map[string][]string, repoNames map[string]string, userFilePaths map[string]bool, emittedFiles map[string]bool) []util.Change {
	if userFilePaths[codeownersPath] {
		return nil
	}
	keys := make([]string, 0, len(ownersByRepo))
	for r := range ownersByRepo {
		keys = append(keys, r)
	}
	sort.Strings(keys)

	var out []util.Change
	for _, r := range keys {
		owners := ownersByRepo[r]
		content := renderCodeowners(owners)
		if content == "" {
			continue
		}
		dedupeKey := r + ":" + codeownersPath
		if emittedFiles[dedupeKey] {
			continue
		}
		repoName := repoNames[r]
		if repoName == "" {
			repoName = r
		}
		out = append(out, util.Change{
			Scope:  "repo-file",
			Target: dedupeKey,
			Action: "ensure",
			Details: map[string]any{
				"org":       org,
				"repo":      repoName,
				"path":      codeownersPath,
				"content":   content,
				"message":   "chore: sync CODEOWNERS",
				"branch":    defaultFileBranch,
				"reconcile": true,
			},
		})
		emittedFiles[dedupeKey] = true
	}
	return out
}

// planCodeownersDeletions emits a repo-file:delete change for every managed
// repo that has no codeowners declared in YAML. It is gated by the caller via
// app.delete_stale_codeowners; when the user has declared CODEOWNERS in
// app.files the deletion is skipped (hand-authored file wins).
//
// The apply handler is idempotent — a delete against a repo with no
// .github/CODEOWNERS no-ops — so this can safely fire for repos that never
// had the file.
func planCodeownersDeletions(org string, managedRepos map[string]bool, repoNames map[string]string, ownersByRepo map[string][]string, userFilePaths map[string]bool, emittedFiles map[string]bool) []util.Change {
	if userFilePaths[codeownersPath] {
		return nil
	}
	keys := make([]string, 0, len(managedRepos))
	for r := range managedRepos {
		if len(ownersByRepo[r]) > 0 {
			continue
		}
		keys = append(keys, r)
	}
	sort.Strings(keys)

	var out []util.Change
	for _, r := range keys {
		dedupeKey := r + ":" + codeownersPath
		if emittedFiles[dedupeKey] {
			continue
		}
		repoName := repoNames[r]
		if repoName == "" {
			repoName = r
		}
		out = append(out, util.Change{
			Scope:  "repo-file",
			Target: dedupeKey,
			Action: "delete",
			Details: map[string]any{
				"org":     org,
				"repo":    repoName,
				"path":    codeownersPath,
				"message": "chore: remove stale CODEOWNERS",
				"branch":  defaultFileBranch,
			},
		})
		emittedFiles[dedupeKey] = true
	}
	return out
}
