// Package diagnostics assembles and formats execution observability records.
// Formatters in this file are credential-safe by construction; they operate on
// ExecutionRecord values whose sensitive fields are already masked by the engine.
// TOON output is not implemented; use YAML as a structurally identical alternative.
package diagnostics

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// headerRule is the 33-character U+2501 divider used in FormatRecordText headers.
const headerRule = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

// recordWaterfallWidth is the bar-area width passed to WaterfallText inside
// FormatRecordText. 54 chars leaves room for the 4-space indent in terminal output.
const recordWaterfallWidth = 54

// FormatRecordText renders an ExecutionRecord as human-readable terminal output.
// Sections are separated by blank lines and appear in the following order:
// header block, per-workflow step timing + waterfall + provenance, footer duration.
func FormatRecordText(r ExecutionRecord) string {
	var sb strings.Builder

	// 1. Header block.
	sb.WriteString(headerRule + "\n")
	if r.TraceID != "" {
		fmt.Fprintf(&sb, " Execution Record  [trace: %s]\n", r.TraceID)
	} else {
		sb.WriteString(" Execution Record\n")
	}
	sb.WriteString(headerRule + "\n")

	for wfIdx, wf := range r.Workflows {
		sb.WriteByte('\n')

		// 2a. Workflow header.
		outcomeLabel := wf.Outcome
		if outcomeLabel == "" {
			outcomeLabel = "unknown"
		}
		line := fmt.Sprintf("Workflow: %s  [%s]  %dms", wf.WorkflowID, outcomeLabel, wf.Duration.Milliseconds())
		if wf.Outcome == "skipped" {
			line += "  (dependency-skipped)"
		}
		sb.WriteString(line + "\n")

		// 2b. Step timing table.
		if len(wf.Steps) > 0 {
			labelWidth := 0
			for _, step := range wf.Steps {
				if l := len(step.StepID); l > labelWidth {
					labelWidth = l
				}
			}
			labelWidth += 2

			for _, step := range wf.Steps {
				isSkipped := step.Outcome == "skipped" || step.Outcome == "dependency-skipped"

				var timingPart string
				if isSkipped || len(step.Requests) == 0 {
					timingPart = "(no request)"
				} else {
					t := step.Requests[0].Timeline
					timingPart = fmt.Sprintf("DNS:%dms TCP:%dms TLS:%dms TTFB:%dms",
						t.DNSMs, t.TCPMs, t.TLSMs, t.TTFBMs)
				}

				fmt.Fprintf(&sb, "  %-*s  %6dms  %-20s  %s\n",
					labelWidth, step.StepID,
					step.Duration.Milliseconds(),
					step.Outcome,
					timingPart,
				)
			}
		}

		// 2c. Dependency waterfall (only if ≥ 2 steps).
		if len(wf.Steps) >= 2 {
			sb.WriteByte('\n')
			sb.WriteString("  Waterfall:\n")
			wfText := WaterfallText(BuildWaterfall(wf), recordWaterfallWidth)
			for _, wfLine := range strings.Split(strings.TrimRight(wfText, "\n"), "\n") {
				sb.WriteString("    ")
				sb.WriteString(wfLine)
				sb.WriteByte('\n')
			}
		}

		// 2d. Provenance trace.
		// Pipeline-level events (WorkflowID="") are shown once, under the first workflow only.
		if len(r.Events) > 0 {
			var wfEvents []TraceEvent
			for _, ev := range r.Events {
				if ev.WorkflowID == wf.WorkflowID {
					wfEvents = append(wfEvents, ev)
				} else if ev.WorkflowID == "" && wfIdx == 0 {
					wfEvents = append(wfEvents, ev)
				}
			}
			sortTraceEvents(wfEvents)

			if len(wfEvents) > 0 {
				sb.WriteByte('\n')
				sb.WriteString("  Provenance:\n")
				for _, ev := range wfEvents {
					third := ""
					if ev.StepID != "" {
						third = ev.StepID
					} else if ev.Detail != "" {
						third = ev.Detail
					}
					fmt.Fprintf(&sb, "    +%dms\t%-28s\t%s\n",
						ev.MonotonicMs, string(ev.Kind), third)
				}
			}
		}
	}

	// 3. Footer.
	if r.Duration != 0 {
		sb.WriteByte('\n')
		fmt.Fprintf(&sb, "Total duration: %dms\n", r.Duration.Milliseconds())
	}

	return sb.String()
}

// sortTraceEvents sorts a slice of TraceEvent by MonotonicMs ascending in-place.
func sortTraceEvents(events []TraceEvent) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].MonotonicMs < events[j].MonotonicMs
	})
}

// FormatRecordJSON returns r serialised as indented JSON (2-space indent).
// No additional transformation — ExecutionRecord is already a clean, serialisable struct.
func FormatRecordJSON(r ExecutionRecord) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// FormatRecordYAML returns r serialised as YAML using gopkg.in/yaml.v3.
// No additional transformation.
func FormatRecordYAML(r ExecutionRecord) ([]byte, error) {
	return yaml.Marshal(r)
}
