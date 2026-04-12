package templates

import (
	"bytes"
	"fmt"
	"path"
	"text/template"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// FileData is the context passed to every FileSpec.Content template.
type FileData struct {
	Org  string
	Repo string
}

// RenderFile executes spec.Content as a text/template and returns the result.
// Template errors are wrapped with the file's path so operators can trace the
// offending entry.
func RenderFile(spec config.FileSpec, data FileData) (string, error) {
	tmpl, err := template.New(spec.Path).Option("missingkey=error").Parse(spec.Content)
	if err != nil {
		return "", fmt.Errorf("parse template for %s: %w", spec.Path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template for %s: %w", spec.Path, err)
	}
	return buf.String(), nil
}

// MatchesRepo reports whether a FileSpec applies to the named repository.
// An empty Only list matches every repo; otherwise any single glob match
// (path.Match semantics) is sufficient.
func MatchesRepo(spec config.FileSpec, repo string) bool {
	if len(spec.Only) == 0 {
		return true
	}
	for _, pattern := range spec.Only {
		if ok, _ := path.Match(pattern, repo); ok {
			return true
		}
	}
	return false
}

// DefaultReadmeSpec returns a FileSpec that renders the built-in README
// template. It is emitted when the legacy AddDefaultReadme flag is set.
func DefaultReadmeSpec() config.FileSpec {
	return config.FileSpec{
		Path:    "README.md",
		Content: readmeTemplate,
		Message: "chore: add default README",
		Branch:  "main", // legacy default branch
	}
}
