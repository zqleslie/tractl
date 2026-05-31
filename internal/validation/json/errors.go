// Package json implements the traCtl JSON subset validator defined in
// tractl_json_spec.md §6–§22.
//
// The sole public entry point is [Validate]. All other symbols are internal
// to the package.
//
// Validation scope (per tractl_json_spec.md §21):
//   - Syntactic validity: encoding, RFC 8259 compliance, subset compliance.
//
// Explicitly out of scope for this package:
//   - Canonical schema validation (tractl_spec.md)
//   - Overlay schema / selector validation (tractl_overlay_spec.md)
//   - Planner-level validation (capability resolution, DAG cycles)
//   - Canonical JSON projection (key ordering, _ulid emission)
package json

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
	ErrEncoding ErrorCode = "JSON_ENCODING"

	// ErrParseFailed is produced when the input is not valid RFC 8259 JSON.
	// Spec ref: §6.1.
	ErrParseFailed ErrorCode = "JSON_PARSE_FAILED"

	// ErrNotObject is produced when the top-level JSON value is not an object.
	// Spec ref: §6.3.
	ErrNotObject ErrorCode = "JSON_NOT_OBJECT"

	// ErrMultiValue is produced when the input contains more than one top-level
	// JSON value (NDJSON / concatenated JSON).
	// Spec ref: §6.3.
	ErrMultiValue ErrorCode = "JSON_MULTI_VALUE"

	// ErrComment is produced when a comment (// or /* */) is detected.
	// Comments are a JSON5/JSONC feature and are not permitted.
	// Spec ref: §6.2.
	ErrComment ErrorCode = "JSON_COMMENT"

	// ErrTrailingComma is produced when a trailing comma appears in an array
	// or object. This is a JSON5 feature and is not permitted.
	// Spec ref: §6.2.
	ErrTrailingComma ErrorCode = "JSON_TRAILING_COMMA"

	// ErrSingleQuote is produced when a single-quoted string literal is
	// detected. Only double-quoted strings are permitted per RFC 8259.
	// Spec ref: §6.2.
	ErrSingleQuote ErrorCode = "JSON_SINGLE_QUOTE"

	// ErrDuplicateKey is produced when an object contains two or more members
	// with the same key string, per RFC 7493 (I-JSON).
	// Spec ref: §8.1.
	ErrDuplicateKey ErrorCode = "JSON_DUPLICATE_KEY"

	// ErrULIDKey is produced when a key named "_ulid" appears in an authored
	// document. _ulid is system-assigned and must not be authored.
	// Spec ref: §9.2.
	ErrULIDKey ErrorCode = "JSON_ULID_KEY"

	// ErrReservedPrefix is produced when a key uses a reserved namespace
	// prefix (_traCtl. or traCtl.).
	// Spec ref: §9.1.
	ErrReservedPrefix ErrorCode = "JSON_RESERVED_PREFIX"

	// ErrProhibitedNumber is produced when a number literal uses a form
	// outside the RFC 8259 grammar: positive-sign prefix (+42), leading zeros
	// (0123), radix prefixes (0x…/0o…/0b…), underscored (1_000_000),
	// missing integer part (.5), missing fractional part (5.), or special
	// float tokens (Infinity, NaN).
	// Spec ref: §7.2.
	ErrProhibitedNumber ErrorCode = "JSON_PROHIBITED_NUMBER"
)

func newResult() *ValidationResult {
	return common.NewResult()
}
