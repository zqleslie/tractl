// main_test.go defines code for the tractl package.

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// run is a backward-compat test helper preserving the pre-refactor run(args) signature.
func run(args []string) int {
	return dispatch(map[string]Command{"run": &RunCommand{}}, args)
}

// writeWF writes a minimal YAML workflow to a temp file and returns its path.
func writeWF(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "wf.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func singleStepYAML(url string, expectedStatus int) string {
	return fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: test-flow
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: status-check
            kind: status
            op: equals
            expected: %d
`, url, expectedStatus)
}

// ── no subcommand ─────────────────────────────────────────────────────────────

func TestRun_NoArgs_Returns2(t *testing.T) {
	if got := run([]string{}); got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}

func TestRun_WrongSubcommand_Returns2(t *testing.T) {
	if got := run([]string{"exec"}); got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}

// ── run subcommand: missing workflow file ─────────────────────────────────────

func TestRun_MissingWorkflowFile_Returns2(t *testing.T) {
	if got := run([]string{"run"}); got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}

// ── run subcommand: non-existent file ────────────────────────────────────────

func TestRun_NonExistentFile_Returns2(t *testing.T) {
	if got := run([]string{"run", "/tmp/this-file-does-not-exist-tractl.yaml"}); got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}

// ── happy path: exit 0 ───────────────────────────────────────────────────────

func TestRun_HappyPath_Returns0(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := writeWF(t, singleStepYAML(srv.URL, 200))
	if got := run([]string{"run", path}); got != 0 {
		t.Fatalf("expected exit code 0, got %d", got)
	}
}

// ── assertion failure: exit 1 ────────────────────────────────────────────────

func TestRun_AssertionFailure_Returns1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	path := writeWF(t, singleStepYAML(srv.URL, 200))
	if got := run([]string{"run", path}); got != 1 {
		t.Fatalf("expected exit code 1, got %d", got)
	}
}

// ── --output json flag ────────────────────────────────────────────────────────

func TestRun_OutputJSON_Returns0(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := writeWF(t, singleStepYAML(srv.URL, 200))
	if got := run([]string{"run", "--output", "json", path}); got != 0 {
		t.Fatalf("expected exit code 0, got %d", got)
	}
}

// ── unknown flag ─────────────────────────────────────────────────────────────

func TestRun_UnknownFlag_Returns2(t *testing.T) {
	if got := run([]string{"run", "--notaflag", "wf.yaml"}); got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}
