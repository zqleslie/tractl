# ADR-007 — Provider Abstraction Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl integrates with LLM providers and execution providers. The AI provider landscape changes rapidly — new vendors, new capability profiles, new pricing models, deprecations. Without explicit provider abstraction boundaries, vendor-specific semantics leak into the planner, compiler, and canonical spec, creating coupling that is expensive to reverse.

---

## Decision

### 1 — Provider Neutrality

The core runtime is provider-neutral.

Provider-specific semantics MUST remain inside provider adapters.

Provider-specific behavior MUST NOT appear in:

- the planner
- the compiler
- the canonical spec (traCtlSpec)
- orchestration logic

Rationale: The AI provider landscape evolves fast enough that any vendor coupling baked into core becomes technical debt within a product generation. Provider neutrality is an architectural durability requirement.

### 2 — Capability-Based Provider Matching

The planner matches providers on declared capability contracts.

Not on vendor identity.

Correct form:

```
requires: structured_output@1
```

Prohibited form:

```
if provider == "openai"
if provider == "anthropic"
```

Providers declare the capabilities they support. The planner resolves providers that satisfy required capability contracts. Vendor names are invisible to the planner.

This is consistent with ADR-004 (extension capability model) and ADR-005 (MCP capability normalization). All three sources unify at the planner boundary under the same capability contract model (ADR-011).

### 3 — Provider Failure Model

Provider failures are execution failures. They are not architectural exceptions.

Provider failure categories handled through canonical execution semantics (ADR-003, ADR-008):

- timeout
- quota exceeded
- authentication failure
- malformed response
- degraded capability
- rate limiting

Provider failures propagate through DAG scheduler semantics:

- dependent branches skip
- independent branches continue

No special-case error paths exist for provider failures.

### 4 — Planning-Time Capability Mismatch

Provider capability mismatches MUST be detected at planning time where statically determinable.

Example: workflow requires `structured_output@1`, selected provider declares `structured_output` unsupported → planning MUST fail with explicit diagnostic.

Dynamic runtime failures remain possible for conditions not statically determinable at planning time (quota exceeded, transient auth failure). These are handled through execution failure semantics.

Silent capability substitution is prohibited.

### 5 — Provider Output Trust

Provider output is untrusted input.

This applies to all providers regardless of vendor trust level.

Provider output MUST NOT directly mutate:

- control flow
- canonical workflow state
- orchestration decisions

This is consistent with ADR-001 (prompt safety boundary) and ADR-006 (provider output trust). The trust model is uniform across prompt output and provider output.

### 6 — Capability Semantic Versioning

Provider capabilities carry semantic versions as part of their contract identity (ADR-011).

Capability name matching without version is prohibited.

Rationale: `structured_output` in one provider generation may enforce strict JSON schema. In another it may be best-effort. `tool_calling` has meaningfully different shapes across providers and versions. Name-only matching reintroduces the vendor-specific behavior that decision 1 prohibits, one level down.

Capability version compatibility rules are governed by ADR-011.

---

## Consequences

**Positive:**

- Core runtime survives provider landscape changes
- Vendor-neutral capability planning
- Consistent failure handling
- Unified capability model with extensions and MCP
- Future multi-provider routing without architectural change

**Tradeoffs:**

- Provider adapter maintenance per vendor
- Capability contract governance overhead
- Capability version compatibility complexity

---

## Alternatives Considered

**Vendor-specific planner branches:** rejected. Vendor coupling in core. Expensive to reverse.

**Name-only capability matching:** rejected. Semantic ambiguity. Hidden vendor-specific behavior. Addressed in ADR-011.

**Provider-native orchestration:** rejected. Execution lifecycle authority outside core. Determinism unenforceable.

**Runtime best-effort capability substitution:** rejected. Nondeterminism. Silent degradation.

---

## Amendment — 2026-05-24

Amended by: ADR-013 (Traffic Acquisition & Capture Architecture)

### Context

ADR-007 defines the provider abstraction model in terms of LLM providers and API execution providers. ADR-013 introduces Capture Adapters as a distinct architectural category. Without explicit categorization, Capture Adapters risk being conflated with Providers, which would incorrectly couple capture semantics to the provider capability model.

This amendment clarifies the three distinct categories of components that interact with the traCtl execution boundary.

### Additional Decisions

#### 7 — Provider Category Clarification

traCtl has three distinct architectural component categories at the platform boundary. These are NOT interchangeable terms.

**Execution Providers**

Deliver execution capabilities to the runtime.

Examples: HTTP runtime, LLM provider adapters, future protocol runtimes.

Governed by: ADR-007 (this ADR), ADR-004, ADR-011.

Capability model: declared capability contracts, capability-based matching by planner.

Lifecycle: installable, hot-loadable via provider host.

**Compatibility Providers**

Translate external ecosystem artifact formats into traCtlSpec.

Examples: OpenAPI provider, Postman provider, Bruno provider, WSDL provider.

Governed by: ADR-001 (universal normalization), ADR-007 (this ADR).

Capability model: declared source format → traCtlSpec normalization contracts.

Lifecycle: installable, hot-loadable via provider host.

**Capture Adapters**

Ingest live traffic from capture sources into the Acquisition Layer.

Examples: Browser Interception Adapter, Local Proxy Adapter, HAR Import Adapter.

Governed by: ADR-013.

Capability model: NOT provider capability contracts. Capture Adapters operate under the Acquisition Layer boundary, upstream of the canonical execution pipeline.

Lifecycle: bundled with surface delivery (not independently installable as providers).

**Critical distinction:**

Capture Adapters MUST NOT be implemented as Execution Providers or Compatibility Providers.

Capture Adapters operate at the acquisition boundary, before normalization.

Providers operate at or after the normalization boundary.

Conflating these categories would allow capture logic to influence execution semantics, which is prohibited.

Rationale: Naming precision at architectural category boundaries prevents category confusion from propagating into implementation. The three categories have different lifecycle models, different trust treatment (ADR-006), and different positions in the canonical pipeline (ADR-003). They must remain explicitly distinct.
