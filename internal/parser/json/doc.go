// Package json implements the traCtl JSON format parser.
//
// Pipeline role: raw JSON bytes → *spec.TraCtlSpec
//
// Parsing runs in three phases: format validation via internal/validation/json
// (returns error on any subset violation), unmarshal via encoding/json into
// *spec.TraCtlSpec, and enrichment (ULID assignment to all identifiable
// entities, metadata.sourceFormat = "json", metadata.sourceRef from the caller).
// The canonical SpecValidator is not called here; callers run it separately.
package json
