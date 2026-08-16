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
- ✅ **Rulesets**: org-wide and repo-specific guard rails (branch protection, tag protection, push rules) with built-in presets — see [Rulesets & guard rails](#rulesets--guard-rails)
- ✅ **Adopt what's already there**: `gomgr import teams` / `gomgr import rulesets` bootstrap an unmanaged org into YAML, comments in your config left intact — see [Adopting an existing organization](#adopting-an-existing-organization)
- ✅ **Custom repository roles**: fully managed - define in YAML, gomgr creates/updates them (GitHub Enterprise Cloud)
- ✅ **Repository topics**: add topics/labels to repositories for organization
- ✅ **Repository pinning**: pin important repositories to organization profile (⚠️ *GitHub API limitation: not currently supported for organizations - configuration accepted but manual pinning required via web UI*)
- ✅ **Optional**: create repos that don’t exist (`create_repo: true`)
- ✅ **Optional**: inject `.github/renovate.json` into repos
- ✅ Warnings & cleanups: unmanaged teams, members without team, unmanaged repos, unmanaged custom roles, unmanaged rulesets
- ✅ **Optional** hard cleanups: delete unmanaged teams, remove members without team, delete unmanaged repos, delete unmanaged custom roles, delete unmanaged rulesets
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

- **App credentials on the command line** — for a config directory shared
  through a repository, which has nowhere safe to keep a private key:
  ```bash
  gomgr sync -c <config> --app-id 1719369 --private-key ~/keys/gomgr.pem
  ```
  Credentials resolve in the order **flags → `app.yaml` → environment**, so the
  committed `app.yaml` can name the org and the behavior flags while CI
  supplies the secrets and you supply them locally.

  There is deliberately no `--token` flag: an App ID is not a secret and a
  private key is passed by *path*, but a PAT on `argv` is visible in `ps` and
  lands in your shell history, so `GITHUB_TOKEN` stays the only way in.

  A relative `private_key:` in `app.yaml` is resolved next to that `app.yaml`
  if it does not resolve from your working directory — so `-c ../../org-config`
  works from anywhere.

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
  warn_unmanaged_rulesets: true      # warn about rulesets not defined in YAML

# Optional enforcement / extras:
remove_members_without_team: true   # remove org members not in any team
delete_unconfigured_teams: true     # delete teams not defined in YAML
delete_unmanaged_repos: false       # delete repos not defined in any team (DESTRUCTIVE!)
delete_unmanaged_custom_roles: false # delete custom roles not in org.yaml (DESTRUCTIVE!)
delete_unmanaged_rulesets: false    # delete rulesets not in YAML (removes guard rails!)
create_repo: true                   # create repos if missing when referenced by teams

# Sign every commit gomgr writes with a Signed-off-by trailer. Required when
# the org enforces DCO via a ruleset `commit_message_pattern` rule — gomgr
# commits straight to the default branch, so no pull-request DCO check ever
# runs on them and an unsigned message is rejected at push time.
# Empty (the default) appends nothing.
sign_off: "my-app[bot] <12345+my-app[bot]@users.noreply.github.com>"

# Legacy convenience flags — still honoured, but `files:` is the preferred
# way to declare per-repo content. Legacy flags are materialised into
# FileSpec entries at load time.
add_renovate_config: true
add_default_readme: false
renovate_config: |
  {
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }

# Templated files that gomgr ensures exist in every managed repository.
# Each `content` is rendered through Go's text/template package with
# {Org, Repo} as context. `only` restricts an entry to specific repos via
# path.Match-style globs; omitting `only` matches every managed repo.
files:
  - path: README.md
    message: "chore: add README"
    branch: main
    content: |
      # {{.Repo}}

      Part of the [{{.Org}}](https://github.com/{{.Org}}) organization.

  - path: LICENSE
    only:
      - "public-*"
      - "oss-*"
    content: |
      MIT License
      Copyright (c) {{.Org}} ...

  - path: .github/CODEOWNERS
    content: |
      * @{{.Org}}/platform-team
```

### Files & templating (`app.files`)

The `files:` list on `app.yaml` replaces the older `add_default_readme` /
`renovate_config` special cases with a single mechanism for declaring any
file that should exist in every managed repo.

- **`path`** (required): repo-relative path, e.g. `README.md`,
  `.github/workflows/ci.yml`.
- **`content`** (required): Go text/template source. `{{.Org}}` and
  `{{.Repo}}` are available; referencing an unknown field is a hard error at
  plan time (via `missingkey=error`).
- **`message`** (optional): commit message; defaults to `chore: add <path>`.
- **`branch`** (optional): target branch; defaults to `main`.
- **`only`** (optional): list of `path.Match` globs against the repo name.
  Empty/omitted matches every repo.

Legacy `add_default_readme` and `add_renovate_config` still work — at load
time they are converted into FileSpec entries and prepended to `files:`. If
you list the same `path:` yourself, your entry overrides the legacy one, so
you can keep the flags on and still supply a custom README.

### Repository visibility

Advanced repo configs accept a `visibility:` field: `public`, `private`
(the default), or `internal` (GHEC only). It is respected when gomgr creates
the repository:

```yaml
repositories:
  public-docs:
    permission: push
    visibility: public      # created public; pairs well with FileSpec `only: [public-*]`
  internal-playbook:
    permission: push
    visibility: internal    # GHEC-only; visible to the whole enterprise
  private-notes:
    permission: admin       # visibility omitted → private (backwards-compatible)
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

`org.yaml` is also where organization-wide **rulesets** live — see
[Rulesets & guard rails](#rulesets--guard-rails) below.

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

## Rulesets & guard rails

Rulesets are GitHub's successor to branch protection, and gomgr manages them the
same way it manages everything else: declaratively, idempotently, and with a
`--dry` plan you can read before anything changes.

They come in two scopes, and **both apply at once** — GitHub evaluates every
ruleset that matches a push and enforces the strictest outcome, so a repository
ruleset tightens the org baseline rather than replacing it:

| Scope | Declared in | Covers |
| --- | --- | --- |
| Organization | `org.yaml` → `rulesets:` | Every repository the `repository_name` condition matches |
| Repository | `teams/*.yaml` → a repo's `rulesets:` | That repository alone |

### Presets

A preset is a named guard rail. Reference one and you get its target,
conditions and rules without restating them:

| Preset | What it enforces |
| --- | --- |
| `branch-protection` | Default branch: no deletion, no force-push, changes via a pull request with 1 approval, stale reviews dismissed, review threads resolved |
| `strict-branch-protection` | The same, tightened: 2 approvals, code-owner review, last-push approval, linear history |
| `tag-protection` | Tags cannot be moved or deleted |
| `no-force-push` | Every branch is append-only. The minimum guard rail if you are not ready to require reviews |
| `require-signed-commits` | Every commit on the default branch carries a verified signature |
| `require-dco` | Commit messages must contain `Signed-off-by:`, enforced at the ref rather than by a pull-request check |
| `no-committed-keys` | Push rule rejecting `.pem`, `.key`, `.p12`, `.pfx`, `.jks`, `.keystore`, `.ppk` |

Anything you set alongside a preset wins, **at whole-rule granularity**: naming
a rule key replaces the preset's version of that rule outright rather than
merging into it field by field. Setting a boolean rule to `false` switches the
preset's version off.

```yaml
# org.yaml
rulesets:
  - name: default-branch-protection
    preset: branch-protection
    conditions:
      repository_name:
        include: ["~ALL"]
        exclude: ["sandbox-*"]
    bypass_actors:
      - type: Integration
        app: self          # gomgr's own GitHub App — see the warning below
        mode: always

  # Same preset, two approvals instead of one, only on production services.
  - name: production-branch-protection
    preset: branch-protection
    conditions:
      repository_name:
        include: ["svc-*"]
    rules:
      pull_request:
        required_approving_review_count: 2
        require_code_owner_review: true
        required_review_thread_resolution: true
```

```yaml
# teams/security-team.yaml
repositories:
  vulnerability-reports:
    permission: admin
    rulesets:
      - name: locked-down-main
        preset: strict-branch-protection
      - name: no-committed-keys
        preset: no-committed-keys
```

### ⚠️ Do not lock gomgr out of its own repositories

gomgr commits templated files (`app.files`, CODEOWNERS, `renovate.json`)
straight to the default branch. A ruleset that requires a pull request or
signed commits on that branch **rejects those pushes**. Either exempt the app:

```yaml
bypass_actors:
  - type: Integration
    app: self       # resolves to app.app_id; requires GitHub App auth, not a PAT
    mode: always
```

…or scope the ruleset away from the repositories gomgr writes to. gomgr raises
this as a plan warning when it spots the combination, so `--dry` tells you
before the sync does.

### Rolling a guard rail out safely

Set `enforcement: evaluate` to have GitHub record what *would* have been blocked
without blocking it. Read the ruleset's insights page, then flip to `active`.

```yaml
  - name: require-ci-green
    target: branch
    enforcement: evaluate     # active | evaluate | disabled
    conditions:
      ref_name:
        include: ["~DEFAULT_BRANCH"]
      repository_name:
        include: ["~ALL"]
    rules:
      required_status_checks:
        strict: true
        checks:
          - context: build
          - context: test
```

### Reference

**Ruleset fields**

| Field | Notes |
| --- | --- |
| `name` | Required. Also the identity gomgr matches on, case-insensitively |
| `preset` | Optional built-in guard rail (table above) |
| `target` | `branch` (default), `tag`, or `push` |
| `enforcement` | `active` (default), `evaluate`, or `disabled` |
| `conditions.ref_name` | `include` / `exclude` fnmatch patterns, plus `~ALL` and `~DEFAULT_BRANCH`. Defaults to `~ALL`; not used by push rulesets |
| `conditions.repository_name` | `include` / `exclude` / `protected`. Organization rulesets only; defaults to `~ALL` |
| `bypass_actors` | Who may bypass. Compared exactly — an actor nobody configured is removed |
| `rules` | The rules themselves |

**Bypass actors**

```yaml
bypass_actors:
  - type: Team
    team: platform-team       # slug, resolved to a team ID
    mode: pull_request        # always | pull_request (default always)
  - type: Integration
    app: self                 # "self" or a numeric GitHub App ID
  - type: RepositoryRole
    actor_id: 5               # role ID from the repository-roles API
  # Identified by type alone — these take no ID:
  - type: OrganizationAdmin
  - type: EnterpriseOwner
  - type: DeployKey
```

`OrganizationAdmin`, `EnterpriseOwner` and `DeployKey` are recognized by their
type; GitHub reports them with no `actor_id` and gomgr sends none. `Team`,
`Integration` and `RepositoryRole` each need an identity — a slug, an app, or a
role ID respectively.

Role IDs are not guessable and gomgr does not map role names to them; read the
IDs your organization actually uses from the
[repository roles API](https://docs.github.com/en/rest/orgs/custom-roles).

**Rules**

Branch and tag targets: `creation`, `update`, `deletion`,
`required_linear_history`, `required_signatures`, `non_fast_forward`,
`pull_request`, `required_status_checks`, `required_deployments`, `merge_queue`,
`commit_message_pattern`, `commit_author_email_pattern`,
`committer_email_pattern`, `branch_name_pattern`, `tag_name_pattern`,
`workflows`, `code_scanning`.

Push target: `file_extension_restriction`, `file_path_restriction`,
`max_file_path_length`, `max_file_size`. Push rulesets need GitHub Enterprise
Cloud on private repositories.

`gomgr validate -c <config>` checks all of this offline — unknown presets,
invalid enumerations, duplicate names, and rules used on the wrong target — so
you find mistakes before GitHub answers with an opaque 422.

### Adopting rulesets that already exist

Most orgs do not start with gomgr. Somebody protected `main` in the web UI two
years ago, somebody else added a tag rule, and none of it is in your YAML.

```bash
gomgr import rulesets -c <config>            # show what could be adopted
gomgr import rulesets -c <config> --write    # splice it into the config files
```

The scan reads the organization **and every repository** — wider than `sync`,
which only looks at repositories your teams declare, precisely because a
ruleset on a repository nobody has adopted yet is the one you most want to know
about.

What it does:

- **Skips what you already declare.** A ruleset your YAML names is gomgr's to
  define; re-importing it would overwrite your config with the live state,
  which is backwards.
- **Collapses to a preset** when one describes the ruleset exactly, so you get
  `preset: branch-protection` rather than forty lines of expanded rules.
- **Restores names, not IDs.** Team IDs become slugs, and gomgr's own app ID
  becomes `app: self`, so the adopted config survives a team being recreated.
- **Drops GitHub's defaults.** `negate: false`, a `~ALL`/`~ALL` selector, and
  the full merge-method list are what GitHub reports when nothing was chosen —
  writing them down would state a default as a decision.
- **Leaves your files alone otherwise.** Entries are spliced into `org.yaml`
  and the `teams/*.yaml` file that already declares each repository. Comments,
  blank lines and quoting elsewhere in those files are untouched, so what you
  review is a small diff. A repository written as `infra: push` grows into a
  settings map with its permission intact.
- **Validates before and after.** The adopted rulesets go through the same
  checks a hand-written config does, and the whole directory is reloaded after
  writing — if either fails, you hear about it there and then.
- **Reports what it cannot express, and carries on.** GitHub's ruleset schema
  is wider than gomgr's and keeps growing. A ruleset using something this build
  does not know is named, with the reason, and left exactly as it is on GitHub —
  one such ruleset does not abort a scan of fifty.

```console
$ gomgr import rulesets -c . --write
adopted 3 rulesets -> org.yaml
adopted 1 ruleset  -> teams/platform-team.yaml

4 rulesets adopted across 2 files.
2 rulesets already declared in your configuration were left alone.

1 repository holds rulesets but appears in no team file, so there is
nowhere to write them. Add the repository to a team first:
  - legacy-scratch

Review with `git diff`, then commit and open a pull request.
```

gomgr stops at the working tree; committing and opening the pull request stays
yours, because a guard rail landing on `main` is a decision a person should
make. Once adopted, the rulesets are ordinary config — the next `sync` will
keep them in step, and `warn_unmanaged_rulesets` goes quiet.

### Cleanup

`warn_unmanaged_rulesets: true` reports rulesets that exist on GitHub but are
not in your YAML. `delete_unmanaged_rulesets: true` removes them. Rulesets
inherited from the organization or an enterprise are never touched at the
repository scope — they are not that scope's to delete.

⚠️ **Before turning `delete_unmanaged_rulesets` on, run `gomgr import
rulesets`.** Otherwise the flag deletes exactly the hand-made guard rails the
import would have adopted.

Two more things worth knowing about how `sync` treats rulesets it did not
create:

- **A matching name is a takeover.** Rulesets are matched by name,
  case-insensitively. If your YAML declares `protect main` and someone hand-made
  a "Protect Main", gomgr treats them as the same ruleset and replaces the live
  definition wholesale with yours. Different names never collide; near-identical
  ones do.
- **`sync` only inspects managed repositories.** Rulesets on repositories that
  appear in no team file are invisible to `sync`, warnings included. `import` is
  what sees those.

---

## Adopting an existing organization

gomgr is declarative, which is only useful for things that are actually
declared. An organization that predates it has teams, permissions and guard
rails that gomgr can see but has no way to adopt — and turning on a cleanup
flag against that state deletes precisely the things you never wrote down.

`gomgr import` closes that gap. Every subcommand is read-only against GitHub,
prints what it found by default, and only writes to your config files with
`--write`.

```bash
gomgr import teams    -c <config>            # bootstrap teams, members, permissions
gomgr import rulesets -c <config>            # adopt guard rails
```

### `gomgr import teams`

Renders every team the config does not declare as a `teams/<slug>.yaml`: name,
description, privacy, maintainers, members, and the repositories that team
reaches with the permission it holds.

```console
$ gomgr import teams -c . --write
adopted team platform                       -> teams/platform.yaml
adopted team security                       -> teams/security.yaml

2 teams adopted.
Skipped 1 team already declared in your configuration.

Review with `git status` and `git diff`, then commit and open a pull request.
```

- **Files are never overwritten.** A `teams/<slug>.yaml` that already exists
  stops the import for that team rather than replacing something you edited.
- **Slugs are only recorded when they matter.** A team whose name derives its
  slug gets no `slug:` line, so an imported file reads like a hand-written one.
- **Members and maintainers are sorted**, so re-importing produces the same
  bytes.
- **The result is validated and reloaded** before the command reports success.

Two things it deliberately does not do:

- **Team hierarchy is reported, not written.** gomgr does not manage nested
  teams, so a team with a parent is flagged as a warning rather than given a
  `parents:` line that nothing would act on.
- **Repositories reached by no team are reported, not adopted.** A repository
  only enters gomgr's config by being granted to a team, so there is no
  meaningful place to put one that no team reaches. **If
  `delete_unmanaged_repos` is set, those are exactly the repositories your next
  sync deletes** — the import shouts about that case specifically:

```console
⚠  No team reaches 2 repositories in this organization, and
delete_unmanaged_repos is set. THE NEXT SYNC WOULD DELETE THEM:
  - forgotten-service
  - old-prototype
Grant them to a team before your next sync.
```

That warning is the reason to run `import teams` *before* your first sync
against an organization gomgr has not managed before.

### Suggested order

```bash
gomgr import teams    -c <config> --write   # 1. teams, members, permissions
gomgr import rulesets -c <config> --write   # 2. guard rails (needs the teams)
gomgr validate        -c <config>           # 3. offline check
gomgr sync            -c <config> --dry     # 4. should report no changes
```

Step 4 is the one that tells you the adoption was faithful: if the import
captured everything, a dry run against the same organization has nothing to do.
Anything it *does* report is a genuine gap worth reading before you commit.

Rulesets come after teams because a bypass actor can name a team, and the
importer resolves team IDs back to slugs using the teams it can see.

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

- `gomgr import teams -c <config> [--write]`  
  Adopts teams that exist on GitHub but are not in your YAML, writing one
  `teams/<slug>.yaml` each. See
  [Adopting an existing organization](#adopting-an-existing-organization).

- `gomgr import rulesets -c <config> [--write] [--only <glob>]`  
  Adopts rulesets that exist on GitHub but are not in your YAML. Prints them by
  default; `--write` splices them into your config files. See
  [Adopting rulesets that already exist](#adopting-rulesets-that-already-exist).

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
