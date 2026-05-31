// Package normalize provides implicit dependency resolution for traCtl
// specifications, bridging the gap between what an author writes and what
// the planner needs to schedule.
//
// Pipeline role: *spec.TraCtlSpec (raw) → *spec.TraCtlSpec (normalised)
//
// Normalize scans all expression-bearing string fields of every step
// (When, Request.Target, Request.Headers values, Request.Body content) for
// ${steps.<id>...} template references, then merges any discovered same-workflow
// step IDs into the step's DependsOn list — deduped and sorted for
// deterministic output. Cross-workflow references are not added to DependsOn;
// they are resolved at runtime via shared execution context. The source spec
// is never mutated; a new *spec.TraCtlSpec is always returned.
package normalize
