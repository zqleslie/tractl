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
