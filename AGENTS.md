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
