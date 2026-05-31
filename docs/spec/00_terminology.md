# traCtl Domain Terminology

## Purpose

This document defines the canonical terms and groupings used across all traCtl documents.

When these category names appear in the PRD, HLD, roadmap, or engineering specs — this is what they mean.

---

## 1. Authoring Formats

How traCtl groups the formats used to write native workflows.

| traCtl Category | Members | What it means |
|---|---|---|
| **Native Authoring Format** | TOON, YAML, JSON | First-class formats for writing traCtl workflows directly |

| Format | Description | Default |
|---|---|---|
| **TOON** | traCtl's own human-readable format. Purpose-built for workflow authoring. Long-term preferred format. | No |
| **YAML** | Standard structured format. Most familiar to developers. Default save format in Alpha. | Yes |
| **JSON** | Machine-friendly structured format. Preferred for programmatic generation. | No |

> All three formats produce equivalent traCtlSpec after normalisation. Choosing a format is an authoring preference, not an execution difference.

---

## 2. Protocol Categories

How traCtl groups communication protocols.

| traCtl Category | Members | What it means |
|---|---|---|
| **Request-Response** | REST, GraphQL, gRPC, SOAP | Client sends a request, server returns a response |
| **Streaming** | gRPC streaming, GraphQL subscriptions, WebSocket | Persistent connection, continuous data flow |
| **Event / Messaging** | MQTT, AMQP, Kafka | Asynchronous, broker-mediated message passing |
| **Raw Transport** | TCP, UDP | No application-layer protocol — raw bytes |

> HTTP/1.1 and HTTP/2 are transports, not protocols in this grouping. REST and GraphQL ride on top of them.

---

## 3. Provider Categories

How traCtl groups everything it can ingest or execute against.

### 3a. Native Spec Providers

Specification languages that traCtl understands natively. These describe a service. traCtl normalises them into traCtlSpec with a declared fidelity boundary.

| Spec | Describes |
|---|---|
| OpenAPI | REST APIs |
| WSDL | SOAP services |
| GraphQL SDL | GraphQL schemas |
| Protocol Buffers (`.proto`) | gRPC services |
| AsyncAPI | Event-driven / messaging APIs |

### 3b. External Compatibility Providers

Ecosystem tool artifacts that teams already have. Installable. traCtl executes them without requiring migration.

| Provider | Ingests |
|---|---|
| Postman | Collection v2.1 JSON |
| Bruno | `.bru` files |
| Insomnia | Export JSON |
| HAR | Browser capture files |
| curl | Command strings |

### 3c. Delegated Specialist Providers

Specialist engines traCtl delegates to for depth outside its core scope. traCtl orchestrates; the engine executes.

| Provider | Delegates To | Capability |
|---|---|---|
| Schemathesis | Schemathesis | Schema-driven fuzz testing |
| k6 | k6 | Advanced load testing |
| OWASP ZAP | OWASP ZAP | Deep security scanning |

---

## 4. Validation Categories

How traCtl groups what it validates. Each stage is independently adoptable — never all-or-nothing.

| Stage | traCtl Category | What is checked | Ownership |
|---|---|---|---|
| 1 | **Correctness** | Status codes, response shape, field types, required fields | Native |
| 2 | **Auth Behavior** | Token rejection, scope enforcement, permission boundaries, missing credential handling | Native |
| 3 | **Contract Drift** | Response shape vs declared spec or prior baseline | Native |
| 4 | **Resilience** | Timeout behavior, upstream error handling, retry semantics, failure-mode responses | Native |
| 5 | **Performance Smoke** | P99 latency thresholds, basic concurrency behavior | Native |
| Future | **Fuzz** | Schema boundary violations, malformed input handling | Delegated |
| Future | **Deep Security** | Vulnerability scanning, injection, auth bypass | Delegated |
| Future | **Deep Load** | High-concurrency load engineering | Delegated |

> Stages 1–5 are native traCtl capabilities. Fuzz, Deep Security, and Deep Load are delegated through the specialist provider architecture.

---

## 5. Diagnostics Boundary

How traCtl defines the observability scope of workflow validation and execution.

| traCtl Category     | What it means |
|---|---|
| **traCtl Diagnostics** | Execution-context observability for backend validation workflows: visibility into communication health, protocol behavior, and session lifecycle during test execution |

### Included in traCtl Diagnostics

- **DNS diagnostics** — resolution, query timing, result validation
- **TLS diagnostics** — handshake completion, certificate validation, cipher suite selection, session lifecycle
- **Transport timing** — connection establishment, time-to-first-byte, total request/response duration
- **Proxy routing visibility** — proxy negotiation, request routing, upstream selection
- **Connection lifecycle telemetry** — connection reuse, pooling behavior, closure events
- **Protocol session lifecycle visibility** — HTTP keep-alive, WebSocket handshake, gRPC stream states, where supported

### Excluded from traCtl Diagnostics

- Packet capture, sniffing, or decoding
- Traceroute or route discovery tooling
- Routing table introspection
- Kernel, NIC, or driver telemetry
- Infrastructure monitoring (CPU, memory, disk, network saturation)
- Network forensic tooling or deep packet analysis

> traCtl diagnostics are scoped to application-layer protocol behavior and execution context observability. Network infrastructure monitoring and packet-level analysis are out of scope and delegated to dedicated network observability products.

---

## 6. Capability Contracts

How traCtl groups the capabilities that workflows require and that platform components provide.

A **capability** is a declared unit of execution behavior that a workflow may require and that the engine, an extension, a provider adapter, or an MCP adapter may provide. The planner resolves capabilities by contract, never by the name of the component that supplies them.

A **capability contract** is the versioned semantic definition of a capability. It carries the capability identifier, its semantic contract version, declared inputs and outputs, behavioral guarantees, execution constraints, and compatibility metadata.

Capabilities fall into two tiers.

| traCtl Category | Owner | Versioning | Reference Form | Examples |
|---|---|---|---|---|
| **Platform Capability** | Engine | Versions with the engine | Stable flag, no `@version` suffix | `protocol.http`, `protocol.grpc`, `scripting.js` |
| **Extension / Provider Capability** | Extension, provider adapter, or MCP adapter | Independent semantic version, mandatory | `identifier@version` | `structured_output@1`, `tool_calling@2`, `secret.ext.vault@1` |

Platform capabilities are owned and provided by the engine. They are referenced as stable flags. Their version is the engine version, declared through the consuming context — the `schemaVersion` and capability flags in traCtlSpec, or the `engineVersion` range in an extension manifest.

Extension and provider capabilities are owned by components outside the engine. Their semantics drift independently of the engine and vary across sources, so they carry mandatory semantic versions in `identifier@version` form.

The planner matches both tiers through the same contract-matching mechanism. The two tiers differ only in how the version is sourced: a platform capability derives its version from the engine, while an extension or provider capability declares its version explicitly in the contract identifier.

Capability resolution rules:

- Capability matching uses contract compatibility, not symbolic name matching.
- Unknown capability contracts fail planning.
- Incompatible capability versions fail planning.
- Silent capability substitution is not permitted.
- Capability sources normalise into one canonical capability contract model regardless of origin — native extension, provider adapter, MCP adapter, or core engine.

---

## 6. Quick Reference — Category at a Glance

| What you are talking about | traCtl category name |
|---|---|
| TOON, YAML, JSON | **Native Authoring Format** |
| REST, GraphQL, gRPC, SOAP | **Protocol** |
| OpenAPI, WSDL, proto, GraphQL SDL | **Native Spec Provider** |
| Postman, Bruno, Insomnia, HAR, curl | **External Compatibility Provider** |
| k6, Schemathesis, OWASP ZAP | **Delegated Specialist Provider** |
| Correctness, Auth, Contract, Resilience, Performance Smoke | **Validation Stage** |
| `protocol.http`, `scripting.js` | **Platform Capability** |
| `structured_output@1`, `tool_calling@2` | **Extension / Provider Capability** |
| The versioned semantic definition of a capability | **Capability Contract** |
