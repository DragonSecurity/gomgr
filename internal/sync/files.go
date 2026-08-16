package sync

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const defaultFileBranch = "main"

// codeownersPath is the canonical CODEOWNERS location gomgr writes to.
// GitHub also reads CODEOWNERS and docs/CODEOWNERS; .github/ is preferred so
// the file stays out of the repo root.
const codeownersPath = ".github/CODEOWNERS"

// signOffPrefix is the trailer key DCO tooling and commit_message_pattern
// rulesets look for.
const signOffPrefix = "Signed-off-by:"

// withSignOff appends a Signed-off-by trailer to a commit message so the
// commit satisfies DCO rulesets. It is a no-op when signOff is empty or when
// the message already carries a trailer — user-supplied app.files[].message
// values may sign themselves off, and double trailers are not harmful but are
// noise in the log.
//
// The trailer is separated by a blank line so it forms a proper git trailer
// block rather than being folded into the subject or body.
func withSignOff(message, signOff string) string {
	if signOff == "" || strings.Contains(message, signOffPrefix) {
		return message
	}
	return strings.TrimRight(message, "\n") + "\n\n" + signOffPrefix + " " + signOff
}

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

// remoteFile is the current state of a file on a branch.
type remoteFile struct {
	exists  bool
	content string
}

// fileProbe reads a file's current state so planning can tell a real change
// from a no-op.
//
// A nil probe means "assume every file needs writing", which is what a
// repository being created in this same run requires — there is nothing to read
// yet — and what tests use when the answer is not what they are checking.
type fileProbe func(ctx context.Context, org, repo, path, branch string) (remoteFile, error)

// newFileProbe returns a probe backed by the API, memoized per file so that a
// repository referenced by several teams is only read once.
func newFileProbe(c *gh.Client) fileProbe {
	cache := map[string]remoteFile{}
	return func(ctx context.Context, org, repo, path, branch string) (remoteFile, error) {
		key := org + "/" + repo + "@" + branch + ":" + path
		if hit, ok := cache[key]; ok {
			return hit, nil
		}
		file, _, resp, err := c.REST.Repositories.GetContents(ctx, org, repo, path,
			&github.RepositoryContentGetOptions{Ref: branch})
		switch {
		case err != nil && resp != nil && resp.StatusCode == http.StatusNotFound:
			// Also covers a branch that does not exist yet.
		case err != nil:
			return remoteFile{}, fmt.Errorf("read %s/%s:%s: %w", org, repo, path, err)
		}

		out := remoteFile{}
		if file != nil {
			content, err := file.GetContent()
			if err != nil {
				return remoteFile{}, fmt.Errorf("decode %s/%s:%s: %w", org, repo, path, err)
			}
			out = remoteFile{exists: true, content: content}
		}
		cache[key] = out
		return out, nil
	}
}

// fileNeedsWrite reports whether a file has to be written, given what is
// already on the branch. It is the same decision applyRepoFileEnsure makes,
// lifted to plan time so that `--dry` reports files that will actually change
// rather than files that will merely be checked.
//
// Without this, every managed repository contributes a repo-file:ensure to
// every plan forever, whether or not anything differs — which both hides real
// drift in the noise and, under reconcile, makes the plan claim commits to
// default branches that will never happen.
func fileNeedsWrite(current remoteFile, desired string, reconcile bool) bool {
	switch {
	case !current.exists:
		return true
	case !reconcile:
		// The file is there and we are not tracking drift for it.
		return false
	default:
		return current.content != desired
	}
}

// planRepoFiles renders each FileSpec for the given repo (when it matches the
// Only filter) and returns a list of repo-file:ensure changes. emittedFiles is
// updated in place so the same path is only emitted once per repo, even when
// multiple teams reference the same repository.
//
// probe, when non-nil, is consulted so files that already match are left out of
// the plan entirely.
//
// signOff, when non-empty, is appended to every commit message as a
// Signed-off-by trailer. It is applied here rather than at apply time so a dry
// run shows the message that will actually be committed.
func planRepoFiles(ctx context.Context, probe fileProbe, route fileRouter, org, repo, repoKey string, specs []config.FileSpec, signOff string, emittedFiles map[string]bool) ([]util.Change, error) {
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
		message = withSignOff(message, signOff)
		branch := spec.Branch
		if branch == "" {
			branch = defaultFileBranch
		}

		if probe != nil {
			current, err := probe(ctx, org, repo, spec.Path, branch)
			if err != nil {
				return nil, err
			}
			if !fileNeedsWrite(current, content, spec.Reconcile) {
				emittedFiles[dedupeKey] = true
				continue
			}
		}

		details := map[string]any{
			"org":       org,
			"repo":      repo,
			"path":      spec.Path,
			"content":   content,
			"message":   message,
			"branch":    branch,
			"reconcile": spec.Reconcile,
		}
		scope := "repo-file"
		if route != nil {
			decision, err := route(repo, branch)
			if err != nil {
				return nil, err
			}
			if decision.UsePullRequest {
				scope = scopeRepoFilePR
				decision.decorate(details)
			}
		}

		out = append(out, util.Change{
			Scope:   scope,
			Target:  repoKey + ":" + spec.Path,
			Action:  "ensure",
			Details: details,
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
func planCodeowners(org string, ownersByRepo map[string][]string, repoNames map[string]string, userFilePaths map[string]bool, signOff string, emittedFiles map[string]bool) []util.Change {
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
				"message":   withSignOff("chore: sync CODEOWNERS", signOff),
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
func planCodeownersDeletions(org string, managedRepos map[string]bool, repoNames map[string]string, ownersByRepo map[string][]string, userFilePaths map[string]bool, signOff string, emittedFiles map[string]bool) []util.Change {
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
				"message": withSignOff("chore: remove stale CODEOWNERS", signOff),
				"branch":  defaultFileBranch,
			},
		})
		emittedFiles[dedupeKey] = true
	}
	return out
}

// markFinalPullRequestFiles flags the last pull-request-routed file change for
// each repository.
//
// Several files in one repository share a head branch and therefore one pull
// request. Auto-merge must only be enabled once the last commit is on that
// branch, or GitHub could merge the pull request while gomgr is still adding to
// it — and the rest of the files would silently never land.
func markFinalPullRequestFiles(changes []util.Change) {
	lastForRepo := map[string]int{}
	for i, ch := range changes {
		if ch.Scope != scopeRepoFilePR {
			continue
		}
		d, ok := ch.Details.(map[string]any)
		if !ok {
			continue
		}
		lastForRepo[fmt.Sprint(d["org"], "/", d["repo"])] = i
	}
	for _, i := range lastForRepo {
		if d, ok := changes[i].Details.(map[string]any); ok {
			d["final"] = true
		}
	}
}
