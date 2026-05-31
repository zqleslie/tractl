
# traCtl — Agent Context

## Read First
Before writing any code, read the architecture documents in this repo:
- All `docs/adr/ADR-*.md` files (authoritative — highest priority)
- `docs/spec/tractl_spec.md` + overlay/yaml/json/toon spec files in `docs/spec/`
- `docs/spec/00_terminology.md`
- `docs/spec/02_hld.md` (high-level design)
- `docs/spec/tractl_roadmap.md` (phase status and milestones)
- `docs/spec/tractl_delivery_handoff_log.md` (update after every session)

Authority order: ADRs → Specs → Terminology → HLD → Roadmap

## Mandatory Guardrail
**Do not implement roadmap phases marked Not Started** (Phase 10+, Acquisition Layer, Extension Platform, etc.) unless the user explicitly requests that scope.

**Active implementation focus** (see `docs/spec/tractl_roadmap.md`):
- **Phase 7** — Diagnostics and Reporting (in progress: TOON output formatter, per-step diagnostics directives)
- **UI track** — Web Surface Foundation (in progress: persistence, status bar, server Tier 2)

Phases **1–6, 8, and 9** are **complete** in the engine (validation through script sandbox, parallel DAG, CLI harness). Do not remove or bypass pipeline stages. Do not re-implement completed milestones unless fixing a defect.

If any architecture conflict is found: **STOP** and report. Do not improvise.

## Project
tractl (TraceCtl) — open-source protocol-native API and workflow execution
platform. Module: `github.com/tractl/tractl`

## Prerequisites
- Go 1.26+
- Node.js 24+ (frontend)
- Wails v2 CLI (desktop builds only)

## Stack
- **Go 1.26+** — engine and binaries
- **CLI** — `cmd/tractl/`
- **Web App (Tier 1)** — `frontend/` + `cmd/wasm/` (Go → WebAssembly, browser `fetch`)
- **Web App (Tier 2, opt-in)** — `cmd/localapi/` HTTP API; React uses `frontend/src/platform/web/http/` when connected
- **Desktop** — `cmd/desktop/` (Wails v2; in-process engine via bindings, no HTTP in desktop binary)
- **Frontend** — `frontend/` shared React TypeScript for web + desktop (`build:web` vs `build:desktop`)
- **Core engine** — `internal/` — never exported outside this module
- **Public SDK** — `pkg/`

Surface rules (ADR-012, ADR-016): Desktop and Web App are independently delivered; they share source at build time only. Web Tier 1 must not require Desktop at runtime.

## Build Commands (GNU Make)

```
make setup                 install git hooks (once per clone)
make install               Go modules + frontend npm dependencies
make build-cli             CLI binary → ./bin/
make build-wasm            engine WASM → frontend/public/
make build-frontend-web    WASM + web UI bundle
make build-frontend-desktop  desktop UI bundle → cmd/desktop/frontend/dist
make build-desktop         Wails production app
make build                 cli + web frontend + desktop

make dev-cli               run CLI without building
make dev-web               WASM + Vite (Web App)
make dev-desktop           Wails + desktop frontend
make dev-localapi          localhost API :7428 (optional Tier 2 dev)

make test                  all Go tests
make test-phase1           Phase 1 validation tests only
make test-frontend         Vitest
make test-frontend-e2e     Playwright
make test-cover            Go coverage HTML report
make check                 vet + fmt-check + lint + test (CI gate)
make check-full            check + race + integration + e2e
make lint                  golangci-lint
make tidy                  tidy all go modules
make clean                 remove build artifacts
```

Windows: `winget install GnuWin32.Make` (or use WSL). Run `make help` for the full target list.

## Structure
```
cmd/tractl/           CLI entrypoint
cmd/wasm/             browser WASM bridge (Tier 1 web)
cmd/desktop/          Wails desktop shell (separate go.mod)
cmd/localapi/         optional HTTP server (Tier 2 web / tooling)

frontend/             shared React UI + platform adapters
  src/platform/desktop/   Wails bindings
  src/platform/web/wasm/  WASM runtime
  src/platform/web/http/  connected-server client

internal/spec/        canonical traCtlSpec Go structs
internal/validation/  contract enforcement (Phase 1)
internal/parser/      yaml / json / toon parsers (Phase 2)
internal/overlay/     overlay engine (Phase 3)
internal/planner/     execution planner (Phase 4)
internal/compiler/    DAG compiler (Phase 4)
internal/runtime/     runtime layer (Phase 5)
internal/executor/    HTTP execution + timeline
internal/scheduler/   parallel DAG scheduler (Phase 8)
internal/assertion/   assertion evaluator (Phase 6)
internal/extract/     extract engine (Phase 6)
internal/context/     execution context + expressions
internal/engine/      pipeline coordinator + CLI wiring
internal/sandbox/     JS script sandbox (Phase 9)
internal/diagnostics/ observability assembly (Phase 7, active)
internal/output/      renderers
internal/localapi/    localhost API handlers
internal/provider/    provider host + contract layer (stubs / future)
internal/normalize/   implicit dependency normalization

pkg/capability/       public capability contract types
```

Phases **not started** in engine: composite steps (10), extension platform (11), OpenAPI provider (12), fuzz (13), acquisition layer (14), multi-surface distribution packaging (15). Do not implement these without explicit direction.

## Conventions
- `internal/` packages are never imported outside this module
- No pipeline stage may be skipped or bypassed
- Validation must fail deterministically — no silent degradation
- Default authoring format: YAML
- CLI default output: concise human-readable
- Structured output: `--output json|yaml|toon`
- Exit codes are deterministic
- Every session must update `docs/spec/tractl_delivery_handoff_log.md`
