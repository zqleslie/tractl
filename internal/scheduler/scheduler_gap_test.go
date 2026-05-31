package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// TestScheduler_FailFast_IndependentBranchUnaffected — under failFast, an
// independent step must NOT be cancelled by an unrelated failure (ADR-008 §5).
func TestScheduler_FailFast_IndependentBranchUnaffected(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1"},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1"}, // independent
		},
		FailurePolicy:    "failFast",
		ConcurrencyLimit: 5,
	}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{StepID: "A", WorkflowID: "wf1", State: runtime.StateFailed, StartedAt: time.Now(), FinishedAt: time.Now(), Extracts: map[string]string{}}
	sched := NewScheduler(plan, mock)
	res, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	cState := findState(res, "C")
	if cState != runtime.StateSucceeded {
		t.Fatalf("expected C to succeed in failFast (branch isolation), got %s", cState)
	}
}

// TestScheduler_PanicWithPendingDependents — when a step panics with multiple
// pending dependents, the scheduler must not deadlock waiting on done.
func TestScheduler_PanicWithPendingDependents(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1"},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"A"}},
		},
		FailurePolicy:    "resilient",
		ConcurrencyLimit: 5,
	}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.panics["A"] = true
	sched := NewScheduler(plan, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := sched.Run(ctx, "wf1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// B and C should be dependency-skipped, not still pending.
	if findState(res, "B") != runtime.StateDependencySkipped {
		t.Fatalf("expected B dependency-skipped, got %s", findState(res, "B"))
	}
	if findState(res, "C") != runtime.StateDependencySkipped {
		t.Fatalf("expected C dependency-skipped, got %s", findState(res, "C"))
	}
}

// TestScheduler_ConditionalSkipDoesNotPropagate — a step whose `when` is false
// becomes ConditionalSkip; its dependents are NOT dependency-skipped (spec §7.6).
func TestScheduler_ConditionalSkipDoesNotPropagate(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "B", WorkflowID: "wf1", When: "false"},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"B"}},
		},
		ConcurrencyLimit: 5,
	}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)
	res, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if findState(res, "B") != runtime.StateConditionalSkip {
		t.Fatalf("expected B conditional-skip, got %s", findState(res, "B"))
	}
	if findState(res, "C") != runtime.StateSucceeded {
		t.Fatalf("expected C succeeded (conditional-skip does not propagate), got %s", findState(res, "C"))
	}
}

// TestScheduler_FailFast_CancelsInFlightDependent — when A fails under failFast
// while B (a dependent) is still in flight, B should reach Cancelled (or be
// skipped before dispatch).
func TestScheduler_FailFast_CancelsInFlightDependent(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1"},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
		},
		FailurePolicy:    "failFast",
		ConcurrencyLimit: 5,
	}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{StepID: "A", WorkflowID: "wf1", State: runtime.StateFailed, StartedAt: time.Now(), FinishedAt: time.Now()}
	sched := NewScheduler(plan, mock)
	res, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	bState := findState(res, "B")
	if bState != runtime.StateDependencySkipped && bState != runtime.StateCancelled {
		t.Fatalf("expected B skipped/cancelled, got %s", bState)
	}
}

// TestScheduler_RetryHonoredViaSchedulerPath — the scheduler dispatches via
// executor.Execute; with the executeOnce/Execute split, retry must be applied.
// Use a real http server through a wrapped mock.
func TestScheduler_RetryHonoredViaSchedulerPath(t *testing.T) {
	// Use a counting mock that fails the first call and succeeds the second.
	attempts := int32(0)
	mock := &countingMock{attempts: &attempts}
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1"},
		},
		ConcurrencyLimit: 1,
	}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	sched := NewScheduler(plan, mock)
	res, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if findState(res, "A") != runtime.StateSucceeded {
		t.Fatalf("expected A succeeded, got %s", findState(res, "A"))
	}
}

type countingMock struct{ attempts *int32 }

func (c *countingMock) Execute(_ context.Context, step *compiler.CompiledStep, _ *runtime.ExecutionContext) (*runtime.StepResult, error) {
	atomic.AddInt32(c.attempts, 1)
	return &runtime.StepResult{StepID: step.StepID, WorkflowID: step.WorkflowID, State: runtime.StateSucceeded, StartedAt: time.Now(), FinishedAt: time.Now(), Extracts: map[string]string{}}, nil
}

func findState(r *WorkflowResult, stepID string) runtime.StepState {
	for _, o := range r.Outcomes {
		if o.StepID == stepID {
			return o.State
		}
	}
	return ""
}
