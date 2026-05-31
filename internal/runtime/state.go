// state.go defines code for the runtime package.

package runtime

import "time"

// StepState is a string typed state for serialization and logs.
type StepState string

// Step state constants.
const (
	StatePending           StepState = "pending"
	StateRunning           StepState = "running"
	StateSucceeded         StepState = "succeeded"
	StateFailed            StepState = "failed"
	StateConditionalSkip   StepState = "conditional-skip"
	StateDependencySkipped StepState = "dependency-skipped"
	StateCancelled         StepState = "cancelled"
)

// IsTerminal returns true when the state is terminal.
func IsTerminal(s StepState) bool {
	switch s {
	case StateSucceeded, StateFailed, StateConditionalSkip, StateDependencySkipped, StateCancelled:
		return true
	case StatePending, StateRunning:
		return false
	default:
		return false
	}
}

// RequestTimeline is populated by the executor (Phase 5.2) and left as a
// struct here so executors can reference it without circular imports.
type RequestTimeline struct {
	DNSResolution    time.Duration
	TCPConnect       time.Duration
	TLSHandshake     time.Duration
	RequestSent      time.Duration
	TimeToFirstByte  time.Duration
	ResponseTransfer time.Duration
	TotalDuration    time.Duration
	WallClockStart   time.Time
	StatusCode       int
	ResponseSize     int64
}

// StepResult contains the outcome of executing a step.
type StepResult struct {
	StepID     string
	WorkflowID string
	State      StepState
	Extracts   map[string]string // extract.As -> resolved value
	StartedAt  time.Time
	FinishedAt time.Time
	Error      error
	// Timeline is populated by the executor (Phase 5.2)
	Timeline *RequestTimeline

	// HTTP response fields — populated by the HTTP executor.
	Status   int               // HTTP status code; 0 if no HTTP request
	Headers  map[string]string // canonicalised lowercase header keys
	Body     []byte            // raw response body; may be nil
	Metadata map[string]string // "url", "method", "attempt" keyed metadata
}
