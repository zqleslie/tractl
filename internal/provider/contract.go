// Package provider defines the provider contract layer for capability adapters.
package provider

// TODO: provider contract layer
// Reference: HLD §4.3, ADR-007, ADR-011
// Responsibilities:
//   - source parsing and source validation
//   - compatibility interpretation with declared fidelity guarantees
//   - capability declaration (normalised into canonical capability contract model)
// Constraint: providers stop at interpretation boundaries.
//   Core execution semantics remain engine-owned unless explicitly delegated
//   through governed specialist provider architecture.
