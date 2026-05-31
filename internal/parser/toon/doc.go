// Package toon implements the traCtl TOON format parser.
//
// Pipeline role: raw TOON bytes → *spec.TraCtlSpec
//
// TOON is a strict YAML subset (tractl_toon_spec.md §3); unmarshaling reuses
// gopkg.in/yaml.v3 via parser/common.UnmarshalYAMLBytes — only the format
// validator differs from the yaml parser. Parsing runs in three phases:
// format validation via internal/validation/toon, unmarshal, and enrichment
// (ULID assignment, metadata.sourceFormat = "toon", metadata.sourceRef).
// The canonical SpecValidator is not called here; callers run it separately.
package toon
