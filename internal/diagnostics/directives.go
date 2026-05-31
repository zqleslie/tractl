// Package diagnostics assembles and formats execution observability records.
// directives.go defines the output-side types for diagnostics collection.
// These constants mirror spec.DiagnosticsKind values but are defined independently
// so internal/diagnostics remains free of internal/ imports.
// Runtime population of StepDiagnosticsOutput is deferred to HTTP client
// instrumentation in a future phase; the types are stubs in Phase 7.5.
package diagnostics

// Kind constants mirror spec.DiagnosticsKind values (§15.2).
// TestKindConstants_MirrorSpecValues guards against drift.
const (
	KindTLS        = "tls"
	KindTCP        = "tcp"
	KindTransport  = "transport"
	KindLifecycle  = "lifecycle"
	KindDNS        = "dns"
	KindConnection = "connection"
)

// Retention constants mirror spec.DiagnosticsRetention values (§15.1).
// TestRetentionConstants_MirrorSpecValues guards against drift.
const (
	RetentionStep     = "step"
	RetentionWorkflow = "workflow"
	RetentionSpec     = "spec"
)

// ValidKinds is the complete set of permitted diagnostics kind identifiers (§15.2).
// No additional kinds are admitted without explicit architecture extension.
var ValidKinds = map[string]struct{}{
	KindTLS:        {},
	KindTCP:        {},
	KindTransport:  {},
	KindLifecycle:  {},
	KindDNS:        {},
	KindConnection: {},
}

// ValidRetentions is the complete set of permitted retention scope values (§15.1).
var ValidRetentions = map[string]struct{}{
	RetentionStep:     {},
	RetentionWorkflow: {},
	RetentionSpec:     {},
}

// StepDiagnosticsOutput holds diagnostics events captured during one step's
// execution. Fields are populated by the runtime HTTP client instrumentation
// when specific kinds are enabled. In Phase 7.5 this type is a stub;
// runtime population is deferred to HTTP client instrumentation.
//
// Diagnostics output MUST NOT alter execution semantics (§15.3).
type StepDiagnosticsOutput struct {
	// EnabledKinds records which kinds were active during this step's execution.
	EnabledKinds []string

	// Events holds the captured diagnostics events in chronological order.
	// Empty until runtime instrumentation is implemented.
	Events []DiagnosticsEvent
}

// DiagnosticsEvent is one captured event from a diagnostics kind's instrumentation.
// The Detail map carries kind-specific key-value pairs (e.g., for "tls":
// cipher, version, serverName; for "dns": host, resolvedIP, durationMs).
// All values in Detail are strings; numeric values are formatted by the producer.
type DiagnosticsEvent struct { //nolint:revive
	MonotonicMs int64             // elapsed ms from workflow start
	Kind        string            // one of the Kind* constants
	Detail      map[string]string // kind-specific key-value pairs; never nil
}
