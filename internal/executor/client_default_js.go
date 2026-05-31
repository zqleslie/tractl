//go:build js

package executor

import "net/http"

// Browser WASM must use the default client so net/http routes through fetch.
func defaultHTTPClient() *http.Client {
	return http.DefaultClient
}
