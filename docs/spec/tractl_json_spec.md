# traCtl JSON Serialization Specification v1.0

> Version 1.0 · Normative JSON Serialization & Canonical Projection Specification
> Classification: **Normative**
> Date: 2026-05-24

## 1. Purpose

This specification defines the normative JSON serialization contract for traCtl.

JSON serves two roles in the traCtl platform:

1. **Native authoring format.** JSON is one of three native authoring formats (TOON, YAML, JSON) per ADR-002 §2 and `00_terminology.md` §1, intended for programmatic generation and machine integration.
2. **Canonical interchange projection.** JSON is the stable wire format for traCtlSpec across tooling, MCP responses, on-disk caches, and cross-tool exchange, per `tractl_spec.md` §17.

These two roles share a single set of representation rules but differ in normative obligations: authored JSON MAY use any valid ordering of mapping keys, while the canonical projection MUST emit lexicographically sorted keys.

This document defines:

- the JSON subset accepted as native authoring input
- the canonical JSON projection contract for traCtlSpec
- representation rules for all canonical traCtlSpec constructs
- canonical normalization expectations
- producer, parser, and serializer conformance obligations
- cross-format parity obligations with TOON and YAML
- forward compatibility rules

This document does **not** define:

- canonical semantic execution behavior (governed by `tractl_spec.md`)
- overlay merge algorithms (governed by `tractl_overlay_spec.md`)
- parser implementation
- runtime, planner, or scheduler behavior
- protocol-specific transport semantics

---

## 2. Scope / Non-Scope

### In Scope

- JSON authoring input contract
- Canonical JSON projection contract
- JSON representation of all canonical traCtlSpec constructs
- JSON representation of overlay documents
- Canonical JSON normalization rules
- Producer, parser, and serializer conformance obligations
- JSON ↔ TOON ↔ YAML parity obligations
- Forward compatibility rules

### Out of Scope

- canonical semantic execution behavior
- overlay merge semantics
- planner or runtime behavior
- protocol provider semantics
- JSON editor or tooling implementation
- JSON parser implementation internals beyond conformance obligations

---

## 3. Design Principles

1. **JSON is a serialization format, not a semantic model.** The canonical semantic model is traCtlSpec.
2. **Canonical projection clarity.** JSON serves as the stable, deterministic interchange format for traCtlSpec across tools and surfaces. Its determinism rules are stricter than its authoring rules.
3. **Cross-format parity.** JSON MUST NOT carry semantics expressible in JSON but not expressible in TOON or YAML.
4. **No JSON dialects.** JSON5, JSONC (JSON with comments), trailing-comma JSON, and unquoted-key JSON MUST NOT be accepted. The accepted dialect is strict RFC 8259 JSON.
5. **Determinism.** Two semantically equivalent documents MUST produce identical canonical JSON projection output across conformant implementations.
6. **Machine first.** JSON's authoring ergonomics target programmatic generation. Human-authored input is permitted but is not the optimization target; TOON and YAML serve that purpose.
7. **No implicit type coercion.** JSON's strict type system is preserved; no string-to-number, number-to-string, or boolean coercion is performed during normalization.

---

## 4. Serialization Philosophy

JSON is **both** a native authoring format and the canonical interchange projection.

JSON is **not**:

- canonical execution truth
- more authoritative than TOON or YAML as an authoring format
- a runtime input format
- a planner or executor input format

Per ADR-001 §2 and ADR-002 §1, the canonical semantic model and the serialization format are separate concerns. JSON is one of several representations that normalize into traCtlSpec, and is the stable projection of traCtlSpec for interchange.

A document authored in JSON, when semantically equivalent to a document authored in TOON or YAML, MUST produce identical canonical traCtlSpec output per ADR-002 §3 and §3a.

### 4.1 Authoring JSON vs Canonical JSON Projection

| Aspect                    | Authoring JSON | Canonical JSON Projection |
|---------------------------|----------------|---------------------------|
| Source                    | User-authored or programmatically generated | Emitted from canonical traCtlSpec |
| Key ordering              | Any            | Lexicographic              |
| Whitespace                | Any            | Normalized (see §16)       |
| `_ulid` fields            | Absent         | Present                    |
| Round-trip determinism    | Not required   | Required                   |
| Use                       | Input to canonical pipeline | Interchange wire format |

This distinction is normative. The two roles share a representation grammar but diverge on serialization obligations.

---

## 5. Canonical Relationship to traCtlSpec

JSON sits in two positions in the canonical architecture:

**As authoring input,** JSON sits in the Input Layer of the canonical pipeline defined in `tractl_spec.md` §2:

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
Planning Layer
    ↓
Execution
```

**As canonical projection,** JSON is emitted from canonical traCtlSpec per `tractl_spec.md` §17 for tooling, MCP responses, and cross-tool interchange. The canonical projection is consumed by external tools but is NOT consumed by the planner, scheduler, or runtime — those consume the internal canonical semantic model directly.

---

## 6. Accepted JSON Subset

The accepted dialect is **strict RFC 8259 JSON** (also known as ECMA-404), with the additional restrictions enumerated in this specification.

### 6.1 Base Standard

The base standard is RFC 8259, with the I-JSON profile (RFC 7493) applied for interoperability:

- numeric values SHOULD be representable as IEEE 754 double precision
- string values MUST NOT contain unpaired surrogates
- duplicate keys MUST NOT appear within the same object

### 6.2 Forbidden JSON Dialects

The following JSON variants MUST NOT be accepted:

- **JSON5** (trailing commas, unquoted keys, single-quoted strings, hex literals, infinity/NaN)
- **JSONC** (JSON with comments)
- **NDJSON** / **JSON Lines** (newline-delimited streams)
- **HJSON**, **JSONNet**, **CSON**, **YAMLscript JSON**, or any other dialect
- **BSON**, **CBOR**, **MessagePack**, or any binary JSON variant
- trailing commas in arrays or objects
- single-quoted strings
- unquoted object keys
- comments in any position
- `Infinity`, `-Infinity`, `NaN`
- hex, octal, or binary number literals

A conformant parser MUST reject any document containing the above constructs.

### 6.3 Document Structure

A traCtl JSON document MUST contain exactly one top-level JSON value.

Multi-value streams (NDJSON, concatenated JSON values) MUST be rejected.

The top-level value MUST be a JSON object.

### 6.4 Encoding

JSON documents MUST be encoded as UTF-8 without a byte-order mark (BOM).

UTF-16 and UTF-32 JSON encodings (technically permitted by older RFCs) MUST be rejected.

Conformant parsers MAY accept and silently strip a leading BOM for compatibility, but MUST NOT emit a BOM during serialization.

### 6.5 Line Endings

Line endings in JSON documents are not semantically significant.

Canonical JSON projection (§16) MUST NOT emit line breaks between tokens unless explicitly producing pretty-printed output for diagnostic purposes. Pretty-printed JSON is NOT canonical and MUST NOT be used as interchange wire format.

---

## 7. Scalar Representation

### 7.1 Strings

Strings follow RFC 8259 string grammar:

- delimited by double quotes
- escapes recognized: `\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`, `\uXXXX`
- surrogate pairs MUST be well-formed
- raw control characters (`U+0000` through `U+001F`) MUST be escaped

#### 7.1.1 Escape Normalization

Canonical JSON projection MUST:

- escape only characters that require escaping (U+0000–U+001F, `"`, `\`)
- emit two-character escapes (`\n`, `\t`, etc.) for characters with two-character forms
- emit `\uXXXX` only for characters lacking two-character forms
- emit lowercase hex digits in `\uXXXX` escapes
- NOT escape the solidus (`/`) — the `\/` form is permitted on input but MUST NOT be emitted in canonical projection

### 7.2 Numbers

Numbers follow RFC 8259 §6 grammar:

```
number = [ "-" ] int [ frac ] [ exp ]
```

Numbers MUST be representable as IEEE 754 double precision per I-JSON. Integer values that exceed safe integer range (`2^53 - 1`) SHOULD be represented as strings if precise integer fidelity is required; canonical projection MUST NOT silently truncate.

Forbidden number forms:

- `+42` (positive sign prefix)
- `0123` (leading zeros, except the bare value `0`)
- `0x1F`, `0o17`, `0b101` (radix prefixes)
- `1_000_000` (underscored numerics)
- `.5` (missing integer part)
- `5.` (missing fractional part after decimal point)
- `Infinity`, `-Infinity`, `NaN`

#### 7.2.1 Number Normalization

Canonical JSON projection MUST:

- emit integers without fractional part or exponent (e.g., `42`, not `42.0` or `4.2e1`)
- emit floating-point numbers using shortest round-trip representation per RFC 7493
- emit exponents using lowercase `e`
- emit negative zero as `0` (not `-0`) unless the value is semantically distinct (no traCtlSpec field carries semantic negative zero)

### 7.3 Booleans

Booleans are the literal tokens `true` and `false`, lowercase, as defined by RFC 8259.

### 7.4 Null

Null is the literal token `null`, lowercase, as defined by RFC 8259.

### 7.5 Durations

Durations MUST be represented as ISO 8601 duration strings (e.g., `PT5S`, `PT1M30S`, `PT2H`), per `tractl_spec.md` §17.

Durations are encoded as JSON strings, not numbers.

### 7.6 No Distinct Date Type

JSON has no native date or timestamp type. Timestamps and dates required by canonical traCtlSpec fields MUST be encoded as JSON strings in ISO 8601 format. Their interpretation as dates is governed by canonical schema field definitions, not by JSON syntax.

---

## 8. Structural Constructs

### 8.1 Objects

JSON objects represent canonical mappings. Object key ordering rules differ between authoring JSON and canonical projection (see §15.1).

Object keys MUST be JSON strings. Numeric keys are NOT permitted (this is a JSON5 feature).

Duplicate keys within the same object MUST be rejected.

### 8.2 Arrays

JSON arrays represent canonical sequences. Array element order MUST be preserved.

### 8.3 Empty Containers

Empty objects and arrays are represented as `{}` and `[]` respectively.

### 8.4 Nesting

Nesting depth is unbounded.

---

## 9. Identity Representation

Identity rules align with `tractl_spec.md` §4 and are equivalent across TOON, YAML, and JSON.

### 9.1 User IDs

User IDs MUST be represented as JSON string values bound to the `id` key:

```json
{ "id": "login-step" }
```

User IDs MUST conform to the pattern in `tractl_spec.md` §4.1. Reserved prefixes (`_traCtl.`, `traCtl.`) MUST NOT be used in authored JSON.

### 9.2 ULIDs

ULIDs (`_ulid` fields) MUST NOT appear in **authoring JSON**.

ULIDs MUST appear in **canonical JSON projection** for every identifiable entity per `tractl_spec.md` §4.2 and §17.

A JSON document submitted as authoring input that contains `_ulid` keys MUST be rejected, unless the submission is explicitly marked as a round-trip preservation operation by the consuming tool.

### 9.3 References

Cross-entity references in JSON use scalar string values containing dotted paths or expression interpolation per §10:

```json
{
  "dependsOn": ["login"],
  "extracts": [
    { "id": "token", "source": "body", "path": "$.access_token" }
  ]
}
```

---

## 10. Expression Serialization

Per `tractl_spec.md` §13.5, traCtlSpec supports `${ ... }` interpolation within scalar values.

### 10.1 Inline Expressions

Expressions MUST be embedded as substring content within JSON string values:

```json
{
  "url": "https://api.example.com/users/${vars.userId}",
  "authorization": "Bearer ${steps.login.extracts.token}"
}
```

### 10.2 String-Only Encoding

Expressions are always encoded as JSON strings. There is no distinct JSON expression type; the `${ ... }` syntax is opaque substring content interpreted by the canonical expression evaluator at runtime.

### 10.3 Expression Grammar

The expression grammar is owned by `tractl_spec.md` §13.5 and is NOT redefined here. JSON serialization treats expressions as opaque substring content within string values.

### 10.4 Escaping

A literal `${` sequence that MUST NOT be interpreted as an expression MUST be escaped as `\${` within the JSON string value. This is a content-level escape interpreted by the canonical expression evaluator; the JSON layer sees a normal backslash escape sequence per §7.1.

### 10.5 No Expression Evaluation at Parse Time

JSON parsers MUST NOT evaluate, partially evaluate, or transform expression content during normalization.

---

## 11. Secret Reference Serialization

Per ADR-006 §3, secrets are never directly embedded in workflow definitions.

### 11.1 Secret Reference Form

Secret references MUST be expressed as expressions resolving against a canonical secret namespace:

```json
{ "apiKey": "${secrets.prod.stripe_api_key}" }
```

### 11.2 Prohibitions

The following MUST be rejected by conformant JSON producers and MUST be flagged by conformant JSON parsers as policy violations:

- literal secret material in any string position
- inline base64-encoded credentials
- raw API tokens
- raw passwords
- raw private keys

Detection of literal secrets is best-effort. The architectural prohibition on literal secrets is owned by ADR-006 §3 and is enforced by the secret resolution pipeline; JSON-layer detection is an authoring-time hint only.

### 11.3 No Secret-Specific JSON Syntax

JSON MUST NOT introduce secret-specific structural constructs. Secrets are referenced through the standard expression mechanism (§10). This preserves cross-format parity with TOON and YAML.

---

## 12. Capability Serialization

Per `tractl_spec.md` §3.2, capabilities are tiered: platform capabilities (stable flags, no version suffix) and extension/provider capabilities (mandatory `identifier@version`).

### 12.1 Platform Capabilities

Platform capabilities MUST be serialized as JSON string values without version suffix:

```json
{
  "capabilities": [
    "protocol.http",
    "scripting.js",
    "diagnostics.tls"
  ]
}
```

### 12.2 Extension and Provider Capabilities

Extension and provider capabilities MUST be serialized as JSON string values in canonical `identifier@version` form:

```json
{
  "capabilities": [
    "protocol.ext.amqp@1",
    "secret.ext.vault@1",
    "assertion.ext.jsonschema@2"
  ]
}
```

JSON strings naturally accommodate the `@` character without quoting concerns; no additional escape is required.

### 12.3 Capability Source Independence

JSON capability serialization MUST NOT vary by capability source per ADR-011 §2.

### 12.4 Capability Ordering

Capability arrays MUST preserve user-authored order during serialization round-trips.

---

## 13. Overlay Serialization

Overlay documents authored in JSON MUST conform to the canonical overlay document model defined in `tractl_overlay_spec.md` §8.

### 13.1 Document Shape

```json
{
  "metadata": {
    "name": "dev-overlay"
  },
  "patches": [
    {
      "target": { "path": "workflows.user-check.steps.login" },
      "patch":  { "timeout": "PT30S" }
    }
  ]
}
```

Only `metadata` (optional) and `patches` (required) are valid top-level keys.

### 13.2 Target Selectors

Target selectors MUST follow the three-mode model in `tractl_overlay_spec.md` §9. Each target MUST use exactly one of `path`, `match`, or `source`.

### 13.3 Action Field

The `action` field, when present, MUST contain one of the canonical action literals defined in `tractl_overlay_spec.md` §10.4:

- `replace`
- `deepMerge`
- `append`
- `appendUnique`
- `remove`

### 13.4 Removal Patches

Removal patches (`action: remove`) MUST NOT contain a `patch` field, per `tractl_overlay_spec.md` §11.

### 13.5 Cross-Format Parity

A JSON overlay MUST normalize identically to a semantically equivalent TOON or YAML overlay. The following combinations MUST produce identical canonical traCtlSpec output per ADR-002 §3a:

- `workflow.json + overlay.toon`
- `workflow.json + overlay.yaml`
- `workflow.json + overlay.json`
- `workflow.toon + overlay.json`
- `workflow.yaml + overlay.json`
- `openapi.json + overlay.yaml`
- `postman.json + overlay.json`

---

## 14. Array Ordering Rules

### 14.1 Element Order Preservation

JSON array element order MUST be preserved across normalization and round-trips, per `tractl_spec.md` §17 and `tractl_overlay_spec.md` §20.

### 14.2 Semantically Ordered Arrays

The following arrays carry semantic order. Reordering during round-trip serialization is prohibited:

- `patches` (overlay application order)
- `dependsOn` (declarative; order preserved for diagnostic purposes)
- `workflows`, `steps`
- `capabilities`
- `assertions`, `extracts`
- `extensions`
- `hooks`

### 14.3 No Implicit Sort

Canonical JSON projection MUST NOT sort array elements. Sort order applies only to object keys (§15.1).

---

## 15. Object Ordering Rules

### 15.1 Canonical Projection Key Order

Canonical JSON projection MUST emit object keys in **lexicographic order** by Unicode code point per `tractl_spec.md` §17.

Lexicographic ordering is well-defined for all keys in canonical traCtlSpec because keys are restricted to the identifier pattern in `tractl_spec.md` §4.1 (ASCII alphanumerics, dot, underscore, hyphen).

### 15.2 Authoring JSON Key Order

Authoring JSON input MAY use any key order. Canonical normalization re-orders keys per §15.1 during projection emission.

A round-trip from authored JSON through canonical normalization and back to canonical JSON projection will reorder keys lexicographically. This is expected behavior, not data loss. Tooling that requires preservation of authored key order MUST retain the original authored document; canonical projection is not a substitute.

### 15.3 No User-Defined Sort

Custom key orderings (semantic priority orderings, schema-declared orderings) MUST NOT be applied to canonical JSON projection. Lexicographic ordering is the sole canonical ordering.

---

## 16. Canonical JSON Projection Contract

The canonical JSON projection is the **stable interchange format** for traCtlSpec per `tractl_spec.md` §17.

### 16.1 Mandatory Projection Rules

A conformant canonical JSON projection MUST:

- emit UTF-8 without BOM
- emit object keys in lexicographic order per §15.1
- preserve array order per §14.1
- emit `_ulid` fields for every identifiable entity per `tractl_spec.md` §4.2
- emit durations as ISO 8601 strings per `tractl_spec.md` §17
- emit enum values as lowercase strings per `tractl_spec.md` §17
- emit numbers using shortest round-trip representation per RFC 7493
- emit no trailing whitespace
- emit no comments
- emit no JSON5 dialect features

### 16.2 Whitespace in Canonical Projection

Canonical JSON projection MUST emit **minified JSON** — no whitespace between tokens except where required by JSON grammar.

Pretty-printed JSON (with indentation and line breaks) is NOT canonical. Pretty-printed JSON MAY be produced by tooling for human consumption but MUST NOT be used as interchange wire format.

This rule ensures byte-identical canonical projections for semantically equivalent inputs and enables content-addressable storage and integrity verification.

### 16.3 Stable Interchange Hash Invariant

Two semantically equivalent traCtlSpec documents MUST produce byte-identical canonical JSON projection. This means a cryptographic hash of the canonical JSON projection is a stable content identifier for the underlying canonical semantics.

This invariant is the operational basis for:

- content-addressable workflow caches
- audit fingerprinting
- cross-tool integrity verification
- MCP response equivalence checks

---

## 17. Deterministic Normalization Expectations

### 17.1 Whitespace

Insignificant whitespace in authoring JSON MUST NOT affect canonical output.

### 17.2 Comments

Comments MUST NOT appear in JSON documents (§6.2). There is no normalization rule for comments because they are syntactically prohibited.

### 17.3 Scalar Normalization

- Strings: Unicode normalization form NFC MUST be applied.
- Booleans: emitted as `true` / `false`.
- Null: emitted as `null`.
- Numbers: integers emitted without fractional part; floats emitted with shortest round-trip representation.
- Durations: normalized to canonical ISO 8601 duration strings.

### 17.4 Expression Normalization

Expression content within strings MUST be preserved verbatim. JSON normalization MUST NOT modify expression syntax, whitespace within expressions, or expression structure.

### 17.5 Determinism Invariant

For any two JSON documents D₁ and D₂ that are semantically equivalent:

- canonical traCtlSpec(D₁) = canonical traCtlSpec(D₂)
- canonical JSON projection(D₁) = canonical JSON projection(D₂) (byte-identical)
- canonical JSON projection from canonical traCtlSpec is byte-stable across re-emissions

---

## 18. Conformance Requirements

### 18.1 Producer Conformance — Authoring JSON

A conformant authoring-JSON producer MUST:

- emit strict RFC 8259 JSON
- emit UTF-8 without BOM
- emit no comments, trailing commas, single-quoted strings, or unquoted keys
- emit no `_ulid` fields
- emit ISO 8601 durations as JSON strings
- emit no literal secret material
- emit no JSON5 dialect features

### 18.2 Producer Conformance — Canonical JSON Projection

A conformant canonical JSON projection emitter MUST:

- satisfy all rules in §18.1
- emit object keys in lexicographic order per §15.1
- emit `_ulid` fields for all identifiable entities
- emit minified output (no insignificant whitespace) per §16.2
- emit byte-identical output for semantically equivalent canonical traCtlSpec inputs per §16.3
- emit numbers using shortest round-trip representation
- emit enum values as lowercase strings
- emit escape sequences per §7.1.1

### 18.3 Parser Conformance

A conformant JSON parser MUST:

- accept strict RFC 8259 JSON
- accept UTF-8 input (BOM accepted and stripped)
- reject UTF-16 and UTF-32 encodings
- reject all JSON5 / JSONC dialect features per §6.2
- reject trailing commas
- reject single-quoted strings
- reject unquoted object keys
- reject comments
- reject `Infinity`, `-Infinity`, `NaN`
- reject duplicate object keys per RFC 7493
- reject multi-value document streams
- reject top-level non-object values
- reject `_ulid` fields in authoring contexts (per §9.2)
- preserve array order
- produce a canonical source model equivalent to the canonical source model produced by a conformant TOON or YAML parser for a semantically equivalent input

### 18.4 Serializer Conformance

A conformant JSON serializer MUST:

- produce output round-trippable through a conformant parser
- produce byte-identical canonical projection for semantically equivalent inputs
- not emit serializer-specific metadata or markers
- not emit JSON5 dialect features

### 18.5 Cross-Implementation Equivalence

Independent conformant implementations MUST produce byte-identical canonical JSON projection for identical canonical traCtlSpec inputs. This is the operational test of JSON canonical projection determinism and is the basis for cross-tool integrity verification.

---

## 19. Canonical Examples

### 19.1 Minimal Workflow (Authoring JSON)

```json
{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "health-check",
      "steps": [
        {
          "id": "ping",
          "kind": "request",
          "request": {
            "protocol": "http",
            "target": "https://api.example.com/health",
            "operation": "GET"
          },
          "assertions": [
            {
              "id": "ok",
              "kind": "status",
              "op": "equals",
              "expected": 200
            }
          ]
        }
      ]
    }
  ]
}
```

### 19.2 Same Workflow as Canonical JSON Projection

Note the lexicographic key ordering, presence of `_ulid` fields, and minified emission:

```json
{"_ulid":"01HXYZ...","capabilities":["protocol.http"],"schemaVersion":1,"workflows":[{"_ulid":"01HXYZ...","id":"health-check","steps":[{"_ulid":"01HXYZ...","assertions":[{"_ulid":"01HXYZ...","expected":200,"id":"ok","kind":"status","op":"equals"}],"id":"ping","kind":"request","request":{"operation":"GET","protocol":"http","target":"https://api.example.com/health"}}]}]}
```

(ULIDs above are abbreviated for readability; conformant projections emit full 26-character ULIDs.)

### 19.3 Workflow with Expressions and Capabilities (Authoring JSON)

```json
{
  "schemaVersion": 1,
  "capabilities": [
    "protocol.http",
    "scripting.js",
    "secret.ext.vault@1"
  ],
  "variables": {
    "baseUrl": "https://api.example.com"
  },
  "workflows": [
    {
      "id": "user-flow",
      "steps": [
        {
          "id": "login",
          "kind": "request",
          "request": {
            "protocol": "http",
            "target": "${vars.baseUrl}/auth/login",
            "operation": "POST",
            "headers": {
              "Content-Type": "application/json"
            },
            "body": {
              "encoding": "json",
              "content": {
                "user": "${vars.username}",
                "key": "${secrets.prod.api_key}"
              }
            }
          },
          "extracts": [
            {
              "id": "token",
              "source": "body",
              "path": "$.access_token"
            }
          ]
        },
        {
          "id": "fetch-profile",
          "kind": "request",
          "dependsOn": ["login"],
          "request": {
            "protocol": "http",
            "target": "${vars.baseUrl}/me",
            "operation": "GET",
            "headers": {
              "Authorization": "Bearer ${steps.login.extracts.token}"
            }
          },
          "assertions": [
            {
              "id": "profile-ok",
              "kind": "status",
              "op": "equals",
              "expected": 200
            }
          ]
        }
      ]
    }
  ]
}
```

### 19.4 Overlay Document (Authoring JSON)

```json
{
  "metadata": {
    "name": "prod-overlay",
    "description": "Production environment augmentation"
  },
  "patches": [
    {
      "target": { "path": "workflows.user-flow.steps.login" },
      "patch": {
        "timeout": "PT30S",
        "retry": {
          "maxAttempts": 3,
          "backoff": "exponential",
          "delay": "PT1S"
        }
      }
    },
    {
      "target": {
        "mode": "all",
        "match": {
          "kind": "request",
          "protocol": "http"
        }
      },
      "patch": {
        "diagnostics": {
          "enabled": true,
          "kinds": ["tls", "transport"]
        }
      }
    },
    {
      "action": "remove",
      "target": {
        "path": "workflows.user-flow.steps.login.request.headers.Accept"
      }
    }
  ]
}
```

### 19.5 Failure Examples

The following MUST be rejected by conformant parsers:

```json
{
  "trailing": "comma",
}
```

```json
{
  // comments are not permitted
  "key": "value"
}
```

```json
{
  unquoted: "key"
}
```

```json
{
  "value": Infinity
}
```

```json
[
  { "schemaVersion": 1 },
  { "schemaVersion": 1 }
]
```

(Top-level array; multi-value streams; both forbidden.)

---

## 20. Cross-Format Parity Guarantees

### 20.1 Bidirectional Parity

Every JSON document conforming to this specification MUST be representable in TOON and YAML without semantic loss. Every TOON or YAML document conforming to the traCtl serialization specifications MUST be representable in JSON without semantic loss.

### 20.2 Overlay Parity

The following combinations MUST produce identical canonical traCtlSpec output per ADR-002 §3a:

- `workflow.json + overlay.toon`
- `workflow.json + overlay.yaml`
- `workflow.json + overlay.json`
- `workflow.toon + overlay.json`
- `workflow.yaml + overlay.json`
- `openapi.json + overlay.yaml`
- `postman.json + overlay.json`

### 20.3 No JSON-Exclusive Semantics

JSON MUST NOT introduce any construct that lacks an equivalent in TOON or YAML. The set of semantic constructs expressible in JSON is the intersection of constructs expressible across all three native authoring formats, governed by the canonical traCtlSpec model.

### 20.4 No Format Hierarchy

JSON is not semantically richer or poorer than TOON or YAML, despite serving as the canonical interchange projection. All three are equivalent native authoring formats per `00_terminology.md` §1.

The canonical projection role is a **serialization role**, not a semantic elevation. JSON does not become more authoritative than TOON or YAML as a result of being the projection format.

---

## 21. Validation Expectations

JSON parsing produces a canonical source model. Validation owners are defined by `tractl_overlay_spec.md` §15:

- **JSON parser:** syntactic validity (encoding, RFC 8259 compliance, subset compliance).
- **Overlay validator:** overlay schema, selector validity, target existence, merge legality, removal patch shape.
- **Canonical validator:** canonical schema, cross-reference integrity, dependency graph validity, capability declaration shape.
- **Planner:** capability resolution, runtime compatibility, composite cycle detection, execution planning.

JSON parsers MUST NOT perform canonical, overlay, or planner-owned validation.

---

## 22. Restrictions Summary

The following constructs MUST NOT appear in traCtl JSON documents:

| Construct                                    | Reason                                     |
|----------------------------------------------|--------------------------------------------|
| JSON5 features (any)                         | Strict RFC 8259 only                       |
| JSONC comments                               | Strict RFC 8259 only                       |
| Trailing commas                              | Strict RFC 8259 only                       |
| Single-quoted strings                        | Strict RFC 8259 only                       |
| Unquoted object keys                         | Strict RFC 8259 only                       |
| Hex, octal, binary number literals           | Strict RFC 8259 only                       |
| `+` positive sign prefix                     | Strict RFC 8259 only                       |
| Leading zeros on numbers                     | Strict RFC 8259 only                       |
| `Infinity`, `-Infinity`, `NaN`               | Strict RFC 8259 only                       |
| Underscored numerics                         | Strict RFC 8259 only                       |
| Top-level non-object values                  | Document shape                             |
| Multi-value streams (NDJSON, concatenated)   | Single-document only                       |
| Duplicate object keys                        | I-JSON requirement; determinism            |
| UTF-16 / UTF-32 encoding                     | UTF-8 only                                 |
| BOM in serialized output                     | UTF-8 conformance                          |
| Binary JSON variants (BSON, CBOR, etc.)      | Text JSON only                             |
| `_ulid` keys in authoring documents          | System-assigned identity                   |
| Reserved ID prefixes (`_traCtl.`, `traCtl.`) | Reserved namespace                         |
| Literal secret material                      | Per ADR-006 §3                             |
| Pretty-printed output as interchange         | Canonical projection is minified           |
| Unicode unpaired surrogates in strings       | I-JSON requirement                         |
| Raw control characters in strings            | Strict RFC 8259 (must be escaped)          |

---

## 23. Forward Compatibility

### 23.1 JSON Specification Versioning

This specification is versioned independently of traCtlSpec (`schemaVersion`).

A JSON document does NOT carry an explicit JSON-subset version marker. The `schemaVersion` field within the document declares the canonical traCtlSpec version; JSON serialization conformance is governed by the version of this specification active at parse time.

### 23.2 Additive Changes

The following changes are additive and do NOT bump this specification's version:

- new canonical traCtlSpec fields representable through existing JSON syntax
- new capability identifiers
- new platform capability flags
- new extension capability declarations

### 23.3 Breaking Changes

The following changes MUST bump this specification's version:

- changes to canonical JSON projection key ordering
- changes to canonical projection whitespace rules
- changes to scalar normalization
- changes to number representation rules
- changes to escape handling
- relaxation of forbidden dialect features

### 23.4 Schema Version Independence

A JSON document declaring `schemaVersion: N` MUST be parsable by any JSON-specification version. Canonical compatibility is governed by `tractl_spec.md` §18.4.

### 23.5 Canonical Projection Stability

The canonical JSON projection contract (§16) is treated as a wire-format stability commitment. Changes to canonical projection that would alter the byte output for equivalent inputs are breaking changes and MUST bump this specification's version. Tooling that relies on canonical projection content hashing depends on this stability.

---

## 24. Governance

Per ADR-010, this specification sits at hierarchy position 4 (schemas and wire contracts).

Conflicts between this specification and:

- ADRs (position 1) → ADR wins
- canonical architecture specifications (position 2) → canonical spec wins
- `00_terminology.md` (position 3) → terminology registry wins
- `tractl_spec.md` (position 4, canonical semantic) → traCtlSpec wins
- `tractl_overlay_spec.md` (position 4, overlay) → overlay spec wins
- HLD, PRD, roadmap, README (lower positions) → this specification wins

Where this specification overlaps with `tractl_spec.md` §17 (canonical JSON projection), the rules in §16 of this document MUST be read as the expanded, normative form of the §17 contract. Any conflict resolves in favor of `tractl_spec.md` §17 per ADR-010 hierarchy.

Cross-format parity obligations are owned jointly with the TOON and YAML serialization specifications. A parity violation in any of the three is a conformance defect in all three.
