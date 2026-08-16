package config

import (
	"fmt"
	"strings"
)

// How gomgr lands the files it writes.
const (
	// FileStrategyAuto derives the route from the rulesets the configuration
	// declares: a pull request where a direct push would be rejected, a direct
	// commit where it would not. This is the default.
	FileStrategyAuto = "auto"
	// FileStrategyDirect always commits straight to the target branch, which is
	// what gomgr did before there was a choice.
	FileStrategyDirect = "direct"
	// FileStrategyPullRequest always routes through a branch and pull request.
	FileStrategyPullRequest = "pull_request"
)

const defaultFileChangeBranch = "gomgr/sync-files"

// FileChangeConfig controls how the files gomgr writes reach their branch.
//
// The strategy is derived rather than declared by default. gomgr already knows
// the organization's guard rails — it manages them — so it can work out whether
// a direct push to a given branch would be rejected instead of asking for a
// flag that has to be kept in step with the rulesets by hand.
//
// Merge behavior is deliberately absent. gomgr decides that at repository
// creation, not per sync: applyRepoEnsure sets AllowAutoMerge true,
// AllowMergeCommit false and DeleteBranchOnMerge true, the same way repository
// visibility is decided there. Offering merge_method here would let a
// configuration ask for a merge commit that gomgr itself disabled.
type FileChangeConfig struct {
	// Strategy is auto (default), direct, or pull_request.
	Strategy string `yaml:"strategy,omitempty"`
	// Branch is the head branch gomgr opens its pull requests from. One branch
	// is reused per repository, so a repeated sync updates the open pull
	// request rather than opening another.
	Branch string `yaml:"branch,omitempty"`
}

// ResolvedStrategy returns the configured strategy, defaulting to auto.
func (f FileChangeConfig) ResolvedStrategy() string {
	if f.Strategy == "" {
		return FileStrategyAuto
	}
	return strings.ToLower(f.Strategy)
}

// ResolvedBranch returns the head branch for gomgr's pull requests.
func (f FileChangeConfig) ResolvedBranch() string {
	if strings.TrimSpace(f.Branch) == "" {
		return defaultFileChangeBranch
	}
	return f.Branch
}

var validFileStrategies = map[string]bool{
	FileStrategyAuto:        true,
	FileStrategyDirect:      true,
	FileStrategyPullRequest: true,
}

// Validate checks the file-change settings.
func (f FileChangeConfig) Validate() error {
	if !validFileStrategies[f.ResolvedStrategy()] {
		return fmt.Errorf("app.file_changes: invalid strategy %q (must be auto, direct or pull_request)", f.Strategy)
	}
	if strings.HasPrefix(f.ResolvedBranch(), "refs/") {
		return fmt.Errorf("app.file_changes: branch must be a branch name, not a ref: %q", f.Branch)
	}
	return nil
}
