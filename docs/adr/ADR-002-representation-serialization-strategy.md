# ADR-002 — Representation & Serialization Strategy

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl requires:

- interoperable persistence
- version control compatibility
- import/export flexibility
- machine integration support
- human authoring ergonomics

A single representation format cannot serve all of these concerns simultaneously. A canonical semantic model (ADR-001) separates execution truth from serialization format, enabling multiple representations to coexist.

---

## Decision

### 1 — Representation Philosophy

The canonical semantic model is NOT a serialization format.

This is an explicit architecture invariant:

```
traCtlSpec (semantic model) ≠ serialization format
```

No serialization format is canonical execution truth. All formats are interchange artifacts.

This invariant prevents any single format from accumulating implicit semantic authority through usage patterns.

### 2 — TOON Positioning

TOON is the preferred human-readable portable DSL for workflow authoring.

TOON is:

- first-class authoring format
- preferred for human-readable portability
- preferred for Git-native storage

TOON is not:

- the canonical execution model
- more authoritative than YAML or JSON
- exclusive authoring surface

Authors choosing YAML or JSON do not produce semantically inferior workflows.

### 3 — Format Semantic Parity

Full bidirectional semantic parity is required across all three formats.

- Anything expressible in TOON MUST be representable in YAML and JSON.
- Anything expressible in YAML or JSON MUST be representable in TOON.

This is bidirectional. "TOON-only" semantics are prohibited.

Rationale: A secretly richer format breaks portability. Machine-generated workflows in JSON must round-trip cleanly to TOON for human review in Git. Asymmetric expressiveness makes this impossible.

### 3a — Cross-Format Overlay Parity

Semantically equivalent source + overlay combinations MUST normalize identically, regardless of which native authoring format is used for the source or the overlay.

Native examples:

```
workflow.toon + overlay.yaml
workflow.yaml + overlay.toon
workflow.json + overlay.json
```

Interoperability source examples:

```
openapi.yaml + overlay.toon
openapi.json + overlay.yaml
postman.json + overlay.toon
```

Rules:

- Cross-format overlay application MUST preserve canonical equivalence.
- Authoring format choice for source or overlay MUST NOT introduce semantic difference in canonical `traCtlSpec` output.
- Any combination producing semantically different canonical output is a conformance defect.

This extends the bidirectional parity guarantee in §3 to cover the source + overlay composition surface defined by the Overlay Specification.

### 4 — Supported Representations

**Import formats:**

- TOON
- YAML
- JSON
- API structured payloads
- Prompt-generated workflow definitions

**Export formats:**

- TOON
- YAML
- JSON

**Default save format (Alpha):**

YAML

### 5 — UI Authoring Model

Primary user interaction is structured UI and API authoring.

Textual DSL authoring is not the default end-user interaction model.

Text representations are portability and interchange artifacts.

The UI MUST NOT be implemented as a raw TOON text editor.

UI internal state MUST use canonical workflow semantics. Serialization targets are YAML, JSON, and TOON.

---

## Consequences

**Positive:**

- Flexible interoperability
- Developer familiarity (YAML default)
- Source control compatibility
- Representation portability
- Clean semantic model separation

**Tradeoffs:**

- Schema compatibility maintenance across three formats
- Serializer versioning complexity
- Bidirectional parity test coverage required

---

## Alternatives Considered

**TOON-only representation:** rejected. Excludes machine integration, forces unfamiliar authoring on all users.

**YAML-only representation:** rejected. Loses TOON portability and human ergonomics advantage.

**JSON-only representation:** rejected. Poor human authoring ergonomics.

**TOON as canonical truth:** rejected. Format gravity accumulates implicit authority, violates ADR-001.
