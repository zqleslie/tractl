// Package toon implements the traCtl TOON subset validator defined in
// tractl_toon_spec.md §6–§22.
//
// Pipeline role: raw TOON bytes → *ValidationResult
//
// TOON is a strict YAML subset; validation runs in four sequential phases:
// UTF-8 encoding pre-check, lexical pre-checks (tab indentation, YAML
// directives, document separators, invalid escape sequences), structural
// parse via gopkg.in/yaml.v3, and an AST walk enforcing all TOON subset
// restrictions (anchors, aliases, explicit tags, single-quoted strings,
// merge keys, duplicate keys, and reserved key prefixes). The sole public
// entry point is Validate([]byte).
package toon
