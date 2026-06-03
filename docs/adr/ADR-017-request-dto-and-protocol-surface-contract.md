# ADR-017 — Request DTO and Protocol Surface Contract

Status: Accepted
Date: 2026-06-02

---

## Context

traCtl supports multiple delivery surfaces — CLI, Web WASM, Desktop, Local API — all
executing the same canonical engine. The engine operates on `*spec.TraCtlSpec` as its
sole input. However, interactive surfaces and external tools need to construct requests
without knowing the full spec structure.

Two unresolved questions arose during pre-release design:

**1. DTO placement and mapper ownership.**
`RequestDef` exists in `internal/localapi/types.go` as the single-request DTO. An
equivalent `RequestDef` type exists in `frontend/src/types/requestDef.ts`. These are
maintained manually in sync with no code generation. The React frontend routes
single-request execution through the Go mapper (`request_mapper.go`), which introduces
an unnecessary coupling: the frontend depends on the Go mapper for a transformation it
can perform itself using its own TypeScript types.

**2. Protocol surface contract.**
`RequestDef` is HTTP-centric: `Method string` (GET/POST/DELETE), `URL string`,
`Body.encoding` (json/form/multipart). traCtl's execution model (`RequestDescriptor`)
is already protocol-neutral — it uses `Protocol string`, `Operation string`, and
`Metadata map[string]interface{}`. The DTO does not reflect this. As GraphQL, SOAP,
OData, and future protocols are added, a `RequestDef` with no `Protocol` field becomes
a breaking API change for every external caller that has already integrated.

Additionally, three protocol-adjacent questions required formal decisions before the
first release: WSDL support scope, HTTP/3 support scope, and SOAP version handling.

---

## Decision

### 1 — RequestDef Is the Public API Contract for Single-Request Execution

`RequestDef` (`internal/localapi/types.go`) is not merely a frontend DTO. It is the
**public API contract** for any external tool that wants to trigger a single request
without constructing a full `TraCtlSpec` document.

Consumers include:

- IDE extensions (VS Code, JetBrains)
- CI pipeline integrations
- External scripts and automation tools
- Any HTTP client calling `POST /api/v1/run`

`RequestDef` hides all engine internals from these callers. They work with
HTTP-familiar concepts (URL, method, headers, assertions) rather than
`schemaVersion`, `capabilities`, `workflows[]`, `steps[]`, and `RequestDescriptor`.

`RequestDef` MUST remain stable as a public API contract. Breaking changes require
a major version increment of the Local API.

### 2 — Protocol Field Required Pre-Release

`RequestDef` MUST include a `Protocol string` field before the first public release.

```go
type RequestDef struct {
    Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
    // ... existing fields unchanged
}
```

Default behaviour when `Protocol` is empty: `"http"`.

Rationale: Adding `Protocol` after the first release is a breaking change for every
existing external integration. Adding it now with a default of `"http"` is fully
backward-compatible. Omitting it now forecloses the ability to add non-HTTP protocols
without a breaking API change.

The corresponding TypeScript type MUST be updated simultaneously:

```typescript
export type Protocol = 'http' | 'graphql' | 'soap' | 'odata'
export interface RequestDef {
    protocol?: Protocol  // default 'http' when absent
    // ... existing fields unchanged
}
```

### 3 — Mapper Ownership Split

Two distinct mapper paths serve two distinct caller populations.

**Go mapper (`request_mapper.go`) — for external callers only.**

The Go mapper (`requestDefToDocument()`) converts `RequestDef` to a single-step
`TraCtlSpec` YAML document. It exists for external HTTP callers who cannot construct
YAML. It MUST remain in `internal/localapi/`.

**TypeScript mapper — for the React frontend.**

The React frontend MUST own its own TypeScript mapper that converts `RequestDef`
to a spec-shaped JSON object and sends it as `WorkflowRunRequest`. The frontend
MUST NOT route single-request execution through the Go mapper.

```
Current (incorrect for frontend):
  RequestDef → JSON → Go: request_mapper.go → YAML → parser → *TraCtlSpec

Correct for frontend:
  RequestDef → TypeScript mapper → spec JSON → WorkflowRunRequest → parser → *TraCtlSpec

Correct for external callers:
  RequestDef → JSON → Go: request_mapper.go → YAML → parser → *TraCtlSpec
```

This eliminates the manual sync risk between `localapi.RequestDef` (Go) and
`types/requestDef.ts` (TypeScript) for the mapping logic. Both type definitions
still exist and must be kept in sync — only the mapping responsibility changes.

Rationale: The frontend already has full TypeScript type information. It is better
positioned than Go to produce spec-shaped JSON from its own form state. The Go
mapper serves a different caller population and should not be on the frontend's
critical path.

### 4 — HTTP-Based Protocol Classification

All of REST, GraphQL, OData, SOAP, and JSON-RPC use HTTP as their transport layer.
The treatment of each is explicitly classified:

**REST — no changes required.**

Standard HTTP. Fully supported. `Protocol: "http"`, existing method/URL/body/headers.

**OData — no changes required.**

OData is REST with `$`-prefixed query parameters (`$filter`, `$select`, `$expand`,
`$orderby`, `$top`, `$skip`). These are standard URL query parameters handled by
the existing `Params []KVRow` field. OData responses are JSON addressable by
existing JSONPath assertions. No protocol-specific engine changes needed.

```
Protocol: "http"
Params:
  - { key: "$filter", value: "Name eq 'Alice'" }
  - { key: "$select", value: "Id,Name" }
```

**GraphQL — first-class body encoding, pre-release.**

GraphQL requires a structured request body: `{query, variables, operationName}`.
It is common enough to warrant first-class support before the first release.

`BodyDef.encoding` MUST accept `"graphql"` as a valid value. When `encoding` is
`"graphql"`, the executor MUST:

1. Serialize `content` as `{"query": ..., "variables": ..., "operationName": ...}`
2. Set `Content-Type: application/json`
3. Use HTTP POST regardless of the `Method` field

GraphQL responses are standard JSON. Existing JSONPath assertions (`body.data.*`,
`body.errors.0.message`) work without modification. No response-side engine changes
are needed.

**JSON-RPC — no changes required.**

JSON-RPC is HTTP POST with `{jsonrpc, method, params, id}` as the body. Use
`encoding: "json"` with the JSON-RPC envelope as `content`. Response assertions
use existing JSONPath (`body.result.*`, `body.error.*`). No protocol-specific
support needed.

**SOAP — deferred to v1.1.**

SOAP requires two engine capabilities not yet implemented:

1. `encoding: "xml"` — the executor must send raw XML bodies
2. XPath extractor — SOAP responses are XML; `gjson` (JSONPath) cannot evaluate them

Both SOAP 1.1 (`Content-Type: text/xml`, separate `SOAPAction` header) and
SOAP 1.2 (`Content-Type: application/soap+xml; action="..."`) are handled
identically at the engine level — the version is expressed entirely through
headers and XML namespace declarations authored by the workflow developer.
No `soapVersion` field in `RequestDef` is required or permitted.

SOAP is deferred to v1.1. The engine changes (XML body serialisation, XPath
assertion evaluation) deserve a dedicated implementation milestone.

### 5 — BodyDef Encoding Extensions

`BodyDef.encoding` MUST accept the following values in the first release:

```
"json"       — application/json body (existing)
"form"       — application/x-www-form-urlencoded (existing)
"multipart"  — multipart/form-data (existing)
"raw"        — raw bytes with explicit Content-Type (existing)
"binary"     — binary payload (existing)
"none"       — no body (existing)
"graphql"    — GraphQL query envelope (NEW, pre-release)
"xml"        — raw XML body (NEW, v1.1 — SOAP support)
```

The executor MUST reject unknown encoding values at runtime with an explicit error.

### 6 — WSDL Is an Import Format, Not an Execution Protocol

WSDL (Web Services Description Language) describes SOAP service contracts. It is
a **spec provider** (alongside OpenAPI, Protocol Buffers, GraphQL SDL, AsyncAPI)
not a runtime execution protocol.

WSDL support means: read a WSDL file and auto-generate traCtl workflow steps from
it. The generated steps execute via SOAP over HTTP. WSDL is never declared in
`capabilities[]` as a protocol — it is a tooling concern.

WSDL import is a Phase 4+ feature (provider adapter tier, consistent with the
OpenAPI import capability). It is explicitly out of scope for the first release.

### 7 — HTTP/3 Deferral with Capability Guard

HTTP/3 uses QUIC over UDP. Go's standard `net/http` supports HTTP/1.1 and HTTP/2
only. HTTP/3 requires the third-party `quic-go` library and is not in the Go
standard library.

**Exception — WASM surface:** Go compiled to WebAssembly delegates network calls
to the browser's `fetch()` API. All major browsers (Chrome 87+, Firefox 88+,
Safari 14+) negotiate HTTP/3 natively when the server supports it. The WASM surface
therefore inherits HTTP/3 support from the browser without any Go-side changes.

**All other surfaces (CLI, Desktop, Local API):** HTTP/3 is not supported. HTTP/1.1
and HTTP/2 are the ceiling.

**Required pre-release:** The planner MUST reject specs that declare
`transport.version: "h3"` on non-WASM surfaces with an explicit capability error
rather than silently falling back to HTTP/2.

```
capability negotiation error:
  transport.version "h3" is not supported on this surface.
  HTTP/3 is available on the Web WASM surface via the browser's fetch API.
  CLI, Desktop, and Local API surfaces support HTTP/1.1 and HTTP/2 only.
```

HTTP/3 support for non-WASM surfaces is a post-release feature dependent on
`quic-go` integration. It MUST be declared as a capability (`transport.h3@1`)
when implemented, following the capability contract model (ADR-011).

---

## Consequences

**Positive:**

- `Protocol` field addition before v1 future-proofs the external API contract
  permanently without a breaking change
- Frontend owns its mapper — eliminates one source of Go/TypeScript sync debt
- Go mapper remains as the stable API surface for IDE plugins and external tools
- OData and JSON-RPC require zero engine changes — they are just HTTP
- GraphQL gets first-class support in the first release
- SOAP deferral gives the XML/XPath implementation the space it needs
- HTTP/3 behaviour is deterministic across surfaces — no silent degradation
- WSDL scope is formally bounded — no scope creep into the execution layer

**Tradeoffs:**

- Two mapper implementations (Go and TypeScript) must produce equivalent output
- TypeScript mapper is a new maintenance surface with no Go-side test coverage
- GraphQL executor changes required pre-release (small but real scope addition)
- Manual type sync (`localapi.RequestDef` ↔ `types/requestDef.ts`) still required
  for field additions — code generation is not yet implemented

---

## Alternatives Considered

**Single mapper in Go, frontend always routes through it:** rejected. Couples the
frontend to the Go mapper for a transformation the frontend can perform with its
own types. Preserves the manual sync risk. No benefit over TypeScript-owned mapping
for the frontend path.

**Single mapper in TypeScript, external callers send full spec JSON:** rejected.
External callers (IDE plugins, curl, CI scripts) would need to construct a full
`TraCtlSpec`-shaped document to call a single HTTP request. This is a significantly
worse developer experience. `RequestDef` exists to protect these callers from engine
internals.

**protocol: "odata" as a named protocol:** rejected. OData is REST. Declaring it
as a distinct protocol would add engine complexity with no benefit — OData's
conventions (`$filter`, `$select`, etc.) are query parameter conventions handled
entirely by the existing `Params []KVRow` mechanism.

**SOAP in first release:** rejected. XML body serialisation and XPath assertion
evaluation are meaningful engine additions. Rushing them into the first release
risks getting the XPath assertion model wrong in ways that are hard to change once
the API contract is published. SOAP usage is declining in new API development.
Deferring to v1.1 is the lower-risk choice.

**HTTP/3 via quic-go pre-release:** rejected. quic-go is a substantial dependency.
HTTP/3 provides no correctness benefit for API testing — it is a performance
optimisation. The WASM surface already handles the most common HTTP/3 use case
via browser fetch. Non-WASM HTTP/3 is post-release, demand-driven.

**No Protocol field until a non-HTTP protocol is actually added:** rejected.
Adding `Protocol` after v1 requires all external callers to update their integration
code. The cost of adding it now (near zero — one field with a default) is orders of
magnitude lower than the cost of a breaking API change after release.

---

## Amendment — 2026-06-02

Amended by: RequestDef as Shared Authoring Contract for Requests and Workflow Steps

### Context

ADR-017 §1 defines `RequestDef` as "the public API contract for single-request
execution." ADR-016 §4 defines the single-request editor as using `RequestFormState`
as its authoring type, which serializes to `RequestDef` for API submission.

A gap existed: the workflow canvas step editor also uses `RequestFormState`
internally (the step detail panel is the same form-based UI as the single-request
editor), but this was not formally recognised as sharing the same authoring contract.
The result was that `WorkflowStep` — a display projection — was mistakenly treated
as the step authoring model, causing silent data loss on serialization (see
ADR-016 Amendment 2026-06-02).

### Decision

#### §3 Extension — RequestDef Is the Authoring Contract for Both Request Types

`RequestDef` and `RequestFormState` are the authoring contracts for **both**
single-request execution and workflow request step editing. The two are the same
concept at the authoring layer. The distinction is at the execution layer only:

| | Single-request editor | Workflow step editor |
|---|---|---|
| Authoring type | `RequestFormState` → `RequestDef` | `RequestFormState` → `RequestDef` |
| Mapper | `requestFormStateToTraCtlSpec()` | `requestFormStateToTraCtlSpec()` (same) |
| Output | Single-step `TraCtlSpecDocument` | `TraCtlStep` merged into multi-step `TraCtlSpecDocument` |
| Execution | Immediate, single step | As part of workflow, respects `dependsOn` |

The TypeScript mapper `requestFormStateToTraCtlSpec()` MUST be the single
serialization path for both cases. It MUST NOT be forked or duplicated for the
workflow step case.

The Go mapper `requestDefToDocument()` (`request_mapper.go`) serves external callers
via `POST /api/v1/run` only. It is not involved in workflow step authoring.

#### Shared Mapper Invariant

The following invariant MUST hold:

> For any `RequestFormState` value, `requestFormStateToTraCtlSpec()` produces a
> `TraCtlStep` whose `request`, `assertions`, `extracts`, `hooks`, `timeout`, and
> `retry` fields are semantically equivalent whether the step is embedded in a
> single-step document or a multi-step workflow document.

The only fields that differ between the two execution contexts are workflow-level
concerns (`dependsOn`, `id` assignment, `failurePolicy`) — not the request content
itself.

#### Auth in Workflow Steps

Auth (`RequestFormState.auth`) is resolved into request headers by the TypeScript
mapper for both single-request and workflow step execution on the WASM surface
(ADR-017 §3, ADR-016 Amendment 2026-06-02 §3). When a workflow YAML file is
authored externally (not via the editor), auth credentials appear as resolved
`Authorization` or custom headers in `request.headers`. The editor restores
`hasAuth: true` from header presence on load but cannot recover the original
auth type or credential value — this is a known limitation of plain-text YAML
as the storage format.

### Consequences

**Positive:**

- One mapper for both cases — no divergence in serialization behaviour
- `RequestDef` / `RequestFormState` formally covers 100% of request authoring
  surface in the UI
- Workflow step editing bugs are caught by single-request editor tests and vice versa

**Tradeoffs:**

- The shared mapper must handle workflow-step-specific concerns (`dependsOn` is
  not a `RequestFormState` field — it is a canvas-level concern passed separately
  to the step assembly function)
- External YAML authors must be aware that `auth` is an editor abstraction; in YAML
  the resolved header is the canonical form

---

## Amendment — 2026-06-02 (B)

Amended by: Go Is the Sole Transformation Layer — TypeScript Is Presentation Only

### Context

ADR-017 §3 (original) stated: "The React frontend MUST own its own TypeScript mapper
that converts `RequestDef` to a spec-shaped JSON object."

ADR-017 Amendment 2026-06-02 extended this, declaring the TypeScript mapper
`requestFormStateToTraCtlSpec()` as the single serialization path for both
single-request and workflow step execution on the WASM surface.

Both decisions were incorrect. The TypeScript mapper was doing business logic that
belongs exclusively in Go:

- ISO 8601 duration formatting (`timeoutMs: 30000 → "PT30S"`)
- Auth resolution (bearer/basic/apikey credentials → `Authorization` header)
- GraphQL body serialization (`query`/`variables`/`operationName` envelope)
- Form row encoding (KVRow[] → `application/x-www-form-urlencoded`)

This logic already exists in the Go mapper (`requestDefToDocument()` in
`internal/localapi/request_mapper.go`). The TypeScript mapper was a second
implementation of the same logic, introducing:

1. **Parity risk** — subtle behavioural differences (e.g. `Math.round` vs integer
   division for timeout conversion) that produce different spec documents from the
   same user input depending on which surface ran the request.

2. **Future surface cost** — every new delivery surface (mobile, IDE extension,
   desktop companion) would need to reimplement the same transformation logic.
   The engine result is consistent; the input construction is not.

3. **Wrong responsibility** — TypeScript's job is to render UI, manage local state,
   and translate form state into a `RequestDef` DTO. Spec document construction is
   an engine concern, not a presentation concern.

The WASM surface was the specific driver: because the browser runs Go compiled to
WebAssembly, there is no reason to build the spec document in TypeScript before
calling the engine. The Go code is already there, inside the WASM bundle.

### Decision

#### §3 Superseded — Go Owns All Transformation Logic for All Surfaces

The original §3 and the 2026-06-02 §3 extension are both superseded by this
amendment.

**The Go mapper is the single transformation path for all delivery surfaces,
including WASM.**

The role of the TypeScript layer is strictly:

1. **Collect user input** — form state (`RequestFormState`)
2. **Translate to `RequestDef`** — thin field mapping only: rename UI encoding labels
   (`'JSON'` → `'json'`, `'Form data'` → `'form'`), pass numeric values as-is
   (`timeoutMs: 30000`, not `"PT30S"`). No duration formatting. No auth injection.
   No protocol-specific body construction.
3. **Dispatch to the appropriate transport** — Wails binding (Desktop), WASM bridge
   (Web Tier 1), or HTTP (Web Tier 2). The transport receives `RequestDef` JSON.
4. **Display results exactly as the engine produces them** — no reformatting of
   engine output in TypeScript.

**The WASM bridge MUST expose a `runRequest` entry point** that accepts `RequestDef`
JSON, calls Go's `RunRequestDef()` internally (which calls `requestDefToDocument()`
and then the engine), and returns the `RunResult` JSON.

```
Web WASM (corrected):
  RequestFormState
    → [thin TS: form state → RequestDef]
    → window.tractl.runRequest(JSON.stringify(requestDef))
        → Go (WASM): localapi.RunRequestDef(def)
            → requestDefToDocument(def)     ← all logic here, in Go
            → engine.RunDocument(spec)
            → RunResult

Desktop:
  RequestFormState
    → [thin TS: form state → RequestDef]
    → Wails.RunRequestDef(requestDef)
        → Go: localapi.RunRequestDef(def)
            → requestDefToDocument(def)
            → engine.RunDocument(spec)
            → RunResult

Server / external callers (unchanged):
  RequestDef → POST /api/v1/run
    → Go: localapi.RunRequestDef(def)
        → requestDefToDocument(def)
        → engine.RunDocument(spec)
        → RunResult
```

All three paths call the same Go function. Duration formatting, auth injection,
GraphQL serialisation, and all other spec construction logic live in exactly one
place.

#### TypeScript Presentation Boundary

The following MUST be in TypeScript (presentation and storage):

- Form state management (`RequestFormState`, Zustand stores)
- Thin `RequestFormState → RequestDef` field mapping (no business logic)
- Transport dispatch (`RequestDef` JSON to Wails / WASM bridge / HTTP fetch)
- Result rendering (display what the engine returns, unchanged)
- Local persistence (IndexedDB for Web Tier 1, auto-save coordination)
- Code view / spec preview — `requestDefToDocumentObject` is permitted here
  because showing the user a preview of the generated spec is a presentation
  concern, not an execution concern. It MUST NOT be on the execution code path.

The following MUST NOT be in TypeScript:

- ISO 8601 duration formatting
- Auth credential resolution into headers
- Protocol-specific body serialisation (GraphQL, form encoding, multipart)
- Any logic that is also present in `request_mapper.go`

#### `requestDefToDocument.ts` Scope Restriction

`frontend/src/platform/mappers/requestDefToDocument.ts` is retained for the code
view / spec preview use case only. It MUST NOT be imported by any execution runner
(`webRequestExecutionRunner`, `desktopRequestExecutionRunner`, or any future runner).
Its role is: given a `RequestDef`, show the user what spec YAML the engine will
receive. It is a display utility, not a transformation pipeline.

#### Protocol Support Is Deferred to Go

Protocol-specific handling (GraphQL body envelope, form encoding, multipart
boundaries, future SOAP XML) is implemented in Go only — in `requestDefToDocument()`
and its helpers. TypeScript sends `body.encoding: "graphql"` and `body.content`
as-is in the `RequestDef`; Go handles what that means at execution time.

This directly resolves the concern that added protocol support would require
parallel TypeScript implementations. It never will. Protocols are added in Go
once and available on all surfaces immediately.

#### Consistency Invariant

The following invariant MUST hold across all delivery surfaces:

> For any `RequestDef` value, `localapi.RunRequestDef(def)` on any surface
> (CLI, Desktop, Web WASM, Web Server) produces semantically equivalent
> `RunResult` output. Surface-level differences (CORS, socket timing availability)
> are declared exceptions per ADR-012, not behavioural differences in request
> construction.

### Consequences

**Positive:**

- Single transformation implementation in Go — no parity risk between surfaces
- New delivery surfaces (mobile, IDE extension) send `RequestDef` JSON and get
  consistent results without implementing any mapping logic
- Protocol support (GraphQL, SOAP, future) added in Go once, available everywhere
- TypeScript codebase is smaller and has a clear, bounded responsibility
- The existing `RunRequestDef()` function in `mapper.go` is already correct —
  this amendment primarily removes TypeScript code, not adds Go code

**Tradeoffs:**

- WASM bridge requires a new `runRequest` entry point — small Go addition
- `requestDefToDocument.ts` remains in the codebase for preview use, creating
  a risk that future engineers use it on the execution path — the import boundary
  stated above is the guard against this
- The `requestFormStateToTraCtlSpec()` TypeScript mapper, and the related
  `requestDefToDocumentObject()` function, are demoted from execution-path code
  to preview utilities — any tests that were validating execution semantics via
  these functions need to be reframed as preview tests

### Superseded Content

The following is superseded by this amendment:

- ADR-017 §3 original: "The React frontend MUST own its own TypeScript mapper
  that converts `RequestDef` to a spec-shaped JSON object and sends it as
  `WorkflowRunRequest`. The frontend MUST NOT route single-request execution
  through the Go mapper."
- ADR-017 Amendment 2026-06-02 §3 extension: "The TypeScript mapper
  `requestFormStateToTraCtlSpec()` MUST be the single serialization path for both
  cases."
- ADR-017 original Tradeoffs: "Two mapper implementations (Go and TypeScript) must
  produce equivalent output" and "TypeScript mapper is a new maintenance surface
  with no Go-side test coverage" — both eliminated by this amendment.
- ADR-017 original Alternatives Considered: "Single mapper in Go, frontend always
  routes through it: rejected. Couples the frontend to the Go mapper." — this
  rejection is itself reversed. The frontend does route through the Go mapper; it
  does so via the WASM bridge `runRequest` entry point, which is not a coupling
  problem because the bridge is the defined interface between the two layers.
