//go:build js

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// runWorkflowJS executes workflow steps sequentially on the calling goroutine.
// Go WASM cannot safely combine errgroup dispatch with blocking fetch callbacks.
func (s *Scheduler) runWorkflowJS(
	ctx context.Context,
	workflow *compiler.CompiledWorkflow,
	execCtx *runtime.ExecutionContext,
) (*WorkflowResult, error) {
	started := time.Now()
	dependents := buildDependents(workflow.Steps)

	var stepCancelMu sync.Mutex
	stepCancels := make(map[string]context.CancelFunc, len(workflow.Steps))
	stepWaits := make(map[string]time.Duration, len(workflow.Steps))

	pending := make([]*compiler.CompiledStep, len(workflow.Steps))
	for i := range workflow.Steps {
		pending[i] = &workflow.Steps[i]
	}
	completed := make([]string, 0, len(workflow.Steps))

	for len(completed) < len(workflow.Steps) {
		if ctx.Err() != nil {
			for _, step := range pending {
				execCtx.SetStepState(step.StepID, runtime.StateCancelled)
				completed = append(completed, step.StepID)
			}
			break
		}

		var toRemove []*compiler.CompiledStep
		progress := false

		for _, step := range pending {
			if !execCtx.IsEligible(step.StepID, step.DependsOn) {
				continue
			}

			toRemove = append(toRemove, step)
			progress = true

			if !execCtx.ShouldExecute(step.StepID, step.DependsOn) {
				execCtx.SetStepState(step.StepID, runtime.StateDependencySkipped)
				completed = append(completed, step.StepID)
				continue
			}

			if step.When != "" {
				resolver := runtime.NewVariableResolver(execCtx)
				eval := runtime.NewExpressionEvaluator(resolver, execCtx)
				result, _ := eval.Evaluate(ctx, step.When)
				if result == "" || result == "false" {
					execCtx.SetStepState(step.StepID, runtime.StateConditionalSkip)
					completed = append(completed, step.StepID)
					continue
				}
			}

			stepCtx, stepCancel := context.WithCancel(ctx)
			stepCancelMu.Lock()
			stepCancels[step.StepID] = stepCancel
			stepCancelMu.Unlock()

			dispatchedAt := time.Now()
			func() {
				defer stepCancel()
				defer func() {
					if r := recover(); r != nil {
						execCtx.SetStepState(step.StepID, runtime.StateFailed)
						execCtx.SetStepResult(&runtime.StepResult{
							StepID:     step.StepID,
							WorkflowID: step.WorkflowID,
							State:      runtime.StateFailed,
							StartedAt:  time.Now(),
							FinishedAt: time.Now(),
							Extracts:   map[string]string{},
							Error:      fmt.Errorf("panic: %v", r),
						})
					}
				}()

				stepWaits[step.StepID] = time.Since(dispatchedAt)
				execCtx.SetStepState(step.StepID, runtime.StateRunning)
				result, err := s.executor.Execute(stepCtx, step, execCtx)

				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						execCtx.SetStepState(step.StepID, runtime.StateCancelled)
					} else {
						execCtx.SetStepState(step.StepID, runtime.StateFailed)
					}
					return
				}
				if result != nil {
					execCtx.SetStepState(step.StepID, result.State)
					if result.State == runtime.StateFailed && workflow.FailurePolicy == "failFast" {
						deps := transitiveDependents(step.StepID, dependents)
						stepCancelMu.Lock()
						for depID := range deps {
							if cancel, ok := stepCancels[depID]; ok {
								cancel()
							}
						}
						stepCancelMu.Unlock()
					}
				}
			}()
			completed = append(completed, step.StepID)
		}

		for _, toRm := range toRemove {
			for i, p := range pending {
				if p.StepID == toRm.StepID {
					pending = append(pending[:i], pending[i+1:]...)
					break
				}
			}
		}

		if !progress {
			break
		}
	}

	finished := time.Now()
	outcomes := make([]StepOutcome, 0, len(workflow.Steps))
	overallState := runtime.StateSucceeded
	hasFailed := false
	hasCancelled := false

	for _, step := range workflow.Steps {
		state := execCtx.GetStepState(step.StepID)
		result, _ := execCtx.GetStepResult(step.StepID)

		outcomes = append(outcomes, StepOutcome{
			StepID:       step.StepID,
			State:        state,
			Result:       result,
			WaitDuration: stepWaits[step.StepID],
		})

		switch state {
		case runtime.StateFailed:
			hasFailed = true
		case runtime.StateCancelled:
			hasCancelled = true
		case runtime.StatePending, runtime.StateRunning, runtime.StateSucceeded,
			runtime.StateConditionalSkip, runtime.StateDependencySkipped:
		}
	}

	if hasFailed {
		overallState = runtime.StateFailed
	} else if hasCancelled {
		overallState = runtime.StateCancelled
	}

	return &WorkflowResult{
		WorkflowID:   workflow.WorkflowID,
		ExecutionID:  execCtx.ExecutionID(),
		Outcomes:     outcomes,
		OverallState: overallState,
		StartedAt:    started,
		FinishedAt:   finished,
	}, nil
}
