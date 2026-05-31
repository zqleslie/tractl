//go:build !js

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/runtime"
)

func (s *Scheduler) runWorkflow(
	ctx context.Context,
	workflow *compiler.CompiledWorkflow,
	execCtx *runtime.ExecutionContext,
) (*WorkflowResult, error) {
	started := time.Now()

	// set concurrency limit: user overrides via concurrency: in workflow spec, else DefaultStepConcurrency
	concLimit := workflow.ConcurrencyLimit
	if concLimit <= 0 {
		concLimit = planner.DefaultStepConcurrency
	}

	// structured concurrency: semaphore and errgroup
	sem := make(chan struct{}, concLimit)
	g, ctx := errgroup.WithContext(ctx)

	// done channel: buffered to exactly number of steps so goroutines never block on send
	done := make(chan string, len(workflow.Steps))

	// dependents map for failFast transitive cancellation
	dependents := buildDependents(workflow.Steps)

	// per-step cancellation funcs (failFast targets the transitive subgraph only)
	var stepCancelMu sync.Mutex
	stepCancels := make(map[string]context.CancelFunc, len(workflow.Steps))

	// per-step wait duration: time from dispatch to semaphore acquisition
	var waitMu sync.Mutex
	stepWaits := make(map[string]time.Duration, len(workflow.Steps))

	// tracking: pending steps, completed steps
	pending := make([]*compiler.CompiledStep, len(workflow.Steps))
	for i := range workflow.Steps {
		pending[i] = &workflow.Steps[i]
	}
	completed := make([]string, 0, len(workflow.Steps))

	// main scheduling loop
	for len(completed) < len(workflow.Steps) {
		// re-scan pending for eligible steps
		var toRemove []*compiler.CompiledStep
		for _, step := range pending {
			if !execCtx.IsEligible(step.StepID, step.DependsOn) {
				continue
			}

			toRemove = append(toRemove, step)

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

			// per-step cancellation context derived from the workflow group context
			stepCtx, stepCancel := context.WithCancel(ctx)
			stepCancelMu.Lock()
			stepCancels[step.StepID] = stepCancel
			stepCancelMu.Unlock()

			// dispatch step as goroutine
			step := step // capture loop variable
			dispatchedAt := time.Now()
			g.Go(func() error {
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
					stepCancel()
					done <- step.StepID
				}()

				waitStart := dispatchedAt
				select {
				case sem <- struct{}{}:
				case <-stepCtx.Done():
					execCtx.SetStepState(step.StepID, runtime.StateCancelled)
					return nil
				}
				waitMu.Lock()
				stepWaits[step.StepID] = time.Since(waitStart)
				waitMu.Unlock()
				defer func() { <-sem }()

				execCtx.SetStepState(step.StepID, runtime.StateRunning)
				result, err := s.executor.Execute(stepCtx, step, execCtx)

				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						execCtx.SetStepState(step.StepID, runtime.StateCancelled)
					} else {
						execCtx.SetStepState(step.StepID, runtime.StateFailed)
					}
				} else if result != nil {
					execCtx.SetStepState(step.StepID, result.State)
					if result.State == runtime.StateFailed && workflow.FailurePolicy == "failFast" {
						// ADR-008 §5: cancel only transitive dependents, not independent branches.
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

				return nil
			})
		}

		// remove processed steps from pending
		for _, toRm := range toRemove {
			for i, p := range pending {
				if p.StepID == toRm.StepID {
					pending = append(pending[:i], pending[i+1:]...)
					break
				}
			}
		}

	drainDone:
		for {
			select {
			case stepID := <-done:
				completed = append(completed, stepID)
			default:
				break drainDone
			}
		}

		if len(pending) == 0 {
			break
		}

		cancelled := false
		select {
		case stepID := <-done:
			completed = append(completed, stepID)
		case <-ctx.Done():
			for _, step := range pending {
				execCtx.SetStepState(step.StepID, runtime.StateCancelled)
			}
			cancelled = true
		}
		if cancelled {
			break
		}
	}

	_ = g.Wait()

	for {
		select {
		case <-done:
		default:
			goto buildResult
		}
	}

buildResult:
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
