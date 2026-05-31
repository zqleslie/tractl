// Package output provides output renderers for CLI result formatting.
package output

// TODO: output renderers — concise, JSON, YAML, TOON
// Reference: HLD §4.9, ADR-002 §4, CLAUDE.md conventions
// Output contract modes:
//   - concise human-readable (CLI default)
//   - structured JSON  (--output json)
//   - structured YAML  (--output yaml)
//   - structured TOON  (--output toon)
//   - expanded diagnostic analysis
// Constraint: exit codes are deterministic (CLAUDE.md)
