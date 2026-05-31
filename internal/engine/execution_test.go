// execution_test.go tests direct Engine.Run and RunDocument execution outcomes.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestRun_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := writeWorkflow(t, singleStepWorkflow(t, srv.URL, "equals", 200))
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if !result.Passed {
		t.Fatalf("expected Passed=true, got false; validation=%q plan=%q parse=%q",
			result.ValidationError, result.PlanError, result.ParseError)
	}
	if result.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode())
	}
	if len(result.Workflows) != 1 {
		t.Fatalf("expected 1 workflow outcome, got %d", len(result.Workflows))
	}
	if len(result.Workflows[0].Steps) != 1 {
		t.Fatalf("expected 1 step outcome, got %d", len(result.Workflows[0].Steps))
	}
}

func TestRunDocument_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	doc := singleStepWorkflow(t, srv.URL, "equals", 200)
	result := engine.New().RunDocument(engine.DocumentConfig{
		Document:  doc,
		Format:    "yaml",
		SourceRef: "test-document",
	})

	if !result.Passed {
		t.Fatalf("expected Passed=true, got false; validation=%q plan=%q parse=%q",
			result.ValidationError, result.PlanError, result.ParseError)
	}
}

func TestRun_AssertionFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound) // 404
	}))
	defer srv.Close()

	// Expect 200 but server returns 404.
	path := writeWorkflow(t, singleStepWorkflow(t, srv.URL, "equals", 200))
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.Passed {
		t.Fatal("expected Passed=false, got true")
	}
	if result.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode())
	}
	if len(result.Workflows) == 0 {
		t.Fatal("expected workflow outcomes")
	}
	step := result.Workflows[0].Steps[0]
	if !step.CausesFailure {
		t.Fatal("expected step.CausesFailure=true")
	}
}

func TestRun_WarningSeverityDoesNotFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound) // 404
	}))
	defer srv.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: warn-flow
    steps:
      - id: step-warn
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: warn-assert
            kind: status
            op: equals
            expected: 200
            severity: warning
`, srv.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if !result.Passed {
		t.Fatal("expected Passed=true for warning-only failure")
	}
	if result.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode())
	}
	// The step itself should not be marked as causing failure.
	if len(result.Workflows[0].Steps) > 0 && result.Workflows[0].Steps[0].CausesFailure {
		t.Fatal("warning assertion must not cause step failure")
	}
}

func TestRun_ExtractAndDownstream(t *testing.T) {
	const tokenValue = "abc-token-123"

	// Step A: returns JSON body with a token field.
	stepA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token":"%s"}`, tokenValue)
	}))
	defer stepA.Close()

	// Step B: verifies the injected token arrived in the Authorization header.
	var capturedAuth string
	stepB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("X-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer stepB.Close()

	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: extract-flow
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        extracts:
          - id: token-extract
            source: body
            path: token
            as: mytoken
            scope: workflow
      - id: step-b
        kind: request
        dependsOn:
          - step-a
        request:
          protocol: http
          target: %s
          operation: GET
          headers:
            X-Token: ${vars.mytoken}
        assertions:
          - id: status-ok
            kind: status
            op: equals
            expected: 200
`, stepA.URL, stepB.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if !result.Passed {
		for _, wf := range result.Workflows {
			for _, s := range wf.Steps {
				t.Logf("step %s state=%s causesFailure=%v err=%q", s.StepID, s.State, s.CausesFailure, s.Error)
				for _, ar := range s.AssertionResults {
					t.Logf("  assertion %s: %s", ar.AssertionID, ar.Message)
				}
			}
		}
		t.Fatalf("expected Passed=true, got false; parse=%q validation=%q plan=%q",
			result.ParseError, result.ValidationError, result.PlanError)
	}

	if capturedAuth != tokenValue {
		t.Fatalf("expected X-Token header %q in step B, got %q", tokenValue, capturedAuth)
	}
}

func TestExitCode_Passed(t *testing.T) {
	r := &engine.RunResult{Passed: true}
	if r.ExitCode() != 0 {
		t.Fatalf("want 0, got %d", r.ExitCode())
	}
}

func TestExitCode_AssertionFailure(t *testing.T) {
	r := &engine.RunResult{Passed: false}
	if r.ExitCode() != 1 {
		t.Fatalf("want 1, got %d", r.ExitCode())
	}
}

func TestExitCode_ParseError(t *testing.T) {
	r := &engine.RunResult{ParseError: "bad yaml"}
	if r.ExitCode() != 2 {
		t.Fatalf("want 2, got %d", r.ExitCode())
	}
}

func TestExitCode_ValidationError(t *testing.T) {
	r := &engine.RunResult{ValidationError: "invalid schema"}
	if r.ExitCode() != 2 {
		t.Fatalf("want 2, got %d", r.ExitCode())
	}
}

func TestExitCode_PlanError(t *testing.T) {
	r := &engine.RunResult{PlanError: "no capability"}
	if r.ExitCode() != 2 {
		t.Fatalf("want 2, got %d", r.ExitCode())
	}
}
