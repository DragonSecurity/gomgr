package config

import (
	"fmt"
	"sort"
	"strings"
)

// knownRepoKeys is every key a repository entry may carry.
var knownRepoKeys = map[string]bool{
	"permission": true,
	"topics":     true,
	"pinned":     true,
	"template":   true,
	"from":       true,
	"visibility": true,
	"rulesets":   true,
	"codeowners": true,
	"settings":   true,
	"files":      true,
	"archived":   true,
}

// RejectUnknownRepoKeys refuses a key gomgr does not understand.
//
// Ignoring one is not harmless. A configuration writing "permissions: admin"
// instead of "permission: admin" declared no permission at all, so gomgr
// granted the empty string, GitHub read that as its default of read, and the
// team sat on read access to a repository the configuration said was admin —
// for as long as nobody read the plan closely, because gomgr re-granted it on
// every run and reported success each time.
func RejectUnknownRepoKeys(m map[string]any) error {
	var unknown []string
	for k := range m {
		if !knownRepoKeys[k] {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)

	var hints []string
	for _, k := range unknown {
		if near := NearestRepoKey(k); near != "" {
			hints = append(hints, fmt.Sprintf("%q (did you mean %q?)", k, near))
			continue
		}
		hints = append(hints, fmt.Sprintf("%q", k))
	}
	return fmt.Errorf("unknown setting %s", strings.Join(hints, ", "))
}

// NearestRepoKey returns a known key the given one is probably a slip of,
// which for the mistakes that actually happen — a plural, a missing letter, a
// transposition — means one edit away.
func NearestRepoKey(k string) string {
	best, bestDist := "", 3
	for known := range knownRepoKeys {
		if d := editDistance(strings.ToLower(k), known); d < bestDist {
			best, bestDist = known, d
		}
	}
	return best
}

// editDistance is Levenshtein distance, used only for the hint above.
func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(min(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}
