# traCtl Overlay Specification v1.0

> Version 1.0 · Normative Overlay Augmentation & Composition Specification Classification: **Normative**

## 1. Purpose

This specification defines the normative overlay model for traCtl.

An overlay is a deterministic augmentation and composition mechanism that modifies a source representation before canonical traCtlSpec creation.

Overlay enables users to:

- augment native traCtl sources
- enrich interoperability sources with native traCtl semantics
- compose environment-specific behavior
- apply deterministic structural transformations
- retain provenance across transformations

Overlay is a pre-execution transformation mechanism.

Overlay MUST NOT become runtime execution input directly.

---

## 2. Scope / Non-Scope

### In Scope

- overlay document structure
- targeting semantics
- merge semantics
- precedence rules
- provenance retention
- interpolation inheritance
- validation ownership
- extension participation constraints
- source adapter targeting requirements
- conformance obligations

### Out of Scope

This specification does NOT define:

- runtime execution behavior
- planner algorithms
- provider execution semantics
- canonical traCtlSpec semantics
- protocol execution models
- extension runtime SDK contracts

---

## 3. Architectural Position

Overlay exists in the pre-execution transformation pipeline.

### Native Sources

```text
TOON / YAML / JSON
    ↓
Native Parser
    ↓
Canonical Source Model
    ↓
Overlay Engine
    ↓
Canonical traCtlSpec
    ↓
Validation
    ↓
Planner
    ↓
Runtime Execution
```

### Interoperability Sources

```text
OpenAPI / WSDL / Postman / HAR / curl / Extension Sources
    ↓
Source Adapter
    ↓
Canonical Source Model + Provenance
    ↓
Overlay Engine
    ↓
Canonical traCtlSpec
    ↓
Validation
    ↓
Planner
    ↓
Runtime Execution
```

Overlay resolution MUST complete before canonical traCtlSpec validation.

---

## 4. Design Principles

1. **Source Agnostic** — overlays apply to native and interoperability sources.
2. **Serialization Neutral** — overlays may be authored in TOON, YAML, or JSON regardless of source format.
3. **Deterministic** — identical inputs MUST produce identical canonical outputs.
4. **Schema Aware** — merge semantics are governed by canonical schema contracts.
5. **Governance Constrained** — overlays MUST NOT bypass validation, capability declaration, or security policy.
6. **Parity Preserving** — overlays MUST preserve representation equivalence.
7. **Composable** — multiple overlays MAY be stacked deterministically.
8. **Provenance Preserving** — overlay contributions MUST remain traceable.

---

## 5. Applicability

Overlay applies to all supported source representations.

### Native Sources

- TOON
- YAML
- JSON

Examples:

```text
workflow.toon + overlay.yaml
workflow.yaml + overlay.json
workflow.json + overlay.toon
```

### Interoperability Sources

- OpenAPI
- WSDL
- Postman
- HAR
- curl
- Bruno
- extension-defined source formats

Examples:

```text
openapi.yaml + overlay.toon
wsdl.xml + overlay.yaml
postman.json + overlay.json
```

Cross-format overlay application is first-class.

---

## 6. Representation Parity

Equivalent overlay combinations MUST normalize identically.

Examples:

```text
workflow.toon + overlay.yaml
workflow.yaml + overlay.toon
workflow.json + overlay.json
```

If semantically equivalent, canonical output MUST be identical.

This extends representation parity guarantees.

---

## 7. TOON Positioning

TOON is the preferred ergonomic native authoring syntax.

TOON, YAML, and JSON remain semantically equivalent native representations.

No representation hierarchy is introduced.

---

## 8. Overlay Document Model

Canonical overlay structure:

```yaml
metadata:
  name: optional
  description: optional
patches:
  - target:
    action:
    patch:
```

The top-level overlay document has no wrapping key. `metadata` and `patches` are the only top-level fields.

### Fields

| Field    | Required | Description            |
| -------- | -------- | ---------------------- |
| metadata | optional | informational metadata |
| patches  | required | ordered patch list     |
| target   | required | selector definition    |
| action   | optional | merge override         |
| patch    | required for non-remove actions; absent for `remove` (see §11) | overlay contribution |

A single overlay file MAY contain multiple patches.

Patch order MUST be preserved.

---

## 9. Target Selection Model

Overlay supports three targeting modes. Each target MUST use exactly one of the three modes (`path`, `match`, or `source`). Combining modes within a single target is prohibited and MUST fail overlay validation.

### 9.1 Canonical Path Targeting

Example:

```yaml
target:
  path: workflows.user-check.steps.login
```

Targets canonical structural nodes.

Path targeting resolves to exactly one canonical node. Multiple matches are not possible by construction. Zero matches MUST fail overlay validation (target existence, per §15).

---

### 9.2 Semantic Matching

Example:

```yaml
target:
  mode: one
  match:
    kind: request
    protocol: http
```

Matches canonical nodes by semantic properties.

The `mode` field governs match cardinality:

| Mode  | Behavior                                                                 |
| ----- | ------------------------------------------------------------------------ |
| `one` | Exactly one match required. Zero or multiple matches MUST fail validation. |
| `all` | Zero or more matches accepted. Patch applies to every match.             |

`mode` defaults to `one` if omitted.

Mode `all` traversal order MUST be deterministic and MUST follow canonical normalized traversal order (see §12).

---

### 9.3 Source-Native Targeting

Example:

```yaml
target:
  mode: one
  source:
    type: openapi
    operationId: getUser
```

or:

```yaml
target:
  mode: all
  source:
    type: wsdl
    service: UserService
```

Source-native targeting resolves via source adapter provenance.

The `mode` field follows the same semantics as §9.2 and defaults to `one`.

Source selector schemas are adapter-defined and MUST be:

- deterministic
- documented in the adapter's capability contract
- capability governed (per ADR-004 §7 and ADR-011)
- validation compatible

An overlay referencing a source selector field not exposed by the adapter MUST fail overlay validation.

---

## 10. Merge Contract

Overlay merge is schema-aware.

### 10.1 Default Scalar Behavior

`replace`

Example:

```text
timeout PT10S + PT30S = PT30S
```

---

### 10.2 Default Object Behavior

`deepMerge`

Example:

```yaml
headers:
  Accept: application/json
```

+

```yaml
headers:
  Authorization: token
```

= merged object.

---

### 10.3 Array Behavior

Schema-aware defaults.

#### append

Default for arrays that semantically accumulate. Examples:

- assertions
- extracts
- tags

#### replace

Default for arrays whose membership is a set-like identity (replacement is the only meaningful merge). Examples:

- dependsOn
- capabilities
- invariants

#### appendUnique

Identity-aware arrays. Membership uses the array's declared identity contract (see §10.5).

If an incoming entry's identity collides with an existing entry, canonical identity semantics apply: the existing entry is replaced by the incoming entry (last-wins, consistent with overlay precedence in §12).

---

### 10.4 Explicit Actions

Supported actions:

- `replace`
- `deepMerge`
- `append`
- `appendUnique`
- `remove`

This is the complete set. No other action names are valid.

Explicit action overrides schema defaults.

If no explicit action is declared and no canonical merge semantic exists for the targeted field, overlay validation MUST fail.

---

### 10.5 appendUnique Identity Contracts

`appendUnique` requires a declared identity contract for the target array. The following identity contracts are normative:

| Array      | Identity                          |
| ---------- | --------------------------------- |
| assertions | `id`                              |
| extracts   | `id`                              |
| workflows  | map key                           |
| steps      | map key                           |
| extensions | map key                           |
| bindings   | map key                           |

If `appendUnique` is applied to an array with no declared identity contract, overlay validation MUST fail.

Extension-defined arrays that wish to support `appendUnique` MUST declare an identity contract in their capability declaration (per ADR-004 §7).

---

## 11. Removal Semantics

Removal is explicit only.

A removal patch:

- declares `action: remove`
- declares a deterministic `target`
- MUST NOT declare a `patch` field — removal is patchless

Example:

```yaml
patches:
  - action: remove
    target:
      path: workflows.user.steps.login.request.headers.Accept
```

Removal MAY target:

- scalar properties
- object properties
- array entries
- targeted canonical nodes

Rules:

- A `patch` field present alongside `action: remove` MUST fail overlay validation.
- The target MUST resolve deterministically. For `mode: all` source or semantic targeting, all resolved targets are removed.
- Implicit deletion (omitting a field in a non-remove patch in order to delete it) is prohibited. Deletion is expressed only through `action: remove`.

---

## 12. Overlay Stacking / Precedence

Multiple overlays are supported.

Application order:

```text
base
→ overlay1
→ overlay2
→ overlay3
```

**Inter-overlay precedence (stacking order):** later overlays override earlier overlays. Last overlay wins.

**Intra-overlay precedence (within a single overlay):** patch order wins. Patch ordering inside a single overlay MUST be preserved as authored.

**Deterministic traversal for `mode: all`:**

When a single patch resolves to multiple targets (via `mode: all` in semantic or source-native targeting), patch application across those targets MUST follow canonical normalized traversal order. Implementations MUST NOT use undefined ordering (hash-map iteration, parallel scheduling without ordering, etc.) for overlay merge.

Two conformant implementations MUST produce identical canonical output for identical inputs across:

- inter-overlay precedence (stacking order)
- intra-overlay patch order
- intra-patch target traversal order (under `mode: all`)

---

## 13. Provenance

Overlay provenance MUST be retained.

Examples:

- `overlayRefs` — references to overlay sources merged into the canonical spec
- contribution metadata — which overlay supplied which field
- merge attribution — which patch produced a given canonical node

Provenance is surfaced through `metadata` on the canonical traCtlSpec (see canonical spec §5.1).

Provenance is informational only. It is consumed by:

- diagnostics
- audit
- tooling
- review surfaces

Execution components MUST ignore metadata unless explicitly declared diagnostic-only. The planner MUST NOT use overlay provenance metadata for execution decisions. No execution component MAY alter scheduling, capability resolution, or runtime behavior based on metadata.

---

## 14. Interpolation

Overlay does not define independent interpolation semantics.

Overlay inherits canonical traCtlSpec expression resolution.

This includes:

- variable resolution precedence
- secret resolution rules
- expression evaluation behavior
- deterministic evaluation guarantees

---

## 15. Validation Ownership

Overlay processing involves three validation owners. Responsibilities MUST NOT overlap ambiguously.

### 15.0 Ownership Table for Overlay-Injected Content

When overlay injects content into a canonical traCtlSpec candidate, validation ownership for the injected content is as follows:

| Injected content type | Owner |
|---|---|
| Overlay document structure (metadata, patches, target, action, patch fields) | Overlay Validator |
| Selector validity (path existence, semantic match cardinality, source-native schema) | Overlay Validator |
| Merge action legality (action compatible with target field semantics) | Overlay Validator |
| Canonical schema correctness (injected fields conform to canonical field types and enums) | Canonical Validator |
| Malformed schema contributions (invalid enums, structurally invalid fragments) | Canonical Validator |
| Capability declaration shape and presence (injected capability declarations are well-formed) | Canonical Validator |
| Capability contract resolution (injected capability contracts satisfy available execution environment) | Planner |
| Cross-reference integrity (step refs, variable refs, extract refs) | Canonical Validator |
| DAG acyclicity (dependency graph remains acyclic after overlay) | Canonical Validator |
| Composite cycle detection (across composites, after recursive flattening) | Planner |
| Runtime compatibility of injected capability contracts | Planner |

A given validation rule has exactly one owner. Responsibility boundaries are non-overlapping.

If overlay injects malformed canonical fragments (invalid enums, bad schema, malformed capability declarations), the **Canonical Validator** rejects them after overlay resolution. The Overlay Validator owns only overlay mechanics; it does not re-validate canonical schema correctness of injected content.

### 15.1 Overlay Validator

Owns:

- overlay schema validation
- selector validity (path, semantic, source-native)
- target existence
- merge legality (action compatibility with target field semantics)
- precedence correctness (stacking order preservation)
- provenance integrity
- deterministic merge validation
- explicit action legality (per §10.4)
- mode `one` vs `all` cardinality enforcement (per §9)
- removal patch shape (no `patch` field; per §11)

### 15.2 Canonical Validator

Owns:

- canonical schema validation
- cross-reference integrity (variable refs, extract refs, step refs)
- canonical reference validity
- dependency graph structural validation (acyclicity, reference resolution)
- secret reference integrity
- capability declaration integrity (presence and shape of declarations)

### 15.3 Planner

Owns:

- capability resolution (matching declared capability contracts to available contracts per ADR-011)
- runtime compatibility
- execution planning
- composite cycle resolution via recursive flattening (per canonical spec §16.3)

The planner consumes a structurally valid canonical traCtlSpec. The planner does NOT re-validate canonical schema or basic DAG acyclicity — that is canonical validator responsibility.

Responsibilities MUST NOT overlap ambiguously. A given validation rule has exactly one owner.

---

## 16. Security Constraints

Overlay is augmentation. Overlay is NOT privilege escalation.

Overlay MUST NOT bypass:

- overlay schema validation
- canonical schema validation
- capability declarations
- extension trust boundaries (per ADR-004, ADR-006)
- secret policy (per ADR-006 §3)
- planner admission
- import safety pipeline (per ADR-006 §6)

These constraints are normative and mirror ADR-006 §9 (Overlay Security Constraints). Any conflict between this section and ADR-006 §9 resolves in favor of ADR-006 per ADR-010 normative hierarchy.

---

## 17. Extension Participation

If extensions participate in overlay processing — through overlay target matching, overlay merge processing, or source-native selector evaluation — they MUST be:

- deterministic
- pure
- side-effect free
- reproducible
- capability declared (per ADR-004 §7)
- security governed (per ADR-006 §9)

Explicitly prohibited inside extension overlay processors:

- network access
- filesystem access outside declared overlay inputs
- external mutation of any kind
- nondeterministic evaluation (clocks, randomness without seeded injection, ambient state)
- secret access

Extension-defined overlay behavior MUST NOT compromise deterministic canonical generation or cross-format overlay parity (ADR-002 §3a).

---

## 18. Source Adapter Contracts

### 18.1 Source Adapter vs Runtime Provider

The overlay pipeline distinguishes two component categories that MUST NOT be conflated:

**Source Adapters** — parse interoperability source formats and emit canonical source models with provenance. Examples: OpenAPI adapter, WSDL adapter, Postman adapter, HAR adapter. Source adapter responsibilities:

- parse source artifact
- emit canonical source model
- emit provenance sufficient for source-native overlay targeting
- declare source-native selector schema

**Runtime Providers** — execute planned protocol behavior at runtime. Examples: HTTP, gRPC, GraphQL. Runtime provider responsibilities are execution only, defined by the canonical traCtlSpec and the planner.

Source adapters operate in the pre-execution transformation pipeline (§3). Runtime providers operate in execution. Overlay interacts with source adapters via provenance and source-native selectors. Overlay does NOT interact with runtime providers.

### 18.2 Provenance Requirements

Source adapters MUST expose provenance sufficient for source-native targeting.

Examples:

#### OpenAPI

- operationId
- path
- method
- tags

#### WSDL

- service
- operation
- binding

#### Postman

- itemId
- folder path
- request identity

Overlay source selectors depend on adapter provenance correctness. Missing or inconsistent provenance is a source adapter defect, not an overlay defect.

---

## 19. Canonical Merge Output

Overlay output MUST be a valid canonical traCtlSpec candidate.

Overlay MUST NOT produce:

- partial execution artifacts
- runtime provider contracts
- invalid canonical structures

Canonical validation MUST occur after overlay resolution.

---

## 20. Conformance Requirements

A conformant implementation MUST:

- preserve authored patch order within a single overlay
- preserve overlay stacking order (last-wins precedence)
- apply `mode: all` targets in canonical normalized traversal order (§12)
- preserve overlay provenance and surface it through canonical metadata (§13)
- enforce all merge semantics defined in §10
- enforce identity contracts for `appendUnique` (§10.5)
- reject invalid selectors
- reject ambiguous targeting (multiple selector modes in one target)
- reject `mode: one` targets that match zero or multiple nodes
- reject removal patches that include a `patch` field (§11)
- reject `appendUnique` on arrays without a declared identity contract (§10.5)
- reject governance bypass attempts (§16, ADR-006 §9)
- preserve cross-format parity (§6, ADR-002 §3a)
- validate explicit action legality against target field semantics
- enforce extension safety constraints (§17, ADR-004 §7, ADR-006 §9)
- complete overlay resolution before canonical validation (§3)

Independent conformant implementations MUST produce identical canonical traCtlSpec output for identical source + overlay inputs. This is the operational test of cross-format parity and deterministic merge.

---

## 21. Examples

The following examples illustrate the overlay document model. Examples are normative for syntax and shape; they are not exhaustive.

### 21.1 Basic Overlay (Path Targeting)

```toon
metadata:
  name: dev-overlay

patches:
  - target:
      path: workflows.user-check.steps.login
    patch:
      timeout: PT30S
```

### 21.2 Semantic Selector (mode: all)

```toon
patches:
  - target:
      mode: all
      match:
        protocol: http
    patch:
      diagnostics:
        enabled: true
```

### 21.3 Source-Native Selector

```toon
patches:
  - target:
      source:
        type: openapi
        operationId: getUser
    patch:
      assertions:
        - id: statusOk
          kind: status
          op: equals
          expected: 200
```

### 21.4 Removal (Patchless)

```toon
patches:
  - action: remove
    target:
      path: workflows.user.steps.login.request.headers.Accept
```

### 21.5 Cross-Format Equivalence

The following combinations MUST produce identical canonical traCtlSpec output when the source and overlay are semantically equivalent:

```text
workflow.toon + overlay.yaml
workflow.yaml + overlay.toon
workflow.json + overlay.json
openapi.yaml + overlay.toon
openapi.json + overlay.yaml
```

This is the conformance surface for ADR-002 §3a (Cross-Format Overlay Parity).


