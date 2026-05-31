// errors.go defines code for the runtime package.

package runtime

import "fmt"

// RuntimeError represents an error in runtime operations.
type RuntimeError struct { //nolint:revive
	Code       string
	StepID     string
	WorkflowID string
	Message    string
}

func (r *RuntimeError) Error() string {
	if r.StepID != "" && r.WorkflowID != "" {
		return fmt.Sprintf("[%s] workflow=%s step=%s: %s", r.Code, r.WorkflowID, r.StepID, r.Message)
	}
	if r.WorkflowID != "" {
		return fmt.Sprintf("[%s] workflow=%s: %s", r.Code, r.WorkflowID, r.Message)
	}
	return fmt.Sprintf("[%s] %s", r.Code, r.Message)
}

// Runtime error code constants.
const (
	ErrUnresolvableExpression = "ERR_UNRESOLVABLE_EXPRESSION"
	ErrStepNotFound           = "ERR_STEP_NOT_FOUND"
	ErrInvalidStepTransition  = "ERR_INVALID_STEP_TRANSITION"
)
