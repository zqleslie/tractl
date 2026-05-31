package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// mockExecutor is a test double for executor.Executor
type mockExecutor struct {
	results map[string]*runtime.StepResult
	panics  map[string]bool
	delay   time.Duration
	mu      sync.Mutex
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		results: make(map[string]*runtime.StepResult),
		panics:  make(map[string]bool),
	}
}

func (m *mockExecutor) Execute(ctx context.Context, step *compiler.CompiledStep, _ *runtime.ExecutionContext) (*runtime.StepResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// check for panic scenario
	if m.panics[step.StepID] {
		panic("intentional panic from mock executor")
	}

	// simulate delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return &runtime.StepResult{
				StepID:     step.StepID,
				WorkflowID: step.WorkflowID,
				State:      runtime.StateCancelled,
				StartedAt:  time.Now(),
				FinishedAt: time.Now(),
				Extracts:   map[string]string{},
			}, ctx.Err()
		}
	}

	// check for configured result
	if result, ok := m.results[step.StepID]; ok {
		return result, nil
	}

	// default success
	return &runtime.StepResult{
		StepID:     step.StepID,
		WorkflowID: step.WorkflowID,
		State:      runtime.StateSucceeded,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Extracts:   map[string]string{},
	}, nil
}

func TestScheduler_LinearChain(t *testing.T) {
	// A -> B -> C
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"B"}},
		},
		ConcurrencyLimit: 1,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}
	if len(result.Outcomes) != 3 {
		t.Fatalf("expected 3 outcomes, got %d", len(result.Outcomes))
	}
}

func TestScheduler_ParallelBranches(t *testing.T) {
	// A is root; B and C both depend on A; D depends on B and C
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "D", WorkflowID: "wf1", DependsOn: []string{"B", "C"}},
		},
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}
}

func TestScheduler_DependencySkippedOnFailure(t *testing.T) {
	// A fails; B depends on A
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
		},
		ConcurrencyLimit: 1,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{
		StepID:     "A",
		WorkflowID: "wf1",
		State:      runtime.StateFailed,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Extracts:   map[string]string{},
	}
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// B should be dependency-skipped
	var bOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "B" {
			bOutcome = &result.Outcomes[i]
			break
		}
	}
	if bOutcome == nil || bOutcome.State != runtime.StateDependencySkipped {
		t.Fatalf("expected B to be dependency-skipped, got %v", bOutcome.State)
	}
}

func TestScheduler_IndependentBranchContinuesOnFailure(t *testing.T) {
	// A fails; C is independent
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
		},
		FailurePolicy:    "resilient",
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{
		StepID:     "A",
		WorkflowID: "wf1",
		State:      runtime.StateFailed,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Extracts:   map[string]string{},
	}
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// C should still succeed despite A failing (resilient policy)
	var cOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "C" {
			cOutcome = &result.Outcomes[i]
			break
		}
	}
	if cOutcome == nil || cOutcome.State != runtime.StateSucceeded {
		t.Fatalf("expected C to succeed, got %v", cOutcome.State)
	}
	if result.OverallState != runtime.StateFailed {
		t.Fatalf("expected overall failed, got %s", result.OverallState)
	}
}

func TestScheduler_ConditionalSkip(t *testing.T) {
	// B has when condition that evaluates to "false"
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "B", WorkflowID: "wf1", When: "false", DependsOn: []string{}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"B"}},
		},
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// B should be conditional-skipped
	var bOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "B" {
			bOutcome = &result.Outcomes[i]
			break
		}
	}
	if bOutcome == nil || bOutcome.State != runtime.StateConditionalSkip {
		t.Fatalf("expected B to be conditional-skipped, got %v", bOutcome.State)
	}

	// C should succeed (conditional-skip is treated as succeeded for dependencies)
	var cOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "C" {
			cOutcome = &result.Outcomes[i]
			break
		}
	}
	if cOutcome == nil || cOutcome.State != runtime.StateSucceeded {
		t.Fatalf("expected C to succeed, got %v", cOutcome.State)
	}
}

func TestScheduler_ConcurrencyLimit(t *testing.T) {
	// 3 independent steps with concurrency limit 1
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
		},
		ConcurrencyLimit: 1,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.delay = 10 * time.Millisecond
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	// cancel parent ctx mid-execution
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
		},
		ConcurrencyLimit: 1,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.delay = 500 * time.Millisecond
	sched := NewScheduler(plan, mock)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	result, err := sched.Run(ctx, "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// either cancelled or failed is acceptable since they were slow and context was cancelled
	if result.OverallState != runtime.StateCancelled && result.OverallState != runtime.StateFailed {
		t.Fatalf("expected cancelled or failed, got %s", result.OverallState)
	}
}

func TestScheduler_PanicRecovery(t *testing.T) {
	// mock executor panics for step B
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
		},
		FailurePolicy:    "resilient",
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.panics["B"] = true
	sched := NewScheduler(plan, mock)

	// should not panic; B should fail, C should succeed
	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "B" {
			bOutcome = &result.Outcomes[i]
			break
		}
	}
	if bOutcome == nil || bOutcome.State != runtime.StateFailed {
		t.Fatalf("expected B to fail, got %v", bOutcome.State)
	}
}

func TestScheduler_AllSucceeded_OverallSucceeded(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
		},
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}
}

func TestScheduler_AnyFailed_OverallFailed(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
		},
		FailurePolicy:    "resilient",
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{
		StepID:     "A",
		WorkflowID: "wf1",
		State:      runtime.StateFailed,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Extracts:   map[string]string{},
	}
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateFailed {
		t.Fatalf("expected failed, got %s", result.OverallState)
	}
}

func TestScheduler_NilExecutorPanics(t *testing.T) {
	workflow := &compiler.CompiledWorkflow{WorkflowID: "wf1"}
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when executor is nil")
		}
	}()

	NewScheduler(plan, nil)
}

func TestScheduler_WorkflowNotFound(t *testing.T) {
	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	_, err := sched.Run(context.Background(), "nonexistent")
	if err == nil {
		t.Fatalf("expected error for missing workflow")
	}
}

func TestScheduler_RaceCondition(t *testing.T) {
	// 5 independent steps with ConcurrencyLimit=5
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "D", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "E", WorkflowID: "wf1", DependsOn: []string{}},
		},
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.delay = 5 * time.Millisecond
	sched := NewScheduler(plan, mock)

	// run multiple times to catch races
	for i := 0; i < 3; i++ {
		result, err := sched.Run(context.Background(), "wf1")
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if result.OverallState != runtime.StateSucceeded {
			t.Fatalf("iteration %d: expected succeeded, got %s", i, result.OverallState)
		}
	}
}

func TestScheduler_DiamondDAG(t *testing.T) {
	// diamond: A -> B,C -> D
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "D", WorkflowID: "wf1", DependsOn: []string{"B", "C"}},
		},
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}

	// verify all steps completed
	if len(result.Outcomes) != 4 {
		t.Fatalf("expected 4 outcomes, got %d", len(result.Outcomes))
	}
}

func TestScheduler_FailFastCancelsDependents(t *testing.T) {
	// A fails with failFast; B depends on A; C is independent
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{"A"}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
		},
		FailurePolicy:    "failFast",
		ConcurrencyLimit: 5,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.results["A"] = &runtime.StepResult{
		StepID:     "A",
		WorkflowID: "wf1",
		State:      runtime.StateFailed,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Extracts:   map[string]string{},
	}
	sched := NewScheduler(plan, mock)

	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// B should be dependency-skipped or cancelled (not executed)
	var bOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "B" {
			bOutcome = &result.Outcomes[i]
			break
		}
	}
	if bOutcome == nil || (bOutcome.State != runtime.StateDependencySkipped && bOutcome.State != runtime.StateCancelled) {
		t.Fatalf("expected B to be skipped or cancelled, got %v", bOutcome.State)
	}

	// C should eventually complete (branch isolation)
	var cOutcome *StepOutcome
	for i := range result.Outcomes {
		if result.Outcomes[i].StepID == "C" {
			cOutcome = &result.Outcomes[i]
			break
		}
	}
	if cOutcome == nil || cOutcome.State != runtime.StateSucceeded {
		t.Fatalf("expected C to succeed, got %v", cOutcome.State)
	}
}

func TestScheduler_ConcurrentStepTracking(t *testing.T) {
	// verify concurrent running count doesn't exceed limit
	workflow := &compiler.CompiledWorkflow{
		WorkflowID: "wf1",
		Steps: []compiler.CompiledStep{
			{StepID: "A", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "B", WorkflowID: "wf1", DependsOn: []string{}},
			{StepID: "C", WorkflowID: "wf1", DependsOn: []string{}},
		},
		ConcurrencyLimit: 1,
	}

	plan := &compiler.CompiledPlan{Workflows: []compiler.CompiledWorkflow{*workflow}}
	mock := newMockExecutor()
	mock.delay = 20 * time.Millisecond

	var maxConcurrent int32
	var currentConcurrent int32

	// wrap mock to track concurrency
	wrappedMock := &trackedMockExecutor{
		inner:             mock,
		currentConcurrent: &currentConcurrent,
		maxConcurrent:     &maxConcurrent,
	}

	sched := NewScheduler(plan, wrappedMock)
	result, err := sched.Run(context.Background(), "wf1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", result.OverallState)
	}
	if atomic.LoadInt32(&maxConcurrent) > 1 {
		t.Fatalf("exceeded concurrency limit: max concurrent = %d", atomic.LoadInt32(&maxConcurrent))
	}
}

type trackedMockExecutor struct {
	inner             *mockExecutor
	currentConcurrent *int32
	maxConcurrent     *int32
}

func (t *trackedMockExecutor) Execute(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error) {
	atomic.AddInt32(t.currentConcurrent, 1)
	defer atomic.AddInt32(t.currentConcurrent, -1)

	for {
		curr := atomic.LoadInt32(t.currentConcurrent)
		maxVal := atomic.LoadInt32(t.maxConcurrent)
		if curr <= maxVal {
			break
		}
		if atomic.CompareAndSwapInt32(t.maxConcurrent, maxVal, curr) {
			break
		}
	}

	return t.inner.Execute(ctx, step, execCtx)
}
