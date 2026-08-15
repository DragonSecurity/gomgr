package sync

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v88/github"
)

// rulesetMatches reports whether the ruleset already on GitHub enforces exactly
// what the desired ruleset asks for.
//
// The comparison is deliberately asymmetric per field, because GitHub does not
// echo back what you sent:
//
//   - Name, target and enforcement are compared exactly.
//   - Rules must be the same set of rule types — dropping a rule from the config
//     has to remove it from GitHub — but each rule's parameters are compared as
//     a subset. GitHub fills omitted parameters with its own defaults (a
//     pull_request rule comes back listing every merge method, patterns grow a
//     `negate: false`), and a strict comparison would report a difference on
//     every run and rewrite the ruleset forever.
//   - Bypass actors are compared as an exact set: an actor nobody configured is
//     a hole in the guard rail, not a harmless default.
//   - Conditions are compared as a subset for the same reason as rule
//     parameters, with the include/exclude lists themselves compared as sets.
func rulesetMatches(actual, desired *github.RepositoryRuleset) (bool, error) {
	if actual == nil || desired == nil {
		return false, nil
	}
	if actual.Name != desired.Name ||
		rulesetTarget(actual) != rulesetTarget(desired) ||
		actual.Enforcement != desired.Enforcement {
		return false, nil
	}

	if !bypassActorsEqual(actual.BypassActors, desired.BypassActors) {
		return false, nil
	}

	actualConds, err := toGeneric(actual.Conditions)
	if err != nil {
		return false, fmt.Errorf("decode current conditions: %w", err)
	}
	desiredConds, err := toGeneric(desired.Conditions)
	if err != nil {
		return false, fmt.Errorf("encode desired conditions: %w", err)
	}
	if !jsonSubset(desiredConds, actualConds) {
		return false, nil
	}

	actualRules, err := rulesByType(actual.Rules)
	if err != nil {
		return false, fmt.Errorf("decode current rules: %w", err)
	}
	desiredRules, err := rulesByType(desired.Rules)
	if err != nil {
		return false, fmt.Errorf("encode desired rules: %w", err)
	}
	if len(actualRules) != len(desiredRules) {
		return false, nil
	}
	for ruleType, desiredParams := range desiredRules {
		actualParams, ok := actualRules[ruleType]
		if !ok {
			return false, nil
		}
		if !jsonSubset(desiredParams, actualParams) {
			return false, nil
		}
	}

	return true, nil
}

// rulesetTarget dereferences the target. The generated accessor hands back a
// pointer for this field, so comparing two of them compares addresses.
func rulesetTarget(rs *github.RepositoryRuleset) github.RulesetTarget {
	if rs == nil || rs.Target == nil {
		return ""
	}
	return *rs.Target
}

// rulesByType flattens a rule set into rule type -> parameters. The API
// represents rules as a heterogeneous array, so going through its own JSON
// encoding is what keeps this in step with the rule list rather than a second
// hand-maintained switch.
func rulesByType(rules *github.RepositoryRulesetRules) (map[string]any, error) {
	out := map[string]any{}
	if rules == nil {
		return out, nil
	}
	b, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}
	var entries []struct {
		Type       string          `json:"type"`
		Parameters json.RawMessage `json:"parameters"`
	}
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		if len(e.Parameters) == 0 {
			out[e.Type] = map[string]any{}
			continue
		}
		var params any
		if err := json.Unmarshal(e.Parameters, &params); err != nil {
			return nil, err
		}
		out[e.Type] = params
	}
	return out, nil
}

// toGeneric round-trips a value through JSON into maps and slices, so the
// comparison below can walk it without knowing its type.
func toGeneric(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// jsonSubset reports whether every key present in want is present in got with
// an equal value. Keys only in got are ignored — those are GitHub's defaults.
// Arrays are compared as sets, since GitHub does not preserve the order of
// include lists or status-check lists.
func jsonSubset(want, got any) bool {
	if want == nil {
		return true
	}
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			return false
		}
		for k, wv := range w {
			gv, present := g[k]
			if !present {
				// An explicit null asks for nothing, so an absent key satisfies it.
				if wv == nil {
					continue
				}
				return false
			}
			if !jsonSubset(wv, gv) {
				return false
			}
		}
		return true
	case []any:
		g, ok := got.([]any)
		if !ok || len(w) != len(g) {
			return false
		}
		return canonicalSet(w) == canonicalSet(g)
	default:
		return want == got
	}
}

// canonicalSet renders a JSON array as an order-independent string key.
func canonicalSet(items []any) string {
	encoded := make([]string, 0, len(items))
	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			// Values here came out of json.Unmarshal, so this cannot fail; fall
			// back to a formatting that still distinguishes different values.
			encoded = append(encoded, fmt.Sprintf("%v", item))
			continue
		}
		encoded = append(encoded, string(b))
	}
	sort.Strings(encoded)
	return "[" + strings.Join(encoded, ",") + "]"
}

// bypassActorsEqual compares two bypass actor lists as sets.
func bypassActorsEqual(a, b []*github.BypassActor) bool {
	if len(a) != len(b) {
		return false
	}
	return canonicalActors(a) == canonicalActors(b)
}

func canonicalActors(actors []*github.BypassActor) string {
	keys := make([]string, 0, len(actors))
	for _, actor := range actors {
		if actor == nil {
			continue
		}
		// The accessors for these two fields return pointers, not values.
		var actorType github.BypassActorType
		if t := actor.GetActorType(); t != nil {
			actorType = *t
		}
		var mode github.BypassMode
		if m := actor.GetBypassMode(); m != nil {
			mode = *m
		}
		keys = append(keys, fmt.Sprintf("%s/%d/%s", actorType, actor.GetActorID(), mode))
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}
