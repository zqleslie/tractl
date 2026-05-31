# ADR-010 — Documentation Governance Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl has a growing set of architecture documents, product documents, specifications, schemas, and operational artifacts. Without explicit governance, these documents accumulate contradictions silently. Architecture changes get recorded in READMEs. Terminology drifts across documents. Generated artifacts get treated as authoritative. The document set becomes inconsistent faster than it gets corrected.

---

## Decision

### 1 — Normative Hierarchy

Documentation governance is explicitly hierarchical.

Normative hierarchy (highest to lowest):

1. ADRs
2. Canonical architecture specifications
3. Terminology registry (`00_terminology.md`)
4. Schemas and wire contracts
5. HLD
6. PRD
7. Roadmap
8. Onboarding documentation
9. README

**Conflict resolution:** conflicts between documents resolve toward the higher position in the hierarchy. The lower document is incorrect and must be updated.

Rationale for terminology registry at position 3: terminology drift is silent corruption. When the HLD uses "job" where the terminology registry says "workflow," the HLD is wrong. Placing the registry above the HLD enforces this. Synonym drift (workflow/job/run/task) baked into implementation types is expensive to fix.

### 2 — Source-of-Truth Classification

Every document MUST carry an explicit classification.

Classifications:

- **Normative:** authoritative. Governs behavior, architecture, and implementation. Conflicts resolve in favor of normative documents.
- **Descriptive:** informational. Explains, illustrates, or summarizes normative content. Not authoritative.
- **Generated:** produced from normative sources. Non-authoritative unless explicitly promoted.
- **Deprecated:** superseded. Retained for historical context only. MUST NOT be used as authority.

Unclassified documents are treated as descriptive.

### 3 — Generated Artifact Authority

Generated artifacts are non-authoritative.

Examples of generated artifacts:

- auto-generated API reference docs
- schema documentation generated from code
- changelogs generated from commit history
- diagrams generated from code analysis

Generated artifacts MUST NOT be promoted to normative authority without explicit decision and classification update.

Drift between a generated artifact and its normative source means the normative source is correct and the generated artifact is stale.

### 4 — Architecture Change Process

Architecture changes require ADR updates.

Architecture changes MUST NOT be recorded in:

- README updates
- roadmap notes
- onboarding document changes
- HLD patches without ADR backing

An architecture change without a corresponding ADR update has not been made architecturally. It has only been made in implementation.

Rationale: The hierarchy in decision 1 is advisory without enforcement. This decision is the enforcement mechanism. Architecture changes enter the system at the top of the hierarchy, not the bottom.

### 5 — Terminology Governance

The canonical terminology registry is a normative architecture artifact at hierarchy position 3.

**The terminology registry is the authoritative source for:**

- canonical term definitions
- category groupings (protocol categories, provider categories, validation stages, authoring formats)
- prohibited synonyms
- disambiguation of overlapping concepts

All documents MUST use terminology consistent with the registry.

Conflicting terminology in any document is a defect in that document, not in the registry.

**Synonym drift prevention:**

The following term sets are common drift vectors and MUST be explicitly governed in the registry:

- workflow / job / run / task / execution
- provider / adapter / plugin / extension / connector
- capability / feature / function / tool
- stage / phase / step / node

When a term is introduced in implementation that does not appear in the registry, the registry MUST be updated before the term propagates.

---

## Consequences

**Positive:**

- Document set remains consistent as architecture evolves
- New contributors have explicit authority hierarchy
- Architecture changes are traceable to ADRs
- Terminology remains stable across implementation surface
- Generated artifact drift is detectable

**Tradeoffs:**

- Architecture change process overhead
- Registry maintenance burden
- Requires active enforcement — hierarchy is advisory without discipline

---

## Alternatives Considered

**Flat documentation model:** rejected. Contradictions accumulate. No resolution mechanism.

**Code as documentation:** rejected. Architecture intent not expressible in code structure alone. ADR reasoning cannot be inferred from implementation.

**Informal terminology conventions:** rejected. Synonym drift is silent and expensive. Formal registry with normative authority prevents it.

**README-first change recording:** rejected. Changes enter at lowest authority level. Architecture becomes whatever was last committed to the README.
