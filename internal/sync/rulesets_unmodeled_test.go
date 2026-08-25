package sync

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
)

// The whole point of the modeled set is that it cannot fall behind go-github.
// If an upgrade adds a rule type, this fails and somebody has to decide whether
// gomgr models it or refuses rulesets that use it — rather than silently
// deleting it from GitHub on the next sync.
func TestModeledRuleFieldsCoverGoGitHub(t *testing.T) {
	known := map[string]bool{
		// Deliberately not modeled. Listing them here is the record that each
		// was considered, so a new field really is new.
		"CopilotCodeReview":    true,
		"MaxFilePathLength":    true,
		"MaxFileSize":          true,
		"RepositoryCreate":     true,
		"RepositoryDelete":     true,
		"RepositoryName":       true,
		"RepositoryTransfer":   true,
		"RepositoryVisibility": true,
	}

	t3 := reflect.TypeOf(github.RepositoryRulesetRules{})
	for i := 0; i < t3.NumField(); i++ {
		name := t3.Field(i).Name
		if !t3.Field(i).IsExported() {
			continue
		}
		if !modeledRuleFields[name] && !known[name] {
			t.Errorf("go-github has rule field %q that gomgr neither models nor knowingly refuses. "+
				"Add it to modeledRuleFields and rulesToConfig, or to the known list here.", name)
		}
	}
}

func TestUnmodeledRuleTypes(t *testing.T) {
	if got := unmodeledRuleTypes(nil); got != nil {
		t.Errorf("no rules means nothing unmodeled, got %v", got)
	}

	// A ruleset gomgr fully understands.
	fine := &github.RepositoryRulesetRules{Deletion: &github.EmptyRuleParameters{}}
	if got := unmodeledRuleTypes(fine); len(got) != 0 {
		t.Errorf("deletion is modeled, got %v", got)
	}

	// The case that started this: a real rule gomgr does not model.
	copilot := &github.RepositoryRulesetRules{
		CopilotCodeReview: &github.CopilotCodeReviewRuleParameters{},
	}
	got := unmodeledRuleTypes(copilot)
	if len(got) != 1 || got[0] != "copilot_code_review" {
		t.Fatalf("want [copilot_code_review], got %v", got)
	}

	// Mixed: the modeled rule must not hide the unmodeled one, because that is
	// the case where adoption succeeds and quietly drops a rule.
	mixed := &github.RepositoryRulesetRules{
		Deletion:          &github.EmptyRuleParameters{},
		CopilotCodeReview: &github.CopilotCodeReviewRuleParameters{},
	}
	if got := unmodeledRuleTypes(mixed); len(got) != 1 || got[0] != "copilot_code_review" {
		t.Errorf("a modeled rule beside an unmodeled one must still report it, got %v", got)
	}
}

func TestRuleTypeName(t *testing.T) {
	for field, want := range map[string]string{
		"CopilotCodeReview": "copilot_code_review",
		"MaxFileSize":       "max_file_size",
		"Deletion":          "deletion",
	} {
		if got := ruleTypeName(field); got != want {
			t.Errorf("%s -> %q, want %q", field, got, want)
		}
	}
}

// The old message said "no rules enabled; a ruleset with no rules enforces
// nothing", which reads as "delete this, it does nothing". Acting on that
// deletes a working rule.
func TestLossyReasonNamesTheRuleRatherThanCallingItEmpty(t *testing.T) {
	rs := &github.RepositoryRuleset{
		Name:  "Code Quality Copilot review for default branch",
		Rules: &github.RepositoryRulesetRules{CopilotCodeReview: &github.CopilotCodeReviewRuleParameters{}},
	}

	reason := lossyReason(rs)

	if !strings.Contains(reason, "copilot_code_review") {
		t.Errorf("the reason must name the rule type: %q", reason)
	}
	if strings.Contains(reason, "enforces nothing") {
		t.Errorf("must not claim the ruleset is empty: %q", reason)
	}
	if !strings.Contains(reason, "delete") {
		t.Errorf("the reason should say what adopting it would cost: %q", reason)
	}
}

func TestLossyReasonIsEmptyForAnAdoptableRuleset(t *testing.T) {
	rs := &github.RepositoryRuleset{
		Name:  "protect",
		Rules: &github.RepositoryRulesetRules{Deletion: &github.EmptyRuleParameters{}},
	}
	if reason := lossyReason(rs); reason != "" {
		t.Errorf("nothing lossy here, got %q", reason)
	}
}

func TestPluralAndPronoun(t *testing.T) {
	if got := plural("rule type", []string{"a"}); got != "rule type a" {
		t.Errorf("got %q", got)
	}
	if got := plural("rule type", []string{"a", "b"}); got != "rule types a and b" {
		t.Errorf("got %q", got)
	}
	if got := plural("rule type", []string{"a", "b", "c"}); got != "rule types a, b and c" {
		t.Errorf("got %q", got)
	}
	if pronounFor([]string{"a"}) != "it" || pronounFor([]string{"a", "b"}) != "them" {
		t.Error("pronoun should follow the count")
	}
}
