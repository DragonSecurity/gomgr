package config

import "sort"

// Built-in preset names. These are the guard rails gomgr ships with, so a
// config can ask for "the usual protection" without restating a dozen rules.
const (
	PresetBranchProtection       = "branch-protection"
	PresetStrictBranchProtection = "strict-branch-protection"
	PresetTagProtection          = "tag-protection"
	PresetNoForcePush            = "no-force-push"
	PresetRequireSignedCommits   = "require-signed-commits"
	PresetRequireDCO             = "require-dco"
	PresetNoCommittedKeys        = "no-committed-keys"
)

// GitHub's magic ref-name and repository-name selectors.
const (
	refAll           = "~ALL"
	refDefaultBranch = "~DEFAULT_BRANCH"
)

func boolPtr(b bool) *bool { return &b }

// RulesetPresets returns the built-in guard rails, freshly constructed on each
// call. Presets hand out pointers into RulesetRules and callers merge those
// pointers into their own config, so sharing one package-level map between
// rulesets would let them alias each other's rule structs.
//
// A preset supplies target, conditions and rules; a RulesetConfig naming one
// overrides any of those it sets itself. Presets deliberately do not set bypass
// actors — who may bypass a guard rail is a decision for the org, not a default.
func RulesetPresets() map[string]RulesetConfig {
	return map[string]RulesetConfig{
		// The baseline every repository should have: the default branch cannot
		// be deleted or force-pushed, and changes arrive through a reviewed
		// pull request.
		PresetBranchProtection: {
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refDefaultBranch}}},
			Rules: RulesetRules{
				Deletion:       boolPtr(true),
				NonFastForward: boolPtr(true),
				PullRequest: &PullRequestRule{
					RequiredApprovingReviewCount:   1,
					DismissStaleReviewsOnPush:      true,
					RequiredReviewThreadResolution: true,
				},
			},
		},

		// The same, tightened for repositories that carry production or
		// security-relevant code: two approvals, code owners must sign off, the
		// last pusher cannot self-approve, and history stays linear.
		PresetStrictBranchProtection: {
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refDefaultBranch}}},
			Rules: RulesetRules{
				Deletion:              boolPtr(true),
				NonFastForward:        boolPtr(true),
				RequiredLinearHistory: boolPtr(true),
				PullRequest: &PullRequestRule{
					RequiredApprovingReviewCount:   2,
					DismissStaleReviewsOnPush:      true,
					RequireCodeOwnerReview:         true,
					RequireLastPushApproval:        true,
					RequiredReviewThreadResolution: true,
				},
			},
		},

		// Releases are immutable: a published tag cannot be moved or deleted.
		PresetTagProtection: {
			Target:      RulesetTargetTag,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refAll}}},
			Rules: RulesetRules{
				Deletion:       boolPtr(true),
				NonFastForward: boolPtr(true),
			},
		},

		// The minimum viable guard rail for orgs not ready to require reviews:
		// history on every branch is append-only.
		PresetNoForcePush: {
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refAll}}},
			Rules: RulesetRules{
				NonFastForward: boolPtr(true),
			},
		},

		// Every commit reaching the default branch must carry a verified
		// signature. Note that this blocks unsigned automation, gomgr's own
		// file-sync commits included — give the app a bypass actor or scope
		// this ruleset away from repositories gomgr writes to.
		PresetRequireSignedCommits: {
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refDefaultBranch}}},
			Rules: RulesetRules{
				RequiredSignatures: boolPtr(true),
			},
		},

		// Enforce the Developer Certificate of Origin at the ref, which the
		// DCO status check cannot do for pushes that never open a pull request.
		PresetRequireDCO: {
			Target:      RulesetTargetBranch,
			Enforcement: RulesetEnforcementActive,
			Conditions:  &RulesetConditions{RefName: &RefNameCondition{Include: []string{refAll}}},
			Rules: RulesetRules{
				CommitMessagePattern: &PatternRule{
					Name:     "Require DCO sign-off",
					Operator: "contains",
					Pattern:  "Signed-off-by:",
				},
			},
		},

		// Refuse pushes carrying private keys and keystores outright, rather
		// than finding them in a secret-scanning alert after the fact.
		PresetNoCommittedKeys: {
			Target:      RulesetTargetPush,
			Enforcement: RulesetEnforcementActive,
			Rules: RulesetRules{
				FileExtensionRestriction: &FileExtensionRestrictionRule{
					RestrictedFileExtensions: []string{
						".pem", ".key", ".p12", ".pfx", ".jks", ".keystore", ".ppk",
					},
				},
			},
		},
	}
}

// PresetNames returns the built-in preset names in sorted order, for error
// messages and documentation.
func PresetNames() []string {
	presets := RulesetPresets()
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
