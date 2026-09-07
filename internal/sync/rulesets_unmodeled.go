package sync

import (
	"reflect"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"
)

// modeledRuleFields are the fields of github.RepositoryRulesetRules that
// rulesToConfig reads and buildRuleset writes back.
//
// Anything outside this set is a rule gomgr cannot express. That matters more
// than it sounds: buildRuleset constructs a ruleset from configuration alone
// and UpdateRuleset replaces what is on GitHub with it, so adopting a ruleset
// containing a rule gomgr does not model would quietly delete that rule on the
// next sync.
//
// TestModeledRuleFieldsCoverGoGitHub fails when go-github grows a field this
// set does not mention, so the list cannot silently fall behind an upgrade.
var modeledRuleFields = map[string]bool{
	"BranchNamePattern":        true,
	"CodeScanning":             true,
	"CommitAuthorEmailPattern": true,
	"CommitMessagePattern":     true,
	"CommitterEmailPattern":    true,
	"Creation":                 true,
	"Deletion":                 true,
	"FileExtensionRestriction": true,
	"FilePathRestriction":      true,
	"MergeQueue":               true,
	"NonFastForward":           true,
	"PullRequest":              true,
	"RequiredDeployments":      true,
	"RequiredLinearHistory":    true,
	"RequiredSignatures":       true,
	"RequiredStatusChecks":     true,
	"TagNamePattern":           true,
	"Update":                   true,
	"Workflows":                true,
}

// unmodeledRuleTypes returns the rule types present on a ruleset that gomgr
// cannot express, named as GitHub names them.
//
// Reflection rather than a hand-written check per field: the point is to notice
// rules gomgr does not know about, and a hand-written check can only look for
// the ones somebody already thought of.
func unmodeledRuleTypes(rules *github.RepositoryRulesetRules) []string {
	if rules == nil {
		return nil
	}
	v := reflect.ValueOf(*rules)
	t := v.Type()

	var out []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || modeledRuleFields[f.Name] {
			continue
		}
		// Every rule field is a pointer; a nil one means the rule is not set.
		if v.Field(i).Kind() == reflect.Pointer && v.Field(i).IsNil() {
			continue
		}
		out = append(out, ruleTypeName(f.Name))
	}
	sort.Strings(out)
	return out
}

// ruleTypeName converts a go-github field name to the snake_case rule type
// GitHub uses in its API and in its user interface, so a message names the
// thing somebody would search for.
func ruleTypeName(field string) string {
	var b strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
