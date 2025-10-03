# github-org-manager-go (gomgr)

A fast, idempotent **GitHub Organization Manager** written in Go. Define your org as YAML and apply it with a single command. Ships with a release workflow and a CI workflow to run sync against one or many org-config folders.

## Highlights

- ✅ YAML-driven org config (`app.yaml`, `org.yaml`, `teams/*.yaml`)
- ✅ Teams, maintainers, members (idempotent add/update)
- ✅ Repo permission grants (pull/triage/push/maintain/admin)
- ✅ **Optional**: create repos that don’t exist (`create_repo: true`)
- ✅ **Optional**: inject `.github/renovate.json` into repos
- ✅ Warnings & cleanups: unmanaged teams, members without team
- ✅ **Optional** hard cleanups: delete unmanaged teams, remove members without team
- ✅ Auth: GitHub App (recommended) or PAT
- ✅ `--dry` plan with stable output; safe apply
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
gomgr sync -c <config> --dry
gomgr sync -c <config>
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

# Optional enforcement / extras:
remove_members_without_team: true   # remove org members not in any team
delete_unconfigured_teams: true     # delete teams not defined in YAML
create_repo: true                   # create repos if missing when referenced by teams
add_renovate_config: true           # create .github/renovate.json in repos
renovate_config: |
  {
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["github>DragonSecurity/renovate-presets"]
  }
```

### `org.yaml`
Currently used for owners (extension point):
```yaml
owners:
  - alice
  - bob
```

### `teams/*.yaml`
```yaml
name: Platform Team
slug: platform-team            # optional; default = kebab(name)
description: Core platform engineers
privacy: closed                # closed | secret
parents: []                    # (future enhancement)
maintainers:
  - alice
members:
  - bob
repositories:
  infra: maintain              # pull|triage|push|maintain|admin
  api: push
```

> Loader ignores non‑YAML files in `teams/` and skips empty/invalid entries.

---

## Auth & Permissions

### GitHub App (recommended)
Set `GITHUB_APP_ID` and `GITHUB_APP_PRIVATE_KEY` (or `app_id`/`private_key` in `app.yaml`). The app must be installed on the org.

**Typical minimum permissions** (tighten to least privilege for your use-case):

- **Organization**: Members (Read/Write), Teams (Read/Write)
- **Repository**: Administration (Read/Write) to grant team access & create repos; Contents (Read/Write) to create files like Renovate config; Metadata (Read)

> Exact permissions depend on which features you enable (e.g., if you don’t create repos or write files, you can drop those).

### Personal Access Token (PAT)
Use a classic PAT with scopes:
- `admin:org` (manage teams/members)
- `repo` (set team repo access and create repos)
- `read:org` (read org metadata)

---

## CLI

- `gomgr sync -c <config> [--dry] [--debug]`  
  Plans and applies org state.

- `gomgr setup-team -n "Team Name" -c <config> [-f out/path.yaml]`  
  Bootstraps a team YAML.

- `gomgr version`  
  Prints version (stamped at build). If built with VCS info, also prints revision/dirty/commit time.

**Order of operations** (apply):  
create teams → set memberships → ensure repos → grant permissions → write renovate (optional) → cleanups (optional)

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
          gh release download "$VERSION" --repo DragonSecurity/github-org-manager-go --pattern "$ASSET" --dir .gomgr
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

## Troubleshooting

- **404 on `/teams//members`**: empty/invalid team YAML or calling membership on a team that doesn’t exist yet. Loader ignores non‑YAML files and planner guards empty slugs; team creation happens before membership.
- **`gomgr version` shows `dev`**: build without `-ldflags -X` or not from a tag. Use the release workflow or pass a version when building.
- **Renovate config not created**: ensure `add_renovate_config: true` and `renovate_config` is non‑empty; repo must exist or `create_repo: true`.

---

## Roadmap / TODO

- Compare & update team fields (description/privacy/parents)
- Optionally remove extra team members / revoke extra repo perms
- Custom default branch for file writes
- Parallel apply with rate‑limit aware workers
- More comprehensive plan diff output

---

## Contributing

PRs welcome! Please:
- open an issue first for larger changes,
- keep commits small & focused,
- add tests where practical,
- run `golangci-lint` (if configured).

---

## Security

This tool modifies org membership and repository access. Use **dry‑run** in CI and restrict credentials using least privilege. Prefer GitHub Apps over PATs.

---

## License

See **[LICENSE](./LICENSE.md)**.
