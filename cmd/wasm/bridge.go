//go:build js && wasm

// bridge.go registers the tractl WASM bridge functions with the JS runtime.
// Handler implementations live in handlers.go; utilities live in helpers.go and helpers_pure.go.

package main

import "syscall/js"

var exportedHandlers []js.Func

// registerBridge wires all handler functions into the JS global tractl object.
func registerBridge() {
	tractl := js.Global().Get("Object").New()

	versionFunc := js.FuncOf(recoveringHandler(handleVersion))
	capabilitiesFunc := js.FuncOf(recoveringHandler(handleCapabilities))
	echoFunc := js.FuncOf(recoveringPromiseHandler(handleEcho))
	parseFunc := js.FuncOf(recoveringPromiseHandler(handleParse))
	validateFunc := js.FuncOf(recoveringPromiseHandler(handleValidate))
	runFunc := js.FuncOf(asyncPromiseHandler(handleRun))
	runRequestFunc := js.FuncOf(asyncPromiseHandler(handleRunRequest))
	layoutFunc := js.FuncOf(recoveringPromiseHandler(handleLayout))
	inferDepsFunc := js.FuncOf(recoveringPromiseHandler(handleInferDeps))
	exportWorkflowFunc := js.FuncOf(recoveringPromiseHandler(handleExportWorkflow))
	defaultsFunc := js.FuncOf(recoveringHandler(handleDefaults))

	exportedHandlers = append(exportedHandlers,
		versionFunc, capabilitiesFunc, echoFunc, parseFunc, validateFunc, runFunc,
		runRequestFunc, layoutFunc, inferDepsFunc, exportWorkflowFunc, defaultsFunc,
	)
	tractl.Set("version", versionFunc)
	tractl.Set("capabilities", capabilitiesFunc)
	tractl.Set("echo", echoFunc)
	tractl.Set("parse", parseFunc)
	tractl.Set("validate", validateFunc)
	tractl.Set("run", runFunc)
	tractl.Set("runRequest", runRequestFunc)
	tractl.Set("layout", layoutFunc)
	tractl.Set("inferDeps", inferDepsFunc)
	tractl.Set("exportWorkflow", exportWorkflowFunc)
	tractl.Set("defaults", defaultsFunc)

	js.Global().Set("tractl", tractl)
}
