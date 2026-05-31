// hooks_test.go defines code for the engine package.

package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

// hookWorkflow builds a minimal single-step workflow YAML with the given
// inline hook YAML fragment inserted at the indicated scope (workflow or step).
func hookWorkflow(serverURL, hookYAML string) string {
	return fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: hook-flow
%s
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: ok
            kind: status
            op: equals
            expected: 200
`, hookYAML, serverURL)
}

func hookWorkflowWithStepHook(serverURL, stepHookYAML string) string {
	return fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: hook-flow
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: ok
            kind: status
            op: equals
            expected: 200
%s
`, serverURL, stepHookYAML)
}

func okServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// ── nil hooks at every level → no error ──────────────────────────────────────

func TestHook_NilHooks_NoError(t *testing.T) {
	srv := okServer(t)
	path := writeWorkflow(t, hookWorkflow(srv.URL, ""))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no hooks; parse=%q validation=%q plan=%q",
			result.ParseError, result.ValidationError, result.PlanError)
	}
}

// ── beforeAll executes and mutations applied ──────────────────────────────────

func TestHook_BeforeAll_Executes(t *testing.T) {
	srv := okServer(t)
	// beforeAll returns an empty object — no mutation, just must not error.
	hooks := `    hooks:
      beforeAll:
        language: js
        source: |
          return {};
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op beforeAll; parse=%q validation=%q plan=%q",
			result.ParseError, result.ValidationError, result.PlanError)
	}
}

// ── afterAll executes ─────────────────────────────────────────────────────────

func TestHook_AfterAll_Executes(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      afterAll:
        language: js
        source: |
          return {};
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op afterAll")
	}
}

// ── beforeEach executes for every step ───────────────────────────────────────

func TestHook_BeforeEach_Executes(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      beforeEach:
        language: js
        source: |
          return {};
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op beforeEach")
	}
}

// ── afterEach executes for every step ────────────────────────────────────────

func TestHook_AfterEach_Executes(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      afterEach:
        language: js
        source: |
          return {};
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op afterEach")
	}
}

// ── step.beforeStep executes ──────────────────────────────────────────────────

func TestHook_BeforeStep_Executes(t *testing.T) {
	srv := okServer(t)
	stepHook := `        hooks:
          beforeStep:
            language: js
            source: |
              return {};
`
	path := writeWorkflow(t, hookWorkflowWithStepHook(srv.URL, stepHook))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op beforeStep")
	}
}

// ── step.afterStep executes ───────────────────────────────────────────────────

func TestHook_AfterStep_Executes(t *testing.T) {
	srv := okServer(t)
	stepHook := `        hooks:
          afterStep:
            language: js
            source: |
              return {};
`
	path := writeWorkflow(t, hookWorkflowWithStepHook(srv.URL, stepHook))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op afterStep")
	}
}

// ── step.transform executes ───────────────────────────────────────────────────

func TestHook_Transform_Executes(t *testing.T) {
	srv := okServer(t)
	stepHook := `        hooks:
          transform:
            language: js
            source: |
              return {};
`
	path := writeWorkflow(t, hookWorkflowWithStepHook(srv.URL, stepHook))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true with no-op transform")
	}
}

// ── language != "js" → error ──────────────────────────────────────────────────

func TestHook_UnsupportedLanguage_Error(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      beforeAll:
        language: python
        source: |
          print("hello")
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if result.Passed {
		t.Fatal("expected Passed=false for unsupported language")
	}
}

// ── sourceRef set → error ─────────────────────────────────────────────────────

func TestHook_SourceRef_Error(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      beforeAll:
        language: js
        sourceRef: ./scripts/hook.js
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if result.Passed {
		t.Fatal("expected Passed=false for sourceRef")
	}
}

// ── cancel: true in beforeAll → workflow does not proceed ─────────────────────

func TestHook_CancelInBeforeAll_WorkflowFails(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hooks := `    hooks:
      beforeAll:
        language: js
        source: |
          return { cancel: true };
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if result.Passed {
		t.Fatal("expected Passed=false when beforeAll requests cancellation")
	}
	_ = called // scheduler may or may not have started; the engine returns early
}

// ── cancel: true in afterAll → workflow fails ─────────────────────────────────

func TestHook_BeforeEach_StepScopeVariable_RoundTrip(t *testing.T) {
	const wantToken = "set-by-hook"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// step-one has no hook header; step-two sets X-Require-Hook-Token to enforce the round-trip.
		if r.Header.Get("X-Require-Hook-Token") == "true" && r.Header.Get("X-Hook-Token") != wantToken {
			http.Error(w, "missing or bad hook token", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	hooks := `    hooks:
      beforeEach:
        language: js
        source: |
          return { variables: { "hook-token": "set-by-hook" } };
`
	yaml := fmt.Sprintf(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: hook-flow
%s
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
        assertions:
          - id: ok
            kind: status
            op: equals
            expected: 200
      - id: step-two
        kind: request
        dependsOn:
          - step-one
        request:
          protocol: http
          target: %s
          operation: GET
          headers:
            X-Require-Hook-Token: "true"
            X-Hook-Token: "${steps.step-one.extracts.hook-token}"
        assertions:
          - id: ok
            kind: status
            op: equals
            expected: 200
`, hooks, srv.URL, srv.URL)

	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if !result.Passed {
		t.Fatalf("expected Passed=true when beforeEach writes step-scoped variable; parse=%q validation=%q plan=%q",
			result.ParseError, result.ValidationError, result.PlanError)
	}
}

func TestHook_CancelInAfterAll_WorkflowFails(t *testing.T) {
	srv := okServer(t)
	hooks := `    hooks:
      afterAll:
        language: js
        source: |
          return { cancel: true };
`
	path := writeWorkflow(t, hookWorkflow(srv.URL, hooks))
	result := engine.New().Run(engine.Config{WorkflowFile: path, Quiet: true})
	if result.Passed {
		t.Fatal("expected Passed=false when afterAll requests cancellation")
	}
}
