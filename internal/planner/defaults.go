// defaults.go defines default tuning constants for the planner.
// These constants are exported so dependent packages (scheduler)
// can reference a single source of truth rather than duplicating
// magic literals.
package planner

// DefaultStepConcurrency is the default maximum number of steps
// that may execute in parallel within a single workflow.
// Both the planner and scheduler reference this constant so that
// changing the default requires editing one location only.
const DefaultStepConcurrency = 4

// DefaultWorkflowConcurrency is the default maximum number of
// workflows that may execute in parallel within a single run.
const DefaultWorkflowConcurrency = 10
