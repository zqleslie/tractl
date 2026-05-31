//go:build integration

// integration_test.go exercises the full parse -> validate -> plan -> compile
// -> execute pipeline end-to-end against a real httptest.Server. Run with:
//
//	go test -tags integration ./internal/engine/...
//
// These tests are excluded from go test ./... unless the integration tag is set.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestEngine_FullPipeline_Integration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	specFile := filepath.Join(t.TempDir(), "pipeline_test.yaml")
	if err := os.WriteFile(specFile, []byte(buildMinimalSpec(t, srv.URL)), 0o600); err != nil {
		t.Fatalf("write spec file: %v", err)
	}

	result := engine.New().Run(engine.Config{
		WorkflowFile: specFile,
		Quiet:        true,
		Verbose:      true,
	})

	if result == nil {
		t.Fatal("engine.Run returned nil result")
	}
	if !result.Passed {
		t.Fatalf("expected result.Passed == true, got false; parse=%q validation=%q plan=%q result=%+v",
			result.ParseError, result.ValidationError, result.PlanError, result)
	}
	if len(result.Workflows) != 1 {
		t.Fatalf("expected 1 workflow outcome, got %d", len(result.Workflows))
	}
	if len(result.Workflows[0].Steps) != 1 {
		t.Fatalf("expected 1 step outcome, got %d", len(result.Workflows[0].Steps))
	}

	step := result.Workflows[0].Steps[0]
	if step.ResponseStatus != http.StatusOK {
		t.Fatalf("expected step response status %d, got %d; step=%+v",
			http.StatusOK, step.ResponseStatus, step)
	}
}

func buildMinimalSpec(t *testing.T, baseURL string) string {
	t.Helper()

	return fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: integration-test-workflow
    steps:
      - id: ping-step
        kind: request
        request:
          protocol: http
          target: %q
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`, baseURL+"/ping")
}
