// helpers.go defines small utility functions shared across the engine pipeline stages.
// traceIDToSeed: converts a trace ID string to a deterministic int64 seed.
// joinValidationErrors: formats validation errors into a human-readable string.
// emitEngineLog: routes engine trace messages through the diagnostics collector.

package engine

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/validation"
)

// fallbackSourceRef is the source reference used when a spec is provided as an
// in-memory document rather than a named file. Checked at multiple points in the pipeline.
const fallbackSourceRef = "document"

// emitEngineLog records an engine trace message on the collector unless cfg.Quiet is set.
func emitEngineLog(collector *diagnostics.Collector, cfg Config, workflowID, stepID, format string, args ...any) {
	if cfg.Quiet {
		return
	}
	collector.Emit(diagnostics.EventEngineLog, workflowID, stepID, "", fmt.Sprintf(format, args...))
}

// traceIDToSeed converts a TraceID string to a deterministic int64 seed.
// Uses FNV-1a hash (hash/fnv from stdlib).
// If traceID is empty, returns 0.
func traceIDToSeed(traceID string) int64 {
	if traceID == "" {
		return 0
	}
	h := fnv.New64a()
	h.Write([]byte(traceID))
	return int64(h.Sum64())
}

// joinValidationErrors formats validation errors into a single string.
func joinValidationErrors(errs []validation.ValidationError) string {
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = fmt.Sprintf("[%s] %s: %s", e.Code, e.Field, e.Message)
	}
	return strings.Join(msgs, "; ")
}
