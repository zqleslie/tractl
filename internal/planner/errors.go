package planner

import "fmt"

// PlannerError describes a planning-phase failure.
type PlannerError struct { //nolint:revive
	Code       string
	WorkflowID string
	StepID     string
	Message    string
}

// Error implements the error interface.
func (e *PlannerError) Error() string {
	if e.WorkflowID != "" && e.StepID != "" {
		return fmt.Sprintf("[%s] workflow %q step %q: %s", e.Code, e.WorkflowID, e.StepID, e.Message)
	}
	if e.WorkflowID != "" {
		return fmt.Sprintf("[%s] workflow %q: %s", e.Code, e.WorkflowID, e.Message)
	}
	if e.StepID != "" {
		return fmt.Sprintf("[%s] step %q: %s", e.Code, e.StepID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Exported error code constants.
const (
	ErrCyclicDependency       = "CYCLIC_DEPENDENCY"
	ErrUnknownCapability      = "UNKNOWN_CAPABILITY"
	ErrIncompatibleCapability = "INCOMPATIBLE_CAPABILITY"
	ErrEmptySpec              = "EMPTY_SPEC"
)

// plannerErr constructs a PlannerError with standard formatting.
func plannerErr(code, workflowID, stepID, msg string) *PlannerError {
	return &PlannerError{
		Code:       code,
		WorkflowID: workflowID,
		StepID:     stepID,
		Message:    msg,
	}
}
