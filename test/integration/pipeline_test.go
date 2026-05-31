//go:build integration

// Package integration_test wires multiple traCtl pipeline stages together
// without real HTTP. Each test exercises the canonical pipeline from
// *spec.TraCtlSpec (or raw YAML/JSON/TOON) through to *scheduler.WorkflowResult.
//
// Run with: go test ./test/integration/... -tags integration
package integration_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/normalize"
	"github.com/tractl/tractl/internal/overlay"
	jsonparser "github.com/tractl/tractl/internal/parser/json"
	toonparser "github.com/tractl/tractl/internal/parser/toon"
	yamlparser "github.com/tractl/tractl/internal/parser/yaml"
	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
	"github.com/tractl/tractl/internal/validation"
)

// mockExecutor implements executor.Executor for integration tests without real HTTP.
type mockExecutor struct {
	results map[string]*runtime.StepResult
}

func (m *mockExecutor) Execute(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error) {
	if r, ok := m.results[step.StepID]; ok {
		return r, nil
	}
	return &runtime.StepResult{StepID: step.StepID, WorkflowID: step.WorkflowID, State: runtime.StateSucceeded, StartedAt: time.Now(), FinishedAt: time.Now(), Extracts: map[string]string{}}, nil
}

// buildPipeline runs the full canonical pipeline up to (but not including)
// scheduler.Run. It returns the compiled plan or an error from any stage.
func buildPipeline(s *spec.TraCtlSpec) (*compiler.CompiledPlan, error) {
	norm, err := normalize.Normalize(s)
	if err != nil {
		return nil, err
	}
	if errs := validation.NewSpecValidator().Validate(norm); len(errs) > 0 {
		return nil, &validationErr{errs: errs}
	}
	p, err := planner.NewPlanner(planner.DefaultRegistry()).Plan(norm)
	if err != nil {
		return nil, err
	}
	return compiler.NewCompiler().Compile(p, norm)
}

func TestIntegration_ParseToPlan_YAML(t *testing.T) {
	const yamlInput = `
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wfA
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: http://example.local/a
      - id: step-b
        kind: request
        dependsOn:
          - step-a
        request:
          protocol: http
          target: http://example.local/b
`
	s, err := yamlparser.Parse([]byte(yamlInput), "test.yaml")
	if err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if len(cp.Workflows[0].Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(cp.Workflows[0].Steps))
	}
	if cp.Workflows[0].Steps[0].StepID != "step-a" {
		t.Fatalf("expected first step step-a, got %s", cp.Workflows[0].Steps[0].StepID)
	}
	if cp.Workflows[0].Steps[1].StepID != "step-b" {
		t.Fatalf("expected second step step-b, got %s", cp.Workflows[0].Steps[1].StepID)
	}
	if !reflect.DeepEqual(cp.Workflows[0].Steps[1].DependsOn, []string{"step-a"}) {
		t.Fatalf("expected step-b.DependsOn=[step-a], got %v", cp.Workflows[0].Steps[1].DependsOn)
	}
}

func TestIntegration_ParseToPlan_JSON(t *testing.T) {
	const jsonInput = `{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {"id": "wfA", "steps": [
      {"id": "step-a", "kind": "request", "request": {"protocol": "http", "target": "http://example.local/a"}},
      {"id": "step-b", "kind": "request", "dependsOn": ["step-a"], "request": {"protocol": "http", "target": "http://example.local/b"}}
    ]}
  ]
}`
	s, err := jsonparser.Parse([]byte(jsonInput), "test.json")
	if err != nil {
		t.Fatalf("json parse: %v", err)
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if cp.Workflows[0].Steps[0].StepID != "step-a" || cp.Workflows[0].Steps[1].StepID != "step-b" {
		t.Fatalf("unexpected step order")
	}
}

func TestIntegration_ParseToPlan_TOON(t *testing.T) {
	const toonInput = `
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wfA
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: http://example.local/a
      - id: step-b
        kind: request
        dependsOn:
          - step-a
        request:
          protocol: http
          target: http://example.local/b
`
	s, err := toonparser.Parse([]byte(toonInput), "test.toon")
	if err != nil {
		t.Fatalf("toon parse: %v", err)
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if cp.Workflows[0].Steps[0].StepID != "step-a" || cp.Workflows[0].Steps[1].StepID != "step-b" {
		t.Fatalf("unexpected step order")
	}
}

func TestIntegration_OverlayApplied(t *testing.T) {
	const yamlInput = `
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wfA
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: http://example.local/a
`
	s, err := yamlparser.Parse([]byte(yamlInput), "base.yaml")
	if err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	// Apply an overlay that sets timeout on step-a using semantic match by ID.
	patches := []*overlay.OverlayDocument{{
		Patches: []overlay.Patch{{
			Target: overlay.Target{Match: overlay.MatchSelector{"id": "step-a"}, Mode: overlay.TargetModeOne},
			Action: overlay.ActionDeepMerge,
			Data:   map[string]any{"timeout": "PT30S"},
		}},
	}}
	engine := overlay.NewEngine()
	patched, err := engine.Apply(s, patches)
	if err != nil {
		t.Fatalf("overlay: %v", err)
	}
	// Convert back to *spec.TraCtlSpec via JSON round-trip.
	out, err := overlayToSpec(patched)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	cp, err := buildPipeline(out)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if cp.Workflows[0].Steps[0].Timeout != "PT30S" {
		t.Fatalf("expected timeout PT30S, got %q", cp.Workflows[0].Steps[0].Timeout)
	}
	// Base spec must be unchanged.
	if s.Workflows[0].Steps[0].Timeout != "" {
		t.Fatalf("base spec mutated: timeout=%q", s.Workflows[0].Steps[0].Timeout)
	}
}

func TestIntegration_ImplicitDependencyMerged(t *testing.T) {
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{
			ID: "wfA",
			Steps: []spec.Step{
				{ID: "login", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://example.local/auth"}},
				{ID: "fetch-profile", Kind: "request", Request: &spec.RequestDescriptor{
					Protocol: "http",
					Target:   "http://example.local/profile",
					Headers:  map[string]string{"Authorization": "Bearer ${steps.login.extracts.token}"},
				}},
			},
		}},
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	// fetch-profile must have implicit dep on login.
	if !reflect.DeepEqual(cp.Workflows[0].Steps[1].DependsOn, []string{"login"}) {
		t.Fatalf("expected fetch-profile.DependsOn=[login], got %v", cp.Workflows[0].Steps[1].DependsOn)
	}
	// Topo order: login before fetch-profile.
	if cp.Workflows[0].Steps[0].StepID != "login" || cp.Workflows[0].Steps[1].StepID != "fetch-profile" {
		t.Fatalf("topo order wrong: %v", []string{cp.Workflows[0].Steps[0].StepID, cp.Workflows[0].Steps[1].StepID})
	}
}

func TestIntegration_CycleRejectedAtValidator(t *testing.T) {
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{
			ID: "wfA",
			Steps: []spec.Step{
				{ID: "step-a", Kind: "request", DependsOn: []string{"step-b"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/a"}},
				{ID: "step-b", Kind: "request", DependsOn: []string{"step-a"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/b"}},
			},
		}},
	}
	_, err := buildPipeline(s)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "cycle") {
		t.Fatalf("expected cycle error from validator, got %v", err)
	}
}

func TestIntegration_UnresolvedDependsOnRejected(t *testing.T) {
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{
			ID: "wfA",
			Steps: []spec.Step{
				{ID: "step-b", Kind: "request", DependsOn: []string{"nonexistent-step"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/b"}},
			},
		}},
	}
	_, err := buildPipeline(s)
	if err == nil {
		t.Fatalf("expected error for unresolved dep")
	}
	if !strings.Contains(err.Error(), "nonexistent-step") {
		t.Fatalf("expected error message to mention 'nonexistent-step', got %v", err)
	}
}

func TestIntegration_FullPipeline_LinearWorkflow(t *testing.T) {
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{
			ID: "wfA",
			Steps: []spec.Step{
				{ID: "A", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/A"}},
				{ID: "B", Kind: "request", DependsOn: []string{"A"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/B"}},
				{ID: "C", Kind: "request", DependsOn: []string{"B"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/C"}},
			},
		}},
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	mock := &mockExecutor{}
	sched := scheduler.NewScheduler(cp, mock)
	res, err := sched.Run(context.Background(), "wfA")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", res.OverallState)
	}
	if len(res.Outcomes) != 3 {
		t.Fatalf("expected 3 outcomes, got %d", len(res.Outcomes))
	}
	// Verify step order: A, B, C
	wantOrder := []string{"A", "B", "C"}
	for i, want := range wantOrder {
		if res.Outcomes[i].StepID != want {
			t.Fatalf("outcome[%d]: want %s, got %s", i, want, res.Outcomes[i].StepID)
		}
	}
}

func TestIntegration_FullPipeline_ResilienceOnFailure(t *testing.T) {
	// A → B → C; D independent. B fails.
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{
			ID:            "wfA",
			Concurrency:   5,
			FailurePolicy: spec.FailurePolicyResilient,
			Steps: []spec.Step{
				{ID: "A", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/A"}},
				{ID: "B", Kind: "request", DependsOn: []string{"A"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/B"}},
				{ID: "C", Kind: "request", DependsOn: []string{"B"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/C"}},
				{ID: "D", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x/D"}},
			},
		}},
	}
	cp, err := buildPipeline(s)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	mock := &mockExecutor{results: map[string]*runtime.StepResult{
		"B": {StepID: "B", WorkflowID: "wfA", State: runtime.StateFailed, StartedAt: time.Now(), FinishedAt: time.Now(), Extracts: map[string]string{}},
	}}
	sched := scheduler.NewScheduler(cp, mock)
	res, err := sched.Run(context.Background(), "wfA")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	states := map[string]runtime.StepState{}
	for _, o := range res.Outcomes {
		states[o.StepID] = o.State
	}
	if states["A"] != runtime.StateSucceeded {
		t.Fatalf("A: expected succeeded, got %s", states["A"])
	}
	if states["B"] != runtime.StateFailed {
		t.Fatalf("B: expected failed, got %s", states["B"])
	}
	if states["C"] != runtime.StateDependencySkipped {
		t.Fatalf("C: expected dependency-skipped, got %s", states["C"])
	}
	if states["D"] != runtime.StateSucceeded {
		t.Fatalf("D: expected succeeded (independent), got %s", states["D"])
	}
	if res.OverallState != runtime.StateFailed {
		t.Fatalf("expected overall failed, got %s", res.OverallState)
	}
}

// overlayToSpec converts the overlay engine's any-typed output back into a
// *spec.TraCtlSpec via JSON round-trip.
func overlayToSpec(v any) (*spec.TraCtlSpec, error) {
	out, err := overlayMarshalUnmarshal(v)
	if err != nil {
		return nil, err
	}
	return out, nil
}
