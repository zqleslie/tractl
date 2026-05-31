// Package common provides shared utilities used by all format-specific parser
// sub-packages (yaml, json, toon).
//
// Pipeline role: shared helpers consumed during the enrichment phase of
// parsing — ULID generation (NewULID), spec entity population (AssignULIDs),
// metadata annotation (SetMetadata), and format-agnostic unmarshaling
// (UnmarshalYAMLBytes, UnmarshalJSONBytes).
//
// Isolation invariant: this package must not import any format-specific parser
// package or any format-specific validator package to avoid cycles. It may
// import internal/spec and internal/validation/common.
package common
