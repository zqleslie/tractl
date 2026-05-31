// result.go defines code for the engine package.

package engine

import (
	"time"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/diagnostics"
	rt "github.com/tractl/tractl/internal/runtime"
)

// StepOutcome carries the complete outcome of one step execution.
type StepOutcome struct {
	StepID           string
	StepULID         string
	State            string // runtime.StepState as string
	AssertionResults []assertion.AssertionResult
	// CausesFailure is true when any assertion caused a step failure.
	CausesFailure bool
	Error         string // non-empty when execution error occurred

	// HTTP response detail — populated when Config.Verbose is set.
	// Always included in JSON output; omitted from text output unless --verbose.
	ResponseStatus  int               `json:",omitempty"`
	ResponseHeaders map[string]string `json:",omitempty"`
	ResponseBody    string            `json:",omitempty"` // UTF-8 decoded; non-UTF-8 bytes replaced

	// Internal timing fields — not serialised; used to build Diagnostics.
	StartedAt       time.Time           `json:"-"`
	Duration        time.Duration       `json:"-"`
	DependsOn       []string            `json:"-"`
	WaitDuration    time.Duration       `json:"-"`
	RequestURL      string              `json:"-"`
	RequestMethod   string              `json:"-"`
	RequestTimeline *rt.RequestTimeline `json:"-"`
}

// WorkflowOutcome carries the complete outcome of one workflow execution.
type WorkflowOutcome struct {
	WorkflowID string
	Passed     bool
	// Skipped is true when the workflow was not executed because a prerequisite workflow failed.
	Skipped bool
	Steps   []StepOutcome

	// Internal timing fields — not serialised; used to build Diagnostics.
	StartedAt time.Time     `json:"-"`
	Duration  time.Duration `json:"-"`
}

// RunResult is the top-level result of a tractl run invocation.
type RunResult struct {
	// Passed is true when all workflows passed (no step caused failure).
	Passed    bool
	Workflows []WorkflowOutcome

	// Set when the pipeline fails before execution; mutually exclusive with Workflows content.
	ParseError      string
	ValidationError string
	PlanError       string

	// Diagnostics is the full observability record for this run. Non-nil after
	// a successful execution; included in JSON output when present.
	Diagnostics *diagnostics.ExecutionRecord `json:"diagnostics,omitempty"`

	// verbose is copied from Config.Verbose at the end of Run(); it gates the
	// diagnostics section in FormatText without changing the public signature.
	verbose bool `json:"-"`
}

// ExitCode returns the process exit code for the result:
//
//	0 — all assertions passed
//	1 — assertion failures (execution completed but some steps failed)
//	2 — pipeline error (parse/validation/plan failed before execution)
func (r *RunResult) ExitCode() int {
	if r.ParseError != "" || r.ValidationError != "" || r.PlanError != "" {
		return 2
	}
	if !r.Passed {
		return 1
	}
	return 0
}
