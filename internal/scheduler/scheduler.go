package scheduler

import (
	"context"
	"fmt"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/executor"
	"github.com/tractl/tractl/internal/runtime"
)

// Scheduler runs a single workflow DAG using a provided executor.
type Scheduler struct {
	executor executor.Executor
	plan     *compiler.CompiledPlan
}

// NewScheduler creates a Scheduler from a compiled plan and an executor.
func NewScheduler(plan *compiler.CompiledPlan, exec executor.Executor) *Scheduler {
	if plan == nil || exec == nil {
		panic("Scheduler: plan and executor must not be nil")
	}
	return &Scheduler{executor: exec, plan: plan}
}

// Run executes a single workflow by ID using a fresh execution context.
func (s *Scheduler) Run(ctx context.Context, workflowID string) (*WorkflowResult, error) {
	execCtx, err := runtime.NewExecutionContext(workflowID, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("scheduler: create execution context: %w", err)
	}
	return s.RunWithContext(ctx, workflowID, execCtx)
}

// RunWithContext runs a single workflow using a caller-supplied ExecutionContext.
// This allows the engine to pre-populate spec/workflow/env variables before
// handing off to the scheduler.
func (s *Scheduler) RunWithContext(ctx context.Context, workflowID string, execCtx *runtime.ExecutionContext) (*WorkflowResult, error) {
	var workflow *compiler.CompiledWorkflow
	for _, w := range s.plan.Workflows {
		if w.WorkflowID == workflowID {
			workflow = &w
			break
		}
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}
	return s.runWorkflow(ctx, workflow, execCtx)
}

// buildDependents builds a map of stepID → direct dependents (steps that depend on it).
func buildDependents(steps []compiler.CompiledStep) map[string][]string {
	out := make(map[string][]string)
	for _, st := range steps {
		for _, d := range st.DependsOn {
			out[d] = append(out[d], st.StepID)
		}
	}
	return out
}

// transitiveDependents returns the set of step IDs that transitively depend on root.
// Excludes root itself.
func transitiveDependents(root string, dependents map[string][]string) map[string]struct{} {
	out := make(map[string]struct{})
	stack := append([]string{}, dependents[root]...)
	for len(stack) > 0 {
		n := len(stack) - 1
		cur := stack[n]
		stack = stack[:n]
		if _, seen := out[cur]; seen {
			continue
		}
		out[cur] = struct{}{}
		stack = append(stack, dependents[cur]...)
	}
	return out
}
