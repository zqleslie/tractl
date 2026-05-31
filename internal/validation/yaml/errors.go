// Package yaml implements the traCtl YAML subset validator defined in
// tractl_yaml_spec.md §6–§22.
//
// The sole public entry point is [Validate]. All other symbols are internal
// to the package.
//
// Validation scope (per tractl_yaml_spec.md §20):
//   - Syntactic validity: encoding, indentation, lexical correctness,
//     structural well-formedness, subset compliance.
//
// Explicitly out of scope for this package:
//   - Canonical schema validation (tractl_spec.md)
//   - Overlay schema / selector validation (tractl_overlay_spec.md)
//   - Planner-level validation (capability resolution, DAG cycles)
package yaml

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
	// Spec ref: §6.4.
	ErrEncoding ErrorCode = "YAML_ENCODING"

	// ErrTabIndentation is produced when a tab character appears at an
	// indentation position (before the first non-whitespace character on the
	// line). traCtl YAML requires space-only indentation.
	// Spec ref: §6.6.
	ErrTabIndentation ErrorCode = "YAML_TAB_INDENTATION"

	// ErrDirective is produced when a YAML directive (%YAML, %TAG, or any
	// %-prefixed preamble line) appears in the document.
	// Spec ref: §6.3.
	ErrDirective ErrorCode = "YAML_DIRECTIVE"

	// ErrMultiDocument is produced when the input stream contains more than
	// one YAML document separated by ---.
	// Spec ref: §6.2.
	ErrMultiDocument ErrorCode = "YAML_MULTI_DOCUMENT"

	// ErrDocumentEndMarker is produced when the YAML document-end marker (...)
	// appears in the input. The document-start marker (---) is permitted.
	// Spec ref: §6.2.
	ErrDocumentEndMarker ErrorCode = "YAML_DOCUMENT_END_MARKER"

	// ErrAnchor is produced when an anchor declaration (&name) is present.
	// Spec ref: §7.1.
	ErrAnchor ErrorCode = "YAML_ANCHOR"

	// ErrAlias is produced when an alias reference (*name) is present.
	// Spec ref: §7.1.
	ErrAlias ErrorCode = "YAML_ALIAS"

	// ErrExplicitTag is produced when an explicit YAML tag (!, !!, !<tag>)
	// appears on any node.
	// Spec ref: §7.3.
	ErrExplicitTag ErrorCode = "YAML_EXPLICIT_TAG"

	// ErrSingleQuote is produced when a single-quoted string literal is
	// present. Only plain and double-quoted scalars are permitted.
	// Spec ref: §7.12.
	ErrSingleQuote ErrorCode = "YAML_SINGLE_QUOTE"

	// ErrFlowStyle is produced when a non-empty flow-style mapping or sequence
	// is present. The empty-container forms {} and [] are permitted.
	// Spec ref: §7.11.
	ErrFlowStyle ErrorCode = "YAML_FLOW_STYLE"

	// ErrMergeKey is produced when a YAML merge key (<<:) is present.
	// Spec ref: §7.2.
	ErrMergeKey ErrorCode = "YAML_MERGE_KEY"

	// ErrComplexKey is produced when a mapping key is not a plain scalar
	// string. Sequence-as-key and mapping-as-key forms are rejected.
	// Spec ref: §7.9.
	ErrComplexKey ErrorCode = "YAML_COMPLEX_KEY"

	// ErrDuplicateKey is produced when a mapping contains two or more entries
	// with the same string key.
	// Spec ref: §7.10.
	ErrDuplicateKey ErrorCode = "YAML_DUPLICATE_KEY"

	// ErrYAML11Boolean is produced when a YAML 1.1 boolean variant
	// (yes/no/on/off/Y/N/True/False and their case variants) is used.
	// Only "true" and "false" (all-lowercase) are valid.
	// Spec ref: §7.4.
	ErrYAML11Boolean ErrorCode = "YAML_11_BOOLEAN"

	// ErrProhibitedNull is produced when a prohibited null form appears:
	//   ~ is rejected as ambiguous.
	//   Null and NULL must be quoted if a string is intended.
	//   An empty value position (key:) is rejected.
	// Spec ref: §7.5.
	ErrProhibitedNull ErrorCode = "YAML_PROHIBITED_NULL"

	// ErrImplicitTimestamp is produced when a scalar is implicitly resolved
	// as a YAML timestamp (e.g. 2026-05-24T12:00:00Z).
	// Date/time values must be double-quoted.
	// Spec ref: §7.7.
	ErrImplicitTimestamp ErrorCode = "YAML_IMPLICIT_TIMESTAMP"

	// ErrProhibitedNumber is produced when a numeric literal uses a form
	// outside the JSON number grammar: octal (0o17, 017), hexadecimal (0xFF),
	// binary (0b101), sexagesimal (12:34:56), underscored (1_000_000),
	// positive-sign prefix (+42), or special floats (.inf, .nan).
	// Spec ref: §7.6.
	ErrProhibitedNumber ErrorCode = "YAML_PROHIBITED_NUMBER"

	// ErrULIDKey is produced when a mapping key named "_ulid" appears in an
	// authored document. _ulid is system-assigned and must never be authored.
	// Spec ref: §10.2.
	ErrULIDKey ErrorCode = "YAML_ULID_KEY"

	// ErrReservedPrefix is produced when a mapping key uses a reserved
	// namespace prefix (_traCtl. or traCtl.).
	// Spec ref: §10.1.
	ErrReservedPrefix ErrorCode = "YAML_RESERVED_PREFIX"

	// ErrParseFailed is produced when the underlying YAML parser cannot build
	// a valid AST from the input. This wraps the library's diagnostic.
	ErrParseFailed ErrorCode = "YAML_PARSE_FAILED"
)

// newResult returns a result in the valid (zero-error) state.
func newResult() *ValidationResult {
	return common.NewResult()
}
