// Package capability defines the public capability contract types for the provider SDK.
package capability

// TODO: public capability contract types for future provider SDK
// Reference: ADR-011, ADR-004, ADR-005, ADR-007, docs/spec/00_terminology.md §6

// CapabilityContract is the public type for a versioned semantic capability contract.
//
// Canonical identifier format: {identifier}@{version} for extension/provider capabilities.
// Platform capabilities are referenced as stable flags (no @version suffix).
//
// All sources (native extensions, MCP adapters, provider adapters, core engine)
// normalize into this single contract model at the planner boundary.
type CapabilityContract struct{} //nolint:revive
