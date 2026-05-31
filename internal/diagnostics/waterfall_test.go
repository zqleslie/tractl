package diagnostics

import (
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ── BuildWaterfall tests ──────────────────────────────────────────────────────

func TestBuildWaterfall_EmptyWorkflow(t *testing.T) {
	wf := WorkflowRecord{WorkflowID: "wf-empty", StartedAt: t0, Duration: 300 * time.Millisecond}
	w := BuildWaterfall(wf)
	if w.TotalMs != 300 {
		t.Errorf("TotalMs: got %d, want 300", w.TotalMs)
	}
	if len(w.Entries) != 0 {
		t.Errorf("Entries: got %d, want 0", len(w.Entries))
	}
}

func TestBuildWaterfall_SingleStep(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-single",
		StartedAt:  t0,
		Duration:   100 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-a", StartedAt: t0, Duration: 100 * time.Millisecond, WaitDuration: 0, Outcome: "pass"},
		},
	}
	w := BuildWaterfall(wf)
	if len(w.Entries) != 1 {
		t.Fatalf("Entries: got %d, want 1", len(w.Entries))
	}
	e := w.Entries[0]
	if e.StartOffsetMs != 0 {
		t.Errorf("StartOffsetMs: got %d, want 0", e.StartOffsetMs)
	}
	if e.WaitMs != 0 {
		t.Errorf("WaitMs: got %d, want 0", e.WaitMs)
	}
	if e.ExecutionMs != 100 {
		t.Errorf("ExecutionMs: got %d, want 100", e.ExecutionMs)
	}
	if e.TotalMs != 100 {
		t.Errorf("TotalMs: got %d, want 100", e.TotalMs)
	}
	if e.Outcome != "pass" {
		t.Errorf("Outcome: got %q, want %q", e.Outcome, "pass")
	}
}

func TestBuildWaterfall_SequentialSteps(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-seq",
		StartedAt:  t0,
		Duration:   300 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-a", StartedAt: t0, Duration: 100 * time.Millisecond, WaitDuration: 0, DependsOn: nil},
			{StepID: "step-b", StartedAt: t0.Add(100 * time.Millisecond), Duration: 200 * time.Millisecond, WaitDuration: 20 * time.Millisecond, DependsOn: []string{"step-a"}},
		},
	}
	w := BuildWaterfall(wf)
	if len(w.Entries) != 2 {
		t.Fatalf("Entries: got %d, want 2", len(w.Entries))
	}
	a := w.Entries[0]
	if a.StepID != "step-a" || a.StartOffsetMs != 0 || a.ExecutionMs != 100 || a.WaitMs != 0 {
		t.Errorf("step-a: got offset=%d exec=%d wait=%d", a.StartOffsetMs, a.ExecutionMs, a.WaitMs)
	}
	b := w.Entries[1]
	if b.StepID != "step-b" || b.StartOffsetMs != 100 || b.ExecutionMs != 180 || b.WaitMs != 20 {
		t.Errorf("step-b: got offset=%d exec=%d wait=%d", b.StartOffsetMs, b.ExecutionMs, b.WaitMs)
	}
}

func TestBuildWaterfall_ParallelSteps(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-par",
		StartedAt:  t0,
		Duration:   150 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-x", StartedAt: t0, Duration: 150 * time.Millisecond, WaitDuration: 0},
			{StepID: "step-y", StartedAt: t0, Duration: 100 * time.Millisecond, WaitDuration: 0},
		},
	}
	w := BuildWaterfall(wf)
	if len(w.Entries) != 2 {
		t.Fatalf("Entries: got %d, want 2", len(w.Entries))
	}
	for _, e := range w.Entries {
		if e.StartOffsetMs != 0 {
			t.Errorf("%s: StartOffsetMs: got %d, want 0", e.StepID, e.StartOffsetMs)
		}
	}
}

func TestBuildWaterfall_WaitDuration(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-wait",
		StartedAt:  t0,
		Duration:   80 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-w", StartedAt: t0, Duration: 80 * time.Millisecond, WaitDuration: 30 * time.Millisecond},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.WaitMs != 30 {
		t.Errorf("WaitMs: got %d, want 30", e.WaitMs)
	}
	if e.ExecutionMs != 50 {
		t.Errorf("ExecutionMs: got %d, want 50", e.ExecutionMs)
	}
	if e.TotalMs != 80 {
		t.Errorf("TotalMs: got %d, want 80", e.TotalMs)
	}
}

func TestBuildWaterfall_SkippedStep(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-skip",
		StartedAt:  t0,
		Duration:   0,
		Steps: []StepRecord{
			{StepID: "step-s", StartedAt: t0, Duration: 0, WaitDuration: 0, Outcome: "dependency-skipped"},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.TotalMs != 0 || e.ExecutionMs != 0 || e.WaitMs != 0 {
		t.Errorf("skipped step: got TotalMs=%d ExecMs=%d WaitMs=%d", e.TotalMs, e.ExecutionMs, e.WaitMs)
	}
	if e.Outcome != "dependency-skipped" {
		t.Errorf("Outcome: got %q, want %q", e.Outcome, "dependency-skipped")
	}
}

func TestBuildWaterfall_SortedByStartOffset(t *testing.T) {
	// Steps added in reverse order (C, B, A) — entries must be sorted A, B, C.
	wf := WorkflowRecord{
		WorkflowID: "wf-sort",
		StartedAt:  t0,
		Duration:   300 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-c", StartedAt: t0.Add(200 * time.Millisecond), Duration: 100 * time.Millisecond},
			{StepID: "step-b", StartedAt: t0.Add(100 * time.Millisecond), Duration: 100 * time.Millisecond},
			{StepID: "step-a", StartedAt: t0, Duration: 100 * time.Millisecond},
		},
	}
	w := BuildWaterfall(wf)
	want := []string{"step-a", "step-b", "step-c"}
	for i, id := range want {
		if w.Entries[i].StepID != id {
			t.Errorf("Entries[%d].StepID: got %q, want %q", i, w.Entries[i].StepID, id)
		}
	}
}

func TestBuildWaterfall_TimingFromFirstRequest(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-timing",
		StartedAt:  t0,
		Duration:   30 * time.Millisecond,
		Steps: []StepRecord{
			{
				StepID:    "step-t",
				StartedAt: t0,
				Duration:  30 * time.Millisecond,
				Requests: []RequestRecord{
					{Timeline: TimingRecord{DNSMs: 5, TCPMs: 10, TLSMs: 15, TotalMs: 30}},
				},
			},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.DNSMs != 5 {
		t.Errorf("DNSMs: got %d, want 5", e.DNSMs)
	}
	if e.TCPMs != 10 {
		t.Errorf("TCPMs: got %d, want 10", e.TCPMs)
	}
	if e.TLSMs != 15 {
		t.Errorf("TLSMs: got %d, want 15", e.TLSMs)
	}
}

func TestBuildWaterfall_NoRequestsZeroTiming(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-noreq",
		StartedAt:  t0,
		Duration:   50 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-nr", StartedAt: t0, Duration: 50 * time.Millisecond},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.DNSMs != 0 || e.TCPMs != 0 || e.TLSMs != 0 || e.RequestSentMs != 0 || e.TTFBMs != 0 || e.TransferMs != 0 {
		t.Errorf("expected all timing fields 0, got DNS=%d TCP=%d TLS=%d Sent=%d TTFB=%d Transfer=%d",
			e.DNSMs, e.TCPMs, e.TLSMs, e.RequestSentMs, e.TTFBMs, e.TransferMs)
	}
}

func TestBuildWaterfall_DependsOnNotNil(t *testing.T) {
	wf := WorkflowRecord{
		WorkflowID: "wf-deps",
		StartedAt:  t0,
		Duration:   50 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-nd", StartedAt: t0, Duration: 50 * time.Millisecond, DependsOn: nil},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.DependsOn == nil {
		t.Error("DependsOn: got nil, want empty slice (so JSON renders [] not null)")
	}
	if len(e.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(e.DependsOn))
	}
}

func TestBuildWaterfall_StartOffsetClampedToZero(t *testing.T) {
	// Synthetic clock skew: step starts 5ms before the workflow StartedAt.
	wf := WorkflowRecord{
		WorkflowID: "wf-clamp",
		StartedAt:  t0,
		Duration:   50 * time.Millisecond,
		Steps: []StepRecord{
			{StepID: "step-ck", StartedAt: t0.Add(-5 * time.Millisecond), Duration: 50 * time.Millisecond},
		},
	}
	w := BuildWaterfall(wf)
	e := w.Entries[0]
	if e.StartOffsetMs != 0 {
		t.Errorf("StartOffsetMs: got %d, want 0 (clamped)", e.StartOffsetMs)
	}
}

// ── WaterfallText tests ───────────────────────────────────────────────────────

func waterfallWithSteps(steps []StepRecord, totalDur time.Duration) Waterfall {
	wf := WorkflowRecord{
		WorkflowID: "test-wf",
		StartedAt:  t0,
		Duration:   totalDur,
		Steps:      steps,
	}
	return BuildWaterfall(wf)
}

func TestWaterfallText_EmptyWorkflow(t *testing.T) {
	w := BuildWaterfall(WorkflowRecord{WorkflowID: "empty-wf", StartedAt: t0, Duration: 300 * time.Millisecond})
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "Workflow:") {
		t.Error("output missing 'Workflow:' header")
	}
	if !strings.Contains(out, "empty-wf") {
		t.Error("output missing workflow ID")
	}
}

func TestWaterfallText_ContainsStepIDs(t *testing.T) {
	steps := []StepRecord{
		{StepID: "login-step", StartedAt: t0, Duration: 50 * time.Millisecond},
		{StepID: "data-step", StartedAt: t0.Add(50 * time.Millisecond), Duration: 50 * time.Millisecond},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "login-step") {
		t.Error("output missing 'login-step'")
	}
	if !strings.Contains(out, "data-step") {
		t.Error("output missing 'data-step'")
	}
}

func TestWaterfallText_ContainsDurationAndOutcome(t *testing.T) {
	steps := []StepRecord{
		{StepID: "check", StartedAt: t0, Duration: 245 * time.Millisecond, Outcome: "pass"},
	}
	w := waterfallWithSteps(steps, 245*time.Millisecond)
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "245ms") {
		t.Error("output missing '245ms'")
	}
	if !strings.Contains(out, "pass") {
		t.Error("output missing 'pass'")
	}
}

func TestWaterfallText_DependencyAnnotation(t *testing.T) {
	steps := []StepRecord{
		{StepID: "main-step", StartedAt: t0.Add(50 * time.Millisecond), Duration: 100 * time.Millisecond,
			WaitDuration: 10 * time.Millisecond, DependsOn: []string{"auth-step"}},
	}
	w := waterfallWithSteps(steps, 150*time.Millisecond)
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "←") {
		t.Error("output missing '←' dependency arrow")
	}
	if !strings.Contains(out, "auth-step") {
		t.Error("output missing dep step ID 'auth-step'")
	}
}

func TestWaterfallText_NoDependencyAnnotation(t *testing.T) {
	steps := []StepRecord{
		{StepID: "solo", StartedAt: t0, Duration: 100 * time.Millisecond, DependsOn: []string{}},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 60)
	if strings.Contains(out, "←") {
		t.Error("output must not contain '←' for step with no dependencies")
	}
}

func TestWaterfallText_LegendPresent(t *testing.T) {
	w := BuildWaterfall(WorkflowRecord{WorkflowID: "wf-leg", StartedAt: t0, Duration: 50 * time.Millisecond})
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "█") {
		t.Error("output missing '█' in legend")
	}
	if !strings.Contains(out, "·") {
		t.Error("output missing '·' in legend")
	}
}

func TestWaterfallText_BarWidth_DefaultsTo60(t *testing.T) {
	steps := []StepRecord{
		{StepID: "w-step", StartedAt: t0, Duration: 100 * time.Millisecond, Outcome: "pass"},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 0) // width=0 → defaults to 60

	// Find the bar between '[' and ']' on the step line.
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "w-step") {
			continue
		}
		open := strings.Index(line, "[")
		closeIdx := strings.Index(line, "]")
		if open < 0 || closeIdx < 0 || closeIdx <= open {
			t.Fatalf("no valid bar brackets found in line: %q", line)
		}
		barContent := []rune(line[open+1 : closeIdx])
		if len(barContent) != 60 {
			t.Errorf("bar width: got %d, want 60; bar=%q", len(barContent), string(barContent))
		}
		break
	}
}

func TestWaterfallText_ZeroTotalMs_NoPanic(t *testing.T) {
	// Must not panic or divide by zero.
	steps := []StepRecord{
		{StepID: "zero-step", StartedAt: t0, Duration: 0, Outcome: "skipped"},
	}
	w := Waterfall{
		WorkflowID: "wf-zero",
		TotalMs:    0,
		Entries:    BuildWaterfall(WorkflowRecord{WorkflowID: "wf-zero", StartedAt: t0, Duration: 0, Steps: steps}).Entries,
	}
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "zero-step") {
		t.Error("output missing step ID")
	}
	if !strings.Contains(out, "skipped") {
		t.Error("output missing outcome")
	}
}

func TestWaterfallText_WaitCharsPresent(t *testing.T) {
	steps := []StepRecord{
		{StepID: "waiter", StartedAt: t0, Duration: 100 * time.Millisecond,
			WaitDuration: 50 * time.Millisecond, Outcome: "pass"},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "·") {
		t.Error("output missing '·' wait bar character")
	}
}

func TestWaterfallText_ExecutionCharsPresent(t *testing.T) {
	steps := []StepRecord{
		{StepID: "runner", StartedAt: t0, Duration: 100 * time.Millisecond,
			WaitDuration: 0, Outcome: "pass"},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 60)
	if !strings.Contains(out, "█") {
		t.Error("output missing '█' execution bar character")
	}
}

func TestWaterfallText_ParallelStepsSameLine(t *testing.T) {
	// Two parallel steps both start at T0 (StartOffsetMs=0).
	// Their bars should both start at position 0 within the brackets.
	steps := []StepRecord{
		{StepID: "par-a", StartedAt: t0, Duration: 100 * time.Millisecond, Outcome: "pass"},
		{StepID: "par-b", StartedAt: t0, Duration: 80 * time.Millisecond, Outcome: "pass"},
	}
	w := waterfallWithSteps(steps, 100*time.Millisecond)
	out := WaterfallText(w, 60)

	// Collect bar starts for each parallel step row.
	barStarts := map[string]int{}
	for _, line := range strings.Split(out, "\n") {
		for _, id := range []string{"par-a", "par-b"} {
			if strings.Contains(line, id) {
				open := strings.Index(line, "[")
				if open < 0 {
					continue
				}
				// Find first '█' rune position after '['.
				barRunes := []rune(line[open+1:])
				for ri, r := range barRunes {
					if r == '█' {
						barStarts[id] = ri
						break
					}
				}
			}
		}
	}
	if len(barStarts) != 2 {
		t.Fatalf("could not find bar starts for both parallel steps; found: %v", barStarts)
	}
	if barStarts["par-a"] != barStarts["par-b"] {
		t.Errorf("parallel steps should start at same bar column: par-a=%d par-b=%d",
			barStarts["par-a"], barStarts["par-b"])
	}
}

// ── buildBar clamping / overflow branches ─────────────────────────────────────

// TestBuildBar_OverflowClamped verifies that when leadingSpaces + waitChars +
// execChars exceeds width, the bar is trimmed to exactly width characters.
func TestBuildBar_OverflowClamped(t *testing.T) {
	// Step starts late (large offset) with long execution: forces overflow.
	e := WaterfallEntry{
		StepID:        "big-step",
		StartOffsetMs: 90,
		WaitMs:        0,
		ExecutionMs:   80,
		TotalMs:       80,
		Outcome:       "pass",
	}
	bar := buildBar(e, 100, 10) // width=10; offset ratio=9, exec ratio=8 → overflow
	runes := []rune(bar)
	if len(runes) != 10 {
		t.Errorf("bar width after overflow clamp: got %d, want 10", len(runes))
	}
}

// TestBuildBar_WaitOverflowClamped forces the case where clamping exec to zero
// still leaves total > width, so waitChars is also trimmed.
func TestBuildBar_WaitOverflowClamped(t *testing.T) {
	e := WaterfallEntry{
		StepID:        "wait-heavy",
		StartOffsetMs: 5,
		WaitMs:        90,
		ExecutionMs:   90,
		TotalMs:       180,
		Outcome:       "pass",
	}
	bar := buildBar(e, 100, 5) // very narrow width forces heavy clamping
	runes := []rune(bar)
	if len(runes) != 5 {
		t.Errorf("bar width with wait overflow clamp: got %d, want 5", len(runes))
	}
}

// TestBuildBar_DeepOverflowWaitClamped forces the path where excess exceeds
// execChars (making it negative), and the shortfall is large enough to also
// make waitChars negative — verifying leadingSpaces absorbs the remainder.
//
// Arithmetic (totalMs=10, width=5):
//
//	leadingSpaces = 16*5/10 = 8
//	waitChars     =  4*5/10 = 2
//	execChars     =  1*5/10 = 0 → bumped to 1 (min-visibility)
//	total=11 > 5; excess=6
//	execChars = 1-6 = -5  → waitChars += -5 = -3 < 0  (line 178 hit)
//	leadingSpaces += -3 = 5; waitChars = 0; execChars = 0
//	result: 5 spaces (exactly width)
func TestBuildBar_DeepOverflowWaitClamped(t *testing.T) {
	e := WaterfallEntry{
		StepID:        "deep-overflow",
		StartOffsetMs: 16,
		WaitMs:        4,
		ExecutionMs:   1,
		TotalMs:       21,
		Outcome:       "pass",
	}
	bar := buildBar(e, 10, 5)
	runes := []rune(bar)
	if len(runes) != 5 {
		t.Errorf("bar width: got %d, want 5", len(runes))
	}
	for _, r := range runes {
		if r != ' ' {
			t.Errorf("expected all spaces when offset absorbs full width, got %q", string(r))
		}
	}
}

// TestBuildBar_ZeroExecMs verifies a step with wait but no active execution
// renders only wait characters (·) and the bar is exactly width wide.
func TestBuildBar_ZeroExecutionMs(t *testing.T) {
	e := WaterfallEntry{
		StepID:        "pure-wait",
		StartOffsetMs: 0,
		WaitMs:        100,
		ExecutionMs:   0,
		TotalMs:       100,
		Outcome:       "pass",
	}
	bar := buildBar(e, 100, 20)
	runes := []rune(bar)
	if len(runes) != 20 {
		t.Errorf("bar width: got %d, want 20", len(runes))
	}
	for _, r := range runes {
		if r != '·' && r != ' ' {
			t.Errorf("expected only '·' or ' ' for zero-exec step, got %q", string(r))
		}
	}
}
