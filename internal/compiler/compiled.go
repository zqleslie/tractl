// Package compiler translates a validated spec.TraCtlSpec into an executable CompiledPlan.
package compiler

import "regexp"

// CompiledRequest represents a protocol-specific invocation payload carried
// forward from the spec so the runtime does not need to re-read the spec.
type CompiledRequest struct {
	Protocol string            // "http" for MVP
	Target   string            // URL (may contain ${...} expressions)
	Method   string            // "GET", "POST", etc.
	Headers  map[string]string // header values (may contain ${...} expressions)
	Body     *CompiledBody
}

// CompiledBody represents a request body descriptor.
type CompiledBody struct {
	Encoding string      // "json", "form", "text", "raw"
	Content  interface{} // raw content from spec (string, map, etc.)
}

// CompiledRetry represents retry policy carried into runtime.
type CompiledRetry struct {
	MaxAttempts int
	Backoff     string   // "fixed" | "linear" | "exponential"
	Delay       string   // ISO 8601 duration e.g. "PT1S"
	RetryOn     []string // status codes or error classes
}

// CompiledAssertion is a single assertion projected into the compiled plan.
type CompiledAssertion struct {
	ID              string
	Kind            string // "status" | "header" | "body" | "schema"
	Target          string
	Op              string // "equals" | "contains" | "matches" | "exists" | "inRange"
	Expected        string
	Severity        string         // "error" | "warning"
	CompiledPattern *regexp.Regexp // pre-compiled for Op=="matches"; nil otherwise
}

// CompiledExtract describes an extraction to bind into runtime context.
type CompiledExtract struct {
	ID     string
	Source string // "status" | "header" | "body" | "metadata" | "timing"
	Path   string // path expression
	As     string // variable name to bind into context
	Scope  string // "step" | "workflow" | "spec"
}

// CompiledStep is the runtime artifact for a single step.
// It now contains the full step payload needed by the Phase 5 runtime.
type CompiledStep struct {
	StepID          string
	WorkflowID      string
	Runtime         string
	FailurePolicy   string
	DependsOn       []string
	ConcurrencySlot int
	CanSkip         bool

	// Payload fields from spec.Step
	When       string
	Request    *CompiledRequest
	Timeout    string
	Retry      *CompiledRetry
	Assertions []CompiledAssertion
	Extracts   []CompiledExtract
}

// CompiledWorkflow is the runtime artifact for a single workflow.
type CompiledWorkflow struct {
	WorkflowID       string
	DependsOn        []string // workflow-level dependencies (other workflow IDs)
	Steps            []CompiledStep
	ConcurrencyLimit int
	FailurePolicy    string
}

// CompiledPlan is the deterministic runtime artifact produced by the compiler.
// The Phase 5 runtime executes this directly without re-reading the source spec.
type CompiledPlan struct {
	PlanID              string
	CompiledAt          string
	WorkflowConcurrency int // max workflows running in parallel; default 10
	Workflows           []CompiledWorkflow
}
