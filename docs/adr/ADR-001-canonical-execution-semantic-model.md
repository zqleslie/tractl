# ADR-001 — Canonical Execution Semantic Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl accepts workflow intent from multiple sources:

- UI-driven structured input
- API payloads
- Prompt-derived generated workflows
- TOON authoring format
- YAML authoring format
- JSON authoring format
- Imported ecosystem artifacts

Multiple representations create execution ambiguity unless a single authoritative semantic model exists.

Without a canonical model, execution behavior becomes representation-specific. Parser-level differences leak into execution semantics. Validation becomes inconsistent. Future runtime evolution becomes dangerous.

---

## Decision

### 1 — Canonical Truth

`traCtlSpec` is the canonical semantic contract.

The Go runtime is the reference implementation.

This does not mean "Go structs are the architecture." The semantic contract is language-independent. The Go runtime implements it as the reference. Future runtimes must conform to the same contract.

### 2 — Universal Normalization

ALL workflow inputs MUST normalize into `traCtlSpec` before any planning, compilation, validation, or execution occurs.

Accepted upstream sources:

- UI structured form input
- API payloads
- Prompt-derived generated workflows
- TOON
- YAML
- JSON
- Imported ecosystem artifacts
- Future SDK-generated workflows

Execution MUST never depend on representation-specific parsing semantics.

No parser-specific execution path is permitted.

### 3 — Prompt Safety Boundary

Prompt output MUST NOT execute directly.

Mandatory flow:

```
Prompt → intermediate generated workflow → schema validation → semantic validation → normalization → traCtlSpec
```

LLM-generated output is untrusted input. No LLM hallucination may directly touch execution. This applies equally to streaming prompt output — real-time generated workflows must complete the same admission pipeline before being promotable to executable.

### 4 — Persistence

`traCtlSpec` MAY be persisted internally for:

- execution snapshots
- state recovery
- migration support
- cached compiled workflows
- debugging surfaces

External interchange remains representation formats (TOON, YAML, JSON).

Internal persistence uses canonical spec snapshots.

These are separate concerns and must remain separate.

### 5 — Versioning

The canonical spec MUST carry an explicit version field.

```
specVersion: "1"
```

Schema evolution requires version bumps.

Validation MUST hard-reject unknown spec versions.

Best-effort interpretation of unknown versions is prohibited.

---

## Consequences

**Positive:**

- Single execution truth
- Deterministic validation behavior
- Representation independence
- Future runtime extensibility
- Simplified runtime guarantees
- No parser-specific execution drift

**Tradeoffs:**

- Normalization layer required for every input source
- Import/export conversion complexity
- Schema versioning governance overhead

---

## Alternatives Considered

**Direct execution from TOON:** rejected. Couples execution semantics to DSL evolution. TOON becomes de facto canonical truth by gravity.

**Direct execution from YAML/JSON:** rejected. Representation leakage into execution semantics. Parser differences become execution differences.

**Multiple canonical models:** rejected. Ambiguity. No single validation boundary.

**Execution from UI state directly:** rejected. Same representation leakage problem. UI internal state must map through canonical model.

---

## Amendment — 2026-05-24

Amended by: ADR-013 (Traffic Acquisition & Capture Architecture)

### Context

ADR-013 introduces the Acquisition Layer as a pre-normalization ingestion boundary for traffic capture sources (browser interception, local proxy, HAR import, curl import, bridge agent, execution capture). These acquisition sources must normalize into traCtlSpec before any execution, consistent with the universal normalization invariant in §2 of this ADR.

However, acquisition sources arrive as Capture Records — an intermediate pre-normalization representation — rather than as directly parseable workflow authoring formats. The source enumeration in §2 did not anticipate this class of upstream input.

### Additional Decisions

#### 6 — Acquisition Sources in Universal Normalization

The universal normalization invariant (§2) is extended to explicitly include acquisition sources.

All acquisition sources MUST normalize into `traCtlSpec` through the Acquisition Layer (ADR-013 §1) before entering the canonical execution pipeline.

Extended upstream source enumeration:

- UI structured form input
- API payloads
- Prompt-derived generated workflows
- TOON
- YAML
- JSON
- Imported ecosystem artifacts
- Future SDK-generated workflows
- **Capture Records from Acquisition Layer (ADR-013 §3)**

The normalization path for acquisition sources is:

```
Traffic Source
    ↓
Capture Adapter (ADR-013 §2)
    ↓
Capture Record (ADR-013 §3)
    ↓
Capture Normalization (ADR-013 §4)
    ↓
traCtlSpec
    ↓
Canonical Execution Pipeline (ADR-003 §2)
```

No stage in this path may be bypassed.

Capture Records are an internal intermediate representation.

Capture Records MUST NOT be treated as traCtlSpec.

Capture Records MUST NOT be directly admitted to execution.

#### 7 — Provenance Field in traCtlSpec

traCtlSpec produced from acquisition sources MUST carry acquisition provenance metadata (ADR-013 §5).

Provenance is metadata only and MUST NOT alter:

- execution semantics
- planning behavior
- assertion evaluation
- extract evaluation

Provenance is preserved through overlay application and remains accessible in execution diagnostics (ADR-014 §5).
