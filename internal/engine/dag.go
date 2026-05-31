// dag.go contains workflow step-DAG dispatch helpers extracted from pipeline.go.
// pipeline.go remains the top-level orchestrator; dag.go handles graph mechanics.

package engine

import (
	goctx "context"
	"strings"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	rt "github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
)

// executeWorkflow runs a single workflow's step DAG and returns its outcome.
func executeWorkflow(
	ctx goctx.Context,
	cwf compiler.CompiledWorkflow,
	sched *scheduler.Scheduler,
	s *spec.TraCtlSpec,
	traceID string,
	cfg Config,
	collector *diagnostics.Collector,
	sb ScriptRunner,
	scriptSem chan struct{},
	seed int64,
) WorkflowOutcome {
	wfDeps := "none"
	if len(cwf.DependsOn) > 0 {
		wfDeps = strings.Join(cwf.DependsOn, ", ")
	}
	emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] workflow start  workflowID=%s  steps=%d  concurrencyLimit=%d  dependsOn=[%s]",
		cwf.WorkflowID, len(cwf.Steps), cwf.ConcurrencyLimit, wfDeps)

	// Log topo-sorted step DAG order before execution.
	emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[dag] workflow=%s  execution order (topo-sorted):", cwf.WorkflowID)
	for idx, st := range cwf.Steps {
		deps := "none"
		if len(st.DependsOn) > 0 {
			deps = strings.Join(st.DependsOn, ", ")
		}
		emitEngineLog(collector, cfg, cwf.WorkflowID, st.StepID, "[dag]   [%d] step=%s  dependsOn=[%s]  concurrencySlot=%d",
			idx+1, st.StepID, deps, st.ConcurrencySlot)
	}

	// Build execution context with spec/workflow/env variables pre-populated.
	specVars := make(map[string]string, len(s.Variables))
	for k, v := range s.Variables {
		specVars[k] = v
	}
	wfVars := make(map[string]string)
	var wfSpec *spec.Workflow
	for i, wf := range s.Workflows {
		if wf.ID == cwf.WorkflowID {
			for k, v := range wf.Variables {
				wfVars[k] = v
			}
			wfSpec = &s.Workflows[i]
			break
		}
	}
	envVars := make(map[string]string) // Phase 4+: populated from cfg.EnvName

	execCtx, err := rt.NewExecutionContext(cwf.WorkflowID, specVars, wfVars, envVars)
	if err != nil {
		emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] create execution context error workflowID=%s err=%v", cwf.WorkflowID, err)
		return WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false}
	}

	// workflow.beforeAll
	if wfSpec != nil && wfSpec.Hooks != nil && wfSpec.Hooks.BeforeAll != nil {
		if err := runHook(ctx, sb, wfSpec.Hooks.BeforeAll, execCtx, rt.ScopeWorkflow, "", seed, scriptSem); err != nil {
			emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] beforeAll hook error workflowID=%s err=%v", cwf.WorkflowID, err)
			return WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false}
		}
	}

	// workflow.beforeEach runs once per step before the current scheduler call.
	if wfSpec != nil && wfSpec.Hooks != nil && wfSpec.Hooks.BeforeEach != nil {
		for _, cs := range cwf.Steps {
			if err := runHook(ctx, sb, wfSpec.Hooks.BeforeEach, execCtx, rt.ScopeStep, cs.StepID, seed, scriptSem); err != nil {
				emitEngineLog(collector, cfg, cwf.WorkflowID, cs.StepID, "[engine] beforeEach hook error workflowID=%s stepID=%s err=%v", cwf.WorkflowID, cs.StepID, err)
				return WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false}
			}
		}
	}

	wfResult, schedErr := sched.RunWithContext(ctx, cwf.WorkflowID, execCtx)
	if schedErr != nil {
		emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] workflow error  workflowID=%s  err=%v", cwf.WorkflowID, schedErr)
		return WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false}
	}

	// Step-level hooks require scheduler-level interleaving; this preserves the
	// existing post-execution approximation until the scheduler owns hook timing.
	if wfSpec != nil && len(wfSpec.Steps) > 0 {
		stepSpecByID := make(map[string]*spec.Step, len(wfSpec.Steps))
		for i := range wfSpec.Steps {
			stepSpecByID[wfSpec.Steps[i].ID] = &wfSpec.Steps[i]
		}
		for _, outcome := range wfResult.Outcomes {
			stepSpec, ok := stepSpecByID[outcome.StepID]
			if !ok || stepSpec.Hooks == nil {
				continue
			}
			if stepSpec.Hooks.BeforeStep != nil {
				if err := runHook(ctx, sb, stepSpec.Hooks.BeforeStep, execCtx, rt.ScopeStep, outcome.StepID, seed, scriptSem); err != nil {
					emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[engine] beforeStep hook error workflowID=%s stepID=%s err=%v", cwf.WorkflowID, outcome.StepID, err)
				}
			}
			if stepSpec.Hooks.AfterStep != nil && outcome.Result != nil {
				if err := runHook(ctx, sb, stepSpec.Hooks.AfterStep, execCtx, rt.ScopeStep, outcome.StepID, seed, scriptSem); err != nil {
					emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[engine] afterStep hook error workflowID=%s stepID=%s err=%v", cwf.WorkflowID, outcome.StepID, err)
				}
			}
			if stepSpec.Hooks.Transform != nil && outcome.Result != nil {
				if err := runHook(ctx, sb, stepSpec.Hooks.Transform, execCtx, rt.ScopeStep, outcome.StepID, seed, scriptSem); err != nil {
					emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[engine] transform hook error workflowID=%s stepID=%s err=%v", cwf.WorkflowID, outcome.StepID, err)
				}
			}
		}
	}

	// workflow.afterEach
	if wfSpec != nil && wfSpec.Hooks != nil && wfSpec.Hooks.AfterEach != nil {
		for _, outcome := range wfResult.Outcomes {
			if outcome.Result != nil {
				if err := runHook(ctx, sb, wfSpec.Hooks.AfterEach, execCtx, rt.ScopeStep, outcome.StepID, seed, scriptSem); err != nil {
					emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[engine] afterEach hook error workflowID=%s stepID=%s err=%v", cwf.WorkflowID, outcome.StepID, err)
				}
			}
		}
	}

	// workflow.afterAll errors, including cancellation, are workflow failures.
	if wfSpec != nil && wfSpec.Hooks != nil && wfSpec.Hooks.AfterAll != nil {
		if err := runHook(ctx, sb, wfSpec.Hooks.AfterAll, execCtx, rt.ScopeWorkflow, "", seed, scriptSem); err != nil {
			emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] afterAll hook error workflowID=%s err=%v", cwf.WorkflowID, err)
			return WorkflowOutcome{WorkflowID: cwf.WorkflowID, Passed: false}
		}
	}

	wo := buildWorkflowOutcome(cwf, wfResult, traceID, cfg, collector)
	emitEngineLog(collector, cfg, cwf.WorkflowID, "", "[engine] workflow end    workflowID=%s  passed=%v  elapsed=%s",
		cwf.WorkflowID, wo.Passed,
		wfResult.FinishedAt.Sub(wfResult.StartedAt).Round(time.Millisecond))
	return wo
}
