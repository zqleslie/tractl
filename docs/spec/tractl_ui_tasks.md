# traCtl UI Task List

Status: Active
Authority: ADR-016 · tractl_ui_spec.md · tractl_ui_context.md
Last updated: 2026-06-03

Tags: [EXISTING] = was already in spec/backlog · [NEW] = added from capability session
Priority: P1 = blocks alpha · P2 = alpha quality · P3 = post-alpha

---

## Active

**T-01 · Workflow Persistence + Workspace Integration** [EXISTING] P1
Step edits in the Workflow Canvas Step Detail Panel do not save back to YAML. `useStepDraftEditor` has explicit TODO stubs throughout. Wires draft state → YAML serialisation → `POST /api/v1/files` (web) / Wails binding (desktop). Also replaces sidebar workflow/request lists which still use alpha fixture data with real workspace file reads.

**T-02 · Desktop File Load API** [EXISTING] P1
`GET /api/v1/files/:id` is not implemented on the desktop local API (noted as TODO in UI-0.3E). Workflow Canvas on desktop cannot load an existing workflow file from disk — it falls back to in-memory workspace and logs a warning. This is a Go-side `cmd/server` endpoint addition that unblocks desktop workflow authoring end-to-end.

---

## Phase 0.4 — Workspace Features

**T-03 · Environment Manager (S6)** [EXISTING] P1
The env switcher and flat variable store exist but the full manager screen is not built. Requires model alignment before implementation begins — flagged in screen inventory. Covers: named environments, variable groups, secret references (`${secrets.env.name}`), and environment-scoped overrides without exposing raw credential values.

**T-04 · Monaco Wiring + ctx Types** [EXISTING] P1
Monaco Editor is stubbed as a styled `<textarea>` in all pre/post script panels. Replace stub with a real Monaco instance. Inject `ctx` TypeScript type definitions generated from the Go `FrozenContext` struct — autocomplete reflects the exact runtime shape per hook location (`ctx.step.response` is only typed in post-script, not pre-script).

**T-05 · Script Snippet Library** [NEW] P1
A chip row above the Monaco editor with insertable snippets: Extract token, Assert status, Cancel on 5xx, Log response, Set from header. Each chip inserts a working, typed code block at cursor that validates against the injected `ctx` types immediately. Eliminates the blank-page problem for first-time script authors and reduces migration friction from pm-based Postman/Apidog scripts.

**T-06 · Script Return Shape Validator** [NEW] P2
A live status bar below the Monaco editor showing whether the script's return shape is valid before Run. Computed via lightweight AST parse of the return statement — no sandbox execution needed. Shows `variables: N keys · assertions: N · cancel: false` or a warning if the return shape doesn't match the mutation set interface for the current hook location (pre vs post have different allowed fields).

**T-07 · Script Editor Screen (S8)** [EXISTING] P2
The dedicated full-screen script editor is not started. The only screen where Monaco is the primary surface. Builds on T-04 and T-05. Adds the ctx Reference explorer panel (collapsible right sidebar with full `ctx` tree, field types, inline docs) and the return shape validator from T-06. Monaco is the only third-party UI component permitted per ADR-016.

**T-08 · Settings Screen (S10)** [EXISTING] P2
App preferences screen: theme (light/dark/system), density (comfortable/default/compact), default environment, default failure policy, WASM memory, desktop API port, surface/version info. Most state exists in `uiStore` — this screen exposes it. Low complexity. Unblocks T-36 (Credential Manager).

**T-09 · S4 Step Detail — Full Completion** [EXISTING] P2
The Step Detail Panel shell (7 tabs) exists but step persistence and result tab polish are explicitly TODO in `useStepDraftEditor`. Distinct from T-01. Wires individual step form state → `useStepDraftEditor` save → workspace store sync. Result tab shows response body, assertions with pass/fail, extracts with resolved values from last run.

**T-10 · Overlay Editor Screen (S7)** [EXISTING] P2
Entirely undesigned beyond a placeholder. Implements a visual patch builder: form-based list of patches (target mode selector, action selector, field + value inputs) generating valid overlay YAML without hand-writing. Patch order is drag-reorderable. Each patch shows a matched node count badge updated live as the WASM overlay engine evaluates against the current workspace spec.

**T-11 · Overlay Before/After Diff Preview** [NEW] P2
A split panel alongside the T-10 patch list: left shows raw source spec (or OpenAPI import), right shows post-overlay canonical result, live-updating as patches are edited. Answers "what does this overlay do?" without a full execution run. Diff data from running the already-wired WASM overlay engine against workspace spec + authored patches.

**T-12 · Overlay Semantic Match Visualiser** [NEW] P2
For `mode: match` and `mode: all` patches in the T-10 Overlay Editor, a visual preview panel highlights which Workflow Canvas steps would be matched by the selector. A `match: { protocol: http }` patch with `mode: all` shades all HTTP step nodes. Makes `mode: all` patches auditable before execution — currently no way to know which nodes a semantic selector targets without running the engine.

---

## Phase 0.4 — Form Panel Enhancements

**T-13 · JSONPath Live Tester — Assertions Tab** [NEW] P1
In the Assertions tab (S2 and S4), when kind is `body` and operator is `jsonpath`, an inline tester panel appears. User types a JSONPath expression, clicks Test, and sees the extracted value from the last response body alongside a pass/fail indicator. Response body already in result store — no extra request needed. The most common pain point in every API testing tool's assertion editor.

**T-14 · Click-to-Extract from Response Body** [NEW] P1
In the Extracts tab, an "Extract from response" button opens a read-only tree of the last response body. Clicking any value node auto-fills an extract row: source=body, path=generated JSONPath, variable name=suggested from key name. Eliminates the most error-prone manual step in API testing — writing JSONPath by hand against a body structure held in memory.

**T-15 · Mutation Set Result Panel** [NEW] P2
After a script executes, the script tab shows what the mutation set actually produced below the editor. Displays each mutation key with its resolved value: `variables.userId → "u_7f3a9b1c" (workflow scope)`, `assertions[0] → status 201 ✓ passed`, `cancel → false`. Apidog shows raw console logs. traCtl's mutation set output is semantically structured and maps directly to the spec contract.

**T-16 · Variable Scope Explorer** [NEW] P2
A panel accessible from any `${...}` field showing the full variable tree across all four scopes: step → workflow → environment → spec. Each scope is collapsible and shows resolved values from the last run with resolution order annotated (most-specific-wins). Includes an expression resolver: paste any `${...}` expression, click Resolve, see which scope it resolved from and the actual value.

**T-17 · Inline Expression Preview** [NEW] P2
In any field containing `${...}` expressions (URL bar, headers, body), a hover tooltip shows the resolved value from the last run. `Bearer ${vars.token}` → hover shows `Bearer eyJhbGci...`. Zero extra UI chrome — tooltip on expression spans. Distinct from T-16: this is the lightweight inline version for fields where a full panel would be disruptive.

**T-18 · Unresolved Variable Warning** [NEW] P2
Any `${...}` expression that cannot be resolved in any active scope gets an amber underline at authoring time — before Run. Tooltip explains which scope was searched and what is missing. Surfaces resolution errors early and eliminates the "run → fail → fix → re-run" cycle that is the most common beginner friction point.

**T-19 · Failure Policy Selector with Live Preview** [NEW] P2
In the workflow Settings tab (S3 and S4), a policy selector (failFast / resilient) with a live mini-diagram that updates the canvas preview when switched: "if this step fails, N steps cancel / N steps continue." Computed statically from `dependsOn` edges — no run required. Currently failure policy exists in the spec but has no visual representation in the UI.

---

## Phase 0.5 — Run Results + Observability

**T-20 · Workflow Run Results Screen (S5)** [EXISTING] P1
S5 is not started despite the engine producing full waterfall, provenance trace, and timing data since Phase 7. Builds the interactive dependency waterfall (clickable horizontal bars per step, expandable to per-segment timing), pipeline stage mini-timeline (validation → planning → compilation → execution with ms labels), and per-step outcome badges. Engine data is fully serialised — pure rendering task.

**T-21 · Wait-Time Annotation in Waterfall** [NEW] P2
Each waterfall bar gets a lighter leading segment representing time the step spent waiting for its dependencies. Makes DAG bottlenecks visually obvious — teams see which steps are blocked waiting versus actively executing. Data from `WaterfallEntry.StartOffsetMs` already computed in `internal/diagnostics/waterfall.go`.

**T-22 · Critical Path Highlighting** [NEW] P2
After a run, compute the critical path (longest dependency chain) and highlight it in both waterfall and canvas. Critical path edges render in blue; steps on the path have a distinct border. Swim lane brackets annotate truly parallel groups. Toggle in the waterfall toolbar. Makes the performance bottleneck immediately legible without timing arithmetic.

**T-23 · Parallelism Swim Lane Annotation** [NEW] P2
In the canvas and waterfall, steps executing in parallel are visually grouped with a swim lane bracket. Shows the group's combined wall-clock time vs what sequential execution would have taken. A "saved Xms by parallelism" annotation makes the DAG execution advantage concrete. Currently parallelism is implicit — users must infer it from `dependsOn` declarations.

**T-24 · Outcome-Filtered Waterfall** [NEW] P2
A filter toggle bar on the S5 waterfall: All · Failed · Slow (above configurable threshold). Reduces visual noise in large workflows where most steps pass. Filter state persists per run history entry. Threshold defaults to 500ms, configurable per workflow. All filter logic operates on already-rendered waterfall data — no engine re-execution.

**T-25 · Pipeline Stage Timeline** [NEW] P2
A compact horizontal timeline above the S5 waterfall showing the four canonical pipeline stages (Validate, Plan, Compile, Execute) with actual millisecond durations from the execution provenance trace. Makes engine overhead visible and comparable across runs. Clickable stages expand a detail panel showing trace events at that stage.

**T-26 · Trace Event Drill-Down** [NEW] P3
An extension of T-25. Each pipeline stage is expandable to a structured event list: `TraceEvent` records with kind, timestamp, and payload. Exposes what the planner decided, which capability contracts resolved, which environment was activated, which overlay patches were applied. Currently the `Provenance:` section only renders as flat text in CLI output.

**T-27 · Validation Coverage Indicator** [NEW] P2
Each step node in the Workflow Canvas shows a 4-segment coverage bar: L0 (no assertions), L1 (status only), L2 (status + body), L3 (full coverage). Canvas footer shows coverage distribution. Derived entirely from assertion definitions already authored — no engine changes required. Makes the progressive validation philosophy visible and actionable.

**T-28 · Validation Maturity Score** [NEW] P2
A workflow-level maturity score in the S5 results header and S3 canvas sidebar. Shows coverage distribution: "L0: 0 · L1: 3 · L2: 2 · L3: 1 steps." A "Next step" prompt suggests the single highest-impact assertion addition to improve coverage. Translates the progressive validation philosophy into a concrete, actionable metric per workflow.

**T-29 · Failure Blast Radius Preview** [NEW] P3
In the Workflow Canvas, hovering a step while blast radius toggle is active shades which downstream steps would cancel under `failFast` vs survive under `resilient`. Computed statically from `dependsOn` edges — no run required. Updates live when the T-19 failure policy selector is toggled. No Apidog equivalent — requires the DAG model they do not have.

---

## Phase 0.6 — Determinism + History

**T-30 · Replay with Same Seed** [NEW] P2
A "Replay with same seed" button on run history entries and the S5 results header. Re-runs the workflow with the identical trace ID seed, producing identical `tractl.now()` and `tractl.random()` values in scripts. Makes script-influenced runs reproducible for debugging. Seed already stored per run history entry — runner invocation with seed passed explicitly rather than generated fresh.

**T-31 · Run Comparison / Diff** [NEW] P3
Select two run history entries → show a structured diff: which assertions changed outcome, which extracts produced different values, which timing changed beyond a threshold. Enabled by the determinism model. No competitor has this because no competitor has seeded deterministic execution as an architectural guarantee.

**T-32 · Workflow YAML Diff View** [NEW] P3
When git detects uncommitted changes to a workflow file, show an inline diff of what changed since last commit — directly in the canvas or editor sidebar. Renders as side-by-side YAML diff using git status data already read via shell (status indicator already specced). Only available when workspace is inside a git repository.

**T-33 · Copy CI Command** [NEW] P2
One-click action copying the fully-resolved CLI command to clipboard: `tractl run ./workflows/user-lifecycle.yaml --env staging --overlay ./overlays/ci.yaml`. Includes `--overlay` flags for any active overlays and `--env` flag. Available in `···` context menu, run results header, and as a keyboard shortcut. Bridges local execution to CI configuration instantly.

---

## Phase 0.7 — Diagnostics Visualisation
*Deferred until Phase 7.5 HTTP instrumentation is wired in Go*

**T-34 · DevTools-Style Request Timeline** [NEW] P3
Once Phase 7.5 HTTP instrumentation is complete, render a Chrome DevTools-style coloured segment bar per step: DNS, TCP, TLS, TTFB, content transfer with ms labels. Types and constants exist (`KindTLS`, `KindTCP`, `KindDNS` in `internal/diagnostics`) but runtime population is deferred. Wires `DiagnosticsEvent` data into S5 step expansion and S4 result tab.

**T-35 · TLS Inspector** [NEW] P3
When `diagnostics.kinds` includes `tls`, show in step detail: negotiated TLS version, cipher suite, certificate chain (subject, issuer, expiry, SANs), handshake duration. Data typically only visible in Wireshark or `curl -v`. Surfaces it alongside assertion outcomes directly in the run results UI.

---

## Phase 1.0 — Alpha Gate

**T-36 · Credential Manager Screen (S11)** [EXISTING] P3
Single-phase screen: list credential names (never raw values), store via OS keychain (Wails on desktop, browser credential store on web), delete, reference in workflow files as `${secrets.env.name}`. Masked display throughout — no raw values per ADR-006. Unblocked after T-08 (Settings).

---

## Summary

| Task | Title | Type | Priority |
|---|---|---|---|
| T-01 | Workflow Persistence + Workspace Integration | EXISTING | P1 |
| T-02 | Desktop File Load API | EXISTING | P1 |
| T-03 | Environment Manager (S6) | EXISTING | P1 |
| T-04 | Monaco Wiring + ctx Types | EXISTING | P1 |
| T-05 | Script Snippet Library | NEW | P1 |
| T-06 | Script Return Shape Validator | NEW | P2 |
| T-07 | Script Editor Screen (S8) | EXISTING | P2 |
| T-08 | Settings Screen (S10) | EXISTING | P2 |
| T-09 | S4 Step Detail Full Completion | EXISTING | P2 |
| T-10 | Overlay Editor Screen (S7) | EXISTING | P2 |
| T-11 | Overlay Before/After Diff Preview | NEW | P2 |
| T-12 | Overlay Semantic Match Visualiser | NEW | P2 |
| T-13 | JSONPath Live Tester | NEW | P1 |
| T-14 | Click-to-Extract from Response Body | NEW | P1 |
| T-15 | Mutation Set Result Panel | NEW | P2 |
| T-16 | Variable Scope Explorer | NEW | P2 |
| T-17 | Inline Expression Preview | NEW | P2 |
| T-18 | Unresolved Variable Warning | NEW | P2 |
| T-19 | Failure Policy Selector with Live Preview | NEW | P2 |
| T-20 | Workflow Run Results Screen (S5) | EXISTING | P1 |
| T-21 | Wait-Time Annotation in Waterfall | NEW | P2 |
| T-22 | Critical Path Highlighting | NEW | P2 |
| T-23 | Parallelism Swim Lane Annotation | NEW | P2 |
| T-24 | Outcome-Filtered Waterfall | NEW | P2 |
| T-25 | Pipeline Stage Timeline | NEW | P2 |
| T-26 | Trace Event Drill-Down | NEW | P3 |
| T-27 | Validation Coverage Indicator | NEW | P2 |
| T-28 | Validation Maturity Score | NEW | P2 |
| T-29 | Failure Blast Radius Preview | NEW | P3 |
| T-30 | Replay with Same Seed | NEW | P2 |
| T-31 | Run Comparison / Diff | NEW | P3 |
| T-32 | Workflow YAML Diff View | NEW | P3 |
| T-33 | Copy CI Command | NEW | P2 |
| T-34 | DevTools-Style Request Timeline | NEW | P3 |
| T-35 | TLS Inspector | NEW | P3 |
| T-36 | Credential Manager Screen (S11) | EXISTING | P3 |

Total: 36 · Existing: 11 · New: 25
P1: 7 · P2: 20 · P3: 9
