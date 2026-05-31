// Package toon implements the traCtl TOON subset validator defined in
// tractl_toon_spec.md §6–§22.
//
// The sole public entry point is [Validate]. All other symbols are internal
// to the package.
//
// Validation scope (per tractl_toon_spec.md §19):
//   - Syntactic validity: encoding, indentation, lexical correctness,
//     structural well-formedness, subset compliance.
//
// Explicitly out of scope for this package:
//   - Canonical schema validation (tractl_spec.md)
//   - Overlay schema / selector validation (tractl_overlay_spec.md)
//   - Planner-level validation (capability resolution, DAG cycles)
package toon

import "github.com/tractl/tractl/internal/validation/common"

// ErrorCode is a re-export of common.ErrorCode for callers of this package.
type ErrorCode = common.ErrorCode

// ValidationError is a re-export of common.ValidationError for callers of this package.
type ValidationError = common.ValidationError

// ValidationResult is a re-export of common.ValidationResult for callers of this package.
type ValidationResult = common.ValidationResult

const (
	// ErrEncoding is produced when the input uses a prohibited encoding
	// (UTF-16, UTF-32) or contains invalid UTF-8 byte sequences.
	// Spec ref: §6.1.
	ErrEncoding ErrorCode = "TOON_ENCODING"

	// ErrTabIndentation is produced when a tab character appears at an
	// indentation position. TOON requires 2-space indentation.
	// Spec ref: §6.3.
	ErrTabIndentation ErrorCode = "TOON_TAB_INDENTATION"

	// ErrDocumentSeparator is produced when a YAML document-start (---)
	// or document-end (...) marker appears. TOON is single-document.
	// Spec ref: §20.
	ErrDocumentSeparator ErrorCode = "TOON_DOCUMENT_SEPARATOR"

	// ErrDirective is produced when a YAML directive (%YAML, %TAG) appears.
	// YAML directives are not applicable to TOON.
	// Spec ref: §20.
	ErrDirective ErrorCode = "TOON_DIRECTIVE"

	// ErrAnchor is produced when an anchor declaration (&name) is present.
	// TOON has no hidden-inheritance mechanism.
	// Spec ref: §20.
	ErrAnchor ErrorCode = "TOON_ANCHOR"

	// ErrAlias is produced when an alias reference (*name) is present.
	// Spec ref: §20.
	ErrAlias ErrorCode = "TOON_ALIAS"

	// ErrExplicitTag is produced when an explicit YAML tag (!, !!, !<tag>)
	// appears on any node. TOON does not support YAML tags.
	// Spec ref: §20.
	ErrExplicitTag ErrorCode = "TOON_EXPLICIT_TAG"

	// ErrSingleQuote is produced when a single-quoted string literal is
	// present. TOON does not support single-quoted strings.
	// Spec ref: §7.1.
	ErrSingleQuote ErrorCode = "TOON_SINGLE_QUOTE"

	// ErrFlowStyle is produced when a non-empty flow-style mapping or sequence
	// is present. Only {} and [] (empty) are permitted.
	// Spec ref: §8.3.
	ErrFlowStyle ErrorCode = "TOON_FLOW_STYLE"

	// ErrMergeKey is produced when a YAML merge key (<<:) is present.
	// TOON has no hidden-inheritance mechanism.
	// Spec ref: §20.
	ErrMergeKey ErrorCode = "TOON_MERGE_KEY"

	// ErrComplexKey is produced when a mapping key is not a plain scalar.
	// Spec ref: §8.1.
	ErrComplexKey ErrorCode = "TOON_COMPLEX_KEY"

	// ErrDuplicateKey is produced when a mapping contains two or more entries
	// with the same key. Spec ref: §8.1.
	ErrDuplicateKey ErrorCode = "TOON_DUPLICATE_KEY"

	// ErrULIDKey is produced when a mapping key named "_ulid" appears in an
	// authored document. _ulid is system-assigned.
	// Spec ref: §9.2.
	ErrULIDKey ErrorCode = "TOON_ULID_KEY"

	// ErrReservedPrefix is produced when a mapping key uses a reserved
	// namespace prefix (_traCtl. or traCtl.).
	// Spec ref: §9.1.
	ErrReservedPrefix ErrorCode = "TOON_RESERVED_PREFIX"

	// ErrInvalidEscape is produced when a double-quoted string contains an
	// unrecognized escape sequence (not in the set \", \\, \n, \r, \t, \uXXXX).
	// Spec ref: §7.1.1.
	ErrInvalidEscape ErrorCode = "TOON_INVALID_ESCAPE"

	// ErrParseFailed is produced when the underlying parser cannot build a
	// valid AST from the input.
	ErrParseFailed ErrorCode = "TOON_PARSE_FAILED"
)

func newResult() *ValidationResult {
	return common.NewResult()
}
