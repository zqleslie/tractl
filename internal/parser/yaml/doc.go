// Package yaml implements the traCtl YAML format parser.
//
// Pipeline role: raw YAML bytes → *spec.TraCtlSpec
//
// Parsing runs in three phases: format validation via internal/validation/yaml
// (returns error on any subset violation), unmarshal via gopkg.in/yaml.v3
// into *spec.TraCtlSpec, and enrichment (ULID assignment to all identifiable
// entities, metadata.sourceFormat = "yaml", metadata.sourceRef from the caller).
// The canonical SpecValidator is not called here; callers run it separately.
package yaml
