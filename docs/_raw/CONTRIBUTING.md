# Contributing to traCtl

Thank you for your interest in contributing. traCtl has a frozen architecture and a phased implementation roadmap. Reading this document before contributing will save you significant time.

---

## Before You Start

- **Read [AGENTS.md](AGENTS.md)** — it contains the architecture baseline, the active phase guardrail, and the authority order every contributor must follow.
- **Read the ADRs** in `docs/adr/` — ADR-001 through ADR-016 are accepted and normative (see [ADR-000-index](docs/adr/ADR-000-index.md)). Implementation must not contradict them. If your code contradicts an ADR, your code is wrong, not the ADR.
- **Architecture is frozen.** Implementation follows the phased roadmap in [`docs/spec/tractl_roadmap.md`](docs/spec/tractl_roadmap.md). Do not implement beyond the active phase and milestones declared in `AGENTS.md`.
- **If you find an architecture conflict** — open a GitHub issue labelled `arch-conflict`. Do not work around it. Do not improvise.

---

## Development Setup

### Prerequisites

- Go 1.26+
- Node.js 24+ (shared React frontend in `frontend/`)
- GNU Make
- Wails v2 CLI (desktop builds only): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Clone and set up

```bash
git clone https://github.com/tractl/tractl.git
cd tractl
make setup        # installs git hooks (run once after every clone)
make install      # Go modules + frontend npm dependencies
make build-cli
make test
```

### Optional: Web and Desktop surfaces

```bash
make build-wasm           # Go engine → frontend/public/tractl.wasm
make dev-web              # WASM + Vite dev server (Web App Tier 1)
make dev-desktop          # Wails desktop + embedded frontend
make test-frontend        # Vitest unit/component tests
make test-frontend-e2e    # Playwright (requires dev-web or built WASM)
```

Web Tier 1 runs the engine in-browser via WebAssembly with no Desktop app required. Desktop uses Wails bindings to the in-process Go engine. Optional `make dev-localapi` serves `cmd/localapi/` on `:7428` for extended web development scenarios.

`make setup` copies `scripts/pre-commit` into `.git/hooks/`. The hook runs the same `make` targets as CI (`fmt-check`, `vet`, `build-cli`, `test`) with `GOWORK=off` so an incomplete root `go.sum` cannot be masked by `go.work`. If `golangci-lint` is not installed it warns but does not block — install it for full parity with CI:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Architecture Rules

- **Active phase** is declared in `AGENTS.md`. Do not implement beyond it.
- Every ADR in `docs/adr/` is normative. Read the relevant ADR before writing code in any area it governs.
- **Terminology must match [`docs/spec/00_terminology.md`](docs/spec/00_terminology.md) exactly.** Do not introduce synonyms. Common drift vectors to avoid: workflow/job/run/task, provider/plugin/adapter/connector, capability/feature/tool/function.
- The canonical pipeline stages are fixed and ordered. No stage may be skipped, collapsed, or bypassed. See [`docs/spec/02_hld.md`](docs/spec/02_hld.md) and `AGENTS.md` for the pipeline map.
- **UI work** must follow [ADR-016](docs/adr/ADR-016-ui-architecture-model.md) and [ADR-012](docs/adr/ADR-012-multi-surface-delivery-model.md): shared React in `frontend/`, transport-agnostic components, and surface-specific adapters under `frontend/src/platform/`.
- `internal/` packages are never imported outside this module. The public SDK surface lives in `pkg/` only.
- Validation must fail deterministically. No silent degradation.

---

## Code Standards

- **Idiomatic Go** — follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) guide.
- All exported symbols must have godoc comments.
- No global mutable state.
- Errors must be explicit — no `panic` in library code.
- Table-driven tests for all validators.
- Run `make check` before every commit — it must pass clean.
- Changes under `frontend/` should also pass `make test-frontend` (and `make test-frontend-e2e` when behavior crosses the WASM bridge).

---

## Branch and Commit Convention

**Branch names** follow `type/short-description`:

```
feat/spec-validator
fix/cycle-detection
test/overlay-validation
refactor/normalize-pipeline
docs/contributing-guide
chore/makefile-targets
```

**Commits** follow [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Use for |
|--------|---------|
| `feat:` | new feature |
| `fix:` | bug fix |
| `test:` | tests only, no behavior change |
| `refactor:` | no behavior change |
| `docs:` | documentation only |
| `chore:` | build, tooling, config |

---

## Pull Request Process

1. **Open an issue first** for any non-trivial change. Get alignment before writing code.
2. Reference the **roadmap phase and milestone** in your PR description.
3. All tests must pass: `make check` (add `make test-frontend` / `make test-frontend-e2e` for UI changes).
4. No architecture rules may be violated. PRs that contradict the frozen architecture will be closed with an explanation.

---

## Issue Reporting

- **Bug** — use the bug report template. Include OS, Go version, Node version (if UI-related), delivery surface (CLI / Web / Desktop), and reproduction steps.
- **Feature request** — use the feature request template. Include the roadmap phase it aligns with.
- **Architecture conflict** — label as `arch-conflict`. This is high priority and will be reviewed promptly by a maintainer.

---

## Questions

Open a [GitHub Discussion](https://github.com/tractl/tractl/discussions) rather than an issue for general questions and design conversations.
