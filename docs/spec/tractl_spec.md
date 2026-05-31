# traCtlSpec — Canonical Specification

> Version 1.5 · Canonical Semantic Model Source of Truth

> **Revision 1.5:** Default `failurePolicy` changed from `failFast` to `resilient`. Changes: (1) §6.1 workflow structure updated: `failurePolicy` default is now `resilient`. (2) §6.3 failure policy table updated: `resilient` is now the default policy; `failFast` remains available as an opt-in. Rationale: `resilient` aligns with the dependency-aware DAG execution model, product onboarding expectations, progressive backend validation workflow model, HLD scheduler semantics, and product differentiation goals. Both policies preserve the branch isolation invariant (ADR-008 §5); the change in default affects only whether in-flight dependent work receives an active cancellation signal (`failFast`) or is passively skipped at the next scheduling cycle (`resilient`).

> **Revision 1.4:** Governance sync with the Overlay Specification, ADR-002 §3a, ADR-004 §7, and ADR-006 §9. Changes: (1) §2 rewritten to make the pre-execution pipeline (parser/source-adapter → canonical source model → overlay engine → canonical traCtlSpec → validation → planner → execution) explicit, and to assert that canonical traCtlSpec is post-overlay normalized output. (2) §5.1 metadata expanded with explicit `provenance` field and explicit rule that the planner MUST ignore metadata for execution decisions. (3) §6.2 dependency validation ownership clarified: basic acyclicity and `dependsOn` reference resolution moved from planner to canonical validator; composite-cycle resolution remains planner-owned (§16.3). (4) §12.1 and §12.2 updated to reflect the validator/planner split. (5) §18.2 / §18.3 conformance updated to reflect the validator-owned dependency graph validation and to add overlay merger requirements. No semantic changes to execution behavior; this revision realigns ownership boundaries only.

> **Revision 1.3:** Three targeted corrections. (1) §3.2 capability wording restructured: "Capability resolution uses semantic contracts" replacing the philosophically inconsistent "Capabilities are versioned semantic contracts" opening. (2) §7.3 / §7.6 / §12.1 step skip semantics defect resolved: introduced canonical step execution states (`succeeded`, `failed`, `conditional-skip`, `dependency-skipped`) in new §7.6 as authoritative reference; §7.3 updated to name conditional-skip explicitly and state it does not propagate; §12.1 updated to reference §7.6 states and resolve the §7.3/§12.1 contradiction. (3) §16.3 added: explicit planner obligation for recursive DAG flattening before composite cycle detection. No other changes.

> **Revision 1.2:** §6.3 Failure Policy corrected. `failFast` redefined as dependency-subgraph cancellation. §12.1 skipped-step rule updated. No other changes.

> Version 1.1 · Canonical Semantic Model Source of Truth
-----

# 1. Purpose

This document defines **traCtlSpec**, the canonical semantic model of the traCtl platform.

traCtlSpec is the normalized, protocol-neutral, runtime-independent, syntax-independent representation of executable intent on which all traCtl planning and execution operates.

This document defines:

- the role of traCtlSpec
- versioning rules
- the capability contract model
- identity rules
- top-level structure
- workflow semantics
- step semantics
- request semantics
- assertion semantics
- extraction semantics
- scripting hooks
- fuzz directives
- extension invocation surface
- diagnostics directives
- dependency semantics
- variable and context semantics
- capability declarations
- conformance rules

This document does **not** define:

- file syntax (TOON, YAML, JSON authoring formats)
- overlay file format
- extension SDK contracts
- MCP tool contracts
- protocol-specific transport semantics
- scheduler implementation
- detailed AI provider contracts

These are specified separately.

-----

# 2. Role of traCtlSpec

traCtlSpec is the **canonical semantic contract** of the traCtl platform.

Per the HLD, all execution operates on normalized traCtlSpec semantics. traCtlSpec sits between the Input Layer (parsers, source adapters, overlay engine) and the Planning Layer (capability resolution, DAG construction, runtime selection).

Canonical pipeline (native and interoperability sources converge here):

```text
Input Layer
    ↓
Native Parser  /  Source Adapter
    ↓
Canonical Source Model (+ provenance for interoperability sources)
    ↓
Overlay Engine
    ↓
Canonical traCtlSpec
    ↓
Validation
    ↓
Planning Layer
    ↓
Execution
```

Canonical traCtlSpec represents **post-overlay normalized output**. Overlay merge MUST be:

- complete (no deferred merge state)
- deterministic
- provenance retaining
- validation compliant

Overlay resolution MUST complete before canonical validation. Overlay is a pre-execution transformation. Overlay MUST NOT become runtime input directly. See the Overlay Specification for the full overlay contract.

traCtlSpec is:

- the **only** representation Planning consumes
- the **only** representation Execution operates on
- the merged result of native inputs, interoperability sources, and overlays

traCtlSpec is **not**:

- a wire format
- an authoring format
- a runtime
- a protocol
- an interoperability source format
- an overlay artifact

Internal representations are implementation-defined. On-disk or on-wire serialization (when needed for tooling, MCP, or transport) uses the stable canonical JSON projection defined in §17.

-----

# 3. Versioning and Capability Model

## 3.1 Schema Version

traCtlSpec uses an **integer schema version** combined with **capability declarations**.

Every traCtlSpec document declares:

- `schemaVersion` — integer, monotonically increasing, starting at `1`
- `capabilities` — declared capability contracts required by the document

Rules:

- Schema version increments only on breaking structural changes.
- Additive, backward-compatible changes do **not** bump the schema version.
- New optional features are introduced via capability declarations, not version bumps.
- A runtime that does not satisfy a declared capability MUST refuse to execute the spec.

## 3.2 Capability Contract Model

Capability resolution uses **semantic contracts**. A declared capability requirement resolves against the contracts available in the execution environment by contract compatibility, never by the name of the component that provides the capability.

Capabilities resolve through one canonical contract model regardless of source. Engine capabilities, extension capabilities, provider adapter capabilities, and MCP adapter capabilities all normalize into the same capability contract representation consumed by the planner.

Capabilities are tiered. The tiers differ in how the contract version is sourced:

- **Platform capabilities** derive their version from engine context: governed by `schemaVersion` and the engine version, not by an independent suffix. Platform capabilities are referenced as stable flags.
- **Extension and provider capabilities** declare their versions explicitly, in `identifier@version` form, because their semantics drift independently of the engine.

Platform capability examples (non-exhaustive):

- `protocol.http`
- `protocol.grpc`
- `protocol.websocket`
- `protocol.mqtt`
- `protocol.tcp`
- `protocol.udp`
- `protocol.graphql`
- `scripting.js`
- `fuzz.builtin`
- `fuzz.advanced`
- `diagnostics.tls`
- `diagnostics.tcp`
- `extension.invoke`
- `ai.assist`

Extension and provider capability examples (non-exhaustive):

- `protocol.ext.amqp@1`
- `secret.ext.vault@1`
- `assertion.ext.jsonschema@2`
- `fuzz.ext.schemathesis@1`

Capability declarations are **requirements**, not feature toggles. They allow:

- runtime selection (browser vs agent vs embedded)
- early rejection of unsupported specs
- deterministic capability negotiation across runtimes

## 3.3 Compatibility Rules

- Specs MUST declare every capability they require.
- Planners MUST resolve declared capabilities against available capability contracts by contract compatibility.
- Planners MUST reject specs requiring unknown capability contracts.
- Planners MUST reject specs requiring incompatible capability versions.
- Planners MUST select the minimal runtime that satisfies declared capabilities.
- Silent capability substitution and best-effort capability guessing are NOT permitted.
- Adding a new platform capability flag is **not** a breaking change.
- Removing or redefining a platform capability flag IS a breaking change and bumps `schemaVersion`.
- For extension and provider capabilities, version compatibility follows semantic versioning. A required `identifier@N` is satisfied by a compatible contract version under the same major version and is not satisfied across major versions absent an explicit compatibility declaration.

-----

# 4. Identity Model

traCtlSpec uses a **hybrid identity model**.

## 4.1 User IDs (Required)

Every identifiable entity (workflow, step, assertion, extract, request) MUST carry a user-defined stable string `id`.

Rules for user IDs:

- non-empty
- unique within their containing scope
- stable across runs
- match the pattern `^[a-zA-Z][a-zA-Z0-9_.-]*$`
- not reserved (reserved prefixes: `_traCtl.`, `traCtl.`)

User IDs are the **primary** identity. They appear in:

- dependency references
- extract references
- variable references
- assertion targets
- reporter outputs

## 4.2 ULID Fallback

Every identifiable entity also carries a system-assigned ULID under the canonical field `_ulid`.

ULIDs are:

- auto-generated at normalization time
- stable for the lifetime of the traCtlSpec instance
- not stable across re-normalization
- used internally by execution, diagnostics, and observability
- never referenced by user authoring

ULIDs serve as the identity used by:

- the execution context engine
- diagnostics and tracing
- internal indexing
- MCP responses requiring globally-unique handles

## 4.3 Identity Scope

|Entity   |ID Scope      |
|---------|--------------|
|Workflow |Spec-global   |
|Step     |Workflow-local|
|Assertion|Step-local    |
|Extract  |Step-local    |
|Request  |Step-local    |

References across scopes use dotted paths (e.g., `workflowA.stepB.extracts.token`).

-----

# 5. Top-Level Structure

A traCtlSpec document has the following top-level shape:

```text
traCtlSpec
├── schemaVersion        (integer, required)
├── capabilities         (list of capability contracts, required)
├── metadata             (descriptive, optional)
├── variables            (spec-scoped variables, optional)
├── auth                 (auth profile definitions, optional)
├── environments         (environment definitions, optional)
├── workflows            (list, required, ≥1)
└── extensions           (extension references, optional)
```

## 5.1 metadata

Free-form descriptive metadata. Examples:

- `name`
- `description`
- `tags`
- `sourceFormat` (e.g., `openapi`, `postman`, `native.toon`)
- `sourceRef` (path or URI of original interoperability source)
- `overlayRefs` (list of overlay sources merged into this spec, in application order)
- `provenance` (per-field contribution attribution emitted by the overlay engine; see Overlay Specification §13)

### 5.1.1 Canonical Provenance Structure

The `provenance` field MUST conform to the following minimal canonical structure when present:

```yaml
metadata:
  provenance:
    <canonical.dot.path>:
      source: <overlay-file-identifier>
      patchIndex: <integer>
```

Where:

- `<canonical.dot.path>` is the canonical dotted path of the contributed field (e.g., `workflows.main.steps.login.request.headers.Authorization`)
- `source` is a string identifier for the overlay file that contributed the field (e.g., `overlay.prod.yaml`)
- `patchIndex` is the zero-based integer index of the patch within that overlay file that contributed the field

Example:

```yaml
metadata:
  provenance:
    workflows.main.steps.login.request.headers.Authorization:
      source: overlay.prod.yaml
      patchIndex: 2
    workflows.main.steps.login.timeout:
      source: overlay.dev.yaml
      patchIndex: 0
```

Rules:

- The provenance map key MUST be a valid canonical dotted path.
- `source` MUST be non-empty and MUST reference a file listed in `overlayRefs`.
- `patchIndex` MUST be a non-negative integer corresponding to a patch in the named overlay file.
- Fields not contributed by an overlay MAY be omitted from `provenance`.
- `provenance` MAY be omitted entirely when no overlay was applied.

Implementations MUST emit provenance in this structure. Tooling, diagnostics, and audit consumers MUST NOT rely on provenance schemas outside this structure.

Metadata is informational only. It is consumed by diagnostics, audit, and tooling surfaces.

Rules:

- Metadata MUST NOT influence execution semantics.
- Execution components MUST ignore metadata unless explicitly declared diagnostic-only. No execution component MAY alter scheduling, capability resolution, or runtime behavior based on metadata.
- The planner MUST ignore metadata for execution decisions.
- Validation MUST NOT consume metadata as a substitute for declared schema, capability, or dependency fields.
- Two specs differing only in metadata MUST produce identical execution behavior.

## 5.2 variables

Spec-scoped variables. See §13 for full variable semantics.

## 5.3 auth

Named, reusable authentication profile definitions. Steps reference profiles by `id`.

Auth profiles are **declarative**; the actual provider implementation is supplied by the engine core or by an auth provider extension (extensionKind: `auth`), per the Extension SDK Architecture.

## 5.4 environments

Named environment definitions providing override values for variables.

A single environment is active per execution. Resolution order is defined in §13.4.

## 5.5 workflows

One or more workflow definitions. See §6.

## 5.6 extensions

Declared extension dependencies required by the spec.

Each entry includes:

- `id` — extension identifier
- `version` — version requirement
- `capabilities` — capability contracts provided by the extension that this spec depends on, each declared in `identifier@version` form

Planners MUST verify extension availability and resolve declared extension capability contracts before execution.

-----

# 6. Workflow Semantics

A **workflow** is an ordered, dependency-aware collection of steps that share execution context.

## 6.1 Workflow Structure

```text
workflow
├── id                   (required)
├── _ulid                (system-assigned)
├── description          (optional)
├── variables            (workflow-scoped, optional)
├── auth                 (default auth profile reference, optional)
├── concurrency          (workflow-level concurrency policy, optional)
├── failurePolicy        (resilient | failFast, default: resilient)
├── steps                (required, ≥1)
└── hooks                (workflow-level lifecycle hooks, optional)
```

## 6.2 Execution Semantics

- Steps within a workflow form a **directed acyclic graph (DAG)** via dependency declarations (§12).
- Steps with no declared dependencies are eligible for parallel execution, subject to `concurrency`.
- Steps with declared dependencies execute only after all dependencies complete successfully (or per `failurePolicy`).
- The DAG MUST be acyclic.

**Validation ownership:**

- Basic DAG acyclicity and `dependsOn` reference resolution are **canonical validator** responsibilities. The canonical validator MUST reject cyclic workflows and unresolved `dependsOn` references before planning.
- Composite-cycle detection requiring recursive sub-workflow flattening is a **planner** responsibility (§16.3).
- Capability resolution, runtime selection, and scheduling decisions are **planner** responsibilities (§18.3).

This ownership split is consistent with the Overlay Specification §15 and the canonical pipeline in §2.

## 6.3 Failure Policy

|Mode        |Behavior |
|------------|---------|
|`resilient` |A step failure is recorded. Steps that would consume the failed step's output are skipped at scheduling time. No additional cancellation signal is issued. Independent branches always continue. **Default.** |
|`failFast`  |A step failure actively cancels all steps in its **transitive dependency subgraph**. Steps with no dependency path to the failed step are not affected and continue executing. |

**DAG invariant (both policies):** a step's dependency subgraph is always skipped on upstream failure, regardless of `failurePolicy`. This invariant is unconditional. The policies differ only in whether in-flight dependent work is actively cancelled (`failFast`) or passively skipped at the next scheduling cycle (`resilient`). Independent branches are never affected by either policy (ADR-008 §5).

`failurePolicy` interacts with `dependsOn` (§12): the dependency subgraph subject to `failFast` cancellation is the transitive closure of `dependsOn` edges from the failed step forward.

## 6.4 Concurrency

Workflow-level `concurrency` declares the maximum parallel step count for the workflow. Absence means "platform default."

Concurrency is a **scheduling hint** that MUST be respected as an upper bound but MAY be reduced by the planner based on runtime constraints.

## 6.5 Hooks

Workflow lifecycle hooks:

- `beforeAll` — runs once before any step
- `afterAll` — runs once after all steps (success or failure)
- `beforeEach` — runs before each step
- `afterEach` — runs after each step

Hooks are scripts (§9) and execute in the Script Sandbox.

-----

# 7. Step Semantics

A **step** is the atomic unit of execution within a workflow.

## 7.1 Step Structure

```text
step
├── id                   (required)
├── _ulid                (system-assigned)
├── description          (optional)
├── kind                 (required: request | script | extensionCall | composite)
├── dependsOn            (list of step IDs, optional)
├── when                 (conditional expression, optional)
├── auth                 (auth profile reference, optional, overrides workflow)
├── retry                (retry policy, optional)
├── timeout              (duration, optional)
├── request              (required if kind=request)
├── script               (required if kind=script)
├── extensionCall        (required if kind=extensionCall)
├── composite            (required if kind=composite)
├── assertions           (list, optional)
├── extracts             (list, optional)
├── fuzz                 (fuzz directive, optional)
├── diagnostics          (diagnostics directive, optional)
└── hooks                (step-level lifecycle hooks, optional)
```

## 7.2 Step Kinds

|Kind           |Purpose                                                           |
|---------------|------------------------------------------------------------------|
|`request`      |Execute a protocol request via the Protocol Dispatcher.           |
|`script`       |Execute a script in the Script Sandbox without a protocol request.|
|`extensionCall`|Invoke an extension-provided capability (§14).                    |
|`composite`    |Reference and inline-execute another workflow (sub-workflow).     |

Exactly one of `request`, `script`, `extensionCall`, or `composite` MUST be present, matching `kind`.

## 7.3 Conditional Execution

`when` is an expression evaluated against the current context. If `when` evaluates to false, the step enters **conditional-skip** state (see §7.6). A conditional-skip is not a failure.

For dependency evaluation, a conditionally-skipped step is treated as **succeeded with empty extractions**: its dependents are eligible to execute and will find no extractions available from it. A conditional-skip does NOT propagate skipping to dependents.

Expression language is defined in §13.5.

## 7.4 Retry and Timeout

`retry` declares:

- `maxAttempts`
- `backoff` (one of `fixed`, `linear`, `exponential`)
- `delay`
- `retryOn` (condition: status codes, error classes, or assertion failures)

`timeout` is a duration applied to the step's execution envelope. Timeouts MUST propagate cancellation through dependent operations.

## 7.5 Step-Level Hooks

`beforeStep` and `afterStep` hooks run scoped to a single step.

## 7.6 Step Execution States

Every step terminates in one of the following canonical states. These states govern dependency propagation and are the authoritative reference for all step-outcome rules in this document.

|State              |Cause                                                      |Dependency propagation to dependents        |
|-------------------|-----------------------------------------------------------|--------------------------------------------|
|`succeeded`        |Step executed; all error-severity assertions passed.       |Dependents are eligible.                    |
|`failed`           |Execution error, or one or more `error`-severity assertion failures. |Dependents enter `dependency-skipped`.      |
|`conditional-skip` |`when` expression evaluated to false.                      |Treated as `succeeded` with empty extractions; dependents are eligible. |
|`dependency-skipped`|A direct dependency has status `failed` or `dependency-skipped`.|Dependents also enter `dependency-skipped`. |

Rules:

- `conditional-skip` and `dependency-skipped` are distinct states with distinct propagation semantics.
- A step whose dependency is `conditional-skip` is NOT `dependency-skipped`; it is eligible and proceeds with no extractions from the skipped step.
- A step whose dependency is `failed` or `dependency-skipped` enters `dependency-skipped` regardless of `failurePolicy`. The `failurePolicy` governs only whether the propagation is immediate active cancellation (`failFast`) or passive scheduling-time skipping (`resilient`); the outcome state is identical under both.
- `warning`-severity assertion failures do not cause `failed` state; the step reaches `succeeded`.

-----

# 8. Request Semantics

A **request** describes a protocol-level invocation in a protocol-neutral manner.

## 8.1 Request Structure

```text
request
├── protocol             (required: http | graphql | grpc | websocket | mqtt | tcp | udp | ext:<id>)
├── target               (required, protocol-specific target descriptor)
├── operation            (protocol-specific operation, optional)
├── headers              (key-value map, optional)
├── metadata             (protocol metadata, optional)
├── body                 (body descriptor, optional)
├── tls                  (TLS profile, optional)
└── transport            (transport options, optional)
```

## 8.2 Protocol Neutrality

Request fields are **shape-neutral**:

- `target` describes "what to talk to" (URL, gRPC service+method, MQTT topic, host:port, etc.)
- `operation` describes "what action" (HTTP verb, gRPC method, GraphQL operation, MQTT pub/sub, etc.)
- `headers` describes protocol metadata pairs (HTTP headers, gRPC metadata, MQTT user properties, etc.)
- `body` describes the payload with declared encoding

Protocol-specific dialects live in protocol provider extension points, not in traCtlSpec itself. The canonical model captures only the protocol-neutral shape; protocol providers interpret the dialect.

## 8.3 Body Descriptors

A `body` carries:

- `encoding` (e.g., `json`, `text`, `binary`, `form`, `multipart`, `protobuf`, `graphql`)
- `content` (the payload, encoding-appropriate)
- `schemaRef` (optional reference to a schema definition for validation/fuzzing)

## 8.4 TLS and Transport

`tls` and `transport` are protocol-orthogonal. Examples:

- `tls.minVersion`, `tls.verify`, `tls.clientCert`
- `transport.proxy`, `transport.localAddress`, `transport.dnsResolver`

Runtimes that cannot honor declared transport options MUST fail capability negotiation.

-----

# 9. Scripting Hooks

Scripting is exposed at workflow, step, and request lifecycle points.

## 9.1 Hook Locations

|Hook                 |Scope   |When                               |
|---------------------|--------|-----------------------------------|
|`workflow.beforeAll` |Workflow|Once, before first step            |
|`workflow.afterAll`  |Workflow|Once, after last step              |
|`workflow.beforeEach`|Workflow|Before every step                  |
|`workflow.afterEach` |Workflow|After every step                   |
|`step.beforeStep`    |Step    |Before the step body               |
|`step.afterStep`     |Step    |After the step body                |
|`step.transform`     |Step    |Transform request/response/extracts|

## 9.2 Script Descriptor

```text
script
├── language             (required: js)
├── source               (inline source) OR
├── sourceRef            (reference to script asset)
└── inputs               (declarative inputs, optional)
```

`language: js` is the only canonical scripting language in schema v1. Additional languages may be introduced via extensions and capability declarations (e.g., `scripting.starlark`).

## 9.3 Script Contract

All scripts receive a **frozen context object** (§13.6) and may produce a **mutation set** describing safe context changes:

- variable writes
- extract writes
- assertion additions
- log entries
- cancellation signals

Scripts MUST NOT perform direct I/O. All I/O is mediated by the runtime. Sandbox boundaries are defined by the Script Sandbox and the Security Architecture.

## 9.4 Determinism

Scripts SHOULD be deterministic. Non-deterministic operations (e.g., random, current time) are provided via injected APIs that the runtime can seed or record for reproducibility.

-----

# 10. Fuzz Directives

Fuzzing is declared inline at the step level via the `fuzz` directive.

## 10.1 Fuzz Directive Structure

```text
fuzz
├── enabled              (boolean, default: false)
├── strategy             (builtin | advanced | ext:<id>)
├── seed                 (integer, optional, for reproducibility)
├── budget               (cases | duration, optional)
├── targets              (list of fuzz target descriptors, optional)
├── invariants           (list of invariant references, optional)
└── config               (strategy-specific config, opaque to core)
```

## 10.2 Strategies

|Strategy   |Behavior                                                                                |
|-----------|----------------------------------------------------------------------------------------|
|`builtin`  |Boundary mutations, invalid payloads, required-field omission, malformed requests.      |
|`advanced` |Schema-driven, stateful, property-based, AI-guided. Requires `fuzz.advanced` capability.|
|`ext:<id>` |Delegated to a fuzz provider extension (e.g., Schemathesis adapter, property-based fuzz engine).|

## 10.3 Targets

A fuzz target identifies a portion of the request to mutate:

- request path parameter
- query parameter
- header
- body field (by JSON pointer or schema path)
- metadata field

## 10.4 Invariants

Invariants are assertions (§11) that MUST hold across all fuzz cases. Invariants are referenced by ID and evaluated per case.

## 10.5 Determinism

Fuzz runs with an explicit `seed` MUST be reproducible across runs on the same runtime version and strategy. Strategies that cannot guarantee reproducibility MUST declare so via capability.

-----

# 11. Assertions and Extractions

## 11.1 Assertion Structure

```text
assertion
├── id                   (required)
├── _ulid                (system-assigned)
├── description          (optional)
├── kind                 (required: status | header | body | schema | script | extension)
├── target               (assertion target, kind-specific)
├── op                   (operator: equals | matches | contains | exists | inRange | jsonpath | custom)
├── expected             (expected value, op-specific)
├── severity             (error | warning, default: error)
└── config               (kind-specific config, optional)
```

## 11.2 Assertion Kinds

|Kind       |Purpose                                                     |
|-----------|------------------------------------------------------------|
|`status`   |Status / code-level checks (HTTP status, gRPC status, etc.).|
|`header`   |Header / metadata key checks.                               |
|`body`     |Payload value checks via path expression.                   |
|`schema`   |Validate payload against a referenced schema.               |
|`script`   |Evaluate a script that returns a boolean.                   |
|`extension`|Delegate to an extension-provided assertion type.           |

## 11.3 Extraction Structure

```text
extract
├── id                   (required)
├── _ulid                (system-assigned)
├── source               (required: status | header | body | metadata | timing | extension)
├── path                 (path expression, source-specific)
├── as                   (variable name to bind, defaults to id)
└── scope                (step | workflow | spec, default: workflow)
```

Extractions write into context (§13). Reference syntax: `${steps.<stepId>.extracts.<extractId>}`.

## 11.4 Severity

- `error` assertions fail the step.
- `warning` assertions record diagnostics but do not fail the step.

`severity` interacts with `failurePolicy`: warnings never cancel dependent steps.

-----

# 12. Dependency Semantics

## 12.1 dependsOn

A step declares dependencies via `dependsOn`, a list of step IDs within the same workflow.

Rules:

- IDs MUST reference existing steps in the same workflow.
- Unresolved `dependsOn` references MUST be rejected by the canonical validator (per §6.2).
- Cycles MUST be rejected before planning. Single-workflow cycles are rejected by the canonical validator. Cycles requiring composite-resolution MUST be rejected by the planner per §16.3.
- A step is **eligible** when all dependencies have reached a terminal state.
- A step is **executed** when eligible and all dependencies are in state `succeeded` or `conditional-skip` (§7.6).
- A step enters **`dependency-skipped`** state when any direct dependency is in state `failed` or `dependency-skipped` (§7.6). A dependency in `conditional-skip` state does NOT cause `dependency-skipped`; the dependent is eligible and proceeds with no extractions from that dependency.
- Under `failFast`, a `dependency-skipped` transition for in-flight work occurs via active cancellation. Under `resilient`, it occurs at the next scheduling cycle. The terminal state is `dependency-skipped` under both policies.

Step execution states and their propagation rules are defined in §7.6, which is the authoritative reference.

## 12.2 Implicit Dependencies

Variable references (§13.5) that resolve to extracts from other steps create **implicit data dependencies**.

Implicit dependency detection is a **normalization-time** operation. The normalizer (producer, per §18.2) MUST:

- detect implicit dependencies from variable references during normalization
- merge implicit dependencies with declared `dependsOn` edges into the canonical DAG

The canonical validator MUST reject cycles produced by the merged DAG (declared edges + implicit edges), per §6.2.

The resulting canonical DAG carries explicit edges only; consumers do not re-derive implicit edges from variable references.

## 12.3 Parallelism

Steps without dependencies (declared or implicit) are eligible for parallel execution.

Parallel execution semantics are deterministic under fixed `seed` and concurrency settings. The Deterministic Scheduler, as defined in the Runtime and Concurrency Architecture, owns execution ordering guarantees.

-----

# 13. Variables, Context, and Expressions

## 13.1 Variable Scopes

|Scope        |Defined In              |Visible To                               |
|-------------|------------------------|-----------------------------------------|
|`spec`       |top-level `variables`   |all workflows, all steps                 |
|`environment`|active environment      |all workflows, all steps (overrides spec)|
|`workflow`   |workflow `variables`    |all steps in that workflow               |
|`step`       |step extracts and locals|that step and downstream steps           |
|`runtime`    |runtime-injected        |read-only, all scopes                    |

## 13.2 Resolution Order

Variable resolution is **most-specific-wins**:

```text
step > workflow > environment > spec > runtime defaults
```

Unresolved variables at execution time produce a **resolution error**, which is a planning/early-execution failure, not a request failure.

## 13.3 Context Object

The execution context exposes:

- `vars` — resolved variables
- `env` — active environment metadata
- `steps.<id>.request`
- `steps.<id>.response`
- `steps.<id>.extracts`
- `steps.<id>.assertions`
- `steps.<id>.timing`
- `workflow.id`, `workflow.metadata`
- `runtime.*` — runtime-injected info (deterministic time, seed, etc.)

Context is the substrate consumed by expressions, scripts, and assertions.

## 13.4 Environment Activation

A single environment is active per execution. Activation source priority:

```text
CLI/MCP flag > spec metadata default > none
```

When no environment is active, only spec-level variables apply.

## 13.5 Expression Language

Inline expressions use `${ ... }` interpolation. The expression grammar supports:

- variable lookup: `${vars.token}`
- nested path: `${steps.login.extracts.token}`
- JSONPath-like body access: `${steps.list.response.body.items[0].id}`
- safe call: `${vars.x ?? "default"}`
- simple comparisons in `when`: `${steps.login.response.status == 200}`

Variable reference rules:

- Use `${...}` everywhere; alternate placeholder syntaxes are invalid.
- Use `${vars.name}` for user-defined variables from spec, workflow, environment, step locals, or script mutations.
- Use `${steps.stepId.extracts.name}` when referencing a specific previous step extract.
- Use `${env.*}` only for predefined active environment metadata, not user-defined environment variables.
- Supported `env.*` keys: `${env.id}`, `${env.name}`, `${env.source}`.
- Use `${workflow.id}` and `${workflow.metadata.*}` for workflow metadata.
- Use `${runtime.*}` only for runtime-injected read-only values.
- Resolve `vars.*` by precedence: step > workflow > environment > spec > runtime defaults.
- Missing variables are execution resolution errors, not request failures.

The expression language is intentionally minimal. For richer logic, use `script` steps or transform hooks.

## 13.6 Frozen Context for Scripts

Scripts receive a frozen snapshot of context at invocation time. Mutations are returned as a mutation set (§9.3) and applied transactionally after the script returns. This preserves determinism and enables replay.

-----

# 14. Extension Invocation Surface

Extensions integrate with traCtlSpec at four points.

## 14.1 Extension Step (`kind: extensionCall`)

```text
extensionCall
├── extension            (required, extension id)
├── operation            (required, operation name)
├── input                (operation input, schema declared by extension)
└── outputBindings       (extract-style bindings from output)
```

The Extension Platform resolves `extension` + `operation` to the registered extension and dispatches the invocation through the Extension Runtime.

## 14.2 Extension Assertion (`assertion.kind: extension`)

Assertions delegated to an extension. The extension declares:

- assertion operator names
- input schema
- pass/fail semantics

## 14.3 Auth, Reporter, Secret, and Data Provider Extensions

Extensions with extensionKind `auth`, `reporter`, `secret`, or `data` are declared via `extensions` (§5.6) and referenced by ID elsewhere in the spec (e.g., `auth.provider: ext:<id>`).

## 14.4 Extension Protocols

A protocol supplied by an extension appears in `request.protocol` as a namespaced value (e.g., `ext:amqp`). The corresponding extension capability is declared as a versioned capability contract on the spec (e.g., `protocol.ext.amqp@1`).

## 14.5 Invocation Contract

The Extension SDK Architecture defines the wire contract. traCtlSpec defines only the **invocation surface** — what fields appear in the spec and how outputs bind back into context.

-----

# 15. Diagnostics Directives

Diagnostics are **opt-in per step** via the `diagnostics` directive.

## 15.1 Diagnostics Structure

```text
diagnostics
├── enabled              (boolean, default: false)
├── kinds                (list: tls | tcp | transport | lifecycle | dns | connection)
├── captureBody          (boolean, default: false)
└── retention            (step | workflow | spec, default: workflow)
```

## 15.2 Semantics and Scope

Diagnostics are **execution-context observability surfaces** attached to workflow execution.

**What diagnostics provide:**

- Request/connection execution visibility (handshake states, protocol negotiation, timing events, error state transitions)
- Step-scoped execution observation surfaces for debugging and validation

**What diagnostics are NOT:**

- Packet inspection tooling (diagnostics do not capture or decode raw packet content)
- Infrastructure observability (diagnostics are not infrastructure monitoring or network-level introspection)
- Host-level telemetry systems (diagnostics do not capture system metrics, process state, or host-level events)

Diagnostics kinds are bounded to execution-context observability:

- `tls` — TLS handshake state and configuration (version, cipher, certificate validation)
- `tcp` — TCP connection state and timing (established, reset, timeout events)
- `transport` — Transport-layer behavior (backpressure, flow control, connection reuse)
- `lifecycle` — Step execution lifecycle events (start, completion, failure)
- `dns` — DNS resolution events and outcomes
- `connection` — Connection pooling and multiplexing state

No additional kinds beyond these six are admitted without explicit architecture extension.

## 15.3 Execution Semantics

- Diagnostics MUST NOT change request execution semantics.
- Diagnostics output is attached to the step result, not to assertions or extracts.
- Diagnostics requiring privileged transport (TLS handshake details, raw TCP metrics) require runtime capability and MAY force runtime selection toward agent or embedded runtimes.

## 15.4 Workflow-Wide Diagnostics

A workflow may declare a `diagnostics` block at workflow scope that applies to all steps unless overridden at step scope.

-----

# 16. Composite Steps and Sub-Workflows

A `composite` step inlines another workflow.

## 16.1 Structure

```text
composite
├── workflowRef          (required, workflow id within the same traCtlSpec document)
├── inputs               (variable bindings into the sub-workflow)
├── outputs              (extract bindings from the sub-workflow)
└── isolation            (shared | isolated, default: isolated)
```

## 16.2 Semantics

- `isolated` (default): the sub-workflow gets its own variable scope; only declared `inputs`/`outputs` cross the boundary.
- `shared`: the sub-workflow shares the parent's workflow scope.
- Composite steps participate in DAG dependencies like any other step.
- Composite steps MUST NOT introduce cycles, including transitively.

## 16.3 Planner Cycle Validation for Composites

Planners MUST validate composite cycles via recursive DAG flattening before execution begins.

Required procedure:

1. For each composite step encountered, resolve its `workflowRef` to the referenced workflow.
2. Recursively resolve any composite steps within that workflow (depth-first), producing a fully-flattened step graph.
3. Graft the flattened sub-workflow steps into the parent DAG as a node group, preserving all inter-step edges.
4. Run cycle detection on the fully-flattened graph.

A workflow that references itself directly or transitively (at any depth) constitutes a cycle and MUST be rejected at planning time. The rejection diagnostic MUST identify the full reference path that forms the cycle (e.g., `workflowA → composite:workflowB → composite:workflowA`).

Planners MUST NOT defer this validation to execution time. Composite cycle detection is a planning-time obligation.

-----

# 17. Canonical JSON Projection

When traCtlSpec is serialized (MCP responses, tooling export, on-disk caches), it uses a **canonical JSON projection** with the following rules:

- object keys sorted lexicographically
- arrays preserve user-authored order
- `_ulid` fields included
- no trailing whitespace
- UTF-8 encoding
- numbers as JSON numbers (no quoted numerics)
- durations as ISO 8601 strings (e.g., `PT5S`)
- enum values as lowercase strings

The canonical projection is the **stable interchange format** for traCtlSpec across tools. It is **not** the authoring format — authoring uses TOON/YAML/JSON syntactic representations that normalize into traCtlSpec.

-----

# 18. Conformance Rules

## 18.1 Document Conformance

A document is a conformant traCtlSpec instance if and only if:

1. It declares `schemaVersion` and `capabilities`.
1. All required fields per §5 are present.
1. All IDs satisfy §4.1 rules.
1. All `dependsOn` references resolve.
1. The dependency graph (including implicit dependencies) is acyclic.
1. All `extensions` declared are available at planning time.
1. All declared capability contracts resolve against the selected runtime by contract compatibility.
1. All variable references resolve under at least one environment.
1. All schema references resolve.
1. No reserved identifier prefixes (§4.1) are used outside the canonical namespace.

## 18.2 Producer Conformance

A producer (parser, source adapter, overlay merger) is conformant if:

- It emits documents satisfying §18.1.
- It populates `metadata.sourceFormat` and `metadata.sourceRef`.
- It populates `metadata.overlayRefs` and `metadata.provenance` when overlays were applied.
- It assigns `_ulid` values to every identifiable entity.
- It preserves user-authored order in arrays.
- It performs implicit-dependency detection and merges implicit edges into the canonical DAG (§12.2).
- It emits extension and provider capability requirements in `identifier@version` form.
- It completes overlay resolution before emitting canonical traCtlSpec (§2).

## 18.3 Consumer Conformance

A consumer (canonical validator, planner, executor, reporter) is conformant if:

- It rejects non-conformant documents with explicit error reasons.
- **Canonical validator** rejects unresolved cross-references, unresolved `dependsOn` references, and cyclic dependency graphs per §6.2 and §12.1.
- **Planner** resolves capabilities by contract compatibility (§3.2, §3.3).
- **Planner** rejects unknown capability contracts and incompatible capability versions.
- **Planner** performs composite cycle detection via recursive flattening (§16.3).
- It produces deterministic execution under fixed seeds and concurrency.
- It applies failure policy per §6.3.
- It honors `severity` rules per §11.4.
- It ignores `metadata` for execution decisions (§5.1).

## 18.4 Forward Compatibility

In schema v1, all declared capabilities are mandatory requirements.

Behavior:

- unknown capability contract → reject with explicit unsupported capability error
- incompatible capability version → reject with explicit version compatibility error

Optional capability semantics are reserved for a future schema version.

-----

# 19. Out of Scope (v1)

The following are explicitly deferred:

- overlay file format and merge algorithm (separate Overlay Specification)
- extension SDK wire contract (separate Extension SDK Architecture)
- MCP tool schemas (separate MCP Architecture)
- detailed AI provider contracts (separate AI Architecture)
- scheduler internals (separate Runtime and Concurrency Architecture)
- TOON authoring syntax (separate TOON Specification)
- secret resolution wire formats (separate Security Architecture)

-----

# 20. Glossary

|Term                     |Meaning                                                             |
|-------------------------|--------------------------------------------------------------------|
|traCtlSpec               |Canonical semantic model of traCtl.                                 |
|Capability contract      |Versioned semantic definition of a capability, resolved by contract compatibility.|
|Platform capability      |Engine-owned capability, versioned with the engine, referenced as a stable flag.|
|Extension/provider capability|Capability owned by an extension, provider adapter, or MCP adapter, carrying a mandatory semantic version in `identifier@version` form.|
|Workflow                 |Ordered, dependency-aware step collection sharing context.          |
|Step                     |Atomic execution unit.                                              |
|Extract                  |Named value pulled from a step result into context.                 |
|Assertion                |Validation rule evaluated against a step result.                    |
|Overlay                  |Augmentation layer applied to interoperability sources.             |
|Composite step           |Step that inlines another workflow.                                 |
|Context                  |Runtime substrate of variables, request/response data, and metadata.|
|Frozen context           |Immutable context snapshot delivered to scripts.                    |
|Mutation set             |Declarative set of context changes returned by a script.            |
|Canonical JSON projection|Stable interchange serialization of traCtlSpec.                     |
|Extension Platform       |The single extensibility architecture governing all providers.      |
|Provider                 |A typed extension implementing a provider contract (Provider ⊂ Extension).|
|extensionKind            |Canonical type declaration in the extension manifest taxonomy.      |
