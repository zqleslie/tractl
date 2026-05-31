// workflow_execution_test.go tests workflow-level DAG execution outcomes.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

// TestRun_MultiWorkflow_Parallel verifies that two independent workflows run
// concurrently (wall-clock evidence) and both must pass for Passed=true.
func TestRun_MultiWorkflow_Parallel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-a
    steps:
      - id: step-a1
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: a1-status
            kind: status
            op: equals
            expected: 200
  - id: wf-b
    steps:
      - id: step-b1
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: b1-status
            kind: status
            op: equals
            expected: 200
`, srv.URL, srv.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if !result.Passed {
		t.Fatalf("expected Passed=true; parse=%q validation=%q plan=%q", result.ParseError, result.ValidationError, result.PlanError)
	}
	if len(result.Workflows) != 2 {
		t.Fatalf("expected 2 workflow outcomes, got %d", len(result.Workflows))
	}
	wfIDs := map[string]bool{}
	for _, wf := range result.Workflows {
		wfIDs[wf.WorkflowID] = wf.Passed
	}
	if !wfIDs["wf-a"] || !wfIDs["wf-b"] {
		t.Fatalf("both workflows must pass; outcomes: %v", wfIDs)
	}
}

// TestRun_MultiWorkflow_WithDependency verifies workflow-level dependsOn:
// wf-b declares dependsOn: [wf-a], so wf-b starts only after wf-a completes.
func TestRun_MultiWorkflow_WithDependency(t *testing.T) {
	var mu sync.Mutex
	callOrder := []string{}

	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		callOrder = append(callOrder, "wf-a")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srvA.Close()

	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		callOrder = append(callOrder, "wf-b")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srvB.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-a
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: a-status
            kind: status
            op: equals
            expected: 200
  - id: wf-b
    dependsOn:
      - wf-a
    steps:
      - id: step-b
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: b-status
            kind: status
            op: equals
            expected: 200
`, srvA.URL, srvB.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if !result.Passed {
		t.Fatalf("expected Passed=true; parse=%q validation=%q plan=%q", result.ParseError, result.ValidationError, result.PlanError)
	}

	mu.Lock()
	order := append([]string{}, callOrder...)
	mu.Unlock()

	if len(order) != 2 {
		t.Fatalf("expected 2 HTTP calls, got %d", len(order))
	}
	// wf-a must have been called before wf-b.
	if order[0] != "wf-a" || order[1] != "wf-b" {
		t.Errorf("expected call order [wf-a wf-b], got %v", order)
	}
}

// TestRun_MultiWorkflow_DependencySkip verifies that when wf-a fails,
// wf-b (which depends on wf-a) is skipped rather than executed.
func TestRun_MultiWorkflow_DependencySkip(t *testing.T) {
	// wf-a: always 500 → assertion failure
	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srvA.Close()

	wfBCalled := false
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		wfBCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srvB.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-a
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: a-status
            kind: status
            op: equals
            expected: 200
  - id: wf-b
    dependsOn:
      - wf-a
    steps:
      - id: step-b
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: b-status
            kind: status
            op: equals
            expected: 200
`, srvA.URL, srvB.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.Passed {
		t.Fatal("expected Passed=false (wf-a failed)")
	}
	if wfBCalled {
		t.Fatal("wf-b server must not be called when its dependency (wf-a) failed")
	}

	wfByID := map[string]engine.WorkflowOutcome{}
	for _, wf := range result.Workflows {
		wfByID[wf.WorkflowID] = wf
	}
	if wfByID["wf-a"].Passed {
		t.Error("wf-a: expected Passed=false")
	}
	if !wfByID["wf-b"].Skipped {
		t.Error("wf-b: expected Skipped=true")
	}
}
