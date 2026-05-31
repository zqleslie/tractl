package localapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/engine"
)

func mapRunResult(result *engine.RunResult) (*RunResult, error) {
	if result.ParseError != "" {
		return nil, fmt.Errorf("parse error: %s", result.ParseError)
	}
	if result.ValidationError != "" {
		return nil, fmt.Errorf("validation error: %s", result.ValidationError)
	}
	if result.PlanError != "" {
		return nil, fmt.Errorf("plan error: %s", result.PlanError)
	}
	if len(result.Workflows) == 0 {
		return nil, fmt.Errorf("run produced no workflow outcomes")
	}

	wf := result.Workflows[0]
	if wf.Skipped {
		return nil, fmt.Errorf("workflow skipped due to dependency failure")
	}
	if len(wf.Steps) == 0 {
		return nil, fmt.Errorf("workflow produced no step outcomes")
	}

	step := wf.Steps[0]
	passedCount := 0
	assertionRows := make([]AssertionResult, 0, len(step.AssertionResults))
	for _, ar := range step.AssertionResults {
		passed := ar.Outcome == assertion.OutcomePass
		if passed {
			passedCount++
		}
		assertionRows = append(assertionRows, AssertionResult{
			ID:       ar.AssertionID,
			Kind:     string(ar.Kind),
			Op:       "",
			Expected: "",
			Received: ar.Message,
			Passed:   passed,
			Severity: "error",
		})
	}

	timing := timingFromDiagnostics(result, step)

	statusLabel := http.StatusText(step.ResponseStatus)
	if statusLabel == "" && step.ResponseStatus > 0 {
		statusLabel = fmt.Sprintf("%d", step.ResponseStatus)
	}
	if step.ResponseStatus > 0 {
		statusLabel = fmt.Sprintf("%d %s", step.ResponseStatus, statusLabel)
	}

	contentType := step.ResponseHeaders["content-type"]
	if contentType == "" {
		contentType = "application/json"
	}

	extractRows := extractResultsFromDiagnostics(result)

	return &RunResult{
		StatusCode:       step.ResponseStatus,
		StatusText:       strings.TrimSpace(statusLabel),
		DurationMs:       timing.Total,
		Body:             step.ResponseBody,
		Headers:          responseHeaders(step.ResponseHeaders, contentType),
		Timing:           timing,
		AssertionResults: assertionRows,
		ExtractResults:   extractRows,
		AssertionsPassed: passedCount,
		AssertionsTotal:  len(step.AssertionResults),
		Error:            step.Error,
	}, nil
}

func responseHeaders(headers map[string]string, contentType string) map[string]string {
	out := make(map[string]string, len(headers)+1)
	for key, value := range headers {
		out[key] = value
	}
	if _, ok := out["content-type"]; !ok && contentType != "" {
		out["content-type"] = contentType
	}
	return out
}

func timingFromDiagnostics(result *engine.RunResult, step engine.StepOutcome) TimingResult {
	if result.Diagnostics != nil && len(result.Diagnostics.Workflows) > 0 {
		wf := result.Diagnostics.Workflows[0]
		if len(wf.Steps) > 0 && len(wf.Steps[0].Requests) > 0 {
			t := wf.Steps[0].Requests[0].Timeline
			return TimingResult{
				DNS:      t.DNSMs,
				TCP:      t.TCPMs,
				TLS:      t.TLSMs,
				TTFB:     t.TTFBMs,
				Transfer: t.TransferMs,
				Total:    t.TotalMs,
				Unit:     "ms",
			}
		}
	}

	totalMs := step.Duration.Milliseconds()
	if totalMs <= 0 {
		totalMs = 0
	}
	return TimingResult{Total: totalMs, Unit: "ms"}
}

func extractResultsFromDiagnostics(result *engine.RunResult) []ExtractResult {
	if result.Diagnostics == nil || len(result.Diagnostics.Workflows) == 0 {
		return []ExtractResult{}
	}
	wf := result.Diagnostics.Workflows[0]
	if len(wf.Steps) == 0 {
		return []ExtractResult{}
	}
	rows := make([]ExtractResult, 0, len(wf.Steps[0].Extracts))
	for _, ex := range wf.Steps[0].Extracts {
		variable := ex.As
		if variable == "" {
			variable = ex.ID
		}
		rows = append(rows, ExtractResult{
			ID:            ex.ID,
			VariableName:  variable,
			Scope:         "workflow",
			ResolvedValue: ex.Source,
		})
	}
	return rows
}
