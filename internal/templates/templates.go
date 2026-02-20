package templates

import (
	"bytes"
	"text/template"
)

// Data represents the data available to templates
type Data struct {
	Org  string
	Repo string
}

// Render renders a template string with the given data
func Render(tmpl string, data Data) (string, error) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderOrPassthrough attempts to render a template, but if it fails or
// the template has no template syntax, it returns the original string
func RenderOrPassthrough(tmpl string, data Data) string {
	result, err := Render(tmpl, data)
	if err != nil {
		return tmpl
	}
	return result
}
