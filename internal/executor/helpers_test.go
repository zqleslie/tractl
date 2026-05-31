package executor

import (
	"testing"

	"github.com/tractl/tractl/internal/runtime"
)

func mustNewEC(t *testing.T, workflowID string, specVars, workflowVars, envVars map[string]string) *runtime.ExecutionContext {
	t.Helper()
	ec, err := runtime.NewExecutionContext(workflowID, specVars, workflowVars, envVars)
	if err != nil {
		t.Fatalf("unexpected error creating ExecutionContext: %v", err)
	}
	return ec
}
