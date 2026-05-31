// testhelpers_test.go provides shared test fixture builders, assertion helpers,
// and the table-driven test runner used by all validation rule test files.
package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

// ── minimal builders ──────────────────────────────────────────────────────────

func requestStep(id string) spec.Step {
	return spec.Step{
		ID:      id,
		Kind:    "request",
		Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://example.com"},
	}
}

func scriptStep(id string) spec.Step {
	return spec.Step{ID: id, Kind: "script", Script: &spec.ScriptDescriptor{}}
}

func extensionCallStep(id string) spec.Step {
	return spec.Step{ID: id, Kind: "extensionCall", ExtensionCall: &spec.ExtensionCallDescriptor{}}
}

func compositeStep(id string) spec.Step {
	return spec.Step{ID: id, Kind: "composite", Composite: &spec.CompositeDescriptor{}}
}

func minimalValidSpec() *spec.TraCtlSpec {
	return &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows:     []spec.Workflow{{ID: "wf1", Steps: []spec.Step{requestStep("step1")}}},
	}
}

// ── assertion helpers ─────────────────────────────────────────────────────────

func hasError(errs []ValidationError, code ErrorCode) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

func noErrors(t *testing.T, errs []ValidationError) {
	t.Helper()
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %d:", len(errs))
		for _, e := range errs {
			t.Errorf("  [%s] %s: %s", e.Code, e.Field, e.Message)
		}
	}
}

// runValidatorTests is the shared table-driven runner used by each focused test file.
func runValidatorTests(t *testing.T, tests []validatorTestCase) {
	t.Helper()
	v := NewSpecValidator()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(tt.spec)
			if tt.wantPass {
				noErrors(t, errs)
				return
			}
			for _, code := range tt.wantCodes {
				if !hasError(errs, code) {
					t.Errorf("expected error with code %q but it was not in results:", code)
					for _, e := range errs {
						t.Errorf("  [%s] %s: %s", e.Code, e.Field, e.Message)
					}
				}
			}
		})
	}
}

type validatorTestCase struct {
	name      string
	spec      *spec.TraCtlSpec
	wantCodes []ErrorCode
	wantPass  bool
}
