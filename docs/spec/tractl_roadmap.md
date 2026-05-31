# traCtl Execution Roadmap

Authority: ADR set · traCtlSpec v1.5 · HLD · Handoff Log
Roadmap Reference: Derived from 05_execution_roadmap.md structure; confirmed against handoff log entries 001–008 and session history through Phase 6.4.

---

## How to Read This Document

Each phase represents a coherent pipeline stage or platform capability. Milestone numbers are sequential with no gaps. Status values match the handoff log: **Complete**, **In Progress**, **Not Started**. ADR references are included where a phase is directly governed by a specific decision record. Phases 4 and 5 milestone names are inferred from ADR-003 §3 and §8, as `05_execution_roadmap.md` was not available in the uploaded project files.

---

## Phase 0: Architecture Governance — Complete

The architecture is frozen and governance rules are established before any implementation begins. All ADRs, canonical specifications, the HLD, the terminology registry, and the execution roadmap are authored and accepted in this phase. No implementation starts without this foundation.

- **Milestone 0.1: Architecture Freeze** — Complete
- **Milestone 0.2: ADR-009 Distribution Channel Model** — Complete

---

## Phase 1: Contract Enforcement Layer — Complete

Validates every input asset before anything enters the execution pipeline. This layer enforces format correctness rules, canonical spec semantic rules, and overlay document rules at the earliest possible boundary so that no malformed or semantically invalid input ever reaches the planner or runtime. Five validators are built independently so their import graphs never cross.

- **Milestone 1.1: traCtlSpec Validation** — Complete
  Canonical struct validator enforcing 12 rules: schemaVersion, capabilities, workflow IDs, step IDs, step kind enum, step kind field presence, dependsOn reference resolution, single-workflow DAG acyclicity, assertion IDs, extract IDs, and failurePolicy. 53 tests. Post-audit compliance fixes applied.

- **Milestone 1.2: Overlay Validation** — Complete
  Overlay document validator enforcing overlay spec §15.1 rules. 11 rules across 2 rule files. Zero dependency on `internal/spec` — preserved as an architectural constraint.

- **Milestone 1.3: YAML Format Validation** — Complete
  YAML syntax and dialect rules. 8 rule sets. YAML 1.1 boolean edge-case (yes/no/on/off under gopkg.in/yaml.v3 YAML 1.2 core schema) fixed. 60 tests.

- **Milestone 1.4: JSON Format Validation** — Complete
  RFC 8259 strict validation. JSON5/JSONC detection (comments, trailing commas, single-quoted strings). `internal/validation/common` shared package introduced to prevent semantic drift across format validators. 37 tests.

- **Milestone 1.5: TOON Format Validation** — Complete
  TOON grammar rules. yes/no/on/off are bare strings in TOON (not boolean errors). Document separators (---) rejected. Stricter escape subset than YAML enforced at lexical pre-check. 38 tests.

---

## Phase 2: Parser Layer — Complete

Converts each supported authoring format (YAML, JSON, TOON) into the canonical in-memory `*spec.TraCtlSpec` model. ULIDs are assigned to every identifiable entity, metadata fields are enriched, and user-authored array order is preserved. Parsers do not call the canonical validator — that is the caller's responsibility. (ADR-002)

- **Milestone 2.1: YAML Parser** — Complete
  Three-phase pipeline: format validate → unmarshal → assign ULIDs → set metadata. 10 tests.

- **Milestone 2.2: JSON Parser** — Complete
  Mirrors YAML parser structure; JSON format validator runs first. 11 tests.

- **Milestone 2.3: TOON Parser** — Complete
  TOON validator + UnmarshalYAMLBytes (TOON is a strict YAML subset; struct tags are identical). 9 tests.

---

## Phase 3: Overlay Engine — Complete

Applies overlay documents to the canonical spec using path, match, and source targeting modes. Supports all merge semantics defined in the Overlay Specification: scalar replace, object deepMerge, array append, array replace, and appendUnique. Records provenance in metadata after every patch. The engine operates on a JSON-clone (`map[string]any`) so it never imports `internal/spec`. (Overlay Specification)

- **Milestone 3.1: Overlay Engine** — Complete
  8 files: errors, clone, path, merge, traverse, provenance, parse, engine. 44 tests covering all Overlay Specification §20 conformance requirements.

---

## Phase 4: Execution Planner and DAG Compiler — Complete

Transforms the validated, post-overlay traCtlSpec into a runtime-executable plan. The planner and compiler are explicitly separate concerns per ADR-003 §3. The planner resolves capability contracts, activates environments, analyses dependencies, and determines execution strategy. The compiler converts planner output into the deterministic runtime artifact — the executable DAG. This phase owns composite cycle detection via recursive DAG flattening (traCtlSpec §16.3). (ADR-003, ADR-011)

- **Milestone 4.1: Execution Planner** — Complete
  Capability contract resolution, environment activation, dependency analysis, concurrency planning, and planning trace event emission (ADR-014 §5 hook). Unknown capability contracts and incompatible versions fail planning.

- **Milestone 4.2: DAG Compiler** — Complete
  Topological compilation, parallel eligibility detection, implicit dependency merging from variable references (traCtlSpec §12.2), and composite step recursive flattening for cycle detection.

---

## Phase 5: Runtime Layer — Complete

Executes planned HTTP steps against real services. Every request produces a Request Timeline capturing the seven mandatory timing segments defined in ADR-014 §2 (DNS resolution, TCP connect, TLS handshake, request sent, TTFB, response transfer, total). Retry and redirect semantics are fully enforced at this layer. (ADR-003, ADR-008, ADR-014)

- **Milestone 5.1: HTTP Request Dispatch** — Complete
  Core HTTP client, request builder, response model, and Request Timeline start/end capture.

- **Milestone 5.2: Retry and Timeout** — Complete
  Retry policies (fixed, linear, exponential backoff), retryOn conditions, timeout propagation via context cancellation. Retry evaluated against remaining timeout budget.

- **Milestone 5.3: Redirect Handling** — Complete
  Redirect chain tracking, redirect hop event emission (ADR-014 §3), and configurable redirect policy.

---

## Phase 6: Assertions, Extracts, and CLI Harness — Complete

Evaluates assertion rules against step results, extracts values into the execution context, and wires all pipeline phases into a runnable `tractl` binary for internal developer testing. After this phase, `go build ./cmd/tractl/` produces a working binary and `tractl run workflow.yaml` executes a real workflow end-to-end. (traCtlSpec §11, §13, ADR-014)

- **Milestone 6.1: Execution Context Engine** — Complete
  Five variable scopes (spec, environment, workflow, step, runtime) with most-specific-wins resolution. `StepResult` and `TimingData` types. Expression evaluator for `${…}` syntax. `FrozenContext` immutable snapshot for scripts. Concurrent-safe via `sync.RWMutex`.

- **Milestone 6.2: Assertion Evaluator** — Complete
  Six assertion kinds: status, header, body, schema (stub), script (stub), extension (stub). Seven operators: equals, matches, contains, exists, inRange, jsonpath, custom. Severity enforcement: error → step fails; warning → step succeeds with diagnostic. Credential-safe header masking in observability events. (ADR-014 §5)

- **Milestone 6.3: Extract Engine** — Complete
  Six extract sources: status, header, body (gjson path), metadata, timing, extension (stub). Scoped context writes (step, workflow, spec). Value placeholder in observability events — actual extracted value never appears in diagnostics. (ADR-014 §6)

- **Milestone 6.4: CLI Harness and Pipeline Wiring** — Complete
  `internal/engine/` coordinator wiring all six pipeline stages in canonical ADR-003 order. `RunResult`, `WorkflowOutcome`, and `StepOutcome` types. `FormatText` and `FormatJSON` output formatters. `cmd/tractl/main.go` with `tractl run` subcommand, `--env`, `--overlay` (repeatable), `--output` (text|json), and `--verbose` flags. Flags accepted before or after the workflow file. Exit codes: 0 = pass, 1 = assertion failures, 2 = pipeline error. `--verbose` includes HTTP response status, headers, and body in output; credential headers are masked with typed placeholders before output (ADR-014 §6). Quiet mode (`--output json`) suppresses engine log lines so JSON is clean on stdout.

---

## Phase 7: Diagnostics and Reporting — In Progress

Assembles all observability events collected across the pipeline into structured, credential-safe output. Implements the full output contract across all four formats and all five delivery surfaces. Includes the Dependency Waterfall and Execution Provenance Trace that make execution results auditable. Also implements opt-in per-step diagnostics directives for TLS, TCP, and transport-level observability. (ADR-014, ADR-002 §4, traCtlSpec §15)

- **Milestone 7.1: Observability Event Assembly and Execution Result Model** — Complete
  Collects planning trace, Request Timeline, retry events, redirect events, assertion evaluation events, and extract evaluation events into a single coherent `ExecutionRecord`. `Collector` is fully wired into `engine.Run()`; goroutine-safe via `sync.Mutex`. Credential-safe at every output surface (ADR-014 §6).

- **Milestone 7.2: Output Formatters** — In Progress
  Text (CLI) — Complete: `diagnostics.FormatRecordText` renders header block, per-workflow step timing table, waterfall, and provenance trace; integrated into `engine.FormatText` under `--verbose`. JSON — Complete: `FormatRecordJSON` (`json.MarshalIndent`). YAML — Complete: `FormatRecordYAML` (`gopkg.in/yaml.v3`). **TOON — Not implemented** (`format.go` defers to YAML as structurally identical; completing this requires a TOON serialiser). (ADR-002 §4)

- **Milestone 7.3: Dependency Waterfall** — Complete
  `BuildWaterfall` and `WaterfallText` in `diagnostics/waterfall.go`. Entries ordered by `StartOffsetMs`; bar-chart ASCII rendering with configurable width. Integrated into `FormatRecordText` for workflows with ≥ 2 steps. Serialisable to JSON/YAML for structured output. (ADR-014 §4)

- **Milestone 7.4: Execution Provenance Trace** — Complete
  `TraceEvent` emitted at all pipeline stages (validation, planning, compilation, workflow start/end). Surfaced as `Provenance:` section in `FormatRecordText`, monotonic-ms ordered. (ADR-014 §5)

- **Milestone 7.5: Diagnostics Directives** — Stub / Not Complete
  `directives.go` defines types (`StepDiagnosticsOutput`, `DiagnosticsEvent`) and all six kind/retention constants (`KindTLS`, `KindTCP`, `KindTransport`, `KindLifecycle`, `KindDNS`, `KindConnection`). Validation rules for spec-level `diagnostics` blocks exist in `internal/validation/rules_diagnostics.go`. Runtime population of `StepDiagnosticsOutput.Events` is not implemented — HTTP client instrumentation for TLS, TCP, DNS, and transport capture is deferred. Diagnostics MUST NOT change execution semantics. (traCtlSpec §15)

---

## Phase 8: Parallel DAG Execution — Complete

Upgrades the sequential executor introduced in Phase 6.4 to true concurrent structured execution. Steps without dependency relationships execute in parallel, bounded by the workflow's declared concurrency setting. All concurrency is structured: every spawned execution unit belongs to a parent lifecycle, orphan work is prohibited, and cancellation propagates deterministically downward. Multi-workflow parallel execution with cross-workflow `dependsOn` added as a workflow-level DAG above the step-level DAG. (ADR-008)

- **Milestone 8.1: Structured Concurrency Scheduler** — Complete
  `internal/scheduler/` goroutine-based DAG execution with bounded concurrency (`concurrencyLimit`, default 4). Every task lifetime is bounded by its parent via `errgroup` + semaphore `chan struct{}`. No orphan work. Topological dispatch (Kahn's algorithm) based on the compiled DAG from Phase 4.2. `RunWithContext` method added for engine integration.

- **Milestone 8.2: Cancellation Propagation and Timeout Composition** — Complete
  Deterministic cancellation propagating downward through the structured concurrency tree via per-step `context.CancelFunc` stored in the DAG node. `failFast` cancels only the transitive dependent subgraph of the failed step; independent branches are unaffected.

- **Milestone 8.3: failFast and Resilient Policy Enforcement** — Complete
  Resilient (default): failed step's dependents are dependency-skipped at the next scheduling cycle; independent branches always continue. failFast (opt-in): active cancellation signal sent to the transitive dependency subgraph of the failed step only. Both policies validated with race-detector tests in `internal/engine/engine_test.go` and `internal/scheduler/`.

- **Milestone 8.4: Multi-Workflow Parallel DAG** — Complete
  Workflows execute as a DAG: cross-workflow `dependsOn` in `spec.Workflow` and `planner.WorkflowPlan`. `workflowConcurrency` (default 10) caps parallel workflow execution. `runWorkflowDAG` in `internal/engine/` uses `errgroup` + semaphore, mirrors the step-level scheduler pattern. Workflow-level topo-sort (`topoSortWorkflows`) with cycle detection in `internal/planner/topo.go`. Unknown and self-referencing `dependsOn` IDs rejected at validation. Skipped workflows (dependency failed) reported in `WorkflowOutcome.Skipped`.

---

## Phase 9: Script Sandbox — Complete

Implements workflow/step lifecycle hooks backed by a bounded JavaScript sandbox. Scripts run inside a fresh `goja.Runtime` per execution, receive a `FrozenContext` snapshot (read-only deep copy), and return a `MutationSet` that the engine applies transactionally after `Execute` returns. This preserves determinism (ADR-003) and prohibits direct I/O (ADR-006). JavaScript (`js`) is the only canonical scripting language in schema v1. (traCtlSpec §9, ADR-003, ADR-006)

- **Milestone 9.1: Script Sandbox Engine and Frozen Context Delivery** — Complete
  `internal/sandbox/` package: `Sandbox.Execute` (per-call fresh `goja.Runtime`, IIFE wrapping, timeout via `time.AfterFunc` + `vm.Interrupt`). I/O prohibition enforced by omission — no `console`, `fetch`, `require`, `process`, `fs`, `setTimeout`, or `setInterval` injected. `injectSafeAPIs` exposes `tractl.now()` (monotonic counter seeded from trace ID), `tractl.random()` (per-call seeded FNV-1a `math/rand`), and `tractl.log()`. `injectFrozenContext` delivers `ctx.spec`, `ctx.workflow`, `ctx.step`, `ctx.environment`, `ctx.runtime` as read-only goja objects. `FrozenContext` is a deep-copy snapshot taken inside `runHook` via `execCtx.Freeze()` (holds `sync.RWMutex` read lock). `classifyError` uses type-assertion on `*goja.InterruptedError` (timeout) and `*goja.Exception` with `SyntaxError` prefix check (compile-time and user-thrown syntax errors both surface as `*goja.Exception`). `traceIDToSeed` (FNV-1a) lives in `internal/engine`; script semaphore `chan struct{}` of capacity `MaxConcurrentScripts` (defaults to `runtime.NumCPU()`) caps concurrent `goja.Runtime` instances. Error codes: `ErrSyntax`, `ErrRuntime`, `ErrTimeout`, `ErrInvalidResult`, `ErrSourceRefUnsupported`, `ErrLanguageUnsupported`.

- **Milestone 9.2: Mutation Set Application and Lifecycle Hooks** — Complete
  `MutationSet` struct with `Variables`, `Extracts`, `Assertions`, `Logs`, `Cancel`. `Apply` writes variables via `execCtx.SetVar`, extracts via `execCtx.SetStepExtract`, returns `ErrCancellationRequested` on `Cancel: true`; `Logs` and `Assertions` no-op (wired to observability in a future phase). All seven hook locations dispatched by `executeWorkflow` in declaration order: `beforeAll` (workflow start; error → fail), `beforeEach` (loop over steps pre-scheduler; error → fail), `sched.RunWithContext`, step-loop for `beforeStep`/`afterStep`/`transform` (post-execution approximation; errors logged, not fatal), `afterEach` (loop over outcomes; errors logged), `afterAll` (workflow end; error → fail). `SourceRef` rejected (Phase 9 inline-only). `runHook` acquires script semaphore slot with `select`/`ctx.Done()` guard before calling `sb.Execute`. 14 example files (`examples/yaml/06–12`, `examples/toon/06–12`) covering all hook locations, MutationSet fields, cancel signal, deterministic APIs, and full dispatch order. Import boundary: `internal/sandbox` imports only `internal/context`; `internal/engine` imports `internal/sandbox`.

---

## Phase 10: Composite Steps — Not Started

Implements the composite step kind, enabling one workflow to inline another as a sub-workflow. The planner performs recursive DAG flattening before execution begins so that composite cycles are rejected at planning time, not at runtime. Variable scope isolation between parent and sub-workflow is configurable (isolated vs shared). (traCtlSpec §16, ADR-003)

- **Milestone 10.1: Composite Step Execution and Variable Scope Isolation** — Not Started
  Composite step dispatch, input/output variable bindings across the workflow boundary, isolated vs shared scope semantics.

- **Milestone 10.2: Planner Recursive DAG Flattening and Composite Cycle Detection** — Not Started
  Depth-first recursive flattening of all composite step references into the parent DAG. Cycle detection on the fully flattened graph. Rejection diagnostic identifies the full reference path forming the cycle (e.g., `workflowA → composite:workflowB → composite:workflowA`). (traCtlSpec §16.3)

---

## Phase 11: Extension Platform — Not Started

Implements the full extension and provider architecture. Extensions are capability providers — not runtime peers — and are untrusted by default. Every extension provides a manifest (governance contract) and an SDK contract (execution contract). Capability contracts from native extensions, MCP adapters, and provider adapters all normalize into the same model consumed by the planner. (ADR-004, ADR-005, ADR-007, ADR-011)

- **Milestone 11.1: Capability Contract Registry and Compatibility Validation** — Not Started
  Unified capability contract model. Version compatibility validation (presence, semantic, input/output, execution constraint). Conflict detection at registration time. Planning-time resolution with explicit diagnostics identifying failing compatibility dimension. (ADR-011)

- **Milestone 11.2: Extension Manifest Registry and Lifecycle** — Not Started
  Manifest ingestion: identifier, version, capability declarations, permission requirements, compatibility metadata. Lifecycle stages: install → register → validate → compatibility check → resolve → authorize → execute → observe → retire. (ADR-004 §2, §6)

- **Milestone 11.3: Extension Invocation Surface** — Not Started
  `extensionCall` step kind dispatch through the Extension Runtime. Extension assertion kind delegation (assertion.kind: extension). Output bindings back into execution context. Extensions receive only declared inputs, return only declared outputs. (traCtlSpec §14.1, §14.2)

- **Milestone 11.4: Auth, Secret, and Data Provider Extensions** — Not Started
  extensionKind auth, secret, and data provider dispatch. Secret access policy: secrets are accessible only through scoped injection at execution time, never directly exposed to workflow definitions or arbitrary extensions. (traCtlSpec §14.3, ADR-006 §3)

- **Milestone 11.5: MCP Adapter and Capability Normalization** — Not Started
  MCP integration as an adapter layer, not as an execution substrate. MCP-exposed capabilities normalize into the canonical capability contract model. MCP failure during planning fails planning explicitly with a timeout-bounded diagnostic. MCP runtime failures propagate through DAG scheduler semantics. (ADR-005)

---

## Phase 12: OpenAPI Compatibility Provider — Not Started

Delivers the first compatibility provider, enabling `tractl run --openapi openapi.yaml` without forced migration. The OpenAPI adapter converts an OpenAPI document into canonical traCtlSpec via the provider adapter model established in Phase 11. Fidelity disclosure is mandatory: traCtl declares which OpenAPI semantics it honours and which it cannot represent. Overlay augmentation allows adding assertions, auth scenarios, and variable extraction on top of OpenAPI sources. (ADR-007, 03_alpha_version_onboarding §8)

- **Milestone 12.1: OpenAPI Parser and Provider Adapter** — Not Started
  OpenAPI document ingestion (YAML and JSON). Canonical traCtlSpec generation from OpenAPI paths, operations, parameters, and request/response schemas. Provenance metadata recording OpenAPI source reference.

- **Milestone 12.2: Fidelity Disclosure and Compatibility Analysis** — Not Started
  Explicit declaration of which OpenAPI semantics are supported and which fall outside traCtlSpec's scope. Surface-visible fidelity report. No silent compatibility degradation permitted.

- **Milestone 12.3: Overlay Augmentation for OpenAPI Sources** — Not Started
  Overlay application on OpenAPI-sourced canonical traCtlSpec: assertion injection, auth enrichment, variable extraction, and validation stage configuration. Enables progressive validation adoption without requiring migration to native TOON/YAML/JSON workflow authoring.

---

## Phase 13: Fuzz Directives — Not Started

Implements the fuzz directive execution model declared in traCtlSpec §10. The builtin strategy covers boundary mutations, invalid payloads, required-field omission, and malformed requests. The advanced strategy and external providers (Schemathesis) require the Extension Platform from Phase 11. Fuzz runs with an explicit seed MUST be reproducible. (traCtlSpec §10)

- **Milestone 13.1: Builtin Fuzz Strategy Engine** — Not Started
  Boundary mutations, invalid type payloads, required-field omission, malformed encoding. Seed-deterministic case generation. Budget enforcement (cases | duration). Invariant assertions evaluated per fuzz case (traCtlSpec §10.4).

- **Milestone 13.2: Schemathesis Provider Adapter** — Not Started
  ext:schemathesis delegated fuzz provider. Schema-driven case generation via the Extension Platform (requires Phase 11). Capability declaration: `fuzz.ext.schemathesis@1`.

---

## Phase 14: Acquisition Layer — Not Started

Implements the pre-normalization ingestion boundary for capture-assisted workflow bootstrapping. All capture adapters emit Capture Records, which are normalized into traCtlSpec by the Capture Normalization pipeline. Credential masking is mandatory at the adapter boundary — before Capture Records exit the adapter — because capture output is intended to become a Git-native artifact. (ADR-013)

- **Milestone 14.1: Capture Record Model and Normalization Pipeline** — Not Started
  Capture Record structure: acquisition source, timestamp, raw request/response, provenance, trust class. Normalization responsibilities: URL parameterization, header canonicalization, body schema inference, request/response pairing, workflow step generation, dependency inference. Normalization is deterministic. (ADR-013 §3, §4)

- **Milestone 14.2: HAR Import Adapter** — Not Started
  Static HAR file ingestion. Trust class: imported artifact (untrusted by default). Schema validation before normalization.

- **Milestone 14.3: curl Import Adapter** — Not Started
  curl command string parsing into structured request representation. Single and multi-command support. Trust class: imported artifact.

- **Milestone 14.4: Credential Masking at Adapter Boundary** — Not Started
  Detection and masking of Authorization headers, API key headers, cookie values, query parameter credentials, and body credential fields before Capture Records enter normalization. Placeholder format preserves credential type without exposing value. Masking is not optional. (ADR-013 §8)

- **Milestone 14.5: Capture Provenance Tracking** — Not Started
  Provenance fields on every Acquisition Layer output: source type, timestamp, adapter version, normalization rule version, original source reference, and provenance chain. Provenance preserved through overlay application and into execution diagnostics. (ADR-013 §5)

- **Milestone 14.6: Browser Interception and Local Proxy Adapters** — Not Started
  Browser Interception Adapter (Browser Extension → CDP / WebExtensions API). Local Proxy Adapter (HTTP CONNECT tunneling with local CA). Trust class: user-initiated / user-configured. Deferred to v1.5 per HLD §9.

---

## Phase 15: Multi-Surface Delivery and Distribution — Not Started

Delivers traCtl across all five canonical delivery surfaces and through the defined distribution channels. The core execution invariant — same workflow definition, same execution semantics, regardless of surface — is the non-negotiable constraint governing this entire phase. (ADR-012, ADR-015, ADR-009)

- **Milestone 15.1: CLI Production Hardening** — Not Started
  Full subcommand surface (`run`, `validate`, `plan`, `version`), flag completions, structured error output, version skew detection. Exit code contract: 0 = pass, 1 = assertion failures, 2 = pipeline error, 3 = capability error.

- **Milestone 15.2: CLI Distribution** — Not Started
  Homebrew formula (macOS and Linux), curl installer with checksum verification and explicit version pinning, goreleaser release pipeline. Scoop and WinGet for Windows. (ADR-015 §2)

- **Milestone 15.3: CI Surface** — Not Started
  CLI-driven deterministic automation execution. Pinned version in CI configuration. Structured machine-consumable output (JSON, YAML, TOON). No interactive prompts. No runtime provider installation. (ADR-012, ADR-009 §7)

- **Milestone 15.4: Container Runtime** — Not Started
  OCI-compliant container images. Multi-arch: linux/amd64 and linux/arm64. Image variants: latest, pinned version, CI runner image, slim (distroless). cosign signing and SBOM attestation. Providers locked at image build time — no runtime provider installation. (ADR-015 §3, ADR-009 §6)

- **Milestone 15.5: Desktop Application** — Not Started
  Primary interactive developer UX surface. macOS (Homebrew Cask, signed DMG, Sparkle auto-update, Apple notarization), Windows (MSIX, Authenticode signed, silent install), Linux (Homebrew/Linuxbrew, .deb, .rpm, AppImage). Full observability visualization. Workflow authoring. Git integration. Inspection mode for observability data. (ADR-015 §1, ADR-012)

- **Milestone 15.6: Browser Extension** — Not Started
  Capture-only surface. MUST NOT execute workflows. Delegates execution to Desktop or CLI via Browser Bridge (WebExtensions native messaging API). Distributed through Chrome Web Store, Firefox Add-ons, and Edge Add-ons. Enterprise-managed deployment via Group Policy supported. Explicitly deferred to v1.5 per HLD §9. (ADR-012 §5, ADR-015 §4)

---

## UI Track: Web Surface Foundation — In Progress

Delivers the shared React frontend and Web Tier 1 WASM runtime path defined by
ADR-016 and the ADR-012 web application amendments. This track keeps screens
transport-agnostic while routing execution through platform adapters.

- **Milestone UI-0.1F: Browser-Native Engine Execution** — Complete
  `window.tractl.run(document, format)` Promise API. Browser WASM executes the
  existing Go engine pipeline using browser `fetch` for HTTP requests. Verified
  parser, validation, execution, assertions, extracts, and browser network
  behavior from Playwright. Request Editor integration was deferred to UI-0.2A.

- **Milestone UI-0.2A: Request Editor → WASM Runtime Integration** — Complete
  Request Editor Run now serializes structured draft state with
  `requestDraftToTraCtlSpec`, routes Web Tier 1 execution through
  `window.tractl.run(document, "json")`, maps canonical WASM `RunResult` into
  the existing `RequestRunResult` model, and renders through the unchanged
  Results Panel. A `RequestExecutionRunner` factory keeps `useRequestEditor`
  transport-agnostic; desktop remains on the existing local API save +
  run-by-fileId path. Web persistence and Tier 2 HTTP/proxy runner remain
  deferred.

- **Milestone UI-0.2B: Fast Start Completion (S1)** — Complete
  Fast Start action cards use real navigation: new request → Request Editor;
  new workflow / OpenAPI import → workflow placeholder (no fake loading);
  open file → picker → WASM parse/validate → execute via
  `RequestExecutionRunner`. Recent list merges persisted run history and
  OpenAPI imports (newest first). Recent Open/Run and View all use S9 run
  history. Empty recent state on Fast Start when no data.

- **Milestone UI-0.2B-web: Web Persistence (IndexedDB)** — Not Started
  Implement `RequestExecutionRunner.saveRequest` for the web surface using
  IndexedDB. Enables autosave and named request files on web without a server.
  Tier 2 HTTP/proxy runner remains a separate future path.

- **Milestone UI-0.2C: Run History Integration** — Complete
  Execution history persisted via Zustand `persist` (localStorage, key
  `tractl-run-history`, cap 200). `runHistoryStore` captures every
  `runRequest` completion — success and failure — with method, URL, status,
  duration, outcome, and full result snapshot. Run History screen (S9)
  replaces placeholder: chronological list newest-first with method badge,
  outcome badge, duration, and timestamp. Selecting any entry navigates to
  the Request Editor and pre-loads the historical result through the unchanged
  `RequestResultsPanel`. No secrets stored (ADR-006). Web and Desktop both
  use localStorage persistence; no server dependency.

- **Milestone UI-0.2D: Status Bar Execution State** — Not Started
- **Milestone UI-0.2D: Environment Selection & Variable Resolution** — Complete
  Added persisted environment management for flat string variables, wired the
  titlebar active-environment badge to quick switching and creation, replaced
  sidebar environment mocks with `EnvironmentStore`, and resolves `{{name}}`
  placeholders in request URL, query params, headers, and body before the
  request runner receives the canonical document. Missing variables fail early
  with `TRACTL_ENV_VARIABLE_MISSING`; no silent substitution or empty-string
  replacement occurs.

- **Milestone UI-0.2E: Status Bar Execution State** — Not Started
  Wire the status bar to reflect Idle → Running → Success / Failed during
  and after request execution. Currently the status bar does not respond to
  execution lifecycle events.

---

## Summary Table

| Phase | Title | Status |
|---|---|---|
| 0 | Architecture Governance | Complete |
| 1 | Contract Enforcement Layer | Complete |
| 2 | Parser Layer | Complete |
| 3 | Overlay Engine | Complete |
| 4 | Execution Planner and DAG Compiler | Complete |
| 5 | Runtime Layer | Complete |
| 6 | Assertions, Extracts, and CLI Harness | Complete |
| 7 | Diagnostics and Reporting | In Progress |
| 8 | Parallel DAG Execution | Complete |
| 9 | Script Sandbox | Complete |
| 10 | Composite Steps | Not Started |
| 11 | Extension Platform | Not Started |
| 12 | OpenAPI Compatibility Provider | Not Started |
| 13 | Fuzz Directives | Not Started |
| 14 | Acquisition Layer | Not Started |
| 15 | Multi-Surface Delivery and Distribution | Not Started |
| UI | Web Surface Foundation | In Progress |
