// parse_test.go defines code for the engine package.

package engine_test

import (
	"testing"

	"github.com/tractl/tractl/internal/engine"
)

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"yaml", true},
		{"yml", true},
		{"json", true},
		{"toon", true},
		{"YAML", true},
		{"", true},
		{"xml", false},
		{"toml", false},
	}
	for _, tc := range tests {
		if got := engine.IsSupportedFormat(tc.format); got != tc.want {
			t.Errorf("IsSupportedFormat(%q) = %v, want %v", tc.format, got, tc.want)
		}
	}
}

func TestSupportedFormats(t *testing.T) {
	for _, f := range engine.SupportedFormats {
		if !engine.IsSupportedFormat(f) {
			t.Errorf("SupportedFormats entry %q rejected by IsSupportedFormat", f)
		}
	}
}

func TestRun_ParseError(t *testing.T) {
	path := writeWorkflow(t, "this: is: not: valid: yaml: [{")
	result := engine.New().Run(engine.Config{WorkflowFile: path})

	if result.ParseError == "" {
		t.Fatal("expected ParseError to be set")
	}
	if result.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", result.ExitCode())
	}
}
