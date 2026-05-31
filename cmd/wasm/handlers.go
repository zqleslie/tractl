//go:build js && wasm

// handlers.go implements the JS-callable handler functions registered by the tractl
// WASM bridge. Each handler is passed to js.FuncOf() in bridge.go's registerBridge().
//
// Format validation: every handler that accepts a format argument delegates to
// engine.IsSupportedFormat via requireSupportedFormat before calling the engine.
// Invalid formats return a bridgeError with code TRACTL_WASM_INVALID_FORMAT.
//
// Note: handlers take js.Value arguments and can only run inside a WASM/JS runtime.
// Direct unit tests are not possible; the underlying logic (parseDocument, validateDocument,
// bridgeError, etc.) is tested via helpers_test.go.

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/tractl/tractl/internal/engine"
	"github.com/tractl/tractl/internal/localapi"
)

// handleVersion returns bridge surface, runtime name, and current binary version.
func handleVersion(this js.Value, args []js.Value) any {
	return map[string]any{
		"surface": surfaceName,
		"runtime": runtimeName,
		"version": projectVersion(),
	}
}

// handleCapabilities returns what the WASM surface supports and explicitly does not support.
func handleCapabilities(this js.Value, args []js.Value) any {
	return map[string]any{
		"surface":         surfaceName,
		"executionMode":   executionMode,
		"networkProvider": networkProvider,
		"supports": []any{
			"http",
			"graphql",
			"workflow",
			"assertions",
			"extracts",
		},
		"unsupported": []any{
			"filesystem",
			"git",
			"playwright",
			"mcp",
			"mtls",
		},
	}
}

// handleEcho echoes back the document string with its byte length and a rune preview.
func handleEcho(this js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", "echo expects one string argument")
	}

	document := args[0].String()
	return map[string]any{
		"surface":  surfaceName,
		"received": true,
		"length":   len(document),
		"preview":  preview(document),
	}
}

// wasmDocumentAndFormat extracts document and format from handler args and validates format.
func wasmDocumentAndFormat(args []js.Value, handlerName string) (document, format string, errResp any) {
	if len(args) != 2 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeString {
		return "", "", bridgeError("TRACTL_WASM_INVALID_ARGUMENT", handlerName+" expects document and format string arguments")
	}
	document = args[0].String()
	format = strings.ToLower(strings.TrimSpace(args[1].String()))
	if bad := requireSupportedFormat(format); bad != nil {
		return "", "", bad
	}
	return document, format, nil
}

// handleParse parses the document and returns the canonical spec object.
func handleParse(this js.Value, args []js.Value) any {
	document, format, errResp := wasmDocumentAndFormat(args, "parse")
	if errResp != nil {
		return errResp
	}

	spec, err := parseDocument(document, format)
	if err != nil {
		return bridgeError("TRACTL_PARSE_ERROR", err.Error())
	}

	canonicalSpec, err := jsonSafeObject(spec)
	if err != nil {
		return bridgeError("TRACTL_PARSE_ERROR", err.Error())
	}

	return map[string]any{
		"surface": surfaceName,
		"parsed":  true,
		"spec":    canonicalSpec,
	}
}

// handleValidate validates the document and returns any schema errors.
func handleValidate(this js.Value, args []js.Value) any {
	document, format, errResp := wasmDocumentAndFormat(args, "validate")
	if errResp != nil {
		return errResp
	}

	errs, err := validateDocument(document, format)
	if err != nil {
		return validationFailure(formatValidationError(err))
	}
	if len(errs) > 0 {
		return validationFailure(specValidationErrors(errs))
	}

	return map[string]any{
		"valid":   true,
		"surface": surfaceName,
	}
}

// handleRun executes the document through the engine and returns the full run result.
func handleRun(this js.Value, args []js.Value) any {
	document, format, errResp := wasmDocumentAndFormat(args, "run")
	if errResp != nil {
		return errResp
	}

	result := engine.New().RunDocument(engine.DocumentConfig{
		Document:  document,
		Format:    format,
		SourceRef: wasmSourceRef,
		Quiet:     true,
		Verbose:   true,
	})

	if result.ParseError != "" {
		return executionError(result.ParseError)
	}
	if result.ValidationError != "" {
		return executionError(result.ValidationError)
	}
	if result.PlanError != "" {
		return executionError(result.PlanError)
	}

	payload, err := runResultPayload(result)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}

// handleRunRequest accepts a RequestDef JSON string.
// Returns RunResult JSON — same shape as POST /api/v1/run.
func handleRunRequest(this js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", "runRequest expects one JSON string argument")
	}
	var def localapi.RequestDef
	if err := json.Unmarshal([]byte(args[0].String()), &def); err != nil {
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

// handleLayout accepts a JSON array of StepRef and returns WorkflowLayoutResponse.
func handleLayout(this js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", "layout expects one JSON string argument")
	}
	var steps []localapi.StepRef
	if err := json.Unmarshal([]byte(args[0].String()), &steps); err != nil {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", err.Error())
	}
	result := localapi.ComputeWorkflowLayout(steps)
	payload, err := jsonSafeObject(result)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}

// handleInferDeps accepts a JSON array of StepScanDef and returns InferDepsResponse.
func handleInferDeps(this js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", "inferDeps expects one JSON string argument")
	}
	var steps []localapi.StepScanDef
	if err := json.Unmarshal([]byte(args[0].String()), &steps); err != nil {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", err.Error())
	}
	result := localapi.InferImplicitDeps(steps)
	payload, err := jsonSafeObject(result)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}

// handleExportWorkflow accepts a WorkflowExportRequest JSON string and returns WorkflowExportResponse.
func handleExportWorkflow(this js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", "exportWorkflow expects one JSON string argument")
	}
	var req localapi.WorkflowExportRequest
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return bridgeError("TRACTL_WASM_INVALID_ARGUMENT", err.Error())
	}
	result, err := localapi.ExportWorkflow(req)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	payload, err := jsonSafeObject(result)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}

// handleDefaults returns the canonical engine defaults — synchronous, no I/O.
func handleDefaults(this js.Value, args []js.Value) any {
	payload, err := jsonSafeObject(localapi.DefaultEngineSettings)
	if err != nil {
		return bridgeError("TRACTL_EXECUTION_ERROR", err.Error())
	}
	return payload
}
