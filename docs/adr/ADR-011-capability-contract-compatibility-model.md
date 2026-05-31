# ADR-011 — Capability Contract & Compatibility Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl resolves capabilities from multiple architectural sources: native extensions (ADR-004), MCP integrations (ADR-005), and provider adapters (ADR-007). Each of these ADRs specifies that capabilities must unify at the planner boundary under a single contract model.

Without a unified capability contract, capability matching degrades to identity-based label matching, semantic inconsistency emerges across sources, and deployment environment drift becomes undetectable. This ADR defines the capability contract model that ADR-004, ADR-005, and ADR-007 depend on.

---

## Decision

### 1 — Capabilities Are Versioned Semantic Contracts

Capabilities are versioned semantic contracts. They are not symbolic labels.

Capability identity consists of:

- capability identifier
- semantic contract version
- declared inputs (schema)
- declared outputs (schema)
- behavioral guarantees
- execution constraints
- compatibility metadata

Canonical format:

```
{identifier}@{version}
```

Examples:

```
structured_output@1
tool_calling@2
streaming@1
http_request@1
embeddings@1
```

Name-only capability matching is prohibited. Version is mandatory for resolution.

Rationale: `structured_output` in one provider generation enforces strict JSON schema. In another it is best-effort. `tool_calling` has meaningfully different parameter shapes across providers and versions. Name-only matching reintroduces the vendor-specific behavior that ADR-007 decision 1 prohibits, hidden one level down. Semantic versioning makes the contract explicit.

### 2 — Unified Capability Source Model

Capability sources MUST normalize into the same canonical capability contract model regardless of origin.

Sources:

- native extensions
- MCP adapters
- provider adapters
- core runtime capabilities

At the planner boundary:

```
native extension capability == MCP adapter capability == provider adapter capability == core capability
```

The planner does not distinguish capability source. It matches against capability contracts.

### 3 — Capability Compatibility Rules

Planning-time compatibility validation MUST evaluate:

- **Presence:** required capability exists in the execution environment
- **Version compatibility:** declared version satisfies workflow version requirement
- **Semantic compatibility:** behavioral guarantees satisfy workflow requirements
- **Input/output compatibility:** declared schemas satisfy workflow usage
- **Execution constraint compatibility:** execution constraints do not conflict with workflow constraints

All five dimensions must pass. Partial compatibility is not compatibility.

**Version compatibility model:**

Version compatibility follows semantic versioning principles:

- patch versions: backward compatible behavioral fixes
- minor versions: backward compatible capability additions
- major versions: breaking contract changes

A workflow requiring `tool_calling@2` is satisfied by `tool_calling@2.x.x`. It is not satisfied by `tool_calling@1.x.x` or `tool_calling@3.x.x` without explicit compatibility declaration.

### 4 — Planning Failure Semantics

Unknown capability contracts MUST fail planning.

Incompatible capability versions MUST fail planning.

Silent substitution is prohibited.

Best-effort capability guessing is prohibited.

Planning failures MUST produce explicit diagnostics identifying:

- the required capability contract
- the available capability contracts in the environment
- the specific compatibility dimension that failed

### 5 — Deployment Consistency

Capability version compatibility MUST be validated against the execution environment at planning time.

Workflow execution in a target environment requires all required capabilities to be present with compatible contracts.

Version skew between authoring and execution environments MUST be explicitly detected.

Environment mismatch MUST fail planning with an explicit diagnostic.

This applies across all deployment topologies (ADR-009):

- local developer environment
- CI environment
- self-hosted environment
- cloud-managed environment

Rationale: The local-versus-CI capability skew scenario is the primary trust failure for traCtl's primary persona. Catching it at planning time — not at execution time — is what makes the core product promise credible.

### 6 — Capability Registration and Conflict Detection

Two sources declaring the same capability identifier and version with incompatible contracts MUST fail at capability registration time.

Conflict detection occurs during extension/provider registration, not at planning time.

Conflict diagnostic MUST identify both conflicting sources and the dimension of incompatibility.

This is consistent with ADR-004 decision 6 (extension lifecycle compatibility check stage).

---

## Consequences

**Positive:**

- Deterministic capability resolution across all sources
- Deployment consistency enforcement
- Provider neutrality preservation (no name-based vendor matching)
- Unified extension, MCP, and provider capability model
- Planning-time failure instead of execution-time surprise
- Explicit conflict detection at registration

**Tradeoffs:**

- Capability contract governance complexity
- Version compatibility maintenance burden across provider/extension ecosystem
- Registration-time conflict detection adds startup complexity

---

## Alternatives Considered

**Capability name-only matching:** rejected. Semantic ambiguity. Hidden vendor-specific behavior. Violates ADR-007.

**Provider-specific capability interpretation:** rejected. Vendor coupling in planner. Violates ADR-007 decision 1.

**Runtime best-effort substitution:** rejected. Nondeterminism. Silent degradation. Violates ADR-003.

**Per-source capability models:** rejected. Three parallel capability systems. Unified planner model impossible.

**Runtime capability discovery:** rejected. Nondeterministic planning. Violates ADR-003 and ADR-004 decision 4.
