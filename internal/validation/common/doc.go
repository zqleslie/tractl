// Package common defines shared types and primitives used by all
// format-specific validator sub-packages (yaml, json, toon).
//
// Pipeline role: provides ValidationError, ValidationResult, and ErrorCode —
// the uniform error surface that every format validator returns and every
// caller of a format validator consumes.
//
// Isolation invariant: this package must not import internal/spec,
// internal/overlay, or any format-specific validator package. The format
// packages import common; the reverse import would create a cycle.
package common
