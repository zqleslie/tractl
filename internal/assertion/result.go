package assertion

import (
	"time"

	"github.com/tractl/tractl/internal/spec"
)

// AssertionOutcome is the verdict of a single assertion evaluation.
type AssertionOutcome string //nolint:revive

// Assertion outcome constants.
const (
	OutcomePass    AssertionOutcome = "pass"
	OutcomeFail    AssertionOutcome = "fail"
	OutcomeSkipped AssertionOutcome = "skipped"
)

// AssertionResult is the output of evaluating one spec.Assertion.
type AssertionResult struct { //nolint:revive
	AssertionID   string
	AssertionULID string
	Kind          string
	Outcome       AssertionOutcome
	Severity      spec.AssertionSeverity
	// Message is human-readable: "expected 200, got 404"
	Message string
	// CausesFailure is true when Outcome==OutcomeFail AND Severity==SeverityError.
	// A warning failure does NOT cause step failure (§11.4).
	CausesFailure bool
	Duration      time.Duration
}

// StepFailed reports whether any result in the slice causes step failure.
func StepFailed(results []AssertionResult) bool {
	for _, r := range results {
		if r.CausesFailure {
			return true
		}
	}
	return false
}
