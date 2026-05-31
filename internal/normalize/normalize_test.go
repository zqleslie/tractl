// normalize_test.go tests the Normalizer: implicit dependency inference from
// header/target/body expressions, explicit + implicit deduplication, self-reference
// errors, dead-reference errors, cross-workflow ref exclusion, and determinism.
package normalize

import (
	"reflect"
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func makeStep(id string, deps []string) spec.Step {
	return spec.Step{ID: id, Kind: "request", DependsOn: deps, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x"}}
}

func TestNormalize_NoImplicitFromVars(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("a", nil),
				{ID: "b", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"K": "${vars.someVar}"}}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Workflows[0].Steps[1].DependsOn != nil {
		t.Fatalf("expected no deps, got %v", out.Workflows[0].Steps[1].DependsOn)
	}
}

func TestNormalize_HeaderRefAddsDep(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("login", nil),
				{ID: "fetch", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"Authorization": "Bearer ${steps.login.extracts.token}"}}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := out.Workflows[0].Steps[1].DependsOn
	if !reflect.DeepEqual(got, []string{"login"}) {
		t.Fatalf("expected [login], got %v", got)
	}
}

func TestNormalize_DedupExplicitAndImplicit(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("login", nil),
				{ID: "fetch", Kind: "request", DependsOn: []string{"login"}, Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"Authorization": "Bearer ${steps.login.extracts.token}"}}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := out.Workflows[0].Steps[1].DependsOn
	if !reflect.DeepEqual(got, []string{"login"}) {
		t.Fatalf("expected single login, got %v", got)
	}
}

func TestNormalize_SelfReferenceErrors(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				{ID: "loop", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"X-Self": "${steps.loop.extracts.x}"}}},
			},
		}},
	}
	_, err := Normalize(s)
	ne, ok := err.(*NormalizerError)
	if !ok || ne.Code != ErrSelfReference {
		t.Fatalf("expected ErrSelfReference, got %v", err)
	}
}

func TestNormalize_DeadReferenceErrors(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("a", nil),
				{ID: "b", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"K": "${steps.ghost.extracts.x}"}}},
			},
		}},
	}
	_, err := Normalize(s)
	ne, ok := err.(*NormalizerError)
	if !ok || ne.Code != ErrDeadReference {
		t.Fatalf("expected ErrDeadReference, got %v", err)
	}
}

func TestNormalize_CrossWorkflowRefNotAdded(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{
			{ID: "wfA", Steps: []spec.Step{makeStep("alpha", nil)}},
			{ID: "wfB", Steps: []spec.Step{
				{ID: "beta", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"K": "${steps.alpha.extracts.x}"}}},
			}},
		},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Workflows[1].Steps[0].DependsOn != nil {
		t.Fatalf("expected no intra-workflow deps for cross-workflow ref, got %v", out.Workflows[1].Steps[0].DependsOn)
	}
}

func TestNormalize_DoesNotMutateSource(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("login", nil),
				{ID: "fetch", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"Authorization": "Bearer ${steps.login.extracts.token}"}}},
			},
		}},
	}
	before := s.Workflows[0].Steps[1].DependsOn
	_, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(before, s.Workflows[0].Steps[1].DependsOn) {
		t.Fatalf("source mutated: before=%v after=%v", before, s.Workflows[0].Steps[1].DependsOn)
	}
}

func TestNormalize_TargetExpressionAddsDep(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("create", nil),
				{ID: "fetch", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "${steps.create.extracts.url}"}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := out.Workflows[0].Steps[1].DependsOn
	if !reflect.DeepEqual(got, []string{"create"}) {
		t.Fatalf("expected [create], got %v", got)
	}
}

func TestNormalize_WhenExpressionAddsDep(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("first", nil),
				{ID: "second", Kind: "request", When: "${steps.first.response.status == 200}", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x"}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := out.Workflows[0].Steps[1].DependsOn
	if !reflect.DeepEqual(got, []string{"first"}) {
		t.Fatalf("expected [first], got %v", got)
	}
}

func TestNormalize_DeterministicOutput(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("a", nil),
				makeStep("b", nil),
				{ID: "c", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Headers: map[string]string{"X": "${steps.b.extracts.x} ${steps.a.extracts.y}"}}},
			},
		}},
	}
	out1, _ := Normalize(s)
	out2, _ := Normalize(s)
	if !reflect.DeepEqual(out1.Workflows[0].Steps[2].DependsOn, out2.Workflows[0].Steps[2].DependsOn) {
		t.Fatalf("non-deterministic output")
	}
	if !reflect.DeepEqual(out1.Workflows[0].Steps[2].DependsOn, []string{"a", "b"}) {
		t.Fatalf("expected sorted [a, b], got %v", out1.Workflows[0].Steps[2].DependsOn)
	}
}

func TestNormalize_BodyContentScanned(t *testing.T) {
	s := &spec.TraCtlSpec{
		Workflows: []spec.Workflow{{
			ID: "wf1",
			Steps: []spec.Step{
				makeStep("login", nil),
				{ID: "post", Kind: "request", Request: &spec.RequestDescriptor{Protocol: "http", Target: "http://x", Body: &spec.BodyDescriptor{Encoding: "json", Content: map[string]any{"token": "${steps.login.extracts.token}"}}}},
			},
		}},
	}
	out, err := Normalize(s)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := out.Workflows[0].Steps[1].DependsOn
	if !reflect.DeepEqual(got, []string{"login"}) {
		t.Fatalf("expected [login], got %v", got)
	}
}
