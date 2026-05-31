# traCtl High-Level Design (Canonical)

## 1. Purpose

This document defines the canonical high-level architecture for traCtl.

This architecture aligns with the PRD and preserves long-term platform extensibility while keeping MVP implementation intentionally focused.

Product scope and architecture scope are intentionally different.

- PRD defines delivered product scope.
- HLD defines architectural boundaries and evolution-safe design.

---

## 2. Architectural Principles

### Product Alignment

Architecture must serve product strategy:

- Git-native backend validation
- deterministic execution
- local and CI parity
- compatibility-first adoption
- local-first execution
- capture-assisted workflow authoring
- execution-context observability
- multi-surface delivery

---

### Core Execution Invariant

traCtl must preserve the core execution promise:

**Same workflow definition. Same execution command. Equivalent execution semantics.**

Across all supported delivery surfaces (ADR-012):

- Desktop
- Browser Extension (capture-only; delegates execution)
- CLI
- CI
- Container Runtime

This is a non-negotiable architectural invariant.

---

### Platform-Ready, Product-Focused

traCtl is architected as a platform from day one.

This does not require full platform delivery in MVP.

Architecture must preserve future extensibility without introducing premature implementation complexity.

---

### Deterministic Execution

Execution behavior must remain deterministic.

Local developer feedback must remain lightweight enough for normal pre-merge workflows.

Heavyweight validations must be explicitly staged.

---

### Capability Contract Model

Workflow requirements and component-provided behaviors are expressed as capability contracts.

A capability contract is a versioned semantic definition carrying a capability identifier, semantic contract version, declared inputs and outputs, behavioral guarantees, execution constraints, and compatibility metadata.

Capabilities resolve through one canonical contract model regardless of source. Native engine capabilities, extension capabilities, provider adapter capabilities, and MCP adapter capabilities all normalize into the same capability contract representation. The planner matches capabilities by contract compatibility, never by the name of the component that provides them.

Capabilities are tiered. Platform capabilities are engine-owned, version with the engine, and are referenced as stable flags. Extension and provider capabilities are owned outside the engine, carry mandatory semantic versions in identifier-and-version form, and drift independently of the engine.

Capability resolution occurs at planning time. Unknown capability contracts and incompatible capability versions fail planning. Silent capability substitution is not permitted.

---

### Provider Host Boundary

Platform architecture must preserve a provider host boundary for installable provider lifecycle support.

This boundary enables:

- provider discovery
- installable provider lifecycle
- hot provider loading
- provider isolation boundaries
- governed execution delegation

Provider host architecture is required for platform evolution.

---

### Git-Native Artifact Model

Workflow assets are plain-text, filesystem-native, repository-native artifacts.

Architecture must support:

- version control
- code review
- merge workflows
- reproducibility
- CI execution
- auditable change history

Opaque workspace-dependent execution is explicitly avoided.

---

### Compatibility-First Execution

traCtl executes existing ecosystem artifacts through provider compatibility layers.

Compatibility execution is preferred over forced migration.

Overlay is a source-agnostic augmentation and composition mechanism. Overlay applies to both native sources (TOON, YAML, JSON) and interoperability sources (OpenAPI, WSDL, Postman, HAR, and extension-defined formats). Overlay enriches native workflows with environment-specific behavior, and interoperability sources with traCtl validation semantics.

Execution fidelity must be explicit and predictable.

No silent compatibility degradation is permitted.

---

## 3. System Architecture Overview

Full canonical pipeline including Acquisition Layer and Observability (ADR-003 §6, ADR-013 §1, ADR-014):

```text
Execution Source / Capture Source
        │
        ├─ Execution Source ─────────────────────────────────────────────────────────────┐
        │   (YAML, JSON, TOON, OpenAPI, Postman, Bruno, HAR, curl, SDK-generated)        │
        │                                                                                 │
        └─ Capture Source                                                                 │
            │                                                                            │
            ▼                                                                            │
    Acquisition Layer (ADR-013)                                                          │
        │                                                                                │
        ├─ Browser Interception Adapter                                                  │
        ├─ Local Proxy Adapter                                                           │
        ├─ Bridge Agent Adapter                                                          │
        ├─ HAR Import Adapter                                                            │
        ├─ curl Import Adapter                                                           │
        └─ Execution Capture Adapter                                                     │
            │                                                                            │
            ▼                                                                            │
        Capture Record                                                                   │
            │                                                                            │
            ▼                                                                            │
        Capture Normalization + Credential Masking (ADR-013 §4, §8)                     │
            │                                                                            │
            ▼                                                                            │
    traCtlSpec (+ Provenance Metadata) ◄────────────────────────────────────────────────┘
        │
        ▼
    Provider Host
        │
        ▼
    Provider Contract  (Native Parser | Compatibility Adapter)
        │
        ▼
    Canonical Source Model  (+ provenance for interoperability sources)
        │
        ▼
    Overlay Engine
        │
        ▼
    Canonical Workflow Specification  (post-overlay traCtlSpec)
        │
        ▼
    Canonical Validation
        │
        ▼
    Execution Planner  [observability: planning trace]
        │
        ▼
    DAG Compiler
        │
        ▼
    Runtime Layer
        │  [observability: Request Timeline per step]
        │  [observability: Retry events]
        │  [observability: Redirect chains]
        │
        ▼
    Assertions + Extracts  [observability: evaluation events]
        │
        ▼
    Diagnostics + Reporting
        │  [observability: Execution Provenance Trace]
        │  [observability: Dependency Waterfall]
        │
        ▼
    Output (human-readable | JSON | YAML | TOON)
```

Pipeline rules:

- Overlay resolution MUST complete before canonical validation.
- Canonical validation MUST complete before planning.
- Overlay MUST NOT become runtime input directly.
- Canonical Workflow Specification is post-overlay normalized output.
- Acquisition Layer output MUST enter the pipeline as traCtlSpec after Capture Normalization.
- No capture source may bypass Canonical Validation.
- Observability instrumentation MUST NOT alter execution semantics.

Capability resolution operates at the planning boundary, consuming the canonical capability contract model. Capability contracts originating from the engine, extensions, provider adapters, and MCP adapters are resolved uniformly by the planner.

---

## 4. Core Architectural Components

### 4.1 Execution Sources

Execution sources define user-facing workflow inputs.

#### Native Sources

Platform-native workflow definitions:

- native JSON
- native YAML
- TOON

#### Native Protocol/Spec Providers

Protocol/spec ecosystems directly understood by the platform with explicitly declared semantic fidelity boundaries.

Initial:

- OpenAPI

Future examples:

- WSDL
- GraphQL SDL
- gRPC descriptors

#### External Compatibility Providers

Vendor ecosystem integrations through installable providers.

Examples:

- Postman
- Bruno
- Insomnia
- future ecosystem providers

#### Delegated Specialist Providers

Specialist execution engines.

Examples:

- performance engines
- security engines
- diagnostics engines

Execution source logic must remain isolated from core execution semantics.

---

### 4.2 Acquisition Layer

The Acquisition Layer is a pre-normalization ingestion boundary for traffic capture sources (ADR-013).

The Acquisition Layer is distinct from the Provider Host and the Execution Pipeline.

Responsibilities:

- receive raw traffic from Capture Adapters
- normalize raw traffic into Capture Records
- apply credential masking at adapter boundary
- assign provenance metadata
- forward to Capture Normalization

The Acquisition Layer does NOT execute workflows, interact with the Runtime Layer, or bypass Canonical Validation.

#### Capture Adapters

| Adapter | Source | Trust Class |
|---|---|---|
| Browser Interception Adapter | Browser Extension DevTools Protocol | User-Initiated |
| Local Proxy Adapter | Local HTTPS proxy | User-Configured |
| Bridge Agent Adapter | Local service observation agent | User-Configured |
| HAR Import Adapter | HAR file | Imported Artifact |
| curl Import Adapter | curl command string | Imported Artifact |
| Execution Capture Adapter | traCtl execution runtime | Internal |

Trust classes govern admission validation strictness and capability resolution constraints (ADR-006 §7, ADR-013 §7).

#### Capture Normalization

Converts Capture Records into traCtlSpec.

Normalization is deterministic.

Responsibilities:

- URL parameterization
- header canonicalization
- credential detection and masking
- body schema inference
- request/response pairing
- workflow step generation
- dependency inference

---

### 4.3 Provider Host

Provider host manages provider lifecycle.

Responsibilities:

- provider discovery
- provider loading
- provider registration
- hot provider lifecycle management
- isolation boundaries
- capability exposure
- delegated execution governance

Provider lifecycle management remains separate from execution planning.

Capability exposure registers each provider's capability contracts into the canonical capability contract model. Capability conflicts between sources are detected at registration, before planning consumes the contracts.

---

### 4.4 Provider Contract Layer

Provider contracts define compatibility interpretation boundaries.

Responsibilities:

- source parsing
- source validation
- compatibility interpretation
- declared fidelity guarantees
- capability declaration

Providers stop at interpretation boundaries.

Core execution semantics remain engine-owned unless explicitly delegated through governed specialist provider architecture.

---

### 4.5 Normalization + Composition Layer

Normalization converts compatibility sources into canonical execution semantics.

Responsibilities:

- structural normalization
- semantic normalization
- dependency extraction
- workflow graph derivation
- capability translation
- compatibility metadata preservation
- unsupported capability detection
- partial semantic mapping disclosure
- fidelity boundary declaration
- fidelity report generation
- overlay composition
- source augmentation merging

Inputs may include:

- provider source artifacts
- native source artifacts (TOON, YAML, JSON)
- traCtlSpec from Acquisition Layer (with provenance)
- native overlay artifacts (applicable to any source type)

Outputs:

- canonical workflow specification
- interoperability fidelity report

No silent data loss is permitted.

This layer owns interoperability truth.

---

### 4.6 Canonical Workflow Specification

Canonical workflow specification is the execution truth.

All execution sources and capture sources normalize into this representation.

Responsibilities:

- requests
- dependencies
- assertions
- variables
- environment bindings
- auth definitions
- retries
- timeout semantics
- validation directives
- execution metadata
- progressive validation configuration
- declared capability contract requirements
- acquisition provenance (when sourced from Acquisition Layer)

No runtime or provider may bypass canonical normalization.

---

### 4.7 Execution Planner

Planner converts canonical workflow semantics into deterministic executable plans.

Responsibilities:

- dependency analysis
- graph validation
- capability resolution
- runtime selection
- validation stage interpretation
- concurrency planning
- execution planning
- failure policy planning

Capability resolution matches each declared capability contract requirement against the capability contracts available in the execution environment. Resolution evaluates capability presence, contract version compatibility, semantic compatibility, declared input and output compatibility, and execution constraint compatibility.

Planning emits a planning trace event (ADR-014 §5, ADR-003 §8).

Unknown capability contracts fail planning. Incompatible capability versions fail planning. Silent substitution is prohibited.

---

### 4.8 DAG Scheduler

Scheduler executes deterministic dependency-aware orchestration.

Responsibilities:

- topological scheduling
- bounded concurrency
- branch execution coordination
- retry orchestration
- timeout propagation
- cancellation propagation
- deterministic failure handling

Default execution policy is `resilient`. `failFast` is an explicit opt-in.

Scheduler does not reinterpret workflow semantics.

---

### 4.9 Runtime Layer

Runtime executes planned protocol behavior.

Responsibilities:

- request execution
- credential transmission
- timeout enforcement
- retry execution
- protocol runtime behavior
- Request Timeline instrumentation (ADR-014 §2)
- retry event emission (ADR-014 §3)
- redirect chain event emission (ADR-014 §3)

**Included diagnostics scope:**

- DNS resolution timing and outcomes
- TCP connection timing
- TLS handshake diagnostics (version, cipher suite, certificate chain, timing)
- TTFB
- response transfer timing
- connection lifecycle telemetry
- retry and redirect event sequences

**Explicitly out of scope:**

- packet capture and raw packet stream access
- deep packet inspection
- NIC-level or kernel telemetry
- infrastructure observability

---

### 4.10 Observability Subsystem

The Observability Subsystem produces structured execution visibility data (ADR-014).

Scope: execution-context diagnostics only.

Not a packet analyzer. Not a network monitor.

Instrumentation points:

| Pipeline Stage | Data Produced |
|---|---|
| Acquisition Layer | Provenance record |
| Execution Planner | Planning trace event |
| Runtime — request | Request Timeline (DNS, TCP, TLS, TTFB, transfer) |
| Runtime — retry | Retry event |
| Runtime — redirect | Redirect hop event |
| Assertions | Assertion evaluation event |
| Extracts | Extract evaluation event (value placeholder only) |
| Workflow completion | Execution Provenance Trace, Dependency Waterfall |

All observability output is credential-safe by default (ADR-014 §6).

Output formats:

- structured JSON
- structured YAML
- human-readable text (CLI)
- rich visual (Desktop only)

---

### 4.11 Diagnostics + Reporting

Diagnostics provide execution visibility and reporting surfaces.

Responsibilities:

- execution telemetry
- assertion reporting
- dependency failure visibility
- interoperability fidelity reporting
- deterministic output rendering
- observability output assembly and delivery

Output contract modes:

- concise human-readable local output
- structured JSON automation output
- structured YAML output
- structured TOON output
- expanded observability output

Persona-aware rendering may vary by delivery surface.

---

### 4.12 Browser Bridge

The Browser Bridge manages communication between the Browser Extension and the Desktop application.

Protocol: native messaging (WebExtensions native messaging API).

Responsibilities:

- authenticated message framing between extension and Desktop
- capture session coordination
- Capture Record forwarding from extension to Acquisition Layer
- capture session state synchronization

Trust model: Browser Extension is external input (ADR-006 §11).

All messages arriving via Browser Bridge are validated before processing.

---

## 5. Delivery Surfaces

Defined by ADR-012. Summarized here for architecture reference.

### Desktop

Primary interactive developer UX surface.

Full observability visualization. Workflow authoring. Git integration. Inspection mode for observability data.

### Browser Extension

Capture-assisted workflow bootstrapping surface.

Does NOT execute workflows. Delegates to Desktop or CLI via Browser Bridge.

### CLI

Canonical automation execution surface.

Structured output. Deterministic exit codes. No interactive UX.

### CI

CLI-driven deterministic automation surface.

Workflow definitions must be committed to repository. Providers pre-installed. Version pinned.

### Container Runtime

Portable isolated execution surface.

Immutable images. Provider versions locked at build time. Environment config via mounted artifacts.

---

## 6. Extensibility Architecture

Future extensibility must preserve:

- installable compatibility providers
- installable execution providers
- delegated specialist providers
- intelligence providers
- governed provider lifecycle
- hot lifecycle evolution

Core engine remains provider-independent.

Capability contracts are the boundary through which extensions, provider adapters, and MCP adapters expose behavior to the planner.

Capture Adapters are NOT providers (ADR-007 §7). They operate at the Acquisition Layer boundary, upstream of the canonical pipeline.

---

## 7. Provenance Tracking

Provenance is a cross-cutting concern threading through multiple layers.

Provenance lifecycle:

- originated at Capture Adapter (ADR-013 §5)
- carried through Capture Normalization
- embedded in traCtlSpec as metadata
- preserved through Overlay Engine
- surfaced in Execution Provenance Trace (ADR-014 §5)
- available in Diagnostics output

Provenance MUST NOT alter execution semantics at any stage.

---

## 8. Security Architecture Summary

Security is execution admission, not middleware (ADR-006).

Key boundaries:

- Acquisition Layer: credential masking at adapter boundary, trust class assignment
- Canonical Validation: import safety pipeline, schema + semantic + policy validation
- Planning: capability version validation, trust class enforcement
- Browser Bridge: authenticated message framing, extension output treated as external input
- Local proxy CA: name-constrained, OS-keychain-protected, revocable
- Observability: credential-safe by default at all output surfaces

Zero implicit trust across all integration surfaces.

---

## 9. Explicit MVP Non-Goals

Not required for MVP:

- public provider marketplace
- signed public provider distribution
- enterprise governance platform
- remote execution fleet orchestration
- full specialist engine parity
- complete protocol ecosystem coverage
- Browser Extension (deferred to v1.5)
- Local proxy adapter (deferred to v1.5)
- Bridge agent (deferred to v2)

---

## 10. Architectural Quality Constraints

traCtl architecture must preserve:

- deterministic execution
- interoperability transparency
- execution reproducibility
- extensibility isolation
- Git-native artifact ownership
- capability contract integrity across sources and environments
- multi-surface execution parity
- credential safety at all system boundaries
- observability without semantic interference
- future platform evolution without architectural rework
