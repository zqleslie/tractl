# ADR-014 — Observability & Traffic Diagnostics Model

Status: Accepted
Date: 2026-05-24

---

## Context

traCtl executes backend workflows and validates their behavior. The execution result — pass or fail — is necessary but insufficient for a developer workflow tool. Developers need to understand what happened during execution: where time was spent, why a request failed, which dependency in a chain was slow, and what the actual wire behavior looked like.

The existing diagnostics model (ADR-003 §3) defines planner output as human-readable and diagnostics hooks at the execution context boundary. This is a correct foundation but scoped to execution mechanics only. It does not define the observability contract for network-layer behavior, timing decomposition, dependency visualization, or credential-safe request inspection.

Without a formal observability model, diagnostics accumulate as ad-hoc implementation decisions, credential-unsafe inspection surfaces, and inconsistent output structures across surfaces.

This ADR defines the canonical observability model for traCtl, the data available at each execution phase, timing decomposition requirements, the dependency waterfall model, and explicit non-goals.

---

## Decision

### 1 — Observability Boundary

traCtl's observability scope is execution-context diagnostics.

traCtl observes what happens inside its own execution context.

traCtl does NOT observe:

- raw network packets
- OS-level socket behavior below the HTTP client
- processes or services outside its execution context
- network infrastructure behavior outside a request lifecycle

This is an explicit product boundary.

Observability is the answer to: "what did traCtl do, and what happened as a result?"

Observability is NOT the answer to: "what is happening on my network?"

---

### 2 — Request Timeline Model

Every HTTP request executed by the traCtl runtime MUST capture a Request Timeline.

**Mandatory timing segments (per request):**

| Segment | Definition |
|---|---|
| DNS Resolution | Time from request initiation to DNS lookup completion |
| TCP Connect | Time from DNS completion to TCP connection established |
| TLS Handshake | Time from TCP connect to TLS negotiation complete (HTTPS only) |
| Request Sent | Time from connection ready to last byte of request sent |
| Time to First Byte (TTFB) | Time from request sent to first byte of response received |
| Response Transfer | Time from first byte received to last byte received |
| Total Duration | End-to-end elapsed time for the request |

**Derived metrics (computed, not captured):**

- Latency attribution: percentage of total duration per segment
- Dominant segment identification: the segment consuming the largest proportion of total duration

Timing capture MUST use monotonic clock sources.

Wall-clock time MUST be captured separately for display purposes and MUST NOT be used for duration calculations.

---

### 3 — Retry and Redirect Observability

Retries and redirects are first-class observable events, not implementation noise.

**Retry observability:**

Each retry attempt MUST produce:

- attempt number
- trigger reason (timeout, error code, policy rule)
- elapsed time at retry decision point
- full Request Timeline for the retry attempt
- backoff duration (if applicable)

**Redirect observability:**

Each redirect hop MUST produce:

- hop number
- redirect source URL
- redirect target URL
- redirect status code
- response headers at redirect hop
- Request Timeline for the hop

Redirect chains MUST be surfaced in their entirety.

Silent redirect following without observability data is prohibited.

---

### 4 — Dependency Waterfall Model

Multi-step workflows produce a Dependency Waterfall.

The Dependency Waterfall is a structured representation of how workflow steps executed relative to each other in time and dependency order.

**Waterfall structure:**

```
Step A ────────────────────────────────────
                 Step B ──────────────────────────
                                  Step C ──────────
         Step D ──────────
```

Each row represents one workflow step.

Each segment within a row represents timing phases within that step.

Step ordering reflects actual execution sequencing, not authoring order.

**Waterfall data per step:**

- step identifier
- dependency edges (which steps this step waited for)
- wait duration (time spent waiting for dependencies)
- execution start timestamp (relative to workflow start)
- Request Timeline segments
- assertion evaluation timing
- extract evaluation timing
- total step duration
- step outcome (pass, fail, skipped)

**Parallel branches** MUST be visually and structurally distinguishable from sequential execution.

The Dependency Waterfall MUST be producible as:

- structured data (JSON, YAML)
- human-readable text form (for CLI output)
- visual representation (for Desktop UX)

---

### 5 — Execution Provenance Traces

Every execution MUST produce an Execution Provenance Trace.

The Execution Provenance Trace records the full lifecycle of a workflow execution as a structured event sequence.

**Mandatory trace events:**

- workflow execution initiated (timestamp, workflow id, environment)
- normalization completed (provenance from ADR-013 §5 if applicable)
- validation completed (outcome)
- planning completed (plan identifier)
- compilation completed
- step execution started (step id, timestamp)
- request dispatched (request id, method, URL, timestamp)
- response received (request id, status, timestamp)
- assertion evaluated (assertion id, outcome)
- extract evaluated (extract id, variable, value placeholder)
- step execution completed (step id, outcome, duration)
- workflow execution completed (outcome, total duration)

**Provenance trace MUST:**

- use monotonic timestamps relative to workflow execution start
- carry workflow-scoped trace identifier
- carry step-scoped span identifiers
- be exportable as structured data

**Provenance trace MUST NOT:**

- expose unmasked credential values
- expose unmasked extracted variable values (placeholder references only)
- include raw response bodies by default (opt-in only, subject to credential safety)

---

### 6 — Credential-Safe Inspection

All observability output is credential-safe by default.

Credential safety rules for observability:

- Authorization header values MUST be masked in all output
- API key header values MUST be masked in all output
- Cookie values MUST be masked unless explicitly opted in
- Query parameter values matching credential patterns MUST be masked
- Request body fields matching credential patterns MUST be masked
- Extracted variable values MUST be masked unless inspection mode is explicitly activated

**Masking representation:**

Masked values MUST be replaced with typed placeholder tokens, not empty strings.

Examples:

```
Authorization: Bearer [masked:bearer_token]
X-API-Key: [masked:api_key]
```

Empty string masking is prohibited. It prevents distinguishing "no value" from "masked value."

**Inspection mode:**

An explicit inspection mode MAY be activated by the user in interactive surfaces (Desktop only).

Inspection mode MUST:

- require explicit user activation
- be scoped to a single execution session
- not persist across sessions
- display a visible indicator when active
- be prohibited in CI and Container surfaces

Rationale: Execution diagnostics and provenance traces are often shared for debugging. Credential-safe defaults prevent accidental credential disclosure in shared diagnostics output.

---

### 7 — Observability Output Contracts

Observability data MUST be available in the following forms.

**Structured output (all surfaces):**

- JSON format: complete observability data as structured JSON
- YAML format: complete observability data as structured YAML

**Human-readable summary (CLI, Desktop):**

- concise timing summary per step
- dominant timing segment highlighted
- retry and redirect counts
- dependency waterfall in text form

**Rich visual (Desktop only):**

- interactive waterfall diagram
- timing breakdown charts
- step-level drill-down
- dependency visualization

Structured output format MUST be deterministic.

Same execution → same structured output shape (values may differ by timing).

---

### 8 — Explicit Non-Goals

The following are explicitly outside traCtl's observability scope.

**Packet analysis:**

traCtl does not capture raw network packets.

traCtl does not perform deep packet inspection.

traCtl does not replace Wireshark, tcpdump, or equivalent tools.

**Infrastructure monitoring:**

traCtl does not monitor service health continuously.

traCtl does not alert on infrastructure failures.

traCtl does not replace APM platforms.

**Log aggregation:**

traCtl does not aggregate application logs.

traCtl does not provide centralized logging.

**Security scanning:**

Execution-context observability does not constitute a security scan.

Credential masking is a safety feature, not a security scanner.

Rationale: Explicit non-goals are as important as goals. traCtl's observability model must remain scoped to execution context or it becomes a general-purpose monitoring platform with an unbounded scope. The product moat is the execution engine, not the observability layer.

---

## Consequences

**Positive:**

- Developers understand exactly what happened during execution without external tools
- Timing decomposition surfaces actionable latency information
- Dependency waterfall makes workflow orchestration behavior visible
- Credential-safe defaults prevent accidental secret disclosure
- Structured observability output enables CI integration and external tooling

**Tradeoffs:**

- Request Timeline instrumentation adds per-request overhead
- Structured trace data volume grows with workflow complexity
- Inspection mode adds UX complexity (activation, session scoping)
- Waterfall visualization complexity for large parallel workflows

---

## Alternatives Considered

**Full packet capture:** rejected. Requires OS-level privileges. Scopes the product as a network analyzer. Conflicts with explicit product positioning (PRD §5).

**OpenTelemetry native output:** deferred. May be appropriate as an extension in later phases. Not MVP-critical. Premature standardization before execution semantics are proven.

**Per-surface observability models:** rejected. Inconsistency between surfaces degrades developer trust. Unified model with surface-specific rendering is correct.

**Opt-in credential masking:** rejected. Credential disclosure risk is too high. Safe-by-default is the only acceptable posture.

**Aggregated timing only (no per-segment breakdown):** rejected. Aggregate duration is insufficient for latency diagnosis. DNS vs TCP vs TLS attribution is the information developers actually need.
