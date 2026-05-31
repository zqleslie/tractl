# ADR-013 — Traffic Acquisition & Capture Architecture

Status: Accepted
Date: 2026-05-24

---

## Context

A significant adoption friction point for backend workflow tools is the authoring cold-start problem: developers must manually describe requests they are already making. Traffic capture eliminates this friction by observing live traffic and normalizing it into executable workflow artifacts.

traCtl's capture model must address several distinct acquisition sources — browser interception, local proxy capture, bridge agents, HAR import, and curl import — without allowing capture-layer variation to pollute the canonical execution pipeline.

The existing architecture (ADR-001, ADR-003) defines normalization as mandatory before execution. Capture is a pre-normalization ingestion concern. Without an explicit acquisition architecture, capture sources either bypass normalization (unsafe) or accumulate ad-hoc normalization paths (inconsistent).

This ADR formally defines the acquisition layer, capture adapters, normalization boundary, and trust model for all capture sources.

---

## Decision

### 1 — Acquisition Layer

traCtl introduces a formal Acquisition Layer as a pre-normalization ingestion boundary.

The Acquisition Layer is distinct from the Provider Host and the Execution Pipeline.

Canonical position:

```
Traffic Source
    ↓
Acquisition Layer
    ↓ (emits normalized Capture Records)
Capture Normalization
    ↓ (emits traCtlSpec)
Canonical Execution Pipeline (ADR-003 §2)
```

The Acquisition Layer:

- receives raw traffic from Capture Adapters
- normalizes raw traffic into Capture Records
- forwards Capture Records to Capture Normalization
- does NOT execute workflows
- does NOT interact with the Runtime Layer
- does NOT bypass Canonical Validation (ADR-003 §2)

---

### 2 — Capture Adapters

Each traffic source connects to the Acquisition Layer through a Capture Adapter.

Defined Capture Adapters:

**Browser Interception Adapter**

Source: Browser Extension (ADR-012 §1)

Mechanism: Browser DevTools Protocol (CDP) or WebExtensions API network interception hooks.

Output: Capture Records from observed browser requests during a live browsing session.

Trust class: User-initiated (§7).

**Local Proxy Adapter**

Source: Locally running HTTP/HTTPS proxy intercepting application traffic.

Mechanism: HTTP CONNECT tunneling with local certificate trust injection.

Output: Capture Records from intercepted proxy traffic.

Trust class: User-configured (§7).

**Bridge Agent Adapter**

Source: Lightweight local agent installed alongside a running backend service.

Mechanism: Transparent request observation at the service boundary (not packet capture).

Output: Capture Records from observed service requests.

Trust class: User-configured (§7).

**HAR Import Adapter**

Source: HTTP Archive (HAR) file.

Mechanism: Static file ingestion with schema validation.

Output: Capture Records from HAR entries.

Trust class: Imported artifact (§7).

**curl Import Adapter**

Source: curl command string (single or multi-command).

Mechanism: curl command parsing into structured request representation.

Output: Capture Records from parsed curl commands.

Trust class: Imported artifact (§7).

**Execution Capture Adapter**

Source: traCtl execution runtime.

Mechanism: Execution context hooks emit observed request/response pairs during workflow execution.

Output: Capture Records from live execution runs.

Trust class: Internal (§7).

---

### 3 — Capture Record Model

All Capture Adapters emit a normalized intermediate representation: the Capture Record.

The Capture Record is NOT traCtlSpec. It is a pre-normalization structure.

Capture Record MUST contain:

- acquisition source identifier
- acquisition timestamp
- raw request representation (method, URL, headers, body)
- raw response representation (status, headers, body) if available
- provenance metadata (§5)
- trust class (§7)

Capture Record MUST NOT contain:

- execution semantics
- assertion definitions
- variable declarations
- workflow dependency definitions

These are added during and after normalization, not during capture.

---

### 4 — Capture Normalization

Capture Normalization converts Capture Records into traCtlSpec.

Normalization is deterministic and rule-based.

Normalization responsibilities:

- URL parameterization (detect and extract path/query variables)
- header canonicalization (normalize header casing, dedup)
- credential detection and masking (§8)
- body schema inference (where applicable)
- request/response pairing
- workflow step generation from request sequences
- dependency inference from observable request ordering

Normalization MUST be deterministic.

Same Capture Record + same normalization rules → same traCtlSpec output.

Normalization MUST NOT:

- generate nondeterministic workflow identifiers
- inject runtime-specific state
- embed acquisition-specific metadata into traCtlSpec
- silently drop captured fields without diagnostics

---

### 5 — Provenance Tracking

Every traCtlSpec artifact produced by the Acquisition Layer MUST carry acquisition provenance.

Provenance fields:

- acquisition source type (enum: browser, proxy, agent, har, curl, execution)
- acquisition timestamp
- capture adapter version
- normalization rule version
- original source reference (where applicable, e.g. HAR filename)
- provenance chain (if artifact was derived from another captured artifact)

Provenance is metadata only.

Provenance MUST NOT alter execution semantics.

Provenance is preserved through overlay application and into execution diagnostics.

Rationale: Provenance enables workflow origin auditability. When a developer asks "where did this request come from," provenance provides a traceable answer without requiring re-capture.

---

### 6 — Replay Semantics

Captured and normalized workflows MUST be replayable.

Replay is first-class, not a secondary feature.

Replay requirements:

- normalized traCtlSpec from capture executes identically to hand-authored traCtlSpec
- execution engine applies no capture-specific execution path
- captured workflows are subject to the same canonical pipeline (ADR-003 §2)
- captured workflows support overlay augmentation (variable injection, auth enrichment, assertion addition)

Replay is the bridge between capture and validation.

Capture without replay produces observation artifacts, not executable workflows.

---

### 7 — Source Trust Classification

Capture sources carry explicit trust classes.

| Trust Class | Sources | Policy |
|---|---|---|
| Internal | Execution Capture Adapter | Full trust. Runtime-generated. No external input. |
| User-Initiated | Browser Interception Adapter | Elevated trust. User explicitly triggered capture. Subject to consent model (ADR-006 §10). |
| User-Configured | Local Proxy Adapter, Bridge Agent Adapter | Elevated trust. User explicitly configured source. Subject to consent model. |
| Imported Artifact | HAR Import Adapter, curl Import Adapter | Untrusted by default. Treated identically to imported workflow artifacts (ADR-006 §6). |

Trust class governs:

- admission validation strictness
- provenance display in UX
- capability resolution constraints
- credential handling (§8)

Lower trust class MUST NOT be promoted by overlay application.

Rationale: A HAR file from an unknown source must not receive the same trust as a workflow produced by the runtime's own execution capture. Trust classification is the acquisition-layer expression of ADR-006 §2 (zero implicit trust).

---

### 8 — Credential Masking

The Acquisition Layer MUST detect and mask credentials before Capture Records exit the adapter boundary.

Credential detection applies to:

- Authorization headers (Bearer tokens, Basic auth, API keys)
- Common API key header patterns (X-API-Key, X-Auth-Token, etc.)
- Cookie values matching authentication patterns
- Query parameter values matching authentication patterns
- Request body fields matching credential patterns (password, token, secret, key)

Masked representation:

- Credential values are replaced with typed placeholders at capture time
- Placeholder format preserves credential type without exposing value
- Placeholders are variable-reference compatible for overlay injection

Masking is not optional.

Masking MUST occur at adapter boundary, before Capture Records enter normalization.

Unmasked credentials MUST NOT appear in:

- Capture Records
- normalized traCtlSpec
- provenance metadata
- execution diagnostics

Rationale: Capture output is intended to become a Git-native artifact. A workflow file committed to Git containing unmasked credentials is a critical security failure. Masking at the earliest possible boundary prevents this.

---

### 9 — Acquisition Boundary vs Runtime Boundary

The Acquisition Layer is explicitly NOT part of the Runtime Layer.

The boundary is enforced structurally:

- Acquisition Layer has no access to Runtime state
- Runtime Layer has no dependency on Acquisition Layer implementation
- The only interface between them is the traCtlSpec emitted by Capture Normalization
- Capture Adapters have no execution lifecycle access

Rationale: Collapsing acquisition into the runtime would couple traffic observation semantics to execution semantics. These are independent concerns. A runtime should execute deterministically without knowledge of how its input was produced.

---

## Consequences

**Positive:**

- Capture-to-execution path is deterministic and auditable
- All acquisition sources normalize through the same interface
- Credential safety is enforced at the earliest architectural boundary
- Replay is first-class, not an afterthought
- Provenance enables workflow origin auditability
- Trust model extends cleanly from ADR-006

**Tradeoffs:**

- Capture Adapter implementation overhead per source type
- Credential detection heuristics require ongoing maintenance
- Normalization determinism must be validated across adapter outputs
- Proxy and browser adapters require OS/browser-level trust configuration

---

## Alternatives Considered

**Direct HAR-to-execution without normalization layer:** rejected. Bypasses canonical pipeline. Creates untrusted execution path.

**Unified proxy-only capture:** rejected. Excludes browser extension and CLI import sources. Limits adoption surface.

**Capture as execution provider:** rejected. Conflates acquisition semantics with execution semantics. Couples runtime to capture implementation.

**Post-execution credential masking:** rejected. Credential exposure window between capture and masking is unacceptable. Masking must occur at adapter boundary.

**Trust promotion by user override:** rejected. Users must not be able to promote imported artifact trust class. Prevents social engineering attacks via crafted HAR files.
