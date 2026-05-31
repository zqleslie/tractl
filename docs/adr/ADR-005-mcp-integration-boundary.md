# ADR-005 — MCP Integration Boundary

Status: Accepted
Date: 2026-05-23

---

## Context

Model Context Protocol (MCP) is an emerging standard for AI agent tool and resource interoperability. As the MCP ecosystem grows, architectural pressure exists to allow MCP to become a first-class execution substrate rather than an adapter layer. This ADR establishes the boundary explicitly to prevent drift.

---

## Decision

### 1 — MCP Role Classification

MCP is an integration protocol. It is not the core runtime execution architecture.

traCtl does NOT become "an MCP workflow engine."

MCP is an adapter layer for interoperability.

This boundary will face pressure as the MCP ecosystem grows. Every proposal to "let the MCP server handle this directly" must be evaluated against this decision. The boundary is frozen.

### 2 — Permitted MCP Roles

MCP integrations are permitted for:

- tool exposure
- resource access
- external capability access
- agent interoperability

MCP integrations are NOT permitted to exercise:

- execution lifecycle authority
- orchestration control
- canonical workflow semantic definition
- planner authority
- scheduler authority

### 3 — Core Runtime Independence

The core runtime MUST function fully without MCP.

MCP is not a required dependency for any core execution capability.

MCP unavailability, protocol changes, or server misbehavior MUST NOT affect core execution behavior.

This prevents architectural dependency inversion — where an external protocol becomes load-bearing for core functionality.

### 4 — Capability Normalization

MCP-exposed capabilities MUST normalize into the same canonical capability contract model as native extensions and provider adapters.

At the planner boundary:

```
native extension capability == MCP adapter capability == provider adapter capability
```

The planner does not distinguish capability source. It matches against capability contracts (ADR-011).

This enables uniform capability conflict detection, version compatibility validation, and planning-time resolution regardless of whether a capability originates from a native extension or an MCP adapter.

### 5 — MCP Failure Semantics

MCP capability failures during execution are handled through canonical execution failure semantics (ADR-003, ADR-008).

MCP server unavailability during planning MUST fail planning deterministically with an explicit diagnostic.

MCP capability resolution at planning time MUST be subject to timeout policy. Planning MUST NOT block indefinitely on MCP discovery.

Timed-out MCP capability resolution MUST fail planning, not degrade silently.

MCP runtime failures during execution propagate through DAG scheduler semantics:

- dependent branches skip
- independent branches continue

---

## Consequences

**Positive:**

- Clean architectural boundary as MCP ecosystem evolves
- Core runtime stability independent of MCP protocol changes
- Unified capability model across native and MCP sources
- Deterministic failure behavior for MCP integrations

**Tradeoffs:**

- MCP capabilities require normalization layer (consistent with ADR-001 philosophy)
- Capability contract governance required for MCP adapters

---

## Alternatives Considered

**MCP as execution substrate:** rejected. Core runtime becomes dependent on external protocol. Vendor coupling at architecture level.

**MCP-native orchestration:** rejected. Execution lifecycle authority leaks outside core. Determinism guarantees become unenforceable.

**Separate capability model for MCP:** rejected. Two parallel capability systems with ad hoc integration points. Breaks unified planner model.

**Best-effort MCP discovery without timeout:** rejected. Non-deterministic planning. Violates ADR-003.
