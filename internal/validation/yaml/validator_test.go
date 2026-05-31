package yaml_test

// Go concept — package naming for tests:
// Using "yaml_test" (not "yaml") is the idiomatic Go pattern for black-box
// testing. It means this test file can only access the package's exported
// symbols — the same surface any external caller sees. This prevents tests
// from accidentally relying on unexported internals.
//
// Think of it as testing the npm package's public API, not its source files.

import (
	"strings"
	"testing"

	// We import our own package under the alias "yamlval" to avoid a name
	// collision with the standard "yaml" naming convention.
	// Replace "github.com/tractl/tractl" with your actual module path from go.mod.
	yamlval "github.com/tractl/tractl/internal/validation/yaml"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

// validate is a convenience wrapper that converts the string to bytes and calls Validate.
func validate(t *testing.T, input string) *yamlval.ValidationResult {
	t.Helper() // marks this as a helper so failures point to the caller line, not here
	return yamlval.Validate([]byte(input))
}

// assertValid fails the test if the document is not accepted.
func assertValid(t *testing.T, input string) {
	t.Helper()
	result := validate(t, input)
	if !result.Valid {
		t.Fatalf("expected valid document but got %d error(s):\n%s",
			len(result.Errors), formatErrors(result))
	}
}

// assertInvalid fails the test if the document is accepted, or if the expected
// error code is not present among the collected violations.
func assertInvalid(t *testing.T, input string, wantCode yamlval.ErrorCode) {
	t.Helper()
	result := validate(t, input)
	if result.Valid {
		t.Fatalf("expected validation failure with code %s but document was accepted", wantCode)
	}
	for _, e := range result.Errors {
		if e.Code == wantCode {
			return // found it
		}
	}
	t.Fatalf("expected error code %s not found; got:\n%s", wantCode, formatErrors(result))
}

// formatErrors renders all errors in a result to a readable multi-line string.
func formatErrors(r *yamlval.ValidationResult) string {
	var sb strings.Builder
	for _, e := range r.Errors {
		sb.WriteString("  • " + e.Error() + "\n")
	}
	return sb.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// Valid documents — the spec's canonical examples must pass
// ─────────────────────────────────────────────────────────────────────────────

func TestValid_MinimalWorkflow(t *testing.T) {
	// spec §18.1
	assertValid(t, `
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
`)
}

func TestValid_WorkflowWithExpressions(t *testing.T) {
	// spec §18.2
	assertValid(t, `
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
`)
}

func TestValid_OverlayDocument(t *testing.T) {
	// spec §18.3
	assertValid(t, `
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
`)
}

func TestValid_DocumentStartMarkerPermitted(t *testing.T) {
	// spec §6.2: the document-start marker (---) MAY appear.
	assertValid(t, `---
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: w
    steps: []
`)
}

func TestValid_EmptyFlowContainers(t *testing.T) {
	// spec §7.11: {} and [] are the permitted empty-container exceptions.
	assertValid(t, `
schemaVersion: 1
capabilities: []
variables: {}
workflows: []
`)
}

func TestValid_CanonicalBooleans(t *testing.T) {
	assertValid(t, "enabled: true\ndisabled: false\n")
}

func TestValid_CanonicalNull(t *testing.T) {
	// "null" (all-lowercase) is the only accepted null form.
	assertValid(t, "value: null\n")
}

func TestValid_CanonicalNumbers(t *testing.T) {
	assertValid(t, `
integer: 42
negative: -7
zero: 0
float: 3.14
negative_float: -0.5
zero_float: 0.0
`)
}

func TestValid_BlockStrings(t *testing.T) {
	assertValid(t, `
literal: |
  line one
  line two
folded: >
  this is
  folded text
stripped: |-
  no trailing newline
`)
}

func TestValid_DoubleQuotedStrings(t *testing.T) {
	// Double-quoted strings, including those with expressions.
	assertValid(t, `
plain: hello world
quoted: "quoted value"
with_expr: "${vars.baseUrl}/users"
nested_quotes: "she said \"hello\""
`)
}

func TestValid_UTF8BOMIsStripped(t *testing.T) {
	// spec §6.4: a leading UTF-8 BOM is silently stripped.
	bom := []byte{0xEF, 0xBB, 0xBF}
	doc := []byte("schemaVersion: 1\ncapabilities: []\nworkflows: []\n")
	result := yamlval.Validate(append(bom, doc...))
	if !result.Valid {
		t.Fatalf("expected BOM-prefixed document to be accepted; got:\n%s", formatErrors(result))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding (§6.4)
// ─────────────────────────────────────────────────────────────────────────────

func TestEncoding_UTF16BE(t *testing.T) {
	input := []byte{0xFE, 0xFF, 0x00, 'k', 0x00, 'e', 0x00, 'y', 0x00, ':', 0x00, '1'}
	result := yamlval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-16 BE input")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == yamlval.ErrEncoding {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ErrEncoding; got:\n%s", formatErrors(result))
	}
}

func TestEncoding_UTF32BE(t *testing.T) {
	input := []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 'k'}
	result := yamlval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-32 BE input")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Indentation (§6.6)
// ─────────────────────────────────────────────────────────────────────────────

func TestIndentation_TabInIndentation(t *testing.T) {
	assertInvalid(t, "key:\n\tvalue: 1\n", yamlval.ErrTabIndentation)
}

func TestIndentation_TabMidLine_IsNotCaught(t *testing.T) {
	// A tab that appears after non-whitespace content is inside a value,
	// not in an indentation position. This is outside our scope.
	// (It may still be invalid YAML for other reasons — the parser decides.)
	result := validate(t, "key: \"value\twith tab\"\n")
	// We only assert the validator doesn't report a spurious tab error.
	for _, e := range result.Errors {
		if e.Code == yamlval.ErrTabIndentation {
			t.Fatal("tab inside a value should not trigger ErrTabIndentation")
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Directives (§6.3)
// ─────────────────────────────────────────────────────────────────────────────

func TestDirective_YAMLVersion(t *testing.T) {
	assertInvalid(t, "%YAML 1.2\n---\nkey: value\n", yamlval.ErrDirective)
}

func TestDirective_TAG(t *testing.T) {
	assertInvalid(t, "%TAG ! tag:example.com,2000:\n---\nkey: value\n", yamlval.ErrDirective)
}

// ─────────────────────────────────────────────────────────────────────────────
// Document structure (§6.2)
// ─────────────────────────────────────────────────────────────────────────────

func TestDocumentEnd_MarkerRejected(t *testing.T) {
	assertInvalid(t, "key: value\n...\n", yamlval.ErrDocumentEndMarker)
}

func TestDocumentEnd_MultiStream(t *testing.T) {
	assertInvalid(t, "key: value\n---\nkey2: value2\n", yamlval.ErrMultiDocument)
}

// ─────────────────────────────────────────────────────────────────────────────
// Anchors and aliases (§7.1)
// ─────────────────────────────────────────────────────────────────────────────

func TestAnchor_Declaration(t *testing.T) {
	assertInvalid(t, "defaults: &defaults\n  timeout: 30\n", yamlval.ErrAnchor)
}

func TestAlias_Reference(t *testing.T) {
	// This document has both an anchor and an alias; both codes should appear.
	input := "a: &def\n  x: 1\nb: *def\n"
	result := validate(t, input)
	if result.Valid {
		t.Fatal("expected invalid document")
	}
	var gotAnchor, gotAlias bool
	for _, e := range result.Errors {
		if e.Code == yamlval.ErrAnchor {
			gotAnchor = true
		}
		if e.Code == yamlval.ErrAlias {
			gotAlias = true
		}
	}
	if !gotAnchor {
		t.Error("expected ErrAnchor")
	}
	if !gotAlias {
		t.Error("expected ErrAlias")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Explicit tags (§7.3)
// ─────────────────────────────────────────────────────────────────────────────

func TestExplicitTag_DoubleExclamation(t *testing.T) {
	assertInvalid(t, "value: !!str hello\n", yamlval.ErrExplicitTag)
}

func TestExplicitTag_IntTag(t *testing.T) {
	assertInvalid(t, "value: !!int 42\n", yamlval.ErrExplicitTag)
}

// ─────────────────────────────────────────────────────────────────────────────
// Single-quoted strings (§7.12)
// ─────────────────────────────────────────────────────────────────────────────

func TestSingleQuote_StringValue(t *testing.T) {
	assertInvalid(t, "value: 'single quoted'\n", yamlval.ErrSingleQuote)
}

func TestSingleQuote_KeyValue(t *testing.T) {
	assertInvalid(t, "key: 'it''s not allowed'\n", yamlval.ErrSingleQuote)
}

// ─────────────────────────────────────────────────────────────────────────────
// Flow style (§7.11)
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowStyle_NonEmptyMapping(t *testing.T) {
	assertInvalid(t, "value: {key: val}\n", yamlval.ErrFlowStyle)
}

func TestFlowStyle_NonEmptySequence(t *testing.T) {
	assertInvalid(t, "values: [a, b, c]\n", yamlval.ErrFlowStyle)
}

func TestFlowStyle_EmptyMappingPermitted(t *testing.T) {
	assertValid(t, "value: {}\n")
}

func TestFlowStyle_EmptySequencePermitted(t *testing.T) {
	assertValid(t, "values: []\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Mapping constraints (§7.2, §7.9, §7.10, §10.1, §10.2)
// ─────────────────────────────────────────────────────────────────────────────

func TestMergeKey_Rejected(t *testing.T) {
	assertInvalid(t, "defaults: &def\n  x: 1\nother:\n  <<: *def\n", yamlval.ErrMergeKey)
}

func TestDuplicateKey_SameLevel(t *testing.T) {
	assertInvalid(t, "key: value1\nkey: value2\n", yamlval.ErrDuplicateKey)
}

func TestULIDKey_Rejected(t *testing.T) {
	assertInvalid(t, "_ulid: 01ARZ3NDEKTSV4RRFFQ69G5FAV\n", yamlval.ErrULIDKey)
}

func TestReservedPrefix_Underscore(t *testing.T) {
	assertInvalid(t, "_traCtl.internal: value\n", yamlval.ErrReservedPrefix)
}

func TestReservedPrefix_Plain(t *testing.T) {
	assertInvalid(t, "traCtl.config: value\n", yamlval.ErrReservedPrefix)
}

func TestReservedPrefix_NotAPrefix(t *testing.T) {
	// "traCtlSpec" does not start with "traCtl." — no dot — so it is not reserved.
	assertValid(t, "traCtlSpec: 1\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// YAML 1.1 boolean variants (§7.4)
// ─────────────────────────────────────────────────────────────────────────────

func TestYAML11Boolean_Yes(t *testing.T) {
	assertInvalid(t, "enabled: yes\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_No(t *testing.T) {
	assertInvalid(t, "enabled: no\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_On(t *testing.T) {
	assertInvalid(t, "flag: on\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_Off(t *testing.T) {
	assertInvalid(t, "flag: off\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_YES_Uppercase(t *testing.T) {
	assertInvalid(t, "flag: YES\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_Y(t *testing.T) {
	assertInvalid(t, "flag: Y\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_TrueMixedCase(t *testing.T) {
	assertInvalid(t, "flag: True\n", yamlval.ErrYAML11Boolean)
}

func TestYAML11Boolean_FalseMixedCase(t *testing.T) {
	assertInvalid(t, "flag: False\n", yamlval.ErrYAML11Boolean)
}

// ─────────────────────────────────────────────────────────────────────────────
// Null variants (§7.5)
// ─────────────────────────────────────────────────────────────────────────────

func TestNull_Tilde(t *testing.T) {
	assertInvalid(t, "value: ~\n", yamlval.ErrProhibitedNull)
}

func TestNull_NullMixedCase(t *testing.T) {
	assertInvalid(t, "value: Null\n", yamlval.ErrProhibitedNull)
}

func TestNull_NullUppercase(t *testing.T) {
	assertInvalid(t, "value: NULL\n", yamlval.ErrProhibitedNull)
}

func TestNull_EmptyValuePosition(t *testing.T) {
	// "key:" with no value is an empty scalar resolved as null.
	assertInvalid(t, "key:\n", yamlval.ErrProhibitedNull)
}

func TestNull_CanonicalFormAccepted(t *testing.T) {
	assertValid(t, "value: null\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Timestamps (§7.7)
// ─────────────────────────────────────────────────────────────────────────────

func TestTimestamp_ISO8601DateTime(t *testing.T) {
	// yaml.v3 resolves ISO 8601 datetimes as !!timestamp.
	assertInvalid(t, "ts: 2026-05-24T12:00:00Z\n", yamlval.ErrImplicitTimestamp)
}

func TestTimestamp_BareDate_LibraryDependent(_ *testing.T) {
	// yaml.v3 may or may not resolve a bare date (2026-05-24) as !!timestamp.
	// If it does, the validator catches it. If it resolves as !!str, the
	// document passes — both outcomes are conformant per spec §7.7 note.
	// We assert only that the validator does not panic.
	input := "date: 2026-05-24\n"
	result := yamlval.Validate([]byte(input))
	_ = result // outcome is library-version dependent
}

// ─────────────────────────────────────────────────────────────────────────────
// Prohibited numeric forms (§7.6)
// ─────────────────────────────────────────────────────────────────────────────

func TestNumber_OctalPrefix_0o(t *testing.T) {
	assertInvalid(t, "mode: 0o755\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_HexPrefix(t *testing.T) {
	assertInvalid(t, "color: 0xFF\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_BinaryPrefix(t *testing.T) {
	assertInvalid(t, "flags: 0b1010\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_PositiveSign(t *testing.T) {
	assertInvalid(t, "value: +42\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_Underscored(t *testing.T) {
	assertInvalid(t, "amount: 1_000_000\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_Inf(t *testing.T) {
	assertInvalid(t, "value: .inf\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_NegativeInf(t *testing.T) {
	assertInvalid(t, "value: -.inf\n", yamlval.ErrProhibitedNumber)
}

func TestNumber_Nan(t *testing.T) {
	assertInvalid(t, "value: .nan\n", yamlval.ErrProhibitedNumber)
}

// ─────────────────────────────────────────────────────────────────────────────
// Exhaustive collection — multiple violations reported in one pass
// ─────────────────────────────────────────────────────────────────────────────

func TestExhaustive_MultipleViolations(t *testing.T) {
	// A document with: anchor + YAML 1.1 boolean + duplicate key.
	// All three must be present in the result.
	input := "a: &anchor 1\nflag: yes\na: duplicate\n"
	result := validate(t, input)
	if result.Valid {
		t.Fatal("expected invalid document")
	}
	codes := make(map[yamlval.ErrorCode]bool)
	for _, e := range result.Errors {
		codes[e.Code] = true
	}
	if !codes[yamlval.ErrAnchor] {
		t.Error("expected ErrAnchor in violations")
	}
	if !codes[yamlval.ErrYAML11Boolean] {
		t.Error("expected ErrYAML11Boolean in violations")
	}
	if !codes[yamlval.ErrDuplicateKey] {
		t.Error("expected ErrDuplicateKey in violations")
	}
}

func TestExhaustive_ValidResultHasEmptyErrors(t *testing.T) {
	result := validate(t, "key: value\n")
	if !result.Valid {
		t.Fatalf("expected valid; got:\n%s", formatErrors(result))
	}
	if len(result.Errors) != 0 {
		t.Fatalf("Valid result must have zero errors; got %d", len(result.Errors))
	}
}
