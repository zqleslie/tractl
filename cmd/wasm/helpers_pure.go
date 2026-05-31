// helpers_pure.go contains bridge helpers that have no syscall/js dependency.
// No build constraint — these can be compiled and tested outside a WASM environment.

package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/tractl/tractl/internal/engine"
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

// runResultPayload converts an engine RunResult into a JS-safe map with surface metadata.
func runResultPayload(result *engine.RunResult) (map[string]any, error) {
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal run result: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal run result: %w", err)
	}

	payload["surface"] = surfaceName
	payload["executionMode"] = executionMode
	payload["networkProvider"] = networkProvider
	return payload, nil
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
// Delegates to engine.ParseByFormat — the single authoritative format-dispatch
// switch — so adding a new format only requires updating internal/engine/parse.go.
func parseDocument(document string, format string) (any, error) {
	return engine.ParseByFormat([]byte(document), format, wasmSourceRef)
}

// validateDocument parses and validates a document, returning any schema errors.
// Delegates parsing to engine.ParseByFormat to avoid duplicating format dispatch.
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

	// Stream D: update internal/localapi to use version.FallbackVersion too.
	return version.FallbackVersion
}
