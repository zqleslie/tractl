// Package validation provides the canonical contract enforcement layer for
// traCtl specification documents.
//
// Pipeline role: *spec.TraCtlSpec → []ValidationError
//
// SpecValidator runs all structural rules (schema version, capabilities,
// workflow and step constraints, DAG acyclicity) and returns the complete
// error list — no short-circuiting. FormatValidator and ValidatorFor provide
// a unified interface for selecting a format-specific byte-level validator
// (yaml, json, toon) by name, enabling the engine to validate raw documents
// without importing each format package concretely.
package validation
