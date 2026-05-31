package extract

import (
	"time"

	"github.com/tractl/tractl/internal/spec"
)

// EvalEvent is the ADR-014 §5 extract evaluation observability event.
// Extracted variable values MUST NOT appear here (credential-safety).
type EvalEvent struct {
	TraceID          string
	SpanID           string
	ExtractID        string
	Source           string
	Variable         string
	ValuePlaceholder string
	Scope            spec.ExtractScope
	MonotonicOffset  time.Duration
}
