// testhelpers_test.go contains shared helpers for engine external-package tests.

package engine_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeWorkflow writes a YAML string to a temp file and returns its path.
func writeWorkflow(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeWorkflow: %v", err)
	}
	return path
}

// singleStepWorkflow builds a minimal single-step workflow YAML against the given URL.
func singleStepWorkflow(t *testing.T, serverURL, assertOp string, expectedStatus int) string {
	t.Helper()
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
          - id: assert-status
            kind: status
            op: %s
            expected: %d
`, serverURL, assertOp, expectedStatus)
}
