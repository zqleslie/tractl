package planner

// StepPlan describes the execution plan for a single step.
type StepPlan struct {
	// StepID is the step identifier from spec.Step.ID.
	StepID string
	// WorkflowID is the parent workflow identifier.
	WorkflowID string
	// Runtime is the resolved runtime name (e.g., "http").
	Runtime string
	// FailurePolicy is either "resilient" or "failFast".
	FailurePolicy string
	// DependsOn is a copy of the step's dependency edges (all explicit + implicit, pre-merged).
	DependsOn []string
	// ConcurrencySlot is the 1-based concurrency slot within the workflow's budget.
	ConcurrencySlot int
	// CanSkip is true if the step has at least one dependency (eligible for skip state).
	CanSkip bool
}

// WorkflowPlan describes the execution plan for a single workflow.
type WorkflowPlan struct {
	// WorkflowID is the workflow identifier.
	WorkflowID string
	// DependsOn lists workflow IDs that must complete before this workflow starts.
	DependsOn []string
	// Steps is the ordered list of steps in topological order (deterministic).
	Steps []StepPlan
	// ConcurrencyLimit is the workflow's concurrency budget.
	ConcurrencyLimit int
	// FailurePolicy is the workflow-level default: "resilient" or "failFast".
	FailurePolicy string
}

// ExecutionPlan describes the complete execution plan for a traCtlSpec.
type ExecutionPlan struct {
	// PlanID is a unique plan identifier (ULID).
	PlanID string
	// WorkflowConcurrency is the maximum number of workflows that may run in parallel.
	// Default 10 when not specified.
	WorkflowConcurrency int
	// Workflows is the topo-sorted list of workflow plans.
	Workflows []WorkflowPlan
	// Trace holds metadata about plan generation.
	Trace PlanTrace
}

// PlanTrace holds telemetry metadata about plan generation.
type PlanTrace struct {
	// PlanID is the plan identifier.
	PlanID string
	// WorkflowCount is the number of workflows in the plan.
	WorkflowCount int
	// TotalSteps is the total number of steps across all workflows.
	TotalSteps int
	// GeneratedAt is the RFC3339-formatted generation timestamp.
	GeneratedAt string
}
