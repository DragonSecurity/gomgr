package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// RepoSettingsConfig declares the repository settings gomgr keeps in step.
//
// Every field is a pointer so that "not mentioned" is distinguishable from
// "set to false". An org-wide default that said false for everything it did not
// mention would switch settings off across the organization the first time
// anybody added the block.
//
// The set is deliberately limited to merge and branch hygiene: settings that
// are idempotent, reversible, and cannot expose anything. Repository visibility
// is handled separately and far more carefully — see AppConfig.ReconcileVisibility.
type RepoSettingsConfig struct {
	AllowAutoMerge      *bool `yaml:"allow_auto_merge,omitempty"`
	AllowSquashMerge    *bool `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch   *bool `yaml:"allow_update_branch,omitempty"`
}

// MergedWith returns these settings overlaid on defaults: a field the override
// states wins, a field it omits falls through.
func (r RepoSettingsConfig) MergedWith(defaults RepoSettingsConfig) RepoSettingsConfig {
	pick := func(override, fallback *bool) *bool {
		if override != nil {
			return override
		}
		return fallback
	}
	return RepoSettingsConfig{
		AllowAutoMerge:      pick(r.AllowAutoMerge, defaults.AllowAutoMerge),
		AllowSquashMerge:    pick(r.AllowSquashMerge, defaults.AllowSquashMerge),
		AllowMergeCommit:    pick(r.AllowMergeCommit, defaults.AllowMergeCommit),
		AllowRebaseMerge:    pick(r.AllowRebaseMerge, defaults.AllowRebaseMerge),
		DeleteBranchOnMerge: pick(r.DeleteBranchOnMerge, defaults.DeleteBranchOnMerge),
		AllowUpdateBranch:   pick(r.AllowUpdateBranch, defaults.AllowUpdateBranch),
	}
}

// IsEmpty reports whether nothing at all is declared.
func (r RepoSettingsConfig) IsEmpty() bool {
	return r.AllowAutoMerge == nil &&
		r.AllowSquashMerge == nil &&
		r.AllowMergeCommit == nil &&
		r.AllowRebaseMerge == nil &&
		r.DeleteBranchOnMerge == nil &&
		r.AllowUpdateBranch == nil
}

// Validate rejects a combination GitHub would refuse.
func (r RepoSettingsConfig) Validate(where string) error {
	// GitHub requires at least one merge method to remain enabled. A config
	// that switches all three off is rejected by the API with an error that
	// does not say which setting caused it.
	stated := func(b *bool) bool { return b != nil }
	off := func(b *bool) bool { return b != nil && !*b }
	if stated(r.AllowSquashMerge) && stated(r.AllowMergeCommit) && stated(r.AllowRebaseMerge) &&
		off(r.AllowSquashMerge) && off(r.AllowMergeCommit) && off(r.AllowRebaseMerge) {
		return &ConfigError{Where: where, Msg: "allow_squash_merge, allow_merge_commit and allow_rebase_merge cannot all be false; GitHub requires at least one merge method"}
	}
	return nil
}

// ConfigError is a configuration problem with the place it was found.
type ConfigError struct {
	Where string
	Msg   string
}

func (e *ConfigError) Error() string { return e.Where + ": " + e.Msg }

// ParseRepoSettings decodes a `settings:` block that arrived as untyped YAML,
// which is the shape a repository entry in teams/*.yaml is held in.
func ParseRepoSettings(v any) (RepoSettingsConfig, error) {
	var out RepoSettingsConfig
	if v == nil {
		return out, nil
	}
	b, err := yaml.Marshal(v)
	if err != nil {
		return out, fmt.Errorf("re-encode settings: %w", err)
	}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return out, fmt.Errorf("parse settings: %w", err)
	}
	return out, nil
}
