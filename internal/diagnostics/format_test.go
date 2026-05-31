package diagnostics

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// ── FormatRecordText ──────────────────────────────────────────────────────────

func TestFormatRecordText_EmptyRecord(t *testing.T) {
	r := ExecutionRecord{TraceID: "t1"}
	out := FormatRecordText(r)
	if !strings.Contains(out, "Execution Record") {
		t.Error("expected 'Execution Record' in output")
	}
	if !strings.Contains(out, "trace: t1") {
		t.Error("expected 'trace: t1' in output")
	}
}

func TestFormatRecordText_NoTraceID(t *testing.T) {
	out := FormatRecordText(ExecutionRecord{})
	if strings.Contains(out, "[trace:") {
		t.Error("output must not contain '[trace:' when TraceID is empty")
	}
}

func TestFormatRecordText_WorkflowHeader(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "auth-flow", Outcome: "pass", Duration: 450 * time.Millisecond},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "Workflow: auth-flow") {
		t.Error("expected 'Workflow: auth-flow'")
	}
	if !strings.Contains(out, "[pass]") {
		t.Error("expected '[pass]'")
	}
	if !strings.Contains(out, "450ms") {
		t.Error("expected '450ms'")
	}
}

func TestFormatRecordText_SkippedWorkflowAnnotation(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "skipped"},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "(dependency-skipped)") {
		t.Error("expected '(dependency-skipped)'")
	}
}

func TestFormatRecordText_StepTimingLine(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{
				WorkflowID: "wf1",
				Outcome:    "pass",
				Steps: []StepRecord{
					{
						StepID:   "step-a",
						Duration: 245 * time.Millisecond,
						Outcome:  "pass",
						Requests: []RequestRecord{
							{
								Attempt: 1,
								Timeline: TimingRecord{
									DNSMs:  5,
									TCPMs:  10,
									TLSMs:  30,
									TTFBMs: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	out := FormatRecordText(r)
	for _, want := range []string{"step-a", "245ms", "pass", "DNS:5ms", "TTFB:80ms"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output", want)
		}
	}
}

func TestFormatRecordText_NoRequestsStep(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{
				WorkflowID: "wf1",
				Outcome:    "pass",
				Steps: []StepRecord{
					{StepID: "step-b", Duration: 10 * time.Millisecond, Outcome: "pass"},
				},
			},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "(no request)") {
		t.Error("expected '(no request)' for step with no Requests")
	}
}

func TestFormatRecordText_WaterfallPresent(t *testing.T) {
	now := time.Now()
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{
				WorkflowID: "wf1",
				Outcome:    "pass",
				StartedAt:  now,
				Duration:   300 * time.Millisecond,
				Steps: []StepRecord{
					{
						StepID:    "step-a",
						StartedAt: now,
						Duration:  100 * time.Millisecond,
						Outcome:   "pass",
					},
					{
						StepID:    "step-b",
						StartedAt: now.Add(100 * time.Millisecond),
						Duration:  200 * time.Millisecond,
						Outcome:   "pass",
					},
				},
			},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "Waterfall:") {
		t.Error("expected 'Waterfall:' for workflow with 2 steps")
	}
	if !strings.Contains(out, "█") {
		t.Error("expected '█' execution bar character in waterfall")
	}
}

func TestFormatRecordText_WaterfallOmittedForSingleStep(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{
				WorkflowID: "wf1",
				Outcome:    "pass",
				Steps: []StepRecord{
					{StepID: "only-step", Duration: 50 * time.Millisecond, Outcome: "pass"},
				},
			},
		},
	}
	out := FormatRecordText(r)
	if strings.Contains(out, "Waterfall:") {
		t.Error("Waterfall must be omitted for single-step workflow")
	}
}

func TestFormatRecordText_ProvenancePresent(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
		},
		Events: []TraceEvent{
			{MonotonicMs: 10, Kind: EventStepStarted, WorkflowID: "wf1", StepID: "step-a"},
			{MonotonicMs: 200, Kind: EventStepCompleted, WorkflowID: "wf1", StepID: "step-a"},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "Provenance:") {
		t.Error("expected 'Provenance:'")
	}
	if !strings.Contains(out, "+10ms") {
		t.Error("expected '+10ms'")
	}
	if !strings.Contains(out, "+200ms") {
		t.Error("expected '+200ms'")
	}
}

func TestFormatRecordText_ProvenanceOmittedWhenEmpty(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
		},
		Events: nil,
	}
	out := FormatRecordText(r)
	if strings.Contains(out, "Provenance:") {
		t.Error("Provenance must be omitted when r.Events is nil")
	}
}

func TestFormatRecordText_PipelineEventsUnderFirstWorkflow(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
			{WorkflowID: "wf2", Outcome: "pass"},
		},
		Events: []TraceEvent{
			{MonotonicMs: 5, Kind: EventPlanningCompleted, WorkflowID: ""},
		},
	}
	out := FormatRecordText(r)
	// The planning.completed event should appear exactly once.
	count := strings.Count(out, string(EventPlanningCompleted))
	if count != 1 {
		t.Errorf("pipeline event should appear exactly once, got %d", count)
	}
}

func TestFormatRecordText_FooterDuration(t *testing.T) {
	r := ExecutionRecord{Duration: 500 * time.Millisecond}
	out := FormatRecordText(r)
	if !strings.Contains(out, "Total duration: 500ms") {
		t.Error("expected 'Total duration: 500ms'")
	}
}

func TestFormatRecordText_FooterOmittedWhenZero(t *testing.T) {
	r := ExecutionRecord{Duration: 0}
	out := FormatRecordText(r)
	if strings.Contains(out, "Total duration:") {
		t.Error("Total duration must be omitted when r.Duration is 0")
	}
}

// TestFormatRecordText_ProvenanceAlreadySorted verifies sortTraceEvents is a
// no-op (no panics, correct output) when events arrive in ascending order.
func TestFormatRecordText_ProvenanceAlreadySorted(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
		},
		Events: []TraceEvent{
			{MonotonicMs: 1, Kind: EventStepStarted, WorkflowID: "wf1", StepID: "s1"},
			{MonotonicMs: 2, Kind: EventStepCompleted, WorkflowID: "wf1", StepID: "s1"},
			{MonotonicMs: 3, Kind: EventWorkflowCompleted, WorkflowID: "wf1"},
		},
	}
	out := FormatRecordText(r)
	if !strings.Contains(out, "+1ms") || !strings.Contains(out, "+3ms") {
		t.Errorf("expected sorted provenance lines, got:\n%s", out)
	}
	// Verify order: +1ms must appear before +3ms in the output.
	pos1 := strings.Index(out, "+1ms")
	pos3 := strings.Index(out, "+3ms")
	if pos1 >= pos3 {
		t.Errorf("provenance lines out of order: +1ms at %d, +3ms at %d", pos1, pos3)
	}
}

// TestFormatRecordText_ProvenanceMergedAndSorted verifies that workflow events
// and pipeline events are merged and re-sorted correctly.
func TestFormatRecordText_ProvenanceMergedAndSorted(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
		},
		Events: []TraceEvent{
			{MonotonicMs: 50, Kind: EventWorkflowStarted, WorkflowID: "wf1"},
			{MonotonicMs: 10, Kind: EventPlanningCompleted, WorkflowID: ""}, // pipeline: earlier
			{MonotonicMs: 20, Kind: EventCompilationCompleted, WorkflowID: ""},
		},
	}
	out := FormatRecordText(r)
	// All three events must appear (pipeline events under first workflow).
	if !strings.Contains(out, "+10ms") {
		t.Errorf("expected '+10ms' in output")
	}
	if !strings.Contains(out, "+50ms") {
		t.Errorf("expected '+50ms' in output")
	}
	// +10ms must come before +50ms.
	p10 := strings.Index(out, "+10ms")
	p50 := strings.Index(out, "+50ms")
	if p10 >= p50 {
		t.Errorf("merged events not in ascending order: +10ms at %d, +50ms at %d", p10, p50)
	}
}

// ── FormatRecordJSON ──────────────────────────────────────────────────────────

func TestFormatRecordJSON_ValidJSON(t *testing.T) {
	b, err := FormatRecordJSON(ExecutionRecord{TraceID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestFormatRecordJSON_ContainsTraceID(t *testing.T) {
	b, err := FormatRecordJSON(ExecutionRecord{TraceID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Accept either "TraceID" or "traceID" depending on struct tags.
	if !strings.Contains(string(b), "t1") {
		t.Error("expected TraceID value 't1' in JSON output")
	}
}

func TestFormatRecordJSON_RoundTrip(t *testing.T) {
	original := ExecutionRecord{
		TraceID: "round-trip-id",
		Workflows: []WorkflowRecord{
			{WorkflowID: "wf1", Outcome: "pass"},
		},
	}
	b, err := FormatRecordJSON(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded ExecutionRecord
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.TraceID != original.TraceID {
		t.Errorf("TraceID: got %q, want %q", decoded.TraceID, original.TraceID)
	}
	if len(decoded.Workflows) != len(original.Workflows) {
		t.Errorf("Workflows length: got %d, want %d", len(decoded.Workflows), len(original.Workflows))
	}
}

// ── FormatRecordYAML ──────────────────────────────────────────────────────────

func TestFormatRecordYAML_ValidYAML(t *testing.T) {
	b, err := FormatRecordYAML(ExecutionRecord{TraceID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(b, &m); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
}

func TestFormatRecordYAML_ContainsWorkflow(t *testing.T) {
	r := ExecutionRecord{
		Workflows: []WorkflowRecord{
			{WorkflowID: "my-workflow", Outcome: "pass"},
		},
	}
	b, err := FormatRecordYAML(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(b), "my-workflow") {
		t.Error("expected workflow ID 'my-workflow' in YAML output")
	}
}
