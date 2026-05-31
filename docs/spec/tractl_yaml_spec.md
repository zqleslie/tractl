# traCtl YAML Serialization Specification v1.0

> Version 1.0 · Normative YAML Authoring & Serialization Specification
> Classification: **Normative**
> Date: 2026-05-24

## 1. Purpose

This specification defines the normative YAML serialization contract for traCtl.

YAML is one of three native authoring formats (TOON, YAML, JSON) per ADR-002 §2 and `00_terminology.md` §1. YAML is the **default save format** for the Alpha release and is the most familiar authoring format for the primary developer persona.

To preserve determinism, cross-format parity, and import safety, traCtl consumes a strictly bounded **subset** of YAML 1.2. The full YAML 1.2 grammar admits constructs (anchors, aliases, merge keys, implicit type coercion, language-specific tags) that introduce nondeterministic parsing, hidden inheritance, and silent semantic divergence across implementations. These constructs are prohibited.

This document defines:

- the YAML 1.2 subset accepted by traCtl
- representation rules for canonical traCtlSpec concepts
- canonical normalization expectations
- producer, parser, and serializer conformance obligations
- cross-format parity obligations with TOON and JSON
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

- YAML 1.2 subset definition
- Forbidden YAML feature enumeration
- YAML representation of all canonical traCtlSpec constructs
- YAML representation of overlay documents
- Canonical YAML normalization rules
- Producer, parser, and serializer conformance obligations
- YAML ↔ TOON ↔ JSON parity obligations
- Forward compatibility rules

### Out of Scope

- canonical semantic execution behavior
- overlay merge semantics
- planner or runtime behavior
- protocol provider semantics
- YAML editor or tooling implementation
- YAML parser implementation internals beyond conformance obligations

---

## 3. Design Principles

1. **YAML is a serialization format, not a semantic model.** The canonical semantic model is traCtlSpec.
2. **Determinism over expressiveness.** YAML features that admit nondeterministic parsing across conformant YAML 1.2 implementations MUST be prohibited.
3. **Cross-format parity.** YAML MUST NOT carry semantics expressible in YAML but not expressible in TOON or JSON.
4. **No implicit type coercion.** YAML implicit typing is a documented source of bugs (the "Norway problem", octal coercion, timestamp ambiguity). Implicit type coercion MUST be eliminated by construction.
5. **No hidden inheritance.** YAML anchors, aliases, and merge keys produce documents whose semantic content cannot be determined from local context. These are prohibited.
6. **Import safety.** Per ADR-006 §6, imported artifacts (including YAML) are untrusted by default and pass through a mandatory validation pipeline. YAML subset restrictions reduce the attack surface of that pipeline.
7. **Familiarity preserved.** Despite restrictions, the YAML subset MUST remain idiomatic and recognizable to YAML authors. Restrictions target footguns, not ergonomics.

---

## 4. Serialization Philosophy

YAML is the default save format for the Alpha release per ADR-002 §4 and `00_terminology.md` §1.

YAML is **not**:

- canonical execution truth
- more authoritative than TOON or JSON
- a runtime input format
- a wire format for planner or executor consumption

Per ADR-001 §2 and ADR-002 §1, the canonical semantic model and the serialization format are separate concerns. YAML is one of several representations that normalize into traCtlSpec.

A document authored in YAML, when semantically equivalent to a document authored in TOON or JSON, MUST produce identical canonical traCtlSpec output per ADR-002 §3 and §3a.

---

## 5. Canonical Relationship to traCtlSpec

YAML sits in the Input Layer of the canonical pipeline defined in `tractl_spec.md` §2:

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

YAML parsing produces a canonical source model. YAML itself is never consumed by the planner, scheduler, or runtime.

A YAML document MUST be normalizable into a structurally valid traCtlSpec document candidate. Canonical validation occurs after overlay resolution per `tractl_spec.md` §2 and `tractl_overlay_spec.md` §3.

---

## 6. Accepted YAML Subset

traCtl consumes a strict subset of YAML 1.2 core schema.

### 6.1 Base Standard

The base standard is **YAML 1.2.2**, **core schema** (RFC-aligned JSON-compatible interpretation), with the additional restrictions enumerated in this specification.

YAML 1.1 implicit typing rules MUST NOT be applied. YAML 1.1 directives MUST be rejected.

### 6.2 Document Structure

A traCtl YAML document MUST contain exactly one YAML document.

Multi-document streams (multiple `---`-separated documents) MUST be rejected.

The document-start marker (`---`) MAY appear at the start of the document. The document-end marker (`...`) MUST NOT appear.

### 6.3 Directives

YAML directives (`%YAML`, `%TAG`, and any other `%`-prefixed lines) MUST NOT appear in traCtl YAML documents.

A YAML document containing directives MUST be rejected during parsing.

### 6.4 Encoding

YAML documents MUST be encoded as UTF-8 without a byte-order mark (BOM).

UTF-16 and UTF-32 YAML encodings (permitted by YAML 1.2) MUST be rejected.

Conformant parsers MAY accept and silently strip a leading BOM for compatibility with editors that emit one, but MUST NOT emit a BOM during serialization.

### 6.5 Line Endings

YAML documents MUST use line feed (`U+000A`) as the line terminator on output.

Conformant parsers MUST accept LF, CRLF, and CR line endings on input and normalize to LF.

### 6.6 Indentation

Indentation MUST use space characters (`U+0020`) only. Tab characters in indentation positions MUST be rejected.

Indentation width is not fixed by this specification but MUST be consistent within a single block-level structure. Producers SHOULD emit two-space indentation for cross-format consistency with TOON.

---

## 7. Forbidden YAML Features

The following YAML 1.2 features are explicitly prohibited in traCtl YAML documents. Conformant parsers MUST reject documents containing them.

### 7.1 Anchors and Aliases

Anchor declarations (`&name`) and alias references (`*name`) MUST NOT appear.

**Rationale:** Anchors and aliases produce documents whose semantic content depends on non-local references. They defeat single-pass parsing, complicate diff review, and admit cyclic structures that crash naive parsers. They are also the most common YAML-billion-laughs denial-of-service vector. Eliminating them by construction closes the entire class.

### 7.2 Merge Keys

The merge key (`<<:`) MUST NOT appear.

**Rationale:** Merge keys are not part of YAML 1.2 core; they are a YAML 1.1 extension preserved by some parsers and not others. Their behavior is implementation-specific (deep vs shallow merge, override semantics) and they violate the no-hidden-inheritance principle.

### 7.3 Explicit Tags

YAML explicit tags (`!`, `!!`, custom `!<tag>` annotations) MUST NOT appear.

**Rationale:** Tags request specific type coercions (`!!int`, `!!str`, `!!binary`, `!!timestamp`). Tag handling is implementation-divergent, and language-specific tags (Python's `!!python/object`, Ruby's `!ruby/object`) are documented remote-code-execution vectors. Eliminating tags eliminates the attack surface.

### 7.4 Implicit Typing — Booleans

YAML 1.1 implicit boolean tokens MUST NOT be coerced to booleans:

| Token   | Treatment in traCtl YAML                |
|---------|-----------------------------------------|
| `yes`   | String                                  |
| `Yes`   | String                                  |
| `YES`   | String                                  |
| `no`    | String                                  |
| `No`    | String                                  |
| `NO`    | String (the "Norway problem")           |
| `on`    | String                                  |
| `On`    | String                                  |
| `ON`    | String                                  |
| `off`   | String                                  |
| `Off`   | String                                  |
| `OFF`   | String                                  |
| `y`     | String                                  |
| `n`     | String                                  |
| `True`  | String                                  |
| `False` | String                                  |
| `TRUE`  | String                                  |
| `FALSE` | String                                  |

The only valid YAML boolean tokens recognized by traCtl are `true` and `false` (all-lowercase).

### 7.5 Implicit Typing — Null

The only valid null token is `null` (all-lowercase).

The YAML 1.2 alternatives MUST NOT be coerced to null:

| Token   | Treatment in traCtl YAML |
|---------|--------------------------|
| `~`     | Rejected as ambiguous    |
| `Null`  | String                   |
| `NULL`  | String                   |
| empty scalar in value position | Rejected as a syntax error |

### 7.6 Implicit Typing — Numbers

Number literals MUST follow the JSON number grammar (RFC 8259 §6).

The following YAML number forms are NOT recognized as numbers and parse as strings:

- octal literals (`0o17`, `017`)
- hexadecimal literals (`0x1F`)
- binary literals (`0b101`)
- sexagesimal literals (`12:34:56`)
- underscored numerics (`1_000_000`)
- positive sign prefix (`+42`)
- the YAML floats `.inf`, `-.inf`, `.nan`, `Infinity`, `NaN`

Numbers with leading zeros MUST be rejected.

### 7.7 Implicit Typing — Timestamps

YAML implicit timestamp coercion MUST NOT be applied.

Tokens such as `2026-05-24`, `2026-05-24T12:00:00Z`, and timezone-less timestamps MUST be parsed as strings. Date and time semantics are governed by canonical traCtlSpec field schemas, not by YAML implicit typing.

> **⚠️ Implementation Note for Tooling Authors:** Many widely-used YAML libraries (PyYAML, Ruby's Psych, Go's `gopkg.in/yaml.v2`, SnakeYAML) apply YAML 1.1 implicit timestamp coercion by default, converting bare date literals like `date: 2026-05-24` into native `Date` or `Time` objects without any explicit tag. This automatic coercion happens silently at parse time and produces a typed object rather than the string `"2026-05-24"`. traCtl parsers MUST configure their YAML library to suppress implicit timestamp coercion — this is typically a non-default option. Parsers that fail to suppress this behavior MUST be considered non-conformant. Test conformance by asserting that `date: 2026-05-24` produces the string `"2026-05-24"`, not a date object.

### 7.8 Binary Tag

The `!!binary` tag and inline base64 binary literals MUST NOT appear.

Binary data, where required, is referenced through canonical traCtlSpec mechanisms (e.g., `body.encoding: binary` with `body.content` carrying an opaque string per the canonical spec) — not embedded as YAML binary literals.

### 7.9 Complex Keys

Complex (non-scalar) mapping keys MUST NOT appear.

The YAML `? key` notation, sequence-as-key, and mapping-as-key forms MUST be rejected.

All mapping keys MUST be scalar strings.

### 7.10 Duplicate Keys

Duplicate keys within the same mapping MUST be rejected.

YAML 1.2 specifies duplicate-key behavior as implementation-defined. traCtl makes it an explicit error.

### 7.11 Flow Style

YAML flow-style mappings (`{key: value}`) and flow-style sequences (`[a, b, c]`) MUST NOT appear, except for the empty-container forms `{}` and `[]`.

**Rationale:** Flow style is harder to diff, introduces quoting rules that diverge from block style, and admits a JSON-incompatible subset (unquoted strings containing commas, braces). Eliminating it closes the divergence.

The empty-container exceptions `{}` and `[]` are permitted because they have no ambiguity and round-trip cleanly to JSON.

### 7.12 Single-Quoted Strings

Single-quoted string literals MUST NOT appear.

**Rationale:** YAML single-quote escape rules (only `''` is recognized) diverge from double-quote escape rules and from JSON string semantics. Eliminating single-quoted strings produces a unified escape model across YAML, TOON, and JSON.

### 7.13 Implicit Document Start in Bare Streams

Multi-document streams (per §6.2) are prohibited. The `---` document-start marker MAY appear once at the start of the document. Subsequent `---` markers MUST be rejected.

---

## 8. Scalar Representation

### 8.1 Strings

Strings have three accepted forms:

| Form         | Delimiter | Escape processing | Multi-line |
|--------------|-----------|-------------------|------------|
| Plain (bare) | none      | none              | no         |
| Double-quoted| `"…"`     | yes (§8.1.1)      | no         |
| Block        | `|` or `>`| no                | yes (§8.4) |

Plain (bare) strings are permitted only when the value does not collide with implicit typing forms enumerated in §7.4–§7.7 and does not contain reserved characters (`:`, `#`, `,`, `[`, `]`, `{`, `}`, leading `&`, leading `*`, leading `!`, leading `?`, leading `|`, leading `>`).

Otherwise, the value MUST be double-quoted.

#### 8.1.1 Escape Sequences

Double-quoted strings recognize the following escapes:

| Escape  | Meaning              |
|---------|----------------------|
| `\"`    | Double quote         |
| `\\`    | Backslash            |
| `\n`    | Line feed (U+000A)   |
| `\r`    | Carriage return      |
| `\t`    | Tab (U+0009)         |
| `\uXXXX`| Unicode code point   |

No other escape sequences are recognized. The YAML-specific escapes `\a`, `\v`, `\0`, `\NEL`, `\L`, `\P`, `\_`, `\N`, and unquoted whitespace-folding rules MUST NOT be applied.

### 8.2 Booleans

The only valid boolean tokens are `true` and `false`. See §7.4 for the prohibition on YAML 1.1 boolean variants.

### 8.3 Null

The only valid null token is `null`. See §7.5 for the prohibition on YAML 1.2 null variants.

### 8.4 Block Strings

Multi-line strings use either of two block indicators:

- `|` — literal block: line breaks preserved.
- `>` — folded block: consecutive non-empty lines folded to single space-separated line; blank lines preserved as line breaks.

Trailing newline behavior:

- Default: single trailing newline preserved.
- `|-` and `>-`: trailing newlines stripped.
- `|+` and `>+`: all trailing newlines preserved.

These are the only YAML chomping indicators recognized. Other YAML block scalar features (explicit indentation indicators such as `|2`, `|+1`) MUST NOT appear.

### 8.5 Durations

Durations MUST be represented as ISO 8601 duration strings (e.g., `PT5S`, `PT1M30S`, `PT2H`), consistent with `tractl_spec.md` §17.

Duration values containing characters that trigger implicit YAML coercion (e.g., values containing `:`) MUST be double-quoted to prevent misinterpretation.

---

## 9. Structural Constructs

### 9.1 Mappings

Mappings (objects) MUST be expressed in block form:

```yaml
metadata:
  name: my-workflow
  description: example
```

Flow-style mappings other than `{}` (the empty-container form) MUST NOT appear.

### 9.2 Sequences

Sequences (arrays) MUST be expressed in block form:

```yaml
patches:
  - target:
      path: workflows.user.steps.login
    patch:
      timeout: "PT30S"
```

Flow-style sequences other than `[]` (the empty-container form) MUST NOT appear.

### 9.3 Empty Containers

Empty mappings and sequences are represented as `{}` and `[]` respectively. These are the only inline structural forms permitted.

### 9.4 Nesting

Nesting depth is unbounded.

---

## 10. Identity Representation

Identity rules align with `tractl_spec.md` §4 and are equivalent across TOON, YAML, and JSON.

### 10.1 User IDs

User IDs MUST be represented as scalar string values bound to the `id` key:

```yaml
id: login-step
```

User IDs MUST conform to the pattern in `tractl_spec.md` §4.1. Reserved prefixes (`_traCtl.`, `traCtl.`) MUST NOT be used in authored YAML.

### 10.2 ULIDs

ULIDs (`_ulid` fields) MUST NOT appear in authored YAML documents. ULIDs are system-assigned at normalization time per `tractl_spec.md` §4.2.

A YAML document containing `_ulid` keys MUST be rejected as a producer error during parsing, unless the document is being re-emitted by a tool that has explicitly opted into round-trip preservation.

### 10.3 References

Cross-entity references in YAML use scalar string values containing dotted paths or expression interpolation per §11:

```yaml
dependsOn:
  - login
extracts:
  - id: token
    source: body
    path: $.access_token
```

Reference resolution is a canonical-validator responsibility per `tractl_spec.md` §6.2.

---

## 11. Expression Serialization

Per `tractl_spec.md` §13.5, traCtlSpec supports `${ ... }` interpolation within scalar values.

### 11.1 Inline Expressions

Expressions MUST be embedded as substring content within scalar string values:

```yaml
url: "https://api.example.com/users/${vars.userId}"
authorization: "Bearer ${steps.login.extracts.token}"
```

### 11.2 Quoting Requirements

Any scalar value containing a `${` token MUST be double-quoted. This avoids ambiguity with plain-scalar parsing and ensures deterministic round-tripping to JSON.

### 11.3 Expression Grammar

The expression grammar is owned by `tractl_spec.md` §13.5 and is NOT redefined here. YAML serialization treats expressions as opaque substring content within string values.

### 11.4 Escaping

A literal `${` sequence that MUST NOT be interpreted as an expression MUST be escaped as `\${`.

### 11.5 No Standalone Expression Type

Expressions are NOT a distinct YAML scalar type. They are string-typed values whose contents are interpreted by the canonical expression evaluator at runtime. YAML parsers MUST NOT evaluate, partially evaluate, or transform expression content during normalization.

---

## 12. Secret Reference Serialization

Per ADR-006 §3, secrets are never directly embedded in workflow definitions. Secrets are referenced through scoped injection.

### 12.1 Secret Reference Form

Secret references MUST be expressed as expressions resolving against a canonical secret namespace:

```yaml
apiKey: "${secrets.prod.stripe_api_key}"
```

### 12.2 Prohibitions

The following MUST be rejected by conformant YAML producers and MUST be flagged by conformant YAML parsers as policy violations:

- literal secret material in any scalar position
- inline base64-encoded credentials
- raw API tokens
- raw passwords
- raw private keys

Detection of literal secrets is best-effort. The architectural prohibition on literal secrets is owned by ADR-006 §3 and is enforced by the secret resolution pipeline; YAML-layer detection is an authoring-time hint only.

### 12.3 No Secret-Specific YAML Syntax

YAML MUST NOT introduce secret-specific syntactic constructs. Secrets are referenced through the standard expression mechanism (§11). This preserves cross-format parity with TOON and JSON.

---

## 13. Capability Serialization

Per `tractl_spec.md` §3.2, capabilities are tiered: platform capabilities (stable flags, no version suffix) and extension/provider capabilities (mandatory `identifier@version`).

### 13.1 Platform Capabilities

Platform capabilities MUST be serialized as bare or quoted identifier strings without version suffix:

```yaml
capabilities:
  - protocol.http
  - scripting.js
  - diagnostics.tls
```

### 13.2 Extension and Provider Capabilities

Extension and provider capabilities MUST be serialized as identifier-with-version strings in canonical `identifier@version` form:

```yaml
capabilities:
  - "protocol.ext.amqp@1"
  - "secret.ext.vault@1"
  - "assertion.ext.jsonschema@2"
```

Capability strings containing `@` MUST be double-quoted to ensure deterministic JSON projection and to avoid any ambiguity with YAML implicit typing.

### 13.3 Capability Source Independence

YAML capability serialization MUST NOT vary by capability source (native engine, extension, MCP adapter, provider adapter), per ADR-011 §2.

### 13.4 Capability Ordering

Capability arrays MUST preserve user-authored order during serialization round-trips.

---

## 14. Overlay Serialization

Overlay documents authored in YAML MUST conform to the canonical overlay document model defined in `tractl_overlay_spec.md` §8.

### 14.1 Document Shape

```yaml
metadata:
  name: dev-overlay

patches:
  - target:
      path: workflows.user-check.steps.login
    patch:
      timeout: "PT30S"
```

Only `metadata` (optional) and `patches` (required) are valid top-level keys.

### 14.2 Target Selectors

Target selectors MUST follow the three-mode model in `tractl_overlay_spec.md` §9. Each target MUST use exactly one of `path`, `match`, or `source`.

### 14.3 Action Field

The `action` field, when present, MUST contain one of the canonical action literals defined in `tractl_overlay_spec.md` §10.4:

- `replace`
- `deepMerge`
- `append`
- `appendUnique`
- `remove`

### 14.4 Removal Patches

Removal patches (`action: remove`) MUST NOT contain a `patch` field, per `tractl_overlay_spec.md` §11.

### 14.5 Cross-Format Parity

A YAML overlay MUST normalize identically to a semantically equivalent TOON or JSON overlay. The following combinations MUST produce identical canonical traCtlSpec output per ADR-002 §3a:

- `workflow.yaml + overlay.toon`
- `workflow.toon + overlay.yaml`
- `workflow.json + overlay.yaml`
- `openapi.yaml + overlay.yaml`
- `postman.json + overlay.yaml`

---

## 15. Ordering Rules

### 15.1 Mapping Key Ordering

Mappings MAY be authored in any key order. Canonical normalization MUST preserve authored order for round-trip emission to YAML.

Canonical JSON projection per `tractl_spec.md` §17 sorts keys lexicographically. YAML-to-canonical-JSON normalization MUST apply lexicographic sorting at the projection boundary, not at YAML ingest.

### 15.2 Sequence Element Ordering

Sequence (array) element order MUST be preserved across normalization, per `tractl_spec.md` §17 and `tractl_overlay_spec.md` §20.

### 15.3 Patch Application Order

Within an overlay document, `patches` order MUST be preserved as authored per `tractl_overlay_spec.md` §12. YAML serializers MUST NOT reorder patch entries.

---

## 16. Deterministic Normalization Expectations

### 16.1 Whitespace

Insignificant whitespace MUST NOT affect canonical output. Trailing whitespace, blank-line counts between blocks, and indentation width (within consistency requirements) are not semantic.

### 16.2 Comments

YAML comments (`# ...`) MUST be discarded during normalization and MUST NOT appear in canonical traCtlSpec output.

### 16.3 Scalar Normalization

- Strings: Unicode normalization form NFC MUST be applied.
- Booleans: normalized to `true` / `false`.
- Null: normalized to `null`.
- Numbers: integers preserved as JSON integers; floats preserved as JSON numbers with shortest round-trip representation per RFC 7493.
- Durations: normalized to canonical ISO 8601 duration strings.

### 16.4 Expression Normalization

Expression content within strings MUST be preserved verbatim. YAML normalization MUST NOT modify expression syntax, whitespace within expressions, or expression structure.

### 16.5 Quoting Normalization

Round-trip serialization to YAML MUST emit the minimal valid quoting form for each scalar:

- plain if permitted by §8.1
- double-quoted otherwise

### 16.6 Determinism Invariant

For any two YAML documents D₁ and D₂ that are semantically equivalent, the following MUST hold across conformant implementations:

- canonical traCtlSpec(D₁) = canonical traCtlSpec(D₂)
- canonical JSON projection(D₁) = canonical JSON projection(D₂)
- normalized YAML re-emission(D₁) = normalized YAML re-emission(D₂)

---

## 17. Conformance Requirements

### 17.1 Producer Conformance

A conformant YAML producer MUST:

- emit UTF-8 without BOM
- emit LF line endings
- emit consistent indentation (two spaces SHOULD be used)
- emit minimal valid quoting per §16.5
- preserve sequence ordering
- emit ISO 8601 durations in canonical form
- omit `_ulid` fields from authored documents
- emit no trailing whitespace
- emit no flow-style constructs except `{}` and `[]`
- emit no anchors, aliases, merge keys, tags, or directives
- emit no single-quoted strings
- emit `true`/`false`/`null` lowercase only

### 17.2 Parser Conformance

A conformant YAML parser MUST:

- accept UTF-8 input (BOM accepted and stripped)
- accept LF, CRLF, or CR line endings, normalizing to LF
- reject tab characters in indentation positions
- reject anchors (`&`), aliases (`*`), and merge keys (`<<`)
- reject explicit tags (`!`, `!!`, `!<tag>`)
- reject YAML directives (`%`)
- reject multi-document streams
- reject flow-style mappings and sequences except `{}` and `[]`
- reject single-quoted strings
- reject complex (non-scalar) keys
- reject duplicate mapping keys
- reject `_ulid` fields in non-round-trip contexts
- reject unrecognized escape sequences in double-quoted strings
- treat YAML 1.1 boolean variants (yes/no/on/off/Y/N) as strings
- treat YAML null variants (`~`, `Null`, `NULL`) as strings or errors per §7.5
- reject YAML implicit timestamp coercion
- reject octal, hex, binary number literals
- reject `+`, `.inf`, `.nan` numerics
- preserve sequence order
- discard comments before normalization
- produce a canonical source model equivalent to the canonical source model produced by a conformant TOON or JSON parser for a semantically equivalent input

### 17.3 Serializer Conformance

A conformant YAML serializer MUST:

- produce output round-trippable through a conformant parser
- produce identical output for semantically equivalent inputs
- not emit serializer-specific metadata or markers
- not emit version directives, tags, or schema annotations
- not emit prohibited features (§7)

### 17.4 Cross-Implementation Equivalence

Independent conformant implementations MUST produce identical canonical traCtlSpec for identical YAML input. This is the operational test of YAML determinism and is the basis for cross-format parity verification per ADR-002 §3a.

---

## 18. Canonical Examples

### 18.1 Minimal Workflow

```yaml
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

### 18.2 Workflow with Expressions and Capabilities

```yaml
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

### 18.3 Overlay Document

```yaml
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

### 18.4 Failure Examples

The following YAML constructs MUST be rejected by conformant parsers:

```yaml
# REJECTED: anchor and alias
defaults: &defaults
  timeout: PT5S
workflows:
  - id: a
    <<: *defaults    # merge key — REJECTED
    steps: []
```

```yaml
# REJECTED: implicit boolean coercion
features:
  fast: yes        # parses as string "yes" in traCtl, but the
                   # producer SHOULD have written: true
                   # Acceptable forms: true | false (lowercase only)
```

```yaml
# REJECTED: explicit tag
secret: !!binary |
  R0lGODlhDAAMAIQAAP//9/X...
```

```yaml
# REJECTED: flow-style mapping
inline: { key: value, other: 1 }
```

```yaml
# REJECTED: multi-document stream
schemaVersion: 1
workflows: []
---
schemaVersion: 1
workflows: []
```

---

## 19. Cross-Format Parity Guarantees

### 19.1 Bidirectional Parity

Every YAML document conforming to this specification MUST be representable in TOON and JSON without semantic loss. Every TOON or JSON document conforming to the traCtl serialization specifications MUST be representable in YAML without semantic loss.

### 19.2 Overlay Parity

The following combinations MUST produce identical canonical traCtlSpec output per ADR-002 §3a:

- `workflow.yaml + overlay.toon`
- `workflow.toon + overlay.yaml`
- `workflow.json + overlay.yaml`
- `openapi.yaml + overlay.yaml`
- `postman.json + overlay.yaml`

### 19.3 No YAML-Exclusive Semantics

YAML MUST NOT introduce any construct that lacks an equivalent in TOON or JSON. The set of semantic constructs expressible in YAML is the intersection of constructs expressible across all three native authoring formats, governed by the canonical traCtlSpec model.

### 19.4 No Format Hierarchy

YAML is not semantically richer or poorer than TOON or JSON, despite being the default save format. All three are equivalent native authoring formats per `00_terminology.md` §1.

---

## 20. Validation Expectations

YAML parsing produces a canonical source model. Validation owners are defined by `tractl_overlay_spec.md` §15:

- **YAML parser:** syntactic validity (encoding, indentation, lexical correctness, structural well-formedness, subset compliance).
- **Overlay validator:** overlay schema, selector validity, target existence, merge legality, removal patch shape.
- **Canonical validator:** canonical schema, cross-reference integrity, dependency graph validity, capability declaration shape.
- **Planner:** capability resolution, runtime compatibility, composite cycle detection, execution planning.

YAML parsers MUST NOT perform canonical, overlay, or planner-owned validation.

---

## 21. Restrictions Summary

The following constructs MUST NOT appear in traCtl YAML documents:

| Construct                                           | Reason                                  |
|-----------------------------------------------------|-----------------------------------------|
| Anchors (`&`)                                       | No hidden inheritance; DoS surface      |
| Aliases (`*`)                                       | No hidden inheritance; cycle risk       |
| Merge keys (`<<:`)                                  | Implementation-divergent semantics      |
| Explicit tags (`!`, `!!`, `!<tag>`)                 | RCE surface; implementation divergence  |
| YAML directives (`%YAML`, `%TAG`)                   | Not part of accepted subset             |
| Multi-document streams                              | Single-document only                    |
| UTF-16 / UTF-32 encoding                            | UTF-8 only                              |
| BOM in serialized output                            | UTF-8 conformance                       |
| Tab characters in indentation                       | Determinism                             |
| Single-quoted strings                               | Cross-format escape parity              |
| Flow-style mappings other than `{}`                 | Diff clarity                            |
| Flow-style sequences other than `[]`                | Diff clarity                            |
| Complex (non-scalar) mapping keys                   | Scalar keys only                        |
| Duplicate keys in same mapping                      | Determinism                             |
| YAML 1.1 boolean variants (yes/no/on/off/Y/N)       | Norway problem                          |
| YAML null variants (`~`, `Null`, `NULL`)            | Ambiguity                               |
| Implicit timestamp coercion                         | Cross-format parity                     |
| Implicit binary tag (`!!binary`)                    | Cross-format parity                     |
| Octal, hex, binary number literals                  | JSON number compatibility               |
| `+`, `.inf`, `.nan` numerics                        | JSON number compatibility               |
| Underscored numerics (`1_000_000`)                  | JSON number compatibility               |
| Sexagesimal numerics (`12:34:56`)                   | YAML 1.1 holdover                       |
| `_ulid` keys in authored documents                  | System-assigned identity                |
| Reserved ID prefixes (`_traCtl.`, `traCtl.`)        | Reserved namespace                      |
| Trailing whitespace in serialized output            | Determinism                             |
| Literal secret material                             | Per ADR-006 §3                          |

---

## 22. Forward Compatibility

### 22.1 YAML Specification Versioning

This specification is versioned independently of traCtlSpec (`schemaVersion`).

A YAML document does NOT carry an explicit YAML-subset version marker. The `schemaVersion` field within the document declares the canonical traCtlSpec version; YAML serialization conformance is governed by the version of this specification active at parse time.

### 22.2 Additive Changes

The following changes are additive and do NOT bump this specification's version:

- new canonical traCtlSpec fields representable through existing YAML syntax
- new capability identifiers
- new platform capability flags
- new extension capability declarations

### 22.3 Breaking Changes

The following changes MUST bump this specification's version:

- relaxation of forbidden YAML features
- changes to scalar normalization
- changes to ordering rules
- changes to quoting rules
- changes to escape handling
- changes to subset boundaries

### 22.4 Schema Version Independence

A YAML document declaring `schemaVersion: N` MUST be parsable by any YAML-specification version. Canonical compatibility is governed by `tractl_spec.md` §18.4.

### 22.5 Future Relaxation

This specification deliberately rejects YAML features whose semantics are stable but whose surface is hostile to determinism. Future relaxation (e.g., admitting flow style under strict quoting constraints) is reserved for explicit specification revisions; the default posture is **conservative**.

---

## 23. Governance

Per ADR-010, this specification sits at hierarchy position 4 (schemas and wire contracts).

Conflicts between this specification and:

- ADRs (position 1) → ADR wins
- canonical architecture specifications (position 2) → canonical spec wins
- `00_terminology.md` (position 3) → terminology registry wins
- `tractl_spec.md` (position 4, canonical semantic) → traCtlSpec wins
- `tractl_overlay_spec.md` (position 4, overlay) → overlay spec wins
- HLD, PRD, roadmap, README (lower positions) → this specification wins

Cross-format parity obligations are owned jointly with the TOON and JSON serialization specifications. A parity violation in any of the three is a conformance defect in all three.
