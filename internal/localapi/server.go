package localapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tractl/tractl/internal/engine"
	"github.com/tractl/tractl/internal/spec"
	"github.com/tractl/tractl/internal/validation"
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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedLocalAPIOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedLocalAPIOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme == "wails" && parsed.Hostname() == "wails.localhost" {
		return true
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := parsed.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, StatusResponse{OK: true, Version: defaultVersion})
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req SaveFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json", Details: err.Error()})
		return
	}
	if len(req.Document) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "document required"})
		return
	}

	specDoc, err := decodeSpec(req.Document)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid document", Details: err.Error()})
		return
	}

	validationErrs := validation.NewSpecValidator().Validate(specDoc)
	if len(validationErrs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "validation failed",
			Details: joinValidationErrors(validationErrs),
		})
		return
	}

	yamlBytes, err := MarshalDocumentYAML(req.Document)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "cannot marshal yaml", Details: err.Error()})
		return
	}

	kind := req.Kind
	if kind == "" {
		kind = "request"
	}

	record, err := s.store.SaveYAML(req.ID, kind, req.Name, yamlBytes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "save failed", Details: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SaveFileResponse{
		ID:        record.ID,
		Name:      record.Name,
		Path:      record.Path,
		UpdatedAt: record.UpdatedAt,
	})
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json", Details: err.Error()})
		return
	}
	if strings.TrimSpace(req.FileID) == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "fileId required"})
		return
	}

	path, err := s.store.PathForID(req.FileID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "file not found", Details: err.Error()})
		return
	}

	result := engine.New().Run(engine.Config{
		WorkflowFile: path,
		EnvName:      req.Env,
		Verbose:      true,
		Quiet:        true,
	})

	mapped, err := mapRunResult(result)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: "run failed", Details: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, mapped)
}

func (s *Server) handleWorkflowRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req WorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json", Details: err.Error()})
		return
	}
	if strings.TrimSpace(req.Document) == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "document required"})
		return
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "yaml"
	}

	result := engine.New().RunDocument(engine.DocumentConfig{
		Document:  req.Document,
		Format:    format,
		SourceRef: "desktop-local-api",
		EnvName:   req.Env,
		Quiet:     true,
		Verbose:   true,
	})

	writeJSON(w, http.StatusOK, result)
}

func decodeSpec(document json.RawMessage) (*spec.TraCtlSpec, error) {
	var s spec.TraCtlSpec
	if err := json.Unmarshal(document, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func joinValidationErrors(errs []validation.ValidationError) string {
	if len(errs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(errs))
	for _, e := range errs {
		parts = append(parts, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(parts, "; ")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
