package planner

import (
	"strings"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/spec"
)

// TestTopoSort_Linear tests basic linear dependency chain.
func TestTopoSort_Linear(t *testing.T) {
	steps := []spec.Step{
		{ID: "a", Kind: "request", DependsOn: []string{}},
		{ID: "b", Kind: "request", DependsOn: []string{"a"}},
		{ID: "c", Kind: "request", DependsOn: []string{"b"}},
	}
	sorted, err := topoSort("test-workflow", steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(sorted))
	}
	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("expected order [a, b, c], got [%s, %s, %s]", sorted[0].ID, sorted[1].ID, sorted[2].ID)
	}
}

// TestTopoSort_Diamond tests diamond DAG shape.
func TestTopoSort_Diamond(t *testing.T) {
	steps := []spec.Step{
		{ID: "a", Kind: "request", DependsOn: []string{}},
		{ID: "b", Kind: "request", DependsOn: []string{"a"}},
		{ID: "c", Kind: "request", DependsOn: []string{"a"}},
		{ID: "d", Kind: "request", DependsOn: []string{"b", "c"}},
	}
	sorted, err := topoSort("test-workflow", steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(sorted))
	}
	// A must be first, D must be last
	if sorted[0].ID != "a" {
		t.Errorf("expected first step to be 'a', got %q", sorted[0].ID)
	}
	if sorted[3].ID != "d" {
		t.Errorf("expected last step to be 'd', got %q", sorted[3].ID)
	}
	// B and C must be between A and D, in lexicographic order
	if sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("expected [a, b, c, d], got [%s, %s, %s, %s]", sorted[0].ID, sorted[1].ID, sorted[2].ID, sorted[3].ID)
	}
}

// TestTopoSort_NoDependencies tests deterministic ordering of independent steps.
func TestTopoSort_NoDependencies(t *testing.T) {
	steps := []spec.Step{
		{ID: "c", Kind: "request", DependsOn: []string{}},
		{ID: "a", Kind: "request", DependsOn: []string{}},
		{ID: "b", Kind: "request", DependsOn: []string{}},
	}
	sorted, err := topoSort("test-workflow", steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(sorted))
	}
	// Should be in lexicographic order: [a, b, c]
	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("expected [a, b, c], got [%s, %s, %s]", sorted[0].ID, sorted[1].ID, sorted[2].ID)
	}
}

// TestTopoSort_Cycle tests cycle detection.
func TestTopoSort_Cycle(t *testing.T) {
	steps := []spec.Step{
		{ID: "a", Kind: "request", DependsOn: []string{"b"}},
		{ID: "b", Kind: "request", DependsOn: []string{"a"}},
	}
	_, err := topoSort("test-workflow", steps)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), ErrCyclicDependency) {
		t.Errorf("expected error code %q, got %v", ErrCyclicDependency, err)
	}
}

// TestTopoSort_ImplicitEdgesAlreadyMerged documents the contract with Phase 3.5.
func TestTopoSort_ImplicitEdgesAlreadyMerged(t *testing.T) {
	// This test demonstrates that DependsOn contains BOTH explicit and
	// implicit edges. The normalizer has already merged implicit edges
	// from variable expressions into DependsOn.
	steps := []spec.Step{
		{ID: "login", Kind: "request", DependsOn: []string{}},
		{ID: "fetch-profile", Kind: "request", DependsOn: []string{"login"}},
	}
	sorted, err := topoSort("test-workflow", steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(sorted))
	}
	if sorted[0].ID != "login" || sorted[1].ID != "fetch-profile" {
		t.Errorf("expected [login, fetch-profile], got [%s, %s]", sorted[0].ID, sorted[1].ID)
	}
}

// TestRegistry_HTTP tests protocol.http capability resolution.
func TestRegistry_HTTP(t *testing.T) {
	reg := DefaultRegistry()
	resolved, err := reg.Resolve("protocol.http")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Runtime != "http" {
		t.Errorf("expected Runtime=http, got %q", resolved.Runtime)
	}
	if resolved.Version != "1" {
		t.Errorf("expected Version=1, got %q", resolved.Version)
	}
}

// TestRegistry_UnknownContract tests unknown capability rejection.
func TestRegistry_UnknownContract(t *testing.T) {
	reg := DefaultRegistry()
	_, err := reg.Resolve("protocol.grpc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), ErrUnknownCapability) {
		t.Errorf("expected error code %q, got %v", ErrUnknownCapability, err)
	}
}

// TestRegistry_EmptyContract tests empty contract rejection.
func TestRegistry_EmptyContract(t *testing.T) {
	reg := DefaultRegistry()
	_, err := reg.Resolve("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestPlan_SingleWorkflowLinear tests planning a single linear workflow.
func TestPlan_SingleWorkflowLinear(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
					{ID: "c", Kind: "request", DependsOn: []string{"b"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(plan.Workflows))
	}

	wf := plan.Workflows[0]
	if len(wf.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(wf.Steps))
	}

	if wf.Steps[0].StepID != "a" || wf.Steps[1].StepID != "b" || wf.Steps[2].StepID != "c" {
		t.Errorf("expected order [a, b, c], got [%s, %s, %s]",
			wf.Steps[0].StepID, wf.Steps[1].StepID, wf.Steps[2].StepID)
	}
}

// TestPlan_SingleWorkflowDiamond tests planning a diamond workflow.
func TestPlan_SingleWorkflowDiamond(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
					{ID: "c", Kind: "request", DependsOn: []string{"a"}},
					{ID: "d", Kind: "request", DependsOn: []string{"b", "c"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf := plan.Workflows[0]
	// A first, D last, B before D, C before D
	if wf.Steps[0].StepID != "a" {
		t.Errorf("expected first step 'a', got %q", wf.Steps[0].StepID)
	}
	if wf.Steps[3].StepID != "d" {
		t.Errorf("expected last step 'd', got %q", wf.Steps[3].StepID)
	}
}

// TestPlan_MultipleWorkflows tests planning multiple workflows.
func TestPlan_MultipleWorkflows(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "wf1",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
			{
				ID: "wf2",
				Steps: []spec.Step{
					{ID: "x", Kind: "request", DependsOn: []string{}},
					{ID: "y", Kind: "request", DependsOn: []string{"x"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(plan.Workflows))
	}

	if plan.Workflows[0].WorkflowID != "wf1" || plan.Workflows[1].WorkflowID != "wf2" {
		t.Errorf("expected workflows [wf1, wf2], got [%s, %s]",
			plan.Workflows[0].WorkflowID, plan.Workflows[1].WorkflowID)
	}
}

// TestPlan_DefaultFailurePolicyResilient tests default failure policy.
func TestPlan_DefaultFailurePolicyResilient(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID:            "test",
				FailurePolicy: "",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Workflows[0].FailurePolicy != "resilient" {
		t.Errorf("expected failurePolicy=resilient, got %q", plan.Workflows[0].FailurePolicy)
	}
	if plan.Workflows[0].Steps[0].FailurePolicy != "resilient" {
		t.Errorf("expected step failurePolicy=resilient, got %q", plan.Workflows[0].Steps[0].FailurePolicy)
	}
}

// TestPlan_ExplicitFailurePolicyFailFast tests explicit failure policy.
func TestPlan_ExplicitFailurePolicyFailFast(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID:            "test",
				FailurePolicy: "failFast",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Workflows[0].FailurePolicy != "failFast" {
		t.Errorf("expected failurePolicy=failFast, got %q", plan.Workflows[0].FailurePolicy)
	}
	for i, step := range plan.Workflows[0].Steps {
		if step.FailurePolicy != "failFast" {
			t.Errorf("step[%d]: expected failurePolicy=failFast, got %q", i, step.FailurePolicy)
		}
	}
}

// TestPlan_ConcurrencySlotAssignment tests concurrency slot assignment.
func TestPlan_ConcurrencySlotAssignment(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID:          "test",
				Concurrency: 2,
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{}},
					{ID: "c", Kind: "request", DependsOn: []string{}},
					{ID: "d", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := plan.Workflows[0].Steps
	expectedSlots := []int{1, 2, 1, 2}
	for i, step := range steps {
		if step.ConcurrencySlot != expectedSlots[i] {
			t.Errorf("step[%d]: expected slot %d, got %d", i, expectedSlots[i], step.ConcurrencySlot)
		}
	}
}

// TestPlan_ConcurrencyZeroDefaultsFour tests zero concurrency defaults to 4,
// so two independent steps receive slots 1 and 2 (within the 4-slot pool).
func TestPlan_ConcurrencyZeroDefaultsFour(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID:          "test",
				Concurrency: 0,
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With default concurrency=4, slots are assigned as (index % 4)+1: 1, 2.
	expectedSlots := []int{1, 2}
	for i, step := range plan.Workflows[0].Steps {
		if step.ConcurrencySlot != expectedSlots[i] {
			t.Errorf("step[%d]: expected slot %d, got %d", i, expectedSlots[i], step.ConcurrencySlot)
		}
	}
	if plan.Workflows[0].ConcurrencyLimit != 4 {
		t.Errorf("expected ConcurrencyLimit=4 (default), got %d", plan.Workflows[0].ConcurrencyLimit)
	}
}

// TestPlan_CanSkipFalseForRootStep tests CanSkip is false for root steps.
func TestPlan_CanSkipFalseForRootStep(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Workflows[0].Steps[0].CanSkip {
		t.Error("expected CanSkip=false for root step, got true")
	}
}

// TestPlan_CanSkipTrueForDependentStep tests CanSkip is true for dependent steps.
func TestPlan_CanSkipTrueForDependentStep(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Workflows[0].Steps[1].CanSkip {
		return // this is correct
	}
	t.Error("expected CanSkip=true for dependent step, got false")
}

// TestPlan_UnknownStepKind tests unknown step kind rejection.
func TestPlan_UnknownStepKind(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "extension", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	_, err := planner.Plan(spec)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), ErrUnknownCapability) {
		t.Errorf("expected error code %q, got %v", ErrUnknownCapability, err)
	}
}

// TestPlan_CyclicWorkflow tests cycle detection in plan.
func TestPlan_CyclicWorkflow(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{"b"}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	_, err := planner.Plan(spec)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), ErrCyclicDependency) {
		t.Errorf("expected error code %q, got %v", ErrCyclicDependency, err)
	}
}

// TestPlan_NilSpec tests nil spec rejection.
func TestPlan_NilSpec(t *testing.T) {
	planner := NewPlanner(DefaultRegistry())
	_, err := planner.Plan(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestPlan_EmptyWorkflows tests empty workflows rejection.
func TestPlan_EmptyWorkflows(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows:     []spec.Workflow{},
	}

	planner := NewPlanner(DefaultRegistry())
	_, err := planner.Plan(spec)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), ErrEmptySpec) {
		t.Errorf("expected error code %q, got %v", ErrEmptySpec, err)
	}
}

// TestPlan_PlanTraceFields tests plan trace metadata.
func TestPlan_PlanTraceFields(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "wf1",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
			{
				ID: "wf2",
				Steps: []spec.Step{
					{ID: "x", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.PlanID == "" {
		t.Error("expected non-empty PlanID")
	}
	if plan.Trace.PlanID != plan.PlanID {
		t.Errorf("expected Trace.PlanID == PlanID")
	}
	if plan.Trace.WorkflowCount != 2 {
		t.Errorf("expected WorkflowCount=2, got %d", plan.Trace.WorkflowCount)
	}
	if plan.Trace.TotalSteps != 3 {
		t.Errorf("expected TotalSteps=3, got %d", plan.Trace.TotalSteps)
	}
	if plan.Trace.GeneratedAt == "" {
		t.Error("expected non-empty GeneratedAt")
	}
	// Verify it's a valid RFC3339 timestamp
	if _, err := time.Parse(time.RFC3339, plan.Trace.GeneratedAt); err != nil {
		t.Errorf("GeneratedAt is not valid RFC3339: %v", err)
	}
}

// TestPlan_SourceNotMutated tests that Plan doesn't mutate the source spec.
func TestPlan_SourceNotMutated(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"a"}},
				},
			},
		},
	}

	// Save original DependsOn
	origDeps := spec.Workflows[0].Steps[1].DependsOn

	planner := NewPlanner(DefaultRegistry())
	_, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Run Plan again to verify no mutation
	_, err = planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error on second plan: %v", err)
	}

	// Verify original DependsOn unchanged
	if len(spec.Workflows[0].Steps[1].DependsOn) != len(origDeps) {
		t.Error("source spec was mutated: DependsOn length changed")
	}
}

// TestPlan_Determinism tests deterministic plan generation.
func TestPlan_Determinism(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "c", Kind: "request", DependsOn: []string{}},
					{ID: "a", Kind: "request", DependsOn: []string{}},
					{ID: "b", Kind: "request", DependsOn: []string{"c"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan1, _ := planner.Plan(spec)
	plan2, _ := planner.Plan(spec)

	// PlanID and GeneratedAt will differ; compare Workflows
	if len(plan1.Workflows) != len(plan2.Workflows) {
		t.Error("plan1 and plan2 have different workflow counts")
	}

	wf1 := plan1.Workflows[0]
	wf2 := plan2.Workflows[0]

	if len(wf1.Steps) != len(wf2.Steps) {
		t.Error("workflows have different step counts")
	}

	for i := range wf1.Steps {
		if wf1.Steps[i].StepID != wf2.Steps[i].StepID {
			t.Errorf("step order differs at index %d: %q vs %q", i, wf1.Steps[i].StepID, wf2.Steps[i].StepID)
		}
	}
}

// TestPlan_PlanWorkflowByID tests PlanWorkflow convenience method.
func TestPlan_PlanWorkflowByID(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "wf1",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
				},
			},
			{
				ID: "wf2",
				Steps: []spec.Step{
					{ID: "x", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.PlanWorkflow(spec, "wf2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(plan.Workflows))
	}
	if plan.Workflows[0].WorkflowID != "wf2" {
		t.Errorf("expected workflow wf2, got %s", plan.Workflows[0].WorkflowID)
	}
}

// TestPlan_PlanWorkflowNotFound tests PlanWorkflow with invalid ID.
func TestPlan_PlanWorkflowNotFound(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "wf1",
				Steps: []spec.Step{
					{ID: "a", Kind: "request", DependsOn: []string{}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	_, err := planner.PlanWorkflow(spec, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got %v", err)
	}
}

// TestPlan_HybridDependencies tests the hybrid explicit+implicit dependency contract.
func TestPlan_HybridDependencies(t *testing.T) {
	spec := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "test",
				Steps: []spec.Step{
					{ID: "login", Kind: "request", DependsOn: []string{}},
					{ID: "fetch-profile", Kind: "request", DependsOn: []string{"login"}},
					{ID: "create-order", Kind: "request", DependsOn: []string{"fetch-profile"}},
				},
			},
		},
	}

	planner := NewPlanner(DefaultRegistry())
	plan, err := planner.Plan(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := plan.Workflows[0].Steps
	if steps[0].StepID != "login" || steps[1].StepID != "fetch-profile" || steps[2].StepID != "create-order" {
		t.Errorf("expected [login, fetch-profile, create-order], got [%s, %s, %s]",
			steps[0].StepID, steps[1].StepID, steps[2].StepID)
	}

	// Verify DependsOn edges are preserved
	if len(steps[1].DependsOn) != 1 || steps[1].DependsOn[0] != "login" {
		t.Errorf("fetch-profile should depend on [login], got %v", steps[1].DependsOn)
	}
	if len(steps[2].DependsOn) != 1 || steps[2].DependsOn[0] != "fetch-profile" {
		t.Errorf("create-order should depend on [fetch-profile], got %v", steps[2].DependsOn)
	}

	// Verify CanSkip values
	if steps[0].CanSkip {
		t.Error("login should have CanSkip=false")
	}
	if !steps[1].CanSkip || !steps[2].CanSkip {
		t.Error("fetch-profile and create-order should have CanSkip=true")
	}
}

// h3Spec builds a minimal spec with one step that declares transport.version "h3".
func h3Spec() *spec.TraCtlSpec {
	return &spec.TraCtlSpec{
		SchemaVersion: 1,
		Workflows: []spec.Workflow{
			{
				ID: "wf",
				Steps: []spec.Step{
					{
						ID:   "fetch",
						Kind: "request",
						Request: &spec.RequestDescriptor{
							Protocol:  "http",
							Target:    "https://api.example.com/data",
							Operation: "GET",
							Transport: map[string]interface{}{"version": "h3"},
						},
					},
				},
			},
		},
	}
}

// TestPlannerRejectsHTTP3OnCLI verifies that transport.version "h3" is rejected
// on the CLI surface with an error referencing "h3" and "WASM" (ADR-017 §7).
func TestPlannerRejectsHTTP3OnCLI(t *testing.T) {
	planner := NewPlannerWithConfig(DefaultRegistry(), Config{Surface: "cli"})
	_, err := planner.Plan(h3Spec())
	if err == nil {
		t.Fatal("expected planning error for h3 on CLI surface, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "h3") {
		t.Errorf("expected error message to contain %q, got: %s", "h3", msg)
	}
	if !strings.Contains(msg, "WASM") {
		t.Errorf("expected error message to contain %q, got: %s", "WASM", msg)
	}
	if !strings.Contains(msg, ErrTransportNotSupported) {
		t.Errorf("expected error code %q, got: %s", ErrTransportNotSupported, msg)
	}
}

// TestPlannerAllowsHTTP3OnWASM verifies that transport.version "h3" is accepted
// on the WASM surface without error (ADR-017 §7).
func TestPlannerAllowsHTTP3OnWASM(t *testing.T) {
	planner := NewPlannerWithConfig(DefaultRegistry(), Config{Surface: "wasm"})
	_, err := planner.Plan(h3Spec())
	if err != nil {
		t.Fatalf("expected no planning error for h3 on WASM surface, got: %v", err)
	}
}

// TestPlannerIgnoresTransportWhenNoVersion verifies that a transport map with no
// "version" key does not trigger the H3 guard on any surface (ADR-017 §7).
func TestPlannerIgnoresTransportWhenNoVersion(t *testing.T) {
	surfaces := []string{"cli", "desktop", "localapi", "wasm"}
	for _, surface := range surfaces {
		t.Run(surface, func(t *testing.T) {
			s := &spec.TraCtlSpec{
				SchemaVersion: 1,
				Workflows: []spec.Workflow{
					{
						ID: "wf",
						Steps: []spec.Step{
							{
								ID:   "fetch",
								Kind: "request",
								Request: &spec.RequestDescriptor{
									Protocol:  "http",
									Target:    "https://api.example.com/data",
									Operation: "GET",
									Transport: map[string]interface{}{"proxy": "socks5://proxy.example.com"},
								},
							},
						},
					},
				},
			}
			planner := NewPlannerWithConfig(DefaultRegistry(), Config{Surface: surface})
			_, err := planner.Plan(s)
			if err != nil {
				t.Errorf("expected no error for transport without version on %q surface, got: %v", surface, err)
			}
		})
	}
}
