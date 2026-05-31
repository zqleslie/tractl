// Package json implements the traCtl JSON subset validator defined in
// tractl_json_spec.md §6–§22.
//
// The sole public entry point is [Validate]. All other symbols are internal
// to the package.
package json

import (
	ejson "encoding/json"
	"io"
	"strings"
)

// Validate checks that input conforms to the traCtl JSON subset defined in
// tractl_json_spec.md §6–§22.
//
// # Pipeline
//
//  1. Encoding pre-check (raw bytes): UTF-8 enforcement, BOM detection/strip.
//     Encoding failures are fatal — a broken encoding cannot be parsed further.
//
//  2. Forbidden-syntax scan (raw bytes): detects JSON5/JSONC constructs that
//     Go's standard parser accepts silently or does not reach: // and /* */
//     comments, single-quoted strings, trailing commas.
//
//  3. Structural parse: encoding/json decodes the input to validate RFC 8259
//     correctness. Parse failures are wrapped in ErrParseFailed.
//
//  4. AST walk: the decoded token stream is traversed to enforce remaining
//     subset restrictions: top-level object shape, single document, duplicate
//     keys, _ulid key prohibition, reserved key prefixes, prohibited number
//     forms.
//
// # Scope
//
// Validate enforces only syntactic and subset compliance per
// tractl_json_spec.md §21. Canonical schema validation, overlay validation,
// and planner-level checks are explicitly out of scope.
//
// # Exhaustiveness
//
// Validate collects all violations before returning. It does not
// short-circuit on the first error.
func Validate(input []byte) *ValidationResult {
	result := newResult()

	// ── Phase 1: encoding pre-check ──────────────────────────────────────────

	stripped, encErrs := checkEncoding(input)
	for _, e := range encErrs {
		result.Add(e)
	}
	if len(encErrs) > 0 {
		return result
	}

	// ── Phase 2: forbidden-syntax scan ───────────────────────────────────────

	for _, e := range checkForbiddenSyntax(stripped) {
		result.Add(e)
	}
	for _, e := range checkTrailingCommas(stripped) {
		result.Add(e)
	}

	// ── Phase 3: structural parse ─────────────────────────────────────────────

	// Use a validating parse to confirm RFC 8259 compliance before the walk.
	// encoding/json rejects malformed JSON, duplicate keys in some cases, and
	// all non-RFC-8259 constructs it recognises (e.g. Infinity, NaN).
	dec := ejson.NewDecoder(strings.NewReader(string(stripped)))
	dec.DisallowUnknownFields() // no-op for json.Decoder but explicit intent

	var raw ejson.RawMessage
	if err := dec.Decode(&raw); err != nil {
		result.Add(ValidationError{
			Code:    ErrParseFailed,
			Message: "JSON parse error: " + err.Error(),
		})
		return result
	}

	// Multi-value stream check.
	var dummy ejson.RawMessage
	if err := dec.Decode(&dummy); err == nil || err != io.EOF {
		msg := "multi-value JSON streams (NDJSON, concatenated values) are not permitted; a traCtl JSON document must contain exactly one top-level object"
		result.Add(ValidationError{
			Code:    ErrMultiValue,
			Message: msg,
		})
	}

	// ── Phase 4: AST walk ────────────────────────────────────────────────────

	for _, e := range walkDocument(stripped) {
		result.Add(e)
	}

	return result
}
