// pipeline.go wires the full parse → validate → plan → compile → schedule → assert
// pipeline. runSpec is called by Run and RunDocument after format detection and
// overlay application are complete. executeWorkflow runs a single workflow's step DAG
// and is also called by the workflow DAG dispatcher in workflow_dag_*.go.

package engine

import (
	goctx "context"
	"fmt"
	"runtime"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/executor"
	"github.com/tractl/tractl/internal/planner"
	rt "github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/sandbox"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
	"github.com/tractl/tractl/internal/validation"
)

func (e *Engine) runSpec(s *spec.TraCtlSpec, cfg Config) *RunResult {
	collector := diagnostics.NewCollector(cfg.TraceID)

	// Resolve script semaphore cap and create semaphore channel.
	maxScripts := cfg.MaxConcurrentScripts
	if maxScripts <= 0 {
		maxScripts = runtime.NumCPU()
	}
	scriptSem := make(chan struct{}, maxScripts)

	// Create a ScriptRunner reused across hooks in this Run.
	sb := cfg.ScriptRunner
	if sb == nil {
		sb = sandbox.NewSandbox(0)
	}

	// ── STAGE 3: canonical validation ────────────────────────────────────────

	validationErrs := validation.NewSpecValidator().Validate(s)
	if len(validationErrs) > 0 {
		return &RunResult{ValidationError: joinValidationErrors(validationErrs)}
	}
	collector.Emit(diagnostics.EventValidationCompleted, "", "", "", "")

	// ── STAGE 4: plan + compile ───────────────────────────────────────────────

	pl := cfg.Planner
	if pl == nil {
		pl = planner.NewPlanner(planner.DefaultRegistry())
	}
	execPlan, planErr := pl.Plan(s)
	if planErr != nil {
		return &RunResult{PlanError: planErr.Error()}
	}
	collector.Emit(diagnostics.EventPlanningCompleted, "", "", "", "")

	comp := cfg.Compiler
	if comp == nil {
		comp = compiler.NewCompiler()
	}
	compiledPlan, compileErr := comp.Compile(execPlan, s)
	if compileErr != nil {
		return &RunResult{PlanError: compileErr.Error()}
	}
	collector.Emit(diagnostics.EventCompilationCompleted, "", "", "", "")

	// ── STAGE 5: execute via parallel workflow DAG + per-workflow step DAG ──────

	traceID := cfg.TraceID
	if traceID == "" {
		traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}

	emitEngineLog(collector, cfg, "", "", "[engine] traceID=%s  workflows=%d  totalSteps=%d  workflowConcurrency=%d",
		traceID, len(compiledPlan.Workflows), execPlan.Trace.TotalSteps, compiledPlan.WorkflowConcurrency)

	// seed derived once per Run: hash cfg.TraceID to int64
	seed := traceIDToSeed(cfg.TraceID)

	httpExec := executor.NewHTTPExecutor(nil)
	sched := scheduler.NewScheduler(compiledPlan, httpExec)

	workflowOutcomes := runWorkflowDAG(goctx.Background(), compiledPlan, sched, s, traceID, cfg, collector, sb, scriptSem, seed)

	// ── STAGE 6: assemble RunResult ───────────────────────────────────────────

	allPassed := true
	for _, wo := range workflowOutcomes {
		if !wo.Passed {
			allPassed = false
			break
		}
	}

	result := &RunResult{
		Passed:    allPassed,
		Workflows: workflowOutcomes,
		verbose:   cfg.Verbose,
	}
	result.Diagnostics = buildDiagnosticsRecord(collector, result)
	return result
}

// runHook executes a single script hook if non-nil.
// Returns an error if the script fails or requests cancellation.
// Applies the resulting MutationSet to execCtx at the given scope.
// seed is used for deterministic tractl APIs.
func runHook(ctx goctx.Context, sb ScriptRunner, script *spec.Script, execCtx *rt.ExecutionContext, scope rt.Scope, stepID string, seed int64, sem chan struct{}) error {
	if script == nil {
		return nil
	}
	if script.Language != "" && script.Language != "js" {
		return &sandbox.SandboxError{
			Code:    sandbox.ErrLanguageUnsupported,
			Message: fmt.Sprintf("language %q is not supported; only js is admitted in schema v1", script.Language),
		}
	}
	if script.SourceRef != "" {
		return &sandbox.SandboxError{
			Code:    sandbox.ErrSourceRefUnsupported,
			Message: "sourceRef is not supported in Phase 9; use inline source",
		}
	}
	// Acquire a script execution slot, respecting context cancellation.
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return ctx.Err()
	}
	ms, err := sb.Execute(script.Source, execCtx.Freeze(), seed)
	if err != nil {
		return err
	}
	if ms.Empty() {
		return nil
	}
	if scope == rt.ScopeStep && stepID != "" {
		execCtx.SetActiveStepID(stepID)
		defer execCtx.ClearActiveStepID()
	}
	return ms.Apply(execCtx, scope)
}

// buildDiagnosticsRecord converts RunResult outcome data into WorkflowRecord and
// StepRecord entries on the Collector, then returns the assembled ExecutionRecord.
// Called after runWorkflowDAG returns, once the timing fields are populated.
func buildDiagnosticsRecord(collector *diagnostics.Collector, result *RunResult) *diagnostics.ExecutionRecord {
	for _, wf := range result.Workflows {
		collector.StartWorkflow(wf.WorkflowID) // idempotent: no-op if already started
		for _, s := range wf.Steps {
			collector.AddStep(wf.WorkflowID, stepOutcomeToDiagnostics(s, wf.StartedAt))
		}
		outcomeStr := "pass"
		if !wf.Passed {
			if wf.Skipped {
				outcomeStr = "skipped"
			} else {
				outcomeStr = "fail"
			}
		}
		collector.CompleteWorkflow(wf.WorkflowID, outcomeStr, wf.Duration)
	}
	rec := collector.Record()
	return &rec
}

// stepOutcomeToDiagnostics converts one StepOutcome to a diagnostics.StepRecord.
func stepOutcomeToDiagnostics(s StepOutcome, wfStart time.Time) diagnostics.StepRecord {
	// Map engine State strings to the diagnostics outcome vocabulary.
	outcomeStr := s.State
	switch s.State {
	case string(rt.StateSucceeded):
		outcomeStr = "pass"
	case string(rt.StateFailed):
		outcomeStr = "fail"
	case string(rt.StateDependencySkipped), string(rt.StateConditionalSkip), string(rt.StateCancelled):
		outcomeStr = "dependency-skipped"
	}

	rec := diagnostics.StepRecord{
		StepID:       s.StepID,
		StartedAt:    s.StartedAt,
		Duration:     s.Duration,
		DependsOn:    s.DependsOn,
		WaitDuration: s.WaitDuration,
		Outcome:      outcomeStr,
	}
	if rec.StartedAt.IsZero() {
		rec.StartedAt = wfStart
	}

	// Build RequestRecord if response data is available.
	if s.ResponseStatus != 0 {
		rr := diagnostics.RequestRecord{
			Attempt: 1,
			Method:  s.RequestMethod,
			URL:     s.RequestURL,
			Status:  s.ResponseStatus,
		}
		if t := s.RequestTimeline; t != nil {
			rr.Timeline = diagnostics.DurationToTimingRecord(
				t.DNSResolution,
				t.TCPConnect,
				t.TLSHandshake,
				t.RequestSent,
				t.TimeToFirstByte,
				t.ResponseTransfer,
				t.TotalDuration,
			)
		}
		rec.Requests = []diagnostics.RequestRecord{rr}
	}

	// Record all assertion results; failures carry a message, passes do not.
	for _, ar := range s.AssertionResults {
		outcome := "pass"
		if ar.CausesFailure {
			outcome = "fail"
		}
		rec.Assertions = append(rec.Assertions, diagnostics.AssertionRecord{
			ID:            ar.AssertionID,
			Kind:          string(ar.Kind),
			Outcome:       outcome,
			Severity:      string(ar.Severity),
			CausesFailure: ar.CausesFailure,
			Message:       ar.Message,
		})
	}
	if s.Error != "" {
		rec.Assertions = append(rec.Assertions, diagnostics.AssertionRecord{
			Outcome:       "fail",
			Severity:      "error",
			CausesFailure: true,
			Message:       s.Error,
		})
	}

	return rec
}
