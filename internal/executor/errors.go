package executor

// ExecutorError represents an executor-level error.
type ExecutorError struct { //nolint:revive
	Code       string
	StepID     string
	WorkflowID string
	Message    string
}

func (e *ExecutorError) Error() string {
	if e.StepID != "" && e.WorkflowID != "" {
		return "[" + e.Code + "] workflow=" + e.WorkflowID + " step=" + e.StepID + ": " + e.Message
	}
	if e.WorkflowID != "" {
		return "[" + e.Code + "] workflow=" + e.WorkflowID + ": " + e.Message
	}
	return "[" + e.Code + "] " + e.Message
}

// Executor error code constants.
const (
	ErrRequestBuildFailed  = "ERR_REQUEST_BUILD_FAILED"
	ErrRequestFailed       = "ERR_REQUEST_FAILED"
	ErrTimeout             = "ERR_TIMEOUT"
	ErrMaxRetriesExceeded  = "ERR_MAX_RETRIES_EXCEEDED"
	ErrUnsupportedProtocol = "ERR_UNSUPPORTED_PROTOCOL"
)
