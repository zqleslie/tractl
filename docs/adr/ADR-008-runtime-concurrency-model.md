# ADR-008 — Runtime Concurrency Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl executes dependency-aware parallel workflows. Concurrency introduces ordering nondeterminism, resource contention, cancellation complexity, and timeout interaction hazards. Without explicit concurrency semantics, retry logic, timeout handling, and failure propagation become inconsistent across the execution surface. The core product promise — same workflow, same behavior everywhere — requires concurrency to be architecturally governed, not implementation-specific.

---

## Decision

### 1 — Structured Concurrency

Structured concurrency is mandatory.

Every spawned execution unit belongs to a parent lifecycle.

Orphan work is prohibited by construction.

Implications:

- every task has an owner
- task lifetime is bounded by parent lifetime
- parent completion waits for all children
- parent cancellation propagates to all children

Rationale: Orphan work makes concurrency bugs undebugable. Eliminating it by construction is more reliable than detecting it at runtime.

### 2 — Cancellation Propagation

Cancellation propagates downward deterministically through the structured concurrency tree.

```
Parent workflow cancelled → child tasks cancelled → grandchild tasks cancelled
```

No zombie branches survive parent cancellation.

Cancellation propagation is deterministic and exhaustive.

### 3 — Timeout Semantics

Timeouts are explicit execution semantics, not implementation hints.

Timeout scopes:

- workflow-level timeout
- stage-level timeout
- capability-level timeout
- provider invocation timeout
- MCP invocation timeout
- extension invocation timeout

**Timeout composition rule:**

The tightest applicable bound wins.

When multiple timeout scopes apply to a single operation, the smallest remaining budget across all applicable scopes governs.

Example: a provider invocation with a 30-second provider timeout executing inside a stage with 10 seconds remaining budget is governed by the 10-second stage budget.

Nested timeout behavior is deterministic. Operations MUST NOT exceed the tightest applicable bound regardless of which scope declares a longer duration.

### 4 — Retry Semantics

Retries are orchestration policy. They are not provider behavior. They are not extension behavior.

Extensions MUST NOT implement internal retry logic.

Providers MUST NOT implement internal retry logic.

All retry behavior is declared in workflow definitions and enforced by the orchestration layer.

Rationale: Distributed retry logic produces retry storms, nondeterministic retry counts, and timeout interactions that no single component can reason about. Centralizing retry in orchestration makes it deterministic, observable, and policy-governed.

**Retry and timeout interaction:**

Retry policy MUST be evaluated against remaining timeout budget. A retry MUST NOT be attempted if the remaining budget is insufficient to complete a minimum retry attempt. This interaction is the responsibility of the orchestration layer, not individual providers or extensions.

### 5 — Parallel Branch Isolation

Parallel branch failures are isolated unless dependency semantics require propagation.

```
Branch A fails → Branch B (independent) continues
Branch A fails → Branch C (depends on A) skips
```

This is consistent with DAG scheduler semantics (ADR-003), MCP failure semantics (ADR-005), and provider failure semantics (ADR-007).

The failure isolation boundary is the dependency edge, not the workflow boundary.

### 6 — Resource Governance

The runtime governs execution resources:

- concurrency limits
- queue pressure
- memory pressure
- provider backpressure
- extension resource contention

Resource governance is runtime-owned. Extensions and providers MUST NOT self-govern resource consumption beyond declared capability constraints.

### 7 — Execution Semantic Determinism

Concurrency MUST NOT alter workflow semantic meaning.

Formal definition:

Given the same workflow definition and the same provider/capability outputs, any valid execution schedule MUST produce:

- the same final workflow state
- the same set of assertion results
- the same set of observable outputs

Timing differences between execution schedules are permitted.

Semantic drift between execution schedules is prohibited.

This is a testable invariant: for any workflow W and output set O, all valid concurrent schedules of W with outputs O must produce identical semantic results.

---

## Consequences

**Positive:**

- Debuggable concurrency — no orphan work
- Deterministic cancellation and timeout behavior
- Consistent retry semantics across all execution surfaces
- Predictable failure isolation
- Testable semantic determinism

**Tradeoffs:**

- Structured concurrency constraints limit some concurrency optimization patterns
- Centralized retry orchestration requires orchestration layer sophistication
- Timeout composition adds implementation complexity

---

## Alternatives Considered

**Extension-owned retry:** rejected. Nondeterministic. Retry storms. Timeout interaction hazards.

**Provider-owned cancellation:** rejected. Cancellation propagation becomes unreliable. Zombie work survives.

**Best-effort timeout semantics:** rejected. Violates determinism guarantee. Budget overruns become possible.

**Global fail-fast workflow semantics:** rejected. A policy that cancels all remaining workflow steps on any single failure is explicitly avoided. Independent branches MUST continue on unrelated failure (§5). This rejection applies to *global* cancellation semantics only.

The `failFast` failure policy in `tractl_spec.md` §6.3 is NOT a global fail-fast and is NOT rejected by this ADR. `failFast` cancels only the transitive dependency subgraph of the failed step; steps with no dependency path to the failed step are unaffected and continue executing. This scoped cancellation is fully compatible with the branch-isolation guarantee in §5. `failFast` is an opt-in policy; the default execution policy is `resilient` (passive scheduling-time skipping of dependent steps, no active cancellation signal).
