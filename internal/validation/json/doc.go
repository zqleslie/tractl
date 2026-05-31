// Package json implements the traCtl JSON subset validator defined in
// tractl_json_spec.md §6–§22.
//
// Pipeline role: raw JSON bytes → *ValidationResult
//
// Validation runs in four sequential phases: UTF-8 encoding pre-check,
// forbidden-syntax scan (JSON5/JSONC constructs — comments, single-quoted
// strings, trailing commas), structural parse via encoding/json, and an AST
// walk enforcing top-level object shape, single-document constraint,
// duplicate-key prohibition, and reserved key prefixes. The sole public
// entry point is Validate([]byte).
package json
