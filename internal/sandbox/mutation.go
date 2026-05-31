package sandbox

import (
	"errors"
	"fmt"

	"github.com/tractl/tractl/internal/runtime"
)

// MutationSet is the declarative set of changes a script may request.
// All changes are applied transactionally by the engine after Execute returns.
// Scripts MUST NOT mutate control flow or canonical workflow state directly —
// only through MutationSet entries (spec §9.3, ADR-006 §4).
type MutationSet struct {
	Variables  map[string]string // variable name → value; applied to the caller's scope
	Extracts   map[string]string // extract binding name → value
	Assertions []AssertionAddition
	Logs       []string
	Cancel     bool // true requests workflow cancellation
}

// AssertionAddition is a runtime assertion injected by a script.
// Kind, Operator, Expected mirror spec.Assertion field semantics.
type AssertionAddition struct {
	ID       string // optional; empty = runtime-generated unique ID
	Kind     string // status | header | body | schema | script | extension
	Operator string // equals | matches | contains | exists | inRange | jsonpath | custom
	Target   string // header name for kind=header; gjson path for kind=body
	Expected string // expected value as string
	Severity string // "error" | "warning"; empty defaults to "error"
}

// Empty returns true if the MutationSet requests no changes.
func (m MutationSet) Empty() bool {
	return len(m.Variables) == 0 &&
		len(m.Extracts) == 0 &&
		len(m.Assertions) == 0 &&
		len(m.Logs) == 0 &&
		!m.Cancel
}

// ErrCancellationRequested is returned by Apply when MutationSet.Cancel is true.
// The engine treats this as a workflow-level cancellation signal.
var ErrCancellationRequested = errors.New("script requested workflow cancellation")

// execWriter is the minimal write-side interface expected from an execution context.
// It purposely does not reference internal/runtime so sandbox can stay decoupled.
type execWriter interface {
	SetVar(scope string, key, value string) error
	SetStepExtract(stepID, extractAs, value string)
}

// Apply writes all mutation set entries into execCtx at the given scope.
//
// Behaviour per mutation type:
//
//	Variables: calls execCtx.SetVar(scope, name, value) for each entry.
//	Extracts:  calls execCtx.SetStepExtract("", name, value) for each entry.
//	Logs:      no-op in this phase (log output deferred to future observability wiring).
//	Cancel:    if true, returns ErrCancellationRequested.
//	Assertions: no-op in this phase (runtime assertion injection deferred).
//
// Apply is transactional in the sense that it returns the first error encountered
// and does NOT guarantee rollback of partial writes. The engine is responsible
// for deciding how to handle partial application (typically by failing the hook).
//
// Returns nil if the MutationSet is empty.
func (m MutationSet) Apply(execCtx execWriter, scope runtime.Scope) error {
	if m.Empty() {
		return nil
	}
	if m.Cancel {
		return ErrCancellationRequested
	}
	// Variables
	for k, v := range m.Variables {
		if err := execCtx.SetVar(string(scope), k, v); err != nil {
			return fmt.Errorf("apply variable %s: %w", k, err)
		}
	}
	// Extracts: write into step-level extracts by default; use workflow scope mapping
	for k, v := range m.Extracts {
		// For now, SetStepExtract uses empty stepID which callers may interpret
		execCtx.SetStepExtract("", k, v)
	}
	// Logs and Assertions are no-ops in this phase.
	return nil
}
