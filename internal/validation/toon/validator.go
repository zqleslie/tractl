// Package toon implements the traCtl TOON subset validator defined in
// tractl_toon_spec.md §6–§22.
//
// The sole public entry point is [Validate]. All other symbols are internal
// to the package.
package toon

import (
	"bytes"
	"errors"
	"io"

	goyaml "gopkg.in/yaml.v3"
)

// Validate checks that input conforms to the traCtl TOON subset defined in
// tractl_toon_spec.md §6–§22.
//
// # Pipeline
//
//  1. Encoding pre-check (raw bytes): UTF-8 enforcement, BOM detection/strip.
//     Encoding failures are fatal.
//
//  2. Lexical pre-checks (raw bytes): tab indentation, YAML directives,
//     document separator markers (--- / ...), invalid escape sequences in
//     double-quoted strings.
//
//  3. Structural parse: gopkg.in/yaml.v3 decodes the input into a yaml.Node
//     AST. Parse failures are wrapped in ErrParseFailed.
//
//  4. AST walk: the node tree is traversed to enforce all remaining subset
//     restrictions: anchors, aliases, explicit tags, single-quoted strings,
//     non-empty flow style, merge keys, complex keys, duplicate keys,
//     reserved key prefixes, _ulid key prohibition.
//
// # Scope
//
// Validate enforces only syntactic and subset compliance per
// tractl_toon_spec.md §19. Canonical schema validation, overlay validation,
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

	normalized := normalizeLineEndings(stripped)

	// ── Phase 2: lexical pre-checks ──────────────────────────────────────────

	for _, e := range checkTabIndentation(normalized) {
		result.Add(e)
	}
	for _, e := range checkDirectives(normalized) {
		result.Add(e)
	}
	for _, e := range checkDocumentSeparators(normalized) {
		result.Add(e)
	}
	for _, e := range checkInvalidEscapes(normalized) {
		result.Add(e)
	}

	// ── Phase 3: structural parse ─────────────────────────────────────────────

	dec := goyaml.NewDecoder(bytes.NewReader(normalized))

	var doc goyaml.Node
	if err := dec.Decode(&doc); err != nil {
		result.Add(ValidationError{
			Code:    ErrParseFailed,
			Message: "TOON parse error: " + err.Error(),
		})
		return result
	}

	// Multi-document stream detection.
	var second goyaml.Node
	if secondErr := dec.Decode(&second); secondErr == nil || !errors.Is(secondErr, io.EOF) {
		msg := "TOON documents must contain exactly one document"
		if secondErr != nil && !errors.Is(secondErr, io.EOF) {
			msg += " (second document parse error: " + secondErr.Error() + ")"
		}
		result.Add(ValidationError{
			Code:    ErrDocumentSeparator,
			Message: msg,
		})
	}

	// ── Phase 4: AST walk ────────────────────────────────────────────────────

	for _, e := range walkDocument(&doc) {
		result.Add(e)
	}

	return result
}
