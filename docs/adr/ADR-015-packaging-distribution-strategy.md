# ADR-015 — Packaging & Distribution Strategy

Status: Accepted
Date: 2026-05-24

---

## Context

ADR-009 defines the deployment topology model and the invariant that deployment topology must not change canonical workflow semantics. ADR-012 defines five first-class delivery surfaces (Desktop, Browser, CLI, CI, Container).

Neither ADR defines how traCtl reaches users — what distribution channels exist, what installation expectations each channel carries, and how packaging decisions relate to execution parity guarantees.

Without a formal distribution strategy, packaging decisions accumulate ad-hoc. Platform-specific packaging affects installation trust, update lifecycle, and in some cases runtime assumptions. A browser extension distributed through an enterprise-managed channel has different trust and lifecycle properties than one installed directly. A CLI distributed via Homebrew carries different update expectations than one installed via curl.

This ADR defines canonical distribution channels per surface, packaging format requirements, update lifecycle expectations, and the relationship between distribution decisions and execution parity.

---

## Decision

### 1 — Desktop Distribution

Desktop distribution targets macOS, Windows, and Linux as first-class platforms.

**macOS**

Primary channel: Homebrew Cask

```
brew install --cask tractl
```

Secondary channel: direct DMG download (signed and notarized)

Requirements:

- application MUST be code-signed with a valid Apple Developer certificate
- application MUST be notarized via Apple's notarization service
- DMG MUST be signed
- auto-update mechanism MUST be included (Sparkle or equivalent)

**Windows**

Primary channel: MSIX installer (Microsoft Store compatible)

Secondary channel: direct installer download (EXE, signed)

Requirements:

- installer MUST be code-signed (Authenticode)
- installer MUST support silent installation for enterprise deployment (`/S` flag or equivalent)
- auto-update mechanism MUST be included

**Linux**

Primary channel: Homebrew (Linux) / Linuxbrew

Secondary channel: distribution-native packages

Supported package formats:

- `.deb` (Debian, Ubuntu)
- `.rpm` (Fedora, RHEL, openSUSE)
- `.AppImage` (distribution-agnostic)

Additional channels:

- Flatpak (deferred, post-v1)
- Snap (deferred, post-v1)

Requirements:

- AppImage MUST be self-contained with no external runtime dependencies
- `.deb` and `.rpm` packages MUST declare explicit runtime dependencies

---

### 2 — CLI Distribution

CLI distribution targets developer machines and CI environments.

**Primary channels:**

Homebrew (macOS and Linux):

```
brew install tractl
```

curl installer (all platforms):

```
curl -fsSL https://install.tractl.softwits.com | sh
```

The curl installer MUST:

- verify binary integrity via checksum before installation
- support explicit version pinning
- support offline installation via pre-downloaded binary
- install to a user-owned directory without requiring elevated privileges by default

**Secondary channels:**

- npm global install (deferred)
- pip install (deferred — not idiomatic for a Go binary; only if ecosystem demand warrants)
- Scoop (Windows package manager)
- WinGet (Windows Package Manager)

**Version pinning:**

CLI distribution MUST support explicit version pinning across all channels.

CI environments MUST be able to pin to an exact version without floating channel resolution.

Example:

```
brew install tractl@1.2.3
```

Rationale: CI workflows that float to latest CLI versions silently break when the CLI introduces behavior changes. Explicit pinning is required for CI surface execution parity.

---

### 3 — Container Distribution

Container distribution targets CI runner environments, sidecar execution, and reproducible isolated execution.

**Primary channel:** Docker Hub

Registry: `tractl/tractl` (Docker Hub)

**Image variants:**

| Tag | Contents | Use Case |
|---|---|---|
| `tractl/tractl:web` | tractl-server binary (embeds React + WASM) on distroless | Web app — `docker run -p 7428:7428` |
| `tractl/tractl:web-X.Y.Z` | same, pinned version | Reproducible deployments |
| `tractl/tractl:ci` | tractl CLI binary on debian-slim | CI runner integration |
| `tractl/tractl:ci-X.Y.Z` | same, pinned version | Reproducible CI |

**Image requirements:**

- all images MUST be multi-arch (linux/amd64, linux/arm64)
- images MUST be signed (cosign or equivalent)
- images MUST carry SBOM attestation
- `latest` tag MUST track the latest stable release only (never pre-release)
- pre-release images MUST use explicit pre-release tags

**Immutable execution requirement:**

Container images are immutable execution environments.

Providers and extensions MUST NOT be installed at container runtime.

Provider and extension availability MUST be declared at image build time.

Rationale: Runtime provider installation inside a container creates non-reproducible execution environments. The execution parity guarantee (ADR-009 §1) requires that container execution be deterministic across runs without network dependency.

**CI runner image:**

A dedicated CI runner image MUST be maintained.

CI runner image contents:

- tractl binary (pinned version)
- OpenAPI provider (bundled)
- common CI-compatible output configuration
- no interactive UX dependencies

---

### 4 — Browser Extension Distribution

Browser extension distribution targets the three major extension marketplaces.

**Supported browsers:**

- Google Chrome (Chrome Web Store)
- Mozilla Firefox (Firefox Add-ons / AMO)
- Microsoft Edge (Edge Add-ons)

**Distribution requirements:**

- extensions MUST pass each marketplace's review process before distribution
- extensions MUST be signed by each respective marketplace
- extension permissions MUST be minimal and explicitly declared
- permissions MUST follow the principle of least privilege

**Declared permissions scope:**

The browser extension MUST NOT request:

- access to browser history
- access to stored passwords or autofill data
- access to cross-origin pages not explicitly interacted with by the user
- persistent background execution without user consent

The browser extension MAY request:

- `webRequest` or `declarativeNetRequest` (for capture, with user consent)
- access to pages the user is actively browsing
- communication with the traCtl Desktop application (native messaging)
- local storage for session capture state

**Enterprise deployment:**

Browser extensions MUST support enterprise-managed deployment via:

- Chrome Group Policy (Google Admin)
- Firefox Enterprise Policy
- Edge Group Policy

Enterprise-managed deployments MUST NOT require manual user installation.

---

### 5 — Update Lifecycle Model

Update lifecycle expectations differ by surface.

**Desktop:**

- semantic versioning (MAJOR.MINOR.PATCH)
- in-app update notifications
- auto-update opt-in (not mandatory)
- breaking changes MUST increment MAJOR version

**CLI:**

- semantic versioning
- `tractl update` command for manual update
- no mandatory auto-update
- version check on execution (configurable)

**Container:**

- image tag pinning is the update mechanism
- no in-container auto-update
- SBOM enables automated CVE scanning by consumers
- monthly base image refresh for `ci` and `slim` variants

**Browser Extension:**

- marketplace-managed auto-update
- major permission changes require explicit user re-consent
- extension version MUST be semantically aligned with the traCtl release version it targets

---

### 6 — Distribution and Execution Parity Relationship

Distribution channel MUST NOT introduce execution semantic differences.

Specifically:

- tractl CLI installed via Homebrew MUST behave identically to tractl CLI installed via curl installer
- tractl container MUST execute workflows with identical semantics to tractl CLI on the same workflow definition
- tractl Desktop MUST not expose execution capabilities unavailable to CLI

Distribution is a delivery concern.

Execution semantics are an architecture concern (ADR-009 §1).

No distribution channel is permitted to introduce a fast path, compatibility mode, or capability subset that diverges from canonical execution behavior.

---

## Consequences

**Positive:**

- Each surface has explicit distribution channels with declared requirements
- Container images have strong reproducibility guarantees via immutability and signing
- Browser extension permissions are explicitly scoped
- CLI version pinning enables reproducible CI execution
- Update lifecycle expectations are surface-appropriate

**Tradeoffs:**

- Maintaining multi-arch container images adds CI build complexity
- Browser extension marketplace review introduces release latency
- Code signing and notarization requirements add release infrastructure overhead
- SBOM attestation requires tooling investment

---

## Alternatives Considered

**Single distribution channel per surface:** rejected. Limits adoption. Different developer environments favor different package managers. Multiple channels with consistent semantics is correct.

**Runtime provider installation in containers:** rejected. Non-reproducible. Violates execution parity guarantee (ADR-009 §1).

**Browser extension auto-update of major permissions:** rejected. User consent is mandatory for permission changes. Silent permission expansion is a security violation.

**Snap/Flatpak as primary Linux channels:** deferred. Complexity of Snap confinement for a developer tool is disproportionate. AppImage + deb/rpm covers the primary Linux personas.

**CLI auto-update mandatory:** rejected. CI environments must not change behavior unexpectedly. Mandatory auto-update in a pinned CI context would break execution parity.
