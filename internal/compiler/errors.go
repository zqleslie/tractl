package compiler

import "fmt"

// Compiler error code constants.
const (
	ErrNilPlan            = "ERR_NIL_PLAN"
	ErrEmptyPlan          = "ERR_EMPTY_PLAN"
	ErrEmptyRuntime       = "ERR_EMPTY_RUNTIME"
	ErrEmptyStepID        = "ERR_EMPTY_STEP_ID"
	ErrEmptyWorkflowID    = "ERR_EMPTY_WORKFLOW_ID"
	ErrMissingStepPayload = "ERR_MISSING_STEP_PAYLOAD"
)

// CompilerError is a structured error produced during plan compilation.
type CompilerError struct { //nolint:revive
	Code       string
	WorkflowID string
	StepID     string
	Message    string
}

func (e *CompilerError) Error() string {
	if e.WorkflowID != "" && e.StepID != "" {
		return fmt.Sprintf("[%s] workflow=%s step=%s: %s", e.Code, e.WorkflowID, e.StepID, e.Message)
	}
	if e.WorkflowID != "" {
		return fmt.Sprintf("[%s] workflow=%s: %s", e.Code, e.WorkflowID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func compilerErr(code, workflowID, stepID, msg string) *CompilerError {
	return &CompilerError{
		Code:       code,
		WorkflowID: workflowID,
		StepID:     stepID,
		Message:    msg,
	}
}
