package scheduler

import (
	"time"

	"github.com/tractl/tractl/internal/runtime"
)

// StepOutcome holds the result of a single step execution within a workflow run.
type StepOutcome struct {
	StepID string
	State  runtime.StepState
	Result *runtime.StepResult
	Error  error
	// WaitDuration is the time the step spent waiting for a concurrency slot
	// after its dependencies finished. Zero for steps that started immediately.
	WaitDuration time.Duration
}

// WorkflowResult holds the aggregate outcome of executing a single workflow.
type WorkflowResult struct {
	WorkflowID   string
	ExecutionID  string
	Outcomes     []StepOutcome
	OverallState runtime.StepState
	StartedAt    time.Time
	FinishedAt   time.Time
}
