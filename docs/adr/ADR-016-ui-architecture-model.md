# ADR-016 — UI Architecture Model

Status: Accepted
Date: 2026-05-29

---

## Context

traCtl's engine is complete through Phase 9. Phases 1–9 deliver a fully functional
execution pipeline accessible only via the CLI binary. The product promise — desktop-first
developer experience with CLI parity — requires a UI layer that exposes every engine
capability as a form-based interface without requiring users to author YAML by hand.

Three decisions require formal locking before UI implementation begins:

1. The desktop shell technology
2. The web frontend framework and its relationship to the desktop shell
3. State management and the API contract between UI and engine

Without locking these, multi-agent implementation will diverge on framework assumptions,
import boundaries, and data flow — the same class of semantic drift the handoff log
was designed to prevent.

---

## Decision

### 1 — Desktop Shell: Wails v2

The desktop application uses Wails v2 as the shell.

Rationale:

- Wails embeds a web frontend (React) inside a native OS window using the system
  webview (WKWebView on macOS, WebView2 on Windows, WebKitGTK on Linux)
- The Go engine runs in-process — no separate server, no IPC latency, no extra install
- The same React codebase serves both the desktop shell and the web surface
- Wails exposes Go functions to the frontend via generated TypeScript bindings —
  the engine API surface is typed end-to-end
- Build output is a single native binary per platform — consistent with the CLI
  distribution model and goreleaser pipeline already planned in ADR-015

Wails is explicitly not Electron. There is no Node.js runtime in the desktop binary.
The Go engine and the Wails shell are compiled together. This preserves the
single-binary distribution invariant established in ADR-009.

Platform targets matching ADR-015 §1:
- macOS (arm64 + amd64, signed DMG, Sparkle auto-update, Apple notarization)
- Windows (amd64, MSIX, Authenticode signed)
- Linux (amd64 + arm64, AppImage, .deb, .rpm)

### 2 — Web Frontend: React 18 + TypeScript

The UI is implemented in React 18 with TypeScript (strict mode).

Rationale:

- The primary contributor has a Node.js background — React is the lowest-friction
  capable framework for this team
- React's component model maps cleanly to the traCtl UI spec screen inventory:
  each screen is a top-level component, each panel is a composed sub-component
- TypeScript strict mode enforces the same discipline at the UI layer that Go's
  type system enforces in the engine — no implicit any, no optional chaining on
  unvalidated API responses
- The React codebase is shared between the Wails desktop shell and the web surface
  served by the local API — one codebase, two delivery targets

**Styling: Tailwind CSS (core utilities only)**

No component library is used. All UI components are implemented from scratch per
the tractl_ui_spec.md design system. Tailwind provides the utility layer.

Rationale: A component library (shadcn, MUI, Ant Design) would impose its own
design language and override the minimal flat aesthetic defined in the spec.
Building from scratch is more work upfront but eliminates the ongoing cost of
fighting a third-party design system.

**Code editor: Monaco Editor**

The Pre-script and Post-script panels use Monaco Editor (the VS Code editor engine).
Monaco is the only third-party UI component permitted in the editor panels.
All other panels are form-based — no Monaco elsewhere.

**Icon set: Tabler Icons (outline only)**

Consistent with the mockup phase. Imported via @tabler/icons-react as SVG components.
No icon font — SVG import only to avoid FOUT and loading artifacts.

### 3 — Local API: Go net/http + chi router

The engine is exposed to the UI via a local HTTP API running on localhost.

This API serves two purposes:
- In the Wails desktop shell: Wails bindings call Go functions directly (no HTTP).
  The local API runs alongside for the web surface to consume.
- In the web surface (browser tab): the browser connects to the local API at
  localhost:7428 (default, configurable). The desktop app must be running.

The local API is not a public API. It binds to localhost only. No auth is required
for alpha — the loopback interface is the trust boundary for local development.

Router: chi (lightweight, idiomatic Go, no framework overhead)
Serialization: encoding/json (stdlib)
No OpenAPI spec for the local API in alpha — TypeScript types are generated from
Go structs via a code generation step.

**Core endpoints (alpha):**

```
POST   /api/v1/run              Execute a workflow or request
POST   /api/v1/validate         Validate a workflow document
GET    /api/v1/history          List run history
GET    /api/v1/history/:id      Get a specific run record
DELETE /api/v1/history          Clear all history
GET    /api/v1/files            List workspace files
GET    /api/v1/files/:id        Read a file
POST   /api/v1/files            Create/update a file (auto-save)
GET    /api/v1/environments     List configured environments
GET    /api/v1/status           Engine health + version
```

Credential endpoints (keychain-backed, alpha):
```
GET    /api/v1/credentials      List credential names (never values)
POST   /api/v1/credentials      Store a credential in OS keychain
DELETE /api/v1/credentials/:id  Remove a credential
```

### 4 — State Management: Zustand

Client-side state uses Zustand (lightweight, no boilerplate, no Redux ceremony).

State is divided into stores with clear ownership boundaries:

- **WorkspaceStore**: open files, active request/workflow, editor dirty state
- **ExecutionStore**: current run state, step outcomes, result records
- **EnvironmentStore**: configured environments, active environment
- **HistoryStore**: cached run history (mirrors local API, paginated)
- **UIStore**: theme (light/dark), sidebar collapsed state, active panel tabs

No global state for form fields — form state is local to each panel component.
Forms serialize to the canonical traCtlSpec shape on submit/auto-save, not on
every keystroke.

Server state (API responses) is managed via TanStack Query (React Query).
TanStack Query handles caching, background refresh, and optimistic updates for
API calls. Zustand handles client-only UI state. These are complementary, not
competing.

### 5 — Credential Storage: zalando/go-keyring

User credentials (API keys, bearer tokens, client certificate private keys) are
stored in the OS-native keychain via the zalando/go-keyring library.

- macOS: Keychain Services
- Windows: DPAPI (CryptProtectData)
- Linux: Secret Service API (D-Bus / libsecret)

Namespacing: all entries keyed as `tractl/<environment-name>/<credential-id>`

The web UI never receives raw credential values. It sends credential names to the
local API, which resolves them from the keychain at execution time. The API never
returns credential values in responses — only names and types.

This satisfies ADR-006 §3 (secret access policy) and §13 (certificate trust constraints).

### 6 — Persistence Model

**Auto-save:** All requests and workflows are auto-saved as YAML files to the
traCtl app directory on first change. No explicit save action is exposed in the UI.

App directory locations:
- macOS:   ~/Library/Application Support/tractl/
- Windows: %APPDATA%\tractl\
- Linux:   ~/.config/tractl/   (XDG Base Directory compliant)

**Run history:** Stored in a local SQLite database in the app directory.
- Desktop default retention: 1000 runs
- Web default retention: 50 runs
- Configurable in Settings

**Export:** User-initiated export to a chosen path via the ··· context menu.
Formats: YAML (default), JSON, TOON. This is the only action that prompts for
a file location.

**Web surface storage:** IndexedDB for UI state (active tab, sidebar state).
Run history and workflow files are stored by the local desktop API — the web
surface does not own persistent storage for execution artifacts.

### 7 — Web Surface Constraint

The web surface (browser tab) requires the traCtl Desktop application to be running.

The browser tab connects to the local API at localhost:7428. If the desktop app
is not running, the web surface shows a "Start traCtl Desktop to continue" screen.

This constraint is architectural, not a limitation to be worked around. It preserves:
- OS keychain access (desktop process only)
- Full execution capability (engine runs in the desktop process)
- Execution semantic parity (same engine, same binary, regardless of surface)

The web surface is explicitly not a standalone product. It is a browser-accessible
view of the desktop application for users who prefer not to switch windows.

CORS: the local API sets `Access-Control-Allow-Origin: http://localhost:*` only.
No wildcard origins. No external origins permitted.

---

## Consequences

**Positive:**
- Single React codebase serves desktop and web — no duplication
- Go engine runs in-process in the desktop shell — no IPC latency
- Wails binary is self-contained — consistent with CLI distribution model
- TypeScript strict mode + Go types = typed API surface end-to-end
- OS keychain integration satisfies ADR-006 security requirements
- SQLite for history is zero-infrastructure, portable, and fast for local queries

**Tradeoffs:**
- Wails webview rendering differs slightly per OS — testing required on all three
- Monaco Editor adds ~2MB to the bundle — acceptable for a desktop app, noted for web
- Web surface requires desktop app running — documented constraint, not a bug
- No component library means higher upfront UI implementation cost

**Explicitly deferred:**
- Cloud sync for history and files (paid tier, requires account model)
- Browser extension (v1.5, ADR-012 §5)
- Hardware security key credential storage (v1+)
- Remote secret management (Vault, AWS Secrets Manager) via Phase 11 extension
- WebSocket for real-time execution streaming (alpha uses polling on /api/v1/run)

---

## Alternatives Considered

**Electron:** rejected. Requires Node.js runtime in the binary, increases bundle
size significantly, and adds a second runtime dependency alongside Go. Wails
achieves the same result (native window + web UI) with the Go runtime already
present.

**Tauri (Rust):** rejected. Requires a Rust compilation step alongside Go. The
engine is Go; the contributor has no Rust background. Wails keeps the entire
backend in one language.

**Next.js / full-stack framework:** rejected. traCtl's UI has no server-side
rendering requirement and no need for a framework router. React + Vite is
sufficient and avoids framework-imposed conventions that conflict with the
Wails embedding model.

**Redux / MobX for state:** rejected. Zustand + TanStack Query covers the state
requirements without ceremony. Redux adds indirection with no benefit at this
scale. MobX's observable model conflicts with React's explicit re-render model.

**REST API with OpenAPI spec:** deferred to v1. For alpha, TypeScript types are
generated from Go structs directly. OpenAPI spec and API versioning governance
are v1 concerns when external consumers may exist.

---

## Amendment — 2026-05-29

Amended by: Web Surface Model Correction

### Context

ADR-016 §3 and §7 defined the web surface as a companion to the Desktop
application that required the Desktop app to be running. This model was
incorrect for two reasons:

1. It created an unnecessary hard dependency between two independently
   delivered products.
2. It failed to account for a viable browser-native execution path via
   WebAssembly (WASM).

### Corrections

#### Correction to §3 — Engine Communication Model

ADR-016 §3 is superseded by three clearly separated paths.

**Desktop path — Wails bindings, no HTTP**

- React UI calls Go engine via Wails-generated TypeScript bindings
- Direct in-process call through the Wails bridge
- No HTTP server, no port, no network hop
- chi HTTP server is NOT bundled into the Desktop binary
- Desktop binary structure: Wails shell + React UI (embedded webview) + Go engine (in-process)
- One binary, one process, no server component

**Web Tier 1 — WebAssembly, no server required**

- Go engine compiled to WASM, bundled into the static web application
- Engine runs inside the browser tab — no Desktop app, no server, no install
- Compiled as a separate build target: `GOARCH=wasm GOOS=js go build`
- TypeScript calls WASM engine via Go WASM runtime bridge (syscall/js)
- Platform abstraction layer routes calls to WASM bridge on web build

What works in Tier 1 (browser-native):

- Full workflow and request authoring
- HTTP request execution via fetch() — status, header, body assertions work
- In-browser persistence via IndexedDB (YAML content, run history)
- Full engine pipeline: validation, planning, compilation, runtime

Browser sandbox limitations in Tier 1 (not traCtl limitations):

- CORS: target APIs must permit browser origins
- Raw socket timing: DNS/TCP/TLS not exposed by browser fetch(). ttfb and
  transfer available via PerformanceResourceTiming API. dns/tcp/tls fields
  omitted from Tier 1 results.
- OS keychain unavailable: credentials use browser credential storage
- Filesystem unavailable: workspace files stored in IndexedDB. Export/import
  via browser file picker (File System Access API where supported)

**Web Tier 2 — connected mode, user opt-in, no default**

- User optionally connects web app to a running traCtl server
- Server options: local (`tractl server` on localhost:7428), Docker, self-hosted, hosted (future paid tier)
- When connected, platform abstraction routes engine calls to server via HTTP instead of WASM
- React components unchanged — transport is hidden by platform abstraction

What Tier 2 unlocks:

- No CORS restrictions (server makes HTTP calls, not the browser)
- Full raw socket timing (DNS, TCP, TLS, TTFB, transfer)
- OS keychain access (when server is local)
- Filesystem-backed YAML storage
- Full observability data

Connection is user-initiated. Web app runs fully in Tier 1 by default.
Tier 2 surfaced as optional "Connect to server" action in Settings.

**chi HTTP server — server binary only**

- Lives in `cmd/server/`
- Compiled as standalone binary: `tractl server`
- NOT bundled into Desktop binary
- Distributed for Docker, self-hosted, CI, and hosted use cases

#### Correction to §7 — Web Surface Constraint

ADR-016 §7 is superseded by this correction.

The web surface has no mandatory dependency on the Desktop application.

The web surface is an independently delivered product:

- Served as static web application (React + WASM bundle) from a domain
- Fully functional in any browser with no install, no Desktop app, no server
- Optionally connectable to a traCtl server for extended capabilities
- No relationship to the Desktop application at runtime
- Shares source code with Desktop at build time only

#### Additional Decision — Platform Abstraction Layer

The `platform/` module in the frontend must implement three transport paths,
selected at build time via compile flags:

| Path | Transport |
|---|---|
| `platform/desktop/` | Wails bindings (`window['go']` bridge) |
| `platform/web/wasm/` | WASM Go bridge (`syscall/js`) |
| `platform/web/http/` | `fetch()` to connected traCtl server (Tier 2) |

The active transport in the web build switches from WASM to HTTP when the user
connects to a server. This switch is runtime state, not a build flag.
Desktop always uses Wails bindings only.

### Consequences

**Positive:**

- Web surface is a standalone product — no Desktop dependency
- WASM enables zero-infrastructure browser-native execution
- Desktop binary is leaner — no embedded HTTP server
- Server binary (`tractl server`) is a clean, separate deployment artifact
- Three transport paths share one React codebase and one Go engine codebase

**Tradeoffs:**

- WASM compilation is an additional build target requiring maintenance
- WASM bundle size adds to web app load time (~several MB — acceptable for a
  developer tool)
- Two persistence models: IndexedDB (web Tier 1) vs filesystem (Desktop, Tier 2)
  — data is not automatically shared between surfaces
- CORS limitations in Tier 1 are a real constraint for some users — documented
  clearly, not hidden

### Superseded Content

The following from the original ADR-016 is superseded by this amendment:

- §3 original text: "The local API serves two purposes… In the Wails desktop shell:
  Wails bindings call Go functions directly (no HTTP). The local API runs alongside
  for the web surface to consume."
- §7 original text: entire section — "The web surface requires the traCtl Desktop
  application to be running."
