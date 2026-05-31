// embed.go defines code for the server package.

package main

import "embed"

// webFS holds the compiled React + WASM frontend bundle produced by
// make build-web (output: cmd/server/dist/web/). The entire directory tree is
// embedded into the binary at compile time — no files are needed at runtime.
//
// The path is relative to this file: cmd/server/embed.go
// Go's //go:embed does not allow ".." traversal; Vite outputs directly into
// cmd/server/dist/web/ so the path stays within this package directory.
//
//go:embed all:dist/web
var webFS embed.FS
