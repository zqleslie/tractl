// workflow_test.go tests the workflow-level validation rules (Rules 3–6):
// workflows presence, workflow identifier format, duplicate IDs, and
// failure policy values.
package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func TestWorkflows(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "valid multiple workflows",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{
					{ID: "wf-alpha", Steps: []spec.Step{requestStep("s1")}},
					{ID: "wf-beta", Steps: []spec.Step{requestStep("s1")}},
				},
			},
			wantPass: true,
		},
		{
			name:      "empty workflows list",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.Workflows = []spec.Workflow{}; return s }(),
			wantCodes: []ErrorCode{MissingRequiredField},
		},
		{
			name: "duplicate workflow ids",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{
					{ID: "wf1", Steps: []spec.Step{requestStep("s1")}},
					{ID: "wf1", Steps: []spec.Step{requestStep("s2")}},
				},
			},
			wantCodes: []ErrorCode{DuplicateWorkflowID},
		},
		{
			name: "distinct workflow ids pass",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{
					{ID: "wf1", Steps: []spec.Step{requestStep("s1")}},
					{ID: "wf2", Steps: []spec.Step{requestStep("s1")}},
				},
			},
			wantPass: true,
		},
	})
}

func TestWorkflowID(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "workflow id empty",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{MissingRequiredField},
		},
		{
			name: "workflow id starts with digit",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "1wf", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{InvalidIdentifier},
		},
		{
			name: "workflow id contains space",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "my workflow", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{InvalidIdentifier},
		},
		{
			name: "workflow id reserved prefix traCtl.",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "traCtl.internal", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{ReservedIdentifier},
		},
		{
			name: "workflow id reserved prefix _traCtl.",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "_traCtl.sys", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{ReservedIdentifier},
		},
		{
			name: "valid dotted workflow id",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "auth.login.flow", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantPass: true,
		},
	})
}

func TestFailurePolicy(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "empty failurePolicy is valid (defaults to resilient at normalisation)",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "wf1", FailurePolicy: "", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantPass: true,
		},
		{
			name: "resilient failurePolicy",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "wf1", FailurePolicy: spec.FailurePolicyResilient, Steps: []spec.Step{requestStep("s1")}}},
			},
			wantPass: true,
		},
		{
			name: "failFast failurePolicy",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "wf1", FailurePolicy: spec.FailurePolicyFailFast, Steps: []spec.Step{requestStep("s1")}}},
			},
			wantPass: true,
		},
		{
			name: "invalid failurePolicy value rejected",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows:     []spec.Workflow{{ID: "wf1", FailurePolicy: "stopAll", Steps: []spec.Step{requestStep("s1")}}},
			},
			wantCodes: []ErrorCode{MissingRequiredField},
		},
	})
}
