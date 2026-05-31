package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

// TestAggregateNoFailFast verifies that all rules run independently and every
// error is reported in a single pass. Reference: tractl_spec.md §18.3
func TestAggregateNoFailFast(t *testing.T) {
	v := NewSpecValidator()
	s := &spec.TraCtlSpec{
		SchemaVersion: 0,                     // InvalidSchemaVersion
		Capabilities:  []string{"noDotHere"}, // InvalidCapabilityContract
		Workflows: []spec.Workflow{{
			ID: "1bad", // InvalidIdentifier
			Steps: []spec.Step{
				{
					ID:   "",        // MissingRequiredField
					Kind: "badkind", // InvalidStepKind
				},
				{
					ID:        "ok",
					Kind:      "request",
					Request:   nil,               // StepBodyKindMismatch
					DependsOn: []string{"ghost"}, // UnknownDependencyRef
				},
			},
		}},
	}

	errs := v.Validate(s)
	required := []ErrorCode{
		InvalidSchemaVersion,
		InvalidCapabilityContract,
		InvalidIdentifier,
		MissingRequiredField,
		InvalidStepKind,
		StepBodyKindMismatch,
		UnknownDependencyRef,
	}
	for _, code := range required {
		if !hasError(errs, code) {
			t.Errorf("expected error code %q but it was not reported", code)
		}
	}
}

// TestMultipleErrorsInOneSpec confirms that multiple independent error categories
// are all returned from a single Validate call.
func TestMultipleErrorsInOneSpec(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "schemaVersion + workflow id + duplicate steps + bad dependency",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 0,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{
					ID: "1bad",
					Steps: []spec.Step{
						requestStep("dup"),
						{ID: "dup", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"missing"}},
					},
				}},
			},
			wantCodes: []ErrorCode{InvalidSchemaVersion, InvalidIdentifier, DuplicateStepID, UnknownDependencyRef},
		},
		{
			name: "duplicate workflow + invalid capability",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"badcap"},
				Workflows: []spec.Workflow{
					{ID: "wf1", Steps: []spec.Step{requestStep("s1")}},
					{ID: "wf1", Steps: []spec.Step{requestStep("s2")}},
				},
			},
			wantCodes: []ErrorCode{InvalidCapabilityContract, DuplicateWorkflowID},
		},
	})
}

// TestFieldPaths verifies that error Field values use correct dot-bracket notation.
func TestFieldPaths(t *testing.T) {
	v := NewSpecValidator()
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http", "@bad"},
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{{
				ID:        "s1",
				Kind:      "request",
				Request:   &spec.RequestDescriptor{},
				DependsOn: []string{"ghost"},
			}},
		}},
	}

	errs := v.Validate(s)
	paths := make(map[string]bool, len(errs))
	for _, e := range errs {
		paths[e.Field] = true
	}

	if !paths["capabilities[1]"] {
		t.Errorf("expected error on field 'capabilities[1]', got fields: %v", paths)
	}
	if !paths["workflows[0].steps[0].dependsOn[0]"] {
		t.Errorf("expected error on field 'workflows[0].steps[0].dependsOn[0]', got fields: %v", paths)
	}
}
