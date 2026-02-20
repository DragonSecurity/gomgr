package sync

import (
	"testing"
)

func TestParseRepoConfig(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		wantPerm   string
		wantTopics []string
		wantPinned bool
	}{
		{
			name:       "simple string permission",
			input:      "push",
			wantPerm:   "push",
			wantTopics: nil,
			wantPinned: false,
		},
		{
			name: "advanced config with permission only",
			input: map[string]any{
				"permission": "maintain",
			},
			wantPerm:   "maintain",
			wantTopics: nil,
			wantPinned: false,
		},
		{
			name: "advanced config with topics",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{"backend", "api"},
			},
			wantPerm:   "push",
			wantTopics: []string{"backend", "api"},
			wantPinned: false,
		},
		{
			name: "advanced config with pinning",
			input: map[string]any{
				"permission": "admin",
				"topics":     []any{"documentation"},
				"pinned":     true,
			},
			wantPerm:   "admin",
			wantTopics: []string{"documentation"},
			wantPinned: true,
		},
		{
			name: "map[any]any format (YAML unmarshal variant)",
			input: map[any]any{
				"permission": "pull",
				"topics":     []any{"frontend", "web"},
				"pinned":     false,
			},
			wantPerm:   "pull",
			wantTopics: []string{"frontend", "web"},
			wantPinned: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := parseRepoConfig(tt.input)

			if settings.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", settings.permission, tt.wantPerm)
			}

			if len(settings.topics) != len(tt.wantTopics) {
				t.Errorf("topics length = %d, want %d", len(settings.topics), len(tt.wantTopics))
			} else {
				for i, topic := range settings.topics {
					if topic != tt.wantTopics[i] {
						t.Errorf("topics[%d] = %q, want %q", i, topic, tt.wantTopics[i])
					}
				}
			}

			if settings.pinned != tt.wantPinned {
				t.Errorf("pinned = %v, want %v", settings.pinned, tt.wantPinned)
			}
		})
	}
}
