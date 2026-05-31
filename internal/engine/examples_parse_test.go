// examples_parse_test.go defines code for the engine package.

package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestExamples_HookFiles_ParseAndValidate(t *testing.T) {
	patterns := []string{
		"../../examples/yaml/06-*.yaml",
		"../../examples/yaml/07-*.yaml",
		"../../examples/yaml/08-*.yaml",
		"../../examples/yaml/09-*.yaml",
		"../../examples/yaml/10-*.yaml",
		"../../examples/yaml/11-*.yaml",
		"../../examples/yaml/12-*.yaml",
		"../../examples/toon/06-*.toon",
		"../../examples/toon/07-*.toon",
		"../../examples/toon/08-*.toon",
		"../../examples/toon/09-*.toon",
		"../../examples/toon/10-*.toon",
		"../../examples/toon/11-*.toon",
		"../../examples/toon/12-*.toon",
	}

	for _, pat := range patterns {
		files, err := filepath.Glob(pat)
		if err != nil {
			t.Fatalf("glob %s: %v", pat, err)
		}
		for _, f := range files {
			f := f
			t.Run(filepath.Base(f), func(t *testing.T) {
				abs, _ := filepath.Abs(f)
				if _, err := os.Stat(abs); err != nil {
					t.Fatalf("file not found: %s", abs)
				}
				result := engine.New().Run(engine.Config{WorkflowFile: abs, Quiet: true})
				// We only care that parsing and validation pass; network may fail.
				if result.ParseError != "" {
					t.Fatalf("parse error: %s", result.ParseError)
				}
				if result.ValidationError != "" {
					t.Fatalf("validation error: %s", result.ValidationError)
				}
				if result.PlanError != "" {
					t.Fatalf("plan error: %s", result.PlanError)
				}
			})
		}
	}
}
