# ADR-006 — Security Trust Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl ingests workflow definitions from multiple sources, executes extensions and providers, integrates with LLM providers, and handles secrets. A compatibility-first adoption model means traCtl actively imports artifacts from external ecosystems. Each of these surfaces is a potential attack vector. Without explicit trust boundaries, privilege escalation paths and injection vectors accumulate silently.

---

## Decision

### 1 — Explicit Trust Zones

traCtl operates under explicitly defined trust zones.

Defined zones:

- core runtime
- workflow definitions (native)
- prompt-generated artifacts
- extensions
- providers (LLM and API)
- MCP integrations
- secrets
- user inputs
- imported workflow artifacts (TOON, YAML, JSON, OpenAPI, vendor formats)

Each zone has declared trust level and explicit boundary crossing policy.

### 2 — Zero Implicit Trust

Nothing crosses trust zone boundaries without explicit policy.

No implicit trust exists between:

- core runtime and extensions
- core runtime and providers
- core runtime and imported workflows
- extensions and secrets
- providers and execution state
- MCP integrations and orchestration

Zero implicit trust applies even where familiarity or convenience suggests it would be safe. Default untrusted is a hard default with no exceptions.

### 3 — Secret Access Policy

Secrets are NEVER directly exposed to:

- workflow definitions
- prompt generation context
- arbitrary extensions
- providers (unless explicitly required for that provider's declared function)
- MCP integrations

Secrets are accessible only through scoped injection at execution time.

Scoped injection means: the minimum secret scope required for the declared capability, resolved at execution time, not at authoring time.

### 4 — Provider Output Trust

Provider output (LLM providers, API providers) is untrusted input.

This applies even to providers from trusted vendors.

Provider output:

- MUST pass through validation before influencing execution
- MUST NOT directly mutate control flow
- MUST NOT directly mutate canonical workflow state

This is consistent with the prompt safety boundary in ADR-001. Provider output and prompt output receive identical trust treatment.

### 5 — Extension Permission Model

Extensions operate under capability-scoped permissions.

Blanket extension trust is prohibited.

Permission examples:

- network access (scoped to declared endpoints)
- filesystem access (scoped to declared paths)
- secret access (scoped to declared secret identifiers)
- connector access (scoped to declared connectors)
- execution hook access (scoped to declared lifecycle points)

Extensions MUST declare required permissions in their manifest (ADR-004). Undeclared permission access MUST be denied.

### 6 — Import Safety

Imported artifacts are untrusted by default regardless of format or source.

Imported formats subject to this policy:

- TOON
- YAML
- JSON
- OpenAPI specifications
- Postman collections
- Bruno files
- HAR files
- Any other ecosystem import format

Mandatory validation pipeline before execution admission:

```
Import → schema validation → semantic validation → capability validation → policy validation → normalization → traCtlSpec
```

No step may be bypassed.

Rationale: Compatibility-first adoption is traCtl's primary attack surface. Maliciously crafted Postman collections, poisoned OpenAPI specs, and injected YAML are realistic threat vectors for a tool that markets itself on import compatibility.

### 7 — Security as Execution Admission

Security validation is part of canonical execution admission.

Security validation is NOT optional middleware.

Security validation is NOT a pluggable security layer.

Security validation MUST occur before any workflow enters the execution pipeline.

This prevents future "security plugin" anti-patterns where security becomes optional, bypassable, or environment-specific.

Fast paths, dev modes, and compatibility modes that skip security validation are prohibited.

### 8 — Audit Observability

All trust boundary crossings are observable events.

The architecture MUST preserve the ability to observe boundary crossings regardless of whether logging is active in a given deployment.

Audit observability is an architectural requirement, not a deployment configuration.

### 9 — Overlay Security Constraints

Overlay is augmentation. Overlay is NOT privilege escalation.

Overlay MUST NOT bypass:

- schema validation
- canonical validation
- capability declarations
- trust boundaries (per §1, §2)
- secret resolution policy (per §3)
- planner admission
- import safety pipeline (per §6)

Overlay contributions are subject to the same trust treatment as the source artifact they augment. An overlay applied to an untrusted imported artifact does NOT promote either to trusted.

**Extension overlay processors** (extensions participating in overlay matching, merging, or source selector evaluation, per ADR-004 §7) MUST be:

- deterministic
- pure
- side-effect free
- reproducible
- security governed

The following are explicitly prohibited inside extension overlay processors:

- network access
- filesystem access outside declared overlay inputs
- external mutation of any kind
- nondeterministic evaluation (clocks, randomness without seeded injection, ambient state)
- secret access (secrets MUST NOT be exposed to overlay processing)

Rationale: Overlay resolution occurs before canonical validation and before planner admission. An overlay processor with network access or nondeterministic behavior would create a pre-admission bypass surface for every constraint in this document. Holding overlay processors to the strictest deterministic and isolation contract closes that surface by construction.

---

## Consequences

**Positive:**

- Eliminated implicit privilege escalation paths
- Consistent trust model across all integration surfaces
- Import safety for compatibility-first adoption model
- Future enterprise governance readiness
- No security bypass surface

**Tradeoffs:**

- Permission declaration overhead for extension authors
- Import validation pipeline adds latency
- Secret scope management complexity

---

## Alternatives Considered

**Implicit trust for first-party components:** rejected. Structural privilege escalation risk. See ADR-004.

**Security as optional middleware:** rejected. Creates bypass surface. Fast-path exceptions accumulate.

**Per-surface trust models:** rejected. Inconsistency creates boundary confusion. Unified zero-trust model prevents this.

**Best-effort import validation:** rejected. Compatibility adoption surface is primary attack vector. Best-effort is insufficient.

---

## Amendment — 2026-05-24

Amended by: ADR-012 (Multi-Surface Delivery Model), ADR-013 (Traffic Acquisition & Capture Architecture), ADR-015 (Packaging & Distribution Strategy)

### Context

ADR-012 introduces the Browser Extension and Local Proxy as first-class delivery surfaces. ADR-013 introduces the Acquisition Layer with multiple capture sources, including browser interception and local proxy capture. ADR-015 defines browser extension distribution with explicit permission requirements.

These new surfaces introduce trust surfaces not covered by the original ADR-006 trust model:

- the Browser Extension operates in a sandboxed browser context with its own permission model
- local proxy capture requires certificate trust injection
- localhost agent communication introduces a new inter-process trust boundary
- captured traffic from external sources carries a distinct trust class
- credential masking at the acquisition boundary requires explicit policy

The original §1 trust zone enumeration and §6 import safety pipeline are extended here to cover these surfaces.

### Additional Decisions

#### 10 — Interception Consent Model

Traffic capture via browser interception or local proxy MUST require explicit, informed user consent.

Consent requirements:

- user MUST explicitly activate capture for each session
- capture MUST NOT be active by default on application launch
- the active capture state MUST be visibly indicated in all interactive surfaces while capture is running
- users MUST be able to stop capture at any time without losing previously captured records
- consent state MUST NOT persist across application restarts without explicit user re-confirmation

Capture sessions MUST be scoped.

Scope options available to user:

- all traffic (explicit opt-in, not default)
- traffic to specific hostnames only
- traffic to specific URL patterns only

Default capture scope MUST be the most restrictive option.

Rationale: Background traffic interception without explicit per-session consent is a privacy violation regardless of technical capability. The consent model must be explicit, visible, and revocable.

#### 11 — Browser Extension Trust Boundaries

The Browser Extension operates in a sandboxed browser context.

Trust rules for the Browser Extension:

- the Browser Extension is NOT a trusted internal component
- the Browser Extension communicates with the Desktop application via native messaging only
- native messaging channel MUST be authenticated (origin verification)
- the Browser Extension MUST NOT have access to traCtl execution state beyond what is required for capture session coordination
- data flowing from Browser Extension to Desktop is treated as external input subject to validation

The Browser Extension trust class is: **User-Initiated** (ADR-013 §7).

This applies even though the Browser Extension is a first-party component. The sandboxed browser context means the extension cannot be trusted at the same level as the core runtime.

Rationale: Browser extensions operate in an environment traCtl does not control. Extension compromise, malicious injection into the browser environment, or crafted page content could manipulate extension behavior. Treating extension output as external input with validation is the correct posture.

#### 12 — Localhost Agent Communication Rules

Localhost inter-process communication (Desktop ↔ Browser Extension native messaging, Desktop ↔ Bridge Agent) is subject to explicit trust rules.

Rules:

- localhost IPC channels MUST use authenticated message framing
- message sources MUST be verified (process identity or origin token)
- messages arriving on localhost IPC MUST be validated before processing
- no implicit trust is granted to localhost-originating messages
- the loopback interface does not constitute a trust boundary

Rationale: Localhost does not mean trusted. Malicious local processes, compromised browser extensions, or crafted messages can target localhost IPC. Zero implicit trust applies to localhost the same as to remote sources.

#### 13 — Certificate Trust Constraints

Local proxy capture requires TLS interception via certificate injection.

Certificate trust constraints:

- traCtl MUST generate a unique, locally-scoped root CA per installation
- the local CA MUST NOT be a copy of or derived from any public CA
- installation of the local CA into the system trust store MUST require explicit user action and OS-level privilege confirmation
- the local CA certificate MUST carry a `Name Constraints` extension limiting its validity to localhost and private IP ranges
- the local CA private key MUST be stored with OS keychain protection (Keychain on macOS, DPAPI on Windows, Secret Service on Linux)
- the local CA MUST NOT be exportable without explicit user action
- the local CA MUST be revocable via a traCtl-provided uninstall action

Rationale: An unconstrained local CA is a catastrophic security risk. Name constraints limit the CA's blast radius to local traffic only, preventing it from being used to forge certificates for public domains even if the private key is compromised.

#### 14 — Credential Masking Policy

Credential masking (ADR-013 §8) is a security policy enforced at the Acquisition Layer boundary.

Masking policy for the trust model:

- masking MUST occur before Capture Records exit the adapter boundary
- unmasked credentials MUST NOT appear in any traCtl-persisted artifact
- unmasked credentials MUST NOT appear in observability output (ADR-014 §6)
- unmasked credentials MUST NOT appear in log output at any verbosity level
- credential detection patterns MUST be maintained and versioned
- false negative credential detection (missed masking) is a security defect, not a product limitation

Users MUST be informed when credential masking occurred during capture.

Rationale: The primary risk of capture-assisted workflow authoring is the accidental Git commit of workflow files containing live credentials. Masking at the earliest boundary, with explicit user notification, is the only posture that prevents this class of failure reliably.

#### 15 — Imported Traffic Trust Handling

Imported traffic sources (HAR files, curl imports) are treated as **Imported Artifact** trust class (ADR-013 §7).

This is identical in treatment to imported workflow artifacts (§6 of this ADR).

Mandatory validation pipeline for imported traffic:

```
Import → format validation → schema validation → semantic validation
       → credential detection + masking → provenance assignment
       → Capture Normalization → traCtlSpec → canonical validation
       → execution admission
```

No step may be bypassed.

A HAR file from an unknown or untrusted source is a realistic attack vector.

Maliciously crafted HAR files could attempt to:

- inject workflow definitions via request body content
- escalate privileges via crafted header values
- exfiltrate data via crafted redirect chains

Untrusted import treatment (validation before normalization, masking before persistence) mitigates these vectors.

#### 16 — Proxy Isolation Boundaries

The local proxy capture component MUST be isolated from the core execution runtime.

Isolation requirements:

- the proxy component MUST run as a separate process from the execution runtime
- the proxy component MUST communicate with the Acquisition Layer via a defined IPC interface only
- the proxy component MUST NOT have direct access to execution state, workflow state, or credential store
- the proxy component MUST be stoppable independently of the execution runtime
- proxy process crash MUST NOT crash or corrupt the execution runtime

Rationale: The proxy component handles raw network traffic and is the highest-risk component in the system. Process isolation limits the blast radius of a proxy component compromise to the proxy boundary. The execution runtime remains protected.

#### Updated Trust Zone Enumeration

The trust zones defined in §1 are extended with the following additions:

- Browser Extension (sandboxed browser context, external input)
- Browser Extension native messaging channel (authenticated, validated)
- Local proxy component (process-isolated, IPC-bounded)
- Bridge agent (process-isolated, authenticated IPC)
- Capture Records (internal intermediate representation, not execution input)
- Imported traffic artifacts (HAR, curl — untrusted by default, identical to imported workflow artifacts)
