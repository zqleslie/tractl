// Package common provides shared validation primitives used by the traCtl
// format validators (YAML, JSON, TOON).
//
// Every format validator imports this package for:
//   - [ValidationError] and [ValidationResult] — the uniform error surface
//   - [ErrorCode] — stable string identifiers for violations
//   - Shared rule helpers: encoding checks, reserved-key enforcement
//
// Package isolation invariant: this package MUST NOT import internal/spec,
// internal/overlay, or any format-specific validator package.
package common

import "fmt"

// ErrorCode is a stable string identifier for a validation violation.
// Callers SHOULD match on ErrorCode rather than parsing message text.
// Codes are stable across minor releases.
type ErrorCode string

// ValidationError records a single violation found during validation.
// It implements the standard error interface.
type ValidationError struct {
	Code    ErrorCode
	Message string
	// Line is the 1-based line number in the source document.
	// Zero means position is not applicable or not determinable.
	Line int
	// Column is the 1-based column number in the source document.
	// Zero means position is not applicable or not determinable.
	Column int
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] line %d, col %d: %s", e.Code, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ValidationResult is the outcome of a Validate call.
// When Valid is true, Errors is guaranteed to be empty.
// When Valid is false, Errors contains all violations found; validation is
// exhaustive and does not short-circuit on the first error.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// Add appends a violation and marks the result invalid.
func (r *ValidationResult) Add(e ValidationError) {
	r.Errors = append(r.Errors, e)
	r.Valid = false
}

// NewResult returns a result in the valid (zero-error) state.
func NewResult() *ValidationResult {
	return &ValidationResult{Valid: true}
}
