# ADR-009 — Packaging & Deployment Model

Status: Accepted
Date: 2026-05-23

---

## Context

traCtl is a local-first platform with a commercial managed services strategy for later phases. The core product promise — same workflow, same behavior everywhere — must hold across local developer machines, self-hosted enterprise environments, and future cloud-managed deployments. Without explicit deployment constraints, execution semantics can silently diverge across topologies, destroying the trust the product is built on.

---

## Decision

### 1 — Deployment Topology Invariant

Deployment topology MUST NOT change canonical workflow semantics.

Local execution, self-hosted execution, and cloud-managed execution MUST produce equivalent execution semantics for the same workflow definition.

This is an absolute invariant. It is the deployment-layer expression of the core product promise.

Timing may vary across topologies. Semantic behavior may not.

Rationale: If cloud deployment changed semantics, "same workflow everywhere" would be false. The local-first product commitment and the commercial managed services strategy both depend on this invariant being maintained.

### 2 — Runtime Portability

Runtime portability across deployment topology classes is required.

Supported topology classes:

- local-first (developer machine, no infrastructure)
- server-hosted (self-managed runtime server)
- enterprise self-hosted (governed organizational deployment)
- cloud-managed (future Phase 8 managed platform)

Architecture MUST NOT assume a single topology class.

Topology-specific optimizations are permitted. Topology-specific semantic behavior is prohibited.

### 3 — Capability Packaging Contract Integrity

Extension and provider packaging MUST preserve capability contract integrity.

Packaging a capability MUST NOT alter:

- declared capability contract
- semantic version
- declared inputs and outputs
- behavioral guarantees
- execution constraints

The packaged and installed version of a capability must be semantically equivalent to its declared contract. Packaging is a distribution concern, not a semantic one.

### 4 — Environment Configuration Boundary

Environment-specific configuration MUST remain external to workflow definitions.

Workflow semantic behavior MUST NOT depend on environment-baked configuration.

Environment-specific values:

- base URLs
- credentials
- environment-specific variables
- infrastructure endpoints

These MUST be supplied through environment configuration artifacts (e.g., `--env dev.yaml`, `--env ci.yaml`), not embedded in workflow definitions.

Rationale: Workflow definitions must be portable artifacts. A workflow definition that embeds environment assumptions cannot be reviewed, version-controlled, and executed consistently across environments.

### 5 — Capability Version Consistency Across Environments

Workflows MAY declare capability version requirements.

Planning MUST validate capability version compatibility against the execution environment.

Version skew between the authoring environment and the execution environment MUST be explicitly detected and reported.

Environment mismatch MUST fail planning with an explicit diagnostic.

Best-effort compatibility guessing across environments is prohibited.

Rationale: The local-versus-CI mismatch scenario is the primary trust failure for traCtl's primary persona. A workflow that behaves differently in CI than locally because of a capability version skew destroys exactly the trust the product is built to create.

---

## Consequences

**Positive:**

- Workflow portability across all deployment topologies
- Local-first promise is architecturally enforced
- Commercial managed platform is semantically equivalent to local execution
- Environment configuration is reviewable and version-controlled separately
- Capability version skew is caught before execution

**Tradeoffs:**

- Environment configuration artifact management overhead
- Capability version compatibility validation across environments adds complexity
- Topology-specific optimizations constrained by semantic equivalence requirement

---

## Alternatives Considered

**Topology-specific execution semantics:** rejected. Violates core product promise. Local-first claim becomes false.

**Environment configuration embedded in workflows:** rejected. Breaks portability. Prevents clean code review. Violates Git-first model.

**Best-effort version compatibility:** rejected. Silent degradation. Nondeterministic behavior. Trust destruction.

**Cloud-only semantic authority:** rejected. Violates local-first architectural commitment. Commercial dependency inversion.

---

## Amendment — 2026-05-24

Amended by: ADR-012 (Multi-Surface Delivery Model), ADR-015 (Packaging & Distribution Strategy)

### Context

ADR-009 defines the deployment topology model with four topology classes (local-first, server-hosted, enterprise self-hosted, cloud-managed). ADR-012 introduces five delivery surfaces, two of which — Container Runtime and Browser Extension — have deployment semantics not covered by the original topology model.

ADR-015 defines the packaging and distribution strategy and introduces specific requirements for container images (immutability, provider version locking) and browser extension deployment (marketplace distribution, enterprise managed deployment).

This amendment extends ADR-009 to cover container runtime parity requirements, CI runner expectations, and browser extension deployment implications.

### Additional Decisions

#### 6 — Container Runtime Deployment Semantics

Container Runtime is a first-class deployment topology (ADR-012 §1).

Container Runtime deployment MUST preserve execution semantic parity with CLI and CI surfaces.

**Immutable execution image requirement:**

Container images are immutable execution environments.

This is an absolute constraint with no exceptions.

Providers and extensions MUST NOT be installed at container runtime.

Provider and extension availability MUST be declared and bundled at image build time.

Rationale: Runtime provider installation creates non-reproducible execution environments. The deployment topology invariant (§1 of this ADR) requires semantic equivalence across runs. A container that installs providers on startup violates this invariant because provider availability becomes a function of network state at run time.

**Provider version locking in containers:**

Container images MUST specify exact provider versions at build time.

Provider version resolution MUST occur at image build time, not at container startup.

Provider version floating in container images is prohibited.

**Environment configuration in containers:**

The environment configuration boundary (§4 of this ADR) applies equally to containers.

Environment-specific values MUST be supplied via environment variables or mounted configuration files.

They MUST NOT be baked into the container image.

Workflow definitions MUST be mountable as volumes, not baked into images.

This preserves the portability guarantee: the same container image + different env config = different environment, not a different image.

#### 7 — CI Runner Expectations

CI surfaces (ADR-012 §1) have specific execution expectations that complement the topology invariant.

CI runner requirements:

- traCtl binary version MUST be explicitly pinned in CI configuration
- workflow definitions MUST be committed to the repository being tested (Git-native, not fetched at runtime)
- environment configuration MUST be supplied via CI secrets management (not embedded in workflow files)
- CI execution MUST produce structured machine-consumable output (JSON, YAML, or TOON)
- CI execution MUST produce deterministic exit codes (0 = pass, non-zero = failure with category)
- CI execution MUST complete without interactive prompts

CI execution MUST NOT:

- require network access to install providers (providers must be pre-installed or bundled)
- require interactive authentication flows
- produce output dependent on terminal capability detection

**Version skew detection:**

The version skew detection requirement (§5 of this ADR) applies to CI with heightened urgency.

CI environments that silently run a different traCtl version than the developer's local environment are the primary trust destruction scenario for the core product promise.

CI pipelines MUST fail explicitly when capability version skew is detected.

#### 8 — Browser Extension Deployment Implications

Browser Extension deployment (ADR-015 §4) does not introduce execution topology. It introduces a capture surface (ADR-012 §5).

Browser Extension deployment implications for this ADR:

- Browser Extension deployment does NOT constitute a traCtl execution topology
- Browser Extension deployment does NOT change execution semantic requirements
- Execution delegation from Browser Extension to Desktop or CLI is subject to the topology invariant: the delegated execution surface MUST preserve execution parity

Enterprise-managed Browser Extension deployment (Group Policy):

- Enterprise-managed extension deployment is a distribution concern (ADR-015)
- Enterprise-managed deployment does not grant additional execution trust
- Trust model for enterprise-deployed extensions remains identical to user-installed extensions (ADR-006 §11)

Rationale: Enterprise deployment management controls installation, not trust. An enterprise-deployed extension operating in a browser sandbox has the same security constraints as a user-installed extension.

---

## Amendment — 2026-05-29

Amended by: ADR-012 Amendment (Web App Surface and Browser-Safe Execution Model)

### Context

The 2026-05-24 decision log recorded: "Web surface = basic HTTP only; agent/proxy
required for CORS bypass."

This was imprecise. It implied the web surface cannot execute without backend
infrastructure, which is incorrect. The correct constraint is that CORS is a
browser-enforced security boundary. It does not apply to all web execution — only
to requests targeting APIs without permissive CORS response headers.

ADR-012 Amendment (§7–§11) formally defines the Web App as a sixth delivery surface
with two modes: basic (WASM bundled, no backend) and extended (connects to Docker
container or hosted server).

The Desktop app is not a web upgrade path. Desktop and Web App serve distinct user
personas. Desktop users install a native app for full functionality. Web App users
want zero installation. CLI alone is not an upgrade path for the browser.

### Additional Decisions

#### 9 — Web App Deployment Topology

The Web App surface constitutes a distinct deployment topology: browser-local.

Browser-local topology characteristics:

- execution runs entirely within the browser via WASM
- no server process, no Docker container, no CLI binary required for basic mode
- storage is IndexedDB (browser-local, no filesystem access)
- network calls use browser fetch (subject to CORS as enforced by the browser)
- extended capability available when a Docker container or server is connected

Browser-local topology is subject to the deployment topology invariant (§1):

- browser-safe capabilities executed via WASM MUST produce semantically equivalent
  results to the same capabilities on Desktop, CLI, CI, or Container surfaces
- timing may vary; semantic behavior may not

#### 10 — Corrected Web Surface Capability Statement

Previous (incorrect):
"Web surface = basic HTTP only; agent/proxy required for CORS bypass."

Corrected:
"Web surface has two modes. Basic mode: browser-safe capability subset executed via
bundled WASM, no backend infrastructure required. Extended mode: full capability via
connection to a Docker container or hosted server. Desktop app and CLI are not upgrade
paths for the Web App."

The browser-safe capability subset is defined in ADR-012 §8.
