// validation_test.go tests canonical validation failures surfaced by the engine.

package engine_test

import (
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestRun_ValidationError(t *testing.T) {
	// Valid YAML structure but fails canonical validation (missing schemaVersion).
	yaml := `schemaVersion: 0
capabilities: []
workflows:
  - id: bad-flow
    steps: []
`
	path := writeWorkflow(t, yaml)
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.ValidationError == "" {
		t.Fatalf("expected ValidationError, got none; parseErr=%q planErr=%q", result.ParseError, result.PlanError)
	}
	if result.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", result.ExitCode())
	}
}
