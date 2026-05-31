// step_test.go tests the step-level validation rules (Rules 7–12): step
// identifier format, step kind legality, step body / kind consistency,
// assertion and extract field validation.
package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func TestStepID(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "step id empty",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "", Kind: "request", Request: &spec.RequestDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{MissingRequiredField},
		},
		{
			name: "step id invalid (starts with digit)",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "9step", Kind: "request", Request: &spec.RequestDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{InvalidIdentifier},
		},
		{
			name: "step id reserved prefix traCtl.",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "traCtl.step", Kind: "request", Request: &spec.RequestDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{ReservedIdentifier},
		},
		{
			name: "duplicate step id within same workflow",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					requestStep("login"),
					requestStep("login"),
				}}},
			},
			wantCodes: []ErrorCode{DuplicateStepID},
		},
		{
			name: "duplicate step id across different workflows is allowed",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{
					{ID: "wf1", Steps: []spec.Step{requestStep("login")}},
					{ID: "wf2", Steps: []spec.Step{requestStep("login")}},
				},
			},
			wantPass: true,
		},
		{
			name: "empty steps list",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "wf1", Steps: []spec.Step{}}},
			},
			wantCodes: []ErrorCode{MissingRequiredField},
		},
	})
}

func TestStepKind(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "invalid step kind",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "webhook", Request: &spec.RequestDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{InvalidStepKind},
		},
		{
			name: "missing step kind",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "", Request: &spec.RequestDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{MissingRequiredField},
		},
		{
			name: "all four valid step kinds accepted",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					requestStep("r"),
					scriptStep("s"),
					extensionCallStep("e"),
					compositeStep("c"),
				}}},
			},
			wantPass: true,
		},
	})
}

func TestStepBody(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "kind=request but Request nil",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "request", Request: nil},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
		{
			name: "kind=request but wrong body (Script set)",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "request", Script: &spec.ScriptDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
		{
			name: "kind=request with multiple bodies set",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "request", Request: &spec.RequestDescriptor{}, Script: &spec.ScriptDescriptor{}},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
		{
			name: "kind=script but Script nil",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "script", Script: nil},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
		{
			name: "kind=extensionCall but ExtensionCall nil",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "extensionCall", ExtensionCall: nil},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
		{
			name: "kind=composite but Composite nil",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "s1", Kind: "composite", Composite: nil},
				}}},
			},
			wantCodes: []ErrorCode{StepBodyKindMismatch},
		},
	})
}
