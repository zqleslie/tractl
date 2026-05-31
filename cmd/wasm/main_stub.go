//go:build !js

// main_stub.go exists so `go build ./...` can compile cmd/wasm outside a WASM environment.
// The real entry point is main.go, which only builds under js && wasm.

package main

func main() {}
