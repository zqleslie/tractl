// validator_test.go tests the JSON format validator against the traCtl JSON
// subset rules. Covers: valid canonical examples, encoding violations, JSON5
// comments, trailing commas, duplicate keys, and reserved key prefixes.
package json_test

import (
	"strings"
	"testing"

	jsonval "github.com/tractl/tractl/internal/validation/json"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

func validate(t *testing.T, input string) *jsonval.ValidationResult {
	t.Helper()
	return jsonval.Validate([]byte(input))
}

func assertValid(t *testing.T, input string) {
	t.Helper()
	result := validate(t, input)
	if !result.Valid {
		t.Fatalf("expected valid document but got %d error(s):\n%s",
			len(result.Errors), formatErrors(result))
	}
}

func assertInvalid(t *testing.T, input string, wantCode jsonval.ErrorCode) {
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

func formatErrors(r *jsonval.ValidationResult) string {
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
	// spec §19.1
	assertValid(t, `{
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
}`)
}

func TestValid_WorkflowWithExpressions(t *testing.T) {
	// spec §19.3
	assertValid(t, `{
  "schemaVersion": 1,
  "capabilities": ["protocol.http", "scripting.js", "secret.ext.vault@1"],
  "variables": { "baseUrl": "https://api.example.com" },
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
            "operation": "POST"
          },
          "extracts": [{ "id": "token", "source": "body", "path": "$.access_token" }]
        }
      ]
    }
  ]
}`)
}

func TestValid_OverlayDocument(t *testing.T) {
	// spec §19.4
	assertValid(t, `{
  "metadata": { "name": "prod-overlay" },
  "patches": [
    {
      "target": { "path": "workflows.user-flow.steps.login" },
      "patch": { "timeout": "PT30S" }
    }
  ]
}`)
}

func TestValid_EmptyContainers(t *testing.T) {
	assertValid(t, `{"capabilities": [], "variables": {}, "workflows": []}`)
}

func TestValid_BOMIsStripped(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	doc := []byte(`{"key": "value"}`)
	result := jsonval.Validate(append(bom, doc...))
	if !result.Valid {
		t.Fatalf("expected BOM-prefixed document to be accepted; got:\n%s", formatErrors(result))
	}
}

func TestValid_CanonicalNumbers(t *testing.T) {
	assertValid(t, `{"integer": 42, "negative": -7, "zero": 0, "float": 3.14, "exp": 1e10}`)
}

func TestValid_NullAndBooleans(t *testing.T) {
	assertValid(t, `{"active": true, "disabled": false, "value": null}`)
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding (§6.4)
// ─────────────────────────────────────────────────────────────────────────────

func TestEncoding_UTF16BE(t *testing.T) {
	input := []byte{0xFE, 0xFF, 0x00, '{', 0x00, '}'}
	result := jsonval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-16 BE input")
	}
	for _, e := range result.Errors {
		if e.Code == jsonval.ErrEncoding {
			return
		}
	}
	t.Fatalf("expected ErrEncoding; got:\n%s", formatErrors(result))
}

func TestEncoding_UTF32BE(t *testing.T) {
	input := []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, '{'}
	result := jsonval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for UTF-32 BE input")
	}
}

func TestEncoding_InvalidUTF8(t *testing.T) {
	// 0xFF is never valid UTF-8.
	input := []byte{'{', '"', 'k', '"', ':', '"', 0xFF, '"', '}'}
	result := jsonval.Validate(input)
	if result.Valid {
		t.Fatal("expected ErrEncoding for invalid UTF-8 input")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Document structure (§6.3)
// ─────────────────────────────────────────────────────────────────────────────

func TestStructure_TopLevelArray(t *testing.T) {
	assertInvalid(t, `[{"schemaVersion": 1}]`, jsonval.ErrNotObject)
}

func TestStructure_TopLevelString(t *testing.T) {
	assertInvalid(t, `"just a string"`, jsonval.ErrNotObject)
}

func TestStructure_TopLevelNumber(t *testing.T) {
	assertInvalid(t, `42`, jsonval.ErrNotObject)
}

func TestStructure_MultiValue(t *testing.T) {
	assertInvalid(t, `{"a": 1}{"b": 2}`, jsonval.ErrMultiValue)
}

func TestStructure_EmptyInput(t *testing.T) {
	result := validate(t, "")
	if result.Valid {
		t.Fatal("expected invalid for empty input")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Forbidden JSON5/JSONC constructs (§6.2)
// ─────────────────────────────────────────────────────────────────────────────

func TestComment_LineComment(t *testing.T) {
	assertInvalid(t, "// comment\n{\"key\": \"value\"}", jsonval.ErrComment)
}

func TestComment_InlineLineComment(t *testing.T) {
	assertInvalid(t, "{\"key\": \"value\"} // comment", jsonval.ErrComment)
}

func TestComment_BlockComment(t *testing.T) {
	assertInvalid(t, "/* comment */\n{\"key\": \"value\"}", jsonval.ErrComment)
}

func TestSingleQuote_StringValue(t *testing.T) {
	assertInvalid(t, "{'key': 'value'}", jsonval.ErrSingleQuote)
}

func TestTrailingComma_InObject(t *testing.T) {
	assertInvalid(t, `{"key": "value",}`, jsonval.ErrTrailingComma)
}

func TestTrailingComma_InArray(t *testing.T) {
	assertInvalid(t, `{"items": [1, 2, 3,]}`, jsonval.ErrTrailingComma)
}

// ─────────────────────────────────────────────────────────────────────────────
// Duplicate keys (§8.1)
// ─────────────────────────────────────────────────────────────────────────────

func TestDuplicateKey_TopLevel(t *testing.T) {
	assertInvalid(t, `{"key": "value1", "key": "value2"}`, jsonval.ErrDuplicateKey)
}

func TestDuplicateKey_Nested(t *testing.T) {
	assertInvalid(t, `{"outer": {"inner": 1, "inner": 2}}`, jsonval.ErrDuplicateKey)
}

// ─────────────────────────────────────────────────────────────────────────────
// ULID and reserved key prefixes (§9.1, §9.2)
// ─────────────────────────────────────────────────────────────────────────────

func TestULIDKey_Rejected(t *testing.T) {
	assertInvalid(t, `{"_ulid": "01ARZ3NDEKTSV4RRFFQ69G5FAV"}`, jsonval.ErrULIDKey)
}

func TestReservedPrefix_Underscore(t *testing.T) {
	assertInvalid(t, `{"_traCtl.internal": "value"}`, jsonval.ErrReservedPrefix)
}

func TestReservedPrefix_Plain(t *testing.T) {
	assertInvalid(t, `{"traCtl.config": "value"}`, jsonval.ErrReservedPrefix)
}

func TestReservedPrefix_NotAPrefix(t *testing.T) {
	// "traCtlSpec" does not start with "traCtl." — no dot — so it is not reserved.
	assertValid(t, `{"traCtlSpec": 1}`)
}

// ─────────────────────────────────────────────────────────────────────────────
// Prohibited number forms (§7.2)
// ─────────────────────────────────────────────────────────────────────────────

func TestNumber_PositiveSign(t *testing.T) {
	// encoding/json rejects +42 as invalid JSON; we still get ErrParseFailed or
	// ErrProhibitedNumber depending on which phase catches it first.
	result := validate(t, `{"value": +42}`)
	if result.Valid {
		t.Fatal("expected invalid for +42")
	}
}

func TestNumber_LeadingZero(t *testing.T) {
	result := validate(t, `{"value": 0123}`)
	if result.Valid {
		t.Fatal("expected invalid for leading zero integer")
	}
}

func TestNumber_MissingIntegerPart(t *testing.T) {
	result := validate(t, `{"value": .5}`)
	if result.Valid {
		t.Fatal("expected invalid for .5")
	}
}

func TestNumber_MissingFractionalPart(t *testing.T) {
	result := validate(t, `{"value": 5.}`)
	if result.Valid {
		t.Fatal("expected invalid for 5.")
	}
}

func TestNumber_Infinity(t *testing.T) {
	result := validate(t, `{"value": Infinity}`)
	if result.Valid {
		t.Fatal("expected invalid for Infinity")
	}
}

func TestNumber_NaN(t *testing.T) {
	result := validate(t, `{"value": NaN}`)
	if result.Valid {
		t.Fatal("expected invalid for NaN")
	}
}

func TestNumber_ValidExponent(t *testing.T) {
	assertValid(t, `{"value": 1e10, "neg": -2.5e-3}`)
}

func TestNumber_ValidZero(t *testing.T) {
	assertValid(t, `{"value": 0, "neg": -0}`)
}

// ─────────────────────────────────────────────────────────────────────────────
// Exhaustive collection — multiple violations reported in one pass
// ─────────────────────────────────────────────────────────────────────────────

func TestExhaustive_MultipleViolations(t *testing.T) {
	// Document with: comment + trailing comma + duplicate key.
	input := "// comment\n{\"key\": 1, \"key\": 2,}"
	result := validate(t, input)
	if result.Valid {
		t.Fatal("expected invalid document")
	}
	codes := make(map[jsonval.ErrorCode]bool)
	for _, e := range result.Errors {
		codes[e.Code] = true
	}
	if !codes[jsonval.ErrComment] {
		t.Error("expected ErrComment in violations")
	}
	if !codes[jsonval.ErrTrailingComma] {
		t.Error("expected ErrTrailingComma in violations")
	}
}

func TestExhaustive_ValidResultHasEmptyErrors(t *testing.T) {
	result := validate(t, `{"key": "value"}`)
	if !result.Valid {
		t.Fatalf("expected valid; got:\n%s", formatErrors(result))
	}
	if len(result.Errors) != 0 {
		t.Fatalf("Valid result must have zero errors; got %d", len(result.Errors))
	}
}
