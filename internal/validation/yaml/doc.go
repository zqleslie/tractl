// Package yaml implements the traCtl YAML subset validator defined in
// tractl_yaml_spec.md §6–§22.
//
// Pipeline role: raw YAML bytes → *ValidationResult
//
// Validation runs in three sequential phases: lexical pre-checks on raw bytes
// (encoding, BOM, tab indentation, directives, document-end marker), structural
// parse via gopkg.in/yaml.v3 into an AST, and an AST walk that enforces all
// remaining subset restrictions (anchors, aliases, explicit tags,
// single-quoted strings, duplicate keys, prohibited scalars, and reserved
// key prefixes). The sole public entry point is Validate([]byte).
package yaml
