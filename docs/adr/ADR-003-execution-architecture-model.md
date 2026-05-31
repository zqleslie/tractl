# ADR-003 — Execution Architecture Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl requires:

- deterministic workflow orchestration
- dependency-aware parallel execution
- retry and cancellation support
- timeout enforcement
- provider-independent execution
- observable execution lifecycle
- LLM integration without nondeterminism in control flow

A pure DAG execution model is too restrictive for streaming and retry semantics. A pure actor model is operationally opaque. A generic workflow engine is too vague to enforce determinism guarantees. A hybrid event-driven orchestration model satisfies all constraints.

---

## Decision

### 1 — Execution Architecture

traCtl uses hybrid event-driven workflow orchestration.

The execution model is neither pure DAG nor pure actor. It is orchestrated, state-managed, event-driven execution with a deterministic control plane.

### 2 — Canonical Pipeline

Execution MUST follow this canonical pipeline:

```
Input
  → Normalization
  → Validation
  → Planning
  → Compilation
  → Orchestration
  → Provider/Extension Execution
  → State Transitions
  → Outputs
```

No stage may be bypassed.

No direct parser-to-runtime execution path exists.

### 3 — Planner and Compiler Are Separate Concerns

**Planner responsibilities:**

- dependency analysis
- execution strategy determination
- resource decisions
- task ordering
- capability analysis
- runtime selection

**Compiler responsibilities:**

- converting canonical execution plan into deterministic runtime artifact
- generating executable plan from planner output

The planner's output MUST be human-readable enough to surface in diagnostics. This enables the "what will this workflow do" inspection surface without requiring execution. The compiler converts planner output into the runtime artifact.

### 4 — Determinism

Control flow MUST be deterministic.

LLM outputs MAY be probabilistic. This distinction is critical.

```
Same workflow structure → same orchestration behavior
```

Even if model text responses differ across executions, the control flow path — which nodes execute, in what dependency order, under what failure semantics — must be identical given the same workflow definition.

Determinism applies to:

- execution graph structure
- node ordering within dependency constraints
- failure propagation behavior
- cancellation propagation
- retry sequencing

Determinism does not require:

- identical response content from probabilistic providers
- identical timing
- identical interleaving of concurrent branches

### 5 — Direct Execution Prohibition

No direct execution from parsed representation is permitted.

Everything traverses the canonical execution pipeline.

This is an absolute guardrail with no exceptions.

---

## Consequences

**Positive:**

- Deterministic lifecycle control
- Retry and cancellation support
- Observability boundaries at each pipeline stage
- Provider and execution abstraction
- LLM integration without control flow nondeterminism

**Tradeoffs:**

- Increased architectural complexity relative to direct interpreters
- Orchestration overhead
- Pipeline stage maintenance

---

## Alternatives Considered

**Direct interpreter execution:** rejected. Nondeterministic. Untestable lifecycle guarantees. Provider coupling.

**Provider-native orchestration:** rejected. Vendor coupling. No unified execution contract.

**Pure DAG execution:** rejected. Too restrictive for streaming, retries, and cancellation propagation.

**Pure actor model:** rejected. Operationally opaque. Difficult to enforce determinism guarantees.

**Generic workflow engine:** rejected. Insufficient determinism enforcement for core product promise.

---

## Amendment — 2026-05-24

Amended by: ADR-012 (Multi-Surface Delivery Model), ADR-013 (Traffic Acquisition & Capture Architecture), ADR-014 (Observability & Traffic Diagnostics Model)

### Context

ADR-012 formalizes five delivery surfaces and establishes that the execution engine must be decoupled from all surface UX concerns. ADR-013 introduces the Acquisition Layer as a pre-normalization ingestion boundary upstream of the canonical pipeline. ADR-014 defines the observability model with explicit hooks at each pipeline stage.

The canonical pipeline in §2 of this ADR correctly defines the normalization-to-output flow but does not account for:

- the Acquisition Layer upstream of normalization
- multi-surface execution engine boundary
- observability instrumentation points at each stage
- the distinction between execution-context diagnostics and surface-level diagnostics

### Additional Decisions

#### 6 — Expanded Canonical Pipeline

The canonical pipeline (§2) is expanded to include the Acquisition Layer and explicit observability instrumentation points.

Full canonical pipeline:

```
Execution Source / Capture Source
        ↓
[Acquisition Layer] (ADR-013 §1)
  Capture Adapter → Capture Record → Capture Normalization
        ↓
traCtlSpec (with provenance if acquisition source)
        ↓
Canonical Validation
        ↓
Overlay Engine
        ↓
Canonical Workflow Specification (post-overlay)
        ↓
Execution Planner
  [Observability: planning trace]
        ↓
DAG Compiler
        ↓
Runtime Layer
  [Observability: Request Timeline per step (ADR-014 §2)]
  [Observability: Retry events (ADR-014 §3)]
  [Observability: Redirect events (ADR-014 §3)]
        ↓
Assertions + Extracts
  [Observability: assertion + extract evaluation events (ADR-014 §5)]
        ↓
Diagnostics + Reporting
  [Observability: Execution Provenance Trace (ADR-014 §5)]
  [Observability: Dependency Waterfall (ADR-014 §4)]
```

The invariants of §2 are unchanged:

- no stage may be bypassed
- no direct parser-to-runtime path exists
- the Acquisition Layer path is additive, not a bypass of normalization

Acquisition-sourced workflows enter the pipeline at the `traCtlSpec` boundary, after Capture Normalization (ADR-013 §4). They do not bypass canonical validation.

#### 7 — Execution Engine Surface Boundary

The execution engine MUST expose a surface-neutral interface.

The execution engine MUST NOT:

- reference surface-specific state (Desktop UX state, browser extension state)
- assume interactive execution context
- embed surface-specific rendering or output formatting

All five delivery surfaces (ADR-012 §1) consume the execution engine through this surface-neutral interface.

Surface-specific output rendering occurs after the execution engine emits its result.

The canonical execution result is the authoritative record.

Surface presentation adapts the canonical result; it does not produce an alternative result.

#### 8 — Observability Instrumentation Points

Observability instrumentation (ADR-014) is embedded at the following canonical pipeline positions.

These are not optional additions. They are part of the pipeline execution contract.

| Pipeline Stage | Instrumentation |
|---|---|
| Acquisition Layer | Provenance record (ADR-013 §5) |
| Planner | Plan trace event |
| Runtime — request dispatch | Request Timeline start (ADR-014 §2) |
| Runtime — response receipt | Request Timeline end (ADR-014 §2) |
| Runtime — retry decision | Retry event (ADR-014 §3) |
| Runtime — redirect | Redirect hop event (ADR-014 §3) |
| Assertions | Assertion evaluation event (ADR-014 §5) |
| Extracts | Extract evaluation event (ADR-014 §5) |
| Workflow completion | Execution Provenance Trace (ADR-014 §5), Dependency Waterfall (ADR-014 §4) |

Instrumentation is credential-safe by default (ADR-014 §6).

Instrumentation MUST NOT alter execution semantics.

Observability is a read-only side-effect of execution, not a control-flow participant.
