// Package normalizer implements Phase 3.5 of the canonical pipeline:
// the implicit dependency normalizer.
//
// It scans expression-bearing string fields of every step for
// ${steps.<id>...} references and merges discovered step IDs into
// the step's DependsOn list (deduped, sorted, deterministic).
//
// Pipeline position:
//
//	parser → overlay → normalizer → validation → planner → compiler → ...
//
// The normalizer does NOT call the canonical validator and does NOT
// mutate the input spec. It returns a new *spec.TraCtlSpec.
package normalizer

import "fmt"

// NormalizerError describes a normalization-phase failure.
type NormalizerError struct { //nolint:revive
	Code       string
	WorkflowID string
	StepID     string
	Message    string
}

// Error implements the error interface.
func (e *NormalizerError) Error() string {
	if e.WorkflowID != "" && e.StepID != "" {
		return fmt.Sprintf("[%s] workflow %q step %q: %s", e.Code, e.WorkflowID, e.StepID, e.Message)
	}
	if e.WorkflowID != "" {
		return fmt.Sprintf("[%s] workflow %q: %s", e.Code, e.WorkflowID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error code constants.
const (
	ErrSelfReference = "SELF_REFERENCE"
	ErrDeadReference = "DEAD_REFERENCE"
)

// normalizerErr constructs a NormalizerError with standard formatting.
func normalizerErr(code, workflowID, stepID, msg string) *NormalizerError {
	return &NormalizerError{
		Code:       code,
		WorkflowID: workflowID,
		StepID:     stepID,
		Message:    msg,
	}
}
