// interfaces.go declares consumer-side interfaces for the engine's external dependencies.

package engine

import (
	"time"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/sandbox"
	"github.com/tractl/tractl/internal/spec"
)

// Planner builds an executable plan from a canonical traCtlSpec.
type Planner interface {
	Plan(s *spec.TraCtlSpec) (*planner.ExecutionPlan, error)
}

// Compiler converts a planner execution plan into a compiled runtime plan.
type Compiler interface {
	Compile(plan *planner.ExecutionPlan, s *spec.TraCtlSpec) (*compiler.CompiledPlan, error)
}

// Evaluator is declared here (consumer package) per standard 5.1.
// Implemented implicitly by assertion.AssertionEvaluator in internal/assertion.
//
// Stream D coordination (HIGH #4): after this PR merges, Stream D should remove
// the duplicate Evaluator interface from internal/assertion. The compile-time check
// below confirms the two are in sync.
type Evaluator interface {
	EvaluateCompiled(
		a compiler.CompiledAssertion,
		result *runtime.StepResult,
		ctx *runtime.ExecutionContext,
		traceID, spanID string,
		workflowStart time.Time,
	) (assertion.AssertionResult, assertion.EvalEvent, error)
}

// ScriptRunner executes sandboxed workflow scripts.
type ScriptRunner interface {
	Execute(source string, frozen runtime.FrozenContext, seed int64) (sandbox.MutationSet, error)
}

var (
	_ Planner      = (*planner.Planner)(nil)
	_ Compiler     = (*compiler.Compiler)(nil)
	_ Evaluator    = (*assertion.AssertionEvaluator)(nil)
	_ ScriptRunner = (*sandbox.Sandbox)(nil)
)
