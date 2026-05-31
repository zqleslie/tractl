// Package extract evaluates spec.Extract rules and writes values into execution context.
package extract

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/spec"
)

// Engine evaluates a spec.Extract against a StepResult and writes the
// extracted value into the execution context.
type Engine interface {
	Extract(
		e spec.Extract,
		result *runtime.StepResult,
		ctx *runtime.ExecutionContext,
		traceID, spanID string,
		workflowStart time.Time,
	) (ExtractResult, EvalEvent, error)
}

// compile-time interface check
var _ Engine = (*ExtractEngine)(nil)

// ExtractEngine is the default implementation of Engine.
type ExtractEngine struct{} //nolint:revive

// New returns a ready-to-use *ExtractEngine.
func New() *ExtractEngine { return &ExtractEngine{} }

// Extract evaluates e against result, writes the value into ctx, and returns
// the result and observability event. The actual extracted value never appears
// in the returned EvalEvent.
func (eng *ExtractEngine) Extract(
	e spec.Extract,
	result *runtime.StepResult,
	ctx *runtime.ExecutionContext,
	traceID, spanID string,
	workflowStart time.Time,
) (ExtractResult, EvalEvent, error) {
	start := time.Now()

	variableName := e.As
	if variableName == "" {
		variableName = e.ID
	}

	value, err := extractValue(e, result)
	if err != nil {
		return ExtractResult{}, EvalEvent{}, err
	}

	if err := writeToContext(e.Scope, e.ID, variableName, value, result.StepID, ctx); err != nil {
		return ExtractResult{}, EvalEvent{}, err
	}

	dur := time.Since(start)
	placeholder := "[extract:" + e.ID + "]"

	res := ExtractResult{
		ExtractID:        e.ID,
		ExtractULID:      e.ULID,
		Source:           e.Source,
		As:               variableName,
		Scope:            e.Scope,
		ValuePlaceholder: placeholder,
		Resolved:         true,
		Duration:         dur,
	}

	ev := EvalEvent{
		TraceID:          traceID,
		SpanID:           spanID,
		ExtractID:        e.ID,
		Source:           e.Source,
		Variable:         variableName,
		ValuePlaceholder: placeholder,
		Scope:            e.Scope,
		MonotonicOffset:  time.Since(workflowStart),
	}

	return res, ev, nil
}

// writeToContext writes the extracted value into the correct scope on ctx.
// The value is always mirrored into the step's extract map (for expression
// resolution via steps.<id>.extracts.<name>) and additionally written to the
// declared variable scope. Scope defaults to workflow per traCtlSpec §11.3.
func writeToContext(scope spec.ExtractScope, _ string, variableName, value, stepID string, ctx *runtime.ExecutionContext) error {
	// Always mirror into step extracts so ${steps.X.extracts.Y} resolution works.
	ctx.SetStepExtract(stepID, variableName, value)

	switch scope {
	case spec.ExtractScopeStep:
		return nil // already written above
	case spec.ExtractScopeWorkflow, "":
		return ctx.SetVar("workflow", variableName, value)
	case spec.ExtractScopeSpec:
		return ctx.SetVar("spec", variableName, value)
	default:
		return &ExtractError{
			Code:    ErrInvalidScope,
			Message: fmt.Sprintf("unknown scope %q", scope),
		}
	}
}

// extractValue dispatches on e.Source and returns the raw string value.
func extractValue(e spec.Extract, result *runtime.StepResult) (string, error) {
	switch e.Source {
	case "status":
		return strconv.Itoa(result.Status), nil

	case "header":
		return extractHeader(e.Path, result.Headers)

	case "body":
		return extractBody(e.Path, result.Body)

	case "metadata":
		return extractMetadata(e.Path, result.Metadata)

	case "timing":
		return extractTiming(e.Path, result.Timeline)

	case "extension":
		return "", &ExtractError{
			Code:    ErrExtensionStub,
			Message: "extension source is not yet implemented",
		}

	default:
		return "", &ExtractError{
			Code:    ErrUnknownSource,
			Message: fmt.Sprintf("unknown source %q", e.Source),
		}
	}
}

func extractHeader(path string, headers map[string]string) (string, error) {
	lp := strings.ToLower(path)
	for k, v := range headers {
		if strings.ToLower(k) == lp {
			return v, nil
		}
	}
	return "", &ExtractError{
		Code:    ErrPathNotFound,
		Message: fmt.Sprintf("header %q not found", path),
	}
}

func extractBody(path string, body []byte) (string, error) {
	if len(body) == 0 {
		return "", &ExtractError{
			Code:    ErrPathNotFound,
			Message: "body is nil or empty",
		}
	}
	r := gjson.GetBytes(body, path)
	if !r.Exists() {
		return "", &ExtractError{
			Code:    ErrPathNotFound,
			Message: fmt.Sprintf("path %q not found in body", path),
		}
	}
	return r.String(), nil
}

func extractMetadata(path string, metadata map[string]string) (string, error) {
	switch path {
	case "url", "method", "attempt":
		v, ok := metadata[path]
		if !ok {
			return "", &ExtractError{
				Code:    ErrMetadataField,
				Message: fmt.Sprintf("metadata key %q not present", path),
			}
		}
		return v, nil
	default:
		return "", &ExtractError{
			Code:    ErrMetadataField,
			Message: fmt.Sprintf("unknown metadata field %q", path),
		}
	}
}

// extractTiming reads a named timing segment from the request timeline.
// Returns "" for all fields when timeline is nil (non-HTTP steps).
func extractTiming(path string, tl *runtime.RequestTimeline) (string, error) {
	if tl == nil {
		return "", &ExtractError{
			Code:    ErrTimingField,
			Message: "no timing data available (timeline is nil)",
		}
	}
	var d time.Duration
	switch path {
	case "dns":
		d = tl.DNSResolution
	case "tcp":
		d = tl.TCPConnect
	case "tls":
		d = tl.TLSHandshake
	case "sent":
		d = tl.RequestSent
	case "ttfb":
		d = tl.TimeToFirstByte
	case "transfer":
		d = tl.ResponseTransfer
	case "total":
		d = tl.TotalDuration
	default:
		return "", &ExtractError{
			Code:    ErrTimingField,
			Message: fmt.Sprintf("unknown timing field %q", path),
		}
	}
	return strconv.FormatInt(d.Milliseconds(), 10) + "ms", nil
}
