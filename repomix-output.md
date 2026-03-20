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
.github/
  ISSUE_TEMPLATE/
    bug_report.md
    feature_request.md
  workflows/
    ci.yaml
    release.yaml
  renovate.json
cmd/
  root.go
  setup_team.go
  sync.go
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
    loader.go
    types_test.go
    types.go
  gh/
    client.go
    rate.go
  sync/
    custom_roles_test.go
    custom_roles.go
    orchestrator.go
    teams_test.go
    teams.go
  templates/
    readme_test.go
    readme.go
  util/
    diff_test.go
    diff.go
    log.go
  version/
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

## File: internal/sync/custom_roles.go
````go
package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
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
				return fmt.Errorf("update custom role %q (ID %d): %w", d.Name, d.ID, err)
			}

		case "custom-role:delete":
			_, err := c.REST.Organizations.DeleteCustomRepoRole(ctx, d.Org, d.ID)
			if err != nil {
				return fmt.Errorf("delete custom role %q (ID %d): %w", d.Name, d.ID, err)
			}
		}
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

## File: internal/util/diff_test.go
````go
package util

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

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

## File: internal/util/log.go
````go
package util

import (
	"fmt"
	"log"
	"os"
)

func EnableDebug() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Warnf prints a warning message to stderr
func Warnf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", v...)
}
````

## File: .golangci.yml
````yaml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly
  go: '1.26'

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gocyclo
    - misspell
    - unconvert
    - goconst
    - goimports
    - revive
    - gosec

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
  misspell:
    locale: US
  revive:
    confidence: 0.8
  gosec:
    excludes:
      - G204 # Audit use of command execution - we need this for bash commands
      - G304 # Audit use of file path - we need this for config files
  goimports:
    local-prefixes: github.com/DragonSecurity/gomgr

issues:
  exclude-rules:
    # Exclude some linters from running on tests files.
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - gosec
  max-issues-per-linter: 0
  max-same-issues: 0

output:
  formats:
    - format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
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
        go-version: ['1.26.0']

    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6

      - name: Setup Go
        uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@cdf6c1fa76f9f475f3d7449005a359c84ca0f306 # v5
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
        uses: actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6
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
        uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6
        with:
          go-version: '1.26.0'

      - name: Cache Go modules
        uses: actions/cache@cdf6c1fa76f9f475f3d7449005a359c84ca0f306 # v5
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
        uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6
        with:
          go-version: '1.26.0'

      - name: Cache Go modules
        uses: actions/cache@cdf6c1fa76f9f475f3d7449005a359c84ca0f306 # v5
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
        uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6
        with:
          go-version: '1.26.0'

      - name: Cache Go modules
        uses: actions/cache@cdf6c1fa76f9f475f3d7449005a359c84ca0f306 # v5
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
        uses: actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6
        with:
          name: gomgr-binary
          path: build/gomgr
          if-no-files-found: error
````

## File: cmd/root.go
````go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgDir string
	debug  bool
	dryRun bool
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
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose debug logs")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry", false, "Show a plan without applying changes")
}
````

## File: cmd/setup_team.go
````go
package cmd

import (
	"path/filepath"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/spf13/cobra"
)

var teamName string
var outFile string

var setupTeamCmd = &cobra.Command{
	Use:   "setup-team",
	Short: "Bootstrap a team YAML file for a given team name",
	RunE: func(cmd *cobra.Command, args []string) error {
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

## File: internal/config/types_test.go
````go
package config

import (
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

func PrintPlan(p Plan) {
	b, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println(string(b))
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

## File: main.go
````go
package main

import "github.com/DragonSecurity/gomgr/cmd"

func main() {
	cmd.Execute()
}
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
	$(GOTEST) -v -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./internal/config ./internal/sync
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

## File: internal/config/loader.go
````go
package config

import (
	"errors"
	"os"
	"path/filepath"
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
	entries, _ := os.ReadDir(teamDir)
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
	return r, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
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

## File: internal/gh/rate.go
````go
package gh

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v83/github"
)

func RespectRate(ctx context.Context, c *github.Client) error {
	r, _, err := c.RateLimits(ctx)
	if err != nil {
		return nil
	}
	if core := r.GetCore(); core.Remaining < 50 {
		sleep := time.Until(core.Reset.Time)
		log.Printf("rate-limit: sleeping until %s", core.Reset.Time)
		time.Sleep(sleep + time.Second)
	}
	return nil
}
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

## File: cmd/version.go
````go
package cmd

import (
	"fmt"

	"github.com/DragonSecurity/gomgr/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
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

## File: internal/config/types.go
````go
package config

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
````

## File: internal/sync/teams_test.go
````go
package sync

import (
	"strings"
	"testing"

	"github.com/google/go-github/v83/github"
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
````

## File: cmd/sync.go
````go
package cmd

import (
	"context"
	"log"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if debug {
			util.EnableDebug()
		}

		cfg, err := config.Load(cfgDir)

		if err != nil {
			return err
		}

		client, appInfo, err := gh.NewClientFromEnv(ctx, cfg.App)
		if err != nil {
			return err
		}
		if appInfo != "" {
			log.Printf("auth: %s", appInfo)
		}

		plan, err := insync.BuildPlan(ctx, client, cfg)
		if err != nil {
			return err
		}

		util.PrintPlan(plan)

		if dryRun {
			util.PrintSummary(plan)
			log.Println("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, plan)
	},
}

func init() {
	syncCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "", "Path to config directory (required)")
	_ = syncCmd.MarkPersistentFlagRequired("config")
	rootCmd.AddCommand(syncCmd)
}
````

## File: internal/gh/client.go
````go
package gh

import (
	"bytes"
	"context"
	"crypto/x509"
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

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"
)

type Client struct {
	REST       *github.Client
	httpClient *http.Client
}

func NewClientFromEnv(ctx context.Context, app config.AppConfig) (*Client, string, error) {
	// PAT
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
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
	httpClient := &http.Client{Transport: itr, Timeout: 30 * time.Second}
	return &Client{REST: github.NewClient(httpClient), httpClient: httpClient}, "Github App", nil
}

func maybeReadPEM(s string) ([]byte, error) {
	if strings.Contains(s, "BEGIN") {
		return []byte(s), nil
	}
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM at %s", s)
	}
	if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		_, _ = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	return b, nil
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute graphql request: %w", err)
	}
	defer resp.Body.Close()

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
		errMsg := gqlResp.Errors[0].Message
		return fmt.Errorf("graphql error: %s", errMsg)
	}

	if result != nil && len(gqlResp.Data) > 0 {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("decode graphql data: %w", err)
		}
	}

	return nil
}
````

## File: internal/sync/orchestrator.go
````go
package sync

import (
	"context"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

type State struct {
	Org          string
	ManagedRepos map[string]bool

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

	// Owners (stub - optional)
	// ownerChanges, err := planOwners(ctx, c, cfg, st)
	// if err != nil { return plan, err }

	// Custom roles must be created before teams/repos use them
	customRoleChanges, err := planCustomRoles(ctx, c, cfg, st)
	if err != nil {
		return plan, err
	}

	teamChanges, desiredBySlug, err := planTeams(ctx, c, cfg, st)
	if err != nil {
		return plan, err
	}

	memChanges, err := planTeamMembership(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, err
	}

	repoChanges, err := planRepoPerms(ctx, c, cfg, st)
	if err != nil {
		return plan, err
	}

	cleanupChanges, warnings, err := planCleanups(ctx, c, cfg, st, desiredBySlug)
	if err != nil {
		return plan, err
	}

	customRoleCleanups, roleWarnings, err := planCustomRoleCleanups(ctx, c, cfg, st)
	if err != nil {
		return plan, err
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

func Apply(ctx context.Context, c *gh.Client, plan util.Plan) error {
	return applyChanges(ctx, c, plan.Changes)
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

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
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
	validTopic := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validTopic.MatchString(topic) {
		return fmt.Errorf("topic contains invalid characters (must be lowercase alphanumeric with hyphens): %q", topic)
	}
	return nil
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
	case map[string]any:
		// Advanced case: RepoConfig structure
		if perm, ok := v["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := v["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", v["permission"])
		}
		// Permission is optional if using advanced config for topics/pinning only

		if topics, ok := v["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := v["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := v["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := v["from"].(string); ok {
			settings.from = from
		}
	case map[any]any:
		// YAML might unmarshal as map[any]any
		if perm, ok := v["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := v["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", v["permission"])
		}

		if topics, ok := v["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := v["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := v["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := v["from"].(string); ok {
			settings.from = from
		}
	}

	return settings, nil
}

// resolveTemplate resolves template inheritance for a repository configuration.
// If the repo has a "from" field, it looks up the template repository and merges settings.
// Topics are combined (union), template flag is not inherited, and permission can be overridden.
func resolveTemplate(repoName string, settings repoSettings, allRepos map[string]repoSettings, defaultOrg string) (repoSettings, error) {
	if settings.from == "" {
		return settings, nil
	}

	// Parse template reference (supports "repo-name" or "org/repo-name")
	templateOrg := defaultOrg
	templateRepo := settings.from
	if strings.Contains(settings.from, "/") {
		parts := strings.SplitN(settings.from, "/", 2)
		if len(parts) != 2 {
			return settings, fmt.Errorf("invalid template reference format: %q (expected 'repo' or 'org/repo')", settings.from)
		}
		templateOrg = parts[0]
		templateRepo = parts[1]
	}

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

// ---- planning ----

func planTeams(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, map[string]config.TeamConfig, error) {
	var out []util.Change
	desired := map[string]config.TeamConfig{}

	// build desired map
	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" && t.Name != "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		if slug == "" {
			continue
		}
		t.Slug = slug
		desired[slug] = t
	}

	// list actual teams
	var actual []*github.Team
	opt := &github.ListOptions{PerPage: 100}
	for {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, st.Org, opt)
		if err != nil {
			return nil, nil, err
		}
		actual = append(actual, ts...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	actualBySlug := map[string]*github.Team{}
	for _, t := range actual {
		actualBySlug[t.GetSlug()] = t
	}

	// Track state
	st.CurrentTeams = len(actual)
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
		// TODO: compare & update description/privacy/parents as needed
	}
	return out, desired, nil
}

func planTeamMembership(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desiredBySlug map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	totalCurrentMembers := 0
	totalDesiredMembers := 0

	for slug, want := range desiredBySlug {
		// actual role map
		got := map[string]string{}
		// maintainers
		mopts := &github.TeamListTeamMembersOptions{Role: "maintainer", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, mopts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				got[strings.ToLower(u.GetLogin())] = "maintainer"
			}
			if resp.NextPage == 0 {
				break
			}
			mopts.Page = resp.NextPage
		}
		// members
		opts := &github.TeamListTeamMembersOptions{Role: "member", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				if _, ok := got[strings.ToLower(u.GetLogin())]; !ok {
					got[strings.ToLower(u.GetLogin())] = "member"
				}
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		// desired role map
		wantRole := map[string]string{}
		for _, u := range want.Maintainers {
			wantRole[strings.ToLower(u)] = "maintainer"
		}
		for _, u := range want.Members {
			if _, ok := wantRole[strings.ToLower(u)]; !ok {
				wantRole[strings.ToLower(u)] = "member"
			}
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

func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	existing := map[string]bool{}
	existingRepos := map[string]*github.Repository{}
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}, Type: "all"}
	for {
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			repoName := strings.ToLower(r.GetName())
			existing[repoName] = true
			existingRepos[repoName] = r
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Track which repos are managed
	managedRepos := map[string]bool{}

	// Map to track desired topics and pinned state per repo
	desiredTopics := map[string][]string{}
	desiredPinned := map[string]bool{}
	desiredTemplates := map[string]bool{}

	// First pass: collect all repository settings
	allRepoSettings := map[string]repoSettings{}
	repoToTeams := map[string][]string{} // track which teams reference each repo

	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		for repo, val := range t.Repositories {
			r := strings.ToLower(repo)
			managedRepos[r] = true

			settings, err := parseRepoConfig(val)
			if err != nil {
				return nil, fmt.Errorf("invalid config for repo %s in team %s: %w", repo, slug, err)
			}

			// Store settings for later template resolution
			allRepoSettings[r] = settings
			repoToTeams[r] = append(repoToTeams[r], slug)
		}
	}

	// Second pass: resolve templates
	resolvedSettings := make(map[string]repoSettings)
	for repo, settings := range allRepoSettings {
		resolved, err := resolveTemplate(repo, settings, allRepoSettings, org)
		if err != nil {
			return nil, fmt.Errorf("error resolving template for repo %s: %w", repo, err)
		}
		resolvedSettings[repo] = resolved
	}

	// Third pass: process repositories with resolved settings
	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		for repo := range t.Repositories {
			r := strings.ToLower(repo)
			settings := resolvedSettings[r]

			if !existing[r] && cfg.App.CreateRepo {
				details := map[string]any{
					"org":     org,
					"name":    repo,
					"private": true,
				}
				// Include template information if present
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

			// Mark repository as template if configured
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

			// Aggregate topics from all teams (union)
			if len(settings.topics) > 0 {
				existingTopics := desiredTopics[r]
				topicSet := map[string]bool{}
				for _, topic := range existingTopics {
					topicSet[topic] = true
				}
				for _, topic := range settings.topics {
					// Validate topic before adding
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

			// Pinned state: if any team wants it pinned, pin it
			if settings.pinned {
				desiredPinned[r] = true
			}

			if cfg.App.AddDefaultReadme {
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
			}
			if cfg.App.AddRenovateConfig && cfg.App.RenovateConfig != "" {
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
			}
		}
	}

	// Plan topic updates for managed repos - only if different from current state
	for repo, topics := range desiredTopics {
		if len(topics) > 0 {
			// GitHub allows max 20 topics per repo
			if len(topics) > 20 {
				return nil, fmt.Errorf("repo %s has %d topics (max 20 allowed)", repo, len(topics))
			}

			// Check if topics differ from current state
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				currentTopics := existingRepo.Topics
				if len(currentTopics) != len(topics) {
					needsUpdate = true
				} else {
					// Compare topics (order-independent)
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
				// Repo doesn't exist yet, will be created
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
	}

	// Plan pinning changes - check current pinning state
	// Note: GitHub REST API doesn't provide pinning status directly
	// We'll generate changes for all repos marked as pinned
	// A future enhancement could use GraphQL to query current pinning state
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
			// Check if repo needs to be marked as template
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				if !existingRepo.GetIsTemplate() {
					needsUpdate = true
				}
			} else {
				// Repo will be created, needs template marking
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

	// Store managed repos in state for cleanup phase
	st.ManagedRepos = managedRepos

	// Track repository counts
	st.CurrentRepos = len(existing)
	st.DesiredRepos = len(managedRepos)

	// Count permissions (team-repo grants)
	// Note: This requires additional API calls to get accurate current state.
	// These calls are intentional for precise state tracking and run only during
	// dry-run planning. The overhead is acceptable for the visibility benefit.
	currentPermsCount := 0
	desiredPermsCount := 0

	// Count current permissions from GitHub
	for _, t := range cfg.Team {
		teamSlug := t.Slug
		if teamSlug == "" {
			teamSlug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		// List team repos to count current permissions
		repoOpts := &github.ListOptions{PerPage: 100}
		for {
			teamRepos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, org, teamSlug, repoOpts)
			if err != nil {
				// If team doesn't exist yet, skip counting
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				// Ignore other errors for counting purposes
				break
			}
			currentPermsCount += len(teamRepos)
			if resp.NextPage == 0 {
				break
			}
			repoOpts.Page = resp.NextPage
		}
	}

	// Count desired permissions
	for _, t := range cfg.Team {
		desiredPermsCount += len(t.Repositories)
	}

	st.CurrentRepoPerms = currentPermsCount
	st.DesiredRepoPerms = desiredPermsCount

	return out, nil
}

func planCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org
	if cfg.App.DeleteUnconfiguredTeams {
		var actual []*github.Team
		opt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, opt)
			if err != nil {
				return nil, nil, err
			}
			actual = append(actual, ts...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		for _, at := range actual {
			if _, ok := desired[at.GetSlug()]; !ok {
				out = append(out, util.Change{Scope: "team", Target: at.GetSlug(), Action: "delete", Details: map[string]any{"org": org, "slug": at.GetSlug()}})
			}
		}
	}

	if cfg.App.RemoveMembersWithoutTeam {
		// list all org members
		memOpt := &github.ListMembersOptions{
			Role: "member",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		var members []*github.User
		for {
			us, resp, err := c.REST.Organizations.ListMembers(ctx, org, memOpt)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, us...)
			if resp.NextPage == 0 {
				break
			}
			memOpt.Page = resp.NextPage
		}
		// compute members who are in any team
		inAnyTeam := map[string]bool{}
		teamOpt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, teamOpt)
			if err != nil {
				return nil, nil, err
			}
			for _, t := range ts {
				page := &github.TeamListTeamMembersOptions{Role: "all", ListOptions: github.ListOptions{PerPage: 100}}
				for {
					us, r2, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, t.GetSlug(), page)
					if err != nil {
						return nil, nil, err
					}
					for _, u := range us {
						inAnyTeam[strings.ToLower(u.GetLogin())] = true
					}
					if r2.NextPage == 0 {
						break
					}
					page.Page = r2.NextPage
				}
			}
			if resp.NextPage == 0 {
				break
			}
			teamOpt.Page = resp.NextPage
		}
		for _, u := range members {
			login := strings.ToLower(u.GetLogin())
			if !inAnyTeam[login] {
				out = append(out, util.Change{Scope: "org-member", Target: login, Action: "remove", Details: map[string]any{"org": org, "user": login}})
			}
		}
	}

	// Warn about or delete unmanaged repositories
	if cfg.App.DeleteUnmanagedRepos || cfg.App.DryWarnings.WarnUnmanagedRepos {
		var actualRepos []*github.Repository
		repoOpt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}, Type: "all"}
		for {
			repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, repoOpt)
			if err != nil {
				return nil, nil, err
			}
			actualRepos = append(actualRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			repoOpt.Page = resp.NextPage
		}
		var unmanagedRepos []string
		for _, repo := range actualRepos {
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
		// Add warning if configured and there are unmanaged repos
		if cfg.App.DryWarnings.WarnUnmanagedRepos && len(unmanagedRepos) > 0 {
			warnings = append(warnings, fmt.Sprintf("Found %d unmanaged repositories: %v", len(unmanagedRepos), unmanagedRepos))
		}
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
	precedence := map[string]int{
		"custom-role:create":   5, // Create custom roles first, before teams/repos
		"custom-role:update":   5,
		"team:create":          10,
		"repo:ensure":          10,
		"team-repo:grant":      20,
		"team-member:ensure":   30,
		"repo-file:ensure":     40,
		"repo-topics:ensure":   45,
		"repo-template:ensure": 46,
		"repo-pin:ensure":      47,
		"team:delete":          90,
		"repo:delete":          90,
		"custom-role:delete":   95, // Delete custom roles last
	}

	sort.Slice(changes, func(i, j int) bool {
		ai := changes[i].Scope + ":" + changes[i].Action
		aj := changes[j].Scope + ":" + changes[j].Action
		return precedence[ai] < precedence[aj]
	})

	// Apply custom role changes first
	if err := applyCustomRoleChanges(ctx, c, changes); err != nil {
		return err
	}

	for _, ch := range changes {
		// Skip custom role changes - already handled above
		if strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		switch ch.Scope + ":" + ch.Action {
		case "team:create":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			var privacyPtr, descPtr *string
			if v, ok := d["privacy"]; ok && fmt.Sprint(v) != "" {
				pv := fmt.Sprint(v)
				privacyPtr = github.Ptr(pv)
			}
			if v, ok := d["description"]; ok && fmt.Sprint(v) != "" {
				dv := fmt.Sprint(v)
				descPtr = github.Ptr(dv)
			}
			newTeam := github.NewTeam{Name: name, Privacy: privacyPtr, Description: descPtr}
			_, _, err := c.REST.Teams.CreateTeam(ctx, org, newTeam)
			if err != nil {
				return fmt.Errorf("create team %q: %w", name, err)
			}

		case "team:delete":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			_, err := c.REST.Teams.DeleteTeamBySlug(ctx, org, slug)
			if err != nil {
				return fmt.Errorf("delete team %s: %w", slug, err)
			}

		case "team-member:ensure":
			d, ok := ch.Details.(teamMemberChange)
			if !ok {
				return fmt.Errorf("invalid details for team-member:ensure")
			}
			_, _, err := c.REST.Teams.AddTeamMembershipBySlug(ctx, d.Org, d.Slug, d.User, &github.TeamAddTeamMembershipOptions{Role: d.Role})
			if err != nil {
				return fmt.Errorf("add %s as %s to %s: %w", d.User, d.Role, d.Slug, err)
			}

		case "repo:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			private := true
			if v, ok := d["private"]; ok {
				private = fmt.Sprint(v) != "false"
			}
			isTemplate := false
			if v, ok := d["template"]; ok {
				isTemplate = fmt.Sprint(v) == "true"
			}

			// Check if this repo should be created from a template
			if fromTemplate, ok := d["from"]; ok && fromTemplate != "" {
				templateRef := fmt.Sprint(fromTemplate)
				// Parse template reference (supports "repo-name" or "org/repo-name")
				templateOrg := org
				templateRepo := templateRef
				if strings.Contains(templateRef, "/") {
					parts := strings.SplitN(templateRef, "/", 2)
					if len(parts) == 2 {
						templateOrg = parts[0]
						templateRepo = parts[1]
					}
				}

				// Create repository from template
				_, _, err := c.REST.Repositories.CreateFromTemplate(ctx, templateOrg, templateRepo, &github.TemplateRepoRequest{
					Name:    github.Ptr(name),
					Owner:   github.Ptr(org),
					Private: github.Ptr(private),
				})
				if err != nil {
					var ghErr *github.ErrorResponse
					if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
						// already exists race
					} else {
						return fmt.Errorf("create repo %s/%s from template %s/%s: %w", org, name, templateOrg, templateRepo, err)
					}
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
					if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
						// already exists race
					} else {
						return fmt.Errorf("create repo %s/%s: %w", org, name, err)
					}
				}
			}

		case "team-repo:grant":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			repo := fmt.Sprint(d["repo"])
			perm := normalizePermission(fmt.Sprint(d["permission"]))
			_, err := c.REST.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repo, &github.TeamAddTeamRepoOptions{Permission: perm})
			if err != nil {
				return fmt.Errorf("grant %s on %s/%s to %s: %w", perm, org, repo, slug, err)
			}

		case "repo-file:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])
			path := fmt.Sprint(d["path"])
			content := []byte(fmt.Sprint(d["content"]))
			message := fmt.Sprint(d["message"])
			branch := fmt.Sprint(d["branch"])
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
					// This can happen due to timing - template files are copied asynchronously.
					// GitHub returns 422 with "sha wasn't supplied" or 409 with "reference already exists"
					// when trying to create a file that already exists.
					var ghErr *github.ErrorResponse
					if errors.As(err, &ghErr) && ghErr.Response != nil {
						// Check if this is a race condition error
						isRaceCondition := (ghErr.Response.StatusCode == 422 && containsErrorMessage(ghErr, "sha", "wasn't supplied")) ||
							(ghErr.Response.StatusCode == 409 && containsErrorMessage(ghErr, "reference already exists"))

						if !isRaceCondition {
							return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
						}
						// File already exists (likely from template), which is what we want - skip error
					} else {
						return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
					}
				}
			} else {
				// optional: update if differs (skipped for now)
			}

		case "repo-topics:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

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

			_, _, err := c.REST.Repositories.ReplaceAllTopics(ctx, org, repo, topicsRaw)
			if err != nil {
				return fmt.Errorf("set topics on %s/%s: %w", org, repo, err)
			}

		case "repo-template:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

			// Mark repository as a template
			_, _, err := c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
				IsTemplate: github.Ptr(true),
			})
			if err != nil {
				return fmt.Errorf("mark repo %s/%s as template: %w", org, repo, err)
			}

		case "repo-pin:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

			// Note: GitHub's GraphQL API does not support pinning repositories to organization profiles.
			// The pinRepository mutation only works for user profiles, not organizations.
			// This is a known limitation of the GitHub API.
			// See: https://github.com/orgs/community/discussions/184845
			util.Warnf("Skipping pin for %s/%s: GitHub API does not support pinning to organization profiles", org, repo)

		case "repo:delete":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])
			_, err := c.REST.Repositories.Delete(ctx, org, repo)
			if err != nil {
				return fmt.Errorf("delete repo %s/%s: %w", org, repo, err)
			}

		default:
			// no-op for unhandled changes
		}
	}
	return nil
}

func normalizePermission(p string) string {
	// Use lowercase comparison to match built-in roles case-insensitively
	switch strings.ToLower(p) {
	case "read", "pull":
		return "pull"
	case "triage":
		return "triage"
	case "write", "push":
		return "push"
	case "maintain":
		return "maintain"
	case "admin":
		return "admin"
	default:
		// For custom repository roles (GitHub Enterprise Cloud), pass through the role name as-is
		// preserving the original case since custom role names may be case-sensitive
		// Custom roles must be created in the GitHub organization before use
		// Examples: "actions-manager", "release-manager", "runner-admin"
		return p
	}
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
        uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6
        with:
          go-version: "1.26.0"

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
        uses: actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6
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
        uses: actions/download-artifact@37930b1c2abaa49bbe596cd826c3c89aef350131 # v7
        with:
          path: dist

      - name: Generate checksums
        run: |
          cd dist
          find . -type f \( -name "*.tar.gz" -o -name "*.zip" \) -print0 | xargs -0 -I {} sh -c 'sha256sum "{}" >> checksums.txt'
          ls -la

      - name: Create Release
        uses: softprops/action-gh-release@a06a81a03ee405af7f2048a818ed3f03bbf83c7b # v2
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

go 1.26.0

require (
	github.com/bradleyfalzon/ghinstallation/v2 v2.17.0
	github.com/google/go-github/v83 v83.0.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/oauth2 v0.35.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/go-github/v75 v75.0.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
)
````
