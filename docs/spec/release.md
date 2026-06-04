# traCtl Release Operations

Status: Active  
Audience: Maintainers with org-level GitHub and Docker Hub access  

This document describes how releases are triggered, which credentials are required, and how draft vs public releases are decided. Implementation lives in `.github/workflows/release.yml`, `.goreleaser.yml`, and `.github/actions/build-web-assets/`.

---

## CI vs release workflows

| Event | Workflow | Purpose |
|-------|----------|---------|
| Pull request to `main` / `master` | `.github/workflows/pr.yml` | Format, vet, lint, tests, `build-cli` + `build-wasm` |
| Tag push `v*.*.*` or `v*.*.*-*` | `.github/workflows/release.yml` | Binaries, Docker images, Homebrew formula, macOS DMGs, cask update |
| Manual | **Actions → Release → Run workflow** | Same pipeline as a tag; `draft` input controls GitHub Release state |

There is no separate `ci.yml`. PR checks are the only pre-merge gate; merging does not re-run the full matrix.

---

## Org secrets (GitHub)

Configure these at the **organization** (or repository) level under **Settings → Secrets and variables → Actions**.

| Secret | Purpose | How to set up |
|--------|---------|----------------|
| `HOMEBREW_TAP_TOKEN` | Push formula/cask updates to [`tractl/homebrew-tap`](https://github.com/tractl/homebrew-tap) | Create a **fine-grained** GitHub PAT. Repository access: `tractl/homebrew-tap` only. Permission: **Contents → Read and write**. Store the token as this secret. |
| `DOCKER_USERNAME` | Log in to Docker Hub for image push | Your Docker Hub username (plain text secret). |
| `DOCKER_PASSWORD` | Docker Hub login | Docker Hub → Account settings → **Security** → **New access token** (read/write for repos you publish). Store the token as this secret—not your account password. |
| `GITHUB_TOKEN` | Create/upload GitHub Release assets | Provided automatically by Actions; `release.yml` sets `contents: write`. No manual setup. |

Without `HOMEBREW_TAP_TOKEN`, GoReleaser still builds CLI artifacts; Homebrew tap commits and the post-release cask job are skipped when the release policy disables Homebrew publish (see below).

---

## Trigger a release (tags)

1. Ensure `main` is green on the PR workflow.
2. Choose a [SemVer](https://semver.org/) tag on `main`:

   ```bash
   git checkout main
   git pull
   git tag v0.1.0-alpha.1
   git push origin v0.1.0-alpha.1
   ```

   Or push all local tags: `git push --tags` (only if you intend every local tag to release).

3. **Actions → Release** runs automatically for tags matching `v*.*.*` and `v*.*.*-*`.

### Tag → draft / Homebrew policy

| Tag pattern | GitHub Release | Homebrew formula (`goreleaser`) | Cask update job |
|-------------|----------------|----------------------------------|-----------------|
| Contains `-alpha`, `-beta`, or `-rc` (e.g. `v0.1.0-alpha.1`, `v0.1.0-rc1`) | **Draft** | Skipped (`--skip=homebrew`) | Skipped |
| Stable (e.g. `v0.1.0`) | **Published** (not draft) | Published to tap | Runs after macOS DMGs upload |

Policy is implemented in the **Resolve release policy** step in `release.yml` (not by editing `.goreleaser.yml` for each release).

### Manual release (no new tag)

**Actions → Release → Run workflow**

- **draft** (boolean, default `true`): when checked, creates a draft release and skips Homebrew publish—useful for dry runs or recovering a failed pipeline.

---

## What the release workflow does

```text
tag push / workflow_dispatch
        │
        ▼
  goreleaser job (ubuntu)
        ├── checkout (full history)
        ├── ./.github/actions/build-web-assets  (make build-web)
        ├── resolve draft / Homebrew policy
        ├── Docker buildx + push (web + ci images)
        └── goreleaser release (CLI + archives + formula when enabled)
        │
        ▼
  desktop-macos (matrix: arm64, amd64)
        └── Wails build → DMG → upload to GitHub Release
        │
        ▼
  update-cask (ubuntu, stable releases only)
        └── SHA256 DMGs → commit tractl-desktop cask to homebrew-tap
```

**GoReleaser `before` hooks:** `go mod tidy`, then `make build-web` only if `cmd/server/dist/web/index.html` is missing (CI already built web assets).

**Desktop** is not built by GoReleaser; macOS runners build DMGs in `release.yml`.

---

## Promote a draft release to public

1. Open the release on GitHub → **Edit** → uncheck **Set as a draft** → **Update release**.
2. For pre-releases that should ship Homebrew: run a **stable** tag when ready, or manually update `tractl/homebrew-tap` (not recommended—prefer a stable tag).

CLI: `gh release edit v0.1.0-alpha.1 --draft=false`

---

## Local maintainer commands

```bash
# Build embed tree + WASM + web bundle (required before server release artifacts)
make build-web

# Local snapshot (no publish, no tag required)
make release-snapshot

# Full publish (needs git tag on HEAD, GITHUB_TOKEN, HOMEBREW_TAP_TOKEN)
make release
```

Install GoReleaser: `brew install goreleaser` (see `Makefile` `release` target).

---

## Related files

- `.github/workflows/pr.yml` — PR CI
- `.github/workflows/release.yml` — tag / manual release
- `.github/actions/build-web-assets/action.yml` — shared Go/Node setup and `make` target
- `.goreleaser.yml` — cross-platform CLI, server binary, Docker, Homebrew formula
- `scripts/install.sh` — curl installer (referenced in release notes template)
