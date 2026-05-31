package toon_test

import (
	"strings"
	"testing"

	toonval "github.com/tractl/tractl/internal/validation/toon"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

func validate(t *testing.T, input string) *toonval.ValidationResult {
	t.Helper()
	return toonval.Validate([]byte(input))
}

func assertValid(t *testing.T, input string) {
	t.Helper()
	result := validate(t, input)
	if !result.Valid {
		t.Fatalf("expected valid document but got %d error(s):\n%s",
			len(result.Errors), formatErrors(result))
	}
}

func assertInvalid(t *testing.T, input string, wantCode toonval.ErrorCode) {
	t.Helper()
	result := validate(t, input)
	if result.Valid {
		t.Fatalf("expected validation failure with code %s but document was accepted", wantCode)
	}
	for _, e := range result.Errors {
		if e.Code == wantCode {
			return
		}
	}
	t.Fatalf("expected error code %s not found; got:\n%s", wantCode, formatErrors(result))
}

func formatErrors(r *toonval.ValidationResult) string {
	var sb strings.Builder
	for _, e := range r.Errors {
		sb.WriteString("  • " + e.Error() + "\n")
	}
	return sb.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// Valid documents — spec canonical examples must pass
// ─────────────────────────────────────────────────────────────────────────────

func TestValid_MinimalWorkflow(t *testing.T) {
	// spec §17.1
	assertValid(t, `schemaVersion: 1
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
	// spec §17.2
	assertValid(t, `schemaVersion: 1
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
        extracts:
          - id: token
            source: body
            path: $.access_token
`)
}

func TestValid_OverlayDocument(t *testing.T) {
	// spec §17.3
	assertValid(t, `metadata:
  name: prod-overlay

patches:
  - target:
      path: workflows.user-flow.steps.login
    patch:
      timeout: "PT30S"
`)
}

func TestValid_EmptyContainers(t *testing.T) {
	// spec §8.3: {} and [] are the only permitted inline forms.
	assertValid(t, "capabilities: []\nvariables: {}\n")
}

func TestValid_BOMIsStripped(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	doc := []byte("key: value\n")
	result := toonval.Validate(append(bom, doc...))
	if !result.Valid {
		t.Fatalf("expected BOM-prefixed document to be accepted; got:\n%s", formatErrors(result))
	}
}

func TestValid_CommentsAreDiscarded(t *testing.T) {
	// spec §6.5: comments must be silently discarded.
	assertValid(t, "# this is a comment\nkey: value\n")
}

func TestValid_YesNoBareStringsPermitted(t *testing.T) {
	// TOON spec §7.2: yes/no/on/off are bare strings in TOON, not booleans.
	// Unlike YAML, TOON validators must NOT reject these as boolean variants.
	assertValid(t, "answer: yes\nflag: no\nstate: on\nmode: off\n")
}

func TestValid_CanonicalBooleans(t *testing.T) {
	assertValid(t, "enabled: true\ndisabled: false\n")
}

func TestValid_CanonicalNull(t *testing.T) {
	assertValid(t, "value: null\n")
}

func TestValid_CanonicalNumbers(t *testing.T) {
	assertValid(t, "integer: 42\nnegative: -7\nfloat: 3.14\n")
}

func TestValid_BlockStrings(t *testing.T) {
	// spec §7.6: literal (|) and folded (>) block strings.
	assertValid(t, "description: |\n  line one\n  line two\nnotes: >\n  folded content\n")
}

func TestValid_DoubleQuotedStrings(t *testing.T) {
	assertValid(t, "value: \"hello world\"\nescaped: \"say \\\"hi\\\"\"\n")
}

func TestValid_ValidEscapeSequences(t *testing.T) {
	// All permitted escape sequences per spec §7.1.1.
	assertValid(t, "a: \"tab:\\there\"\nb: \"newline:\\nhere\"\nc: \"cr:\\rhere\"\nd: \"unicode:\\u0041\"\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding (§6.1)
// ─────────────────────────────────────────────────────────────────────────────

func TestEncoding_UTF16BE(t *testing.T) {
	input := []byte{0xFE, 0xFF, 0x00, 'k', 0x00, 'e', 0x00, 'y', 0x00, ':', 0x00, '1'}
	result := toonval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-16 BE input")
	}
	for _, e := range result.Errors {
		if e.Code == toonval.ErrEncoding {
			return
		}
	}
	t.Fatalf("expected ErrEncoding; got:\n%s", formatErrors(result))
}

func TestEncoding_UTF32BE(t *testing.T) {
	input := []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 'k'}
	result := toonval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-32 BE input")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Indentation (§6.3)
// ─────────────────────────────────────────────────────────────────────────────

func TestIndentation_TabInIndentation(t *testing.T) {
	assertInvalid(t, "key:\n\tvalue: 1\n", toonval.ErrTabIndentation)
}

// ─────────────────────────────────────────────────────────────────────────────
// Document separators (§20)
// ─────────────────────────────────────────────────────────────────────────────

func TestDocumentSeparator_DashMarker(t *testing.T) {
	assertInvalid(t, "---\nkey: value\n", toonval.ErrDocumentSeparator)
}

func TestDocumentSeparator_DotMarker(t *testing.T) {
	assertInvalid(t, "key: value\n...\n", toonval.ErrDocumentSeparator)
}

// ─────────────────────────────────────────────────────────────────────────────
// Directives (§20)
// ─────────────────────────────────────────────────────────────────────────────

func TestDirective_YAMLVersion(t *testing.T) {
	assertInvalid(t, "%YAML 1.2\nkey: value\n", toonval.ErrDirective)
}

// ─────────────────────────────────────────────────────────────────────────────
// Anchors and aliases (§20)
// ─────────────────────────────────────────────────────────────────────────────

func TestAnchor_Declaration(t *testing.T) {
	assertInvalid(t, "defaults: &defaults\n  timeout: 30\n", toonval.ErrAnchor)
}

func TestAlias_Reference(t *testing.T) {
	input := "a: &def\n  x: 1\nb: *def\n"
	result := validate(t, input)
	if result.Valid {
		t.Fatal("expected invalid document")
	}
	var gotAnchor, gotAlias bool
	for _, e := range result.Errors {
		if e.Code == toonval.ErrAnchor {
			gotAnchor = true
		}
		if e.Code == toonval.ErrAlias {
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
// Explicit tags (§20)
// ─────────────────────────────────────────────────────────────────────────────

func TestExplicitTag_DoubleExclamation(t *testing.T) {
	assertInvalid(t, "value: !!str hello\n", toonval.ErrExplicitTag)
}

// ─────────────────────────────────────────────────────────────────────────────
// Single-quoted strings (§7.1)
// ─────────────────────────────────────────────────────────────────────────────

func TestSingleQuote_StringValue(t *testing.T) {
	assertInvalid(t, "value: 'single quoted'\n", toonval.ErrSingleQuote)
}

// ─────────────────────────────────────────────────────────────────────────────
// Flow style (§8.3)
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowStyle_NonEmptyMapping(t *testing.T) {
	assertInvalid(t, "value: {key: val}\n", toonval.ErrFlowStyle)
}

func TestFlowStyle_NonEmptySequence(t *testing.T) {
	assertInvalid(t, "values: [a, b, c]\n", toonval.ErrFlowStyle)
}

func TestFlowStyle_EmptyMappingPermitted(t *testing.T) {
	assertValid(t, "value: {}\n")
}

func TestFlowStyle_EmptySequencePermitted(t *testing.T) {
	assertValid(t, "values: []\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Mapping constraints (§8.1, §9.1, §9.2, §20)
// ─────────────────────────────────────────────────────────────────────────────

func TestMergeKey_Rejected(t *testing.T) {
	assertInvalid(t, "defaults: &def\n  x: 1\nother:\n  <<: *def\n", toonval.ErrMergeKey)
}

func TestDuplicateKey_SameLevel(t *testing.T) {
	assertInvalid(t, "key: value1\nkey: value2\n", toonval.ErrDuplicateKey)
}

func TestULIDKey_Rejected(t *testing.T) {
	assertInvalid(t, "_ulid: 01ARZ3NDEKTSV4RRFFQ69G5FAV\n", toonval.ErrULIDKey)
}

func TestReservedPrefix_Underscore(t *testing.T) {
	assertInvalid(t, "_traCtl.internal: value\n", toonval.ErrReservedPrefix)
}

func TestReservedPrefix_Plain(t *testing.T) {
	assertInvalid(t, "traCtl.config: value\n", toonval.ErrReservedPrefix)
}

func TestReservedPrefix_NotAPrefix(t *testing.T) {
	assertValid(t, "traCtlSpec: 1\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Invalid escape sequences (§7.1.1)
// ─────────────────────────────────────────────────────────────────────────────

func TestInvalidEscape_BackslashA(t *testing.T) {
	// \a is not in the TOON-permitted escape set.
	assertInvalid(t, "value: \"hello\\aworld\"\n", toonval.ErrInvalidEscape)
}

func TestInvalidEscape_BackslashV(t *testing.T) {
	assertInvalid(t, "value: \"\\v\"\n", toonval.ErrInvalidEscape)
}

func TestInvalidEscape_BackslashZero(t *testing.T) {
	assertInvalid(t, "value: \"\\0\"\n", toonval.ErrInvalidEscape)
}

// ─────────────────────────────────────────────────────────────────────────────
// Exhaustive collection — multiple violations in one pass
// ─────────────────────────────────────────────────────────────────────────────

func TestExhaustive_MultipleViolations(t *testing.T) {
	// anchor + duplicate key + single-quoted string
	input := "a: &anchor 1\nflag: 'quoted'\na: duplicate\n"
	result := validate(t, input)
	if result.Valid {
		t.Fatal("expected invalid document")
	}
	codes := make(map[toonval.ErrorCode]bool)
	for _, e := range result.Errors {
		codes[e.Code] = true
	}
	if !codes[toonval.ErrAnchor] {
		t.Error("expected ErrAnchor in violations")
	}
	if !codes[toonval.ErrSingleQuote] {
		t.Error("expected ErrSingleQuote in violations")
	}
	if !codes[toonval.ErrDuplicateKey] {
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
