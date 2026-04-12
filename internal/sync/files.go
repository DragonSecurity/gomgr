package sync

import (
	"fmt"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const defaultFileBranch = "main"

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
				"org":     org,
				"repo":    repo,
				"path":    spec.Path,
				"content": content,
				"message": message,
				"branch":  branch,
			},
		})
		emittedFiles[dedupeKey] = true
	}
	return out, nil
}
