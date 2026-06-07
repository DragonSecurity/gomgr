This file is a merged representation of the entire codebase, combined into a single document by Repomix.

# File Summary

## Purpose
This file contains a packed representation of the entire repository's contents.
It is designed to be easily consumable by AI systems for analysis, code review,
or other automated processes.

## File Format
The content is organized as follows:
1. This summary section
2. Repository information
3. Directory structure
4. Repository files (if enabled)
5. Multiple file entries, each consisting of:
  a. A header with the file path (## File: path/to/file)
  b. The full contents of the file in a code block

## Usage Guidelines
- This file should be treated as read-only. Any changes should be made to the
  original repository files, not this packed version.
- When processing this file, use the file path to distinguish
  between different files in the repository.
- Be aware that this file may contain sensitive information. Handle it with
  the same level of security as you would the original repository.

## Notes
- Some files may have been excluded based on .gitignore rules and Repomix's configuration
- Binary files are not included in this packed representation. Please refer to the Repository Structure section for a complete list of file paths, including binary files
- Files matching patterns in .gitignore are excluded
- Files matching default ignore patterns are excluded
- Files are sorted by Git change count (files with more changes are at the bottom)

# Directory Structure
```
.claude/
  settings.local.json
.github/
  ISSUE_TEMPLATE/
    bug_report.md
    feature_request.md
  workflows/
    ci.yaml
    release.yaml
  renovate.json
cmd/
  cmd_test.go
  root.go
  setup_team.go
  sync.go
  validate.go
  version.go
config/
  example/
    .github/
      workflows/
        sync.yaml
    teams/
      example-team.yaml
    app.yaml
    org.yaml
examples/
  config/
    .github/
      workflows/
        sync.yaml
    teams/
      backend-team.yaml
      devops-team.yaml
      example-team.yaml
      frontend-team.yaml
      github-actions-team.yaml
      security-team.yaml
    app.yaml
    org.yaml
internal/
  config/
    loader_test.go
    loader.go
    types_test.go
    types.go
  gh/
    client_test.go
    client.go
    rate_test.go
    rate.go
    retry_test.go
    retry.go
  sync/
    apply_handlers_test.go
    apply_handlers.go
    apply_registry_test.go
    apply_registry.go
    custom_roles_test.go
    custom_roles.go
    orchestrator_test.go
    orchestrator.go
    teams_test.go
    teams.go
  templates/
    readme_test.go
    readme.go
  util/
    diff_test.go
    diff.go
    log_test.go
    log.go
  version/
    version_test.go
    version.go
.gitignore
.golangci.yml
.repomixignore
AGENTS.md
go.mod
LICENSE.md
main.go
Makefile
README.md
repomix.config.json
```

# Files

## File: .claude/settings.local.json
````json
{
  "permissions": {
    "allow": [
      "Bash(wc -l /Users/dragon/projects/DragonSecurity/gomgr/internal/sync/*.go)",
      "Bash(wc -l /Users/dragon/projects/DragonSecurity/gomgr/internal/**/*.go)",
      "Bash(go build:*)",
      "Bash(go test:*)",
      "Bash(go vet:*)",
      "Bash(make vet:*)",
      "Bash(make test:*)",
      "Bash(/tmp/gomgr sync:*)",
      "Bash(go mod:*)",
      "Bash(grep -n \"applyChanges\\\\|^func Apply\" internal/sync/*.go)",
      "Bash(go tool:*)",
      "Bash(grep:*)",
      "Bash(golangci-lint run:*)",
      "Bash(golangci-lint:*)",
      "Bash(make lint:*)",
      "WebSearch",
      "WebFetch(domain:golangci-lint.run)",
      "Bash(gofmt -w internal/sync/orchestrator.go)",
      "Bash(goimports -w cmd/setup_team.go cmd/sync.go)",
      "Bash(gofmt:*)",
      "Bash(go install:*)",
      "Bash(git stash:*)",
      "Bash(gh api:*)",
      "Bash(git add:*)",
      "Bash(git commit:*)",
      "Bash(git --version)",
      "Read(//Users/dragon/.buddy/**)",
      "Bash(claude mcp:*)",
      "mcp__buddy__buddy_status",
      "mcp__buddy__buddy_hatch",
      "Bash(wc -l internal/**/*.go)",
      "Bash(ls -la /Users/dragon/projects/DragonSecurity/gomgr/cmd/*_test.go)",
      "mcp__buddy__buddy_observe",
      "Bash(cat examples/config/teams/*.yaml)"
    ]
  }
}
````

## File: .github/ISSUE_TEMPLATE/bug_report.md
````markdown
---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Desktop (please complete the following information):**
 - OS: [e.g. iOS]
 - Browser [e.g. chrome, safari]
 - Version [e.g. 22]

**Smartphone (please complete the following information):**
 - Device: [e.g. iPhone6]
 - OS: [e.g. iOS8.1]
 - Browser [e.g. stock browser, safari]
 - Version [e.g. 22]

**Additional context**
Add any other context about the problem here.
````

## File: .github/ISSUE_TEMPLATE/feature_request.md
````markdown
---
name: Feature request
about: Suggest an idea for this project
title: ''
labels: ''
assignees: ''

---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
````

## File: .github/renovate.json
````json
{ 
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }
````

## File: cmd/cmd_test.go
````go
package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// runCmd invokes rootCmd with the given args, capturing anything written to
// either the cobra Out/Err sinks or os.Stdout (several commands use fmt.Println
// directly). The caller gets stdout, stderr, and the command's error back.
//
// Tests that use this helper must not run in parallel — rootCmd is a package
// singleton with shared flag state.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cfgDir = ""
	debug = false
	dryRun = false
	timeout = 10 * time.Minute
	auditLog = false
	teamName = ""
	outFile = ""
	resetFlagsChanged(rootCmd)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)

	origStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	os.Stdout = w

	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	_ = w.Close()
	os.Stdout = origStdout
	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	return stdoutBuf.String() + outBuf.String(), errBuf.String(), execErr
}

// resetFlagsChanged walks a cobra command tree and clears the Changed flag on
// every pflag so required-flag detection works across sequential Execute calls.
func resetFlagsChanged(c *cobra.Command) {
	clear := func(f *pflag.Flag) { f.Changed = false }
	c.Flags().VisitAll(clear)
	c.PersistentFlags().VisitAll(clear)
	for _, child := range c.Commands() {
		resetFlagsChanged(child)
	}
}

// writeConfigDir builds a minimal but valid gomgr config tree under dir and
// returns the dir path. The team file references a repo called "api".
func writeConfigDir(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, contents string) {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(dir, "app.yaml"), "org: testorg\n")
	write(filepath.Join(dir, "org.yaml"), "owners:\n  - alice\n")
	write(filepath.Join(dir, "teams", "backend.yaml"), `name: Backend
slug: backend
privacy: closed
maintainers:
  - alice
members:
  - bob
repositories:
  api:
    permission: push
`)
	return dir
}

func TestVersionCommand(t *testing.T) {
	stdout, _, err := runCmd(t, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(stdout, "Version:") {
		t.Errorf("expected 'Version:' in output, got %q", stdout)
	}
}

func TestValidate_MissingConfigFlag(t *testing.T) {
	_, _, err := runCmd(t, "validate")
	if err == nil {
		t.Fatal("expected error when --config is missing")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error to mention --config, got %v", err)
	}
}

func TestValidate_NonexistentConfigDir(t *testing.T) {
	_, _, err := runCmd(t, "validate", "-c", "/nonexistent/path/gomgr-test")
	if err == nil {
		t.Fatal("expected error for nonexistent config dir")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	dir := writeConfigDir(t, t.TempDir())
	stdout, _, err := runCmd(t, "validate", "-c", dir)
	if err != nil {
		t.Fatalf("validate failed on good config: %v", err)
	}
	if !strings.Contains(stdout, "Configuration is valid") {
		t.Errorf("expected success message, got %q", stdout)
	}
}

func TestValidate_InvalidConfig_EmptyOrg(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("org: \"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "org.yaml"), []byte("owners: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := runCmd(t, "validate", "-c", dir)
	if err == nil {
		t.Fatal("expected validation error for empty org")
	}
	if !strings.Contains(err.Error(), "org") {
		t.Errorf("expected error to mention 'org', got %v", err)
	}
}

func TestValidate_InvalidConfig_BadTeamPrivacy(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile := func(path, contents string) {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(filepath.Join(dir, "app.yaml"), "org: testorg\n")
	writeFile(filepath.Join(dir, "org.yaml"), "owners:\n  - alice\n")
	writeFile(filepath.Join(dir, "teams", "bad.yaml"), "name: Bad\nprivacy: open\n")

	_, _, err := runCmd(t, "validate", "-c", dir)
	if err == nil {
		t.Fatal("expected validation error for bad privacy")
	}
	if !strings.Contains(err.Error(), "privacy") {
		t.Errorf("expected error to mention 'privacy', got %v", err)
	}
}

func TestSetupTeam_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	_, _, err := runCmd(t, "setup-team", "-c", dir, "-n", "Backend Team")
	if err != nil {
		t.Fatalf("setup-team failed: %v", err)
	}
	// slug should be lowercase with dashes.
	expected := filepath.Join(dir, "teams", "backend-team.yaml")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected team file at %s, got: %v", expected, err)
	}
	b, err := os.ReadFile(expected)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Backend Team") {
		t.Errorf("expected team file to contain team name, got %q", string(b))
	}
}

func TestSetupTeam_MissingNameFlag(t *testing.T) {
	dir := t.TempDir()
	_, _, err := runCmd(t, "setup-team", "-c", dir)
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestSetupTeam_ExplicitOutFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "frontend.yaml")
	_, _, err := runCmd(t, "setup-team", "-c", dir, "-n", "Frontend", "-f", out)
	if err != nil {
		t.Fatalf("setup-team failed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected file at %s, got: %v", out, err)
	}
}

func TestSync_MissingConfigFlag(t *testing.T) {
	_, _, err := runCmd(t, "sync")
	if err == nil {
		t.Fatal("expected error when --config is missing")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error to mention --config, got %v", err)
	}
}

func TestRoot_UnknownCommand(t *testing.T) {
	_, _, err := runCmd(t, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
````

## File: examples/config/teams/backend-team.yaml
````yaml
name: Backend Team
slug: backend-team
description: Backend developers responsible for APIs and services
privacy: closed

# Multiple maintainers - team leads and senior engineers
maintainers:
  - alice-backend-lead
  - bob-senior-engineer
  - charlie-tech-lead

# Multiple members - regular team members
members:
  - david-developer
  - emma-engineer
  - frank-junior-dev
  - grace-contractor
  - henry-intern

repositories:
  # Core backend services with admin access
  backend-api:
    permission: admin
    topics:
      - backend
      - api
      - core-service
      - project-platform
  
  # Microservices with push access
  user-service:
    permission: push
    topics:
      - backend
      - microservice
      - users
  
  auth-service:
    permission: push
    topics:
      - backend
      - microservice
      - authentication
  
  # Shared libraries with maintain access
  backend-common:
    permission: maintain
    topics:
      - backend
      - library
      - shared
  
  # Documentation with pull access
  backend-docs:
    permission: pull
    topics:
      - backend
      - documentation
  
  # Infrastructure code with triage access (can manage issues but not code)
  backend-infra:
    permission: triage
    topics:
      - backend
      - infrastructure
      - terraform
````

## File: examples/config/teams/devops-team.yaml
````yaml
name: DevOps Team
slug: devops-team
description: DevOps engineers responsible for CI/CD, infrastructure, and deployments
privacy: closed

# DevOps team leads
maintainers:
  - rachel-devops-lead
  - sam-sre-manager

# DevOps engineers and specialists
members:
  - tom-k8s-specialist
  - uma-cicd-engineer
  - victor-security-ops
  - wendy-cloud-architect
  - xavier-monitoring-expert

repositories:
  # Infrastructure as code repositories
  terraform-infra:
    permission: admin
    topics:
      - infrastructure
      - terraform
      - aws
      - project-platform
  
  kubernetes-config:
    permission: admin
    topics:
      - infrastructure
      - kubernetes
      - k8s
      - project-platform
  
  # CI/CD pipelines
  github-workflows:
    permission: maintain
    topics:
      - cicd
      - github-actions
      - automation
  
  # Monitoring and observability
  monitoring-config:
    permission: push
    topics:
      - monitoring
      - prometheus
      - grafana
  
  # Security and compliance
  security-policies:
    permission: push
    topics:
      - security
      - compliance
      - policies
  
  # Scripts and automation
  automation-scripts:
    permission: push
    topics:
      - automation
      - scripts
      - tooling
````

## File: examples/config/teams/frontend-team.yaml
````yaml
name: Frontend Team
slug: frontend-team
description: Frontend developers responsible for web and mobile interfaces
privacy: closed

# Multiple maintainers
maintainers:
  - isabella-frontend-lead
  - jack-ui-architect

# Multiple members with diverse roles
members:
  - karen-react-dev
  - liam-vue-specialist
  - maya-ux-engineer
  - noah-accessibility-expert
  - olivia-junior-dev
  - peter-design-engineer
  - quinn-mobile-dev

repositories:
  # Main frontend applications
  web-app:
    permission: admin
    topics:
      - frontend
      - react
      - web
      - project-platform
  
  mobile-app:
    permission: admin
    topics:
      - frontend
      - react-native
      - mobile
      - project-platform
  
  # Shared UI components library
  component-library:
    permission: maintain
    topics:
      - frontend
      - react
      - components
      - library
  
  # Design system
  design-system:
    permission: push
    topics:
      - frontend
      - design
      - storybook
  
  # Documentation
  frontend-docs:
    permission: push
    topics:
      - frontend
      - documentation
````

## File: examples/config/teams/security-team.yaml
````yaml
name: Security Team
slug: security-team
description: Security engineers responsible for application security and compliance
privacy: secret

# Security team is typically secret for sensitive access
maintainers:
  - yara-security-lead
  - zack-appsec-manager

# Security team members
members:
  - anna-pentester
  - ben-security-analyst
  - clara-compliance-officer

repositories:
  # Security has read access to most repos for auditing
  backend-api:
    permission: pull
    topics:
      - security-audit
  
  web-app:
    permission: pull
    topics:
      - security-audit
  
  # Maintain access to security-specific repos
  security-scanning:
    permission: maintain
    topics:
      - security
      - scanning
      - sast
      - dast
  
  vulnerability-reports:
    permission: admin
    topics:
      - security
      - vulnerabilities
      - private
  
  # Admin access to security policies
  security-policies:
    permission: admin
    topics:
      - security
      - compliance
      - policies
````

## File: internal/config/loader_test.go
````go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg
create_repo: true
`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners:
  - alice
`)
	teamsDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(teamsDir, "backend.yaml"), `name: Backend
slug: backend
members:
  - alice
repositories:
  api: push
`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.App.Org != "myorg" {
		t.Errorf("expected org=myorg, got %q", root.App.Org)
	}
	if !root.App.CreateRepo {
		t.Error("expected CreateRepo=true")
	}
	if len(root.Org.Owners) != 1 || root.Org.Owners[0] != "alice" {
		t.Errorf("expected owners=[alice], got %v", root.Org.Owners)
	}
	if len(root.Team) != 1 {
		t.Fatalf("expected 1 team, got %d", len(root.Team))
	}
	if root.Team[0].Name != "Backend" {
		t.Errorf("expected team name=Backend, got %q", root.Team[0].Name)
	}
}

func TestLoad_MissingAppYaml(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing app.yaml")
	}
	if !strings.Contains(err.Error(), "app.yaml") {
		t.Errorf("expected error about app.yaml, got: %v", err)
	}
}

func TestLoad_MissingOrg(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `create_repo: true`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty org")
	}
	if !strings.Contains(err.Error(), "app.org is required") {
		t.Errorf("expected 'app.org is required' error, got: %v", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `{{{invalid yaml`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse YAML") {
		t.Errorf("expected parse YAML error, got: %v", err)
	}
}

func TestLoad_NoTeamsDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Team) != 0 {
		t.Errorf("expected 0 teams, got %d", len(root.Team))
	}
}

func TestLoad_IgnoresNonYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `org: myorg`)
	writeFile(t, filepath.Join(dir, "org.yaml"), `owners: []`)
	teamsDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(teamsDir, ".DS_Store"), "binary junk")
	writeFile(t, filepath.Join(teamsDir, "README.md"), "# Teams")
	writeFile(t, filepath.Join(teamsDir, "backend.yaml"), `name: Backend
members:
  - alice
`)

	root, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Team) != 1 {
		t.Errorf("expected 1 team (ignoring non-YAML), got %d", len(root.Team))
	}
}

func TestBootstrapTeamYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "teams", "new-team.yaml")

	if err := BootstrapTeamYAML(path, "New Team"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "name: New Team") {
		t.Errorf("expected 'name: New Team' in output, got:\n%s", content)
	}
}

func TestResolvedSlug(t *testing.T) {
	tests := []struct {
		name string
		tc   TeamConfig
		want string
	}{
		{
			name: "explicit slug",
			tc:   TeamConfig{Name: "Backend", Slug: "be-team"},
			want: "be-team",
		},
		{
			name: "derived from name",
			tc:   TeamConfig{Name: "Backend Team"},
			want: "backend-team",
		},
		{
			name: "empty both",
			tc:   TeamConfig{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tc.ResolvedSlug()
			if got != tt.want {
				t.Errorf("ResolvedSlug() = %q, want %q", got, tt.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
````

## File: internal/gh/retry_test.go
````go
package gh

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRetryTransport_SuccessOnFirstAttempt(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestRetryTransport_RetriesOn502(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 3 {
		t.Errorf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 2),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 after exhausting retries, got %d", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3 total
	if c := atomic.LoadInt32(&calls); c != 3 {
		t.Errorf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_NoRetryOnClientError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Errorf("expected 1 call (no retry on 404), got %d", c)
	}
}

func TestRetryTransport_RespectsRetryAfter(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Errorf("expected 2 calls, got %d", c)
	}
}
````

## File: internal/gh/retry.go
````go
package gh

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// retryTransport wraps an http.RoundTripper and retries on transient failures
// (5xx responses and 429 rate limits) with exponential backoff and jitter.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

// newRetryTransport wraps the given transport with retry logic.
func newRetryTransport(base http.RoundTripper, maxRetries int) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base, maxRetries: maxRetries}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			// Network-level error: only retry if the request is idempotent or retryable
			if !isRetryableMethod(req.Method) || attempt == t.maxRetries {
				return resp, err
			}
			backoff := calcBackoff(attempt)
			time.Sleep(backoff)
			continue
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Don't retry if we've exhausted attempts
		if attempt == t.maxRetries {
			return resp, nil
		}

		// Use Retry-After header if present (GitHub sends it on 429)
		backoff := retryAfterDuration(resp)
		if backoff == 0 {
			backoff = calcBackoff(attempt)
		}

		// Drain and close response body before retry
		_ = resp.Body.Close()
		time.Sleep(backoff)
	}

	return resp, err
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || // 429
		status == http.StatusInternalServerError || // 500
		status == http.StatusBadGateway || // 502
		status == http.StatusServiceUnavailable || // 503
		status == http.StatusGatewayTimeout // 504
}

func isRetryableMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

// calcBackoff returns exponential backoff with jitter: base * 2^attempt + random jitter.
func calcBackoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	exp := time.Duration(math.Pow(2, float64(attempt))) * base
	if exp > 30*time.Second {
		exp = 30 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond))) //nolint:gosec
	return exp + jitter
}

// retryAfterDuration parses the Retry-After header if present.
func retryAfterDuration(resp *http.Response) time.Duration {
	ra := resp.Header.Get("Retry-After")
	if ra == "" {
		return 0
	}
	if secs, err := strconv.Atoi(ra); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}
````

## File: internal/sync/apply_registry_test.go
````go
package sync

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestHandlerRegistry_RegisterAndLookup(t *testing.T) {
	r := NewHandlerRegistry()
	want := errors.New("boom")
	r.Register("team", "create", 10, HandlerFunc(func(context.Context, *gh.Client, util.Change) error {
		return want
	}))

	h, ok := r.Lookup("team", "create")
	if !ok {
		t.Fatal("expected Lookup to find team:create")
	}
	if err := h.Apply(context.Background(), nil, util.Change{}); !errors.Is(err, want) {
		t.Errorf("expected registered handler to run, got %v", err)
	}

	if _, ok := r.Lookup("repo", "delete"); ok {
		t.Error("expected Lookup to miss unregistered keys")
	}
}

func TestHandlerRegistry_Precedence(t *testing.T) {
	r := NewHandlerRegistry()
	r.Register("team", "create", 10, HandlerFunc(noopHandler))
	r.Register("repo", "delete", 90, HandlerFunc(noopHandler))

	if got := r.Precedence("team", "create"); got != 10 {
		t.Errorf("team:create precedence = %d, want 10", got)
	}
	if got := r.Precedence("repo", "delete"); got != 90 {
		t.Errorf("repo:delete precedence = %d, want 90", got)
	}
	if got := r.Precedence("unknown", "action"); got != math.MaxInt {
		t.Errorf("unknown precedence = %d, want MaxInt (sort last)", got)
	}
}

func TestHandlerRegistry_DuplicatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	r := NewHandlerRegistry()
	r.Register("team", "create", 10, HandlerFunc(noopHandler))
	r.Register("team", "create", 20, HandlerFunc(noopHandler))
}

func TestDefaultRegistry_HasKnownKinds(t *testing.T) {
	kinds := []struct{ scope, action string }{
		{"team", "create"},
		{"team", "update"},
		{"team", "delete"},
		{"team-member", "ensure"},
		{"repo", "ensure"},
		{"repo", "delete"},
		{"team-repo", "grant"},
		{"repo-file", "ensure"},
		{"repo-topics", "ensure"},
		{"repo-template", "ensure"},
		{"repo-pin", "ensure"},
		{"org-member", "remove"},
		{"custom-role", "create"},
		{"custom-role", "update"},
		{"custom-role", "delete"},
	}
	for _, k := range kinds {
		if _, ok := defaultRegistry.Lookup(k.scope, k.action); !ok {
			t.Errorf("defaultRegistry missing handler for %s:%s", k.scope, k.action)
		}
	}
}

func TestDefaultRegistry_PrecedenceOrdering(t *testing.T) {
	// Create must run before member additions, which must run before deletions.
	if defaultRegistry.Precedence("team", "create") >= defaultRegistry.Precedence("team-member", "ensure") {
		t.Error("team:create should precede team-member:ensure")
	}
	if defaultRegistry.Precedence("team-member", "ensure") >= defaultRegistry.Precedence("team", "delete") {
		t.Error("team-member:ensure should precede team:delete")
	}
	if defaultRegistry.Precedence("repo", "ensure") >= defaultRegistry.Precedence("repo", "delete") {
		t.Error("repo:ensure should precede repo:delete")
	}
}

func noopHandler(context.Context, *gh.Client, util.Change) error { return nil }
````

## File: internal/sync/apply_registry.go
````go
package sync

import (
	"context"
	"fmt"
	"math"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// Handler applies a single planned change against GitHub.
//
// Implementations should be idempotent where possible; the planner already
// filters out no-op changes, but partial failures mid-apply are possible.
type Handler interface {
	Apply(ctx context.Context, c *gh.Client, ch util.Change) error
}

// HandlerFunc adapts an ordinary function to the Handler interface.
type HandlerFunc func(ctx context.Context, c *gh.Client, ch util.Change) error

// Apply implements Handler.
func (f HandlerFunc) Apply(ctx context.Context, c *gh.Client, ch util.Change) error {
	return f(ctx, c, ch)
}

// registration bundles a handler with the ordering precedence used during apply.
type registration struct {
	handler    Handler
	precedence int
}

// HandlerRegistry resolves change kinds (scope:action) to handlers and orders them.
type HandlerRegistry struct {
	entries map[string]registration
}

// NewHandlerRegistry returns an empty registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{entries: map[string]registration{}}
}

// Register binds a Handler to scope:action with the given precedence.
// Lower precedence runs first. Duplicate registrations panic so misconfigured
// registries fail fast at startup rather than during an apply.
func (r *HandlerRegistry) Register(scope, action string, precedence int, h Handler) {
	key := changeKey(scope, action)
	if _, exists := r.entries[key]; exists {
		panic(fmt.Sprintf("handler already registered for %s", key))
	}
	r.entries[key] = registration{handler: h, precedence: precedence}
}

// Lookup returns the Handler registered for scope:action, if any.
func (r *HandlerRegistry) Lookup(scope, action string) (Handler, bool) {
	e, ok := r.entries[changeKey(scope, action)]
	if !ok {
		return nil, false
	}
	return e.handler, true
}

// Precedence returns the ordering weight for scope:action. Unregistered keys
// sort last so unknown change kinds don't silently jump the queue.
func (r *HandlerRegistry) Precedence(scope, action string) int {
	if e, ok := r.entries[changeKey(scope, action)]; ok {
		return e.precedence
	}
	return math.MaxInt
}

func changeKey(scope, action string) string { return scope + ":" + action }

// defaultRegistry is the set of built-in handlers applyChanges consults.
var defaultRegistry = buildDefaultRegistry()

func buildDefaultRegistry() *HandlerRegistry {
	r := NewHandlerRegistry()

	// Creation / mutation phase (low precedence = runs first).
	r.Register("custom-role", "create", precedenceCustomRoleCreate, HandlerFunc(applyCustomRoleNoop))
	r.Register("custom-role", "update", precedenceCustomRoleUpdate, HandlerFunc(applyCustomRoleNoop))
	r.Register("team", "create", precedenceTeamCreate, HandlerFunc(applyTeamCreate))
	r.Register("repo", "ensure", precedenceRepoEnsure, HandlerFunc(applyRepoEnsure))
	r.Register("team", "update", precedenceTeamUpdate, HandlerFunc(applyTeamUpdate))
	r.Register("team-repo", "grant", precedenceTeamRepoGrant, HandlerFunc(applyTeamRepoGrant))
	r.Register("team-member", "ensure", precedenceTeamMemberEnsure, HandlerFunc(applyTeamMemberEnsure))
	r.Register("repo-file", "ensure", precedenceRepoFileEnsure, HandlerFunc(applyRepoFileEnsure))
	r.Register("repo-topics", "ensure", precedenceRepoTopicsEnsure, HandlerFunc(applyRepoTopicsEnsure))
	r.Register("repo-template", "ensure", precedenceRepoTemplateEnsure, HandlerFunc(applyRepoTemplateEnsure))
	r.Register("repo-pin", "ensure", precedenceRepoPinEnsure, HandlerFunc(applyRepoPinEnsure))

	// Cleanup phase (high precedence = runs last).
	r.Register("org-member", "remove", precedenceOrgMemberRemove, HandlerFunc(applyOrgMemberRemove))
	r.Register("team", "delete", precedenceTeamDelete, HandlerFunc(applyTeamDelete))
	r.Register("repo", "delete", precedenceRepoDelete, HandlerFunc(applyRepoDelete))
	r.Register("custom-role", "delete", precedenceCustomRoleDelete, HandlerFunc(applyCustomRoleNoop))

	return r
}

// applyCustomRoleNoop is a placeholder for custom-role changes; they are
// dispatched via applyCustomRoleChanges before the main loop runs, so this
// handler is only consulted for precedence ordering and should never execute.
func applyCustomRoleNoop(_ context.Context, _ *gh.Client, ch util.Change) error {
	return fmt.Errorf("custom-role change reached generic apply loop: %s:%s %s", ch.Scope, ch.Action, ch.Target)
}
````

## File: internal/sync/custom_roles_test.go
````go
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
````

## File: internal/templates/readme_test.go
````go
package templates

import (
	"strings"
	"testing"
)

func TestGenerateReadme(t *testing.T) {
	org := "TestOrg"
	repo := "test-repo"

	content, err := GenerateReadme(org, repo)
	if err != nil {
		t.Fatalf("GenerateReadme failed: %v", err)
	}

	// Verify content has expected elements
	expectedElements := []string{
		"# test-repo",
		"git clone git@github.com:TestOrg/test-repo.git",
		"git remote add origin git@github.com:TestOrg/test-repo.git",
		"Quick setup",
		"README, LICENSE, and .gitignore",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated README missing expected element: %q", expected)
		}
	}

	// Verify the repo name appears multiple times
	if strings.Count(content, repo) < 3 {
		t.Errorf("Expected repo name to appear at least 3 times, found %d", strings.Count(content, repo))
	}

	// Verify the org name appears multiple times
	if strings.Count(content, org) < 3 {
		t.Errorf("Expected org name to appear at least 3 times, found %d", strings.Count(content, org))
	}
}

func TestGenerateReadmeConsistency(t *testing.T) {
	// Test that generating the same README twice produces the same result
	org := "ConsistentOrg"
	repo := "consistent-repo"

	content1, err1 := GenerateReadme(org, repo)
	content2, err2 := GenerateReadme(org, repo)

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateReadme failed: %v, %v", err1, err2)
	}

	if content1 != content2 {
		t.Error("GenerateReadme should produce consistent results")
	}
}
````

## File: internal/templates/readme.go
````go
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
````

## File: internal/version/version_test.go
````go
package version

import "testing"

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()
	if info.Version == "" {
		t.Error("expected non-empty Version field")
	}
}
````

## File: .repomixignore
````
*.pem

# gomgr binary
/gomgr
/build/

# Test coverage
/coverage/

# Created by https://www.toptal.com/developers/gitignore/api/go,goland
# Edit at https://www.toptal.com/developers/gitignore?templates=go,goland

### Go ###
# If you prefer the allow list template instead of the deny list, see community template:
# https://github.com/github/gitignore/blob/main/community/Golang/Go.AllowList.gitignore
#
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

### GoLand ###
# Covers JetBrains IDEs: IntelliJ, RubyMine, PhpStorm, AppCode, PyCharm, CLion, Android Studio, WebStorm and Rider
# Reference: https://intellij-support.jetbrains.com/hc/en-us/articles/206544839

# User-specific stuff
.idea/**/workspace.xml
.idea/**/tasks.xml
.idea/**/usage.statistics.xml
.idea/**/dictionaries
.idea/**/shelf

# AWS User-specific
.idea/**/aws.xml

# Generated files
.idea/**/contentModel.xml

# Sensitive or high-churn files
.idea/**/dataSources/
.idea/**/dataSources.ids
.idea/**/dataSources.local.xml
.idea/**/sqlDataSources.xml
.idea/**/dynamic.xml
.idea/**/uiDesigner.xml
.idea/**/dbnavigator.xml

# Gradle
.idea/**/gradle.xml
.idea/**/libraries

# Gradle and Maven with auto-import
# When using Gradle or Maven with auto-import, you should exclude module files,
# since they will be recreated, and may cause churn.  Uncomment if using
# auto-import.
# .idea/artifacts
# .idea/compiler.xml
# .idea/jarRepositories.xml
# .idea/modules.xml
# .idea/*.iml
# .idea/modules
# *.iml
# *.ipr

# CMake
cmake-build-*/

# Mongo Explorer plugin
.idea/**/mongoSettings.xml

# File-based project format
*.iws

# IntelliJ
out/

# mpeltonen/sbt-idea plugin
.idea_modules/

# JIRA plugin
atlassian-ide-plugin.xml

# Cursive Clojure plugin
.idea/replstate.xml

# SonarLint plugin
.idea/sonarlint/

# Crashlytics plugin (for Android Studio and IntelliJ)
com_crashlytics_export_strings.xml
crashlytics.properties
crashlytics-build.properties
fabric.properties

# Editor-based Rest Client
.idea/httpRequests

# Android studio 3.1+ serialized cache file
.idea/caches/build_file_checksums.ser

### GoLand Patch ###
# Comment Reason: https://github.com/joeblau/gitignore.io/issues/186#issuecomment-215987721

# *.iml
# modules.xml
# .idea/misc.xml
# *.ipr

# Sonarlint plugin
# https://plugins.jetbrains.com/plugin/7973-sonarlint
.idea/**/sonarlint/

# SonarQube Plugin
# https://plugins.jetbrains.com/plugin/7238-sonarqube-community-plugin
.idea/**/sonarIssues.xml

# Markdown Navigator plugin
# https://plugins.jetbrains.com/plugin/7896-markdown-navigator-enhanced
.idea/**/markdown-navigator.xml
.idea/**/markdown-navigator-enh.xml
.idea/**/markdown-navigator/

# Cache file creation bug
# See https://youtrack.jetbrains.com/issue/JBR-2257
.idea/$CACHE_FILE$

# CodeStream plugin
# https://plugins.jetbrains.com/plugin/12206-codestream
.idea/codestream.xml

# Azure Toolkit for IntelliJ plugin
# https://plugins.jetbrains.com/plugin/8053-azure-toolkit-for-intellij
.idea/**/azureSettings.xml

# End of https://www.toptal.com/developers/gitignore/api/go,goland

.idea

.env
.pem
````

## File: LICENSE.md
````markdown
Apache License
                           Version 2.0, January 2025
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.

   END OF TERMS AND CONDITIONS

   APPENDIX: How to apply the Apache License to your work.

      To apply the Apache License to your work, attach the following
      boilerplate notice, with the fields enclosed by brackets "[]"
      replaced with your own identifying information. (Don't include
      the brackets!)  The text should be enclosed in the appropriate
      comment syntax for the file format. We also recommend that a
      file or class name and description of purpose be included on the
      same "printed page" as the copyright notice for easier
      identification within third-party archives.

   Copyright [yyyy] [name of copyright owner]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
````

## File: repomix.config.json
````json
{
  "$schema": "https://repomix.com/schemas/latest/schema.json",
  "input": {
    "maxFileSize": 52428800
  },
  "output": {
    "filePath": "repomix-output.md",
    "style": "markdown",
    "parsableStyle": false,
    "fileSummary": true,
    "directoryStructure": true,
    "files": true,
    "removeComments": false,
    "removeEmptyLines": false,
    "compress": false,
    "topFilesLength": 5,
    "showLineNumbers": false,
    "truncateBase64": false,
    "copyToClipboard": false,
    "includeFullDirectoryStructure": false,
    "tokenCountTree": false,
    "git": {
      "sortByChanges": true,
      "sortByChangesMaxCommits": 100,
      "includeDiffs": false,
      "includeLogs": false,
      "includeLogsCount": 50
    }
  },
  "include": [],
  "ignore": {
    "useGitignore": true,
    "useDotIgnore": true,
    "useDefaultPatterns": true,
    "customPatterns": []
  },
  "security": {
    "enableSecurityCheck": true
  },
  "tokenCount": {
    "encoding": "o200k_base"
  }
}
````

## File: cmd/validate.go
````go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration without connecting to GitHub",
	Example: `  gomgr validate -c ./config
  gomgr validate --config /path/to/config`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfgDir == "" {
			return fmt.Errorf("--config/-c flag is required")
		}
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		_ = cfg
		fmt.Println("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
````

## File: config/example/org.yaml
````yaml
owners:
  - octocat

# Custom repository roles (requires GitHub Enterprise Cloud)
# Define fine-grained permissions for specialized access patterns
custom_roles:
  - name: actions-manager
    description: Manage CI/CD workflows and runners
    base_role: read
    permissions:
      - write_actions
      - read_actions_variables
      - write_actions_variables

  - name: release-manager
    description: Manage releases and deployments
    base_role: write
    permissions:
      - create_releases
      - edit_releases
      - manage_environments
````

## File: examples/config/.github/workflows/sync.yaml
````yaml
name: Synchronise organization users and teams (gomgr)

on:
  push:
    branches: [ "main" ]
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-24.04
    continue-on-error: true
    strategy:
      max-parallel: 5
      fail-fast: false
      matrix:
        config:
          - { folder: "examples/example", gom_version: "v0.0.3" }

    env:
      GH_TOKEN: ${{ github.token }}   # for gh release download

    steps:
      - name: Checkout repository
        uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v6

      - name: Determine platform
        id: plat
        run: |
          echo "os=linux"   >> $GITHUB_OUTPUT
          echo "arch=amd64" >> $GITHUB_OUTPUT

      - name: Download gomgr binary from releases
        run: |
          VERSION="${{ matrix.config.gom_version }}"
          OS="${{ steps.plat.outputs.os }}"
          ARCH="${{ steps.plat.outputs.arch }}"
          ASSET_TGZ="gomgr_${VERSION}_${OS}_${ARCH}.tar.gz"
          ASSET_ZIP="gomgr_${VERSION}_${OS}_${ARCH}.zip"

          mkdir -p .gomgr
          gh release download "$VERSION" --repo DragonSecurity/gomgr --pattern "$ASSET_TGZ"                 --dir .gomgr || true

          if [ ! -f ".gomgr/$ASSET_TGZ" ]; then
            gh release download "$VERSION" --repo DragonSecurity/gomgr --pattern "$ASSET_ZIP"                   --dir .gomgr
          fi

          if [ -f ".gomgr/$ASSET_TGZ" ]; then
            tar -xzf ".gomgr/$ASSET_TGZ" -C .gomgr
          else
            unzip -o ".gomgr/$ASSET_ZIP" -d .gomgr
          fi

          GOMGR_PATH=$(find .gomgr -type f -name "gomgr" -o -name "gomgr.exe" | head -n1)
          sudo mv "$GOMGR_PATH" /usr/local/bin/gomgr
          sudo chmod +x /usr/local/bin/gomgr

      - name: Show gomgr version
        run: gomgr version

      # - name: Dry-run sync
      #   run: gomgr sync -c ${{ matrix.examples.folder }} --dry

      - name: Synchronise settings
        run: gomgr sync -c ${{ matrix.config.folder }}
        env:
          GITHUB_APP_PRIVATE_KEY: ${{ secrets.DSEC_USER_MANAGEMENT_APP_PRIVATE_KEY }}
          GITHUB_APP_ID: "1719369"
````

## File: examples/config/teams/example-team.yaml
````yaml
name: Platform Team
description: Core platform engineers
privacy: closed
parents: []
maintainers:
  - allanice001
repositories:
  infra: maintain
  
  # Template repository - can be used by other repos
  template-go-api:
    permission: push
    template: true
    topics:
      - backend
      - api
      - go-template
  
  # Repository using simple config with topics
  api: 
    permission: push
    topics:
      - backend
      - api
      - project-platform
  
  # Repository using template (inherits topics and permission)
  my-api:
    from: template-go-api
    topics:
      - my-project
      # Will inherit: backend, api, go-template from template
  
  # Repository using template with permission override
  admin-api:
    from: template-go-api
    permission: admin
    topics:
      - admin-service

  platform-index:
    permission: admin
    topics:
      - project-platform
      - documentation
    pinned: true
````

## File: examples/config/teams/github-actions-team.yaml
````yaml
name: GitHub Actions Team
slug: github-actions-team
description: Team managing CI/CD workflows with custom repository roles for GitHub Actions
privacy: closed

# Note: This example demonstrates GitHub's custom repository roles feature
# Custom roles must be defined in org.yaml and gomgr will create/update them automatically
# See: https://docs.github.com/en/enterprise-cloud@latest/organizations/managing-user-access-to-your-organizations-repositories/managing-repository-roles/managing-custom-repository-roles-for-an-organization

maintainers:
  - derek-cicd-lead
  - emily-actions-expert

members:
  - frank-workflow-engineer
  - gina-automation-dev
  - hans-release-manager

repositories:
  # Standard permission levels (built-in roles)
  # These work for all GitHub plans
  backend-api:
    permission: pull
    topics:
      - cicd
      - backend
  
  web-app:
    permission: pull
    topics:
      - cicd
      - frontend
  
  # Custom repository role examples
  # These require GitHub Enterprise Cloud and must be defined in org.yaml
  # gomgr will create/update these roles automatically
  # Custom roles allow fine-grained permissions like:
  # - Managing GitHub Actions without full repository admin access
  # - Managing environments and secrets
  # - Managing runners
  # - Configuring Actions settings
  
  # Example 1: Custom role for managing workflows without code access
  # This role might have permissions to:
  # - Edit workflow files
  # - Manage Actions secrets and variables
  # - Manage self-hosted runners
  # - Cancel and re-run workflow runs
  ci-workflows:
    permission: actions-manager  # Custom role name (must be created in GitHub org)
    topics:
      - cicd
      - github-actions
      - workflows
  
  # Example 2: Custom role for release management
  # This role might have permissions to:
  # - Create and edit releases
  # - Manage deployment environments
  # - Manage environment secrets
  # - Trigger deployments
  release-automation:
    permission: release-manager  # Custom role name (must be created in GitHub org)
    topics:
      - cicd
      - releases
      - deployments
  
  # Example 3: Custom role for runner management
  # This role might have permissions to:
  # - Manage self-hosted runners
  # - View workflow runs
  # - Access runner logs
  runner-infrastructure:
    permission: runner-admin  # Custom role name (must be created in GitHub org)
    topics:
      - cicd
      - runners
      - infrastructure
  
  # Example 4: Custom role for security scanning
  # This role might have permissions to:
  # - Manage code scanning alerts
  # - Configure secret scanning
  # - Manage Dependabot settings
  security-scanning:
    permission: security-scanner  # Custom role name (must be created in GitHub org)
    topics:
      - security
      - scanning
      - cicd
````

## File: examples/config/org.yaml
````yaml
owners:
  - allanice001

# Custom repository roles (requires GitHub Enterprise Cloud)
# These roles will be created/updated automatically by gomgr
custom_roles:
  - name: actions-manager
    description: Manage GitHub Actions workflows, runners, and secrets without code access
    base_role: read
    permissions:
      - write_actions
      - read_actions_variables
      - write_actions_variables
      - read_organization_secrets
      - write_organization_secrets

  - name: release-manager
    description: Create and manage releases and deployment environments
    base_role: write
    permissions:
      - create_releases
      - edit_releases
      - delete_releases
      - manage_environments

  - name: runner-admin
    description: Manage self-hosted runners for CI/CD infrastructure
    base_role: read
    permissions:
      - admin_self_hosted_runners
      - read_self_hosted_runners

  - name: security-scanner
    description: Configure security scanning without repository admin access
    base_role: read
    permissions:
      - read_code_scanning_alerts
      - write_code_scanning_alerts
      - read_secret_scanning_alerts
      - write_secret_scanning_alerts
````

## File: internal/gh/client_test.go
````go
package gh

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeReadPEM_InlineKey(t *testing.T) {
	inline := "-----BEGIN RSA PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END RSA PRIVATE KEY-----\n"
	b, err := maybeReadPEM(inline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != inline {
		t.Errorf("got %q, want inline key back", string(b))
	}
}

func TestMaybeReadPEM_InlineGarbage(t *testing.T) {
	_, err := maybeReadPEM("BEGIN but not a real pem")
	if err == nil {
		t.Fatal("expected error for malformed inline key")
	}
	if !strings.Contains(err.Error(), "invalid PEM") {
		t.Errorf("expected 'invalid PEM' in error, got: %v", err)
	}
}

func TestMaybeReadPEM_WrongBlockType(t *testing.T) {
	cert := "-----BEGIN CERTIFICATE-----\nMIIBogIBAAJBALRiMLAH\n-----END CERTIFICATE-----\n"
	_, err := maybeReadPEM(cert)
	if err == nil {
		t.Fatal("expected error for non-private-key block")
	}
	if !strings.Contains(err.Error(), "expected a private key block") {
		t.Errorf("expected block-type error, got: %v", err)
	}
}

func TestMaybeReadPEM_PKCS8(t *testing.T) {
	pkcs8 := "-----BEGIN PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END PRIVATE KEY-----\n"
	if _, err := maybeReadPEM(pkcs8); err != nil {
		t.Fatalf("unexpected error for PKCS#8 block: %v", err)
	}
}

func TestMaybeReadPEM_FromFile(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nMIIBogIBAAJBALRiMLAH\n-----END RSA PRIVATE KEY-----\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(path, []byte(pem), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := maybeReadPEM(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != pem {
		t.Errorf("got %q, want file contents", string(b))
	}
}

func TestMaybeReadPEM_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(path, []byte("not a pem"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := maybeReadPEM(path)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
	if !strings.Contains(err.Error(), "invalid PEM") {
		t.Errorf("expected 'invalid PEM' in error, got: %v", err)
	}
}

func TestMaybeReadPEM_MissingFile(t *testing.T) {
	_, err := maybeReadPEM("/nonexistent/path/key.pem")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func newGraphQLClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		httpClient: srv.Client(),
		GraphQLURL: srv.URL,
	}, srv
}

func TestDoGraphQL_Success(t *testing.T) {
	var gotBody map[string]any
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type=application/json, got %q", ct)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"viewer":{"login":"octocat"}}}`))
	})

	var out struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}
	err := c.DoGraphQL(context.Background(), "query { viewer { login } }", map[string]any{"x": 1}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Viewer.Login != "octocat" {
		t.Errorf("expected login=octocat, got %q", out.Viewer.Login)
	}
	if gotBody["query"] == nil {
		t.Error("expected request body to include query")
	}
	if gotBody["variables"] == nil {
		t.Error("expected request body to include variables")
	}
}

func TestDoGraphQL_OmitsEmptyVariables(t *testing.T) {
	var gotBody map[string]any
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})

	if err := c.DoGraphQL(context.Background(), "query { viewer { login } }", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := gotBody["variables"]; present {
		t.Error("expected variables to be omitted when empty")
	}
}

func TestDoGraphQL_HTTPError(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "status 401") {
		t.Errorf("expected 'status 401' in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Bad credentials") {
		t.Errorf("expected upstream message to be surfaced, got %v", err)
	}
}

func TestDoGraphQL_GraphQLErrors(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errors":[{"message":"field missing"},{"message":"another issue"}]}`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for graphql errors")
	}
	if !strings.Contains(err.Error(), "field missing") || !strings.Contains(err.Error(), "another issue") {
		t.Errorf("expected both error messages to be joined, got %v", err)
	}
}

func TestDoGraphQL_MalformedJSON(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": not-json`))
	})

	err := c.DoGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "decode graphql response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestDoGraphQL_NilResultSkipsDataUnmarshal(t *testing.T) {
	c, _ := newGraphQLClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"anything":123}}`))
	})

	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err != nil {
		t.Errorf("expected nil error when result is nil, got %v", err)
	}
}

func TestDoGraphQL_EmptyQueryRejected(t *testing.T) {
	c := &Client{httpClient: &http.Client{}}
	if err := c.DoGraphQL(context.Background(), "   ", nil, nil); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestDoGraphQL_NilClientRejected(t *testing.T) {
	var c *Client
	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestDoGraphQL_NilHTTPClientRejected(t *testing.T) {
	c := &Client{}
	if err := c.DoGraphQL(context.Background(), "query {}", nil, nil); err == nil {
		t.Fatal("expected error for nil httpClient")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"hello", "world", "hello"},
		{"", "world", "world"},
		{"", "", ""},
		{"a", "", "a"},
	}
	for _, tt := range tests {
		got := firstNonEmpty(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("firstNonEmpty(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}
````

## File: internal/gh/rate_test.go
````go
package gh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v84/github"
)

func TestRespectRate_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"resources": map[string]any{
				"core": map[string]any{
					"limit":     5000,
					"remaining": 4999,
					"reset":     time.Now().Add(time.Hour).Unix(),
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)

	err := RespectRate(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRespectRate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "server error"})
	}))
	defer server.Close()

	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)

	err := RespectRate(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "rate limit check") {
		t.Errorf("expected 'rate limit check' in error, got: %v", err)
	}
}
````

## File: internal/sync/apply_handlers_test.go
````go
package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// newTestClient creates a gh.Client backed by the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *gh.Client {
	t.Helper()
	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)
	return &gh.Client{REST: client}
}

func TestApplyTeamCreate(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/orgs/myorg/teams" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "slug": "backend"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "backend",
		Action: "create",
		Details: map[string]any{
			"org":         "myorg",
			"name":        "Backend",
			"privacy":     "closed",
			"description": "Backend team",
		},
	}

	err := applyTeamCreate(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "Backend" {
		t.Errorf("expected name=Backend, got %v", gotBody["name"])
	}
}

func TestApplyTeamDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/teams/old-team" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "old-team",
		Action: "delete",
		Details: map[string]any{
			"org":  "myorg",
			"slug": "old-team",
		},
	}

	err := applyTeamDelete(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyRepoEnsure(t *testing.T) {
	t.Run("regular repo", func(t *testing.T) {
		var gotBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/orgs/myorg/repos" {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "api"})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope:  "repo",
			Target: "api",
			Action: "ensure",
			Details: map[string]any{
				"org":     "myorg",
				"name":    "api",
				"private": true,
			},
		}

		err := applyRepoEnsure(context.Background(), c, ch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotBody["name"] != "api" {
			t.Errorf("expected name=api, got %v", gotBody["name"])
		}
	})

	t.Run("from template", func(t *testing.T) {
		created := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/repos/myorg/template-go/generate" {
				created = true
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 2, "name": "new-api"})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope:  "repo",
			Target: "new-api",
			Action: "ensure",
			Details: map[string]any{
				"org":     "myorg",
				"name":    "new-api",
				"private": true,
				"from":    "template-go",
			},
		}

		err := applyRepoEnsure(context.Background(), c, ch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected template creation request to be made")
		}
	})
}

func TestApplyRepoFileEnsure_RaceCondition(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errBody    map[string]any
		wantErr    bool
	}{
		{
			name:       "422 sha not supplied (race condition)",
			statusCode: 422,
			errBody: map[string]any{
				"message": `"sha" wasn't supplied`,
			},
			wantErr: false,
		},
		{
			name:       "409 reference already exists (race condition)",
			statusCode: 409,
			errBody: map[string]any{
				"message": "reference already exists",
			},
			wantErr: false,
		},
		{
			name:       "422 unrelated error",
			statusCode: 422,
			errBody: map[string]any{
				"message": "Validation Failed",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// GetContents returns 404 (file not found)
				if r.Method == "GET" {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
					return
				}
				// CreateFile returns the race condition error
				if r.Method == "PUT" {
					w.WriteHeader(tt.statusCode)
					_ = json.NewEncoder(w).Encode(tt.errBody)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c := newTestClient(t, server)
			ch := util.Change{
				Scope:  "repo-file",
				Target: "api:README.md",
				Action: "ensure",
				Details: map[string]any{
					"org":     "myorg",
					"repo":    "api",
					"path":    "README.md",
					"content": "# API",
					"message": "add readme",
					"branch":  "main",
				},
			}

			err := applyRepoFileEnsure(context.Background(), c, ch)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRepoFileEnsure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyRepoTopicsEnsure(t *testing.T) {
	tests := []struct {
		name   string
		topics any
	}{
		{
			name:   "string slice",
			topics: []string{"backend", "api"},
		},
		{
			name:   "any slice",
			topics: []any{"backend", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "PUT" && r.URL.Path == "/repos/myorg/api/topics" {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{"names": []string{"backend", "api"}})
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c := newTestClient(t, server)
			ch := util.Change{
				Scope:  "repo-topics",
				Target: "api",
				Action: "ensure",
				Details: map[string]any{
					"org":    "myorg",
					"repo":   "api",
					"topics": tt.topics,
				},
			}

			err := applyRepoTopicsEnsure(context.Background(), c, ch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestApplyHandlers_InvalidDetails(t *testing.T) {
	handlers := map[string]func(context.Context, *gh.Client, util.Change) error{
		"team:create":          applyTeamCreate,
		"team:update":          applyTeamUpdate,
		"team:delete":          applyTeamDelete,
		"repo:ensure":          applyRepoEnsure,
		"team-repo:grant":      applyTeamRepoGrant,
		"repo-file:ensure":     applyRepoFileEnsure,
		"repo-topics:ensure":   applyRepoTopicsEnsure,
		"repo-template:ensure": applyRepoTemplateEnsure,
		"repo-pin:ensure":      applyRepoPinEnsure,
		"repo:delete":          applyRepoDelete,
	}

	for key, handler := range handlers {
		t.Run(key, func(t *testing.T) {
			ch := util.Change{
				Scope:   key[:4], // doesn't matter much
				Target:  "test",
				Action:  "test",
				Details: "not-a-map", // wrong type
			}
			err := handler(context.Background(), nil, ch)
			if err == nil {
				t.Errorf("expected error for invalid details type, got nil")
			}
			if err != nil && !containsSubstr(err.Error(), "invalid details") {
				t.Errorf("expected 'invalid details' in error, got: %v", err)
			}
		})
	}

	// Test team-member:ensure separately (expects teamMemberChange, not map)
	t.Run("team-member:ensure", func(t *testing.T) {
		ch := util.Change{
			Scope:   "team-member",
			Target:  "test",
			Action:  "ensure",
			Details: "not-a-struct",
		}
		err := applyTeamMemberEnsure(context.Background(), nil, ch)
		if err == nil {
			t.Error("expected error for invalid details type, got nil")
		}
	})
}

func containsSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestApplyTeamUpdate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/orgs/myorg/teams/backend" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "slug": "backend"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "backend",
		Action: "update",
		Details: map[string]any{
			"org":         "myorg",
			"slug":        "backend",
			"name":        "Backend",
			"description": "Updated description",
			"privacy":     "secret",
		},
	}

	err := applyTeamUpdate(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "Backend" {
		t.Errorf("expected name=Backend, got %v", gotBody["name"])
	}
	if gotBody["description"] != "Updated description" {
		t.Errorf("expected description='Updated description', got %v", gotBody["description"])
	}
	if gotBody["privacy"] != "secret" {
		t.Errorf("expected privacy=secret, got %v", gotBody["privacy"])
	}
}

func TestApplyTeamMemberEnsure(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/orgs/myorg/teams/backend/memberships/alice" {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"state": "active", "role": "member"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:   "team-member",
		Target:  "backend",
		Action:  "ensure",
		Details: teamMemberChange{Org: "myorg", Slug: "backend", User: "alice", Role: "member"},
	}

	err := applyTeamMemberEnsure(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/orgs/myorg/teams/backend/memberships/alice" {
		t.Errorf("expected PUT to memberships path, got %s", gotPath)
	}
}

func TestApplyRepoTemplateEnsure(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/repos/myorg/api" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "api", "is_template": true})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-template",
		Target: "api",
		Action: "ensure",
		Details: map[string]any{
			"org":      "myorg",
			"repo":     "api",
			"template": true,
		},
	}

	err := applyRepoTemplateEnsure(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["is_template"] != true {
		t.Errorf("expected is_template=true, got %v", gotBody["is_template"])
	}
}

func TestApplyRepoPinEnsure(t *testing.T) {
	ch := util.Change{
		Scope:  "repo-pin",
		Target: "api",
		Action: "ensure",
		Details: map[string]any{
			"org":    "myorg",
			"repo":   "api",
			"pinned": true,
		},
	}

	err := applyRepoPinEnsure(context.Background(), nil, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyRepoDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/repos/myorg/old-repo" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo",
		Target: "old-repo",
		Action: "delete",
		Details: map[string]any{
			"org":  "myorg",
			"repo": "old-repo",
		},
	}

	err := applyRepoDelete(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyTeamRepoGrant(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/orgs/myorg/teams/backend/repos/myorg/api" {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team-repo",
		Target: "backend/api",
		Action: "grant",
		Details: map[string]any{
			"org":        "myorg",
			"slug":       "backend",
			"repo":       "api",
			"permission": "push",
		},
	}

	err := applyTeamRepoGrant(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/orgs/myorg/teams/backend/repos/myorg/api" {
		t.Errorf("expected PUT to team repos path, got %s", gotPath)
	}
}

func TestApplyOrgMemberRemove(t *testing.T) {
	removed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/memberships/bob" {
			removed = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "org-member",
		Target: "bob",
		Action: "remove",
		Details: map[string]any{
			"org":  "myorg",
			"user": "bob",
		},
	}

	err := applyOrgMemberRemove(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removed {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyOrgMemberRemove_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "org-member",
		Target: "ghost",
		Action: "remove",
		Details: map[string]any{
			"org":  "myorg",
			"user": "ghost",
		},
	}

	err := applyOrgMemberRemove(context.Background(), c, ch)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !containsSubstr(err.Error(), "remove member") {
		t.Errorf("expected 'remove member' in error, got: %v", err)
	}
}

func TestApplyOrgMemberRemove_InvalidDetails(t *testing.T) {
	ch := util.Change{
		Scope:   "org-member",
		Target:  "test",
		Action:  "remove",
		Details: "not-a-map",
	}
	err := applyOrgMemberRemove(context.Background(), nil, ch)
	if err == nil {
		t.Error("expected error for invalid details type")
	}
}

func TestApplyCustomRoleCreate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/orgs/myorg/custom-repository-roles" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "deployer"})
			return
		}
		// Rate limit endpoint
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "create",
			Details: customRoleChange{
				Org:         "myorg",
				Name:        "deployer",
				Description: "Deploy role",
				BaseRole:    "read",
				Permissions: []string{"manage_actions"},
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "deployer" {
		t.Errorf("expected name=deployer, got %v", gotBody["name"])
	}
}

func TestApplyCustomRoleUpdate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/orgs/myorg/custom-repository-roles/42" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42, "name": "deployer"})
			return
		}
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "update",
			Details: customRoleChange{
				Org:         "myorg",
				ID:          42,
				Name:        "deployer",
				Description: "Updated desc",
				BaseRole:    "write",
				Permissions: []string{"manage_actions", "create_releases"},
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "deployer" {
		t.Errorf("expected name=deployer, got %v", gotBody["name"])
	}
}

func TestApplyCustomRoleDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/custom-repository-roles/42" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "delete",
			Details: customRoleChange{
				Org:  "myorg",
				ID:   42,
				Name: "deployer",
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyCustomRoleChanges_SkipsNonCustomRole(t *testing.T) {
	changes := []util.Change{
		{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg"}},
	}
	// Should not error - just skip non-custom-role changes
	err := applyCustomRoleChanges(context.Background(), nil, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlanCustomRoleCleanups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/custom-repository-roles" && r.Method == "GET" {
			resp := map[string]any{
				"total_count": 2,
				"custom_roles": []map[string]any{
					{"id": 1, "name": "deployer"},
					{"id": 2, "name": "stale-role"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{
			Org:                        "myorg",
			DeleteUnmanagedCustomRoles: true,
		},
		Org: config.OrgConfig{
			CustomRoles: []config.CustomRoleConfig{
				{Name: "deployer", BaseRole: "read"},
			},
		},
	}
	st := &State{Org: "myorg"}

	changes, warnings, err := planCustomRoleCleanups(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
	// Should have a delete change for "stale-role"
	found := false
	for _, ch := range changes {
		if ch.Scope == "custom-role" && ch.Action == "delete" && ch.Target == "stale-role" {
			found = true
		}
	}
	if !found {
		t.Error("expected custom-role:delete change for stale-role")
	}
}
````

## File: internal/sync/apply_handlers.go
````go
package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// extractDetails performs a safe type assertion on ch.Details to map[string]any.
func extractDetails(ch util.Change) (map[string]any, error) {
	d, ok := ch.Details.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid details for %s:%s: expected map[string]any, got %T", ch.Scope, ch.Action, ch.Details)
	}
	return d, nil
}

func detailString(d map[string]any, key string) string {
	if v, ok := d[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	return ""
}

func detailBool(d map[string]any, key string) bool {
	if v, ok := d[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
		return fmt.Sprint(v) == "true"
	}
	return false
}

func applyTeamCreate(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	name := detailString(d, "name")
	var privacyPtr, descPtr *string
	if pv := detailString(d, "privacy"); pv != "" {
		privacyPtr = github.Ptr(pv)
	}
	if dv := detailString(d, "description"); dv != "" {
		descPtr = github.Ptr(dv)
	}
	newTeam := github.NewTeam{Name: name, Privacy: privacyPtr, Description: descPtr}
	_, _, err = c.REST.Teams.CreateTeam(ctx, org, newTeam)
	if err != nil {
		return fmt.Errorf("create team %q: %w", name, err)
	}
	return nil
}

func applyTeamUpdate(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	name := detailString(d, "name")
	newTeam := github.NewTeam{Name: name}
	if dv := detailString(d, "description"); dv != "" {
		newTeam.Description = github.Ptr(dv)
	}
	if pv := detailString(d, "privacy"); pv != "" {
		newTeam.Privacy = github.Ptr(pv)
	}
	_, _, err = c.REST.Teams.EditTeamBySlug(ctx, org, slug, newTeam, false)
	if err != nil {
		return fmt.Errorf("update team %q: %w", slug, err)
	}
	return nil
}

func applyTeamDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	_, err = c.REST.Teams.DeleteTeamBySlug(ctx, org, slug)
	if err != nil {
		return fmt.Errorf("delete team %q in org %q: %w", slug, org, err)
	}
	return nil
}

func applyTeamMemberEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, ok := ch.Details.(teamMemberChange)
	if !ok {
		return fmt.Errorf("invalid details for team-member:ensure: expected teamMemberChange, got %T", ch.Details)
	}
	_, _, err := c.REST.Teams.AddTeamMembershipBySlug(ctx, d.Org, d.Slug, d.User, &github.TeamAddTeamMembershipOptions{Role: d.Role})
	if err != nil {
		return fmt.Errorf("add %q as %q to %q in org %q: %w", d.User, d.Role, d.Slug, d.Org, err)
	}
	return nil
}

func applyRepoEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	name := detailString(d, "name")
	private := true
	if v, ok := d["private"]; ok {
		if b, isBool := v.(bool); isBool {
			private = b
		} else {
			private = fmt.Sprint(v) != "false"
		}
	}
	isTemplate := detailBool(d, "template")

	// Check if this repo should be created from a template
	if templateRef := detailString(d, "from"); templateRef != "" {
		templateOrg, templateRepo := parseTemplateRef(templateRef, org)

		// Create repository from template
		_, _, err := c.REST.Repositories.CreateFromTemplate(ctx, templateOrg, templateRepo, &github.TemplateRepoRequest{
			Name:    github.Ptr(name),
			Owner:   github.Ptr(org),
			Private: github.Ptr(private),
		})
		if err != nil {
			var ghErr *github.ErrorResponse
			if !errors.As(err, &ghErr) || ghErr.Response == nil || ghErr.Response.StatusCode != 422 {
				return fmt.Errorf("create repo %s/%s from template %s/%s: %w", org, name, templateOrg, templateRepo, err)
			}
			// already exists race — ignore
		}
	} else {
		// Create regular repository
		_, _, err := c.REST.Repositories.Create(ctx, org, &github.Repository{
			Name:                github.Ptr(name),
			Private:             github.Ptr(private),
			IsTemplate:          github.Ptr(isTemplate),
			AllowAutoMerge:      github.Ptr(true),
			AllowMergeCommit:    github.Ptr(false),
			DeleteBranchOnMerge: github.Ptr(true),
			HasIssues:           github.Ptr(true),
		})
		if err != nil {
			var ghErr *github.ErrorResponse
			if !errors.As(err, &ghErr) || ghErr.Response == nil || ghErr.Response.StatusCode != 422 {
				return fmt.Errorf("create repo %s/%s: %w", org, name, err)
			}
			// already exists race — ignore
		}
	}
	return nil
}

func applyTeamRepoGrant(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	repo := detailString(d, "repo")
	perm := normalizePermission(detailString(d, "permission"))
	_, err = c.REST.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repo, &github.TeamAddTeamRepoOptions{Permission: perm})
	if err != nil {
		return fmt.Errorf("grant %q on %s/%s to %q: %w", perm, org, repo, slug, err)
	}
	return nil
}

func applyRepoFileEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	path := detailString(d, "path")
	content := []byte(detailString(d, "content"))
	message := detailString(d, "message")
	branch := detailString(d, "branch")
	file, _, resp, err := c.REST.Repositories.GetContents(ctx, org, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		return fmt.Errorf("check %s/%s:%s: %w", org, repo, path, err)
	}
	if file == nil {
		_, _, err := c.REST.Repositories.CreateFile(ctx, org, repo, path, &github.RepositoryContentFileOptions{
			Message: github.Ptr(message),
			Content: content,
			Branch:  github.Ptr(branch),
		})
		if err != nil {
			// Handle race condition: If repository was created from template,
			// files may exist even though GetContents returned nil.
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response != nil {
				isRaceCondition := (ghErr.Response.StatusCode == 422 && containsErrorMessage(ghErr, errTermSHA, errTermSHANotSupplied)) ||
					(ghErr.Response.StatusCode == 409 && containsErrorMessage(ghErr, errTermRefExists))

				if !isRaceCondition {
					return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
				}
				// File already exists (likely from template), which is what we want - skip error
			} else {
				return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
			}
		}
	}
	return nil
}

func applyRepoTopicsEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	// Handle topics - may come as []string or []any from planning
	var topicsRaw []string
	if v, ok := d["topics"]; ok {
		switch topics := v.(type) {
		case []string:
			topicsRaw = topics
		case []any:
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					topicsRaw = append(topicsRaw, tStr)
				}
			}
		default:
			return fmt.Errorf("invalid type for topics for %s/%s: %T", org, repo, v)
		}
	}

	_, _, err = c.REST.Repositories.ReplaceAllTopics(ctx, org, repo, topicsRaw)
	if err != nil {
		return fmt.Errorf("set topics on %s/%s: %w", org, repo, err)
	}
	return nil
}

func applyRepoTemplateEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	_, _, err = c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
		IsTemplate: github.Ptr(true),
	})
	if err != nil {
		return fmt.Errorf("mark repo %s/%s as template: %w", org, repo, err)
	}
	return nil
}

func applyRepoPinEnsure(_ context.Context, _ *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	util.Warnf("Skipping pin for %s/%s: GitHub API does not support pinning to organization profiles", org, repo)
	return nil
}

func applyRepoDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	_, err = c.REST.Repositories.Delete(ctx, org, repo)
	if err != nil {
		return fmt.Errorf("delete repo %q in org %q: %w", repo, org, err)
	}
	return nil
}

func applyOrgMemberRemove(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	user := detailString(d, "user")
	_, err = c.REST.Organizations.RemoveOrgMembership(ctx, user, org)
	if err != nil {
		return fmt.Errorf("remove member %q from org %q: %w", user, org, err)
	}
	return nil
}

func normalizePermission(p string) string {
	switch strings.ToLower(p) {
	case "read", permPull:
		return permPull
	case permTriage:
		return permTriage
	case "write", permPush:
		return permPush
	case permMaintain:
		return permMaintain
	case permAdmin:
		return permAdmin
	default:
		return p
	}
}
````

## File: internal/sync/orchestrator_test.go
````go
package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestBuildPlan_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return empty lists for all endpoints
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
	}

	plan, err := BuildPlan(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(plan.Changes))
	}
	if plan.Stats == nil {
		t.Fatal("expected stats to be populated")
	}
	if plan.Stats.Teams.Current != 0 || plan.Stats.Teams.Desired != 0 {
		t.Errorf("expected 0/0 teams, got %d/%d", plan.Stats.Teams.Current, plan.Stats.Teams.Desired)
	}
}

func TestBuildPlan_WithTeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/teams" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case r.URL.Path == "/orgs/myorg/repos" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "Backend", Slug: "backend"},
		},
	}

	plan, err := BuildPlan(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Stats.Teams.Desired != 1 {
		t.Errorf("expected 1 desired team, got %d", plan.Stats.Teams.Desired)
	}
	// Should have at least a team:create change
	found := false
	for _, ch := range plan.Changes {
		if ch.Scope == "team" && ch.Action == "create" {
			found = true
		}
	}
	if !found {
		t.Error("expected a team:create change")
	}
}

func TestApply_Empty(t *testing.T) {
	plan := util.Plan{Changes: []util.Change{}}
	err := Apply(context.Background(), nil, plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	plan := util.Plan{
		Changes: []util.Change{
			{Scope: "team", Target: "test", Action: "create", Details: map[string]any{"org": "myorg", "name": "test"}},
		},
	}
	err := Apply(ctx, nil, plan)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
````

## File: internal/util/diff_test.go
````go
package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintPlan(t *testing.T) {
	plan := Plan{
		Changes: []Change{
			{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg"}},
		},
		Warnings: []string{"test warning"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintPlan(plan)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify it's valid JSON
	var parsed Plan
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, output)
	}
	if len(parsed.Changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(parsed.Changes))
	}
}

func TestPrintStatePair(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		pair     StatePair
		contains string
		empty    bool
	}{
		{name: "increase", label: "Teams:", pair: StatePair{Current: 2, Desired: 5}, contains: "(+3)"},
		{name: "decrease", label: "Teams:", pair: StatePair{Current: 5, Desired: 2}, contains: "(-3)"},
		{name: "no change", label: "Teams:", pair: StatePair{Current: 3, Desired: 3}, contains: "(no change)"},
		{name: "both zero", label: "Teams:", pair: StatePair{Current: 0, Desired: 0}, empty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printStatePair(tt.label, tt.pair)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.empty {
				if output != "" {
					t.Errorf("expected no output for zero pair, got %q", output)
				}
				return
			}
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, output)
			}
		})
	}
}

func TestPrintSummary(t *testing.T) {
	tests := []struct {
		name             string
		plan             Plan
		expectedContains []string
	}{
		{
			name: "with changes and warnings",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
					{Scope: "team", Target: "team2", Action: "create", Details: nil},
					{Scope: "team-member", Target: "user1", Action: "ensure", Details: nil},
					{Scope: "team-repo", Target: "repo1", Action: "grant", Details: nil},
					{Scope: "repo-pin", Target: "repo1", Action: "ensure", Details: nil},
				},
				Warnings: []string{"Test warning 1", "Test warning 2"},
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Total changes: 5",
				"Changes by scope:",
				"team:",
				"team-member:",
				"team-repo:",
				"repo-pin:",
				"Changes by action:",
				"create:",
				"ensure:",
				"grant:",
				"Warnings: 2",
				"Test warning 1",
				"Test warning 2",
			},
		},
		{
			name: "no changes",
			plan: Plan{
				Changes:  []Change{},
				Warnings: nil,
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"No changes required - configuration is in sync",
			},
		},
		{
			name: "changes without warnings",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
				},
				Warnings: nil,
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Total changes: 1",
				"Changes by scope:",
				"team:",
				"Changes by action:",
				"create:",
			},
		},
		{
			name: "with state statistics",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
					{Scope: "team-member", Target: "user1", Action: "ensure", Details: nil},
				},
				Warnings: nil,
				Stats: &StateStats{
					Teams: StatePair{
						Current: 2,
						Desired: 3,
					},
					TeamMembers: StatePair{
						Current: 5,
						Desired: 7,
					},
					Repositories: StatePair{
						Current: 10,
						Desired: 12,
					},
					RepoPermissions: StatePair{
						Current: 15,
						Desired: 18,
					},
				},
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Current State vs Desired State:",
				"Teams:",
				"2 → 3",
				"Team Members:",
				"5 → 7",
				"Repositories:",
				"10 → 12",
				"Repo Permissions:",
				"15 → 18",
				"Total changes: 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			PrintSummary(tt.plan)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r) // explicitly discard error
			output := buf.String()

			// Verify expected strings are in output
			for _, expected := range tt.expectedContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
				}
			}
		})
	}
}
````

## File: internal/util/log_test.go
````go
package util

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestWarnf(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Warnf("something %s happened", "bad")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "WARNING:") {
		t.Errorf("expected WARNING: prefix, got %q", output)
	}
	if !strings.Contains(output, "something bad happened") {
		t.Errorf("expected formatted message, got %q", output)
	}
}

func TestAudit_Enabled(t *testing.T) {
	oldVal := AuditLog
	AuditLog = true
	defer func() { AuditLog = oldVal }()

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Audit("team", "backend", "create", "ok")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	var entry map[string]string
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, output)
	}
	if entry["scope"] != "team" {
		t.Errorf("expected scope=team, got %q", entry["scope"])
	}
	if entry["target"] != "backend" {
		t.Errorf("expected target=backend, got %q", entry["target"])
	}
	if entry["action"] != "create" {
		t.Errorf("expected action=create, got %q", entry["action"])
	}
	if entry["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", entry["status"])
	}
	if entry["ts"] == "" {
		t.Error("expected non-empty ts field")
	}
}

func TestAudit_Disabled(t *testing.T) {
	oldVal := AuditLog
	AuditLog = false
	defer func() { AuditLog = oldVal }()

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Audit("team", "backend", "create", "ok")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() != 0 {
		t.Errorf("expected no output when audit disabled, got %q", buf.String())
	}
}

func TestEnableDebug(t *testing.T) {
	oldLevel := levelVar.Level()
	defer levelVar.Set(oldLevel)

	EnableDebug()

	if !Logger().Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected debug level enabled after EnableDebug")
	}
}
````

## File: internal/version/version.go
````go
package version

import (
	"runtime/debug"
)

// BuildInfo contains version and build information
type BuildInfo struct {
	Version    string
	Revision   string
	Modified   bool
	CommitTime string
}

// GetBuildInfo returns version information from runtime/debug
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version: "dev",
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	// Extract VCS information from build settings
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			info.Revision = setting.Value
			if len(info.Revision) > 12 {
				info.Revision = info.Revision[:12]
			}
		case "vcs.time":
			info.CommitTime = setting.Value
		case "vcs.modified":
			info.Modified = setting.Value == "true"
		}
	}

	// Use VCS revision as version if available, otherwise check for version in main module
	if info.Revision != "" {
		if info.Modified {
			info.Version = info.Revision + "-dirty"
		} else {
			info.Version = info.Revision
		}
	} else if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		info.Version = buildInfo.Main.Version
	}

	return info
}
````

## File: main.go
````go
package main

import "github.com/DragonSecurity/gomgr/cmd"

func main() {
	cmd.Execute()
}
````

## File: cmd/root.go
````go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgDir   string
	debug    bool
	dryRun   bool
	timeout  time.Duration
	auditLog bool
)

var rootCmd = &cobra.Command{
	Use:   "gomgr",
	Short: "GitHub Organization Manager (Go)",
	Long:  "Sync GitHub org owners, teams, members, and repo permissions from YAML.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "", "Path to config directory (required)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose debug logs")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry", false, "Show a plan without applying changes")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Minute, "Overall context timeout for the sync operation")
	rootCmd.PersistentFlags().BoolVar(&auditLog, "audit-log", false, "Emit structured JSON audit log entries to stderr")
}
````

## File: config/example/teams/example-team.yaml
````yaml
name: Platform Team
description: Core platform engineers
privacy: closed
parents: []
maintainers:
  - octocat
repositories:
  # Simple permission string (backward compatible)
  infra: maintain
  
  # Template repository - can be reused by other repos
  template-go-api:
    permission: push
    template: true
    topics:
      - backend
      - api
      - go-template
  
  # Advanced config with topics and pinning
  api:
    permission: push
    topics:
      - backend
      - api
      - project-platform
  
  # Repository using template (inherits permission and topics)
  my-api:
    from: template-go-api
    topics:
      - my-project
      # Will inherit: backend, api, go-template from template
  
  # Repository with pinning
  platform-index:
    permission: admin
    topics:
      - project-platform
      - documentation
    pinned: true
  
  # Simple read-only access
  web: pull
````

## File: examples/config/app.yaml
````yaml
org: KaMuses
# GitHub App auth (preferred):
app_id: 1719369
private_key: ./dsec-gom.2026-02-20.private-key.pem

dry_warnings:
  warn_unmanaged_teams: true
  warn_members_without_any_team: true
  warn_unmanaged_repos: true
  warn_unmanaged_custom_roles: true  # warn about custom roles not defined in org.yaml

remove_members_without_team: true   # remove org members not in any team
delete_unconfigured_teams: true     # delete teams not defined in YAML
delete_unmanaged_repos: true        # delete repos not defined in any team (DESTRUCTIVE!)
delete_unmanaged_custom_roles: false # delete custom roles not defined in org.yaml
create_repo: true                   # create repos if missing when referenced by teams
add_renovate_config: true           # create .github/renovate.json in repos
add_default_readme: false           # create default README.md in repos (optional)

renovate_config: | 
  { 
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }
````

## File: internal/config/types_test.go
````go
package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRepoConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantName string
		wantType string // "simple" or "advanced"
	}{
		{
			name: "simple string permission",
			yaml: `
name: Test Team
repositories:
  simple-repo: push
`,
			wantName: "simple-repo",
			wantType: "simple",
		},
		{
			name: "advanced config with topics",
			yaml: `
name: Test Team
repositories:
  advanced-repo:
    permission: maintain
    topics:
      - backend
      - project-test
`,
			wantName: "advanced-repo",
			wantType: "advanced",
		},
		{
			name: "pinned repository",
			yaml: `
name: Test Team
repositories:
  test-index:
    permission: admin
    topics:
      - documentation
    pinned: true
`,
			wantName: "test-index",
			wantType: "advanced",
		},
		{
			name: "mixed configuration",
			yaml: `
name: Test Team
repositories:
  simple: push
  advanced:
    permission: admin
    topics: [backend]
    pinned: true
`,
			wantName: "advanced",
			wantType: "advanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var team TeamConfig
			if err := yaml.Unmarshal([]byte(tt.yaml), &team); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if team.Name != "Test Team" {
				t.Errorf("Expected team name 'Test Team', got %q", team.Name)
			}

			if len(team.Repositories) == 0 {
				t.Fatal("No repositories found")
			}

			val, ok := team.Repositories[tt.wantName]
			if !ok {
				t.Fatalf("Repository %q not found", tt.wantName)
			}

			switch tt.wantType {
			case "simple":
				if _, ok := val.(string); !ok {
					t.Errorf("Expected string type, got %T", val)
				}
			case "advanced":
				if _, ok := val.(map[string]any); !ok {
					t.Errorf("Expected map[string]any type, got %T", val)
				}
			}
		})
	}
}

func TestAppConfigParsing(t *testing.T) {
	yamlData := `
org: TestOrg
dry_warnings:
  warn_unmanaged_teams: true
  warn_members_without_any_team: true
  warn_unmanaged_repos: true
remove_members_without_team: false
delete_unconfigured_teams: false
delete_unmanaged_repos: true
create_repo: true
`

	var app AppConfig
	if err := yaml.Unmarshal([]byte(yamlData), &app); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if app.Org != "TestOrg" {
		t.Errorf("Expected org 'TestOrg', got %q", app.Org)
	}

	if !app.DryWarnings.WarnUnmanagedRepos {
		t.Error("Expected WarnUnmanagedRepos to be true")
	}

	if !app.DeleteUnmanagedRepos {
		t.Error("Expected DeleteUnmanagedRepos to be true")
	}

	if !app.CreateRepo {
		t.Error("Expected CreateRepo to be true")
	}
}

func TestOrgConfigParsingWithCustomRoles(t *testing.T) {
	yamlData := `
owners:
  - alice
  - bob
custom_roles:
  - name: actions-manager
    description: Manage GitHub Actions without code access
    base_role: read
    permissions:
      - manage_actions
      - manage_runners
  - name: release-manager
    description: Manage releases and deployments
    base_role: write
    permissions:
      - create_releases
      - manage_environments
`

	var org OrgConfig
	if err := yaml.Unmarshal([]byte(yamlData), &org); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(org.Owners) != 2 {
		t.Errorf("Expected 2 owners, got %d", len(org.Owners))
	}

	if len(org.CustomRoles) != 2 {
		t.Fatalf("Expected 2 custom roles, got %d", len(org.CustomRoles))
	}

	// Check first role
	role1 := org.CustomRoles[0]
	if role1.Name != "actions-manager" {
		t.Errorf("Expected name 'actions-manager', got %q", role1.Name)
	}
	if role1.BaseRole != "read" {
		t.Errorf("Expected base_role 'read', got %q", role1.BaseRole)
	}
	if len(role1.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(role1.Permissions))
	}

	// Check second role
	role2 := org.CustomRoles[1]
	if role2.Name != "release-manager" {
		t.Errorf("Expected name 'release-manager', got %q", role2.Name)
	}
	if role2.BaseRole != "write" {
		t.Errorf("Expected base_role 'write', got %q", role2.BaseRole)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		root      Root
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config",
			root: Root{
				App: AppConfig{Org: "myorg"},
				Team: []TeamConfig{
					{Name: "backend", Privacy: "closed"},
					{Name: "frontend", Privacy: "secret"},
					{Name: "ops"},
				},
				Org: OrgConfig{
					CustomRoles: []CustomRoleConfig{
						{Name: "deployer", BaseRole: "write"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid team privacy",
			root: Root{
				App:  AppConfig{Org: "myorg"},
				Team: []TeamConfig{{Name: "backend", Privacy: "public"}},
			},
			wantErr:   true,
			errSubstr: "invalid privacy",
		},
		{
			name: "invalid custom role base_role",
			root: Root{
				App: AppConfig{Org: "myorg"},
				Org: OrgConfig{
					CustomRoles: []CustomRoleConfig{
						{Name: "deployer", BaseRole: "superadmin"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "invalid base_role",
		},
		{
			name: "empty team name",
			root: Root{
				App:  AppConfig{Org: "myorg"},
				Team: []TeamConfig{{Name: ""}},
			},
			wantErr:   true,
			errSubstr: "team name must not be empty",
		},
		{
			name: "empty custom role name",
			root: Root{
				App: AppConfig{Org: "myorg"},
				Org: OrgConfig{
					CustomRoles: []CustomRoleConfig{
						{Name: "", BaseRole: "read"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "custom role name must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.root.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Validate() error = %q, want substring %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestValidateRepoName(t *testing.T) {
	tests := []struct {
		name      string
		repoName  string
		wantErr   bool
		errSubstr string
	}{
		{name: "valid simple", repoName: "my-repo", wantErr: false},
		{name: "valid with dots", repoName: "repo.name", wantErr: false},
		{name: "valid with underscore", repoName: "repo_name", wantErr: false},
		{name: "valid single char", repoName: "a", wantErr: false},
		{name: "invalid space", repoName: "repo name", wantErr: true, errSubstr: "invalid characters"},
		{name: "invalid at sign", repoName: "repo@name", wantErr: true, errSubstr: "invalid characters"},
		{name: "invalid dot dot", repoName: "..", wantErr: true, errSubstr: "cannot be"},
		{name: "invalid single dot", repoName: ".", wantErr: true, errSubstr: "cannot be"},
		{name: "invalid too long", repoName: strings.Repeat("a", 101), wantErr: true, errSubstr: "1-100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Root{
				App: AppConfig{Org: "myorg"},
				Team: []TeamConfig{
					{Name: "test", Repositories: map[string]any{tt.repoName: "push"}},
				},
			}
			err := r.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Validate() error = %q, want substring %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		wantErr   bool
		errSubstr string
	}{
		{name: "valid simple", username: "alice", wantErr: false},
		{name: "valid with hyphen", username: "bob-smith", wantErr: false},
		{name: "valid alphanumeric", username: "a1", wantErr: false},
		{name: "valid single char", username: "a", wantErr: false},
		{name: "invalid consecutive hyphens", username: "a--b", wantErr: true, errSubstr: "consecutive hyphens"},
		{name: "invalid starts with hyphen", username: "-bob", wantErr: true, errSubstr: "invalid characters"},
		{name: "invalid ends with hyphen", username: "bob-", wantErr: true, errSubstr: "invalid characters"},
		{name: "invalid special chars", username: "alice@bob", wantErr: true, errSubstr: "invalid characters"},
		{name: "invalid too long", username: strings.Repeat("a", 40), wantErr: true, errSubstr: "1-39 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Root{
				App: AppConfig{Org: "myorg"},
				Team: []TeamConfig{
					{Name: "test", Members: []string{tt.username}},
				},
			}
			err := r.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Validate() error = %q, want substring %q", err.Error(), tt.errSubstr)
			}
		})
	}
}
````

## File: internal/util/diff.go
````go
package util

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Change struct {
	Scope   string      `json:"scope"`
	Target  string      `json:"target"`
	Action  string      `json:"action"`
	Details interface{} `json:"details"`
}

type StateStats struct {
	Teams           StatePair `json:"teams"`
	TeamMembers     StatePair `json:"team_members"`
	Repositories    StatePair `json:"repositories"`
	RepoPermissions StatePair `json:"repo_permissions"`
	CustomRoles     StatePair `json:"custom_roles"`
}

type StatePair struct {
	Current int `json:"current"`
	Desired int `json:"desired"`
}

type Plan struct {
	Changes  []Change    `json:"changes"`
	Warnings []string    `json:"warnings"`
	Stats    *StateStats `json:"stats,omitempty"`
}

func PrintPlan(p Plan) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// PrintSummary prints a human-readable summary of the plan
func PrintSummary(p Plan) {
	separator := "\n" + "================================================================"
	fmt.Println(separator)
	fmt.Println("Summary of Proposed Changes")
	fmt.Println("================================================================")

	// Show current vs desired state if available
	if p.Stats != nil {
		fmt.Println("\nCurrent State vs Desired State:")
		fmt.Println("--------------------------------")

		printStatePair("Teams:", p.Stats.Teams)
		printStatePair("Team Members:", p.Stats.TeamMembers)
		printStatePair("Repositories:", p.Stats.Repositories)
		printStatePair("Repo Permissions:", p.Stats.RepoPermissions)
		printStatePair("Custom Roles:", p.Stats.CustomRoles)
		fmt.Println()
	}

	if len(p.Changes) == 0 {
		fmt.Println("No changes required - configuration is in sync")
		fmt.Println("\n" + "================================================================")
		return
	}

	// Count changes by scope
	scopeCounts := make(map[string]int)
	actionCounts := make(map[string]int)

	for _, ch := range p.Changes {
		scopeCounts[ch.Scope]++
		actionCounts[ch.Action]++
	}

	fmt.Printf("Total changes: %d\n\n", len(p.Changes))

	// Print by scope
	fmt.Println("Changes by scope:")
	scopes := make([]string, 0, len(scopeCounts))
	for scope := range scopeCounts {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)

	for _, scope := range scopes {
		count := scopeCounts[scope]
		fmt.Printf("  %-20s %d\n", scope+":", count)
	}

	fmt.Println("\nChanges by action:")
	actions := make([]string, 0, len(actionCounts))
	for action := range actionCounts {
		actions = append(actions, action)
	}
	sort.Strings(actions)

	for _, action := range actions {
		count := actionCounts[action]
		fmt.Printf("  %-20s %d\n", action+":", count)
	}

	if len(p.Warnings) > 0 {
		fmt.Printf("\nWarnings: %d\n", len(p.Warnings))
		for _, w := range p.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println("\n" + "================================================================")
}

// printStatePair prints a single state comparison line with delta
func printStatePair(label string, pair StatePair) {
	if pair.Current == 0 && pair.Desired == 0 {
		return
	}

	fmt.Printf("  %-20s %d → %d", label, pair.Current, pair.Desired)
	delta := pair.Desired - pair.Current
	if delta > 0 {
		fmt.Printf(" (+%d)\n", delta)
	} else if delta < 0 {
		fmt.Printf(" (%d)\n", delta)
	} else {
		fmt.Println(" (no change)")
	}
}
````

## File: internal/util/log.go
````go
package util

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// AuditLog controls whether structured audit log entries are emitted to stderr.
var AuditLog bool

var (
	levelVar = new(slog.LevelVar)
	logger   = newLogger(false)
)

func newLogger(addSource bool) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: addSource,
	}))
}

// Logger returns the package-level structured logger.
func Logger() *slog.Logger { return logger }

// EnableDebug switches the package logger to debug level with source info.
func EnableDebug() {
	levelVar.Set(slog.LevelDebug)
	logger = newLogger(true)
	slog.SetDefault(logger)
}

// Infof emits a formatted info-level log line via slog.
func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Debugf emits a formatted debug-level log line via slog.
func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Warnf prints a warning message to stderr.
func Warnf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", v...)
}

// Audit emits a structured JSON audit log entry to stderr when AuditLog is enabled.
func Audit(scope, target, action, status string) {
	if !AuditLog {
		return
	}
	entry := map[string]string{
		"ts":     time.Now().UTC().Format(time.RFC3339),
		"scope":  scope,
		"target": target,
		"action": action,
		"status": status,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit marshal error: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}
````

## File: .gitignore
````
*.pem

# gomgr binary
/gomgr
/build/

# Test coverage
/coverage/

# Created by https://www.toptal.com/developers/gitignore/api/go,goland
# Edit at https://www.toptal.com/developers/gitignore?templates=go,goland

### Go ###
# If you prefer the allow list template instead of the deny list, see community template:
# https://github.com/github/gitignore/blob/main/community/Golang/Go.AllowList.gitignore
#
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

### GoLand ###
# Covers JetBrains IDEs: IntelliJ, RubyMine, PhpStorm, AppCode, PyCharm, CLion, Android Studio, WebStorm and Rider
# Reference: https://intellij-support.jetbrains.com/hc/en-us/articles/206544839

# User-specific stuff
.idea/**/workspace.xml
.idea/**/tasks.xml
.idea/**/usage.statistics.xml
.idea/**/dictionaries
.idea/**/shelf

# AWS User-specific
.idea/**/aws.xml

# Generated files
.idea/**/contentModel.xml

# Sensitive or high-churn files
.idea/**/dataSources/
.idea/**/dataSources.ids
.idea/**/dataSources.local.xml
.idea/**/sqlDataSources.xml
.idea/**/dynamic.xml
.idea/**/uiDesigner.xml
.idea/**/dbnavigator.xml

# Gradle
.idea/**/gradle.xml
.idea/**/libraries

# Gradle and Maven with auto-import
# When using Gradle or Maven with auto-import, you should exclude module files,
# since they will be recreated, and may cause churn.  Uncomment if using
# auto-import.
# .idea/artifacts
# .idea/compiler.xml
# .idea/jarRepositories.xml
# .idea/modules.xml
# .idea/*.iml
# .idea/modules
# *.iml
# *.ipr

# CMake
cmake-build-*/

# Mongo Explorer plugin
.idea/**/mongoSettings.xml

# File-based project format
*.iws

# IntelliJ
out/

# mpeltonen/sbt-idea plugin
.idea_modules/

# JIRA plugin
atlassian-ide-plugin.xml

# Cursive Clojure plugin
.idea/replstate.xml

# SonarLint plugin
.idea/sonarlint/

# Crashlytics plugin (for Android Studio and IntelliJ)
com_crashlytics_export_strings.xml
crashlytics.properties
crashlytics-build.properties
fabric.properties

# Editor-based Rest Client
.idea/httpRequests

# Android studio 3.1+ serialized cache file
.idea/caches/build_file_checksums.ser

### GoLand Patch ###
# Comment Reason: https://github.com/joeblau/gitignore.io/issues/186#issuecomment-215987721

# *.iml
# modules.xml
# .idea/misc.xml
# *.ipr

# Sonarlint plugin
# https://plugins.jetbrains.com/plugin/7973-sonarlint
.idea/**/sonarlint/

# SonarQube Plugin
# https://plugins.jetbrains.com/plugin/7238-sonarqube-community-plugin
.idea/**/sonarIssues.xml

# Markdown Navigator plugin
# https://plugins.jetbrains.com/plugin/7896-markdown-navigator-enhanced
.idea/**/markdown-navigator.xml
.idea/**/markdown-navigator-enh.xml
.idea/**/markdown-navigator/

# Cache file creation bug
# See https://youtrack.jetbrains.com/issue/JBR-2257
.idea/$CACHE_FILE$

# CodeStream plugin
# https://plugins.jetbrains.com/plugin/12206-codestream
.idea/codestream.xml

# Azure Toolkit for IntelliJ plugin
# https://plugins.jetbrains.com/plugin/8053-azure-toolkit-for-intellij
.idea/**/azureSettings.xml

# End of https://www.toptal.com/developers/gitignore/api/go,goland

.idea

.env

.pem
````

## File: .golangci.yml
````yaml
version: "2"

run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly
  go: '1.26'

formatters:
  enable:
    - gofmt
    - goimports
  settings:
    goimports:
      local-prefixes:
        - github.com/DragonSecurity/gomgr

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
    - gocyclo
    - misspell
    - unconvert
    - goconst
    - revive
    - gosec
  settings:
    gocyclo:
      min-complexity: 45
    goconst:
      min-len: 3
      min-occurrences: 3
    misspell:
      locale: US
    revive:
      confidence: 0.8
      rules:
        - name: package-comments
          disabled: true
        - name: exported
          disabled: true
    gosec:
      excludes:
        - G304 # Audit use of file path - we need this for config files
        - G301 # Directory permissions 0755 are fine for config dirs
        - G306 # File permissions 0644 are fine for config files
        - G101 # False positives on test PEM fixtures
        - G703 # Path traversal via taint analysis - config paths come from user
  exclusions:
    rules:
      - path: _test\.go
        linters:
          - gocyclo
          - errcheck
          - gosec
          - goconst

output:
  formats:
    text:
      print-issued-lines: true
      print-linter-name: true
````

## File: AGENTS.md
````markdown
# Agents

This document describes the automation agents and workflows used in the gomgr (GitHub Organization Manager) project.

## GitHub Organization Manager Agent

The core of this project is an automated agent that manages GitHub organization resources in an idempotent, declarative manner. The agent reads YAML configuration files and applies them to your GitHub organization.

### Agent Capabilities

The gomgr agent can:

- **Team Management**
  - Create and configure teams with specified privacy levels
  - Assign team maintainers and members
  - Optionally delete teams not defined in configuration
  - Warn about unmanaged teams

- **Member Management**
  - Add members to teams with appropriate roles
  - Optionally remove members not assigned to any team
  - Warn about members without team assignments

- **Repository Permissions**
  - Grant team-level repository access (pull/triage/push/maintain/admin)
  - Support custom repository roles (GitHub Enterprise Cloud)
  - Optionally create repositories if they don't exist
  - Inject Renovate configuration into repositories
  - Optionally inject default README into repositories

- **Repository Management**
  - Add topics/labels to repositories for better organization
  - Mark repositories as templates for reuse
  - Support template inheritance (permission and topics)
  - Pin important repositories to organization profile (API limitation: must be done manually via web UI)
  - Optionally delete unmanaged repositories
  - Warn about repositories not defined in any team configuration

- **Custom Repository Roles**
  - Create and update custom repository roles (GitHub Enterprise Cloud)
  - Delete unmanaged custom roles (optional)
  - Warn about unmanaged custom roles
  - Support fine-grained permissions for specialized access patterns

- **Synchronization**
  - Idempotent apply: safe to run repeatedly
  - Dry-run mode for safe planning before applying changes
  - Stable output for predictable CI/CD integration
  - State comparison showing current vs desired state

### Agent Authentication

The agent supports two authentication methods:

1. **GitHub App (Recommended)**
   - More secure with fine-grained permissions
   - Can be scoped to specific organizations
   - Requires `GITHUB_APP_ID` and `GITHUB_APP_PRIVATE_KEY`

2. **Personal Access Token (PAT)**
   - Simpler setup for smaller use cases
   - Requires `GITHUB_TOKEN` environment variable
   - Uses classic PAT with `admin:org`, `repo`, and `read:org` scopes

### Agent Operations

The agent performs operations in the following order:

1. **Create Custom Roles** - Creates/updates custom repository roles (GitHub Enterprise Cloud)
2. **Create Teams** - Ensures all teams defined in YAML exist
3. **Set Memberships** - Assigns maintainers and members to teams
4. **Ensure Repos** - Creates repositories if configured to do so
5. **Mark Templates** - Marks repositories as templates if configured
6. **Grant Permissions** - Applies repository access permissions to teams (including custom roles)
7. **Write Files** - Optionally injects default README and `.github/renovate.json` into repos
8. **Set Topics** - Applies topics/labels to repositories for organization
9. **Pin Repos** - Attempts to pin repositories (warning issued due to API limitation)
10. **Cleanups** - Optionally removes unmanaged resources (teams, members, repositories)
11. **Delete Custom Roles** - Optionally removes unmanaged custom roles (if configured)

## CI/CD Automation

### Release Agent

The project includes a GitHub Actions workflow (`.github/workflows/release.yaml`) that automates the release process:

- **Trigger**: Tag push (`v*.*.*`) or manual workflow dispatch
- **Platforms**: Builds for Linux, macOS, and Windows
- **Architectures**: Supports both amd64 and arm64
- **Artifacts**: Creates packaged binaries with version stamping
- **Distribution**: Uploads release artifacts and checksums to GitHub Releases

**Usage:**
```bash
git tag v0.1.0
git push origin v0.1.0
```

### Organization Sync Agent (Template)

While not included in this repository, the README provides a template workflow for automating organization synchronization in your org-config repository:

**Features:**
- Runs on push to main branch or manual trigger
- Supports multiple organizations via matrix strategy
- Downloads and installs the appropriate gomgr version
- Executes sync with GitHub App authentication
- Continues on error to allow other matrix jobs to complete

**Example workflow structure:**
```yaml
name: Synchronise organization users and teams (gomgr)
on:
  push:
    branches: [ "main" ]
  workflow_dispatch:
jobs:
  sync:
    runs-on: ubuntu-24.04
    strategy:
      fail-fast: false
      matrix:
        config:
          - { folder: "org1", gom_version: "v0.12.2" }
          - { folder: "org2", gom_version: "v0.10.2" }
```

## Agent Configuration

Agents are configured through YAML files in a config directory:

### `app.yaml` - Agent Settings
Defines the target organization, authentication method, and behavioral flags:
- Organization name
- GitHub App credentials or PAT
- Warning flags for dry-run mode (unmanaged teams, members without teams, unmanaged repos, unmanaged custom roles)
- Optional enforcement features (remove members, delete teams, delete unmanaged repos, delete custom roles, create repos)
- Renovate configuration injection
- Default README injection (optional)

### `org.yaml` - Organization Metadata
Defines organization owners and custom repository roles (GitHub Enterprise Cloud):
- List of organization owners
- Custom repository role definitions with base roles and permissions

### `teams/*.yaml` - Team Definitions
Each file defines a team with:
- Name and slug
- Description and privacy level
- Maintainers and members
- Repository access permissions with optional advanced configuration:
  - Simple string permission (backward compatible): `repo: push`
  - Advanced object with topics, pinning, and templates:
    ```yaml
    repo:
      permission: push
      topics: [backend, api, project-name]
      pinned: true
      template: true
      from: template-repo  # inherit from template
    ```

## Agent Safety Features

- **Dry-run Mode**: Preview changes without applying them (`--dry` flag)
- **Stable Output**: Predictable output format for CI/CD validation
- **Idempotent Operations**: Safe to run multiple times without side effects
- **Least Privilege**: GitHub App authentication with minimal required permissions
- **Fail-safe Warnings**: Alerts about unmanaged resources before cleanup

## Agent Observability

- **Debug Mode**: Detailed logging with `--debug` flag
- **Version Information**: Built-in version reporting with VCS details
- **Rate Limit Awareness**: Respects GitHub API rate limits
- **Error Reporting**: Clear error messages for troubleshooting

## Future Agent Enhancements

The roadmap includes:

- Compare and update team fields (description, privacy, parents)
- Optionally remove extra team members or revoke excess permissions
- Parallel apply with rate-limit aware workers
- More comprehensive plan diff output
- Custom default branch for file writes

## Security Considerations

This tool acts as a powerful automation agent that can modify organization membership and repository access. To use it safely:

- **Always test with `--dry` flag first** in CI environments
- **Use least privilege credentials** - GitHub Apps preferred over PATs
- **Review changes** before applying in production
- **Restrict workflow permissions** to prevent unauthorized modifications
- **Store credentials securely** using GitHub Secrets or secure vault solutions
- **Audit changes** by reviewing agent logs and GitHub audit logs

## Contributing to Agent Development

When contributing new agent capabilities:

1. Open an issue first for larger changes
2. Keep commits small and focused
3. Add tests where practical
4. Run `golangci-lint` if configured
5. Document new configuration options in README.md
6. Update this AGENTS.md file with new capabilities

## License

See [LICENSE](./LICENSE.md).
````

## File: Makefile
````makefile
.PHONY: help build test test-coverage test-verbose lint fmt vet security clean install-tools all check ci

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=gomgr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s"
BUILD_DIR=build
COVERAGE_DIR=coverage

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOMOD=$(GOCMD) mod

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: clean fmt vet lint test build ## Run all checks and build

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) .

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./internal/config ./internal/sync ./internal/util ./internal/gh
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1

test-verbose: ## Run tests with verbose output
	@echo "Running verbose tests..."
	$(GOTEST) -v -race ./...

test-short: ## Run short tests only
	@echo "Running short tests..."
	$(GOTEST) -short ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -w -s .
	@echo "Code formatted"

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Code is properly formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "go vet passed"

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m ./...; \
		echo "Linting passed"; \
	elif [ -x "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		$(shell go env GOPATH)/bin/golangci-lint run --timeout 5m ./...; \
		echo "Linting passed"; \
	else \
		echo "golangci-lint not found. Install it with: make install-tools"; \
		exit 1; \
	fi

security: ## Run security checks with gosec (requires gosec to be installed)
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
		echo "Security check passed"; \
	elif [ -x "$(shell go env GOPATH)/bin/gosec" ]; then \
		$(shell go env GOPATH)/bin/gosec -quiet ./...; \
		echo "Security check passed"; \
	else \
		echo "gosec not found. Install it with: make install-tools"; \
		exit 1; \
	fi

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	$(GOMOD) tidy
	@echo "Modules tidied"

mod-verify: ## Verify go modules
	@echo "Verifying go modules..."
	$(GOMOD) verify
	@echo "Modules verified"

mod-download: ## Download go modules
	@echo "Downloading go modules..."
	$(GOMOD) download
	@echo "Modules downloaded"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Cleaned"

install-tools: ## Install development tools (golangci-lint, gosec)
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	else \
		echo "golangci-lint already installed"; \
	fi
	@echo "Installing gosec..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	else \
		echo "gosec already installed"; \
	fi
	@echo "All tools installed"

check: fmt-check vet test ## Run basic checks (format, vet, test)

check-all: fmt-check vet lint security test ## Run all checks including lint and security

ci: clean check build ## Run CI pipeline (clean, run basic checks, build)

version: ## Display version information
	@echo "Version: $(VERSION)"

.PHONY: mod-tidy mod-verify mod-download version
````

## File: cmd/setup_team.go
````go
package cmd

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
)

var teamName string
var outFile string

var setupTeamCmd = &cobra.Command{
	Use:   "setup-team",
	Short: "Bootstrap a team YAML file for a given team name",
	Example: `  gomgr setup-team -c ./config -n "Backend"
  gomgr setup-team -n "Frontend" -f ./teams/frontend.yaml`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slug := strings.ToLower(strings.ReplaceAll(teamName, " ", "-"))
		path := outFile
		if path == "" {
			path = filepath.Join(cfgDir, "teams", slug+".yaml")
		}
		return config.BootstrapTeamYAML(path, teamName)
	},
}

func init() {
	setupTeamCmd.Flags().StringVarP(&teamName, "name", "n", "", "Team display name (required)")
	_ = setupTeamCmd.MarkFlagRequired("name")
	setupTeamCmd.Flags().StringVarP(&outFile, "file", "f", "", "Force output file path")
	rootCmd.AddCommand(setupTeamCmd)
}
````

## File: config/example/app.yaml
````yaml
org: KaMuses
# PAT path: set env GITHUB_TOKEN
# GitHub App path:
app_id: 1719369
private_key: ./dsec-gom.2025-10-02.private-key.pem
dry_warnings:
  warn_unmanaged_teams: true
  warn_members_without_any_team: true
  warn_unmanaged_repos: true
  warn_unmanaged_custom_roles: true
remove_members_without_team: true
delete_unconfigured_teams: true
delete_unmanaged_repos: false  # Set to true to delete repos not defined in any team
delete_unmanaged_custom_roles: false  # Set to true to delete custom roles not in org.yaml
create_repo: true
add_default_readme: true
add_renovate_config: true
renovate_config: | 
  { 
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }
````

## File: internal/config/loader.go
````go
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(dir string) (*Root, error) {
	r := &Root{}
	// app.yaml
	if err := readYAML(filepath.Join(dir, "app.yaml"), &r.App); err != nil {
		return nil, err
	}
	// org.yaml
	if err := readYAML(filepath.Join(dir, "org.yaml"), &r.Org); err != nil {
		return nil, err
	}
	// teams/*.yaml
	teamDir := filepath.Join(dir, "teams")
	entries, err := os.ReadDir(teamDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read teams directory %s: %w", teamDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue // ignore non-YAML files like .DS_Store, README, etc.
		}
		var t TeamConfig
		if err := readYAML(filepath.Join(teamDir, e.Name()), &t); err != nil {
			return nil, err
		}
		r.Team = append(r.Team, t)
	}
	if r.App.Org == "" {
		return nil, errors.New("app.org is required")
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parse YAML in %s: %w", path, err)
	}
	return nil
}

// Validate checks that the loaded configuration is semantically correct.
func (r *Root) Validate() error {
	validPrivacy := map[string]bool{"": true, "closed": true, "secret": true}
	validBaseRole := map[string]bool{"read": true, "triage": true, "write": true, "maintain": true, "admin": true}

	for _, t := range r.Team {
		if t.Name == "" {
			return fmt.Errorf("team name must not be empty")
		}
		if !validPrivacy[t.Privacy] {
			return fmt.Errorf("team %q has invalid privacy %q (must be closed or secret)", t.Name, t.Privacy)
		}
		for repo := range t.Repositories {
			if err := validateRepoName(repo); err != nil {
				return fmt.Errorf("team %q: %w", t.Name, err)
			}
		}
		for _, u := range t.Maintainers {
			if err := validateUsername(u); err != nil {
				return fmt.Errorf("team %q maintainer: %w", t.Name, err)
			}
		}
		for _, u := range t.Members {
			if err := validateUsername(u); err != nil {
				return fmt.Errorf("team %q member: %w", t.Name, err)
			}
		}
	}
	for _, cr := range r.Org.CustomRoles {
		if cr.Name == "" {
			return fmt.Errorf("custom role name must not be empty")
		}
		if !validBaseRole[cr.BaseRole] {
			return fmt.Errorf("custom role %q has invalid base_role %q (must be read|triage|write|maintain|admin)", cr.Name, cr.BaseRole)
		}
	}
	for _, u := range r.Org.Owners {
		if err := validateUsername(u); err != nil {
			return fmt.Errorf("org owner: %w", err)
		}
	}
	return nil
}

var validRepoName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func validateRepoName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return fmt.Errorf("repo name must be 1-100 characters: %q", name)
	}
	if !validRepoName.MatchString(name) {
		return fmt.Errorf("repo name contains invalid characters: %q", name)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("repo name cannot be %q", name)
	}
	return nil
}

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

func validateUsername(name string) error {
	if len(name) == 0 || len(name) > 39 {
		return fmt.Errorf("username must be 1-39 characters: %q", name)
	}
	if !validUsername.MatchString(name) {
		return fmt.Errorf("username contains invalid characters: %q", name)
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("username cannot contain consecutive hyphens: %q", name)
	}
	return nil
}

func BootstrapTeamYAML(path string, name string) error {
	t := TeamConfig{
		Name:         name,
		Maintainers:  []string{},
		Members:      []string{},
		Repositories: map[string]any{},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
````

## File: internal/sync/custom_roles.go
````go
package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// customRoleChange represents a custom role modification
type customRoleChange struct {
	Org         string
	ID          int64
	Name        string
	Description string
	BaseRole    string
	Permissions []string
}

// planCustomRoles determines what custom repository role changes are needed
func planCustomRoles(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	if len(cfg.Org.CustomRoles) == 0 {
		// No custom roles configured
		return out, nil
	}

	// Fetch existing custom roles from GitHub
	existingRolesResp, _, err := c.REST.Organizations.ListCustomRepoRoles(ctx, org)
	if err != nil {
		// If the org doesn't have custom roles enabled (not Enterprise Cloud),
		// return an error with helpful context
		return out, fmt.Errorf("list custom repo roles: %w (note: custom roles require GitHub Enterprise Cloud)", err)
	}

	// Build map of existing roles by name (lowercase for case-insensitive comparison)
	existingByName := make(map[string]*github.CustomRepoRoles)
	for _, role := range existingRolesResp.CustomRepoRoles {
		if role.Name != nil {
			existingByName[strings.ToLower(*role.Name)] = role
		}
	}

	// Track state
	st.CurrentCustomRoles = len(existingRolesResp.CustomRepoRoles)
	st.DesiredCustomRoles = len(cfg.Org.CustomRoles)

	// Plan changes for each desired role
	for _, desiredRole := range cfg.Org.CustomRoles {
		roleName := desiredRole.Name
		roleNameLower := strings.ToLower(roleName)

		existingRole, exists := existingByName[roleNameLower]

		if !exists {
			// Create new role
			out = append(out, util.Change{
				Scope:  "custom-role",
				Target: roleName,
				Action: "create",
				Details: customRoleChange{
					Org:         org,
					Name:        roleName,
					Description: desiredRole.Description,
					BaseRole:    desiredRole.BaseRole,
					Permissions: desiredRole.Permissions,
				},
			})
		} else {
			// Check if update is needed
			needsUpdate := false

			// Check description changes
			existingDesc := ""
			if existingRole.Description != nil {
				existingDesc = *existingRole.Description
			}
			if existingDesc != desiredRole.Description {
				needsUpdate = true
			}

			// Check base role changes
			if existingRole.BaseRole != nil && *existingRole.BaseRole != desiredRole.BaseRole {
				needsUpdate = true
			}

			// Check permission changes
			if !permissionsEqual(existingRole.Permissions, desiredRole.Permissions) {
				needsUpdate = true
			}

			if needsUpdate {
				out = append(out, util.Change{
					Scope:  "custom-role",
					Target: roleName,
					Action: "update",
					Details: customRoleChange{
						Org:         org,
						ID:          existingRole.GetID(),
						Name:        roleName,
						Description: desiredRole.Description,
						BaseRole:    desiredRole.BaseRole,
						Permissions: desiredRole.Permissions,
					},
				})
			}
		}
	}

	return out, nil
}

// planCustomRoleCleanups determines which custom roles should be deleted
func planCustomRoleCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	if !cfg.App.DeleteUnmanagedCustomRoles && !cfg.App.DryWarnings.WarnUnmanagedCustomRoles {
		return out, warnings, nil
	}

	// Fetch existing custom roles
	existingRolesResp, _, err := c.REST.Organizations.ListCustomRepoRoles(ctx, org)
	if err != nil {
		// If custom roles aren't available, skip cleanup
		return out, warnings, nil
	}

	// Build set of desired role names (case-insensitive)
	desiredNames := make(map[string]bool)
	for _, role := range cfg.Org.CustomRoles {
		desiredNames[strings.ToLower(role.Name)] = true
	}

	// Find unmanaged roles
	var unmanagedRoles []string
	for _, role := range existingRolesResp.CustomRepoRoles {
		if role.Name == nil {
			continue
		}
		roleName := *role.Name
		if !desiredNames[strings.ToLower(roleName)] {
			unmanagedRoles = append(unmanagedRoles, roleName)
			if cfg.App.DeleteUnmanagedCustomRoles {
				out = append(out, util.Change{
					Scope:  "custom-role",
					Target: roleName,
					Action: "delete",
					Details: customRoleChange{
						Org:  org,
						ID:   role.GetID(),
						Name: roleName,
					},
				})
			}
		}
	}

	if cfg.App.DryWarnings.WarnUnmanagedCustomRoles && len(unmanagedRoles) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d unmanaged custom repository roles: %v", len(unmanagedRoles), unmanagedRoles))
	}

	return out, warnings, nil
}

// applyCustomRoleChanges handles creating, updating, and deleting custom roles
func applyCustomRoleChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	for _, ch := range changes {
		if !strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		util.Infof("custom-role:%s %s", ch.Action, ch.Target)

		if err := gh.RespectRate(ctx, c.REST); err != nil {
			util.Warnf("rate limit check failed: %v", err)
		}

		d, ok := ch.Details.(customRoleChange)
		if !ok {
			return fmt.Errorf("invalid details for custom-role change")
		}

		switch ch.Scope + ":" + ch.Action {
		case "custom-role:create":
			opts := &github.CreateOrUpdateCustomRepoRoleOptions{
				Name:        github.Ptr(d.Name),
				BaseRole:    github.Ptr(d.BaseRole),
				Permissions: d.Permissions,
			}
			if d.Description != "" {
				opts.Description = github.Ptr(d.Description)
			}

			_, _, err := c.REST.Organizations.CreateCustomRepoRole(ctx, d.Org, opts)
			if err != nil {
				util.Audit(ch.Scope, ch.Target, ch.Action, "error")
				return fmt.Errorf("create custom role %q: %w", d.Name, err)
			}

		case "custom-role:update":
			opts := &github.CreateOrUpdateCustomRepoRoleOptions{
				Name:        github.Ptr(d.Name),
				BaseRole:    github.Ptr(d.BaseRole),
				Permissions: d.Permissions,
			}
			if d.Description != "" {
				opts.Description = github.Ptr(d.Description)
			}

			_, _, err := c.REST.Organizations.UpdateCustomRepoRole(ctx, d.Org, d.ID, opts)
			if err != nil {
				util.Audit(ch.Scope, ch.Target, ch.Action, "error")
				return fmt.Errorf("update custom role %q (ID %d): %w", d.Name, d.ID, err)
			}

		case "custom-role:delete":
			_, err := c.REST.Organizations.DeleteCustomRepoRole(ctx, d.Org, d.ID)
			if err != nil {
				util.Audit(ch.Scope, ch.Target, ch.Action, "error")
				return fmt.Errorf("delete custom role %q (ID %d): %w", d.Name, d.ID, err)
			}
		}

		util.Audit(ch.Scope, ch.Target, ch.Action, "ok")
	}

	return nil
}

// permissionsEqual checks if two permission lists are equivalent
func permissionsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aSet := make(map[string]bool)
	for _, p := range a {
		aSet[p] = true
	}
	for _, p := range b {
		if !aSet[p] {
			return false
		}
	}
	return true
}
````

## File: cmd/version.go
````go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		info := version.GetBuildInfo()
		fmt.Println("Version:", info.Version)

		if info.Revision != "" {
			fmt.Printf("Revision: %s\n", info.Revision)
			fmt.Printf("Modified: %v\n", info.Modified)
		}

		if info.CommitTime != "" {
			fmt.Printf("LastCommit: %s\n", info.CommitTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
````

## File: internal/config/types.go
````go
package config

import "strings"

type AppConfig struct {
	AppID      int64  `yaml:"app_id,omitempty"`
	PrivateKey string `yaml:"private_key,omitempty"`
	Org        string `yaml:"org"`

	DryWarnings struct {
		WarnUnmanagedTeams        bool `yaml:"warn_unmanaged_teams"`
		WarnMembersWithoutAnyTeam bool `yaml:"warn_members_without_any_team"`
		WarnUnmanagedRepos        bool `yaml:"warn_unmanaged_repos"`
		WarnUnmanagedCustomRoles  bool `yaml:"warn_unmanaged_custom_roles"`
	} `yaml:"dry_warnings"`
	RemoveMembersWithoutTeam   bool   `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams    bool   `yaml:"delete_unconfigured_teams"`
	DeleteUnmanagedRepos       bool   `yaml:"delete_unmanaged_repos"`
	DeleteUnmanagedCustomRoles bool   `yaml:"delete_unmanaged_custom_roles"`
	CreateRepo                 bool   `yaml:"create_repo"`
	AddRenovateConfig          bool   `yaml:"add_renovate_config"`
	RenovateConfig             string `yaml:"renovate_config"`
	AddDefaultReadme           bool   `yaml:"add_default_readme"`
}

type OrgConfig struct {
	Owners      []string           `yaml:"owners"`
	CustomRoles []CustomRoleConfig `yaml:"custom_roles,omitempty"`
}

// CustomRoleConfig defines a custom repository role for the organization
// Requires GitHub Enterprise Cloud
type CustomRoleConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	BaseRole    string   `yaml:"base_role"` // read, triage, write, maintain, admin
	Permissions []string `yaml:"permissions,omitempty"`
}

type RepoConfig struct {
	Permission string   `yaml:"permission,omitempty"` // pull|triage|push|maintain|admin
	Topics     []string `yaml:"topics,omitempty"`
	Pinned     bool     `yaml:"pinned,omitempty"`
}

type TeamConfig struct {
	Name        string   `yaml:"name"`
	Slug        string   `yaml:"slug,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Privacy     string   `yaml:"privacy,omitempty"` // closed, secret
	Parents     []string `yaml:"parents,omitempty"`

	Maintainers []string `yaml:"maintainers,omitempty"`
	Members     []string `yaml:"members,omitempty"`

	// repo => permission (pull|triage|push|maintain|admin) or RepoConfig for advanced settings
	// For backward compatibility, supports both:
	//   repositories:
	//     infra: maintain               # simple string permission
	//     api:                          # or advanced RepoConfig
	//       permission: push
	//       topics: [backend, api]
	//       pinned: true
	Repositories map[string]any `yaml:"repositories,omitempty"`
}

type Root struct {
	App  AppConfig    `yaml:"app"`
	Org  OrgConfig    `yaml:"org"`
	Team []TeamConfig `yaml:"teams"`
}

// ResolvedSlug returns the team's slug, deriving it from the name if not explicitly set.
func (t TeamConfig) ResolvedSlug() string {
	if t.Slug != "" {
		return t.Slug
	}
	return strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
}
````

## File: internal/gh/rate.go
````go
package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/util"
)

func RespectRate(ctx context.Context, c *github.Client) error {
	r, _, err := c.RateLimit.Get(ctx)
	if err != nil {
		return fmt.Errorf("rate limit check: %w", err)
	}
	if r == nil {
		return nil
	}
	if core := r.GetCore(); core.Remaining < 50 {
		sleep := time.Until(core.Reset.Time) + time.Second
		util.Infof("rate-limit: sleeping %s until %s", sleep, core.Reset.Time)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
````

## File: internal/sync/teams_test.go
````go
package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestParseRepoConfig(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		wantPerm   string
		wantTopics []string
		wantPinned bool
		wantError  bool
	}{
		{
			name:       "simple string permission",
			input:      "push",
			wantPerm:   "push",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "advanced config with permission only",
			input: map[string]any{
				"permission": "maintain",
			},
			wantPerm:   "maintain",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
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
			wantError:  false,
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
			wantError:  false,
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
			wantError:  false,
		},
		{
			name: "missing permission field (topics only)",
			input: map[string]any{
				"topics": []any{"backend"},
			},
			wantPerm:   "",
			wantTopics: []string{"backend"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "empty topics array",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{},
			},
			wantPerm:   "push",
			wantTopics: nil,
			wantPinned: false,
			wantError:  false,
		},
		{
			name: "non-string values in topics array (should be ignored)",
			input: map[string]any{
				"permission": "push",
				"topics":     []any{"valid", 123, "another"},
			},
			wantPerm:   "push",
			wantTopics: []string{"valid", "another"},
			wantPinned: false,
			wantError:  false,
		},
		{
			name:       "empty string permission",
			input:      "",
			wantPerm:   "",
			wantTopics: nil,
			wantPinned: false,
			wantError:  true,
		},
		{
			name: "permission as non-string type",
			input: map[string]any{
				"permission": 123,
			},
			wantPerm:   "",
			wantTopics: nil,
			wantPinned: false,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return // Skip validation if error was expected
			}

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

func TestParseRepoConfigWithTemplate(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		wantPerm     string
		wantTopics   []string
		wantPinned   bool
		wantTemplate bool
		wantFrom     string
		wantError    bool
	}{
		{
			name: "template repository",
			input: map[string]any{
				"permission": "push",
				"template":   true,
				"topics":     []any{"backend", "template"},
			},
			wantPerm:     "push",
			wantTopics:   []string{"backend", "template"},
			wantPinned:   false,
			wantTemplate: true,
			wantFrom:     "",
			wantError:    false,
		},
		{
			name: "repository using template (same org)",
			input: map[string]any{
				"permission": "push",
				"from":       "template-go-api",
				"topics":     []any{"my-project"},
			},
			wantPerm:     "push",
			wantTopics:   []string{"my-project"},
			wantPinned:   false,
			wantTemplate: false,
			wantFrom:     "template-go-api",
			wantError:    false,
		},
		{
			name: "repository using template (cross-org)",
			input: map[string]any{
				"from":   "some-org/template-repo",
				"topics": []any{"backend"},
			},
			wantPerm:     "",
			wantTopics:   []string{"backend"},
			wantPinned:   false,
			wantTemplate: false,
			wantFrom:     "some-org/template-repo",
			wantError:    false,
		},
		{
			name: "template with from (both should work)",
			input: map[string]any{
				"permission": "admin",
				"template":   true,
				"from":       "another-template",
			},
			wantPerm:     "admin",
			wantTopics:   nil,
			wantPinned:   false,
			wantTemplate: true,
			wantFrom:     "another-template",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return
			}

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

			if settings.template != tt.wantTemplate {
				t.Errorf("template = %v, want %v", settings.template, tt.wantTemplate)
			}

			if settings.from != tt.wantFrom {
				t.Errorf("from = %q, want %q", settings.from, tt.wantFrom)
			}
		})
	}
}

func TestResolveTemplate(t *testing.T) {
	tests := []struct {
		name          string
		repoName      string
		settings      repoSettings
		allRepos      map[string]repoSettings
		defaultOrg    string
		wantPerm      string
		wantTopics    []string
		wantError     bool
		errorContains string
	}{
		{
			name:     "no template reference",
			repoName: "api",
			settings: repoSettings{
				permission: "push",
				topics:     []string{"backend"},
			},
			allRepos:   map[string]repoSettings{},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend"},
			wantError:  false,
		},
		{
			name:     "inherit from template",
			repoName: "my-api",
			settings: repoSettings{
				from:   "template-go-api",
				topics: []string{"my-project"},
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend", "api"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend", "api", "my-project"},
			wantError:  false,
		},
		{
			name:     "override permission from template",
			repoName: "my-api",
			settings: repoSettings{
				permission: "admin",
				from:       "template-go-api",
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "admin",
			wantTopics: []string{"backend"},
			wantError:  false,
		},
		{
			name:     "template not found",
			repoName: "my-api",
			settings: repoSettings{
				from: "nonexistent-template",
			},
			allRepos:      map[string]repoSettings{},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "not found",
		},
		{
			name:     "referenced repo not marked as template",
			repoName: "my-api",
			settings: repoSettings{
				from: "regular-repo",
			},
			allRepos: map[string]repoSettings{
				"regular-repo": {
					permission: "push",
					template:   false,
				},
			},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "not marked with template: true",
		},
		{
			name:     "cross-org template reference",
			repoName: "my-api",
			settings: repoSettings{
				from: "other-org/template-repo",
			},
			allRepos:      map[string]repoSettings{},
			defaultOrg:    "myorg",
			wantError:     true,
			errorContains: "cross-organization template references not yet supported",
		},
		{
			name:     "deduplicate topics",
			repoName: "my-api",
			settings: repoSettings{
				from:   "template-go-api",
				topics: []string{"backend", "my-service"},
			},
			allRepos: map[string]repoSettings{
				"template-go-api": {
					permission: "push",
					topics:     []string{"backend", "api"},
					template:   true,
				},
			},
			defaultOrg: "myorg",
			wantPerm:   "push",
			wantTopics: []string{"backend", "api", "my-service"},
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveTemplate(tt.repoName, tt.settings, tt.allRepos, tt.defaultOrg)

			if (err != nil) != tt.wantError {
				t.Errorf("resolveTemplate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if result.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", result.permission, tt.wantPerm)
			}

			if len(result.topics) != len(tt.wantTopics) {
				t.Errorf("topics = %v, want %v", result.topics, tt.wantTopics)
			} else {
				for i, topic := range result.topics {
					if topic != tt.wantTopics[i] {
						t.Errorf("topics[%d] = %q, want %q", i, topic, tt.wantTopics[i])
					}
				}
			}
		})
	}
}

func TestValidateTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "valid topic",
			topic:     "backend",
			wantError: false,
		},
		{
			name:      "valid topic with hyphens",
			topic:     "my-project-backend",
			wantError: false,
		},
		{
			name:      "valid topic with numbers",
			topic:     "project123",
			wantError: false,
		},
		{
			name:      "empty topic",
			topic:     "",
			wantError: true,
		},
		{
			name:      "topic too long (>50 chars)",
			topic:     "this-is-a-very-long-topic-name-that-exceeds-fifty-characters-limit",
			wantError: true,
		},
		{
			name:      "topic starting with hyphen",
			topic:     "-invalid",
			wantError: true,
		},
		{
			name:      "topic with uppercase",
			topic:     "Backend",
			wantError: true,
		},
		{
			name:      "topic with underscore",
			topic:     "my_project",
			wantError: true,
		},
		{
			name:      "topic with space",
			topic:     "my project",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTopic(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("validateTopic() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNormalizePermission(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Standard permissions
		{
			name:  "pull permission",
			input: "pull",
			want:  "pull",
		},
		{
			name:  "read permission (normalized to pull)",
			input: "read",
			want:  "pull",
		},
		{
			name:  "triage permission",
			input: "triage",
			want:  "triage",
		},
		{
			name:  "push permission",
			input: "push",
			want:  "push",
		},
		{
			name:  "write permission (normalized to push)",
			input: "write",
			want:  "push",
		},
		{
			name:  "maintain permission",
			input: "maintain",
			want:  "maintain",
		},
		{
			name:  "admin permission",
			input: "admin",
			want:  "admin",
		},
		// Case insensitive
		{
			name:  "uppercase PUSH",
			input: "PUSH",
			want:  "push",
		},
		{
			name:  "mixed case Admin",
			input: "Admin",
			want:  "admin",
		},
		// Custom repository roles (GitHub Enterprise Cloud)
		{
			name:  "custom role: actions-manager",
			input: "actions-manager",
			want:  "actions-manager",
		},
		{
			name:  "custom role: release-manager",
			input: "release-manager",
			want:  "release-manager",
		},
		{
			name:  "custom role: runner-admin",
			input: "runner-admin",
			want:  "runner-admin",
		},
		{
			name:  "custom role: security-scanner",
			input: "security-scanner",
			want:  "security-scanner",
		},
		{
			name:  "custom role with mixed case (preserved)",
			input: "Custom-Role-Name",
			want:  "Custom-Role-Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePermission(tt.input)
			if got != tt.want {
				t.Errorf("normalizePermission(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRepoConfigWithCustomRoles(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantPerm  string
		wantError bool
	}{
		{
			name:      "custom role as simple string",
			input:     "actions-manager",
			wantPerm:  "actions-manager",
			wantError: false,
		},
		{
			name: "custom role in advanced config",
			input: map[string]any{
				"permission": "release-manager",
				"topics":     []any{"cicd", "releases"},
			},
			wantPerm:  "release-manager",
			wantError: false,
		},
		{
			name: "custom role with hyphens",
			input: map[string]any{
				"permission": "github-actions-admin",
				"topics":     []any{"cicd"},
			},
			wantPerm:  "github-actions-admin",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseRepoConfig(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseRepoConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return
			}

			if settings.permission != tt.wantPerm {
				t.Errorf("permission = %q, want %q", settings.permission, tt.wantPerm)
			}
		})
	}
}

func TestContainsErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		errResp     *github.ErrorResponse
		searchTerms []string
		want        bool
	}{
		{
			name: "message in main Message field",
			errResp: &github.ErrorResponse{
				Message: `"sha" wasn't supplied`,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in main Message field without quotes",
			errResp: &github.ErrorResponse{
				Message: `sha wasn't supplied`,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in Errors array",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: `"sha" wasn't supplied`},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "message in Errors array among multiple errors",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: "some other error"},
					{Message: `"sha" wasn't supplied`},
					{Message: "yet another error"},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        true,
		},
		{
			name: "reference already exists in main Message",
			errResp: &github.ErrorResponse{
				Message: "reference already exists",
			},
			searchTerms: []string{"reference already exists"},
			want:        true,
		},
		{
			name: "reference already exists in Errors array",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors: []github.Error{
					{Message: "reference already exists"},
				},
			},
			searchTerms: []string{"reference already exists"},
			want:        true,
		},
		{
			name: "partial match should fail",
			errResp: &github.ErrorResponse{
				Message: "sha is required",
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
		{
			name: "no match in Message or Errors",
			errResp: &github.ErrorResponse{
				Message: "some other error",
				Errors: []github.Error{
					{Message: "different error"},
				},
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
		{
			name: "empty ErrorResponse",
			errResp: &github.ErrorResponse{
				Message: "",
				Errors:  nil,
			},
			searchTerms: []string{"sha", "wasn't supplied"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsErrorMessage(tt.errResp, tt.searchTerms...)
			if got != tt.want {
				t.Errorf("containsErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---- Planning Function Tests ----

func TestPlanTeams(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Team: []config.TeamConfig{
			{Name: "Backend", Slug: "backend", Description: "Backend team", Privacy: "closed"},
			{Name: "Frontend", Slug: "frontend", Description: "New desc", Privacy: "closed"},
			{Name: "Infra", Slug: "infra"},
		},
	}
	st := &State{
		Org: "myorg",
		ActualTeams: []*github.Team{
			{ID: github.Ptr(int64(1)), Slug: github.Ptr("backend"), Name: github.Ptr("Backend"), Description: github.Ptr("Backend team"), Privacy: github.Ptr("closed")},
			{ID: github.Ptr(int64(2)), Slug: github.Ptr("frontend"), Name: github.Ptr("Frontend"), Description: github.Ptr("Old desc"), Privacy: github.Ptr("closed")},
		},
	}

	changes, desired, err := planTeams(context.Background(), nil, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(desired) != 3 {
		t.Errorf("expected 3 desired teams, got %d", len(desired))
	}

	var creates, updates int
	for _, ch := range changes {
		switch ch.Action {
		case "create":
			creates++
			if ch.Target != "infra" {
				t.Errorf("expected create for infra, got %s", ch.Target)
			}
		case "update":
			updates++
			if ch.Target != "frontend" {
				t.Errorf("expected update for frontend, got %s", ch.Target)
			}
		}
	}
	if creates != 1 {
		t.Errorf("expected 1 create, got %d", creates)
	}
	if updates != 1 {
		t.Errorf("expected 1 update, got %d", updates)
	}
}

func TestPlanTeamMembership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/myorg/teams/backend/members" && r.URL.Query().Get("role") == "maintainer":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"login": "alice"},
			})
		case r.URL.Path == "/orgs/myorg/teams/backend/members" && r.URL.Query().Get("role") == "member":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"login": "bob"},
			})
		case r.URL.Path == "/users/alice" || r.URL.Path == "/users/bob" || r.URL.Path == "/users/charlie":
			_ = json.NewEncoder(w).Encode(map[string]any{"login": strings.TrimPrefix(r.URL.Path, "/users/")})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	st := &State{Org: "myorg"}
	desiredBySlug := map[string]config.TeamConfig{
		"backend": {
			Name:        "Backend",
			Slug:        "backend",
			Maintainers: []string{"alice"},
			Members:     []string{"charlie"}, // bob removed, charlie added
		},
	}

	changes, err := planTeamMembership(context.Background(), c, st, desiredBySlug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have a change for charlie (new member)
	found := false
	for _, ch := range changes {
		if ch.Scope == "team-member" && ch.Action == "ensure" {
			d := ch.Details.(teamMemberChange)
			if d.User == "charlie" && d.Role == "member" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected team-member:ensure change for charlie")
	}
}

func TestPlanRepoPerms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/orgs/myorg/teams/") && strings.HasSuffix(r.URL.Path, "/repos"):
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg", CreateRepo: true},
		Team: []config.TeamConfig{
			{
				Name: "Backend",
				Slug: "backend",
				Repositories: map[string]any{
					"api": map[string]any{
						"permission": "push",
						"topics":     []any{"backend", "go"},
					},
					"new-service": "maintain",
				},
			},
		},
	}
	st := &State{
		Org: "myorg",
		ActualRepos: []*github.Repository{
			{Name: github.Ptr("api"), Topics: []string{"backend"}},
		},
	}

	changes, err := planRepoPerms(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var repoEnsures, teamRepoGrants, topicChanges int
	for _, ch := range changes {
		switch ch.Scope + ":" + ch.Action {
		case "repo:ensure":
			repoEnsures++
		case "team-repo:grant":
			teamRepoGrants++
		case "repo-topics:ensure":
			topicChanges++
		}
	}
	if repoEnsures != 1 {
		t.Errorf("expected 1 repo:ensure (new-service), got %d", repoEnsures)
	}
	if teamRepoGrants != 2 {
		t.Errorf("expected 2 team-repo:grant, got %d", teamRepoGrants)
	}
	if topicChanges != 1 {
		t.Errorf("expected 1 repo-topics:ensure, got %d", topicChanges)
	}
}

func TestPlanCleanups(t *testing.T) {
	cfg := &config.Root{
		App: config.AppConfig{
			Org:                     "myorg",
			DeleteUnconfiguredTeams: true,
			DeleteUnmanagedRepos:    true,
		},
	}
	desired := map[string]config.TeamConfig{
		"backend": {Name: "Backend", Slug: "backend"},
	}
	st := &State{
		Org:          "myorg",
		ManagedRepos: map[string]bool{"api": true},
		ActualTeams: []*github.Team{
			{ID: github.Ptr(int64(1)), Slug: github.Ptr("backend")},
			{ID: github.Ptr(int64(2)), Slug: github.Ptr("old-team")},
		},
		ActualRepos: []*github.Repository{
			{Name: github.Ptr("api")},
			{Name: github.Ptr("legacy-app")},
		},
	}

	changes, _, err := planCleanups(context.Background(), nil, cfg, st, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var teamDeletes, repoDeletes int
	for _, ch := range changes {
		switch ch.Scope + ":" + ch.Action {
		case "team:delete":
			teamDeletes++
			if ch.Target != "old-team" {
				t.Errorf("expected delete for old-team, got %s", ch.Target)
			}
		case "repo:delete":
			repoDeletes++
			if ch.Target != "legacy-app" {
				t.Errorf("expected delete for legacy-app, got %s", ch.Target)
			}
		}
	}
	if teamDeletes != 1 {
		t.Errorf("expected 1 team:delete, got %d", teamDeletes)
	}
	if repoDeletes != 1 {
		t.Errorf("expected 1 repo:delete, got %d", repoDeletes)
	}
}

func TestApplyChanges_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	changes := []util.Change{
		{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg", "name": "Backend"}},
	}

	err := applyChanges(ctx, nil, changes)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestPlanCustomRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/custom-repository-roles" && r.Method == "GET" {
			resp := map[string]any{
				"total_count": 1,
				"custom_roles": []map[string]any{
					{
						"id":          1,
						"name":        "deployer",
						"description": "Old desc",
						"base_role":   "read",
						"permissions": []string{"manage_actions"},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{Org: "myorg"},
		Org: config.OrgConfig{
			CustomRoles: []config.CustomRoleConfig{
				{Name: "deployer", Description: "Updated desc", BaseRole: "read", Permissions: []string{"manage_actions"}},
				{Name: "release-manager", Description: "New role", BaseRole: "write", Permissions: []string{"create_releases"}},
			},
		},
	}
	st := &State{Org: "myorg"}

	changes, err := planCustomRoles(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var creates, updates int
	for _, ch := range changes {
		switch ch.Action {
		case "create":
			creates++
		case "update":
			updates++
		}
	}
	if creates != 1 {
		t.Errorf("expected 1 custom-role:create, got %d", creates)
	}
	if updates != 1 {
		t.Errorf("expected 1 custom-role:update, got %d", updates)
	}
}
````

## File: cmd/sync.go
````go
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/DragonSecurity/gomgr/internal/util"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	Example: `  gomgr sync -c ./config
  gomgr sync -c ./config --dry
  gomgr sync -c ./config --timeout 5m --audit-log`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfgDir == "" {
			return fmt.Errorf("--config/-c flag is required")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if debug {
			util.EnableDebug()
		}
		util.AuditLog = auditLog

		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		client, appInfo, err := gh.NewClientFromEnv(ctx, cfg.App)
		if err != nil {
			return err
		}
		if appInfo != "" {
			util.Infof("auth: %s", appInfo)
		}

		plan, err := insync.BuildPlan(ctx, client, cfg)
		if err != nil {
			return err
		}

		if err := util.PrintPlan(plan); err != nil {
			return fmt.Errorf("print plan: %w", err)
		}

		if dryRun {
			util.PrintSummary(plan)
			util.Infof("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, plan)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
````

## File: config/example/.github/workflows/sync.yaml
````yaml
name: Synchronise organization users and teams (gomgr)

on:
  push:
    branches: [ "main" ]
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-24.04
    continue-on-error: true
    strategy:
      max-parallel: 5
      fail-fast: false
      matrix:
        config:
          - { folder: "config/example", gom_version: "v0.0.3" }

    env:
      GH_TOKEN: ${{ github.token }}   # for gh release download

    steps:
      - name: Checkout repository
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Determine platform
        id: plat
        run: |
          echo "os=linux"   >> $GITHUB_OUTPUT
          echo "arch=amd64" >> $GITHUB_OUTPUT

      - name: Download gomgr binary from releases
        run: |
          VERSION="${{ matrix.config.gom_version }}"
          OS="${{ steps.plat.outputs.os }}"
          ARCH="${{ steps.plat.outputs.arch }}"
          ASSET_TGZ="gomgr_${VERSION}_${OS}_${ARCH}.tar.gz"
          ASSET_ZIP="gomgr_${VERSION}_${OS}_${ARCH}.zip"

          mkdir -p .gomgr
          gh release download "$VERSION" --repo DragonSecurity/gomgr --pattern "$ASSET_TGZ"                 --dir .gomgr || true

          if [ ! -f ".gomgr/$ASSET_TGZ" ]; then
            gh release download "$VERSION" --repo DragonSecurity/gomgr --pattern "$ASSET_ZIP"                   --dir .gomgr
          fi

          if [ -f ".gomgr/$ASSET_TGZ" ]; then
            tar -xzf ".gomgr/$ASSET_TGZ" -C .gomgr
          else
            unzip -o ".gomgr/$ASSET_ZIP" -d .gomgr
          fi

          GOMGR_PATH=$(find .gomgr -type f -name "gomgr" -o -name "gomgr.exe" | head -n1)
          sudo mv "$GOMGR_PATH" /usr/local/bin/gomgr
          sudo chmod +x /usr/local/bin/gomgr

      - name: Show gomgr version
        run: gomgr version

      # - name: Dry-run sync
      #   run: gomgr sync -c ${{ matrix.config.folder }} --dry

      - name: Synchronise settings
        run: gomgr sync -c ${{ matrix.config.folder }}
        env:
          GITHUB_APP_PRIVATE_KEY: ${{ secrets.DSEC_USER_MANAGEMENT_APP_PRIVATE_KEY }}
          GITHUB_APP_ID: "1719369"
````

## File: internal/sync/orchestrator.go
````go
package sync

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

type State struct {
	Org          string
	ManagedRepos map[string]bool

	// Cached API results to avoid duplicate calls
	ActualTeams []*github.Team
	ActualRepos []*github.Repository

	// Current state from GitHub
	CurrentTeams       int
	CurrentTeamMembers int
	CurrentRepos       int
	CurrentRepoPerms   int
	CurrentCustomRoles int

	// Desired state from config
	DesiredTeams       int
	DesiredTeamMembers int
	DesiredRepos       int
	DesiredRepoPerms   int
	DesiredCustomRoles int
}

func BuildPlan(ctx context.Context, c *gh.Client, cfg *config.Root) (util.Plan, error) {
	st := &State{Org: cfg.App.Org}
	var plan util.Plan

	// Prefetch teams and repos once to avoid duplicate API calls
	if err := prefetchState(ctx, c, st); err != nil {
		return plan, fmt.Errorf("prefetch state: %w", err)
	}

	// Custom roles must be created before teams/repos use them
	customRoleChanges, err := planCustomRoles(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan custom roles: %w", err)
	}

	teamChanges, desiredBySlug, err := planTeams(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan teams: %w", err)
	}

	memChanges, err := planTeamMembership(ctx, c, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan team membership: %w", err)
	}

	repoChanges, err := planRepoPerms(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan repo permissions: %w", err)
	}

	cleanupChanges, warnings, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, fmt.Errorf("plan cleanups: %w", err)
	}

	customRoleCleanups, roleWarnings, err := planCustomRoleCleanups(ctx, c, cfg, st)
	if err != nil {
		return plan, fmt.Errorf("plan custom role cleanups: %w", err)
	}

	plan.Changes = append(plan.Changes, customRoleChanges...)
	plan.Changes = append(plan.Changes, teamChanges...)
	plan.Changes = append(plan.Changes, memChanges...)
	plan.Changes = append(plan.Changes, repoChanges...)
	plan.Changes = append(plan.Changes, cleanupChanges...)
	plan.Changes = append(plan.Changes, customRoleCleanups...)
	plan.Warnings = append(warnings, roleWarnings...)

	// Populate stats
	plan.Stats = &util.StateStats{
		Teams: util.StatePair{
			Current: st.CurrentTeams,
			Desired: st.DesiredTeams,
		},
		TeamMembers: util.StatePair{
			Current: st.CurrentTeamMembers,
			Desired: st.DesiredTeamMembers,
		},
		Repositories: util.StatePair{
			Current: st.CurrentRepos,
			Desired: st.DesiredRepos,
		},
		RepoPermissions: util.StatePair{
			Current: st.CurrentRepoPerms,
			Desired: st.DesiredRepoPerms,
		},
		CustomRoles: util.StatePair{
			Current: st.CurrentCustomRoles,
			Desired: st.DesiredCustomRoles,
		},
	}

	return plan, nil
}

// prefetchState fetches teams and repos from GitHub once, caching them in State
// so that both planning and cleanup phases can reuse the data.
func prefetchState(ctx context.Context, c *gh.Client, st *State) error {
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, st.Org, opts)
		if err != nil {
			return nil, err
		}
		st.ActualTeams = append(st.ActualTeams, ts...)
		return resp, nil
	}); err != nil {
		return fmt.Errorf("list teams: %w", err)
	}

	repoOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
		Type:        "all",
	}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		repoOpt.ListOptions = *opts
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, st.Org, repoOpt)
		if err != nil {
			return nil, err
		}
		st.ActualRepos = append(st.ActualRepos, repos...)
		return resp, nil
	}); err != nil {
		return fmt.Errorf("list repos: %w", err)
	}

	return nil
}

func Apply(ctx context.Context, c *gh.Client, plan util.Plan) error {
	return applyChanges(ctx, c, plan.Changes)
}
````

## File: .github/workflows/ci.yaml
````yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

permissions:
  contents: read

jobs:
  test:
    name: Test
    runs-on: ubuntu-24.04
    strategy:
      matrix:
        go-version: ['1.26.2']

    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@668228422ae6a00e4ad889ee87cd7109ec5666a7 # v5
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: make mod-download

      - name: Verify dependencies
        run: make mod-verify

      - name: Run tests
        run: make test-coverage

      - name: Upload coverage to artifacts
        uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7
        with:
          name: coverage-report
          path: coverage/
          if-no-files-found: error

  lint:
    name: Lint
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6
        with:
          go-version: '1.26.2'

      - name: Cache Go modules
        uses: actions/cache@668228422ae6a00e4ad889ee87cd7109ec5666a7 # v5
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Check formatting
        run: make fmt-check

      - name: Run go vet
        run: make vet

      - name: Install golangci-lint
        run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.0
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      - name: Run golangci-lint
        run: make lint

  security:
    name: Security
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6
        with:
          go-version: '1.26.2'

      - name: Cache Go modules
        uses: actions/cache@668228422ae6a00e4ad889ee87cd7109ec5666a7 # v5
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install gosec
        run: go install github.com/securego/gosec/v2/cmd/gosec@latest

      - name: Run security checks
        run: make security

  build:
    name: Build
    runs-on: ubuntu-24.04
    needs: [test, lint, security]
    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6
        with:
          go-version: '1.26.2'

      - name: Cache Go modules
        uses: actions/cache@668228422ae6a00e4ad889ee87cd7109ec5666a7 # v5
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Build
        run: make build

      - name: Upload binary
        uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7
        with:
          name: gomgr-binary
          path: build/gomgr
          if-no-files-found: error
````

## File: internal/gh/client.go
````go
package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v84/github"
	"golang.org/x/oauth2"

	"github.com/DragonSecurity/gomgr/internal/config"
)

type Client struct {
	REST       *github.Client
	httpClient *http.Client
	// GraphQLURL is the endpoint used by DoGraphQL. Empty means GitHub's public
	// GraphQL API. Tests may override it to point at a local server.
	GraphQLURL string
}

const defaultMaxRetries = 3
const defaultGraphQLURL = "https://api.github.com/graphql"

func NewClientFromEnv(ctx context.Context, app config.AppConfig) (*Client, string, error) {
	// PAT
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		tc.Transport = newRetryTransport(tc.Transport, defaultMaxRetries)
		return &Client{REST: github.NewClient(tc), httpClient: tc}, "PAT", nil
	}
	// App
	appID := app.AppID
	if v := os.Getenv("GITHUB_APP_ID"); v != "" && appID == 0 {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			appID = id
		}
	}
	key := firstNonEmpty(app.PrivateKey, os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	if appID == 0 || key == "" {
		return nil, "", errors.New("no auth found: set GITHUB_TOKEN or app_id+private_key")
	}
	pemBytes, err := maybeReadPEM(key)
	if err != nil {
		return nil, "", err
	}
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, pemBytes)
	if err != nil {
		return nil, "", fmt.Errorf("app transport: %w", err)
	}
	tmp := github.NewClient(&http.Client{Transport: atr})
	inst, _, err := tmp.Apps.FindOrganizationInstallation(ctx, app.Org)
	if err != nil {
		return nil, "", fmt.Errorf("find installation for org %q: %w", app.Org, err)
	}
	itr := ghinstallation.NewFromAppsTransport(atr, inst.GetID())
	httpClient := &http.Client{Transport: newRetryTransport(itr, defaultMaxRetries), Timeout: 30 * time.Second}
	return &Client{REST: github.NewClient(httpClient), httpClient: httpClient}, "Github App", nil
}

func maybeReadPEM(s string) ([]byte, error) {
	var (
		data   []byte
		source string
	)
	if strings.Contains(s, "BEGIN") {
		data = []byte(s)
		source = "inline key"
	} else {
		b, err := os.ReadFile(s)
		if err != nil {
			return nil, err
		}
		data = b
		source = s
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM at %s", source)
	}
	if !isPrivateKeyBlockType(block.Type) {
		return nil, fmt.Errorf("invalid PEM at %s: expected a private key block, got %q", source, block.Type)
	}
	return data, nil
}

func isPrivateKeyBlockType(t string) bool {
	switch t {
	case "RSA PRIVATE KEY", "PRIVATE KEY", "EC PRIVATE KEY":
		return true
	}
	return false
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// DoGraphQL executes a GraphQL query or mutation
func (c *Client) DoGraphQL(ctx context.Context, query string, variables map[string]any, result any) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf("graphql client httpClient is nil")
	}
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("graphql query must not be empty")
	}
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}

	reqBody := map[string]any{
		"query": query,
	}
	if len(variables) > 0 {
		reqBody["variables"] = variables
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal graphql request: %w", err)
	}

	url := c.GraphQLURL
	if url == "" {
		url = defaultGraphQLURL
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute graphql request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("graphql request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to check for GraphQL errors
	var gqlResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Path    []any  `json:"path,omitempty"`
		} `json:"errors"`
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read graphql response: %w", err)
	}

	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("graphql error: %s", strings.Join(msgs, "; "))
	}

	if result != nil && len(gqlResp.Data) > 0 {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("decode graphql data: %w", err)
		}
	}

	return nil
}
````

## File: README.md
````markdown
# github-org-manager-go (gomgr)

[![Build](https://github.com/DragonSecurity/gomgr/actions/workflows/release.yaml/badge.svg)](https://github.com/DragonSecurity/gomgr/actions/workflows/release.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/DragonSecurity/gomgr.svg)](https://pkg.go.dev/github.com/DragonSecurity/gomgr)
[![Go Report Card](https://goreportcard.com/badge/github.com/DragonSecurity/gomgr)](https://goreportcard.com/report/github.com/DragonSecurity/gomgr)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/DragonSecurity/gomgr)](https://github.com/DragonSecurity/gomgr/releases)
[![GitHub issues](https://img.shields.io/github/issues/DragonSecurity/gomgr)](https://github.com/DragonSecurity/gomgr/issues)

A fast, idempotent **GitHub Organization Manager** written in Go. Define your org as YAML and apply it with a single command. Ships with a release workflow and a CI workflow to run sync against one or many org-config folders.

## Highlights

- ✅ YAML-driven org config (`app.yaml`, `org.yaml`, `teams/*.yaml`)
- ✅ Teams, maintainers, members (idempotent add/update)
- ✅ Repo permission grants (pull/triage/push/maintain/admin)
- ✅ **Custom repository roles**: fully managed - define in YAML, gomgr creates/updates them (GitHub Enterprise Cloud)
- ✅ **Repository topics**: add topics/labels to repositories for organization
- ✅ **Repository pinning**: pin important repositories to organization profile (⚠️ *GitHub API limitation: not currently supported for organizations - configuration accepted but manual pinning required via web UI*)
- ✅ **Optional**: create repos that don’t exist (`create_repo: true`)
- ✅ **Optional**: inject `.github/renovate.json` into repos
- ✅ Warnings & cleanups: unmanaged teams, members without team, unmanaged repos, unmanaged custom roles
- ✅ **Optional** hard cleanups: delete unmanaged teams, remove members without team, delete unmanaged repos, delete unmanaged custom roles
- ✅ Auth: GitHub App (recommended) or PAT
- ✅ `--dry` plan with **state comparison** showing current GitHub state vs desired config state
- ✅ Cross‑platform binaries via GitHub Releases; `gomgr version` stamped at build

---

## Install

### Option 1 — Download a release
Grab a binary from Releases and put it on your `$PATH`.
```bash
gomgr version
```

### Option 2 — Build from source
```bash
go build -trimpath -buildvcs=true -ldflags "-s -w -X github.com/DragonSecurity/github-org-manager-go/internal/version.Version=$(git describe --tags --always --dirty)" -o gomgr .
```

---

## Quickstart

1. **Prepare config directory** (example shown below):
```
<config>/
├─ app.yaml
├─ org.yaml
└─ teams/
   └─ platform-team.yaml
```

2. **Auth** (choose one):

- **GitHub App** (recommended)
  ```bash
  export GITHUB_APP_ID=1719369
  export GITHUB_APP_PRIVATE_KEY="$(cat /path/to/private-key.pem)"   # or set app.yaml: private_key: <path or PEM>
  ```
  *Installation*: install the app on your target org and grant the permissions listed below.

- **PAT**
  ```bash
  export GITHUB_TOKEN=<personal-access-token>
  ```

3. **Run a dry run, then apply**
```bash
gomgr sync -c <config> --dry  # Shows JSON plan + summary of changes
gomgr sync -c <config>         # Actually applies changes
```

The dry run output includes:
- Complete JSON plan with all change details
- **Current vs Desired State comparison** - shows what exists in GitHub vs what's in your config
- Summary showing counts by scope and action
- List of any warnings

**Example summary output:**
```
================================================================
Summary of Proposed Changes
================================================================

Current State vs Desired State:
--------------------------------
  Teams:              5 → 6 (+1)
  Team Members:       12 → 14 (+2)
  Repositories:       15 → 18 (+3)
  Repo Permissions:   22 → 28 (+6)

Total changes: 7

Changes by scope:
  repo-file:           3
  repo-pin:            1
  repo-topics:         1
  team-repo:           3

Changes by action:
  ensure:              5
  grant:               3

Warnings: 1
  - Skipping pin for KaMuses/platform-index: GitHub API does not support pinning to organization profiles

================================================================
```

---

## Configuration

### `app.yaml`
```yaml
org: KaMuses
# GitHub App auth (preferred):
app_id: 1719369                 # or set via env GITHUB_APP_ID
private_key: ./app-private.pem  # file path or raw PEM; env GITHUB_APP_PRIVATE_KEY also works

dry_warnings:
  warn_unmanaged_teams: true
  warn_members_without_any_team: true
  warn_unmanaged_repos: true         # warn about repos not defined in any team
  warn_unmanaged_custom_roles: true  # warn about custom roles not in org.yaml

# Optional enforcement / extras:
remove_members_without_team: true   # remove org members not in any team
delete_unconfigured_teams: true     # delete teams not defined in YAML
delete_unmanaged_repos: false       # delete repos not defined in any team (DESTRUCTIVE!)
delete_unmanaged_custom_roles: false # delete custom roles not in org.yaml (DESTRUCTIVE!)
create_repo: true                   # create repos if missing when referenced by teams
add_renovate_config: true           # create .github/renovate.json in repos
add_default_readme: false           # create default README.md in repos (optional)
renovate_config: |
  {
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }
```

### `org.yaml`
Define organization owners and custom repository roles:
```yaml
owners:
  - alice
  - bob

# Custom repository roles (requires GitHub Enterprise Cloud)
# gomgr will create/update these roles automatically
custom_roles:
  - name: actions-manager
    description: Manage GitHub Actions workflows and runners
    base_role: read  # read, triage, write, maintain, admin
    permissions:
      - write_actions
      - read_actions_variables
      - write_actions_variables
      
  - name: release-manager
    description: Create and manage releases
    base_role: write
    permissions:
      - create_releases
      - edit_releases
      - manage_environments
```

**Available Permissions** (partial list - see [GitHub Docs](https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/custom-roles#list-repository-fine-grained-permissions-for-an-organization) for full list):
- Actions: `write_actions`, `read_actions_variables`, `write_actions_variables`
- Releases: `create_releases`, `edit_releases`, `delete_releases`
- Environments: `manage_environments`, `read_deployment_environments`, `write_deployment_environments`
- Runners: `admin_self_hosted_runners`, `read_self_hosted_runners`
- Security: `read_code_scanning_alerts`, `write_code_scanning_alerts`, `read_secret_scanning_alerts`, `write_secret_scanning_alerts`
- And many more...


### `teams/*.yaml`
```yaml
name: Platform Team
slug: platform-team            # optional; default = kebab(name)
description: Core platform engineers
privacy: closed                # closed | secret
parents: []                    # (future enhancement)

# Multiple maintainers (team leads, senior engineers)
maintainers:
  - alice-backend-lead
  - bob-senior-engineer
  - charlie-tech-lead

# Multiple members (regular team members)
members:
  - david-developer
  - emma-engineer
  - frank-junior-dev
  - grace-contractor

repositories:
  # Simple permission string (backward compatible)
  # Built-in roles: pull|triage|push|maintain|admin
  infra: maintain
  
  # Advanced config with topics and pinning
  api:
    permission: push
    topics:
      - backend
      - api
      - project-platform
  
  # Custom repository roles (requires GitHub Enterprise Cloud)
  # Custom roles allow fine-grained permissions like managing GitHub Actions
  # without full repository admin access. Custom roles must be defined in org.yaml
  # and gomgr will create/update them automatically.
  # See: https://docs.github.com/en/enterprise-cloud@latest/organizations/managing-user-access-to-your-organizations-repositories/managing-repository-roles/managing-custom-repository-roles-for-an-organization
  ci-workflows:
    permission: actions-manager  # Custom role name (must be defined in org.yaml)
    topics:
      - cicd
      - github-actions
  
  # Template repository - can be reused by other repos
  # Mark a repository as a template with `template: true`
  # Templates are marked in GitHub and can be inherited by other repos in config
  template-go-api:
    permission: push
    template: true
    topics:
      - backend
      - api
      - go-template
  
  # Repository using template (inherits permission and topics)
  my-api:
    from: template-go-api      # Reference to template repo (currently only same-org supported)
    topics:
      - my-project             # Additional topics (merged with template topics)
    # Will inherit: permission: push, and topics: backend, api, go-template
  
  # Repository using template with override
  admin-api:
    from: template-go-api
    permission: admin          # Override template permission
    topics:
      - admin-service
    # Will inherit topics: backend, api, go-template from template
  
  # Repository with pinning (note: pinning is not supported by GitHub API for organizations)
  platform-index:
    permission: admin
    topics:
      - project-platform
      - documentation
    pinned: true  # Will be shown in plan but skipped with a warning - pin manually via GitHub web UI
```

> Loader ignores non‑YAML files in `teams/` and skips empty/invalid entries.

---

## Extended Team Examples

The `examples/config/teams/` directory includes comprehensive team definition examples demonstrating various organizational patterns:

### Example Teams

**Backend Team** (`backend-team.yaml`)
- Multiple maintainers (team leads, senior engineers)
- Multiple members (developers, contractors, interns)
- Demonstrates different permission levels (admin, push, maintain, triage, pull)
- Shows diverse repository types (APIs, microservices, libraries, documentation)

**Frontend Team** (`frontend-team.yaml`)
- Cross-functional team with specialized roles (React, Vue, UX, accessibility)
- Web and mobile application management
- Shared component libraries and design systems

**DevOps Team** (`devops-team.yaml`)
- Infrastructure and CI/CD management
- Terraform, Kubernetes, and cloud configurations
- Monitoring, security, and automation repositories

**Security Team** (`security-team.yaml`)
- Uses `privacy: secret` for sensitive access
- Read access to multiple repos for security audits
- Admin access to security-specific repositories
- Compliance and vulnerability management

**GitHub Actions Team** (`github-actions-team.yaml`)
- **Demonstrates custom repository roles** (requires GitHub Enterprise Cloud)
- Shows how to use fine-grained permissions for CI/CD management
- Examples of custom roles: `actions-manager`, `release-manager`, `runner-admin`, `security-scanner`

### Best Practices Demonstrated

1. **Multiple Maintainers**: Include multiple team leads to avoid single points of failure
2. **Diverse Membership**: Mix senior engineers, regular developers, contractors, and interns
3. **Descriptive Privacy**: Use `closed` for most teams, `secret` for sensitive security teams
4. **Clear Descriptions**: Write meaningful team descriptions for easy discovery
5. **Permission Hierarchy**: Use appropriate permission levels based on responsibility
6. **Topic Organization**: Tag repositories with relevant topics for discoverability
7. **Custom Roles**: Leverage fine-grained permissions for specialized access patterns

---

## Template Repository Pattern

gomgr supports marking repositories as templates and referencing them from other repositories. This enables consistent configuration across multiple repositories:

**Template Repository Features:**
- Mark a repository as a template with `template: true`
- Template repositories can define permission and topics that other repos inherit
- Reference templates using `from: template-repo-name` (same-org only - cross-org not yet supported)
- New repositories with `from:` are created using GitHub's template repository feature
- Topics are automatically merged (template topics + repo-specific topics)
- Permissions can be inherited or overridden
- Templates are marked using the GitHub API's template repository flag

**How it works:**
1. Define a template repository with `template: true`
2. Other repositories reference it with `from: template-name`
3. When creating a new repo with `from:`, GitHub's CreateFromTemplate API is used
4. The referencing repo inherits permission (if not specified) and topics from the template
5. Add repo-specific topics to extend the template's topics
6. Override permission if needed for specific use cases

**Benefits:**
- Consistency across similar repositories (e.g., all microservices)
- DRY principle - define common configuration once
- Easy to update multiple repos by changing the template
- Clear relationships between repos in your configuration

**Limitations:**
- Currently only supports same-organization templates
- Cross-organization template references are not yet supported

---

## Custom Repository Roles Management

**Requires GitHub Enterprise Cloud**

gomgr now fully manages GitHub's custom repository roles, automatically creating and updating them based on your configuration. Custom roles allow fine-grained permissions beyond the standard roles (pull, triage, push, maintain, admin).

**Key Features:**
- **Automated Role Management**: Define roles in `org.yaml` and gomgr creates/updates them automatically
- **Fine-grained permissions**: Grant access to specific capabilities (Actions, runners, secrets, environments)
- **Separation of concerns**: Allow CI/CD management without code modification access
- **Idempotent updates**: Roles are kept in sync with your configuration
- **Optional cleanup**: Warn about or delete unmanaged custom roles

**Configuration Workflow:**

1. **Define roles in `org.yaml`**:
   ```yaml
   custom_roles:
     - name: actions-manager
       description: Manage CI/CD workflows
       base_role: read
       permissions:
         - write_actions
         - read_actions_variables
         - write_actions_variables
   ```

2. **Use role names in team configurations**:
   ```yaml
   # teams/cicd-team.yaml
   repositories:
     ci-workflows:
       permission: actions-manager  # Custom role
       topics: [cicd, github-actions]
   ```

3. **Apply configuration** - gomgr will:
   - Create custom roles if they don't exist
   - Update roles if configuration changed
   - Warn about unmanaged roles (if configured)
   - Optionally delete unmanaged roles (if configured)

**Example Use Cases:**

- **Actions Manager**: Manage workflows, runners, and secrets without code access
- **Release Manager**: Create releases and manage deployment environments
- **Security Scanner**: Configure security scanning without repository admin access
- **Runner Admin**: Manage self-hosted runners for CI/CD infrastructure

**Order of Operations:**
1. Custom roles are created/updated first (before teams/repositories)
2. Teams can then use the custom roles in repository permissions
3. Custom roles are deleted last (if cleanup is enabled)

**Configuration Options (app.yaml):**
```yaml
dry_warnings:
  warn_unmanaged_custom_roles: true  # Warn about roles not in config

delete_unmanaged_custom_roles: false  # Delete roles not in config (DESTRUCTIVE!)
```

**Important Notes:**
- Custom roles require GitHub Enterprise Cloud
- Role creation requires "Custom repository roles" (write) or "Administration" (write) permission
- Once created, roles can be assigned to teams just like built-in roles
- See `examples/config/org.yaml` for complete examples

---

## Project Organization Pattern

gomgr supports organizing repositories by project with topics, pinning, and naming conventions:

> **Note**: Repository pinning is not currently supported by the GitHub API for organization profiles. The `pinned` field is accepted in configuration but the actual pinning operation will be skipped with a warning. You can manually pin repositories through the GitHub web interface.

**Example: Multi-repo project setup**

1. Define a project name (slug), e.g., `platform`
2. Prefix all project repositories: `platform-api`, `platform-web`, `platform-infra`
3. Tag all repos with topic: `project-platform`
4. Create an index repository: `platform-index` with README linking to all project repos
5. Pin the index repo to make it prominent on the org profile (must be done manually via GitHub web UI due to API limitations)

**Example configuration:**

```yaml
name: Platform Team
repositories:
  platform-index:
    permission: admin
    topics:
      - project-platform
      - documentation
    pinned: true
  
  platform-api:
    permission: push
    topics:
      - project-platform
      - backend
  
  platform-web:
    permission: push
    topics:
      - project-platform
      - frontend
  
  platform-infra:
    permission: maintain
    topics:
      - project-platform
      - infrastructure
```

This pattern makes it easy to:
- Discover all repositories belonging to a project using GitHub's topic search
- Provide project documentation via the index repository (can be manually pinned via GitHub web UI)
- Maintain consistent naming and organization across projects

---

## Auth & Permissions

### GitHub App (recommended)
Set `GITHUB_APP_ID` and `GITHUB_APP_PRIVATE_KEY` (or `app_id`/`private_key` in `app.yaml`). The app must be installed on the org.

**Required Organization Permissions:**

Core features:
- **Members**: Read/Write - manage org members
- **Administration**: Read/Write - manage teams and repositories
  - Or minimum: **Teams**: Read/Write + **Repository**: Administration (Read/Write)
- **Metadata**: Read - read org metadata

Custom Repository Roles (if using this feature):
- **Custom repository roles**: Read/Write - create and manage custom roles
  - Alternative: **Administration**: Read/Write (includes custom roles)

**Required Repository Permissions:**
- **Administration**: Read/Write - grant team access, create repos, mark templates
- **Contents**: Read/Write - create files (if using Renovate config injection or default README)
- **Metadata**: Read - read repository metadata

> **Note**: If you don't use certain features (e.g., creating repos, custom roles, file injection), you can reduce permissions accordingly.

### Personal Access Token (PAT)
Use a classic PAT with scopes:
- `admin:org` - manage teams, members, and custom repository roles
- `repo` - set team repo access, create repos, and manage repository settings
- `read:org` - read org metadata

**Fine-grained PAT** (alternative):
- Organization permissions: Administration (Read/Write), Custom repository roles (Read/Write)
- Repository permissions: Administration (Read/Write), Contents (Read/Write)

---

## CLI

- `gomgr sync -c <config> [--dry] [--debug]`  
  Plans and applies org state. With `--dry`, shows a JSON plan followed by a human-readable summary of proposed changes without applying them.

- `gomgr setup-team -n "Team Name" -c <config> [-f out/path.yaml]`  
  Bootstraps a team YAML.

- `gomgr version`  
  Prints version (stamped at build). If built with VCS info, also prints revision/dirty/commit time.

**Order of operations** (apply):  
create custom roles → create teams → set memberships → ensure repos → mark templates → grant permissions → write files (renovate/readme) → set topics → pin repos → cleanups (optional) → delete custom roles (optional)

---

## CI: Releases

This repo includes `.github/workflows/release.yml`:

- Trigger on tag push (`v*.*.*`) or manual.
- Builds for linux/darwin/windows × amd64/arm64.
- Stamps `gomgr version` via `-ldflags`.
- Uploads packaged artifacts and `checksums.txt`.

**Tag to release:**
```bash
git tag v0.1.0
git push origin v0.1.0
```

---

## CI: Org sync workflow (for your org-config repo)

Example (`.github/workflows/org-sync.yml`):

```yaml
name: Synchronise organization users and teams (gomgr)

on:
  push:
    branches: [ "main" ]
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-24.04
    strategy:
      fail-fast: false
      matrix:
        config:
          - { folder: "dragonsecurity/dragonsecurity", gom_version: "v0.12.2" }
          - { folder: "dragonsecurity/dragondevcc",   gom_version: "v0.10.2" }
          - { folder: "dragonsecurity/kamuses",       gom_version: "v0.10.2" }
    continue-on-error: true  # allow other matrix jobs to finish if one fails

    steps:
      - uses: actions/checkout@v4
      - name: Install gh
        uses: cli/gh-action@v2
      - name: Download gomgr
        env: { GH_TOKEN: ${{ github.token }} }
        run: |
          VERSION="${{ matrix.config.gom_version }}"
          OS=linux ARCH=amd64
          ASSET="gomgr_${VERSION}_${OS}_${ARCH}.tar.gz"
          mkdir -p .gomgr
          gh release download "$VERSION" --repo DragonSecurity/gomgr --pattern "$ASSET" --dir .gomgr
          tar -xzf ".gomgr/$ASSET" -C .gomgr
          sudo mv $(find .gomgr -type f -name gomgr) /usr/local/bin/gomgr
          sudo chmod +x /usr/local/bin/gomgr
      - run: gomgr version
      - name: Synchronise settings
        run: gomgr sync -c ${{ matrix.config.folder }}
        env:
          GITHUB_APP_PRIVATE_KEY: ${{ secrets.DSEC_USER_MANAGEMENT_APP_PRIVATE_KEY }}
          GITHUB_APP_ID: "1719369"
```

---

## Development

### Testing

The project includes a comprehensive test suite and Makefile for development:

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run verbose tests
make test-verbose

# Check code formatting
make fmt-check

# Format code
make fmt

# Run go vet
make vet
```

### Code Quality Tools

Install development tools (golangci-lint, gosec):
```bash
make install-tools
```

Run code quality checks:
```bash
# Run linter
make lint

# Run security scanner
make security

# Run all basic checks (format, vet, test)
make check

# Run all checks including lint and security
make check-all
```

### Building

```bash
# Build binary
make build

# Clean build artifacts
make clean

# Run full CI pipeline
make ci
```

### CI/CD

The repository includes GitHub Actions workflows:
- **`.github/workflows/ci.yaml`**: Runs tests, linting, and security checks on every push/PR
- **`.github/workflows/release.yaml`**: Builds and releases binaries on version tags

---

## Troubleshooting

- **404 on `/teams//members`**: empty/invalid team YAML or calling membership on a team that doesn’t exist yet. Loader ignores non‑YAML files and planner guards empty slugs; team creation happens before membership.
- **`gomgr version` shows `dev`**: build without `-ldflags -X` or not from a tag. Use the release workflow or pass a version when building.
- **Renovate config not created**: ensure `add_renovate_config: true` and `renovate_config` is non‑empty; repo must exist or `create_repo: true`.
- **Repository pinning warnings**: The GitHub API does not support pinning repositories to organization profiles programmatically. The `pinned: true` configuration is accepted but the operation is skipped with a warning. You must manually pin repositories through the GitHub web interface.
- **Template reference not found**: ensure the template repository is defined in the same team configuration or another team file with `template: true` set. Cross-organization templates are not yet supported.
- **Custom role not found**: ensure custom roles are defined in `org.yaml` before using them in team repository permissions. Custom roles require GitHub Enterprise Cloud.

---

## Roadmap / TODO

- Compare & update team fields (description/privacy/parents)
- Optionally remove extra team members / revoke extra repo perms
- Optionally remove extra topics from repos (current behavior: union of all topics)
- Custom default branch for file writes
- Parallel apply with rate‑limit aware workers
- More comprehensive plan diff output

---

## Contributing

PRs welcome! Please:
- open an issue first for larger changes,
- keep commits small & focused,
- add tests where practical,
- run `make check` before submitting (or `make check-all` for full checks including linting),
- ensure `make build` succeeds.

See the **Development** section above for available commands and tooling.

---

## Security

This tool modifies org membership and repository access. Use **dry‑run** in CI and restrict credentials using least privilege. Prefer GitHub Apps over PATs.

---

## License

See **[LICENSE](./LICENSE.md)**.
````

## File: internal/sync/teams.go
````go
package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const defaultPerPage = 100

var validTopicRe = regexp.MustCompile(`^[a-z0-9-]+$`)

const (
	roleMaintainer = "maintainer"
	roleMember     = "member"
)

// Repository permission levels used across planning and apply.
const (
	permPull     = "pull"
	permTriage   = "triage"
	permPush     = "push"
	permMaintain = "maintain"
	permAdmin    = "admin"
)

const (
	precedenceCustomRoleCreate   = 5
	precedenceCustomRoleUpdate   = 5
	precedenceTeamCreate         = 10
	precedenceTeamUpdate         = 15
	precedenceRepoEnsure         = 10
	precedenceTeamRepoGrant      = 20
	precedenceTeamMemberEnsure   = 30
	precedenceRepoFileEnsure     = 40
	precedenceRepoTopicsEnsure   = 45
	precedenceRepoTemplateEnsure = 46
	precedenceRepoPinEnsure      = 47
	precedenceOrgMemberRemove    = 85
	precedenceTeamDelete         = 90
	precedenceRepoDelete         = 90
	precedenceCustomRoleDelete   = 95
)

const (
	errTermSHA            = "sha"
	errTermSHANotSupplied = "wasn't supplied"
	errTermRefExists      = "reference already exists"
)

type teamMemberChange struct {
	Org  string
	Slug string
	User string
	Role string // "member" or "maintainer"
}

type repoSettings struct {
	permission string
	topics     []string
	pinned     bool
	template   bool
	from       string
}

// validateTopic checks if a topic name meets GitHub requirements:
// - lowercase alphanumeric with hyphens
// - max 50 characters
// - cannot start with a hyphen
func validateTopic(topic string) error {
	if len(topic) == 0 {
		return fmt.Errorf("topic cannot be empty")
	}
	if len(topic) > 50 {
		return fmt.Errorf("topic exceeds 50 characters: %q", topic)
	}
	if topic[0] == '-' {
		return fmt.Errorf("topic cannot start with hyphen: %q", topic)
	}
	// Match lowercase alphanumeric and hyphens only
	if !validTopicRe.MatchString(topic) {
		return fmt.Errorf("topic contains invalid characters (must be lowercase alphanumeric with hyphens): %q", topic)
	}
	return nil
}

// normalizeYAMLMap converts both map[string]any and map[any]any (from YAML) to map[string]any.
func normalizeYAMLMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[fmt.Sprint(k)] = val
		}
		return result, true
	default:
		return nil, false
	}
}

// parseRepoConfig parses a repository value which can be either:
// - a simple string (permission only)
// - a map with permission, topics, pinned fields
func parseRepoConfig(val any) (repoSettings, error) {
	settings := repoSettings{}

	switch v := val.(type) {
	case string:
		// Simple case: just a permission string
		if v == "" {
			return settings, fmt.Errorf("permission cannot be empty string")
		}
		settings.permission = v
	default:
		m, ok := normalizeYAMLMap(val)
		if !ok {
			return settings, nil
		}
		if perm, ok := m["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := m["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", m["permission"])
		}
		// Permission is optional if using advanced config for topics/pinning only

		if topics, ok := m["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := m["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := m["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := m["from"].(string); ok {
			settings.from = from
		}
	}

	return settings, nil
}

// parseTemplateRef splits a template reference into org and repo parts.
// Supports "repo-name" (uses defaultOrg) or "org/repo-name".
func parseTemplateRef(ref, defaultOrg string) (org, repo string) {
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		return parts[0], parts[1]
	}
	return defaultOrg, ref
}

// resolveTemplate resolves template inheritance for a repository configuration.
// If the repo has a "from" field, it looks up the template repository and merges settings.
// Topics are combined (union), template flag is not inherited, and permission can be overridden.
func resolveTemplate(_ string, settings repoSettings, allRepos map[string]repoSettings, defaultOrg string) (repoSettings, error) {
	if settings.from == "" {
		return settings, nil
	}

	// Parse template reference (supports "repo-name" or "org/repo-name")
	templateOrg, templateRepo := parseTemplateRef(settings.from, defaultOrg)

	// Only support same-org templates for now
	if templateOrg != defaultOrg {
		return settings, fmt.Errorf("cross-organization template references not yet supported: %q", settings.from)
	}

	// Look up template repository in the current configuration
	templateKey := strings.ToLower(templateRepo)
	templateSettings, exists := allRepos[templateKey]
	if !exists {
		return settings, fmt.Errorf("template repository %q not found in configuration", templateRepo)
	}

	if !templateSettings.template {
		return settings, fmt.Errorf("repository %q is referenced as template but not marked with template: true", templateRepo)
	}

	// Merge settings: inherit from template, override with repo-specific
	result := settings

	// Inherit permission if not specified
	if result.permission == "" && templateSettings.permission != "" {
		result.permission = templateSettings.permission
	}

	// Merge topics (union): template topics + repo-specific topics
	// Clear existing topics first since we'll rebuild the list
	result.topics = nil
	topicSet := make(map[string]bool)

	// Add template topics first
	for _, topic := range templateSettings.topics {
		topicSet[topic] = true
		result.topics = append(result.topics, topic)
	}

	// Add repo-specific topics that aren't already in the set
	for _, topic := range settings.topics {
		if !topicSet[topic] {
			topicSet[topic] = true
			result.topics = append(result.topics, topic)
		}
	}

	// Don't inherit template or pinned flags
	// result.template is already false (or explicitly set)
	// result.pinned is already set from repo config

	return result, nil
}

// paginate calls fn repeatedly, advancing through pages until there are no more.
func paginate(fn func(opts *github.ListOptions) (*github.Response, error)) error {
	opts := &github.ListOptions{PerPage: defaultPerPage}
	for {
		resp, err := fn(opts)
		if err != nil {
			return err
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}

// ---- planning ----

func planTeams(_ context.Context, _ *gh.Client, cfg *config.Root, st *State) ([]util.Change, map[string]config.TeamConfig, error) {
	var out []util.Change
	desired := map[string]config.TeamConfig{}

	// build desired map
	for _, t := range cfg.Team {
		slug := t.ResolvedSlug()
		if slug == "" {
			continue
		}
		t.Slug = slug
		desired[slug] = t
	}

	// use prefetched teams
	actualBySlug := map[string]*github.Team{}
	for _, t := range st.ActualTeams {
		actualBySlug[t.GetSlug()] = t
	}

	// Track state
	st.CurrentTeams = len(st.ActualTeams)
	st.DesiredTeams = len(desired)

	for slug, want := range desired {
		if _, ok := actualBySlug[slug]; !ok {
			out = append(out, util.Change{
				Scope:  "team",
				Target: slug,
				Action: "create",
				Details: map[string]any{
					"org":         st.Org,
					"name":        want.Name,
					"privacy":     want.Privacy,
					"description": want.Description,
				},
			})
			continue
		}
		// Compare & update description/privacy
		existing := actualBySlug[slug]
		needsUpdate := false
		updateDetails := map[string]any{
			"org":  st.Org,
			"slug": slug,
			"name": want.Name,
		}
		if want.Description != existing.GetDescription() {
			needsUpdate = true
			updateDetails["description"] = want.Description
		}
		if want.Privacy != "" && want.Privacy != existing.GetPrivacy() {
			needsUpdate = true
			updateDetails["privacy"] = want.Privacy
		}
		if needsUpdate {
			out = append(out, util.Change{
				Scope:   "team",
				Target:  slug,
				Action:  "update",
				Details: updateDetails,
			})
		}
	}
	return out, desired, nil
}

func planTeamMembership(ctx context.Context, c *gh.Client, st *State, desiredBySlug map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	totalCurrentMembers := 0
	totalDesiredMembers := 0

	validatedUsers := map[string]bool{}

	for slug, want := range desiredBySlug {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// actual role map
		got := map[string]string{}
		// maintainers
		mopts := &github.TeamListTeamMembersOptions{Role: roleMaintainer, ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			mopts.ListOptions = *opts
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, mopts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			for _, u := range users {
				got[strings.ToLower(u.GetLogin())] = roleMaintainer
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}
		// members
		memOpts := &github.TeamListTeamMembersOptions{Role: roleMember, ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			memOpts.ListOptions = *opts
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, memOpts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			for _, u := range users {
				if _, ok := got[strings.ToLower(u.GetLogin())]; !ok {
					got[strings.ToLower(u.GetLogin())] = roleMember
				}
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}

		// desired role map
		wantRole := map[string]string{}
		for _, u := range want.Maintainers {
			wantRole[strings.ToLower(u)] = roleMaintainer
		}
		for _, u := range want.Members {
			if _, ok := wantRole[strings.ToLower(u)]; !ok {
				wantRole[strings.ToLower(u)] = roleMember
			}
		}

		// Validate that all desired users exist on GitHub
		for user := range wantRole {
			if validatedUsers[user] {
				continue
			}
			_, _, err := c.REST.Users.Get(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("user %q in team %q not found on GitHub: %w", user, slug, err)
			}
			validatedUsers[user] = true
		}

		// Track member counts
		totalCurrentMembers += len(got)
		totalDesiredMembers += len(wantRole)

		for user, role := range wantRole {
			if got[user] == role {
				continue
			}
			out = append(out, util.Change{
				Scope:   "team-member",
				Target:  slug,
				Action:  "ensure",
				Details: teamMemberChange{Org: org, Slug: slug, User: user, Role: role},
			})
		}
		// (optional) removals left for later
	}

	// Update state
	st.CurrentTeamMembers = totalCurrentMembers
	st.DesiredTeamMembers = totalDesiredMembers

	return out, nil
}

// collectRepoSettings gathers and validates all repository settings from config.
func collectRepoSettings(cfg *config.Root, _ string) (allSettings map[string]repoSettings, managedRepos map[string]bool, err error) {
	allSettings = map[string]repoSettings{}
	managedRepos = map[string]bool{}

	for _, t := range cfg.Team {
		slug := t.ResolvedSlug()
		for repo, val := range t.Repositories {
			r := strings.ToLower(repo)
			managedRepos[r] = true

			settings, err := parseRepoConfig(val)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid config for repo %s in team %s: %w", repo, slug, err)
			}
			allSettings[r] = settings
		}
	}
	return allSettings, managedRepos, nil
}

// resolveAllTemplates resolves template inheritance for all repository settings.
func resolveAllTemplates(allSettings map[string]repoSettings, org string) (map[string]repoSettings, error) {
	resolved := make(map[string]repoSettings, len(allSettings))
	for repo, settings := range allSettings {
		r, err := resolveTemplate(repo, settings, allSettings, org)
		if err != nil {
			return nil, fmt.Errorf("error resolving template for repo %s: %w", repo, err)
		}
		resolved[repo] = r
	}
	return resolved, nil
}

// teamRepoPermKey is "team-slug/repo-name" (lowercase).
type teamRepoPermKey = string

// fetchCurrentPermissions fetches the current team-repo permission grants from GitHub.
// Returns the total count and a map of "team/repo" -> permission string.
func fetchCurrentPermissions(ctx context.Context, c *gh.Client, cfg *config.Root, org string) (int, map[teamRepoPermKey]string, error) {
	count := 0
	permMap := map[teamRepoPermKey]string{}
	for _, t := range cfg.Team {
		if ctx.Err() != nil {
			return 0, nil, ctx.Err()
		}
		teamSlug := t.ResolvedSlug()
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			teamRepos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			count += len(teamRepos)
			for _, repo := range teamRepos {
				repoName := strings.ToLower(repo.GetName())
				perm := extractRepoPerm(repo)
				permMap[teamSlug+"/"+repoName] = perm
			}
			return resp, nil
		}); err != nil {
			return 0, nil, fmt.Errorf("fetch permissions for team %s: %w", teamSlug, err)
		}
	}
	return count, permMap, nil
}

// extractRepoPerm returns the highest permission level granted to a team for a repo.
func extractRepoPerm(repo *github.Repository) string {
	p := repo.Permissions
	if p == nil {
		return ""
	}
	switch {
	case p.Admin != nil && *p.Admin:
		return permAdmin
	case p.Maintain != nil && *p.Maintain:
		return permMaintain
	case p.Push != nil && *p.Push:
		return permPush
	case p.Triage != nil && *p.Triage:
		return permTriage
	case p.Pull != nil && *p.Pull:
		return permPull
	default:
		return ""
	}
}

func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	// use prefetched repos
	existing := map[string]bool{}
	existingRepos := map[string]*github.Repository{}
	for _, r := range st.ActualRepos {
		repoName := strings.ToLower(r.GetName())
		existing[repoName] = true
		existingRepos[repoName] = r
	}

	allRepoSettings, managedRepos, err := collectRepoSettings(cfg, org)
	if err != nil {
		return nil, err
	}

	resolvedSettings, err := resolveAllTemplates(allRepoSettings, org)
	if err != nil {
		return nil, err
	}

	desiredTopics := map[string][]string{}
	desiredPinned := map[string]bool{}
	desiredTemplates := map[string]bool{}
	emittedFiles := map[string]bool{} // tracks repo-level file changes to avoid duplicates

	for _, t := range cfg.Team {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		slug := t.ResolvedSlug()
		for repo := range t.Repositories {
			r := strings.ToLower(repo)
			settings := resolvedSettings[r]

			if !existing[r] && cfg.App.CreateRepo {
				details := map[string]any{
					"org":     org,
					"name":    repo,
					"private": true,
				}
				if settings.from != "" {
					details["from"] = settings.from
				}
				if settings.template {
					details["template"] = true
				}
				out = append(out, util.Change{
					Scope:   "repo",
					Target:  r,
					Action:  "ensure",
					Details: details,
				})
				existing[r] = true
			}

			if settings.template {
				desiredTemplates[r] = true
			}

			out = append(out, util.Change{
				Scope:  "team-repo",
				Target: slug + "/" + r,
				Action: "grant",
				Details: map[string]any{
					"org":        org,
					"slug":       slug,
					"repo":       repo,
					"permission": settings.permission,
				},
			})

			if len(settings.topics) > 0 {
				existingTopics := desiredTopics[r]
				topicSet := map[string]bool{}
				for _, topic := range existingTopics {
					topicSet[topic] = true
				}
				for _, topic := range settings.topics {
					if err := validateTopic(topic); err != nil {
						return nil, fmt.Errorf("invalid topic for repo %s: %w", repo, err)
					}
					if !topicSet[topic] {
						existingTopics = append(existingTopics, topic)
						topicSet[topic] = true
					}
				}
				desiredTopics[r] = existingTopics
			}

			if settings.pinned {
				desiredPinned[r] = true
			}

			// Emit file changes only once per repo (skip if already emitted from another team)
			if cfg.App.AddDefaultReadme && !emittedFiles[r+":README.md"] {
				readmeContent, err := templates.GenerateReadme(org, repo)
				if err != nil {
					return nil, fmt.Errorf("failed to generate README for %s: %w", repo, err)
				}
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":README.md",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    "README.md",
						"content": readmeContent,
						"message": "chore: add default README",
						"branch":  "main",
					},
				})
				emittedFiles[r+":README.md"] = true
			}
			if cfg.App.AddRenovateConfig && cfg.App.RenovateConfig != "" && !emittedFiles[r+":renovate"] {
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":.github/renovate.json",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    ".github/renovate.json",
						"content": cfg.App.RenovateConfig,
						"message": "chore: add Renovate config",
						"branch":  "main",
					},
				})
				emittedFiles[r+":renovate"] = true
			}
		}
	}

	// Plan topic updates
	for repo, topics := range desiredTopics {
		if len(topics) > 20 {
			return nil, fmt.Errorf("repo %s has %d topics (max 20 allowed)", repo, len(topics))
		}
		needsUpdate := false
		if existingRepo, ok := existingRepos[repo]; ok {
			currentTopics := existingRepo.Topics
			if len(currentTopics) != len(topics) {
				needsUpdate = true
			} else {
				currentSet := make(map[string]bool)
				for _, t := range currentTopics {
					currentSet[t] = true
				}
				for _, t := range topics {
					if !currentSet[t] {
						needsUpdate = true
						break
					}
				}
			}
		} else {
			needsUpdate = true
		}
		if needsUpdate {
			out = append(out, util.Change{
				Scope:  "repo-topics",
				Target: repo,
				Action: "ensure",
				Details: map[string]any{
					"org":    org,
					"repo":   repo,
					"topics": topics,
				},
			})
		}
	}

	// Plan pinning changes
	for repo, shouldPin := range desiredPinned {
		if shouldPin {
			out = append(out, util.Change{
				Scope:  "repo-pin",
				Target: repo,
				Action: "ensure",
				Details: map[string]any{
					"org":    org,
					"repo":   repo,
					"pinned": true,
				},
			})
		}
	}

	// Plan template marking changes
	for repo, shouldBeTemplate := range desiredTemplates {
		if shouldBeTemplate {
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				if !existingRepo.GetIsTemplate() {
					needsUpdate = true
				}
			} else {
				needsUpdate = true
			}
			if needsUpdate {
				out = append(out, util.Change{
					Scope:  "repo-template",
					Target: repo,
					Action: "ensure",
					Details: map[string]any{
						"org":      org,
						"repo":     repo,
						"template": true,
					},
				})
			}
		}
	}

	st.ManagedRepos = managedRepos
	st.CurrentRepos = len(existing)
	st.DesiredRepos = len(managedRepos)

	currentPerms, currentPermMap, err := fetchCurrentPermissions(ctx, c, cfg, org)
	if err != nil {
		return nil, fmt.Errorf("fetch current permissions: %w", err)
	}
	st.CurrentRepoPerms = currentPerms
	desiredPermsCount := 0
	for _, t := range cfg.Team {
		desiredPermsCount += len(t.Repositories)
	}
	st.DesiredRepoPerms = desiredPermsCount

	// Filter out no-op grants where the permission already matches
	filtered := out[:0]
	for _, ch := range out {
		if ch.Scope == "team-repo" && ch.Action == "grant" {
			d := ch.Details.(map[string]any)
			slug := d["slug"].(string)
			repo := strings.ToLower(d["repo"].(string))
			desired := normalizePermission(d["permission"].(string))
			current := currentPermMap[slug+"/"+repo]
			if current == desired {
				continue // skip, already has correct permission
			}
		}
		filtered = append(filtered, ch)
	}

	return filtered, nil
}

// planTeamCleanups generates delete changes for teams not in the desired set.
func planTeamCleanups(st *State, org string, desired map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	for _, at := range st.ActualTeams {
		if _, ok := desired[at.GetSlug()]; !ok {
			out = append(out, util.Change{Scope: "team", Target: at.GetSlug(), Action: "delete", Details: map[string]any{"org": org, "slug": at.GetSlug()}})
		}
	}
	return out, nil
}

// planMemberCleanups generates remove changes for org members not in any team.
func planMemberCleanups(ctx context.Context, c *gh.Client, org string) ([]util.Change, error) {
	var out []util.Change
	memOpt := &github.ListMembersOptions{
		Role:        roleMember,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	var members []*github.User
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		memOpt.ListOptions = *opts
		us, resp, err := c.REST.Organizations.ListMembers(ctx, org, memOpt)
		if err != nil {
			return nil, err
		}
		members = append(members, us...)
		return resp, nil
	}); err != nil {
		return nil, err
	}
	inAnyTeam := map[string]bool{}
	var allTeams []*github.Team
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, ts...)
		return resp, nil
	}); err != nil {
		return nil, err
	}
	for _, t := range allTeams {
		tmOpt := &github.TeamListTeamMembersOptions{Role: "all", ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			tmOpt.ListOptions = *opts
			us, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, t.GetSlug(), tmOpt)
			if err != nil {
				return nil, err
			}
			for _, u := range us {
				inAnyTeam[strings.ToLower(u.GetLogin())] = true
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}
	}
	for _, u := range members {
		login := strings.ToLower(u.GetLogin())
		if !inAnyTeam[login] {
			out = append(out, util.Change{Scope: "org-member", Target: login, Action: "remove", Details: map[string]any{"org": org, "user": login}})
		}
	}
	return out, nil
}

// planRepoCleanups generates delete/warning changes for unmanaged repositories.
func planRepoCleanups(cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org
	var unmanagedRepos []string
	for _, repo := range st.ActualRepos {
		repoName := strings.ToLower(repo.GetName())
		if !st.ManagedRepos[repoName] {
			unmanagedRepos = append(unmanagedRepos, repo.GetName())
			if cfg.App.DeleteUnmanagedRepos {
				out = append(out, util.Change{
					Scope:  "repo",
					Target: repoName,
					Action: "delete",
					Details: map[string]any{
						"org":  org,
						"repo": repo.GetName(),
					},
				})
			}
		}
	}
	if cfg.App.DryWarnings.WarnUnmanagedRepos && len(unmanagedRepos) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d unmanaged repositories: %v", len(unmanagedRepos), unmanagedRepos))
	}
	return out, warnings, nil
}

func planCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	if cfg.App.DeleteUnconfiguredTeams {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, err := planTeamCleanups(st, org, desired)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
	}

	if cfg.App.RemoveMembersWithoutTeam {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, err := planMemberCleanups(ctx, c, org)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
	}

	if cfg.App.DeleteUnmanagedRepos || cfg.App.DryWarnings.WarnUnmanagedRepos {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, w, err := planRepoCleanups(cfg, st)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
		warnings = append(warnings, w...)
	}

	return out, warnings, nil
}

// containsErrorMessage checks if a GitHub ErrorResponse contains a specific error message
// in either the main Message field or in any of the individual Error messages in the Errors array.
func containsErrorMessage(ghErr *github.ErrorResponse, searchTerms ...string) bool {
	// Check main message (only if not empty)
	if ghErr.Message != "" {
		allFound := true
		for _, term := range searchTerms {
			if !strings.Contains(ghErr.Message, term) {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}

	// Check individual errors in the Errors array
	for _, e := range ghErr.Errors {
		allFound := true
		for _, term := range searchTerms {
			if !strings.Contains(e.Message, term) {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}

	return false
}

// ---- apply ----

func applyChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	return applyChangesWith(ctx, c, changes, defaultRegistry)
}

func applyChangesWith(ctx context.Context, c *gh.Client, changes []util.Change, reg *HandlerRegistry) error {
	sort.SliceStable(changes, func(i, j int) bool {
		return reg.Precedence(changes[i].Scope, changes[i].Action) <
			reg.Precedence(changes[j].Scope, changes[j].Action)
	})

	// Apply custom role changes first — they have their own dispatcher.
	if err := applyCustomRoleChanges(ctx, c, changes); err != nil {
		return err
	}

	// Count non-custom-role changes for progress display.
	total := 0
	for _, ch := range changes {
		if !strings.HasPrefix(ch.Scope, "custom-role") {
			total++
		}
	}

	applied := 0
	for _, ch := range changes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		applied++
		util.Infof("[%d/%d] %s:%s %s", applied, total, ch.Scope, ch.Action, ch.Target)

		if err := gh.RespectRate(ctx, c.REST); err != nil {
			util.Warnf("rate limit check failed: %v", err)
		}

		handler, ok := reg.Lookup(ch.Scope, ch.Action)
		if !ok {
			util.Warnf("no handler for change %s:%s on %s", ch.Scope, ch.Action, ch.Target)
			continue
		}
		if err := handler.Apply(ctx, c, ch); err != nil {
			util.Audit(ch.Scope, ch.Target, ch.Action, "error")
			return err
		}
		util.Audit(ch.Scope, ch.Target, ch.Action, "ok")
	}
	return nil
}
````

## File: .github/workflows/release.yaml
````yaml
name: Release gomgr

on:
  push:
    tags:
      - "v*.*.*"
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    name: Build binaries
    runs-on: ubuntu-24.04
    strategy:
      fail-fast: false
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]

    steps:
      - name: Checkout
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6
        with:
          go-version: "1.26.2"

      - name: Set VERSION
        run: echo "VERSION=${GITHUB_REF_NAME}" >> $GITHUB_ENV

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          BIN=gomgr
          if [ "${{ matrix.goos }}" = "windows" ]; then BIN=gomgr.exe; fi
          mkdir -p build
          go build -trimpath -ldflags "-s" -o build/$BIN ./main.go

      - name: Package
        run: |
          PKG="gomgr_${VERSION}_${{ matrix.goos }}_${{ matrix.goarch }}"
          mkdir -p "dist/${PKG}"
          cp build/* "dist/${PKG}/"
          if [ -f README.md ]; then cp README.md "dist/${PKG}/"; fi
          if [ -f LICENSE ]; then cp LICENSE "dist/${PKG}/"; fi
          cd dist
          if [ "${{ matrix.goos }}" = "windows" ]; then
            zip -r "${PKG}.zip" "${PKG}"
          else
            tar -czf "${PKG}.tar.gz" "${PKG}"
          fi

      - name: Upload artifact
        uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7
        with:
          name: "artifacts_${{ matrix.goos }}_${{ matrix.goarch }}"
          path: |
            dist/*.tar.gz
            dist/*.zip
          if-no-files-found: error

  release:
    name: Create GitHub Release
    runs-on: ubuntu-24.04
    needs: build
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c # v8
        with:
          path: dist

      - name: Generate checksums
        run: |
          cd dist
          find . -type f \( -name "*.tar.gz" -o -name "*.zip" \) -print0 | xargs -0 -I {} sh -c 'sha256sum "{}" >> checksums.txt'
          ls -la

      - name: Create Release
        uses: softprops/action-gh-release@153bb8e04406b158c6c84fc1615b65b24149a1fe # v2
        with:
          draft: false
          prerelease: false
          files: |
            dist/**/*.tar.gz
            dist/**/*.zip
            dist/checksums.txt
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
````

## File: go.mod
````
module github.com/DragonSecurity/gomgr

go 1.26.2

require (
	github.com/bradleyfalzon/ghinstallation/v2 v2.18.0
	github.com/google/go-github/v84 v84.0.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	golang.org/x/oauth2 v0.36.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
)
````
