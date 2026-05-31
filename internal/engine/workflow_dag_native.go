//go:build !js

package engine

import (
	goctx "context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
)

// runWorkflowDAG executes all workflows in the compiled plan as a parallel DAG.
// Workflows with no dependsOn start immediately (subject to the concurrency cap).
// A workflow whose dependsOn workflows all passed starts as soon as a slot is free.
// If any prerequisite workflow failed, the dependent is dependency-skipped.
// Results are returned in declaration (topo-sort) order.
func runWorkflowDAG(
	ctx goctx.Context,
	plan *compiler.CompiledPlan,
	sched *scheduler.Scheduler,
	s *spec.TraCtlSpec,
	traceID string,
	cfg Config,
	collector *diagnostics.Collector,
	sb ScriptRunner,
	scriptSem chan struct{},
	seed int64,
) []WorkflowOutcome {
	n := len(plan.Workflows)
	outcomes := make([]WorkflowOutcome, n)

	// Semaphore limiting how many workflows run concurrently.
	sem := make(chan struct{}, plan.WorkflowConcurrency)

	// Track which workflow indices have finished and whether they passed.
	// Protected by a mutex because goroutines write, the dispatch loop reads.
	var mu sync.Mutex
	finished := make(map[string]bool) // workflowID → passed
	done := make(chan string, n)      // signals completion

	// Build index: workflowID → position in outcomes slice (for ordered output).
	idxByID := make(map[string]int, n)
	for i, cwf := range plan.Workflows {
		idxByID[cwf.WorkflowID] = i
	}

	g, gctx := errgroup.WithContext(ctx)

	// pending holds workflow indices not yet dispatched.
	pending := make([]int, n)
	for i := range pending {
		pending[i] = i
	}
	dispatched := make(map[string]bool, n)
	completed := 0

	for completed < n {
		var toRemove []int

		for _, i := range pending {
			cwf := plan.Workflows[i]

			// Check if all dependsOn have finished.
			mu.Lock()
			allPrereqDone := true
			anyPrereqFailed := false
			for _, depID := range cwf.DependsOn {
				passed, ok := finished[depID]
				if !ok {
					allPrereqDone = false
					break
				}
				if !passed {
					anyPrereqFailed = true
				}
			}
			mu.Unlock()

			if !allPrereqDone {
				continue
			}

			toRemove = append(toRemove, i)

			// Dependency-skip: a prerequisite failed, don't run this workflow.
			if anyPrereqFailed {
				emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] workflow skip  workflowID=%s  reason=dependency-failed", cwf.WorkflowID)
				outcomes[i] = WorkflowOutcome{
					WorkflowID: cwf.WorkflowID,
					Passed:     false,
					Skipped:    true,
				}
				mu.Lock()
				finished[cwf.WorkflowID] = false
				mu.Unlock()
				done <- cwf.WorkflowID
				dispatched[cwf.WorkflowID] = true
				continue
			}

			// Dispatch as goroutine.
			dispatched[cwf.WorkflowID] = true
			cwfCopy := cwf
			idx := i
			g.Go(func() error {
				// Acquire concurrency slot.
				select {
				case sem <- struct{}{}:
				case <-gctx.Done():
					outcomes[idx] = WorkflowOutcome{WorkflowID: cwfCopy.WorkflowID, Passed: false}
					mu.Lock()
					finished[cwfCopy.WorkflowID] = false
					mu.Unlock()
					done <- cwfCopy.WorkflowID
					return nil
				}
				defer func() { <-sem }()

				wfStart := time.Now()
				collector.Emit(diagnostics.EventWorkflowStarted, cwfCopy.WorkflowID, "", "", "")
				collector.StartWorkflow(cwfCopy.WorkflowID)

				wo := executeWorkflow(gctx, cwfCopy, sched, s, traceID, cfg, collector, sb, scriptSem, seed)
				wo.StartedAt = wfStart
				wo.Duration = time.Since(wfStart)

				outcomeStr := "pass"
				if !wo.Passed {
					outcomeStr = "fail"
				}
				collector.Emit(diagnostics.EventWorkflowCompleted, cwfCopy.WorkflowID, "", "", outcomeStr)

				outcomes[idx] = wo

				mu.Lock()
				finished[cwfCopy.WorkflowID] = wo.Passed
				mu.Unlock()
				done <- cwfCopy.WorkflowID
				return nil
			})
		}

		// Remove dispatched from pending.
		if len(toRemove) > 0 {
			toRemoveSet := make(map[int]bool, len(toRemove))
			for _, i := range toRemove {
				toRemoveSet[i] = true
			}
			next := pending[:0]
			for _, i := range pending {
				if !toRemoveSet[i] {
					next = append(next, i)
				}
			}
			pending = next
		}

		// If there are still pending workflows that can't start yet, wait for one completion.
		if len(pending) > 0 {
			select {
			case <-done:
				completed++
			case <-gctx.Done():
				// Mark remaining pending as skipped.
				for _, i := range pending {
					cwf := plan.Workflows[i]
					outcomes[i] = WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false, Skipped: true}
					mu.Lock()
					finished[cwf.WorkflowID] = false
					mu.Unlock()
				}
				completed = n
			}
		} else {
			// All workflows dispatched: drain remaining done signals.
			for completed < n {
				<-done
				completed++
			}
		}
	}

	// g.Wait() error intentionally not checked: each workflow goroutine writes
	// its outcome into the outcomes slice; the caller inspects outcomes directly.
	// The errgroup error would duplicate what is already captured.
	_ = g.Wait() // outcomes already carry workflow errors.
	return outcomes
}
