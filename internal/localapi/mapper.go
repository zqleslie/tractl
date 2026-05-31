package localapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/engine"
)

func mapRunResult(result *engine.RunResult) (*RunResponse, error) {
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
			ID:     ar.AssertionID,
			Passed: passed,
			Label:  fmt.Sprintf("%s %s", ar.Kind, ar.AssertionID),
			Detail: ar.Message,
		})
	}

	headers := responseHeaders(step.ResponseHeaders)
	timeline, durationMs := timelineFromDiagnostics(result, step)

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

	return &RunResponse{
		Passed:           result.Passed,
		DurationMs:       durationMs,
		StatusCode:       step.ResponseStatus,
		StatusLabel:      strings.TrimSpace(statusLabel),
		ContentType:      contentType,
		Body:             step.ResponseBody,
		Headers:          headers,
		AssertionResults: assertionRows,
		ExtractResults:   extractRows,
		PassedCount:      passedCount,
		TotalCount:       len(step.AssertionResults),
		Timeline:         timeline,
		Error:            step.Error,
	}, nil
}

func responseHeaders(headers map[string]string) []HeaderRow {
	if len(headers) == 0 {
		return []HeaderRow{}
	}
	rows := make([]HeaderRow, 0, len(headers))
	i := 0
	for key, value := range headers {
		i++
		rows = append(rows, HeaderRow{
			ID:      fmt.Sprintf("response-header-%d", i),
			Enabled: true,
			Key:     key,
			Value:   value,
		})
	}
	return rows
}

func timelineFromDiagnostics(result *engine.RunResult, step engine.StepOutcome) ([]TimelineSegment, int64) {
	if result.Diagnostics != nil && len(result.Diagnostics.Workflows) > 0 {
		wf := result.Diagnostics.Workflows[0]
		if len(wf.Steps) > 0 && len(wf.Steps[0].Requests) > 0 {
			t := wf.Steps[0].Requests[0].Timeline
			return []TimelineSegment{
				{Label: "DNS", Ms: t.DNSMs},
				{Label: "TCP", Ms: t.TCPMs},
				{Label: "TLS", Ms: t.TLSMs},
				{Label: "TTFB", Ms: t.TTFBMs},
				{Label: "Transfer", Ms: t.TransferMs},
			}, t.TotalMs
		}
	}

	totalMs := step.Duration.Milliseconds()
	if totalMs <= 0 {
		return []TimelineSegment{}, 0
	}
	return []TimelineSegment{
		{Label: "Total", Ms: totalMs},
	}, totalMs
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
			Variable: "steps.this.extracts." + variable,
			Value:    ex.Source,
			Scope:    "workflow",
		})
	}
	return rows
}
