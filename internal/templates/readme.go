package templates

import (
	"bytes"
	"text/template"
)

const readmeTemplate = `# {{.Repo}}

Quick setup — if you've done this kind of thing before  
or clone directly:  

` + "```bash" + `
git clone git@github.com:{{.Org}}/{{.Repo}}.git
` + "```" + `

Get started by creating a new file or uploading an existing one.  
We recommend every repository include a README, LICENSE, and .gitignore.

…or create a new repository on the command line

` + "```bash" + `
echo "# {{.Repo}}" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin git@github.com:{{.Org}}/{{.Repo}}.git
git push -u origin main
` + "```" + `

…or push an existing repository from the command line

` + "```bash" + `
git remote add origin git@github.com:{{.Org}}/{{.Repo}}.git
git branch -M main
git push -u origin main
` + "```" + `
`

var readmeTmpl = template.Must(template.New("readme").Parse(readmeTemplate))

// ReadmeData holds the data for README template
type ReadmeData struct {
	Org  string
	Repo string
}

// GenerateReadme generates a README.md content from template
func GenerateReadme(org, repo string) (string, error) {
	data := ReadmeData{
		Org:  org,
		Repo: repo,
	}

	var buf bytes.Buffer
	if err := readmeTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
