// Command localapi runs the traCtl localhost API for web UI development.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tractl/tractl/internal/localapi"
)

func main() {
	root, err := localapi.DefaultRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tractl-localapi: %v\n", err)
		os.Exit(1)
	}

	store, err := localapi.NewStore(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tractl-localapi: %v\n", err)
		os.Exit(1)
	}

	addr := "127.0.0.1:7428"
	server := localapi.NewServer(addr, store)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("tractl local API listening on http://%s\n", addr)
	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "tractl-localapi: %v\n", err)
		os.Exit(1)
	}
}
