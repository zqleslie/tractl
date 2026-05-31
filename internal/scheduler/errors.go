// Package scheduler runs workflow DAGs with parallel step execution and failure policies.
package scheduler

// Scheduler error code constants.
const (
	ErrWorkflowCancelled = "workflow_cancelled"
	ErrSchedulerTimeout  = "scheduler_timeout"
	ErrStepPanicked      = "step_panicked"
)

// SchedulerError represents a scheduler-level failure.
type SchedulerError struct { //nolint:revive
	Code       string
	WorkflowID string
	StepID     string
	Message    string
}

func (e *SchedulerError) Error() string {
	if e.StepID != "" {
		return "scheduler: step " + e.StepID + " (" + e.Code + "): " + e.Message
	}
	return "scheduler: " + e.Code + ": " + e.Message
}
