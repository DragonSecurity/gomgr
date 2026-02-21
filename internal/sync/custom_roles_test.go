package sync

import (
	"testing"
)

func TestPermissionsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "both empty",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "equal permissions",
			a:    []string{"read", "write"},
			b:    []string{"read", "write"},
			want: true,
		},
		{
			name: "equal permissions different order",
			a:    []string{"write", "read"},
			b:    []string{"read", "write"},
			want: true,
		},
		{
			name: "different lengths",
			a:    []string{"read"},
			b:    []string{"read", "write"},
			want: false,
		},
		{
			name: "different permissions",
			a:    []string{"read", "admin"},
			b:    []string{"read", "write"},
			want: false,
		},
		{
			name: "one empty",
			a:    []string{"read"},
			b:    []string{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := permissionsEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("permissionsEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
