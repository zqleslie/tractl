package assertion

import "time"

// EvalEvent is the ADR-014 §5 assertion evaluation observability event.
// It carries only a placeholder for the assertion target value —
// never the actual extracted credential-bearing value.
type EvalEvent struct {
	TraceID     string
	SpanID      string
	AssertionID string
	Kind        string
	Outcome     AssertionOutcome
	// TargetPlaceholder describes what was checked, not the value itself.
	// Example: "response.status", "response.headers.Authorization[masked]"
	TargetPlaceholder string
	MonotonicOffset   time.Duration // relative to workflow execution start
}
