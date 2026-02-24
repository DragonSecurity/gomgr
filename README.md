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
- ✅ **Just-In-Time (JIT) access**: temporary team/repo access with automatic cleanup
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

## Just-In-Time (JIT) Access

gomgr supports Just-In-Time (JIT) access for temporary team and repository access. This is useful for granting short-term access without modifying your YAML configuration.

### Commands

**`gomgr access`** - Grant temporary access (default: 1 hour)
```bash
# Grant team access for 1 hour
gomgr access developer allanice001 --org myorg
gomgr access developer allanice001 maintainer --org myorg

# Grant repository access for 1 hour
gomgr access myrepo allanice001 push --org myorg

# Grant access for custom duration
gomgr access myrepo allanice001 push --org myorg --duration 2h
gomgr access developer allanice001 --org myorg --duration 30m
```

**`gomgr grant`** - Grant permanent access (without YAML config)
```bash
# Grant permanent team membership
gomgr grant developer allanice001 --org myorg
gomgr grant developer allanice001 maintainer --org myorg

# Grant permanent repository access
gomgr grant myrepo allanice001 push --org myorg
gomgr grant myrepo allanice001 admin --org myorg
```

**`gomgr list-jit`** - List active JIT grants
```bash
gomgr list-jit
```

**`gomgr cleanup-jit`** - Manually revoke expired access
```bash
gomgr cleanup-jit --org myorg
```

### How JIT Access Works

1. **Grant Access**: Use `gomgr access` to grant temporary access to a team or repository
2. **Track State**: JIT grants are tracked in `~/.gomgr/jit-state.json`
3. **Auto Cleanup**: Run `gomgr cleanup-jit` periodically to revoke expired access
4. **Manual Cleanup**: Use `gomgr list-jit` to view active grants and `gomgr cleanup-jit` to clean up

### Automation

For automatic cleanup, run `gomgr cleanup-jit` in a cron job or scheduled workflow:

```yaml
# .github/workflows/cleanup-jit.yml
name: Cleanup JIT Access
on:
  schedule:
    - cron: '*/15 * * * *'  # Every 15 minutes
  workflow_dispatch:

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Cleanup expired JIT grants
        run: |
          gomgr cleanup-jit --org myorg
        env:
          GITHUB_TOKEN: ${{ secrets.ORG_ADMIN_TOKEN }}
```

### Slack Bot Integration

Example Slack bot handler for JIT access commands:

```go
// Handle: /gomgr access developer allanice001
func handleSlackCommand(command string) {
    parts := strings.Fields(command)
    if len(parts) < 4 || parts[1] != "access" {
        return
    }
    
    team := parts[2]
    user := parts[3]
    
    exec.Command("gomgr", "access", team, user, "--org", "myorg").Run()
}
```

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
