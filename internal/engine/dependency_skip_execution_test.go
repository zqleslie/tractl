// dependency_skip_execution_test.go tests dependency-skip state propagation.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestRun_DependencySkip_OnParentFailure(t *testing.T) {
	// step-1: returns 500 — assertion will fail
	parent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer parent.Close()

	// step-2: depends on step-1 — should be dependency-skipped, not run
	child := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer child.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: dep-skip-flow
    steps:
      - id: step-1
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: s1-status
            kind: status
            op: equals
            expected: 200
      - id: step-2
        kind: request
        dependsOn:
          - step-1
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: s2-status
            kind: status
            op: equals
            expected: 200
`, parent.URL, child.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.Passed {
		t.Fatal("expected Passed=false because step-1 failed")
	}

	states := make(map[string]string)
	for _, wf := range result.Workflows {
		for _, s := range wf.Steps {
			states[s.StepID] = s.State
		}
	}

	if states["step-1"] != "failed" {
		t.Errorf("step-1: want failed, got %s", states["step-1"])
	}
	// step-2 must be skipped (dependency-skipped), not run
	if states["step-2"] == "succeeded" {
		t.Errorf("step-2: should not succeed when dependency failed, got %s", states["step-2"])
	}
	skipped := states["step-2"] == "dependency-skipped" || states["step-2"] == "cancelled"
	if !skipped {
		t.Errorf("step-2: want dependency-skipped or cancelled, got %s", states["step-2"])
	}
}
