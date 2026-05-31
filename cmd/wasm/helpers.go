//go:build js && wasm

// helpers.go contains bridge helpers that depend on syscall/js.
// Pure helpers without a JS dependency live in helpers_pure.go and can be tested normally.

package main

import (
	"fmt"
	"syscall/js"
)

// jsHandler is the function signature shared by all bridge handler implementations.
type jsHandler func(this js.Value, args []js.Value) any

// recoveringHandler wraps a synchronous handler so any panic becomes a structured bridge error.
func recoveringHandler(handler jsHandler) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) (result any) {
		defer func() {
			if recovered := recover(); recovered != nil {
				result = bridgeError("TRACTL_WASM_PANIC", fmt.Sprint(recovered))
			}
		}()

		return handler(this, args)
	}
}

// recoveringPromiseHandler wraps a promise-returning handler so any panic resolves as a bridge error.
func recoveringPromiseHandler(handler jsHandler) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		return newPromise(func(resolve js.Value) {
			defer func() {
				if recovered := recover(); recovered != nil {
					resolve.Invoke(bridgeError("TRACTL_WASM_PANIC", fmt.Sprint(recovered)))
				}
			}()

			resolve.Invoke(handler(this, args))
		})
	}
}

// asyncPromiseHandler runs the handler in a new goroutine so blocking net/http
// (fetch) does not stall the syscall/js callback thread. See golang/go#41310.
func asyncPromiseHandler(handler jsHandler) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		return newPromise(func(resolve js.Value) {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						resolve.Invoke(bridgeError("TRACTL_WASM_PANIC", fmt.Sprint(recovered)))
					}
				}()
				resolve.Invoke(handler(this, args))
			}()
		})
	}
}

// newPromise creates a JS Promise that resolves with the value returned by execute.
func newPromise(execute func(resolve js.Value)) js.Value {
	executor := js.FuncOf(func(this js.Value, args []js.Value) any {
		execute(args[0])
		return nil
	})
	promise := js.Global().Get("Promise").New(executor)
	executor.Release()
	return promise
}
