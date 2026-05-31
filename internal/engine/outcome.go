// outcome.go converts scheduler step results into engine workflow outcomes.

package engine

import (
	"strings"
	"time"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/executor"
	rt "github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/scheduler"
)

// buildWorkflowOutcome converts a scheduler.WorkflowResult into an engine.WorkflowOutcome,
// evaluating assertion results from each step's StepResult and emitting DAG log lines.
func buildWorkflowOutcome(cwf compiler.CompiledWorkflow, wfResult *scheduler.WorkflowResult, traceID string, cfg Config, collector *diagnostics.Collector) WorkflowOutcome {
	assertEval := cfg.Evaluator
	if assertEval == nil {
		assertEval = assertion.New()
	}
	workflowStart := wfResult.StartedAt

	stepOutcomes := make([]StepOutcome, 0, len(wfResult.Outcomes))
	workflowFailed := false

	// Index the compiled assertions and DependsOn by stepID for fast lookup.
	assertionsByStep := make(map[string][]compiler.CompiledAssertion, len(cwf.Steps))
	dependsOnByStep := make(map[string][]string, len(cwf.Steps))
	for _, cs := range cwf.Steps {
		assertionsByStep[cs.StepID] = cs.Assertions
		dependsOnByStep[cs.StepID] = cs.DependsOn
	}

	for _, outcome := range wfResult.Outcomes {
		state := string(outcome.State)

		// Log every step completion with its final state.
		if outcome.Result != nil {
			emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[dag] step=%s  state=%s  http-status=%d  elapsed=%s",
				outcome.StepID, state, outcome.Result.Status,
				outcome.Result.FinishedAt.Sub(outcome.Result.StartedAt).Round(time.Millisecond))
		} else {
			emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[dag] step=%s  state=%s", outcome.StepID, state)
		}

		// Non-executed steps (skipped/cancelled) have no result to assert against.
		if outcome.Result == nil || outcome.State == rt.StateDependencySkipped ||
			outcome.State == rt.StateConditionalSkip || outcome.State == rt.StateCancelled {
			stepOutcomes = append(stepOutcomes, StepOutcome{
				StepID:    outcome.StepID,
				State:     state,
				DependsOn: dependsOnByStep[outcome.StepID],
			})
			continue
		}

		// Re-evaluate assertions from the compiled step against the executor's result.
		assertionResults := make([]assertion.AssertionResult, 0, len(assertionsByStep[outcome.StepID]))
		stepCausesFailure := false

		for _, a := range assertionsByStep[outcome.StepID] {
			ar, _, _ := assertEval.EvaluateCompiled(a, outcome.Result, nil, traceID, "", workflowStart)
			assertionResults = append(assertionResults, ar)
			if ar.CausesFailure {
				stepCausesFailure = true
				emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[dag] step=%s  assertion=%s  outcome=FAIL  severity=%s  msg=%q",
					outcome.StepID, ar.AssertionID, ar.Severity, ar.Message)
			} else {
				emitEngineLog(collector, cfg, cwf.WorkflowID, outcome.StepID, "[dag] step=%s  assertion=%s  outcome=%s",
					outcome.StepID, ar.AssertionID, ar.Outcome)
			}
		}

		// A step fails if the executor flagged it or an assertion caused failure.
		stepFailed := outcome.State == rt.StateFailed || stepCausesFailure || outcome.Error != nil
		if stepFailed {
			workflowFailed = true
			state = string(rt.StateFailed)
		}

		errMsg := ""
		if outcome.Error != nil {
			errMsg = outcome.Error.Error()
		}

		so := StepOutcome{
			StepID:           outcome.StepID,
			State:            state,
			AssertionResults: assertionResults,
			CausesFailure:    stepFailed,
			Error:            errMsg,
			DependsOn:        dependsOnByStep[outcome.StepID],
		}

		// Capture per-step timing from the executor result.
		if outcome.Result != nil {
			so.StartedAt = outcome.Result.StartedAt
			so.Duration = outcome.Result.FinishedAt.Sub(outcome.Result.StartedAt)
			so.WaitDuration = outcome.WaitDuration
			so.RequestTimeline = outcome.Result.Timeline
			so.RequestURL = executor.NewCredentialMasker().MaskURL(outcome.Result.Metadata["url"])
			so.RequestMethod = outcome.Result.Metadata["method"]
		}

		if outcome.Result != nil && cfg.Verbose {
			so.ResponseStatus = outcome.Result.Status
			so.ResponseHeaders = executor.NewCredentialMasker().MaskResponseHeaders(outcome.Result.Headers)
			so.ResponseBody = strings.ToValidUTF8(string(outcome.Result.Body), "?")
		}

		stepOutcomes = append(stepOutcomes, so)
	}

	return WorkflowOutcome{
		WorkflowID: cwf.WorkflowID,
		Passed:     !workflowFailed,
		Steps:      stepOutcomes,
	}
}
