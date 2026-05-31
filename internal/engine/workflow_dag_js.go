//go:build js

package engine

import (
	goctx "context"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
)

// runWorkflowDAGJS runs workflows one at a time on the calling goroutine for WASM.
func runWorkflowDAGJS(
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
	finished := make(map[string]bool, n)

	pending := make([]int, n)
	for i := range pending {
		pending[i] = i
	}
	completed := 0

	for completed < n {
		var toRemove []int
		progress := false

		for _, i := range pending {
			cwf := plan.Workflows[i]

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
			if !allPrereqDone {
				continue
			}

			toRemove = append(toRemove, i)
			progress = true

			if anyPrereqFailed {
				emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] workflow skip  workflowID=%s  reason=dependency-failed", cwf.WorkflowID)
				outcomes[i] = WorkflowOutcome{
					WorkflowID: cwf.WorkflowID,
					Passed:     false,
					Skipped:    true,
				}
				finished[cwf.WorkflowID] = false
				completed++
				continue
			}

			wfStart := time.Now()
			collector.Emit(diagnostics.EventWorkflowStarted, cwf.WorkflowID, "", "", "")
			collector.StartWorkflow(cwf.WorkflowID)

			wo := executeWorkflow(ctx, cwf, sched, s, traceID, cfg, collector, sb, scriptSem, seed)
			wo.StartedAt = wfStart
			wo.Duration = time.Since(wfStart)

			outcomeStr := "pass"
			if !wo.Passed {
				outcomeStr = "fail"
			}
			collector.Emit(diagnostics.EventWorkflowCompleted, cwf.WorkflowID, "", "", outcomeStr)

			outcomes[i] = wo
			finished[cwf.WorkflowID] = wo.Passed
			completed++
		}

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

		if !progress {
			for _, i := range pending {
				cwf := plan.Workflows[i]
				outcomes[i] = WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false, Skipped: true}
				finished[cwf.WorkflowID] = false
				completed++
			}
		}
	}

	return outcomes
}
