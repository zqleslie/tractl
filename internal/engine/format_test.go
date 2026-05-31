// format_test.go tests engine output formatting and formatter selection.

package engine_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/engine"
)

func TestFormatterFor_KnownFormats(t *testing.T) {
	for _, format := range []string{"text", "", "json"} {
		f, err := engine.FormatterFor(format)
		if err != nil {
			t.Errorf("FormatterFor(%q) unexpected error: %v", format, err)
		}
		if f == nil {
			t.Errorf("FormatterFor(%q) returned nil formatter", format)
		}
	}
}

func TestFormatterFor_UnknownFormat(t *testing.T) {
	_, err := engine.FormatterFor("xml")
	if err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}

func TestFormatText_ContainsWorkflowID(t *testing.T) {
	r := &engine.RunResult{
		Passed: true,
		Workflows: []engine.WorkflowOutcome{
			{WorkflowID: "my-workflow", Passed: true, Steps: []engine.StepOutcome{}},
		},
	}
	out := engine.FormatText(r)
	if !strings.Contains(out, "my-workflow") {
		t.Fatalf("FormatText output missing workflow ID:\n%s", out)
	}
}

func TestFormatText_ContainsStepID(t *testing.T) {
	r := &engine.RunResult{
		Passed: true,
		Workflows: []engine.WorkflowOutcome{
			{
				WorkflowID: "wf",
				Passed:     true,
				Steps: []engine.StepOutcome{
					{StepID: "my-step", State: "succeeded"},
				},
			},
		},
	}
	out := engine.FormatText(r)
	if !strings.Contains(out, "my-step") {
		t.Fatalf("FormatText output missing step ID:\n%s", out)
	}
}

func TestFormatText_ContainsAssertionMessage(t *testing.T) {
	r := &engine.RunResult{
		Passed: false,
		Workflows: []engine.WorkflowOutcome{
			{
				WorkflowID: "wf",
				Passed:     false,
				Steps: []engine.StepOutcome{
					{
						StepID:        "fail-step",
						State:         "failed",
						CausesFailure: true,
						AssertionResults: []assertion.AssertionResult{
							{
								AssertionID:   "assert-status",
								Message:       "expected status 200, got 404",
								CausesFailure: true,
							},
						},
					},
				},
			},
		},
	}
	out := engine.FormatText(r)
	if !strings.Contains(out, "expected status 200, got 404") {
		t.Fatalf("FormatText output missing assertion message:\n%s", out)
	}
}

func TestFormatText_PassAndFailSymbols(t *testing.T) {
	r := &engine.RunResult{
		Passed: false,
		Workflows: []engine.WorkflowOutcome{
			{
				WorkflowID: "wf",
				Passed:     false,
				Steps: []engine.StepOutcome{
					{StepID: "ok-step", State: "succeeded", CausesFailure: false},
					{StepID: "fail-step", State: "failed", CausesFailure: true},
				},
			},
		},
	}
	out := engine.FormatText(r)
	if !strings.Contains(out, "✓") {
		t.Fatalf("FormatText missing pass symbol:\n%s", out)
	}
	if !strings.Contains(out, "✗") {
		t.Fatalf("FormatText missing fail symbol:\n%s", out)
	}
	if !strings.Contains(out, "FAIL") {
		t.Fatalf("FormatText missing FAIL summary:\n%s", out)
	}
}

func TestFormatText_ParseError(t *testing.T) {
	r := &engine.RunResult{ParseError: "unexpected EOF"}
	out := engine.FormatText(r)
	if !strings.Contains(out, "Parse error") {
		t.Fatalf("FormatText missing 'Parse error':\n%s", out)
	}
	if !strings.Contains(out, "unexpected EOF") {
		t.Fatalf("FormatText missing parse error detail:\n%s", out)
	}
}

func TestFormatJSON_ValidJSON(t *testing.T) {
	r := &engine.RunResult{Passed: true, Workflows: []engine.WorkflowOutcome{}}
	b, err := engine.FormatJSON(r)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("FormatJSON not valid JSON: %v\n%s", err, b)
	}
}

func TestFormatJSON_ContainsPassedField(t *testing.T) {
	r := &engine.RunResult{Passed: true}
	b, _ := engine.FormatJSON(r)
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("FormatJSON not valid JSON: %v\n%s", err, b)
	}
	if _, ok := out["Passed"]; !ok {
		t.Fatalf("FormatJSON missing 'Passed' field; keys: %v", keysOf(out))
	}
}

func TestFormatJSON_ContainsNestedData(t *testing.T) {
	r := &engine.RunResult{
		Passed: false,
		Workflows: []engine.WorkflowOutcome{
			{WorkflowID: "wf1", Passed: false, Steps: []engine.StepOutcome{
				{StepID: "s1", State: "failed"},
			}},
		},
	}
	b, _ := engine.FormatJSON(r)

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("FormatJSON not valid JSON: %v\n%s", err, b)
	}

	wfs, ok := out["Workflows"].([]any)
	if !ok || len(wfs) == 0 {
		t.Fatalf("FormatJSON missing Workflows array; got: %s", b)
	}
	wf, ok := wfs[0].(map[string]any)
	if !ok {
		t.Fatalf("FormatJSON workflow entry is not an object")
	}
	if wf["WorkflowID"] != "wf1" {
		t.Fatalf("FormatJSON WorkflowID mismatch: %v", wf["WorkflowID"])
	}
	steps, ok := wf["Steps"].([]any)
	if !ok || len(steps) == 0 {
		t.Fatalf("FormatJSON missing Steps in workflow")
	}
}

// TestFormatText_DiagnosticsGatedOnVerbose verifies the verbose gate end-to-end:
// the "Execution Record" diagnostics section must appear only when Verbose=true.
func TestFormatText_DiagnosticsGatedOnVerbose(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	spec := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf1
    steps:
      - id: step1
        kind: request
        request:
          protocol: http
          target: %s
`, srv.URL)
	f := writeWorkflow(t, spec)

	// Non-verbose: diagnostics section must be absent.
	outQuiet := engine.FormatText(engine.New().Run(engine.Config{WorkflowFile: f, Quiet: true}))
	if strings.Contains(outQuiet, "Execution Record") {
		t.Errorf("FormatText without --verbose must not contain diagnostics section")
	}

	// Verbose: diagnostics section must be present.
	outVerbose := engine.FormatText(engine.New().Run(engine.Config{WorkflowFile: f, Verbose: true, Quiet: true}))
	if !strings.Contains(outVerbose, "Execution Record") {
		t.Errorf("FormatText with --verbose must contain diagnostics section:\n%s", outVerbose)
	}
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
