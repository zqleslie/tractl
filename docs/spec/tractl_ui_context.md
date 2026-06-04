# traCtl UI Implementation Context

Classification: UI Agent Reference — give to every UI agent session
Authority: ADR-016 · tractl_delivery_handoff_log.md (UI entries)
Supersedes: tractl_ui_screen_inventory.md
Last updated: 2026-06-03

---

## How to use this document

Give this file alongside `ADR-016` and `tractl_ui_spec.md` for every UI task.
When this document conflicts with `tractl_ui_spec.md`, **this document is correct** —
it records what was actually built. The spec has stale sections noted below.

---

## Spec Divergences — Read Before Using tractl_ui_spec.md

These sections in `tractl_ui_spec.md` no longer match the built implementation:

| Spec section | What spec says | What was actually built |
|---|---|---|
| Persistence model (top of file) | "No explicit Save button. Changes persist automatically." | Requests use History (automatic, max 50) + Collections (manual save via dialog). Save button in URL bar opens a collection dialog. Autosave label removed from URL bar. |
| S2 URL bar description | Shows `[ autosave indicator ]` as a URL bar element | URL bar has Send menu (dropdown) + Save button (opens dialog). No autosave indicator in the bar. |
| Design system tokens (bottom) | "Auto-save indicator replaces Save button everywhere." | Not true. Save opens a dialog. Export still via ··· menu. |
| Authority line | "Authority: ADR-016 (to be written)" | ADR-016 is written and accepted. |
| Companion line | "Companion: tractl_ui_screen_inventory.md" | Screen inventory is retired. Screen status lives in this document. |
| S6 Environment Manager | Described as full not-started screen | Basic env switcher + flat variable store was partially built in UI-0.2D. Full manager screen still not started but model needs alignment first. |

Everything else in the spec (S1, S2 panel tabs, S3, S9, design tokens, shell chrome) accurately reflects what was built.

---

## Technology Stack

| Concern | Choice |
|---|---|
| Desktop shell | Wails v2 (Go engine in-process, React in webview) |
| Frontend framework | React 18 + TypeScript (strict mode) |
| Styling | Tailwind CSS core utilities only — no component library |
| Icons | Tabler Icons outline only (`@tabler/icons-react`, SVG import only) |
| Code editor | Monaco Editor (pre/post script panels and S8 only — currently stubbed as textarea) |
| State management | Zustand (client) + TanStack Query (server/API) |
| Testing | Vitest (unit) + Playwright (e2e, requires `npx playwright install chromium` on fresh agents) |
| Build | Vite (frontend) + `make build-wasm` (Go WASM) |

---

## Three Execution Paths — Critical

### Desktop (Wails)
React calls Go via Wails-generated TypeScript bindings. Direct in-process call. No HTTP server, no port. Desktop binary: `cmd/desktop/`.

### Web Tier 1 (WASM — zero install)
Go engine compiled to WASM, runs in the browser tab. TypeScript calls via `window.tractl.*`. Loader: `frontend/src/platform/web/wasm/loadTractlWasmRuntime.ts`. CORS is a browser sandbox limit, not a traCtl limit. No DNS/TCP/TLS timing (browser fetch doesn't expose those hooks). `cmd/wasm/` imports `internal/localapi/compute` only — never `internal/localapi`.

### Web Tier 2 (Server)
`cmd/server` is the single server binary — static frontend + all engine API routes. `cmd/localapi` is removed (absorbed). Local dev: `TRACTL_ADDR=:7428`.

**Platform abstraction:** `frontend/src/platform/` routes to the correct path. Single `if` at `getRequestExecutionRunner` only — per ADR-016.

**Critical WASM constraint:** Blocking `net/http` inside `syscall/js.FuncOf` terminates the Go WASM runtime (exit code 2). Bridge MUST use `asyncPromiseHandler` — run blocking work in a goroutine, resolve Promise from that goroutine.

---

## Architectural Decisions (UI-relevant)

**D1** — `cmd/server` is the single server binary. `cmd/localapi` removed.
**D2** — `internal/localapi` split: `internal/localapi/compute` (WASM-safe, no socket) + `internal/localapi` (HTTP server layer, not WASM-safe).
**D3** — `cmd/wasm` imports `internal/localapi/compute` only. `net.Listen` unreachable from WASM import graph by construction.

**Monaco** — Stubbed as styled `<textarea>` in all pre/post script panels (S2 and S4). Wire in the Script Editor milestone (T-04).

**Implicit canvas edges** — `inferImplicitDependsOn` scans `${steps.<id>}` in `when`, `request.target`, headers, and body. `dependencyKind: 'implicit'` edges render dashed on canvas.

**Reuse contract** — All panel components under `components/panels/` are pure controlled components (`value` + `onChange` / row mutators). Ready for S2 and S4 without modification.

---

## Zustand Store Ownership

| Store | Owns |
|---|---|
| `workspaceStore` | Open files, active request/workflow, editor dirty state |
| `executionStore` | Current run state, step outcomes, result records |
| `environmentStore` | Configured environments, active environment |
| `historyStore` | Run history (mirrors API, paginated) — cap 50 on web |
| `requestCollectionsStore` | Folders + saved requests (manual save) |
| `workflowCanvasStore` | Canvas graph state, run outcome, step detail open state |
| `uiStore` | Theme, density, UI scale, sidebar state, active tabs, activity rail, tab bar |
| `openApiImportStore` | Imported OpenAPI specs (import-only, no asset generation) |

No global state for form fields. Form state is local to each panel component.

---

## Frontend Directory Structure

```
frontend/src/
  app/                         App.tsx, ThemeSync.tsx
  components/
    common/                    PanelTabBar, KeyValueTable, AddRow, SeverityBadge, EmptyState
    panels/                    Params, Headers, Body, Auth, Pre/Post-script,
                               Assertions, Extracts, Settings
                               (pure controlled — reused across S2 and S4)
    results/                   ResultsPanel, TimingBar, AssertionResult, ExtractResult
    request-editor/            RequestUrlBar, useRequestEditor, results/ResponseTabContent
    shell/                     ActivityBar, SidePanel, Topbar, TabBar, BottomToolbar,
                               CommandPalette, KeyboardShortcuts, AppShell
    sidebar/                   SidebarRequestsPanel, SaveToCollectionDialog
    primitives/                MethodBadge, EnvPill
  screens/
    RequestEditor.tsx          (re-exported as RequestEditorScreen.tsx)
    WorkflowCanvas/            WorkflowCanvasScreen + GraphView/ + StepDetailPanel/
    RunHistoryScreen/
    FastStartScreen/
  stores/                      (see Store Ownership above)
  platform/
    web/wasm/                  loadTractlWasmRuntime.ts
    web/http/                  stub for desktop surface guard
  lib/
    workflowCanvas/            workflowDocumentToCanvasWorkflow, inferImplicitDependsOn,
                               workflowSerializer.ts, workflowRunAdapter.ts
    history/                   groupHistoryByDate, historyEntryLabel
    fastStart/                 buildRecentItems, formatTimeAgo, recentItemActions
    tractlDocument/            inferDocumentFormat, pickTextFile, parseValidateDocument,
                               isOpenApiDocument, runDocument, buildRecentItems
  e2e/                         Playwright specs

cmd/
  wasm/
    main.go                    WASM entry (//go:build js && wasm)
    bridge.go                  Registers window.tractl — registration only (~40 lines)
    handlers.go                handleVersion, handleCapabilities, handleEcho,
                               handleParse, handleValidate, handleRun
    helpers.go                 asyncPromiseHandler, newPromise, bridgeError, etc.
    helpers_pure.go            parseDocument/validateDocument → delegate to engine.ParseByFormat
  desktop/
    app.go                     Wails binding stubs
  server/
    main.go, server.go, embed.go   Static frontend + engine API routes
```

---

## Local API Endpoints

```
POST   /api/v1/run                   Execute a request document
POST   /api/v1/workflows/run         Execute a workflow document
POST   /api/v1/validate              Validate a workflow document
POST   /api/v1/workflow/layout       Compute workflow DAG layout
POST   /api/v1/workflow/infer-deps   Infer implicit step dependencies
POST   /api/v1/workflow/export       Export workflow to format
GET    /api/v1/engine/defaults       Engine defaults
GET    /api/v1/workspace/status      Workspace status
GET    /api/v1/history               List run history
GET    /api/v1/history/:id           Get a specific run record
DELETE /api/v1/history               Clear all history
GET    /api/v1/files                 List workspace files
GET    /api/v1/files/:id             Read a file  ← TODO: not yet implemented
POST   /api/v1/files                 Create/update a file
GET    /api/v1/environments          List configured environments
GET    /api/v1/status                Engine health + version
GET    /api/v1/credentials           List credential names (never values)
POST   /api/v1/credentials           Store credential in OS keychain
DELETE /api/v1/credentials/:id       Remove a credential
```

---

## WASM Bridge API

```js
await window.tractl.version()          // { version, commit, buildDate }
await window.tractl.capabilities()     // string[]
await window.tractl.parse(doc, fmt)    // canonical spec | TRACTL_PARSE_ERROR
await window.tractl.validate(doc, fmt) // { valid, errors? } | TRACTL_WASM_INVALID_FORMAT
await window.tractl.run(doc, fmt)      // RunResult + { surface, executionMode, networkProvider }
```

Accepted formats: `yaml`, `yml`, `json`, `toon`.
Error codes: `TRACTL_WASM_INVALID_ARGUMENT`, `TRACTL_WASM_INVALID_FORMAT`, `TRACTL_EXECUTION_ERROR`, `TRACTL_WASM_PANIC`.

---

## Make Commands

```bash
make build-wasm          # Compile Go → frontend/public/tractl.wasm
make build-web           # Build frontend → cmd/server/dist/web/
make build-server        # Build cmd/server binary (run build-web first)
make dev-web             # Vite dev server
make dev-desktop         # Wails dev server
make embed-stub          # Minimal index.html for go:embed (CI / fresh clones)
npm --prefix frontend run test         # Vitest unit tests
npm --prefix frontend run build:web    # Production build
```

---

## Screen Status

| ID | Screen | Status | What exists / What's missing |
|---|---|---|---|
| S1 | App Shell + Fast Start | ✓ Complete | Shell v4, activity bar, command palette, keyboard shortcuts, env badge |
| S2 | Request Editor | ✓ Complete | Full panel suite, results, density system. Monaco **stubbed as textarea**. |
| S3 | Workflow Canvas | ✓ Complete | DAG, WASM + desktop run, implicit edges (dashed), step insertion, variable reference hints |
| S4 | Step Detail Panel | ⚠ Shell only | 7-tab shell exists. **Step persistence and result tab are TODO** in `useStepDraftEditor`. |
| S5 | Workflow Run Results | ✗ Not started | Engine produces full data. Pure rendering task. (T-20) |
| S6 | Environment Manager | ✗ Not started | Basic env switcher exists. Full screen needs model alignment first. (T-03) |
| S7 | Overlay Editor | ✗ Not started | No design or implementation. (T-10) |
| S8 | Script Editor | ✗ Not started | Monaco stubbed everywhere. Full screen needed. (T-07) |
| S9 | Run History | ✓ Complete | History (auto, 50 max, grouped by date) + Collections (manual, folders) |
| S10 | Settings | ✗ Not started | State in `uiStore` already. Needs a form screen. (T-08) |
| S11 | Credential Manager | ✗ Deferred | After S10. (T-36) |

---

## Open TODOs (coded in source)

- `useStepDraftEditor` — step persistence and workflow auto-save explicitly TODO. Edits don't save back to YAML.
- Monaco — stubbed as `<textarea>` in all pre/post script panels (S2 and S4). Wire in T-04.
- Desktop file load — `GET /api/v1/files/:id` not implemented. Desktop canvas falls back to in-memory + logs warning.
- Sidebar lists — workflow/request lists partially use alpha fixtures. T-01 wires real workspace files.
- Pre-seeded MOCK_WORKFLOW step results on canvas — retained for demos, deferred cleanup.
- Baseline lint failures (pre-existing, do not fix in unrelated sessions): `e2e/wasm-run.spec.ts`, `VariableReferenceHints.tsx`, `EnvironmentManager.tsx`, `useStepDraftEditor.ts`, `workflowCanvasStore.ts`, generated wailsjs runtime types.

---

## Deferred Canvas Product Decisions

- Remove step: UX and dependsOn rewiring TBD.
- Multiple workflows on canvas: FilterBar is single-workflow today.
- Reorder/move steps: drag-and-drop TBD.
- Implicit deps in DependenciesTab: show separately from explicit dependsOn (non-blocking).

---

## Completed Session Log

**UI-0.1B** — WASM Bootstrap. `window.tractl` registered. `cmd/wasm/bridge.go`, `loadTractlWasmRuntime.ts`, `make build-wasm`.

**UI-0.1C** — JS ↔ Go data exchange. Echo Promise API. Panic recovery via `recoveringPromiseHandler`. Playwright boundary tests (1–500 KiB, Unicode).

**UI-0.1D** — Parser pipeline. `window.tractl.parse(doc, format)`. WASM size ~4.8 MiB after parser linkage.

**UI-0.1E** — Canonical validation. `window.tractl.validate(doc, format)`. Format validate + parser + SpecValidator. WASM ~4.9 MiB.

**UI-0.1F** — Browser engine execution. `window.tractl.run(doc, format)`. Full engine pipeline in browser. `asyncPromiseHandler` pattern established. Sequential WASM scheduler (`scheduler_js.go`) for event-loop safety. DNS/TCP/TLS disabled on `GOOS=js`.

**WASM Bridge Refactor** — `bridge.go` trimmed to ~40 lines. `handlers.go` + `helpers.go` + `helpers_pure.go` split. Format validation added to all handlers (`TRACTL_WASM_INVALID_FORMAT`). `parseDocument`/`validateDocument` delegate to `engine.ParseByFormat`.

**UI-0.2A** — Request Editor → WASM wired via `RequestExecutionRunner` factory. Single platform `if` established.

**UI-0.2B** — Fast Start. File picker WASM run. OpenAPI import plumbing (import-only). Persisted recent list (`buildRecentItems`).

**UI-0.2C** — Run History. Zustand persist (localStorage). `RunHistoryScreen`. Reopen hydrates Request Editor.

**UI-0.2D** — Environment selection. Env switcher in titlebar. `{{name}}` resolver before execution. Flat string variables only. Missing variables → `TRACTL_ENV_VARIABLE_MISSING`.

**UI-0.3B** — Step Detail Panel shell. 7-tab shell. `useStepDraftEditor` (no persistence). Reuses S2 panel primitives. Canvas dims when panel open.

**UI-0.3C** — Workflow Run Results (inline). `ResultBar` (collapsible step chips) + `FullResultsPopup` (summary cards + ASCII waterfall). ASCII waterfall matches engine `WaterfallText` format.

**UI-0.3D** — Canvas WASM run wiring. `workflowSerializer.ts` (canvas → minimal YAML). `workflowRunAdapter.ts` (WASM RunResult → canvas StepResult). Custom YAML emitter (js-yaml not yet added).

**UI-0.3E** — Canvas desktop run. `POST /api/v1/workflows/run` added. `getWorkflowRunRunner()`, `buildWorkflowRunPayload()`, `mapRunResult`. Desktop file load (`GET /api/v1/files/:id`) still TODO.

**Shell v4** — ActivityBar (48px), SidePanel (collapsible), Topbar (centered command, right env), TabBar, BottomToolbar, CommandPalette, KeyboardShortcuts. Body modes: JSON/Form data/Multipart/Raw/Binary/None. Wails binding stubs in `cmd/desktop/app.go`.

**Shell UX refinements** — Centered command trigger. Env switcher right + outside-click close. Activity rail icon-only + expandable. Request tabs rename after Run. JSON response pretty/raw toggle.

**Shell scale control** — Persisted `uiScale` in `uiStore`. `data-ui-scale` CSS variable. Tailwind `text-ui-*` tokens from CSS vars. Scale upward menu in BottomToolbar.

**Phase 15.2 — Request Editor (S2)** — Full screen with stacked/side-by-side layouts, draggable split, results collapse. Route sync `/request/:id` and `/request/new`. 127 Vitest tests. All panel components pure controlled — ready for S4 reuse.

**Requests History/Collections Sidebar** — `requestCollectionsStore.ts` (NEW). History/Collections tabs. `SaveToCollectionDialog`. Send menu + Save button. 132 tests passing.

**ADR-SERVER-CONSOLIDATION** — D1/D2/D3 decisions recorded. No Go code changed. CHANGE_PLAN.md for server consolidation + localapi partition still pending implementation.

---

## Bug Fixes Applied

**UI-BUG-01** — `requestUrl` added to `StepResult`. Mapped from `diagnostics.Workflows[].Steps[].Requests[].URL`. ResultTab shows resolved vs template URL when they differ.

**UI-BUG-02** — `extractCount` from `step.extracts?.length`. Diamond example at `examples/yaml/04-diamond-dependency.yaml`. `inferImplicitDependsOn` scans `${steps.<id>}` in `when`, `request.target`, headers, body. Implicit edges render dashed.

**UI-BUG-04** — `loadCanvasWorkflow` resolves workflows from `workflowWorkspaceStore`. Web: FileReader + WASM parse/validate. Desktop: in-memory only + warning log (no file load API yet).

**UI-BUG-05** — `insertStepAfter`, `insertParallelStep`, `insertStepBetween`. `updateStep`/`renameStep`. `persistCanvasWorkflow`. Serializer preserves spec `variables` + step `assertions`/`extracts`.

**UI-BUG-06** — `${...}` reference chips on step form fields. Collapsible Variables panel (`${vars.*}`, upstream extracts, `${runtime.*}`).

**UI-BUG-07** — Waterfall Gantt offsets from `WorkflowRecord.StartedAt`/`StepRecord.StartedAt` → `StepResult.startMs`. Topology+duration fallback in `deriveRunSummary`. CLI waterfall shows without `--verbose` when diagnostics present.

---

## Decision Log

| Date | Decision | Reason |
|---|---|---|
| Architecture Freeze | Multi-agent governance mandatory | Prevent semantic drift |
| 2026-05-24 | Browser surface scoped to basic HTTP + agent/proxy | Web = nice-to-have; desktop + CLI non-negotiable |
| 2026-05-24 | Storage adapter model per surface | Desktop opaque, web IndexedDB, CLI explicit file |
| 2026-05-29 | `asyncPromiseHandler` mandatory | Blocking `net/http` in `js.FuncOf` deadlocks WASM runtime |
| 2026-05-29 | Custom YAML emitter for canvas serializer | `js-yaml` not yet in frontend deps |
| 2026-05-29 | Single `if` at `getRequestExecutionRunner` | Platform factory boundary per ADR-016 |
| 2026-05-31 | D1: `cmd/server` single server binary | `cmd/localapi` removed |
| 2026-05-31 | D2: `internal/localapi` partitioned | WASM-safe boundary by package |
| 2026-05-31 | D3: `cmd/wasm` imports compute only | `net.Listen` unreachable from WASM by construction |
| 2026-05-31 | Monaco wired in S8 milestone only | Pre/post panels remain textarea stubs until T-04 |
| 2026-05-31 | Pre-seeded mock results retained on canvas | Useful for demos; deferred cleanup |

---

## Test Baseline

- Vitest: 32 files, 132 tests passing
- Playwright: core WASM matrix passing
- Pre-existing lint failures: do not fix in unrelated sessions (listed in Open TODOs above)
