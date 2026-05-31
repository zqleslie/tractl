// Package version exposes build-time metadata injected by the Makefile
// via -ldflags at go build time. Default values are used when the binary
// is built without explicit flags (go run, go test, local development).
package version

// FallbackVersion is the version string used when the binary was built without
// linker-injected version information (e.g. go run, local dev builds).
const FallbackVersion = "0.1.0-alpha"

// Version is the semver string injected at build time via:
//
//	-ldflags "-X github.com/tractl/tractl/internal/version.Version=v1.2.3"
var Version = FallbackVersion

// Commit is the abbreviated git SHA injected at build time.
var Commit = "unknown"

// Date is the UTC build timestamp in RFC 3339 format injected at build time.
var Date = "unknown"

// String returns a single human-readable version line used by
// `tractl version` and window.tractl.version() on the WASM bridge.
func String() string {
	return Version + " (" + Commit + ") built " + Date
}
