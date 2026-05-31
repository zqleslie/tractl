# traCtl WASM Compatibility Assessment

Status: Appendix
Phase: UI-0.1A
Date: 2026-05-29
Authority: ADR-016 Amendment (2026-05-29), ADR-012 §9 Browser Native Execution

This appendix assesses whether the current traCtl CLI execution pipeline can run inside a browser-hosted Go WebAssembly runtime using `GOOS=js GOARCH=wasm`.

This is an assessment only. It does not define a WASM bridge, browser APIs, React changes, `cmd/wasm`, or build scripts.

## Executive Summary

The current execution pipeline has no hard WASM blockers on the `tractl run` path. The pipeline is pure Go, uses `net/http` for outbound requests, has no `os/exec`, no cgo, no raw socket dependencies on the run path, and uses dependencies that are expected to compile with the standard Go WASM target.

The implementation effort is viable, but it requires targeted modifications before browser execution:

- Replace file-path based workflow and overlay loading with byte/string inputs supplied by the UI.
- Treat HTTP transport diagnostics as best-effort because browser WASM routes Go `net/http` through `fetch`, not native DNS/TCP/TLS hooks.
- Accept browser networking constraints, especially CORS, forbidden headers, restricted response visibility, and lack of custom TLS/proxy control.
- Use the standard Go compiler rather than TinyGo because the current dependency set includes reflection-heavy packages such as `goja`.

Recommendation: Proceed with modifications.

## Execution Pipeline Summary

The CLI starts in `cmd/tractl/main.go`.

1. `cmd/tractl/main.go` parses `tractl run` flags and builds `engine.Config`.
2. `engine.New().Run(cfg)` coordinates the full run.
3. `internal/engine` reads the workflow document using `os.ReadFile(cfg.WorkflowFile)`.
4. `parseByExtension` dispatches by extension to `internal/parser/yaml`, `internal/parser/json`, or `internal/parser/toon`.
5. Optional overlays are loaded with `os.ReadFile`, parsed by `internal/overlay`, and applied to the canonical spec.
6. `internal/validation` performs canonical spec validation.
7. `internal/planner` creates an execution plan, including DAG ordering, capability resolution, runtime selection, concurrency, and failure policy.
8. `internal/compiler` converts the plan and spec into a compiled plan.
9. `internal/engine` runs workflow DAGs concurrently using goroutines, channels, and `errgroup`.
10. `internal/scheduler` runs each workflow step DAG.
11. `internal/executor` builds and executes HTTP requests using `net/http`.
12. `internal/extract` evaluates response extraction rules.
13. `internal/assertion` evaluates compiled assertions.
14. `internal/sandbox` runs JavaScript hooks in `goja` where hooks are present.
15. `internal/diagnostics` and `internal/engine` assemble the final `RunResult`.
16. `engine.FormatText` or `engine.FormatJSON` renders output for the CLI.

## Dependency Inventory

### Internal Packages On The Run Path

- `cmd/tractl`
- `internal/engine`
- `internal/spec`
- `internal/parser/common`
- `internal/parser/yaml`
- `internal/parser/json`
- `internal/parser/toon`
- `internal/overlay`
- `internal/validation`
- `internal/validation/common`
- `internal/validation/yaml`
- `internal/validation/json`
- `internal/validation/toon`
- `internal/planner`
- `internal/compiler`
- `internal/scheduler`
- `internal/executor`
- `internal/runtime`
- `internal/context`
- `internal/sandbox`
- `internal/extract`
- `internal/assertion`
- `internal/diagnostics`
- `internal/provider`

### Internal Packages Not On The CLI Run Path

- `internal/localapi`
- `internal/runtime/http`
- `internal/normalize`
- `internal/normalizer`

`internal/localapi` contains server-side local API code and uses listening sockets. It must not be compiled into the browser WASM target, but it is not part of `engine.Run`.

### External Modules

- `github.com/dop251/goja`
- `github.com/dlclark/regexp2`
- `github.com/go-sourcemap/sourcemap`
- `github.com/google/pprof`
- `github.com/oklog/ulid/v2`
- `github.com/tidwall/gjson`
- `github.com/tidwall/match`
- `github.com/tidwall/pretty`
- `golang.org/x/sync`
- `golang.org/x/text`
- `gopkg.in/yaml.v3`

All dependencies used by the execution path are pure Go. No cgo dependency was identified on the run path.

## WASM Safe Components

These components are expected to compile and execute under `GOOS=js GOARCH=wasm` without meaningful runtime changes:

- `internal/spec`
- `internal/validation`
- `internal/validation/common`
- `internal/validation/yaml`
- `internal/validation/json`
- `internal/validation/toon`
- `internal/compiler`
- `internal/context`
- `internal/assertion`
- `internal/parser/yaml`
- `internal/parser/json`
- `internal/parser/toon`
- `internal/parser/common`
- `internal/overlay`
- `internal/extract`
- `internal/scheduler`
- `internal/diagnostics`
- `internal/provider`

The scheduler and engine DAG orchestration use goroutines, channels, and `errgroup`. These are available under standard Go WASM, but browser execution is effectively single-threaded unless explicitly moved into a worker.

`crypto/rand` is used for ULID generation in `internal/parser/common`, `internal/runtime`, and `internal/planner`. Under `js/wasm`, Go backs random bytes with the browser crypto APIs, so this is not a blocker.

## Review Required Components

### `internal/engine`

`engine.Run` reads the workflow document from `os.ReadFile(cfg.WorkflowFile)`. Overlay files are also loaded with `os.ReadFile`.

This compiles under WASM, but a browser-hosted runtime has no normal local filesystem. Browser execution needs an engine entry point that accepts workflow and overlay source bytes directly, with explicit format metadata.

Required change:

- Add a browser-suitable engine input path that bypasses filesystem reads.

### `internal/executor`

`internal/executor` uses `net/http` with an `http.Client` and an `http.Transport`.

Under `GOOS=js GOARCH=wasm`, Go routes `net/http` through the browser `fetch` API when no custom dialer is configured. The current transport sets tuning fields such as `DisableKeepAlives`, `MaxIdleConns`, and `IdleConnTimeout`, but does not set `Dial`, `DialContext`, `DialTLS`, or `DialTLSContext`. That means the browser `fetch` path should be used.

Review points:

- Transport tuning fields are mostly ignored by browser `fetch`.
- `httptrace` DNS, TCP, TLS, and first-byte callbacks are not reliable under browser fetch.
- Redirect behavior is owned by fetch and may not populate current redirect-hop diagnostics.
- Request cancellation maps to fetch abort behavior.

Required change:

- Treat request timeline fields as best-effort in browser WASM. Preserve total duration, but expect DNS, TCP, TLS, and redirect details to be unavailable.

### `internal/sandbox`

`internal/sandbox` uses `github.com/dop251/goja` for JavaScript hooks.

`goja` is pure Go and is expected to compile with standard Go WASM. However, it is reflection-heavy and increases binary size. This makes TinyGo unsuitable for the current execution engine.

Review points:

- Standard Go WASM should be used.
- Binary size and cold-start time need measurement in UI-0.1B.
- Hook timeout via `time.AfterFunc` should be verified in-browser.

## WASM Blockers

No hard WASM blockers were found on the `cmd/tractl/main.go` to `engine.Run` execution path.

The following blocker-class patterns were not found on the run path:

- `os/exec`
- process spawning
- listening sockets
- raw TCP access
- custom TLS handshakes
- syscall-dependent runtime behavior
- cgo
- platform-specific assembly

Known blocker outside the run path:

- `internal/localapi/server.go` uses `net.Listen` and `http.Server`. This package is not part of the CLI execution pipeline and must be excluded from browser WASM builds.

## Browser Networking Review

The current execution engine assumes outbound HTTP through `net/http` only.

Identified networking behavior:

- Requests are built with `http.NewRequestWithContext`.
- Headers are set through `req.Header.Set`.
- Bodies support JSON and form encoding.
- Execution uses `http.Client.Do`.
- Retry behavior is implemented in `internal/executor`, not by the transport.
- Response body is fully read into memory with `io.ReadAll`.
- Response headers are canonicalized into lowercase map keys.
- HTTP status, headers, body, extracts, assertions, and timing are recorded in runtime results.

Not currently used:

- custom transports requiring raw TCP
- explicit proxy support
- custom TLS configuration
- client certificates
- certificate pinning
- raw socket access
- custom DNS resolution

Browser `fetch` differences:

- CORS applies to all cross-origin requests.
- Forbidden browser headers cannot be set even if the spec asks for them.
- Some response headers are not exposed to JavaScript.
- Cookie handling is browser-controlled.
- TLS certificates, client certificates, proxy routing, and certificate pinning are not controllable from Go.
- Low-level DNS, TCP, and TLS timing is unavailable.
- Some failed fetches collapse into generic network errors.

## Browser Runtime Constraints

### Filesystem

The browser cannot supply normal filesystem paths to `os.ReadFile`. Workflow and overlay documents must be provided as in-memory source.

### CORS

Browser-native execution can only call APIs that permit the page origin through CORS. Non-CORS APIs require a relay/proxy or a desktop/local execution path.

### Headers

The browser may reject or ignore forbidden request headers such as `Host`, `Content-Length`, `Connection`, `Cookie`, and similar browser-controlled headers.

### Response Visibility

Response headers and bodies may be hidden for opaque or restricted cross-origin responses.

### Transport Diagnostics

Browser fetch hides DNS, TCP, TLS, proxy, and connection lifecycle details from Go `httptrace`. Existing diagnostics must degrade gracefully.

### Concurrency

Goroutines work in standard Go WASM, but browser execution is effectively single-threaded. CPU-heavy work can block the UI thread unless the runtime is hosted in a Web Worker.

### Binary Size

The current dependency set likely produces a large standard Go WASM binary. `goja`, reflection, `net/http`, parsers, and diagnostics increase bundle size. Compression and caching are required, and actual size should be measured early.

### Environment

Environment variables are not available as a normal OS facility. Browser execution needs explicit UI-provided environment data.

## Recommendation

Proceed with modifications.

The WASM implementation effort is viable. The current CLI execution pipeline is mostly portable to browser-hosted Go WASM, and no architectural dead-end was found in the engine itself.

Required work before implementation:

1. Add an engine entry point that accepts workflow and overlay source bytes directly.
2. Keep browser-only code outside the engine and avoid importing `internal/localapi`.
3. Keep HTTP execution on `net/http`, but account for fetch semantics.
4. Make transport diagnostics optional or degraded in browser execution.
5. Measure standard Go WASM binary size and load performance.
6. Decide how the UI handles non-CORS endpoints, likely via proxy, desktop execution, or explicit unsupported-state messaging.

The next phase can proceed if UI-0.1B is scoped around these modifications rather than a direct compile of the CLI command.
