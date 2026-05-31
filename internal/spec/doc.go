// Package spec defines the canonical traCtl document model — the Go types that
// represent a parsed and normalised traCtl specification.
//
// Pipeline role: consumed by every package that operates on a traCtl document.
// spec is the shared lingua franca of the pipeline.
//
// Primary types: TraCtlSpec (document root), Workflow, Step, Assertion, Extract,
// Metadata, AuthProfile, Environments, and ExtensionRef. spec is a types-only
// package and must not import any other internal traCtl package.
package spec
