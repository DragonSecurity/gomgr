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
// The set is limited to settings that are idempotent, reversible, and cannot
// expose anything: merge and branch hygiene, and the security analyses below.
// Repository visibility is handled separately and far more carefully — see
// AppConfig.ReconcileVisibility.
type RepoSettingsConfig struct {
	AllowAutoMerge      *bool `yaml:"allow_auto_merge,omitempty"`
	AllowSquashMerge    *bool `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch   *bool `yaml:"allow_update_branch,omitempty"`

	// Security analyses. GitHub keeps these in a nested security_and_analysis
	// object and reports each as a status string rather than a flag, but they
	// behave like every other setting here: stating one asks for it, omitting
	// one leaves GitHub's state alone.
	//
	// Secret scanning and push protection are free on public repositories and
	// need GitHub Advanced Security on private ones, so a private repository
	// in an organization without it will be refused by the API.
	//
	// Deliberately absent: advanced_security and code_security. Those turn
	// GHAS itself on, which is a billing decision rather than a hygiene one,
	// and gomgr does not spend money on your behalf. secret_scanning_ai_
	// detection and secret_scanning_non_provider_patterns are absent because
	// go-github does not model them yet.
	SecretScanning               *bool `yaml:"secret_scanning,omitempty"`
	SecretScanningPushProtection *bool `yaml:"secret_scanning_push_protection,omitempty"`
	SecretScanningValidityChecks *bool `yaml:"secret_scanning_validity_checks,omitempty"`
	DependabotSecurityUpdates    *bool `yaml:"dependabot_security_updates,omitempty"`
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
		AllowAutoMerge:               pick(r.AllowAutoMerge, defaults.AllowAutoMerge),
		AllowSquashMerge:             pick(r.AllowSquashMerge, defaults.AllowSquashMerge),
		AllowMergeCommit:             pick(r.AllowMergeCommit, defaults.AllowMergeCommit),
		AllowRebaseMerge:             pick(r.AllowRebaseMerge, defaults.AllowRebaseMerge),
		DeleteBranchOnMerge:          pick(r.DeleteBranchOnMerge, defaults.DeleteBranchOnMerge),
		AllowUpdateBranch:            pick(r.AllowUpdateBranch, defaults.AllowUpdateBranch),
		SecretScanning:               pick(r.SecretScanning, defaults.SecretScanning),
		SecretScanningPushProtection: pick(r.SecretScanningPushProtection, defaults.SecretScanningPushProtection),
		SecretScanningValidityChecks: pick(r.SecretScanningValidityChecks, defaults.SecretScanningValidityChecks),
		DependabotSecurityUpdates:    pick(r.DependabotSecurityUpdates, defaults.DependabotSecurityUpdates),
	}
}

// IsEmpty reports whether nothing at all is declared.
func (r RepoSettingsConfig) IsEmpty() bool {
	return r.AllowAutoMerge == nil &&
		r.AllowSquashMerge == nil &&
		r.AllowMergeCommit == nil &&
		r.AllowRebaseMerge == nil &&
		r.DeleteBranchOnMerge == nil &&
		r.AllowUpdateBranch == nil &&
		r.SecretScanning == nil &&
		r.SecretScanningPushProtection == nil &&
		r.SecretScanningValidityChecks == nil &&
		r.DependabotSecurityUpdates == nil
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

	// Push protection and validity checks are secret scanning features and
	// cannot be on while it is off. Asking for both in one breath is a
	// contradiction the config can be told about now, rather than a 422 later.
	// Leaving secret_scanning unstated is not a contradiction — the repository
	// may already have it on — so that case is left to the plan, which can see
	// the live state.
	for _, dep := range []struct {
		name string
		on   *bool
	}{
		{"secret_scanning_push_protection", r.SecretScanningPushProtection},
		{"secret_scanning_validity_checks", r.SecretScanningValidityChecks},
	} {
		if off(r.SecretScanning) && dep.on != nil && *dep.on {
			return &ConfigError{Where: where, Msg: dep.name + " cannot be true while secret_scanning is false; it is a secret scanning feature"}
		}
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
