package yaml

import (
	"bytes"
	"errors"
	"io"

	goyaml "gopkg.in/yaml.v3"
)

// Validate checks that input conforms to the traCtl YAML subset defined in
// tractl_yaml_spec.md §6–§22.
//
// # Pipeline
//
// Validation runs in three sequential phases:
//
//  1. Lexical pre-checks (raw bytes): encoding, BOM, tab indentation,
//     YAML directives, and the document-end marker (...). These checks
//     operate on the raw byte sequence and produce accurate line numbers.
//     Encoding failures are fatal; all others are collected and parsing
//     continues so that all violations can be reported together.
//
//  2. Structural parse: gopkg.in/yaml.v3 decodes the input into a yaml.Node
//     AST. Parse failures are wrapped in ErrParseFailed and returned
//     immediately — a broken AST cannot be reliably walked.
//
//  3. AST walk: the node tree is traversed to enforce all remaining subset
//     restrictions: anchors, aliases, explicit tags, single-quoted strings,
//     non-empty flow style, merge keys, complex keys, duplicate keys, reserved
//     key prefixes, and prohibited scalar values (YAML 1.1 booleans, ambiguous
//     nulls, implicit timestamps, prohibited numeric forms).
//
// # Scope
//
// Validate enforces only syntactic and subset compliance as defined in
// tractl_yaml_spec.md §20. Canonical schema validation, overlay validation,
// and planner-level checks are explicitly out of scope.
//
// # Exhaustiveness
//
// Validate collects all violations before returning. It does not short-circuit
// on the first error, so tooling can present complete diagnostics in a single
// pass.
//
// Validate is the sole public entry point for this package.
func Validate(input []byte) *ValidationResult {
	result := newResult()

	// ── Phase 1: lexical pre-checks ───────────────────────────────────────

	stripped, encErrs := checkEncoding(input)
	for _, e := range encErrs {
		result.Add(e)
	}
	// Encoding failures make further byte-level and parse analysis unreliable.
	if len(encErrs) > 0 {
		return result
	}

	// Normalize line endings before all subsequent scanning and parsing.
	// This ensures line-number counts are consistent across CRLF and LF inputs.
	normalized := normalizeLineEndings(stripped)

	for _, e := range checkTabIndentation(normalized) {
		result.Add(e)
	}
	for _, e := range checkDirectives(normalized) {
		result.Add(e)
	}
	for _, e := range checkDocumentEndMarker(normalized) {
		result.Add(e)
	}

	// ── Phase 2: structural parse ─────────────────────────────────────────

	// goyaml.NewDecoder accepts an io.Reader. bytes.NewReader wraps our byte
	// slice in a reader without copying it — equivalent to creating a
	// Readable stream from a Buffer in Node.js.
	dec := goyaml.NewDecoder(bytes.NewReader(normalized))

	var doc goyaml.Node
	if err := dec.Decode(&doc); err != nil {
		result.Add(ValidationError{
			Code:    ErrParseFailed,
			Message: "YAML parse error: " + err.Error(),
		})
		return result
	}

	// Multi-document stream detection.
	// We attempt to decode a second document from the same stream.
	// io.EOF means the stream ended normally after the first document — that's
	// correct. Any other outcome means additional content was present.
	//
	// Go concept — errors.Is: this is the idiomatic way to check for a
	// specific sentinel error value, even when the error is wrapped.
	// Equivalent to err === io.EOF in a less pedantic world.
	var second goyaml.Node
	if secondErr := dec.Decode(&second); secondErr == nil || !errors.Is(secondErr, io.EOF) {
		msg := "multi-document YAML streams are not permitted; a traCtl YAML file must contain exactly one document"
		if secondErr != nil && !errors.Is(secondErr, io.EOF) {
			msg += " (second document parse error: " + secondErr.Error() + ")"
		}
		result.Add(ValidationError{
			Code:    ErrMultiDocument,
			Message: msg,
		})
	}

	// ── Phase 3: AST walk ─────────────────────────────────────────────────

	for _, e := range walkDocument(&doc) {
		result.Add(e)
	}

	return result
}
