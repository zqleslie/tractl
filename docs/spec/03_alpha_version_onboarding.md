# traCtl Alpha Onboarding & User Journey (Canonical)

## Purpose

This document defines the canonical alpha user onboarding journey and behavioral product contract for traCtl.

This is a product behavior document.

It defines how early users interact with traCtl, what assumptions the architecture must preserve, and what workflow expectations become product commitments.

This document informs ADRs, UX implementation, CLI design, provider behavior, and roadmap sequencing.

---

# 1. Product Goal for Alpha

Alpha exists to validate the core product wedge.

Goal:

Enable backend developers to perform progressive backend validation inside their normal workflow without context switching, heavyweight setup, or migration pressure.

Alpha validates:

- desktop-first developer experience
- deterministic execution model
- Git-first artifact ownership
- CLI parity
- dependency-aware workflow execution
- progressive validation adoption
- provider compatibility model foundations

Alpha is not a broad ecosystem release.

---

# 2. First-Time Backend Developer Journey

## Goal

A backend developer should be productive immediately without mandatory setup friction.

## Product Decisions

- no mandatory project creation
- no mandatory workspace initialization
- no mandatory environment setup
- workspace-first interaction model
- artifact-backed persistence model
- explicit save model
- undo/redo editing history
- optional Git ownership

Git-first means Git-compatible, not Git-required.

---

## Alpha Flow

```text
Launch Desktop
    ↓
Fast Start Screen
    - Create Request
    - Open Workflow
    - Run Existing File
    - Open OpenAPI Spec
    ↓
Create Request
    ↓
Enter URL + HTTP Method
    ↓
Optional Auth Configuration
    ↓
Optional Assertion Configuration
    ↓
Run Validation
    ↓
Concise Result Output
    ↓
Explicit Save
    ↓
Optional Git Commit
```

---

## UX Expectations

First-run experience must feel lightweight and immediately actionable.

traCtl must not behave like a heavyweight IDE setup wizard.

Immediate execution capability is required.

Example first-run workflow:

```text
GET https://api.example.com/users
Run
```

without mandatory configuration ceremony.

---

# 3. Workspace and Artifact Model

## Product Decisions

Workspace interaction is primary.

Persistent truth is artifact-based.

Users interact visually.

Underlying workflow state maps to canonical workflow semantics.

Saved artifacts are explicit, reviewable, filesystem-native assets.

---

## Artifact Priorities

Preferred artifact ordering:

1. YAML
2. JSON
3. TOON

### Architectural Clarification

UI internal state must use canonical workflow semantics.

UI must not be implemented as a raw TOON text editor.

Serialization targets:

- YAML
- JSON
- TOON

---

## Save Model

Explicit save only.

No hidden mutation persistence.

Advanced save options:

- Save as YAML
- Save as JSON
- Save as TOON

Default save target:

YAML

---

# 4. Multi-Step Workflow Journey

## Goal

Support realistic dependent backend workflows while preserving usability.

Examples:

- login → token extraction → authenticated request
- create resource → fetch resource → validate state
- dependent API chains

---

## Product Decisions

- visual dependency chaining required
- hybrid UI model
- structured variable extraction
- dependency-aware failure visualization

---

## Alpha Flow

```text
Create Request A
    ↓
Run/Test
    ↓
Extract Variable
    ↓
Create Request B
    ↓
Visually Connect Dependency
    ↓
Reference Extracted Variable
    ↓
Add Assertions
    ↓
Run Workflow
    ↓
Dependency-Aware Results
    ↓
Save Workflow
```

---

## UX Model

Single request workflows:

simple request editor UX.

Multi-step workflows:

hybrid visual workflow UX.

traCtl must not force graph complexity for simple workflows.

---

## Variable Extraction

Structured extraction UX.

Example:

```text
Extract $.token → authToken
```

Alpha avoids scripting-heavy workflow authoring.

---

## Failure Behavior

Dependency-aware failure propagation.

Example:

```text
auth FAILED
billing SKIPPED
metrics PASSED
health PASSED
```

Unrelated executable branches continue.

traCtl defaults to `resilient` execution: independent branches always run to completion. Dependent branches are skipped when their upstream step fails. No active cancellation is issued. `failFast` is available as an opt-in for workflows that require scoped dependency-subgraph cancellation.

---

# 5. Git Workflow Journey

## Product Decisions

- Git-first, not Git-required
- filesystem-native artifacts
- repository-native compatibility
- no mandatory repository structure

---

## Alpha Flow

```text
Save Workflow
    ↓
Prefer Repository Location
    ↓
Commit Artifact
    ↓
Code Review
    ↓
Run Same Workflow via CLI
```

---

## UX Guidance

traCtl may recommend repository ownership.

Example:

"This workflow is not inside a Git repository. Git-native workflows work best inside version-controlled projects."

This is advisory, never blocking.

---

# 6. CLI Journey

## Product Decisions

- no initialization required
- same semantics as desktop
- direct execution model
- concise default output
- structured optional output

---

## CLI Contract

Example:

```bash
tractl run workflow.yaml --env dev.yaml
```

Supported output formats:

```bash
--output json
--output yaml
--output toon
```

Preferred automation format:

JSON

---

## Execution Contract

CLI must preserve desktop behavioral parity.

Core invariant:

Same workflow definition. Same execution command. Equivalent execution semantics.

---

# 7. CI Journey

## Product Decisions

- plain CLI first
- deterministic exit codes
- structured machine output
- no CI-specific execution semantics

---

## Alpha Flow

```bash
tractl run workflow.yaml --env ci.yaml --output json
```

Behavior:

- deterministic exit codes
- structured result output
- same workflow semantics as local execution

GitHub Actions wrappers or CI adapters are future convenience layers.

---

# 8. OpenAPI Public v1 Journey

## Product Decisions

- compatibility execution, not forced migration
- direct provider execution
- fidelity disclosure mandatory
- augmentation overlays supported
- desktop drag/drop support

---

## CLI Flow

Direct execution:

```bash
tractl run --openapi openapi.yaml
```

Extended execution:

```bash
tractl run --openapi openapi.yaml --overlay overlay.yaml
tractl run --openapi openapi.yaml --overlay overlay.yaml --fuzz
```

---

## Product Model

OpenAPI provides compatibility execution foundation.

Overlay augments validation depth.

Overlay may define:

- assertions
- auth behavior scenarios
- variable extraction
- validation stages
- fuzz directives
- performance smoke directives
- execution overrides

This avoids forced migration into native workflows.

---

## Desktop UX

Supported:

- drag/drop OpenAPI spec
- provider compatibility analysis
- fidelity report display
- optional overlay augmentation
- workflow execution
- optional native workflow save

---

# 9. Future External Provider Journey

## Product Decisions

Vendor ecosystem compatibility is installable provider based.

Examples:

- Postman
- Bruno
- future vendor ecosystems

---

## CLI Flow

```bash
tractl provider add postman
tractl run --postman collection.json
```

---

## Behavior

- provider installation
- compatibility analysis
- fidelity disclosure
- execution through provider model

Desktop provider management is future UX scope.

---

# 10. Alpha Success Criteria

Alpha succeeds if:

- first-time backend developers become productive quickly
- workflow authoring feels lightweight
- multi-step workflows feel natural
- Git-first workflows feel practical
- desktop and CLI behavior remain trustworthy
- CI parity is credible
- OpenAPI compatibility feels adoption-friendly
- progressive validation feels incrementally adoptable


---

# Amendment — 2026-05-24

Authority: ADR-012, ADR-013, ADR-015, PRD v2

This amendment extends the alpha onboarding document to address multi-surface onboarding assumptions, CLI onboarding, and the container quickstart. It does not alter existing alpha scope or behavioral contracts.

---

# 8. Multi-Surface Onboarding Model

## Desktop Onboarding (Alpha Primary)

Desktop onboarding remains the primary alpha path as defined in §2–§4.

No change to alpha desktop behavioral contract.

---

## CLI Onboarding (Alpha Parity)

CLI onboarding must be available at alpha alongside Desktop.

CLI onboarding flow:

```text
Install traCtl CLI
    ↓
tractl run workflow.yaml
    ↓
Concise result output
    ↓
Deterministic exit code
```

Installation:

```bash
# macOS / Linux
brew install tractl

# or
curl -fsSL https://install.tractl.softwits.com | sh
```

First run:

```bash
tractl run my-workflow.yaml --env dev.yaml
```

CLI onboarding MUST NOT require Desktop installation.

CLI onboarding MUST NOT require interactive setup prompts.

CLI onboarding behavioral contract:

- install → run → output within 2 minutes for a developer with a workflow file
- no mandatory account creation
- no mandatory cloud dependency

---

## Container Quickstart (v1 Path)

Container onboarding is explicitly NOT an alpha deliverable.

Container quickstart becomes available at v1 with the Container Runtime image.

Anticipated v1 container quickstart:

```bash
docker run --rm \
  -v $(pwd)/workflows:/workflows \
  -v $(pwd)/env:/env \
  tractl/tractl:1.0.0 \
  run /workflows/my-workflow.yaml --env /env/dev.yaml
```

Container quickstart behavioral contract:

- no host installation required
- workflow definitions and env config mounted as volumes
- deterministic exit codes
- structured JSON output by default in container context

---

## Browser Capture Onboarding (v1.5 Future Path)

Browser capture onboarding is explicitly NOT an alpha deliverable.

v1.5 anticipated browser onboarding flow:

```text
Install traCtl Desktop
    ↓
Install traCtl Browser Extension (Chrome)
    ↓
Activate Capture Session in Desktop
    ↓
Browse your application in Chrome as normal
    ↓
Stop Capture
    ↓
Review captured requests in Desktop
    ↓
Save as workflow (YAML)
    ↓
Run workflow
    ↓
Optional: add assertions via overlay
    ↓
Git commit
```

Browser capture onboarding design principles:

- capture requires explicit per-session consent activation
- credentials are masked before captured requests are displayed
- captured workflow is immediately runnable without manual editing required
- save-to-Git is encouraged but not mandatory

---

# 9. Onboarding Assumptions By Surface

| Surface | Alpha | v1 | v1.5 |
|---|---|---|---|
| Desktop | ✓ Primary | ✓ | ✓ |
| CLI | ✓ Parity | ✓ | ✓ |
| CI | Implied (CLI) | ✓ Explicit | ✓ |
| Container | ✗ | ✓ | ✓ |
| Browser Extension | ✗ | ✗ | ✓ |

Alpha onboarding does NOT assume container or browser extension availability.

Onboarding documentation MUST NOT reference features from later phases as if they are available at alpha.
