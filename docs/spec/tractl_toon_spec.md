# traCtl TOON Serialization Specification v1.0

> Version 1.0 · Normative TOON Authoring & Serialization Specification
> Classification: **Normative**
> Date: 2026-05-24

## 1. Purpose

This specification defines the normative TOON serialization contract for traCtl.

TOON (Token-Oriented Object Notation) is the preferred ergonomic authoring format for traCtl workflows, overlays, environment artifacts, and capability declarations. TOON is optimized for human readability and Git diffability.

This document defines:

- the TOON syntactic surface used by traCtl
- representation rules for canonical traCtlSpec concepts
- canonical normalization expectations
- producer and consumer conformance obligations
- cross-format parity obligations with YAML and JSON
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

- TOON document structure
- TOON lexical and syntactic rules
- TOON representation of all canonical traCtlSpec constructs
- TOON representation of overlay documents
- TOON canonical normalization rules
- TOON conformance obligations
- TOON ↔ YAML ↔ JSON parity obligations
- TOON forward compatibility rules

### Out of Scope

- canonical semantic execution behavior
- overlay merge semantics
- planner or runtime behavior
- protocol provider semantics
- TOON editor or tooling implementation
- TOON parser implementation details beyond conformance obligations

---

## 3. Design Principles

1. **TOON is a serialization format, not a semantic model.** TOON is one of three native authoring formats (TOON, YAML, JSON). The canonical semantic model is traCtlSpec.
2. **Structural explicitness.** TOON syntax MUST make hierarchy and identity unambiguous at the lexical level. No implicit semantics, no magic shorthand.
3. **Determinism.** Identical TOON documents MUST produce identical canonical traCtlSpec output across conformant implementations.
4. **Human ergonomics.** TOON SHOULD minimize syntactic overhead to preserve readability and reduce authoring friction.
5. **Git friendliness.** TOON SHOULD produce minimal, line-oriented diffs under common editing operations.
6. **Parity preservation.** TOON MUST NOT carry semantics expressible in TOON but not expressible in YAML or JSON.
7. **No hidden inheritance.** TOON MUST NOT support anchors, aliases, merge keys, or any implicit reuse mechanism.
8. **Machine safety.** Ambiguous scalar interpretations (boolean coercion, octal parsing, timezone-less timestamps) MUST be eliminated by construction.

---

## 4. Serialization Philosophy

TOON is the **preferred ergonomic authoring syntax** for traCtl. TOON, YAML, and JSON are semantically equivalent native authoring formats per ADR-002 §3 and `00_terminology.md` §1.

TOON is **not**:

- canonical execution truth
- more authoritative than YAML or JSON
- the only acceptable authoring surface
- a runtime input format
- a wire format for planner or executor consumption

Per ADR-001 §2 and ADR-002 §1, the canonical semantic model and the serialization format are separate concerns. TOON is one of several representations that normalize into traCtlSpec.

A document authored in TOON, when semantically equivalent to a document authored in YAML or JSON, MUST produce identical canonical traCtlSpec output per ADR-002 §3 and §3a.

---

## 5. Canonical Relationship to traCtlSpec

TOON sits in the Input Layer of the canonical pipeline defined in `tractl_spec.md` §2:

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

TOON parsing produces a canonical source model. TOON itself is never consumed by the planner, scheduler, or runtime.

A TOON document MUST be normalizable into a structurally valid traCtlSpec document candidate. Canonical validation occurs after overlay resolution per `tractl_spec.md` §2 and `tractl_overlay_spec.md` §3.

---

## 6. Lexical Rules

### 6.1 Character Encoding

TOON documents MUST be encoded as UTF-8 without a byte-order mark (BOM).

Conformant parsers MUST reject documents containing invalid UTF-8 byte sequences.

Conformant parsers MAY accept and silently strip a leading BOM for compatibility with editors that emit one, but MUST NOT emit a BOM during serialization.

### 6.2 Line Endings

TOON documents MUST use line feed (`U+000A`) as the line terminator.

Conformant parsers MUST accept carriage-return-line-feed (`U+000D U+000A`) and normalize it to LF on ingest.

Conformant serializers MUST emit LF only.

### 6.3 Indentation

Indentation MUST use exactly two space characters (`U+0020`) per nesting level.

Tab characters (`U+0009`) MUST NOT appear in indentation positions. A tab in an indentation position is a lexical error.

Indentation depth determines structural nesting. Adjacent sibling nodes MUST share identical indentation.

### 6.4 Whitespace

Trailing whitespace on any line MUST NOT be emitted by serializers. Conformant parsers MUST accept and ignore trailing whitespace on ingest.

Blank lines (zero-length or whitespace-only) are syntactically permitted between top-level and block-level constructs. Blank lines MUST NOT carry semantic meaning.

### 6.5 Comments

Comments begin with `#` and extend to end of line.

Two comment positions are recognized:

- **Line comments:** `#` at the start of a line (after optional indentation).
- **Trailing comments:** `#` preceded by at least one space character following a complete scalar value.

Comments inside quoted strings are not comments — they are string content.

Comments MUST NOT carry semantic meaning. Comments MUST be discarded during normalization and MUST NOT appear in canonical traCtlSpec output.

### 6.6 Identifiers (Bare Keys)

Bare identifier keys MUST match the pattern:

```
^[a-zA-Z_][a-zA-Z0-9_.-]*$
```

Keys not matching this pattern MUST be quoted as strings per §7.1.

This pattern is intentionally aligned with the user-ID pattern in `tractl_spec.md` §4.1 to permit direct use of canonical entity IDs as keys without quoting.

---

## 7. Scalar Representation

### 7.1 Strings

Strings have three forms:

| Form         | Delimiter | Escape processing | Multi-line |
|--------------|-----------|-------------------|------------|
| Bare         | none      | none              | no         |
| Double-quoted| `"…"`     | yes (§7.1.1)      | no         |
| Block        | `|` or `>`| no                | yes (§7.4) |

**Bare strings** are permitted only when the value:

- matches the identifier pattern in §6.6, OR
- contains no whitespace, no quote characters, no `#`, no `:`, no `[`, no `]`, no `{`, no `}`, no `,`, and does not collide with a reserved literal (§7.2, §7.3, §7.5).

Otherwise, the value MUST be double-quoted.

**Single-quoted strings are not supported in TOON.** This restriction eliminates the YAML 1.2 single-quote escape rules as a source of cross-format divergence.

#### 7.1.1 Escape Sequences

Double-quoted strings recognize the following escapes:

| Escape  | Meaning              |
|---------|----------------------|
| `\"`    | Double quote         |
| `\\`    | Backslash            |
| `\n`    | Line feed (U+000A)   |
| `\r`    | Carriage return      |
| `\t`    | Tab (U+0009)         |
| `\uXXXX`| Unicode code point   |

No other escape sequences are recognized. Unrecognized escapes (e.g., `\a`, `\v`, `\0`) MUST be rejected as lexical errors.

### 7.2 Booleans

The only valid boolean literals are `true` and `false`, lowercase.

The tokens `yes`, `no`, `on`, `off`, `True`, `False`, `TRUE`, `FALSE`, `Y`, `N` and all other variants are **not** booleans in TOON. When such a token appears in a scalar position, it is interpreted as a bare string.

This restriction directly addresses the "Norway problem" inherited from YAML 1.1 implicit typing.

### 7.3 Null

The only valid null literal is `null`, lowercase.

The YAML alternatives `~`, the empty scalar, `Null`, and `NULL` are **not** null in TOON. The token `null` (lowercase) is the sole representation. The empty scalar is an error in any value position.

### 7.4 Numbers

Numbers follow the JSON number grammar (RFC 8259 §6):

```
number = [ "-" ] int [ frac ] [ exp ]
```

The following are NOT valid TOON numbers and MUST be parsed as strings if appearing in scalar position:

- leading zeros (e.g., `0123`)
- octal literals (e.g., `0o17`)
- hexadecimal literals (e.g., `0x1F`)
- binary literals (e.g., `0b101`)
- underscored numerics (e.g., `1_000_000`)
- positive sign prefix (e.g., `+42`)
- the tokens `.inf`, `-.inf`, `.nan`, `Infinity`, `NaN`

Integer values MUST be representable as a 64-bit signed integer. Floating-point values MUST be representable as IEEE 754 double precision.

### 7.5 Durations

Durations MUST be represented as ISO 8601 duration strings (e.g., `PT5S`, `PT1M30S`, `PT2H`), consistent with `tractl_spec.md` §17.

Duration values MUST be double-quoted only when their containing field's schema does not declare duration semantics. Schema-declared duration fields MAY accept unquoted ISO 8601 strings as bare scalars provided the value matches the duration pattern `^P[T0-9HMS.]+$` and is not otherwise ambiguous.

### 7.6 Block Strings

Multi-line strings use either of two block indicators on the line preceding the content:

- `|` — **literal block.** Line breaks are preserved as-is.
- `>` — **folded block.** Consecutive non-empty lines are folded into a single space-separated line; blank lines are preserved as line breaks.

Block content indentation establishes the content baseline. The first content line's indentation depth MUST be deeper than the indicator line, and all subsequent content lines MUST be indented to at least the baseline depth.

Trailing newline handling:

- Default: a single trailing newline is preserved.
- `|-` and `>-`: trailing newlines are stripped.
- `|+` and `>+`: all trailing newlines are preserved.

---

## 8. Structural Constructs

### 8.1 Mappings

Mappings (objects) are expressed as `key: value` pairs, one per line, at consistent indentation:

```toon
metadata:
  name: my-workflow
  description: example
```

Inline mappings (`{ key: value }`) MUST NOT be used in TOON. All mappings MUST be expressed in block form.

Duplicate keys within the same mapping MUST be rejected as a structural error.

### 8.2 Sequences

Sequences (arrays) are expressed as block sequences with leading `-` markers:

```toon
patches:
  - target:
      path: workflows.user.steps.login
    patch:
      timeout: "PT30S"
```

Inline sequences (`[1, 2, 3]`) MUST NOT be used in TOON. All sequences MUST be expressed in block form.

Each `-` marker MUST be followed by exactly one space character before the entry content.

Sequence entry indentation establishes that all entries within a sequence share identical indentation depth.

### 8.3 Empty Containers

Empty mappings and sequences are represented explicitly using JSON-compatible notation:

```toon
emptyMap: {}
emptyList: []
```

These are the only inline structural forms permitted in TOON. They MUST be used only for empty containers.

### 8.4 Nesting

Nesting depth is unbounded. Each level adds two spaces of indentation.

Mixed sequence-of-mappings nesting MUST place the `-` marker at the depth of the enclosing sequence position, with mapping content indented one level deeper:

```toon
workflows:
  - id: user-check
    steps:
      - id: login
        kind: request
```

---

## 9. Identity Representation

Per `tractl_spec.md` §4, traCtlSpec uses a hybrid identity model: user IDs (required, user-authored) and ULIDs (system-assigned).

### 9.1 User IDs

User IDs MUST be represented as scalar string values bound to the `id` key:

```toon
id: login-step
```

User IDs MUST conform to the pattern in `tractl_spec.md` §4.1. Reserved prefixes (`_traCtl.`, `traCtl.`) MUST NOT be used in authored TOON.

### 9.2 ULIDs

ULIDs (`_ulid` fields) MUST NOT appear in authored TOON documents. ULIDs are system-assigned at normalization time.

A TOON document containing `_ulid` keys in any position MUST be rejected as a producer error during parsing, unless the document is being re-emitted by a tool that has explicitly opted into round-trip preservation. Round-trip preservation is an interchange concern and is not a normal authoring path.

### 9.3 References

Cross-entity references in TOON use scalar string values containing dotted paths or expression interpolation per §10:

```toon
dependsOn:
  - login
extracts:
  - id: token
    source: body
    path: $.access_token
```

Reference resolution is a canonical-validator responsibility per `tractl_spec.md` §6.2 and is outside the TOON serialization contract.

---

## 10. Expression Serialization

Per `tractl_spec.md` §13.5, traCtlSpec supports `${ ... }` interpolation within scalar values.

### 10.1 Inline Expressions

Expressions MUST be embedded as substring content within scalar string values:

```toon
url: "https://api.example.com/users/${vars.userId}"
authorization: "Bearer ${steps.login.extracts.token}"
```

### 10.2 Quoting Requirements

Any scalar value containing a `${` token MUST be double-quoted. This avoids ambiguity with bare scalar parsing and ensures deterministic round-tripping to JSON.

### 10.3 Expression Grammar

The expression grammar is owned by `tractl_spec.md` §13.5 and is NOT redefined here. TOON serialization treats expressions as opaque substring content within string values.

### 10.4 Escaping

A literal `${` sequence that MUST NOT be interpreted as an expression MUST be escaped as `\${`. This is the only context-sensitive escape recognized in TOON.

### 10.5 No Standalone Expression Type

Expressions are NOT a distinct TOON scalar type. They are string-typed values whose contents are interpreted by the canonical expression evaluator at runtime. TOON parsers MUST NOT evaluate, partially evaluate, or transform expression content during normalization.

---

## 11. Secret Reference Serialization

Per ADR-006 §3, secrets are never directly embedded in workflow definitions. Secrets are referenced through scoped injection.

### 11.1 Secret Reference Form

Secret references MUST be expressed as expressions resolving against a canonical secret namespace:

```toon
apiKey: "${secrets.prod.stripe_api_key}"
```

The secret namespace (`secrets.*`) is governed by the Security Architecture and is referenced through standard expression interpolation per §10.

### 11.2 Prohibitions

The following MUST be rejected by conformant TOON producers and MUST be flagged by conformant TOON parsers as policy violations:

- literal secret material in any scalar position
- inline base64-encoded credentials
- raw API tokens
- raw passwords
- raw private keys

Detection of literal secrets is best-effort and is not a substitute for the canonical secret resolution policy. The architectural prohibition on literal secrets is owned by ADR-006 §3 and is enforced by the secret resolution pipeline; TOON-layer detection is an authoring-time hint only.

### 11.3 No Secret-Specific Syntax

TOON MUST NOT introduce secret-specific syntactic constructs. Secrets are referenced through the standard expression mechanism (§10). This preserves cross-format parity with YAML and JSON, neither of which carries a secret-specific syntactic form.

---

## 12. Capability Serialization

Per `tractl_spec.md` §3.2, capabilities are tiered: platform capabilities (stable flags, no version suffix) and extension/provider capabilities (mandatory `identifier@version`).

### 12.1 Platform Capabilities

Platform capabilities MUST be serialized as bare or quoted identifier strings without version suffix:

```toon
capabilities:
  - protocol.http
  - scripting.js
  - diagnostics.tls
```

### 12.2 Extension and Provider Capabilities

Extension and provider capabilities MUST be serialized as identifier-with-version strings in canonical `identifier@version` form:

```toon
capabilities:
  - "protocol.ext.amqp@1"
  - "secret.ext.vault@1"
  - "assertion.ext.jsonschema@2"
```

Capability strings containing `@` MUST be double-quoted to remove any ambiguity with reserved characters and to ensure deterministic JSON projection.

### 12.3 Capability Source Independence

TOON capability serialization MUST NOT vary by capability source (native engine, extension, MCP adapter, provider adapter). Capability source is governed by the canonical capability contract model per ADR-011 and is invisible at the TOON layer.

### 12.4 Capability Ordering

Capability arrays MUST preserve user-authored order during serialization round-trips. Canonical traCtlSpec is responsible for capability resolution semantics; TOON preserves authored order as input to that process.

---

## 13. Overlay Serialization

Overlay documents authored in TOON MUST conform to the canonical overlay document model defined in `tractl_overlay_spec.md` §8.

### 13.1 Document Shape

Overlay documents MUST contain only the top-level keys `metadata` (optional) and `patches` (required), per `tractl_overlay_spec.md` §8.

```toon
metadata:
  name: dev-overlay

patches:
  - target:
      path: workflows.user-check.steps.login
    patch:
      timeout: "PT30S"
```

### 13.2 Target Selectors

Target selectors MUST follow the three-mode model in `tractl_overlay_spec.md` §9. Each target MUST use exactly one of `path`, `match`, or `source`. Multi-mode targets MUST be rejected during overlay validation.

### 13.3 Action Field

The `action` field, when present, MUST contain one of the canonical action literals defined in `tractl_overlay_spec.md` §10.4:

- `replace`
- `deepMerge`
- `append`
- `appendUnique`
- `remove`

Action values MUST be serialized as bare strings.

### 13.4 Removal Patches

Removal patches (`action: remove`) MUST NOT contain a `patch` field, per `tractl_overlay_spec.md` §11. A TOON overlay document containing a `patch` field on a removal entry MUST be rejected during overlay validation.

### 13.5 Cross-Format Parity

A TOON overlay MUST normalize identically to a semantically equivalent YAML or JSON overlay. The following combinations MUST produce identical canonical traCtlSpec output per ADR-002 §3a and `tractl_overlay_spec.md` §6:

- `workflow.toon + overlay.yaml`
- `workflow.yaml + overlay.toon`
- `workflow.json + overlay.toon`
- `openapi.yaml + overlay.toon`
- `postman.json + overlay.toon`

---

## 14. Ordering Rules

### 14.1 Mapping Key Ordering

Mappings MAY be authored in any key order. Canonical normalization MUST preserve authored order for round-trip emission to TOON.

Canonical JSON projection per `tractl_spec.md` §17 sorts keys lexicographically. TOON-to-canonical-JSON normalization MUST apply lexicographic sorting at the projection boundary, not at TOON ingest. This permits TOON authors to retain semantically meaningful authoring order (e.g., `id` first, `description` second) for readability without affecting canonical output.

### 14.2 Sequence Element Ordering

Sequence (array) element order MUST be preserved across normalization, per `tractl_spec.md` §17 and `tractl_overlay_spec.md` §20.

Sequence order is semantically meaningful for:

- `patches` (overlay application order)
- `dependsOn` (declarative; order preserved for diagnostic purposes only)
- `workflows`, `steps` (authored order preserved for diagnostics)
- `capabilities` (authored order preserved)
- `assertions`, `extracts` (authored order preserved)

Reordering of sequence elements during round-trip serialization is prohibited.

### 14.3 Patch Application Order

Within an overlay document, `patches` order MUST be preserved as authored per `tractl_overlay_spec.md` §12. TOON serializers MUST NOT reorder patch entries.

---

## 15. Deterministic Normalization Expectations

### 15.1 Whitespace

Insignificant whitespace (leading whitespace beyond required indentation, trailing whitespace, blank-line counts between blocks) MUST NOT affect canonical output.

### 15.2 Comments

Comments MUST be discarded during normalization and MUST NOT appear in canonical traCtlSpec output (§6.5).

### 15.3 Scalar Normalization

- Strings: Unicode normalization form NFC MUST be applied.
- Booleans: normalized to `true` / `false`.
- Null: normalized to `null`.
- Numbers: integers preserved as JSON integers; floats preserved as JSON numbers with shortest round-trip representation per RFC 7493.
- Durations: normalized to canonical ISO 8601 duration strings (e.g., `PT90S` normalizes to `PT1M30S`).

### 15.4 Expression Normalization

Expression content within strings MUST be preserved verbatim. TOON normalization MUST NOT modify expression syntax, whitespace within expressions, or expression structure.

### 15.5 Quoting Normalization

Round-trip serialization to TOON MUST emit the **minimal valid quoting form** for each scalar:

- bare if permitted by §7.1
- double-quoted otherwise

This rule ensures deterministic TOON output regardless of input authoring style.

### 15.6 Determinism Invariant

For any two TOON documents D₁ and D₂ that are semantically equivalent (same structural content, same scalar values), the following MUST hold across conformant implementations:

- canonical traCtlSpec(D₁) = canonical traCtlSpec(D₂)
- canonical JSON projection(D₁) = canonical JSON projection(D₂)
- normalized TOON re-emission(D₁) = normalized TOON re-emission(D₂)

---

## 16. Conformance Requirements

### 16.1 Producer Conformance

A conformant TOON producer MUST:

- emit UTF-8 without BOM
- emit LF line endings
- emit two-space indentation
- emit minimal valid quoting per §15.5
- preserve sequence ordering
- emit ISO 8601 durations in canonical form
- omit `_ulid` fields from authored documents
- emit no trailing whitespace
- emit no inline mappings or sequences except `{}` and `[]` for empty containers
- emit no anchors, aliases, merge keys, or tags (these are not part of the TOON syntax)

### 16.2 Parser Conformance

A conformant TOON parser MUST:

- accept UTF-8 input
- accept LF or CRLF line endings, normalizing to LF
- reject tab characters in indentation positions
- reject documents containing inline mappings or sequences other than `{}` and `[]`
- reject single-quoted strings
- reject duplicate mapping keys
- reject `_ulid` fields in non-round-trip contexts
- reject unrecognized escape sequences in double-quoted strings
- reject the YAML-style boolean tokens enumerated in §7.2 as boolean values (they parse as strings)
- reject the YAML-style null tokens enumerated in §7.3 as null values
- preserve sequence order
- discard comments before normalization
- produce a canonical source model equivalent to the canonical source model produced by a conformant YAML or JSON parser for a semantically equivalent input

### 16.3 Serializer Conformance

A conformant TOON serializer MUST:

- produce output round-trippable through a conformant parser
- produce identical output for semantically equivalent inputs
- not emit serializer-specific metadata or markers
- not emit version directives, tags, or schema annotations

### 16.4 Cross-Implementation Equivalence

Independent conformant implementations MUST produce identical canonical traCtlSpec for identical TOON input. This is the operational test of TOON determinism and is the basis for cross-format parity verification per ADR-002 §3a and `tractl_overlay_spec.md` §20.

---

## 17. Canonical Examples

### 17.1 Minimal Workflow

```toon
schemaVersion: 1
capabilities:
  - protocol.http

workflows:
  - id: health-check
    steps:
      - id: ping
        kind: request
        request:
          protocol: http
          target: "https://api.example.com/health"
          operation: GET
        assertions:
          - id: ok
            kind: status
            op: equals
            expected: 200
```

### 17.2 Workflow with Expressions and Capabilities

```toon
schemaVersion: 1
capabilities:
  - protocol.http
  - scripting.js
  - "secret.ext.vault@1"

variables:
  baseUrl: "https://api.example.com"

workflows:
  - id: user-flow
    steps:
      - id: login
        kind: request
        request:
          protocol: http
          target: "${vars.baseUrl}/auth/login"
          operation: POST
          headers:
            Content-Type: application/json
          body:
            encoding: json
            content:
              user: "${vars.username}"
              key: "${secrets.prod.api_key}"
        extracts:
          - id: token
            source: body
            path: $.access_token

      - id: fetch-profile
        kind: request
        dependsOn:
          - login
        request:
          protocol: http
          target: "${vars.baseUrl}/me"
          operation: GET
          headers:
            Authorization: "Bearer ${steps.login.extracts.token}"
        assertions:
          - id: profile-ok
            kind: status
            op: equals
            expected: 200
```

### 17.3 Overlay Document

```toon
metadata:
  name: prod-overlay
  description: Production environment augmentation

patches:
  - target:
      path: workflows.user-flow.steps.login
    patch:
      timeout: "PT30S"
      retry:
        maxAttempts: 3
        backoff: exponential
        delay: "PT1S"

  - target:
      mode: all
      match:
        kind: request
        protocol: http
    patch:
      diagnostics:
        enabled: true
        kinds:
          - tls
          - transport

  - action: remove
    target:
      path: workflows.user-flow.steps.login.request.headers.Accept
```

### 17.4 Multiline Strings

```toon
metadata:
  description: |
    This workflow validates the user authentication flow
    across staging and production environments.
  notes: >
    Keep retries conservative;
    the auth provider rate-limits aggressively.
```

---

## 18. Cross-Format Parity Guarantees

### 18.1 Bidirectional Parity

Every TOON document MUST be representable in YAML and JSON without semantic loss. Every YAML and JSON document conforming to the traCtl serialization specifications MUST be representable in TOON without semantic loss.

### 18.2 Overlay Parity

The following combinations MUST produce identical canonical traCtlSpec output:

- `workflow.toon + overlay.yaml`
- `workflow.yaml + overlay.toon`
- `workflow.json + overlay.toon`
- `openapi.yaml + overlay.toon`
- `postman.json + overlay.toon`
- `wsdl.xml + overlay.toon`

This is the operational expression of ADR-002 §3a.

### 18.3 No TOON-Exclusive Semantics

TOON MUST NOT introduce any construct that lacks an equivalent in YAML or JSON. The set of semantic constructs expressible in TOON is the intersection of constructs expressible across all three native authoring formats, governed by the canonical traCtlSpec model.

### 18.4 No Format Hierarchy

TOON is not semantically richer or poorer than YAML or JSON. All three are equivalent native authoring formats per `00_terminology.md` §1.

---

## 19. Validation Expectations

TOON parsing produces a canonical source model. Validation owners are defined by `tractl_overlay_spec.md` §15:

- **TOON parser:** syntactic validity (encoding, indentation, lexical correctness, structural well-formedness).
- **Overlay validator:** overlay schema, selector validity, target existence, merge legality, removal patch shape.
- **Canonical validator:** canonical schema, cross-reference integrity, dependency graph validity, capability declaration shape.
- **Planner:** capability resolution, runtime compatibility, composite cycle detection, execution planning.

TOON parsers MUST NOT perform canonical, overlay, or planner-owned validation. Conflating these responsibilities is a conformance defect.

---

## 20. Restrictions

The following constructs MUST NOT appear in TOON documents:

| Construct                          | Reason                                                  |
|------------------------------------|---------------------------------------------------------|
| Tab characters in indentation      | Determinism                                             |
| Single-quoted strings              | Cross-format parity (eliminate YAML 1.2 single-quote escape rules) |
| Inline mappings (other than `{}`)  | Diff clarity, parser simplicity                         |
| Inline sequences (other than `[]`) | Diff clarity, parser simplicity                         |
| YAML anchors (`&`)                 | No hidden inheritance                                   |
| YAML aliases (`*`)                 | No hidden inheritance                                   |
| YAML merge keys (`<<`)             | No hidden inheritance                                   |
| YAML tags (`!`, `!!`)              | No implicit type coercion outside canonical model       |
| YAML directives (`%YAML`, `%TAG`)  | Not applicable to TOON                                  |
| Document separators (`---`, `...`) | TOON is single-document                                 |
| Implicit boolean tokens (yes/no/on/off) | Norway problem                                     |
| Octal, hex, binary number literals | JSON number compatibility                               |
| `+`, `.inf`, `.nan` numerics       | JSON number compatibility                               |
| `_ulid` keys in authored documents | System-assigned identity                                |
| Reserved ID prefixes (`_traCtl.`, `traCtl.`) | Reserved namespace per `tractl_spec.md` §4.1   |
| BOM in serialized output           | UTF-8 conformance                                       |
| Trailing whitespace in serialized output | Determinism                                       |
| Literal secret material            | Per ADR-006 §3                                          |

---

## 21. Forward Compatibility

### 21.1 TOON Specification Versioning

This specification is versioned independently of traCtlSpec (`schemaVersion`). The TOON specification version is declared at the top of this document.

A TOON document does NOT carry an explicit TOON specification version marker. The `schemaVersion` field within the document declares the canonical traCtlSpec version; TOON serialization conformance is governed by the version of this specification active at parse time.

### 21.2 Additive Changes

The following changes are additive and do NOT bump the TOON specification version:

- new canonical traCtlSpec fields representable through existing TOON syntax
- new capability identifiers
- new platform capability flags
- new extension capability declarations

### 21.3 Breaking Changes

The following changes MUST bump the TOON specification version:

- syntactic changes (new lexical forms, removed forms)
- changes to scalar normalization
- changes to ordering rules
- changes to quoting rules
- changes to escape handling
- changes to comment handling
- changes to indentation requirements

### 21.4 Schema Version Independence

A TOON document declaring `schemaVersion: N` MUST be parsable by any TOON specification version. Canonical compatibility is governed by `tractl_spec.md` §18.4.

### 21.5 Forward Compatibility Invariant

A TOON document authored against TOON specification version M MUST remain parsable by TOON specification versions ≥ M, subject to canonical traCtlSpec compatibility rules. Breaking TOON specification changes (§21.3) MAY require document migration; such migration MUST be deterministic and tool-supported.

---

## 22. Governance

Per ADR-010, this specification sits at hierarchy position 4 (schemas and wire contracts).

Conflicts between this specification and:

- ADRs (position 1) → ADR wins
- canonical architecture specifications (position 2) → canonical spec wins
- `00_terminology.md` (position 3) → terminology registry wins
- `tractl_spec.md` (position 4, canonical semantic) → traCtlSpec wins
- `tractl_overlay_spec.md` (position 4, overlay) → overlay spec wins
- HLD, PRD, roadmap, README (lower positions) → this specification wins

Cross-format parity obligations are owned jointly with the YAML and JSON serialization specifications. A parity violation in any of the three is a conformance defect in all three.
