# ADR-012 — Multi-Surface Delivery Model

Status: Accepted
Date: 2026-05-24

---

## Context

traCtl's core execution promise — same workflow definition, same execution semantics, across all execution environments — has always been an architectural invariant. However, the existing ADR set implicitly assumed a narrow set of execution surfaces (desktop, CLI, CI) without formally defining surfaces as first-class architectural citizens.

As traCtl expands toward browser capture, container-native execution, and headless automation, the absence of a formal surface model creates compounding risks:

- UX behavior gets conflated with execution engine behavior
- Interactive surfaces accumulate assumptions not safe for automation surfaces
- Execution parity guarantees become implicit and untestable
- New surfaces get implemented without declared behavioral contracts

This ADR formalizes delivery surfaces, surface categories, invariants, and parity guarantees.

---

## Decision

### 1 — Canonical Delivery Surfaces

traCtl defines five first-class delivery surfaces.

**Desktop**

Primary interactive workflow surface.

- native application (macOS, Windows, Linux)
- designed for backend engineers authoring, running, and inspecting workflows
- full UX: workflow editor, execution output, diagnostics visualization, Git integration
- human-in-the-loop interaction model

**Browser Extension**

Capture-assisted interactive surface.

- installed browser extension (Chrome, Firefox, Edge)
- primary function: traffic capture and workflow bootstrapping
- secondary function: request inspection during live browsing sessions
- NOT a standalone execution surface
- execution delegated to Desktop or CLI

**CLI**

Canonical automation-compatible execution surface.

- tractl binary distributed as a standalone executable
- primary function: workflow execution with deterministic outputs
- designed for script, CI, and local automation consumption
- no interactive UX dependencies
- structured output (JSON, YAML, TOON) required

**CI**

Headless execution surface embedded in continuous integration pipelines.

- executes same workflow definitions as CLI
- no interactive UX
- deterministic exit codes mandatory
- structured output mandatory
- execution semantics must match CLI

**Container Runtime**

Portable isolated execution surface.

- OCI-compliant container image
- designed for CI runners, sidecar execution, and reproducible environments
- execution semantics must match CLI and CI
- no host dependency assumptions permitted
- immutable execution image behavior required

---

### 2 — Surface Category Classification

Surfaces divide into two categories.

**Interactive surfaces:** Desktop, Browser Extension

Characteristics:

- human-in-the-loop interaction model
- real-time visual feedback
- UX state management
- workflow authoring support
- exploratory execution model

**Automation surfaces:** CLI, CI, Container Runtime

Characteristics:

- no human-in-the-loop dependency
- deterministic inputs and outputs
- structured machine-consumable output
- scripting and pipeline composition compatibility
- headless execution model

The distinction between surface categories is load-bearing.

UX behavior, state management, and visual feedback are implementation concerns of interactive surfaces only.

Execution semantics, pipeline behavior, and output contracts are shared across all surfaces without exception.

---

### 3 — UX Boundary vs Execution Engine Boundary

The execution engine MUST be decoupled from all surface UX concerns.

The execution engine boundary is defined as:

```
Canonical Workflow Specification (traCtlSpec)
    ↓
Execution Engine (Planner → Compiler → Runtime)
    ↓
Execution Result
```

Everything above and below this boundary is surface-specific.

UX surfaces consume the execution engine through a defined interface.

The execution engine MUST NOT:

- reference surface-specific state
- assume interactive execution context
- depend on UX lifecycle events
- expose surface-specific APIs

Surface implementations MUST NOT:

- bypass the canonical pipeline (ADR-003 §2)
- implement parallel execution logic
- duplicate normalization behavior
- maintain independent execution state

Rationale: The execution engine is the product moat. Surface proliferation must not fragment execution semantics. Coupling the engine to any surface creates a ceiling on how many surfaces can be safely delivered.

---

### 4 — Surface Parity Guarantees

The following invariants are non-negotiable across all five surfaces.

**Execution semantic parity:**

The same workflow definition MUST produce equivalent execution behavior regardless of which surface executes it.

Equivalent means: same normalization, same planning, same compilation, same runtime behavior, same assertion outcomes, same failure semantics.

Timing may vary across surfaces. Semantic behavior may not.

**Output contract parity:**

Structured output formats (JSON, YAML, TOON) MUST be semantically equivalent across CLI, CI, and Container surfaces.

Desktop MAY present richer visual output but MUST NOT alter the underlying semantic result.

**Failure semantic parity:**

Exit codes, error categories, and failure propagation behavior MUST be consistent across automation surfaces (CLI, CI, Container).

Interactive surfaces MUST map execution failures to equivalent UX states without altering the failure semantics.

**Capability resolution parity:**

Capability contracts available on one surface MUST be available on equivalent surfaces.

Surface-specific capability exclusions MUST be declared explicitly and fail at planning time with a diagnostic, not silently degrade.

---

### 5 — Browser Extension Execution Model

The Browser Extension surface requires explicit scope constraints.

The Browser Extension:

- MUST NOT execute traCtl workflows directly
- MUST delegate execution to Desktop or CLI
- MUST function as a capture and bootstrapping surface only
- MUST export captured traffic as normalized traCtlSpec artifacts

Rationale: Embedding execution inside a browser extension creates an isolated execution context that cannot guarantee parity with the canonical execution engine. Capture-then-delegate preserves execution engine authority.

---

### 6 — Surface-Specific Non-Goals

The following surface behaviors are explicitly out of scope.

Desktop:

- is not a packet capture surface
- is not a network traffic analyzer
- provides execution-context diagnostics only (per ADR-014)

Browser Extension:

- is not an execution runtime
- is not a standalone API client
- is not a protocol analyzer

CLI:

- does not provide interactive UX
- does not manage workflow authoring state

CI:

- does not provide interactive debugging
- assumes pre-validated workflow definitions

Container Runtime:

- does not provide interactive UX
- does not install providers dynamically at runtime (provider versions are locked at image build)

---

## Consequences

**Positive:**

- Execution engine remains authoritative across all surfaces
- Surface proliferation cannot fragment semantics
- New surfaces have a defined integration contract
- Parity guarantees are explicitly testable
- UX and engine concerns cannot accumulate inappropriate coupling

**Tradeoffs:**

- Surface-specific optimizations are constrained by parity requirements
- Browser Extension execution delegation adds latency relative to native execution
- Container Runtime provider versioning requires explicit image governance

---

## Alternatives Considered

**Per-surface execution engines:** rejected. Semantic divergence is guaranteed. Core product promise becomes unenforceable.

**Browser Extension with embedded execution:** rejected. Cannot maintain canonical pipeline parity. Extension sandbox limitations prevent deterministic execution.

**Single-surface initial delivery:** rejected. Architecture must support multi-surface from the start, even if delivery is phased. Late surface addition creates retrofit debt.

**Implicit surface parity (no formal model):** rejected. Existing approach. Creates compounding assumptions without declared invariants or testable guarantees.

---

## Amendment — 2026-05-29

### Context

ADR-012 originally defined five delivery surfaces: Desktop, Browser Extension, CLI,
CI, and Container Runtime. No standalone Web App surface was defined. The decision
log in ADR-009 recorded "agent/proxy required for CORS bypass" as a hard requirement
for web execution, implying the web surface could not operate without backend
infrastructure.

This amendment formalises the Web App as a sixth first-class delivery surface and
defines its browser-safe execution model. The core decision: the engine is partitioned
into browser-safe and non-browser-safe capability subsets. The browser-safe subset is
compiled to WebAssembly and bundled with the web app. No backend infrastructure is
required to use browser-safe capabilities.

The Web App and Desktop are mutually exclusive primary surfaces. Desktop targets users
who want full functionality and are willing to install a native app. Web App targets
users who want zero installation and accept the browser-safe capability subset as the
tradeoff. A user running Desktop has no reason to use the Web App.

### Additional Decisions

#### 7 — Web App Surface

The Web App is a sixth first-class delivery surface.

**Definition:**

- static web application (HTML, JS, WASM) deployable to any static host or CDN
- no backend infrastructure required for browser-safe capability execution
- engine compiled to WebAssembly (GOOS=js GOARCH=wasm) and bundled with the app
- storage: IndexedDB (run history, workspace state)
- execution: browser-safe capability subset only (see §8)

**Target user:**

Users who want zero installation. No native app, no CLI binary, no Docker required
to use basic functionality. Open a browser, go to the URL, use it.

**Upgrade paths for extended capability:**

Extended capability (non-browser-safe) requires connecting to an external server.
Two options:

- User-deployed Docker container or local server → web app connects to it
- Hosted server (traCtl-managed or user-managed) → web app connects to it

Desktop app is NOT an upgrade path for the Web App. Users who have Desktop installed
use Desktop directly, not the Web App. These are distinct user personas with distinct
installation preferences.

CLI alone is NOT an upgrade path for the Web App. The CLI is a one-shot binary, not
a persistent server.

The Web App is an interactive surface. It is NOT an automation surface.

#### 8 — Browser-Safe Capability Partition

The engine is formally partitioned into two capability subsets based on browser
runtime constraints.

**Browser-safe capabilities (bundled in Web App WASM, no backend required):**

| Capability | Engine Package | Notes |
|---|---|---|
| Spec validation | internal/validation | Pure logic |
| Overlay parsing and application | internal/overlay | Pure logic |
| YAML / JSON / TOON parsing | internal/parser | Pure logic |
| Execution planning | internal/planner | Pure logic |
| DAG compilation | internal/compiler | Pure logic |
| DAG scheduling | internal/scheduler | Pure logic |
| HTTP execution (protocol.http) | internal/executor | Go WASM target rewires net/http to browser fetch automatically |
| Assertion evaluation | internal/assertion | Pure logic on response |
| Value extraction | internal/extract | Pure logic on response |
| Execution context and expressions | internal/context | Pure logic |
| Basic response timing | internal/diagnostics | Timing available from fetch response |

**Non-browser-safe capabilities (require Docker container or server):**

| Capability | Reason |
|---|---|
| TLS handshake diagnostics (diagnostics.tls) | Requires raw socket access; browser does not expose TLS internals |
| DNS diagnostics (diagnostics.tcp) | Requires OS-level DNS resolution hooks |
| Scripting sandbox (scripting.js) — full | Full Deno/Node.js runtime not available in browser WASM |
| Secrets vault integration (secret.ext.vault@1) | Requires network access to vault infrastructure |
| Non-CORS API execution | Browser enforces CORS; target must send permissive headers |
| Packet capture | Not available in browser sandbox |

**Capability declaration behaviour on web surface:**

When a workflow declares a non-browser-safe capability and no external server is
connected, planning MUST fail with an explicit diagnostic naming the missing
capability and stating that a Docker container or server connection is required.

Silent degradation is prohibited.

#### 9 — CORS Constraint Clarification

CORS is a browser-enforced security boundary, not a traCtl limitation.

The Web App executes HTTP requests using the browser's native fetch API via Go's
net/http WASM rewiring.

- CORS-enabled APIs (Access-Control-Allow-Origin present): work natively, no server
  required.
- Non-CORS APIs: blocked by the browser. Require connecting to a Docker container
  or server which proxies the request outside the browser sandbox.

The Web App surfaces a clear diagnostic when a request is blocked by CORS, with
instructions to connect to a server for extended capability.

#### 10 — Web App Build Artefact

The Web App requires a dedicated build target separate from the CLI binary.

Build target: GOOS=js GOARCH=wasm go build -o tractl.wasm ./cmd/wasm/

A thin cmd/wasm/ entry point is required. It:

- accepts a traCtlSpec document as input from JavaScript
- runs the full browser-safe pipeline (parse → validate → plan → compile →
  schedule → execute → assert)
- returns structured results to JavaScript
- exposes capability negotiation so the UI can declare which capabilities are
  available in the current session (browser-safe only, or extended if a server
  is connected)

The cmd/wasm/ entry point MUST NOT import any non-browser-safe packages.

#### 11 — Amended Surface Table

| Surface | Category | Execution | Backend Required |
|---|---|---|---|
| Desktop | Interactive | Full engine (native Go) | No |
| Web App (basic) | Interactive | Browser-safe engine (WASM) | No |
| Web App (extended) | Interactive | Full engine via server | Docker container or server |
| Browser Extension | Capture only | Delegates to Desktop or CLI | No |
| CLI | Automation | Full engine (native Go) | No |
| CI | Automation | Full engine (native Go) | No |
| Container Runtime | Automation | Full engine (native Go) | No |

---

## Amendment — 2026-05-29

### Context

ADR-012 §1 defines five delivery surfaces: Desktop, Browser Extension, CLI,
CI, Container Runtime. The web browser application is absent as a named
first-class surface. It was implicitly treated as a view of the Desktop
surface (per ADR-016 §7, now corrected). With the ADR-016 correction, the
web application is an independently delivered, standalone surface that
requires explicit surface classification and parity contracts.

### Additional Decisions

#### 8 — Web Application as Sixth Delivery Surface

Add Web Application as the sixth canonical delivery surface.

**Web Application**

Standalone browser-native interactive surface.

- delivered as static files (React + WebAssembly bundle) served from a domain
- no install required — runs in any modern browser
- Go engine compiled to WebAssembly runs inside the browser tab
- no dependency on Desktop application at runtime
- no dependency on any server at runtime (Tier 1)
- optionally connects to a traCtl server for extended capabilities (Tier 2)
- human-in-the-loop interaction model
- full workflow authoring and execution within browser sandbox constraints

Surface category: Interactive (alongside Desktop and Browser Extension).

#### 9 — Web Application Surface Tiers

The Web Application surface operates in two tiers with different capability profiles.

**Tier 1 — Browser-native**

Engine: WebAssembly (WASM) running in-browser
Persistence: IndexedDB
Credentials: browser credential storage
Execution: fetch() for HTTP requests
Limitations: browser sandbox constraints (CORS, no raw socket timing, no filesystem)
Dependencies: none — zero infrastructure required

**Tier 2 — Server-connected (user opt-in)**

Engine: traCtl server (local, Docker, self-hosted, or hosted)
Persistence: filesystem-backed (server-side)
Credentials: OS keychain (if server is local)
Execution: server makes HTTP calls — no CORS restrictions
Limitations: requires a running traCtl server
Dependencies: user-initiated connection — not a default

Tier 1 is the default. Tier 2 is opt-in. The web app never prompts the user
to connect a server on first load.

#### 10 — Revised Surface Category Classification

Update §2 surface category classification to include Web Application.

Interactive surfaces: Desktop, Web Application, Browser Extension

The Web Application shares the interactive surface characteristics defined
in §2 (human-in-the-loop, real-time visual feedback, workflow authoring,
exploratory execution) with the following web-specific additions:

- WASM engine in Tier 1 (not in-process like Desktop, not delegated like
  Browser Extension)
- browser sandbox constraints apply in Tier 1
- server-delegated execution in Tier 2 follows the same parity guarantees
  as CLI/CI surfaces

#### 11 — Web Application Parity Guarantees

The Web Application surface is subject to the §4 execution semantic parity
invariant with the following declared exceptions:

**Tier 1 exceptions (browser sandbox, not engine limitations):**

- requestTimeline.dns, requestTimeline.tcp, requestTimeline.tls fields are
  omitted from execution results. These require raw socket access unavailable
  in the browser sandbox. ttfb and transfer are available via
  PerformanceResourceTiming API.
- CORS policy of target APIs may prevent execution of some requests. This is
  a network-level constraint, not an engine constraint. The engine reports
  the CORS failure as an execution error, not a silent skip.

These exceptions are declared, not silent. They appear in the result record
with an explicit capability note.

**Tier 2: full parity**

When connected to a traCtl server, the Web Application surface has full
execution parity with CLI and CI surfaces. No exceptions.

#### 12 — Desktop and Web Application Independence

The Desktop application and the Web Application are independently delivered
products with no runtime dependency on each other.

They share:

- Go engine source code (internal/*)
- React UI source code (frontend/src/*)
- traCtlSpec canonical model

They do not share:

- binaries
- processes
- ports
- storage
- sessions

A user running the Desktop app and opening the Web Application in a browser
is using two independent products that happen to share source code. There is
no automatic synchronization, shared state, or required pairing between them.

### Consequences

**Positive:**

- Web Application is a zero-friction entry point — no install, no account,
  no server required
- Increases addressable audience beyond desktop-only developers
- WASM execution maintains engine parity guarantees within declared exceptions
- Desktop binary remains lean — no HTTP server bundled

**Tradeoffs:**

- WASM is an additional compilation and maintenance target
- Two persistence models (IndexedDB vs filesystem) mean workspace files are
  not automatically shared between surfaces
- Tier 1 CORS limitations are a real constraint for some users — clearly
  documented in the UI, not hidden
