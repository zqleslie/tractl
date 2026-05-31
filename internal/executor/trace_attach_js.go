//go:build js

package executor

import "context"

// Browser fetch does not expose DNS/TCP/TLS hooks; skip httptrace on js/wasm.
// ADR-012 §11: dns/tcp/tls fields are declared exceptions for the Web Tier 1 surface.
func attachHTTPTrace(ctx context.Context, _ *TimingRecorder) context.Context {
	return ctx
}
