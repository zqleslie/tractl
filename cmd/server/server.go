// server.go defines code for the server package.

package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/tractl/tractl/internal/version"
)

// statusResponse is returned by GET /api/v1/status.
// The web frontend polls this endpoint on load. A successful response means
// the app is running in connected mode (Web Tier 2) with full engine capability.
type statusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Surface string `json:"surface"`
}

func newMux(static embed.FS) http.Handler {
	mux := http.NewServeMux()

	// ── API routes ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/status", handleStatus)

	// ── Static file server ─────────────────────────────────────────────────
	// fs.Sub strips the "dist/web" prefix from the embedded FS so that
	// dist/web/index.html is served at /, not at /dist/web/index.html.
	stripped, err := fs.Sub(static, "dist/web")
	if err != nil {
		log.Fatalf("embed: failed to sub dist/web: %v", err)
	}

	fileServer := http.FileServer(http.FS(stripped))

	// Any path not matched by an API route falls through to the file server.
	// The React SPA router handles client-side routing — requests for paths
	// like /history or /settings that don't map to real files will get
	// index.html served, which is correct for a SPA.
	mux.Handle("/", fileServer)

	return mux
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statusResponse{
		Status:  "ok",
		Version: version.Version,
		Commit:  version.Commit,
		Surface: "server",
	}); err != nil {
		log.Printf("status encode error: %v", err)
	}
}
