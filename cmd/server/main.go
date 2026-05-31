// main.go defines code for the server package.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tractl/tractl/internal/version"
)

func main() {
	addr := os.Getenv("TRACTL_ADDR")
	if addr == "" {
		addr = ":7428"
	}

	mux := newMux(webFS)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("traCtl server %s\n", version.String())
	fmt.Printf("Web UI → http://localhost%s\n", addr)
	fmt.Printf("API    → http://localhost%s/api/v1/status\n", addr)

	log.Fatal(srv.ListenAndServe())
}
