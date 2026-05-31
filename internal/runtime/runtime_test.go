// runtime_test.go defines code for the runtime package.

package runtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

func mustNewEC(t *testing.T, workflowID string, specVars, workflowVars, envVars map[string]string) *ExecutionContext {
	t.Helper()
	ec, err := NewExecutionContext(workflowID, specVars, workflowVars, envVars)
	if err != nil {
		t.Fatalf("unexpected error creating ExecutionContext: %v", err)
	}
	return ec
}

func TestIsTerminalStates(t *testing.T) {
	cases := map[StepState]bool{
		StatePending:           false,
		StateRunning:           false,
		StateSucceeded:         true,
		StateFailed:            true,
		StateConditionalSkip:   true,
		StateDependencySkipped: true,
		StateCancelled:         true,
	}
	for s, want := range cases {
		if IsTerminal(s) != want {
			t.Fatalf("IsTerminal(%s): want %v", s, want)
		}
	}
}

func TestExecutionContext_InitialStatePending(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	if ec.GetStepState("noexist") != StatePending {
		t.Fatalf("expected initial pending state")
	}
}

func TestExecutionContext_SetAndGetState(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("s1", StateRunning)
	if ec.GetStepState("s1") != StateRunning {
		t.Fatalf("state mismatch")
	}
}

func TestExecutionContext_SetAndGetResult(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	r := &StepResult{StepID: "s1", WorkflowID: "wf1", State: StateSucceeded, StartedAt: time.Now(), FinishedAt: time.Now()}
	ec.SetStepResult(r)
	r2, ok := ec.GetStepResult("s1")
	if !ok || r2 == nil {
		t.Fatalf("expected result present")
	}
	if r2.State != StateSucceeded {
		t.Fatalf("state mismatch")
	}
}

func TestExecutionContext_GetResultMissing(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	if r, ok := ec.GetStepResult("missing"); ok || r != nil {
		t.Fatalf("expected missing result")
	}
}

func TestExecutionContext_SetExtractAndRetrieve(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepExtract("s1", "token", "abc")
	v, ok := ec.GetExtract("s1", "token")
	if !ok || v != "abc" {
		t.Fatalf("extract mismatch")
	}
}

func TestExecutionContext_IsEligible_AllDepsTerminal(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("a", StateSucceeded)
	ec.SetStepState("b", StateConditionalSkip)
	if !ec.IsEligible("s", []string{"a", "b"}) {
		t.Fatalf("expected eligible")
	}
}

func TestExecutionContext_IsEligible_DepStillRunning(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("a", StateRunning)
	if ec.IsEligible("s", []string{"a"}) {
		t.Fatalf("expected not eligible")
	}
}

func TestExecutionContext_ShouldExecute_AllSucceeded(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("a", StateSucceeded)
	ec.SetStepState("b", StateConditionalSkip)
	if !ec.IsEligible("s", []string{"a", "b"}) {
		t.Fatalf("precondition failed")
	}
	if !ec.ShouldExecute("s", []string{"a", "b"}) {
		t.Fatalf("expected ShouldExecute true")
	}
}

func TestExecutionContext_ShouldExecute_DepFailed(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("a", StateFailed)
	if !ec.IsEligible("s", []string{"a"}) {
		t.Fatalf("precondition failed")
	}
	if ec.ShouldExecute("s", []string{"a"}) {
		t.Fatalf("expected ShouldExecute false")
	}
}

func TestExecutionContext_ShouldExecute_DepConditionalSkip(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepState("a", StateConditionalSkip)
	if !ec.IsEligible("s", []string{"a"}) {
		t.Fatalf("precondition failed")
	}
	if !ec.ShouldExecute("s", []string{"a"}) {
		t.Fatalf("expected ShouldExecute true")
	}
}

func TestExecutionContext_ConcurrentStateUpdates(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sid := "s" + string(rune('A'+(i%26)))
			ec.SetStepState(sid, StateRunning)
		}(i)
	}
	wg.Wait()
}

// VariableResolver tests
func TestResolve_SpecVar(t *testing.T) {
	specVars := map[string]string{"x": "1"}
	ec := mustNewEC(t, "wf1", specVars, nil, nil)
	vr := NewVariableResolver(ec)
	v, ok := vr.Resolve("spec", "x")
	if !ok || v != "1" {
		t.Fatalf("expected spec var 1")
	}
}

func TestResolve_WorkflowOverridesSpec(t *testing.T) {
	specVars := map[string]string{"x": "1"}
	wfVars := map[string]string{"x": "2"}
	ec := mustNewEC(t, "wf1", specVars, wfVars, nil)
	vr := NewVariableResolver(ec)
	v, ok := vr.Resolve("workflow", "x")
	if !ok || v != "2" {
		t.Fatalf("expected workflow override")
	}
}

func TestResolve_EnvOverridesWorkflow(t *testing.T) {
	specVars := map[string]string{"x": "1"}
	wfVars := map[string]string{"x": "2"}
	env := map[string]string{"x": "3"}
	ec := mustNewEC(t, "wf1", specVars, wfVars, env)
	vr := NewVariableResolver(ec)
	v, ok := vr.Resolve("env", "x")
	if !ok || v != "3" {
		t.Fatalf("expected env override")
	}
}

func TestResolve_NotFound(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	vr := NewVariableResolver(ec)
	_, ok := vr.Resolve("spec", "missing")
	if ok {
		t.Fatalf("expected not found")
	}
}

// ExpressionEvaluator tests
func TestEvaluate_NoExpression(t *testing.T) {
	ec := mustNewEC(t, "wf1", map[string]string{"token": "tkn"}, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "plain string")
	if err != nil || out != "plain string" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_VarsLookup(t *testing.T) {
	ec := mustNewEC(t, "wf1", map[string]string{"token": "tkn"}, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "Bearer ${vars.token}")
	if err != nil || out != "Bearer tkn" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_ExtractLookup(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ec.SetStepExtract("login", "token", "abc")
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "tok=${steps.login.extracts.token}")
	if err != nil || out != "tok=abc" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_ResponseStatus(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	r := &StepResult{StepID: "login", State: StateSucceeded, Timeline: &RequestTimeline{StatusCode: 201}}
	ec.SetStepResult(r)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "status=${steps.login.response.status}")
	if err != nil || out != "status=201" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_DefaultFallback(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "${vars.missing ?? \"default\"}")
	if err != nil || out != "default" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_MultipleExpressions(t *testing.T) {
	ec := mustNewEC(t, "wf1", map[string]string{"a": "1", "b": "2"}, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	out, err := ev.Evaluate(context.Background(), "x=${vars.a},y=${vars.b}")
	if err != nil || out != "x=1,y=2" {
		t.Fatalf("unexpected evaluate result: %v %v", out, err)
	}
}

func TestEvaluate_UnknownPattern(t *testing.T) {
	ec := mustNewEC(t, "wf1", nil, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	_, err := ev.Evaluate(context.Background(), "${unknown.foo}")
	if err == nil {
		t.Fatalf("expected error for unknown pattern")
	}
}

func TestEvaluate_NestedMap(t *testing.T) {
	ec := mustNewEC(t, "wf1", map[string]string{"token": "tkn"}, nil, nil)
	ev := NewExpressionEvaluator(NewVariableResolver(ec), ec)
	m := map[string]string{"Authorization": "Bearer ${vars.token}", "X-Empty": "${vars.missing ?? \"x\"}"}
	out, err := ev.EvaluateMap(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["Authorization"] != "Bearer tkn" || out["X-Empty"] != "x" {
		t.Fatalf("unexpected map eval: %#v", out)
	}
}
