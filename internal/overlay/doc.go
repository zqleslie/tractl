// Package overlay applies overlay documents to a base traCtl specification,
// merging values according to the overlay rules defined in the overlay spec.
//
// Pipeline role: base spec (any) + []*OverlayDocument → patched spec (any)
//
// The engine operates on the spec through a JSON round-trip deep clone and
// map[string]any traversal, making it format-neutral. Patches are applied in
// authored order; overlays are applied in stack order (first overlay first).
// overlay must not import any format-specific validator sub-package
// (internal/validation/yaml, json, or toon); format validation of overlay
// documents is the caller's responsibility before passing them to Engine.Apply.
package overlay
