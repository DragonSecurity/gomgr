package templates

import (
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func TestRenderFile_InterpolatesOrgRepo(t *testing.T) {
	spec := config.FileSpec{
		Path:    "README.md",
		Content: "# {{.Repo}}\n\nPart of {{.Org}}.",
	}
	got, err := RenderFile(spec, FileData{Org: "KaMuses", Repo: "infra"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "# infra") {
		t.Errorf("expected rendered repo name, got %q", got)
	}
	if !strings.Contains(got, "Part of KaMuses.") {
		t.Errorf("expected rendered org name, got %q", got)
	}
}

func TestRenderFile_ParseError(t *testing.T) {
	spec := config.FileSpec{Path: "bad.md", Content: "{{ .Repo "}
	_, err := RenderFile(spec, FileData{Org: "o", Repo: "r"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "bad.md") {
		t.Errorf("expected error to mention path, got %v", err)
	}
}

func TestRenderFile_UnknownFieldErrors(t *testing.T) {
	spec := config.FileSpec{Path: "t.md", Content: "{{.Nope}}"}
	_, err := RenderFile(spec, FileData{Org: "o", Repo: "r"})
	if err == nil {
		t.Fatal("expected execute error for unknown field")
	}
}

func TestMatchesRepo(t *testing.T) {
	tests := []struct {
		name     string
		only     []string
		repo     string
		expected bool
	}{
		{"empty only matches everything", nil, "anything", true},
		{"glob star", []string{"public-*"}, "public-docs", true},
		{"glob star no match", []string{"public-*"}, "internal-api", false},
		{"exact match", []string{"infra"}, "infra", true},
		{"multiple patterns, one matches", []string{"legacy-*", "infra"}, "infra", true},
		{"multiple patterns, none match", []string{"legacy-*", "infra"}, "frontend", false},
		{"question mark", []string{"api-?"}, "api-1", true},
		{"question mark no match", []string{"api-?"}, "api-10", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := config.FileSpec{Path: "f", Content: "c", Only: tt.only}
			if got := MatchesRepo(spec, tt.repo); got != tt.expected {
				t.Errorf("MatchesRepo(%v, %q) = %v, want %v", tt.only, tt.repo, got, tt.expected)
			}
		})
	}
}

func TestDefaultReadmeSpec_RendersBuiltInTemplate(t *testing.T) {
	spec := DefaultReadmeSpec()
	if spec.Path != "README.md" {
		t.Errorf("expected path README.md, got %q", spec.Path)
	}
	out, err := RenderFile(spec, FileData{Org: "Acme", Repo: "widgets"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "# widgets") {
		t.Errorf("expected rendered repo header, got %q", out)
	}
	if !strings.Contains(out, "git@github.com:Acme/widgets.git") {
		t.Errorf("expected rendered clone URL, got %q", out)
	}
}
