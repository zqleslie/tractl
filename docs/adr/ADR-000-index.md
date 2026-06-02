# traCtl Architecture Decision Records — Index

Status: Normative
Date: 2026-05-23

---

## Purpose

This index is the entry point for the traCtl ADR set.

ADRs are normative architecture artifacts (ADR-010 hierarchy position 1).

Architecture changes require ADR updates. README edits and roadmap notes do not constitute architecture decisions.

---

## Normative Hierarchy

Per ADR-010:

1. **ADRs** ← this document set
2. Canonical architecture specifications
3. Terminology registry (`00_terminology.md`)
4. Schemas and wire contracts
5. HLD
6. PRD
7. Roadmap
8. Onboarding documentation
9. README

Conflicts resolve upward.

---

## ADR Set

| ADR | Title | Status | Key Invariant |
|-----|-------|--------|---------------|
| [ADR-001](ADR-001-canonical-execution-semantic-model.md) | Canonical Execution Semantic Model | Accepted | All inputs normalize into `traCtlSpec` before any execution |
| [ADR-002](ADR-002-representation-serialization-strategy.md) | Representation & Serialization Strategy | Accepted | Semantic model ≠ serialization format; bidirectional parity required |
| [ADR-003](ADR-003-execution-architecture-model.md) | Execution Architecture Model | Accepted | Canonical pipeline mandatory; control flow is deterministic |
| [ADR-004](ADR-004-extension-architecture-model.md) | Extension Architecture Model | Accepted | Extensions are capability providers, not runtime peers |
| [ADR-005](ADR-005-mcp-integration-boundary.md) | MCP Integration Boundary | Accepted | MCP is adapter-layer interoperability, not runtime substrate |
| [ADR-006](ADR-006-security-trust-model.md) | Security Trust Model | Accepted | Zero implicit trust; security is execution admission, not middleware |
| [ADR-007](ADR-007-provider-abstraction-model.md) | Provider Abstraction Model | Accepted | Core runtime is provider-neutral; capability contracts not vendor names |
| [ADR-008](ADR-008-runtime-concurrency-model.md) | Runtime Concurrency Model | Accepted | Structured concurrency mandatory; semantic determinism required |
| [ADR-009](ADR-009-packaging-deployment-model.md) | Packaging & Deployment Model | Accepted | Deployment topology must not change canonical workflow semantics |
| [ADR-010](ADR-010-documentation-governance-model.md) | Documentation Governance Model | Accepted | Normative hierarchy with upward conflict resolution |
| [ADR-011](ADR-011-capability-contract-compatibility-model.md) | Capability Contract & Compatibility Model | Accepted | Capabilities are versioned semantic contracts, not symbolic labels |
| [ADR-012](ADR-012-multi-surface-delivery-model.md) | Multi-Surface Delivery Model | Accepted | Same workflow definition, same execution semantics across all surfaces |
| [ADR-013](ADR-013-traffic-acquisition-capture-architecture.md) | Traffic Acquisition & Capture Architecture | Accepted | Traffic capture is acquisition-layer concern, not execution-layer |
| [ADR-014](ADR-014-observability-traffic-diagnostics-model.md) | Observability & Traffic Diagnostics Model | Accepted | Diagnostics are emitted as events; collection decoupled from execution |
| [ADR-015](ADR-015-packaging-distribution-strategy.md) | Packaging & Distribution Strategy | Accepted | Single binary per surface; no runtime dependencies |
| [ADR-016](ADR-016-ui-architecture-model.md) | UI Architecture Model | Accepted | Shared React in frontend/; transport-agnostic components; surface adapters in platform/ |
| [ADR-017](ADR-017-request-dto-and-protocol-surface-contract.md) | Request DTO & Protocol Surface Contract | Accepted | RequestDef is the public API contract for single-request execution; Protocol field required pre-release |

---

## Core Invariants Cross-Reference

The following invariants are load-bearing across multiple ADRs.

**Universal normalization (ADR-001):**
All inputs → `traCtlSpec` → execution. No exceptions. Enforced by ADR-006 admission boundary.

**Determinism (ADR-003):**
Same workflow definition → same execution behavior. Protected at normalization (001), planning (003, 004, 007), concurrency (008), and deployment (009).

**Capability-over-identity (ADR-004, 005, 007, 011):**
The planner matches on capability contracts, not vendor names or source identity. Unified at planner boundary.

**Zero implicit trust (ADR-006):**
Applies to extensions (004), providers (007), MCP integrations (005), imported artifacts (006), and prompt output (001).

**Prompt and provider output safety (ADR-001, 006, 007):**
LLM and provider output is untrusted input. Same trust treatment regardless of source.

**Security as admission (ADR-006):**
Security validation is part of execution admission. Not optional middleware. Not bypassable.

**Deployment topology invariant (ADR-009):**
Topology does not change semantics. Local == CI == cloud-managed for workflow behavior.

**Cross-format overlay parity (ADR-002 §3a):**
Semantically equivalent source + overlay combinations normalize identically regardless of authoring format. Format choice MUST NOT introduce semantic difference.

**TOON normative governance decision:**
TOON is frozen as a v1 normative specification alongside the YAML and JSON serialization specifications. The `tractl_toon_spec.md` is a normative governance artifact at hierarchy position 4. TOON is governed under the same freeze and change-control process as the YAML and JSON specs. This decision was made explicitly at the v1 architecture freeze; TOON is not a deferred workstream. Any future changes to TOON normative status require an ADR update.

**Overlay as augmentation, not privilege escalation (ADR-006 §9):**
Overlay MUST NOT bypass schema validation, canonical validation, capability declarations, trust boundaries, secret policy, or planner admission. Extension overlay processors MUST be deterministic, pure, side-effect free, and prohibited from network access or nondeterministic evaluation (ADR-004 §7).

---

## Decision Dependency Graph

```
ADR-001 (canonical model)
    ← ADR-002 (representations normalize into it)
    ← ADR-003 (execution pipeline starts from it)
    ← ADR-006 (security admission enforces it)

ADR-003 (execution pipeline)
    ← ADR-004 (extensions resolve at planning stage)
    ← ADR-005 (MCP failures follow pipeline semantics)
    ← ADR-007 (provider failures follow pipeline semantics)
    ← ADR-008 (concurrency within pipeline)

ADR-011 (capability contracts)
    ← ADR-004 (extensions declare capabilities via this model)
    ← ADR-005 (MCP capabilities normalize into this model)
    ← ADR-007 (provider capabilities declared via this model)
    ← ADR-009 (deployment consistency validated via this model)

ADR-006 (trust model)
    ← ADR-004 (extension trust boundaries)
    ← ADR-005 (MCP trust boundaries)
    ← ADR-007 (provider trust boundaries)

ADR-010 (governance)
    ← governs all other documents
```

---

## Adding New ADRs

New ADRs follow this structure:

```
# ADR-NNN — Title

Status: [Proposed | Accepted | Deprecated | Superseded by ADR-NNN]
Date: YYYY-MM-DD

---

## Context
## Decision
## Consequences
## Alternatives Considered
```

Numbering is sequential. No gaps. No reuse of deprecated numbers.

Status transitions require explicit update to this index.

---

## Index Update — 2026-05-24

### New ADRs (ADR-012 to ADR-015)

| ADR | Title | Status | Key Invariant |
|-----|-------|--------|---------------|
| [ADR-012](ADR-012-multi-surface-delivery-model.md) | Multi-Surface Delivery Model | Accepted | Same execution engine, same semantics, across all five delivery surfaces |
| [ADR-013](ADR-013-traffic-acquisition-capture-architecture.md) | Traffic Acquisition & Capture Architecture | Accepted | All capture sources normalize through Acquisition Layer into traCtlSpec before execution |
| [ADR-014](ADR-014-observability-traffic-diagnostics-model.md) | Observability & Traffic Diagnostics Model | Accepted | Execution-context diagnostics only; credential-safe by default |
| [ADR-015](ADR-015-packaging-distribution-strategy.md) | Packaging & Distribution Strategy | Accepted | Distribution channel must not introduce execution semantic differences |

### Amended ADRs

| ADR | Amendment Summary |
|-----|-------------------|
| [ADR-001](ADR-001-canonical-execution-semantic-model.md) | Acquisition sources added to universal normalization enumeration; provenance field defined |
| [ADR-003](ADR-003-execution-architecture-model.md) | Expanded canonical pipeline with Acquisition Layer and observability instrumentation points; surface-neutral engine boundary defined |
| [ADR-006](ADR-006-security-trust-model.md) | Interception consent model, browser extension trust, localhost IPC rules, certificate constraints, credential masking policy, imported traffic handling, proxy isolation |
| [ADR-007](ADR-007-provider-abstraction-model.md) | Three-category clarification: Execution Providers, Compatibility Providers, Capture Adapters |
| [ADR-009](ADR-009-packaging-deployment-model.md) | Container immutability, provider version locking, CI runner expectations, browser extension deployment implications |

### Updated Core Invariants Cross-Reference

**Acquisition normalization (ADR-013):**
All capture sources → Acquisition Layer → Capture Normalization → traCtlSpec → canonical pipeline. No bypass.

**Multi-surface parity (ADR-012):**
Same workflow definition → same execution semantics across Desktop, Browser Extension, CLI, CI, Container Runtime.

**Observability scope (ADR-014):**
Execution-context diagnostics only. Credential-safe by default. Not a packet analyzer.

**Distribution parity (ADR-015):**
Distribution channel is a delivery concern. Execution semantics do not vary by distribution channel.

---

## Index Update — 2026-05-29

### Amended ADRs

| ADR | Amendment Summary |
|-----|-------------------|
| ADR-009 | §9: Web App browser-local deployment topology added. §10: Corrects the 2026-05-24 decision log. Basic mode needs no backend. Extended mode connects to Docker container or hosted server. Desktop and CLI are not web upgrade paths. |
| ADR-012 | §7: Web App added as sixth delivery surface with two modes. Web App and Desktop serve distinct user personas. §8: Browser-safe / non-browser-safe capability partition defined. §9: CORS clarified as browser-enforced boundary. §10: WASM build artefact defined. §11: Surface table updated. |

### Updated Core Invariants Cross-Reference

**Web App two-mode model (ADR-012 §7):**
Web App basic mode requires no backend. Web App extended mode connects to a Docker
container or hosted server. Desktop and Web App serve distinct user personas and are
not interchangeable surfaces.

**Browser-safe capability partition (ADR-012 §8):**
Browser-safe capabilities are bundled in the Web App WASM and require no backend.
Non-browser-safe capabilities require a Docker container or server connection.
The partition boundary is defined by package and capability contract.

**Web App topology invariant (ADR-009 §9):**
Browser-safe capabilities executed via WASM MUST produce semantically equivalent
results to the same capabilities on Desktop, CLI, CI, or Container surfaces.
Timing may vary. Semantic behavior may not.

**CORS is not a traCtl constraint (ADR-012 §9):**
CORS is a browser-enforced boundary. CORS-enabled APIs work natively in basic mode.
Non-CORS APIs require extended mode via Docker container or hosted server. Planning
MUST surface an explicit diagnostic when a non-browser-safe capability is attempted
without a server connection.

---

## Index Update — 2026-05-29

### Amended ADRs

| ADR | Amendment Summary |
|-----|-------------------|
| [ADR-012](ADR-012-multi-surface-delivery-model.md) | Web Application added as sixth delivery surface (§8). Two-tier model defined: Tier 1 browser-native WASM, Tier 2 server-connected opt-in (§9). Surface category classification updated to include Web Application as interactive surface (§10). Parity guarantees with declared Tier 1 exceptions (§11). Desktop and Web Application independence declared (§12). |
| [ADR-016](ADR-016-ui-architecture-model.md) | §3 engine communication model corrected: three separated paths — Desktop via Wails bindings (no HTTP), Web Tier 1 via WASM, Web Tier 2 via fetch() to connected server. chi HTTP server removed from Desktop binary; lives in cmd/server/ as standalone tractl server binary only. §7 web surface constraint superseded: web surface has no dependency on Desktop application at runtime. Platform abstraction layer defined with three transport paths. |

### Updated Core Invariants Cross-Reference

**Web Application independence (ADR-012 §12, ADR-016 amendment):**
Desktop and Web Application share source code at build time only. No runtime
dependency, shared process, shared storage, or shared session between them.

**Web WASM execution (ADR-012 §9, ADR-016 amendment):**
Web Tier 1 executes the full Go engine compiled to WebAssembly inside the
browser tab. No server required. Browser sandbox limitations are declared
exceptions, not silent degradations.

**Platform transport isolation (ADR-016 amendment):**
Desktop uses Wails bindings only. Web Tier 1 uses WASM bridge. Web Tier 2
uses HTTP fetch to connected server. Transports are mutually exclusive per
build target. React components are transport-agnostic.

**chi server scope (ADR-016 amendment):**
chi HTTP server is compiled into cmd/server/ only. It is not bundled into
the Desktop binary. It is not required for Web Tier 1. It is the Tier 2
server binary distributed for local, Docker, self-hosted, and hosted use.
