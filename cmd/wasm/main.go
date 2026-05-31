//go:build js && wasm

package main

func main() {
	registerBridge()
	select {}
}
