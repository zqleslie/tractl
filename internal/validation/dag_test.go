package validation

import (
	"strings"
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func TestDependsOn(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "valid linear dependsOn chain",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{
					ID: "wf1",
					Steps: []spec.Step{
						requestStep("a"),
						{ID: "b", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"a"}},
						{ID: "c", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"a", "b"}},
					},
				}},
			},
			wantPass: true,
		},
		{
			name: "dependsOn unknown step",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					requestStep("a"),
					{ID: "b", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"ghost"}},
				}}},
			},
			wantCodes: []ErrorCode{UnknownDependencyRef},
		},
		{
			name: "dependsOn self-reference",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "a", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"a"}},
				}}},
			},
			wantCodes: []ErrorCode{UnknownDependencyRef},
		},
	})
}

func TestCycleDetection(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name: "two-step cycle A→B→A",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "A", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"B"}},
					{ID: "B", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"A"}},
				}}},
			},
			wantCodes: []ErrorCode{CyclicDependency},
		},
		{
			name: "three-step cycle A→B→C→A",
			spec: &spec.TraCtlSpec{
				SchemaVersion: 1,
				Capabilities:  []string{"protocol.http"},
				Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
					{ID: "A", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"C"}},
					{ID: "B", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"A"}},
					{ID: "C", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"B"}},
				}}},
			},
			wantCodes: []ErrorCode{CyclicDependency},
		},
	})
}

func TestCycleMessageFormat(t *testing.T) {
	v := NewSpecValidator()
	s := &spec.TraCtlSpec{
		SchemaVersion: 1,
		Capabilities:  []string{"protocol.http"},
		Workflows: []spec.Workflow{{ID: "wf1", Steps: []spec.Step{
			{ID: "A", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"C"}},
			{ID: "B", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"A"}},
			{ID: "C", Kind: "request", Request: &spec.RequestDescriptor{}, DependsOn: []string{"B"}},
		}}},
	}

	errs := v.Validate(s)
	if !hasError(errs, CyclicDependency) {
		t.Fatal("expected CyclicDependency error")
	}
	for _, e := range errs {
		if e.Code == CyclicDependency {
			if !strings.Contains(e.Message, "→") {
				t.Errorf("cycle message must contain '→' separator, got: %s", e.Message)
			}
			if !strings.HasPrefix(e.Message, "cycle detected:") {
				t.Errorf("cycle message must start with 'cycle detected:', got: %s", e.Message)
			}
		}
	}
}
