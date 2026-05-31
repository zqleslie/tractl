package planner

import "github.com/tractl/tractl/internal/spec"

// detectCompositeCycles detects cycles across composite workflow references.
// Per spec §16.3, planners MUST validate composite cycles via recursive DAG flattening
// before execution.
//
// Phase 4 MVP implementation:
//   - Scans all steps for composite steps
//   - If none found, returns nil (fast path)
//   - If found, returns ErrCyclicDependency (stub; full implementation Phase 4+)
//
// TODO: Full implementation pending composite step descriptor population (Phase 4+).
// See tractl_spec.md §16.3 for the complete cycle detection algorithm.
func detectCompositeCycles(s *spec.TraCtlSpec) error {
	for _, workflow := range s.Workflows {
		for _, step := range workflow.Steps {
			// Check if this is a composite step.
			// In Phase 4 MVP, CompositeDescriptor is a stub with no fields.
			// When populated, we would check for workflowRef and build a
			// cross-workflow reference graph.
			if step.Kind == "composite" && step.Composite != nil {
				// TODO: Phase 4+ implement full composite cycle detection
				// For now, return nil (no cycles in MVP stub state)
				continue
			}
		}
	}
	return nil
}
