package compiler

import (
	"strings"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/spec"
)

// validPlan builds a valid ExecutionPlan with n steps for use in tests.
func validPlan(n int) *planner.ExecutionPlan {
	steps := make([]planner.StepPlan, n)
	for i := range steps {
		id := string(rune('A' + i))
		steps[i] = planner.StepPlan{
			StepID:          "step-" + id,
			WorkflowID:      "wf-1",
			Runtime:         "http",
			FailurePolicy:   "resilient",
			DependsOn:       []string{},
			ConcurrencySlot: i + 1,
			CanSkip:         i > 0,
		}
	}
	return &planner.ExecutionPlan{
		PlanID: "plan-001",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID:       "wf-1",
				Steps:            steps,
				ConcurrencyLimit: 4,
				FailurePolicy:    "resilient",
			},
		},
	}
}

// helper to build a minimal spec matching a plan
func specFromPlan(plan *planner.ExecutionPlan) *spec.TraCtlSpec {
	if plan == nil || len(plan.Workflows) == 0 {
		return &spec.TraCtlSpec{Workflows: nil}
	}
	wfs := make([]spec.Workflow, 0, len(plan.Workflows))
	for _, wp := range plan.Workflows {
		st := make([]spec.Step, 0, len(wp.Steps))
		for _, sp := range wp.Steps {
			st = append(st, spec.Step{ID: sp.StepID})
		}
		wfs = append(wfs, spec.Workflow{ID: wp.WorkflowID, Steps: st})
	}
	return &spec.TraCtlSpec{Workflows: wfs}
}

func TestCompile_HappyPath(t *testing.T) {
	plan := validPlan(3)
	c := NewCompiler()
	compiled, err := c.Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(compiled.Workflows[0].Steps) != 3 {
		t.Errorf("want 3 steps, got %d", len(compiled.Workflows[0].Steps))
	}
}

func TestCompile_PlanIDInherited(t *testing.T) {
	plan := validPlan(1)
	plan.PlanID = "my-plan-id-xyz"
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compiled.PlanID != "my-plan-id-xyz" {
		t.Errorf("want PlanID %q, got %q", "my-plan-id-xyz", compiled.PlanID)
	}
}

func TestCompile_CompiledAtPopulated(t *testing.T) {
	plan := validPlan(1)
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compiled.CompiledAt == "" {
		t.Fatal("CompiledAt is empty")
	}
	if _, err := time.Parse(time.RFC3339, compiled.CompiledAt); err != nil {
		t.Errorf("CompiledAt %q does not parse as RFC3339: %v", compiled.CompiledAt, err)
	}
}

func TestCompile_StepFieldsCopied(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID:       "wf-1",
				ConcurrencyLimit: 2,
				FailurePolicy:    "failFast",
				Steps: []planner.StepPlan{
					{
						StepID:          "s1",
						WorkflowID:      "wf-1",
						Runtime:         "http",
						FailurePolicy:   "resilient",
						DependsOn:       []string{"s0"},
						ConcurrencySlot: 3,
						CanSkip:         true,
					},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cs := compiled.Workflows[0].Steps[0]
	sp := plan.Workflows[0].Steps[0]

	if cs.StepID != sp.StepID {
		t.Errorf("StepID: want %q got %q", sp.StepID, cs.StepID)
	}
	if cs.WorkflowID != sp.WorkflowID {
		t.Errorf("WorkflowID: want %q got %q", sp.WorkflowID, cs.WorkflowID)
	}
	if cs.Runtime != sp.Runtime {
		t.Errorf("Runtime: want %q got %q", sp.Runtime, cs.Runtime)
	}
	if cs.FailurePolicy != sp.FailurePolicy {
		t.Errorf("FailurePolicy: want %q got %q", sp.FailurePolicy, cs.FailurePolicy)
	}
	if len(cs.DependsOn) != 1 || cs.DependsOn[0] != "s0" {
		t.Errorf("DependsOn: want [s0] got %v", cs.DependsOn)
	}
	if cs.ConcurrencySlot != sp.ConcurrencySlot {
		t.Errorf("ConcurrencySlot: want %d got %d", sp.ConcurrencySlot, cs.ConcurrencySlot)
	}
	if cs.CanSkip != sp.CanSkip {
		t.Errorf("CanSkip: want %v got %v", sp.CanSkip, cs.CanSkip)
	}
}

func TestCompile_DependsOnIsIndependentCopy(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: "http", DependsOn: []string{"s0"}},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	compiled.Workflows[0].Steps[0].DependsOn[0] = "MUTATED"
	if plan.Workflows[0].Steps[0].DependsOn[0] != "s0" {
		t.Error("mutating CompiledStep.DependsOn affected the source StepPlan.DependsOn")
	}
}

func TestCompile_WorkflowOrderPreserved(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{WorkflowID: "alpha", Steps: []planner.StepPlan{{StepID: "s1", WorkflowID: "alpha", Runtime: "http"}}},
			{WorkflowID: "beta", Steps: []planner.StepPlan{{StepID: "s2", WorkflowID: "beta", Runtime: "http"}}},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compiled.Workflows[0].WorkflowID != "alpha" {
		t.Errorf("want Workflows[0].WorkflowID == %q, got %q", "alpha", compiled.Workflows[0].WorkflowID)
	}
	if compiled.Workflows[1].WorkflowID != "beta" {
		t.Errorf("want Workflows[1].WorkflowID == %q, got %q", "beta", compiled.Workflows[1].WorkflowID)
	}
}

func TestCompile_StepOrderPreserved(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "A", WorkflowID: "wf-1", Runtime: "http"},
					{StepID: "B", WorkflowID: "wf-1", Runtime: "http"},
					{StepID: "C", WorkflowID: "wf-1", Runtime: "http"},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	steps := compiled.Workflows[0].Steps
	for i, want := range []string{"A", "B", "C"} {
		if steps[i].StepID != want {
			t.Errorf("Steps[%d].StepID: want %q got %q", i, want, steps[i].StepID)
		}
	}
}

func TestCompile_MultipleWorkflows(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: "http"},
					{StepID: "s2", WorkflowID: "wf-1", Runtime: "http"},
				},
			},
			{
				WorkflowID: "wf-2",
				Steps: []planner.StepPlan{
					{StepID: "s3", WorkflowID: "wf-2", Runtime: "http"},
					{StepID: "s4", WorkflowID: "wf-2", Runtime: "http"},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(compiled.Workflows) != 2 {
		t.Fatalf("want 2 workflows, got %d", len(compiled.Workflows))
	}
	if len(compiled.Workflows[0].Steps) != 2 {
		t.Errorf("wf-1: want 2 steps, got %d", len(compiled.Workflows[0].Steps))
	}
	if len(compiled.Workflows[1].Steps) != 2 {
		t.Errorf("wf-2: want 2 steps, got %d", len(compiled.Workflows[1].Steps))
	}
}

func TestCompile_NilPlan(t *testing.T) {
	compiled, err := NewCompiler().Compile(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil plan, got nil")
	}
	if compiled != nil {
		t.Error("expected nil CompiledPlan on error")
	}
	if !strings.Contains(err.Error(), ErrNilPlan) {
		t.Errorf("expected error code %q in: %v", ErrNilPlan, err)
	}
}

func TestCompile_EmptyWorkflows(t *testing.T) {
	plan := &planner.ExecutionPlan{PlanID: "p1", Workflows: nil}
	_, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err == nil {
		t.Fatal("expected ErrEmptyPlan")
	}
	if !strings.Contains(err.Error(), ErrEmptyPlan) {
		t.Errorf("expected error code %q in: %v", ErrEmptyPlan, err)
	}
}

func TestCompile_EmptyRuntime(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: ""},
				},
			},
		},
	}
	_, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err == nil {
		t.Fatal("expected ErrEmptyRuntime")
	}
	if !strings.Contains(err.Error(), ErrEmptyRuntime) {
		t.Errorf("expected error code %q in: %v", ErrEmptyRuntime, err)
	}
	if !strings.Contains(err.Error(), "wf-1") {
		t.Errorf("expected WorkflowID in error: %v", err)
	}
	if !strings.Contains(err.Error(), "s1") {
		t.Errorf("expected StepID in error: %v", err)
	}
}

func TestCompile_EmptyStepID(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "", WorkflowID: "wf-1", Runtime: "http"},
				},
			},
		},
	}
	_, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err == nil {
		t.Fatal("expected ErrEmptyStepID")
	}
	if !strings.Contains(err.Error(), ErrEmptyStepID) {
		t.Errorf("expected error code %q in: %v", ErrEmptyStepID, err)
	}
}

func TestCompile_EmptyWorkflowID(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "", Runtime: "http"},
				},
			},
		},
	}
	_, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err == nil {
		t.Fatal("expected ErrEmptyWorkflowID")
	}
	if !strings.Contains(err.Error(), ErrEmptyWorkflowID) {
		t.Errorf("expected error code %q in: %v", ErrEmptyWorkflowID, err)
	}
}

func TestCompile_ErrorsAggregated(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: ""},
					{StepID: "s2", WorkflowID: "wf-1", Runtime: ""},
				},
			},
		},
	}
	_, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err == nil {
		t.Fatal("expected errors")
	}
	msg := err.Error()
	if strings.Count(msg, ErrEmptyRuntime) < 2 {
		t.Errorf("expected at least 2 occurrences of %q in error, got: %s", ErrEmptyRuntime, msg)
	}
}

func TestCompile_SourceNotMutated(t *testing.T) {
	plan := validPlan(2)
	origID0 := plan.Workflows[0].Steps[0].StepID
	origID1 := plan.Workflows[0].Steps[1].StepID

	c := NewCompiler()
	_, err := c.Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("first Compile: %v", err)
	}
	_, err = c.Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("second Compile: %v", err)
	}

	if plan.Workflows[0].Steps[0].StepID != origID0 {
		t.Errorf("Steps[0].StepID mutated: want %q got %q", origID0, plan.Workflows[0].Steps[0].StepID)
	}
	if plan.Workflows[0].Steps[1].StepID != origID1 {
		t.Errorf("Steps[1].StepID mutated: want %q got %q", origID1, plan.Workflows[0].Steps[1].StepID)
	}
}

func TestCompile_Determinism(t *testing.T) {
	plan := validPlan(3)
	c := NewCompiler()

	cp1, err := c.Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("first Compile: %v", err)
	}
	cp2, err := c.Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("second Compile: %v", err)
	}

	if len(cp1.Workflows) != len(cp2.Workflows) {
		t.Fatalf("workflow count differs: %d vs %d", len(cp1.Workflows), len(cp2.Workflows))
	}
	for wi := range cp1.Workflows {
		w1, w2 := cp1.Workflows[wi], cp2.Workflows[wi]
		if w1.WorkflowID != w2.WorkflowID {
			t.Errorf("Workflows[%d].WorkflowID differs: %q vs %q", wi, w1.WorkflowID, w2.WorkflowID)
		}
		if len(w1.Steps) != len(w2.Steps) {
			t.Fatalf("Workflows[%d] step count differs", wi)
		}
		for si := range w1.Steps {
			s1, s2 := w1.Steps[si], w2.Steps[si]
			if s1.StepID != s2.StepID {
				t.Errorf("Steps[%d].StepID differs: %q vs %q", si, s1.StepID, s2.StepID)
			}
			if s1.Runtime != s2.Runtime {
				t.Errorf("Steps[%d].Runtime differs", si)
			}
		}
	}
}

func TestCompile_NilDependsOn(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: "http", DependsOn: nil},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// nil or empty []string are both acceptable; assert no panic and length == 0
	if len(compiled.Workflows[0].Steps[0].DependsOn) != 0 {
		t.Errorf("expected empty DependsOn, got %v", compiled.Workflows[0].Steps[0].DependsOn)
	}
}

func TestCompile_EmptyDependsOn(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps: []planner.StepPlan{
					{StepID: "s1", WorkflowID: "wf-1", Runtime: "http", DependsOn: []string{}},
				},
			},
		},
	}
	compiled, err := NewCompiler().Compile(plan, specFromPlan(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(compiled.Workflows[0].Steps[0].DependsOn) != 0 {
		t.Errorf("expected empty DependsOn, got %v", compiled.Workflows[0].Steps[0].DependsOn)
	}
}

func TestCompile_RequestPayloadCopied(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID: "p1",
		Workflows: []planner.WorkflowPlan{
			{
				WorkflowID: "wf-1",
				Steps:      []planner.StepPlan{{StepID: "req1", WorkflowID: "wf-1", Runtime: "http"}},
			},
		},
	}
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{
			{
				ID: "wf-1",
				Steps: []spec.Step{
					{
						ID: "req1",
						Request: &spec.RequestDescriptor{
							Protocol:  "http",
							Target:    "https://example.local/${vars.host}",
							Operation: "POST",
							Headers:   map[string]string{"X-Req": "${vars.token}"},
							Body: &spec.BodyDescriptor{
								Encoding: "json",
								Content:  map[string]interface{}{"k": "v"},
							},
						},
					},
				},
			},
		},
	}
	cp, err := NewCompiler().Compile(plan, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cs := cp.Workflows[0].Steps[0]
	if cs.Request == nil {
		t.Fatal("expected Request to be set")
	}
	if cs.Request.Protocol != "http" || cs.Request.Target != "https://example.local/${vars.host}" || cs.Request.Method != "POST" {
		t.Fatalf("request fields mismatch: %#v", cs.Request)
	}
	if cs.Request.Body == nil || cs.Request.Body.Encoding != "json" {
		t.Fatalf("body mismatch: %#v", cs.Request.Body)
	}
}

func TestCompile_AssertionsCopied(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID:    "p1",
		Workflows: []planner.WorkflowPlan{{WorkflowID: "wf-1", Steps: []planner.StepPlan{{StepID: "s1", WorkflowID: "wf-1", Runtime: "http"}}}},
	}
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{
			{
				ID: "wf-1",
				Steps: []spec.Step{
					{
						ID: "s1",
						Assertions: []spec.Assertion{{
							ID:       "a1",
							Kind:     "status",
							Op:       "equals",
							Target:   "status",
							Expected: 200,
							Severity: spec.SeverityError,
						}},
					},
				},
			},
		},
	}
	cp, err := NewCompiler().Compile(plan, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cs := cp.Workflows[0].Steps[0]
	if len(cs.Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(cs.Assertions))
	}
	if cs.Assertions[0].ID != "a1" || cs.Assertions[0].Op != "equals" {
		t.Fatalf("assertion mismatch: %#v", cs.Assertions[0])
	}
}

func TestCompile_ExtractsCopied(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID:    "p1",
		Workflows: []planner.WorkflowPlan{{WorkflowID: "wf-1", Steps: []planner.StepPlan{{StepID: "s1", WorkflowID: "wf-1", Runtime: "http"}}}},
	}
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{
			{
				ID: "wf-1",
				Steps: []spec.Step{
					{
						ID: "s1",
						Extracts: []spec.Extract{{
							ID:     "e1",
							Source: "body",
							Path:   "$.token",
							As:     "tok",
							Scope:  spec.ExtractScopeStep,
						}},
					},
				},
			},
		},
	}
	cp, err := NewCompiler().Compile(plan, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cs := cp.Workflows[0].Steps[0]
	if len(cs.Extracts) != 1 {
		t.Fatalf("expected 1 extract, got %d", len(cs.Extracts))
	}
	if cs.Extracts[0].ID != "e1" || cs.Extracts[0].As != "tok" || cs.Extracts[0].Scope != string(spec.ExtractScopeStep) {
		t.Fatalf("extract mismatch: %#v", cs.Extracts[0])
	}
}

func TestCompile_MissingStepPayload(t *testing.T) {
	plan := &planner.ExecutionPlan{
		PlanID:    "p1",
		Workflows: []planner.WorkflowPlan{{WorkflowID: "wf-1", Steps: []planner.StepPlan{{StepID: "s-x", WorkflowID: "wf-1", Runtime: "http"}}}},
	}
	// spec has no matching step
	s := &spec.TraCtlSpec{Workflows: []spec.Workflow{{ID: "wf-1", Steps: []spec.Step{{ID: "other"}}}}}
	_, err := NewCompiler().Compile(plan, s)
	if err == nil {
		t.Fatal("expected ErrMissingStepPayload")
	}
	if !strings.Contains(err.Error(), ErrMissingStepPayload) {
		t.Fatalf("expected %s in error: %v", ErrMissingStepPayload, err)
	}
}
