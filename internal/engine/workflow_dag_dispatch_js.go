//go:build js

package engine

import (
	goctx "context"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/diagnostics"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
)

// runWorkflowDAG routes to the sequential JS implementation on GOOS=js.
// The parallel errgroup path in workflow_dag_native.go is excluded on this build.
func runWorkflowDAG(
	ctx goctx.Context,
	plan *compiler.CompiledPlan,
	sched *scheduler.Scheduler,
	s *spec.TraCtlSpec,
	traceID string,
	cfg Config,
	collector *diagnostics.Collector,
	sb ScriptRunner,
	scriptSem chan struct{},
	seed int64,
) []WorkflowOutcome {
	return runWorkflowDAGJS(ctx, plan, sched, s, traceID, cfg, collector, sb, scriptSem, seed)
}
