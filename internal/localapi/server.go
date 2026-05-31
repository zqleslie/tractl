package localapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const defaultVersion = "0.1.0-alpha"

// Server exposes the traCtl localhost API (ADR-016).
type Server struct {
	addr    string
	store   *Store
	handler http.Handler
	server  *http.Server
}

// NewServer creates a local API server bound to addr (e.g. "127.0.0.1:7428").
func NewServer(addr string, store *Store) *Server {
	s := &Server{addr: addr, store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/files", s.handleFiles)
	mux.HandleFunc("/api/v1/files/{path...}", s.handleFilePath)
	mux.HandleFunc("/api/v1/run", s.handleRun)
	mux.HandleFunc("/api/v1/workflows/run", s.handleWorkflowRun)
	s.handler = withCORS(mux)
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Start listens until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("localapi: listen %s: %w", s.addr, err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
	}()

	if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
