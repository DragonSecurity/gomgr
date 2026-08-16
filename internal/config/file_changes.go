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

const (
	defaultFileChangeBranch = "gomgr/sync-files"
	defaultMergeMethod      = "squash"
)

// FileChangeConfig controls how the files gomgr writes reach their branch.
//
// The strategy is derived rather than declared by default. gomgr already knows
// the organization's guard rails — it manages them — so it can work out whether
// a direct push to a given branch would be rejected instead of asking for a
// flag that has to be kept in step with the rulesets by hand.
type FileChangeConfig struct {
	// Strategy is auto (default), direct, or pull_request.
	Strategy string `yaml:"strategy,omitempty"`
	// Branch is the head branch gomgr opens its pull requests from. One branch
	// is reused per repository, so a repeated sync updates the open pull
	// request rather than opening another.
	Branch string `yaml:"branch,omitempty"`
	// MergeMethod is squash (default), merge, or rebase.
	MergeMethod string `yaml:"merge_method,omitempty"`
	// AutoMerge asks GitHub to merge the pull request once the required checks
	// pass. Defaults to true: it keeps sync unattended without letting gomgr
	// bypass a rule, since GitHub does the merging and only when the rules
	// allow. Set false to leave pull requests for a human.
	AutoMerge *bool `yaml:"auto_merge,omitempty"`
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

// ResolvedMergeMethod returns the merge method, defaulting to squash.
func (f FileChangeConfig) ResolvedMergeMethod() string {
	if f.MergeMethod == "" {
		return defaultMergeMethod
	}
	return strings.ToLower(f.MergeMethod)
}

// ShouldAutoMerge reports whether GitHub should be asked to merge the pull
// request once its checks pass.
func (f FileChangeConfig) ShouldAutoMerge() bool {
	return f.AutoMerge == nil || *f.AutoMerge
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
	if !validMergeMethods[f.ResolvedMergeMethod()] {
		return fmt.Errorf("app.file_changes: invalid merge_method %q (must be merge, squash or rebase)", f.MergeMethod)
	}
	if strings.HasPrefix(f.ResolvedBranch(), "refs/") {
		return fmt.Errorf("app.file_changes: branch must be a branch name, not a ref: %q", f.Branch)
	}
	return nil
}
