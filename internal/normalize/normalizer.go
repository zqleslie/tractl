// Package normalize is the normalization layer between parser output
// (Canonical Source Model) and the Overlay Engine.
//
// For native sources (YAML, JSON, TOON), normalization is performed
// directly by the parser layer (internal/parser/*). This package is
// reserved for interoperability source adapters (OpenAPI, Postman,
// Bruno, HAR) that must translate foreign formats into *spec.TraCtlSpec.
//
// References: ADR-001 §2, ADR-002, HLD §4.4
// Implementation target: Phase 4 (interoperability source adapters)
package normalize
