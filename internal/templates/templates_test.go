package templates

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     Data
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.Org}}",
			data:     Data{Org: "TestOrg"},
			want:     "Hello TestOrg",
			wantErr:  false,
		},
		{
			name:     "template with both fields",
			template: "Org: {{.Org}}, Repo: {{.Repo}}",
			data:     Data{Org: "MyOrg", Repo: "MyRepo"},
			want:     "Org: MyOrg, Repo: MyRepo",
			wantErr:  false,
		},
		{
			name:     "no template syntax",
			template: "plain text",
			data:     Data{Org: "TestOrg"},
			want:     "plain text",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "{{.InvalidField}}",
			data:     Data{Org: "TestOrg"},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderOrPassthrough(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     Data
		want     string
	}{
		{
			name:     "valid template",
			template: "Hello {{.Org}}",
			data:     Data{Org: "TestOrg"},
			want:     "Hello TestOrg",
		},
		{
			name:     "invalid template returns original",
			template: "{{.InvalidField}}",
			data:     Data{Org: "TestOrg"},
			want:     "{{.InvalidField}}",
		},
		{
			name:     "no template syntax",
			template: "plain text",
			data:     Data{Org: "TestOrg"},
			want:     "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderOrPassthrough(tt.template, tt.data)
			if got != tt.want {
				t.Errorf("RenderOrPassthrough() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderRenovateConfig(t *testing.T) {
	// Test a realistic Renovate config template
	template := `{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["github>{{.Org}}/renovate-presets"]
}`
	data := Data{Org: "DragonSecurity"}
	got, err := Render(template, data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(got, "github>DragonSecurity/renovate-presets") {
		t.Errorf("Rendered template doesn't contain expected org reference: %s", got)
	}
}
