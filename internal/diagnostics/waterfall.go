package diagnostics

// Waterfall and WaterfallText implement the Dependency Waterfall visualisation
// defined in ADR-014 §4.
//
// WaterfallText output is for human display only; do not parse it
// programmatically. For machine-readable output, use Waterfall directly —
// it is JSON/YAML serialisable.

import (
	"fmt"
	"sort"
	"strings"
)

// Waterfall is the structured dependency waterfall for one workflow execution.
// It is serialisable to JSON and YAML for structured output (ADR-014 §4).
// Entries are ordered by StartOffsetMs ascending (actual execution order,
// not authoring order).
type Waterfall struct {
	WorkflowID string
	TotalMs    int64
	Entries    []WaterfallEntry
}

// WaterfallEntry is one row in the waterfall — one step.
// All timing values are milliseconds (int64) for serialisation friendliness.
type WaterfallEntry struct {
	StepID        string
	DependsOn     []string // step IDs this step waited for (nil → empty slice in output)
	StartOffsetMs int64    // elapsed ms from workflow start to step start; clamped to ≥ 0
	WaitMs        int64    // time spent waiting for dependency steps
	ExecutionMs   int64    // active execution time = TotalMs - WaitMs; clamped to ≥ 0
	TotalMs       int64    // total step duration
	Outcome       string   // "pass" | "fail" | "skipped" | "dependency-skipped"
	// Request timeline breakdown (from Requests[0] if present; zero otherwise).
	DNSMs         int64
	TCPMs         int64
	TLSMs         int64
	RequestSentMs int64
	TTFBMs        int64
	TransferMs    int64
	// Reserved: always 0 until Phase 7.2 engine wiring.
	AssertionMs int64
	ExtractMs   int64
}

// BuildWaterfall computes a Waterfall from a WorkflowRecord.
// Start offsets are derived from step.StartedAt.Sub(wf.StartedAt), clamped to ≥ 0.
// Entries are sorted by StartOffsetMs ascending.
func BuildWaterfall(wf WorkflowRecord) Waterfall {
	entries := make([]WaterfallEntry, 0, len(wf.Steps))

	for _, step := range wf.Steps {
		offsetMs := step.StartedAt.Sub(wf.StartedAt).Milliseconds()
		if offsetMs < 0 {
			offsetMs = 0
		}

		waitMs := step.WaitDuration.Milliseconds()
		totalMs := step.Duration.Milliseconds()
		execMs := totalMs - waitMs
		if execMs < 0 {
			execMs = 0
		}

		deps := step.DependsOn
		if deps == nil {
			deps = []string{}
		}

		entry := WaterfallEntry{
			StepID:        step.StepID,
			DependsOn:     deps,
			StartOffsetMs: offsetMs,
			WaitMs:        waitMs,
			ExecutionMs:   execMs,
			TotalMs:       totalMs,
			Outcome:       step.Outcome,
		}

		if len(step.Requests) > 0 {
			t := step.Requests[0].Timeline
			entry.DNSMs = t.DNSMs
			entry.TCPMs = t.TCPMs
			entry.TLSMs = t.TLSMs
			entry.RequestSentMs = t.RequestSentMs
			entry.TTFBMs = t.TTFBMs
			entry.TransferMs = t.TransferMs
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartOffsetMs < entries[j].StartOffsetMs
	})

	return Waterfall{
		WorkflowID: wf.WorkflowID,
		TotalMs:    wf.Duration.Milliseconds(),
		Entries:    entries,
	}
}

// WaterfallText renders a Waterfall as a human-readable ASCII string for CLI
// output (ADR-014 §4 "human-readable text form").
//
// width is the bar area width in characters. If ≤ 0, defaults to 60.
// Output is for human display only; do not parse it programmatically.
func WaterfallText(w Waterfall, width int) string {
	if width <= 0 {
		width = 60
	}

	// Compute label column width from longest step ID.
	maxLen := 0
	for _, e := range w.Entries {
		if len(e.StepID) > maxLen {
			maxLen = len(e.StepID)
		}
	}
	labelWidth := maxLen + 2

	var sb strings.Builder

	// Header.
	fmt.Fprintf(&sb, "Workflow: %s  (total: %dms)\n", w.WorkflowID, w.TotalMs)

	for _, e := range w.Entries {
		bar := buildBar(e, w.TotalMs, width)

		label := fmt.Sprintf("%-*s", labelWidth, e.StepID)
		dur := fmt.Sprintf("%6dms", e.TotalMs)

		line := fmt.Sprintf("%s[%s]  %s  %s", label, bar, dur, e.Outcome)

		if len(e.DependsOn) > 0 {
			line += "  ← " + strings.Join(e.DependsOn, " ")
		}

		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	sb.WriteString("Legend: █ = executing  · = waiting for dependencies\n")

	return sb.String()
}

// buildBar constructs the bar string of exactly `width` characters for one entry.
func buildBar(e WaterfallEntry, totalMs int64, width int) string {
	// Skipped steps or zero-duration workflows get a blank bar.
	if totalMs == 0 || e.TotalMs == 0 {
		return strings.Repeat(" ", width)
	}

	leadingSpaces := int(e.StartOffsetMs * int64(width) / totalMs)
	waitChars := int(e.WaitMs * int64(width) / totalMs)
	execChars := int(e.ExecutionMs * int64(width) / totalMs)

	// Ensure minimum 1 char visibility when the duration is non-zero.
	if e.WaitMs > 0 && waitChars == 0 {
		waitChars = 1
	}
	if e.ExecutionMs > 0 && execChars == 0 {
		execChars = 1
	}

	// Clamp total to width — trim execChars first, then waitChars.
	total := leadingSpaces + waitChars + execChars
	if total > width {
		excess := total - width
		execChars -= excess
		if execChars < 0 {
			waitChars += execChars // waitChars reduced by the shortfall
			execChars = 0
			if waitChars < 0 {
				leadingSpaces += waitChars
				waitChars = 0
				if leadingSpaces < 0 {
					leadingSpaces = 0
				}
			}
		}
	}

	var bar strings.Builder
	bar.Grow(width)
	bar.WriteString(strings.Repeat(" ", leadingSpaces))
	bar.WriteString(strings.Repeat("·", waitChars))
	bar.WriteString(strings.Repeat("█", execChars))

	// Pad trailing spaces to exactly width.
	used := leadingSpaces + waitChars + execChars
	if used < width {
		bar.WriteString(strings.Repeat(" ", width-used))
	}

	return bar.String()
}
