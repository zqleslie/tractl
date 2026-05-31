// dependency_execution_test.go tests branch and dependency failure propagation.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

// TestRun_MultiWorkflow_IndependentContinuesOnSiblingFailure verifies resilient
// behaviour: wf-b fails, but wf-c (independent of wf-b) still executes and passes.
func TestRun_MultiWorkflow_IndependentContinuesOnSiblingFailure(t *testing.T) {
	srvFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srvFail.Close()

	srvOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srvOK.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-b
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
  - id: wf-c
    steps:
      - id: step-c
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: c-status
            kind: status
            op: equals
            expected: 200
`, srvFail.URL, srvOK.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.Passed {
		t.Fatal("expected Passed=false (wf-b failed)")
	}

	wfByID := map[string]engine.WorkflowOutcome{}
	for _, wf := range result.Workflows {
		wfByID[wf.WorkflowID] = wf
	}
	if wfByID["wf-b"].Passed {
		t.Error("wf-b: expected Passed=false")
	}
	// wf-c has no dependency on wf-b — must still run and pass.
	if !wfByID["wf-c"].Passed {
		t.Errorf("wf-c: expected Passed=true (independent of wf-b), got Passed=false Skipped=%v", wfByID["wf-c"].Skipped)
	}
}

func TestRun_ParallelBranchFailure_ResilientPolicy(t *testing.T) {
	// step-a: always 200 — independent
	stepA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer stepA.Close()

	// step-b: always 500 — independent, will fail assertion
	stepB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer stepB.Close()

	// step-c: depends on step-a only — should still run and succeed
	stepC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer stepC.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: resilient-flow
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
      - id: step-c
        kind: request
        dependsOn:
          - step-a
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: c-status
            kind: status
            op: equals
            expected: 200
`, stepA.URL, stepB.URL, stepC.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	// overall: FAIL because step-b failed
	if result.Passed {
		t.Fatal("expected Passed=false")
	}

	states := make(map[string]string)
	causes := make(map[string]bool)
	for _, wf := range result.Workflows {
		for _, s := range wf.Steps {
			states[s.StepID] = s.State
			causes[s.StepID] = s.CausesFailure
		}
	}

	if states["step-a"] != "succeeded" {
		t.Errorf("step-a: want succeeded, got %s", states["step-a"])
	}
	if !causes["step-b"] {
		t.Errorf("step-b: want CausesFailure=true")
	}
	// step-c depends only on step-a (which succeeded) — must still run and pass
	if states["step-c"] != "succeeded" {
		t.Errorf("step-c: want succeeded, got %s (resilient policy must not cancel independent branches)", states["step-c"])
	}
}

func TestRun_FailFast_DownstreamSkipped(t *testing.T) {
	// step-root: fails (500)
	root := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer root.Close()

	// step-child: depends on root — should be dependency-skipped or cancelled
	child := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// this server must NOT be called
		w.WriteHeader(http.StatusOK)
	}))
	defer child.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: failfast-flow
    failurePolicy: failFast
    steps:
      - id: step-root
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: root-status
            kind: status
            op: equals
            expected: 200
      - id: step-child
        kind: request
        dependsOn:
          - step-root
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: child-status
            kind: status
            op: equals
            expected: 200
`, root.URL, child.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.Passed {
		t.Fatal("expected Passed=false")
	}

	states := make(map[string]string)
	for _, wf := range result.Workflows {
		for _, s := range wf.Steps {
			states[s.StepID] = s.State
		}
	}

	if states["step-root"] != "failed" {
		t.Errorf("step-root: want failed, got %s", states["step-root"])
	}
	// child must not have succeeded — it should be skipped or cancelled, never "succeeded"
	if states["step-child"] == "succeeded" {
		t.Errorf("step-child: should not succeed when parent failed under failFast, got %s", states["step-child"])
	}
}
