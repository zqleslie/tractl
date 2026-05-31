package diagnostics

// TraceEvent is one entry in the Execution Provenance Trace (ADR-014 §5).
// MonotonicMs is elapsed milliseconds from ExecutionRecord.StartedAt,
// computed by Collector using time.Since on its monotonic origin time.
// Detail must be pre-masked by the caller per ADR-014 §6.
type TraceEvent struct {
	MonotonicMs int64
	Kind        TraceEventKind
	WorkflowID  string // empty for workflow-agnostic events
	StepID      string // empty for non-step events
	SpanID      string // step-scoped identifier; empty for workflow-level events
	Detail      string // credential-safe human-readable note
}

// TraceEventKind identifies the phase or action that produced a TraceEvent.
type TraceEventKind string

// Trace event kind constants.
const (
	EventWorkflowStarted      TraceEventKind = "workflow.started"
	EventValidationCompleted  TraceEventKind = "validation.completed"
	EventPlanningCompleted    TraceEventKind = "planning.completed"
	EventCompilationCompleted TraceEventKind = "compilation.completed"
	EventStepStarted          TraceEventKind = "step.started"
	EventRequestDispatched    TraceEventKind = "request.dispatched"
	EventResponseReceived     TraceEventKind = "response.received"
	EventAssertionEvaluated   TraceEventKind = "assertion.evaluated"
	EventExtractEvaluated     TraceEventKind = "extract.evaluated"
	EventStepCompleted        TraceEventKind = "step.completed"
	EventWorkflowCompleted    TraceEventKind = "workflow.completed"

	// EventEngineLog is emitted by the engine for diagnostic trace messages.
	// Added Stream C HIGH #4. Stream D: integrate with structured log levels as needed.
	EventEngineLog TraceEventKind = "engine.log"
)
