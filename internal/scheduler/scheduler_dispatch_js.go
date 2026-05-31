//go:build js

package scheduler

import (
	"context"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// runWorkflow routes to the sequential JS implementation on GOOS=js.
// The parallel errgroup path in scheduler_native.go is excluded on this build.
func (s *Scheduler) runWorkflow(
	ctx context.Context,
	workflow *compiler.CompiledWorkflow,
	execCtx *runtime.ExecutionContext,
) (*WorkflowResult, error) {
	return s.runWorkflowJS(ctx, workflow, execCtx)
}
