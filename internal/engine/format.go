// format.go defines code for the engine package.

package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tractl/tractl/internal/diagnostics"
)

// OutputFormatter formats a RunResult into bytes for a specific output format.
// Implement this interface to add a new output format without modifying existing code.
type OutputFormatter interface {
	Format(result *RunResult) ([]byte, error)
}

// FormatterFor returns the OutputFormatter for the given format name.
// Supported formats: "text", "json". Returns an error for unknown formats.
func FormatterFor(format string) (OutputFormatter, error) {
	switch format {
	case "text", "":
		return textFormatter{}, nil
	case "json":
		return jsonFormatter{}, nil
	default:
		return nil, fmt.Errorf("engine: unknown output format %q: supported formats are text, json", format)
	}
}

type textFormatter struct{}

func (textFormatter) Format(result *RunResult) ([]byte, error) {
	return []byte(FormatText(result)), nil
}

type jsonFormatter struct{}

func (jsonFormatter) Format(result *RunResult) ([]byte, error) {
	return FormatJSON(result)
}

// FormatText returns a concise human-readable summary of a RunResult.
// Targets developer terminal output; uses no external dependencies.
func FormatText(r *RunResult) string {
	var sb strings.Builder
	const divider = "──────────────────────────────────────"

	sb.WriteString("traCtl run\n")
	sb.WriteString(divider + "\n")

	// Pipeline errors.
	if r.ParseError != "" {
		sb.WriteString("Parse error: ")
		sb.WriteString(r.ParseError)
		sb.WriteString("\n")
		sb.WriteString(divider + "\n")
		sb.WriteString("FAIL  pipeline error\n")
		return sb.String()
	}
	if r.ValidationError != "" {
		sb.WriteString("Validation error: ")
		sb.WriteString(r.ValidationError)
		sb.WriteString("\n")
		sb.WriteString(divider + "\n")
		sb.WriteString("FAIL  pipeline error\n")
		return sb.String()
	}
	if r.PlanError != "" {
		sb.WriteString("Plan error: ")
		sb.WriteString(r.PlanError)
		sb.WriteString("\n")
		sb.WriteString(divider + "\n")
		sb.WriteString("FAIL  pipeline error\n")
		return sb.String()
	}

	// Workflow outcomes.
	totalSteps := 0
	failedSteps := 0
	skippedWFs := 0
	for _, wf := range r.Workflows {
		if wf.Skipped {
			skippedWFs++
			fmt.Fprintf(&sb, "workflow: %s  [skipped — dependency failed]\n", wf.WorkflowID)
			continue
		}
		fmt.Fprintf(&sb, "workflow: %s\n", wf.WorkflowID)
		for _, step := range wf.Steps {
			totalSteps++
			if step.CausesFailure {
				failedSteps++
				fmt.Fprintf(&sb, "  ✗ %-20s failed\n", step.StepID)
				for _, ar := range step.AssertionResults {
					if ar.CausesFailure {
						fmt.Fprintf(&sb, "    • %s: %s [%s]\n", ar.AssertionID, ar.Message, ar.Severity)
					}
				}
				if step.Error != "" {
					fmt.Fprintf(&sb, "    • error: %s\n", step.Error)
				}
			} else {
				fmt.Fprintf(&sb, "  ✓ %-20s %s\n", step.StepID, step.State)
			}
			// Verbose: show HTTP response detail below each step.
			if step.ResponseStatus != 0 {
				fmt.Fprintf(&sb, "    status: %d\n", step.ResponseStatus)
				if len(step.ResponseHeaders) > 0 {
					sb.WriteString("    headers:\n")
					// Stable sort for deterministic output.
					keys := make([]string, 0, len(step.ResponseHeaders))
					for k := range step.ResponseHeaders {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						fmt.Fprintf(&sb, "      %s: %s\n", k, step.ResponseHeaders[k])
					}
				}
				if step.ResponseBody != "" {
					fmt.Fprintf(&sb, "    body: %s\n", step.ResponseBody)
				}
			}
		}
	}

	sb.WriteString(divider + "\n")

	wfCount := len(r.Workflows)
	if r.Passed {
		fmt.Fprintf(&sb, "PASS  %d workflow(s), %d step(s) passed\n", wfCount, totalSteps)
	} else {
		if skippedWFs > 0 {
			fmt.Fprintf(&sb, "FAIL  %d workflow(s), %d step(s) failed, %d workflow(s) skipped\n", wfCount, failedSteps, skippedWFs)
		} else {
			fmt.Fprintf(&sb, "FAIL  %d workflow(s), %d step(s) failed\n", wfCount, failedSteps)
		}
	}

	const cliWaterfallWidth = 60
	if r.Diagnostics != nil {
		for _, wf := range r.Diagnostics.Workflows {
			if len(wf.Steps) < 2 {
				continue
			}
			sb.WriteByte('\n')
			sb.WriteString("Waterfall:\n")
			wfText := diagnostics.WaterfallText(diagnostics.BuildWaterfall(wf), cliWaterfallWidth)
			for _, wfLine := range strings.Split(strings.TrimRight(wfText, "\n"), "\n") {
				sb.WriteString(wfLine)
				sb.WriteByte('\n')
			}
		}
	}

	// Verbose diagnostics section — only rendered when --verbose was passed.
	if r.verbose && r.Diagnostics != nil {
		sb.WriteByte('\n')
		sb.WriteString(diagnostics.FormatRecordText(*r.Diagnostics))
	}

	return sb.String()
}

// FormatJSON returns the RunResult serialised as indented JSON (2-space indent).
func FormatJSON(r *RunResult) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
