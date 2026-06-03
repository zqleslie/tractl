// helpers_pure.go contains bridge helpers that have no syscall/js dependency.
// No build constraint — these can be compiled and tested outside a WASM environment.

package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/tractl/tractl/internal/engine"
	"github.com/tractl/tractl/internal/localapi"
	"github.com/tractl/tractl/internal/validation"
	"github.com/tractl/tractl/internal/version"
)

const (
	surfaceName      = "web-wasm"
	executionMode    = "browser"
	networkProvider  = "fetch"
	runtimeName      = "go"
	previewRuneLimit = 120
	wasmSourceRef    = "web-wasm"
)

// bridgeError returns a JS-safe error map with the given code and message.
func bridgeError(code string, message string) map[string]any {
	return map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
}

// executionError returns a TRACTL_EXECUTION_ERROR bridge error.
func executionError(message string) map[string]any {
	return bridgeError("TRACTL_EXECUTION_ERROR", message)
}

// runResultPayload serialises engine.RunResult to a JS-safe map for the workflow canvas.
// Field names are PascalCase (Go default) — workflowRunAdapter.ts reads Workflows, Steps,
// AssertionResults, etc. by that exact casing. The diagnostics field retains its lowercase
// json tag so workflowRunAdapter.ts can reach it as run.diagnostics.
// Pipeline errors (ParseError/ValidationError/PlanError) are checked by handleRun before
// this is called, so they will be empty strings in the serialised payload.
func runResultPayload(result *engine.RunResult) (map[string]any, error) {
	payload, err := jsonSafeObject(result)
	if err != nil {
		return nil, fmt.Errorf("marshal run result: %w", err)
	}
	convertDiagnosticDurationsToMs(payload)
	payload["surface"] = surfaceName
	payload["executionMode"] = executionMode
	payload["networkProvider"] = networkProvider
	return payload, nil
}

// convertDiagnosticDurationsToMs walks the diagnostics section of a serialised
// engine.RunResult payload and converts Duration values from nanoseconds to
// milliseconds in-place. diagnostics.WorkflowRecord.Duration and StepRecord.Duration
// are time.Duration fields with no json tag, so json.Marshal emits them as raw int64
// nanoseconds under the key "Duration". TypeScript receives milliseconds after this call.
func convertDiagnosticDurationsToMs(payload map[string]any) {
	diag, ok := payload["diagnostics"].(map[string]any)
	if !ok {
		return
	}
	convertDurationField(diag)
	workflows, _ := diag["Workflows"].([]any)
	for _, wfAny := range workflows {
		wf, ok := wfAny.(map[string]any)
		if !ok {
			continue
		}
		convertDurationField(wf)
		steps, _ := wf["Steps"].([]any)
		for _, stepAny := range steps {
			step, ok := stepAny.(map[string]any)
			if !ok {
				continue
			}
			convertDurationField(step)
		}
	}
}

// convertDurationField converts the "Duration" field in m from nanoseconds to
// milliseconds using integer division. json.Unmarshal decodes int64 as float64.
func convertDurationField(m map[string]any) {
	v, ok := m["Duration"]
	if !ok {
		return
	}
	ns, ok := v.(float64)
	if !ok || ns <= 0 {
		return
	}
	m["Duration"] = int64(ns) / 1_000_000
}

// requireSupportedFormat returns a TRACTL_WASM_INVALID_FORMAT bridge error when format
// is not supported by engine.ParseByFormat, or nil when the format is valid.
func requireSupportedFormat(format string) map[string]any {
	if engine.IsSupportedFormat(format) {
		return nil
	}
	return bridgeError("TRACTL_WASM_INVALID_FORMAT",
		fmt.Sprintf("unsupported format %q: must be %s", format, strings.Join(engine.SupportedFormats, ", ")))
}

// parseDocument dispatches to the correct parser based on format.
func parseDocument(document string, format string) (any, error) {
	return engine.ParseByFormat([]byte(document), format, wasmSourceRef)
}

// validateDocument parses and validates a document, returning any schema errors.
func validateDocument(document string, format string) ([]validation.ValidationError, error) {
	spec, err := engine.ParseByFormat([]byte(document), format, wasmSourceRef)
	if err != nil {
		return nil, err
	}
	return validation.NewSpecValidator().Validate(spec), nil
}

// validationFailure returns the JS response shape for a failed validation.
func validationFailure(errors []any) map[string]any {
	return map[string]any{
		"valid":   false,
		"surface": surfaceName,
		"errors":  errors,
	}
}

// formatValidationError wraps a parse-time error as a single-item validation error list.
func formatValidationError(err error) []any {
	return []any{
		map[string]any{
			"message": err.Error(),
		},
	}
}

// specValidationErrors converts engine ValidationError values to JS-safe maps.
func specValidationErrors(errs []validation.ValidationError) []any {
	errors := make([]any, 0, len(errs))
	for _, err := range errs {
		errors = append(errors, map[string]any{
			"field":   err.Field,
			"code":    string(err.Code),
			"message": err.Message,
		})
	}
	return errors
}

// jsonSafeObject round-trips a value through JSON to produce a plain map[string]any.
func jsonSafeObject(value any) (map[string]any, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("canonical spec marshal: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("canonical spec unmarshal: %w", err)
	}

	return result, nil
}

// runRequestFromJSON unmarshals a RequestDef JSON string, calls RunRequestDef,
// and returns a JS-safe result map or a bridgeError map on failure.
// Extracted from handleRunRequest to allow testing without syscall/js.
func runRequestFromJSON(jsonStr string) map[string]any {
	var def localapi.RequestDef
	if err := json.Unmarshal([]byte(jsonStr), &def); err != nil {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", err.Error())
	}
	result, err := localapi.RunRequestDef(def)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	payload, err := jsonSafeObject(result)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}

// preview returns the first previewRuneLimit runes of the document.
func preview(document string) string {
	runes := []rune(document)
	if len(runes) <= previewRuneLimit {
		return document
	}

	return string(runes[:previewRuneLimit])
}

// projectVersion reads the binary's embedded module version or returns the fallback.
func projectVersion() string {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
			return buildInfo.Main.Version
		}
	}

	return version.FallbackVersion
}
