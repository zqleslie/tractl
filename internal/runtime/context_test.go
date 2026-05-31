// context_test.go defines code for the runtime package.

package runtime

import (
	"testing"
)

func TestNewExecutionContext_ReturnsContext(t *testing.T) {
	ec, err := NewExecutionContext("wf1", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ec == nil {
		t.Fatal("expected non-nil ExecutionContext")
	}
	if ec.WorkflowID() != "wf1" {
		t.Fatalf("workflowID mismatch: got %q", ec.WorkflowID())
	}
	if ec.ExecutionID() == "" {
		t.Fatal("expected non-empty execution ID")
	}
}

func TestNewExecutionContext_UniqueIDs(t *testing.T) {
	a, err := NewExecutionContext("wf1", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := NewExecutionContext("wf1", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ExecutionID() == b.ExecutionID() {
		t.Fatal("expected unique execution IDs for separate contexts")
	}
}

func TestSetVar_StepScope_WritesStepExtract(t *testing.T) {
	ec, err := NewExecutionContext("wf-step-scope", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ec.SetActiveStepID("step-one")
	defer ec.ClearActiveStepID()

	if err := ec.SetVar(string(ScopeStep), "hook-token", "set-by-hook"); err != nil {
		t.Fatalf("SetVar step scope: %v", err)
	}
	got, ok := ec.GetExtract("step-one", "hook-token")
	if !ok || got != "set-by-hook" {
		t.Fatalf("GetExtract: got %q ok=%v", got, ok)
	}
}

func TestSetVar_StepScope_RequiresActiveStep(t *testing.T) {
	ec, err := NewExecutionContext("wf-step-scope", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = ec.SetVar(string(ScopeStep), "hook-token", "x")
	if err == nil {
		t.Fatal("expected error when active step ID is unset")
	}
}

func TestContext_ScopeRuntime_IsReadOnly(t *testing.T) {
	ec, err := NewExecutionContext("wf-runtime-scope", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ec.SetVar(string(ScopeRuntime), "key", "value")
	if err == nil {
		t.Error("expected error when setting a value in read-only ScopeRuntime, got nil")
	}
}
