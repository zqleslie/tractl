# ADR-004 — Extension Architecture Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl is architected as a platform from day one. Platform extensibility requires a governed model for external capability contribution. Without architectural constraints, extensions accumulate implicit runtime authority, introduce nondeterminism, create security vulnerabilities, and undermine the deterministic execution guarantees in ADR-003.

---

## Decision

### 1 — Extension Execution Model

Extensions are capability providers. They are not runtime peers.

Extensions:

- provide declared capabilities (connectors, actions, transformers, validators, custom execution handlers)
- execute within bounded capability contracts

Extensions MUST NOT:

- alter orchestration semantics
- mutate scheduler behavior
- override canonical execution lifecycle stages
- access runtime internals directly

The runtime core remains architecture-governed and isolated from extension implementation concerns.

Rationale: Plugin systems fail when plugins acquire implicit authority over orchestration. Prohibiting this at the architecture level prevents future pressure toward special-case exceptions.

### 2 — Extension Contract Model

Every extension MUST provide:

**Manifest (governance contract):**

- extension identifier
- version
- capability declarations
- permission requirements
- compatibility metadata
- runtime requirements

**SDK contract (execution contract):**

- declared input schema
- declared output schema
- behavioral guarantees
- execution constraints

The manifest governs authorization policy. The SDK contract governs execution behavior. These are separate concerns and evolve independently.

### 3 — Trust Model

Extensions are untrusted by default.

First-party extensions are untrusted by default unless explicitly granted elevated privilege through governance policy.

This applies even to extensions authored by the traCtl core team.

Rationale: Implicit trust for first-party plugins is how privilege escalation vulnerabilities become permanent architectural features. Default untrusted prevents this by construction.

### 4 — Capability Resolution Stage

Capability dependency resolution occurs at planning time.

Runtime capability discovery is prohibited.

A workflow referencing a missing or incompatible extension capability MUST fail at planning, not during execution.

This preserves determinism (ADR-003) when extensions are involved and provides clear developer feedback before workflow execution begins.

### 5 — Extension Isolation

Extensions execute inside bounded execution environments.

Extensions MUST receive only inputs declared in their capability contract.

Extensions MUST return only outputs declared in their capability contract.

Extensions MUST NOT access:

- scheduler internals
- canonical workflow mutation paths
- unrelated execution state
- secrets not explicitly scoped to their capability
- other extensions' execution state

Data flow isolation is the architectural invariant. Process isolation is an implementation detail that may or may not accompany it.

### 6 — Extension Lifecycle

Extension lifecycle stages:

```
install
  → register
  → validate
  → compatibility check
  → resolve
  → authorize
  → execute
  → observe
  → retire
```

**Compatibility check** (between validate and resolve): verifies that declared capabilities are compatible with the current spec version and do not conflict with other registered extensions. Two extensions declaring the same capability with incompatible semantics MUST fail at registration time, not at planning time.

Each stage is a governed boundary. Stage failures are explicit and observable.

### 7 — Extension Participation in Overlay Processing

Extensions MAY participate in overlay processing through:

- overlay target matching (extension-defined selector schemas)
- overlay merge processing (extension-defined merge semantics for extension-owned fields)
- source selector provenance (extension-defined source adapter targeting)

Any extension participating in overlay processing MUST be:

- deterministic — identical inputs MUST produce identical outputs across invocations
- pure — no side effects beyond the declared overlay contribution
- side-effect free — no I/O, no global state mutation
- reproducible — repeated invocation across runs MUST produce identical results
- capability declared — overlay participation surface MUST be declared as a capability contract per ADR-011
- validation compatible — overlay contributions MUST pass canonical validation after merge
- security governed — subject to the trust constraints in ADR-006

Network access, external mutation, and nondeterministic evaluation during overlay processing are prohibited.

Rationale: Overlay resolution sits in the pre-execution transformation pipeline. Nondeterministic overlay behavior breaks the canonical equivalence guarantees in ADR-002 §3a and the deterministic execution guarantees in ADR-003. Extension-defined overlay behavior MUST NOT compromise deterministic canonical generation.

---

## Consequences

**Positive:**

- Stable runtime core independent of extension evolution
- Governed extensibility surface
- Deterministic capability resolution
- Security boundary enforcement
- Platform evolution without core architectural rework

**Tradeoffs:**

- Extension authoring complexity relative to uncontrolled plugins
- Manifest and SDK governance overhead
- Compatibility check complexity increases with ecosystem size

---

## Alternatives Considered

**Unrestricted plugin access:** rejected. Plugin chaos. Nondeterminism. Security vulnerabilities.

**Extensions as runtime peers:** rejected. Orchestration authority leaks. Core invariants become unenforceable.

**Runtime capability discovery:** rejected. Violates ADR-003 determinism. Execution failures replace planning failures.

**Implicit first-party trust:** rejected. Structural privilege escalation risk.
