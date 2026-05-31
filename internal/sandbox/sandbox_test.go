package sandbox_test

import (
	"testing"
	"time"

	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/sandbox"
)

func TestSandbox_EmptySource_NoError(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute("", frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ms.Empty() {
		t.Fatalf("expected empty MutationSet")
	}
}

func TestSandbox_ReturnUndefined_EmptyMutationSet(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute("var x = 1;", frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ms.Empty() {
		t.Fatalf("expected empty MutationSet")
	}
}

func TestSandbox_VariableWrite(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { variables: { "k": "v" } };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ms.Variables["k"] != "v" {
		t.Fatalf("want v got %v", ms.Variables["k"])
	}
}

func TestSandbox_ExtractWrite(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { extracts: { "token": "abc" } };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ms.Extracts["token"] != "abc" {
		t.Fatalf("want abc got %v", ms.Extracts["token"])
	}
}

func TestSandbox_LogViaTracTlAPI(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`tractl.log("hello"); return {};`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	found := false
	for _, l := range ms.Logs {
		if l == "hello" {
			found = true
		}
	}
	if !found {
		t.Fatalf("log not recorded: %v", ms.Logs)
	}
}

func TestSandbox_CancellationSignal(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { cancel: true };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ms.Cancel {
		t.Fatalf("expected cancel true")
	}
}

func TestSandbox_ReadFrozenContext_WorkflowScope(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{Workflow: map[string]string{"env": "prod"}}
	ms, err := sb.Execute(`return { variables: { "r": ctx.workflow["env"] } };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ms.Variables["r"] != "prod" {
		t.Fatalf("want prod got %v", ms.Variables["r"])
	}
}

func TestSandbox_FrozenContextGoSideUnchanged(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{Spec: map[string]string{"x": "original"}}
	_, err := sb.Execute(`ctx.spec["x"] = "mutated"; return {};`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if frozen.Spec["x"] != "original" {
		t.Fatalf("frozen mutated: %v", frozen.Spec["x"])
	}
}

func TestSandbox_IOProhibited_FetchUndefined(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	_, err := sb.Execute(`if (typeof fetch !== "undefined") { throw new Error("fetch present"); } return {};`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestSandbox_IOProhibited_RequireUndefined(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	_, err := sb.Execute(`if (typeof require !== "undefined") { throw new Error("require present"); } return {};`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestSandbox_SyntaxError_ReturnsErrSyntax(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	_, err := sb.Execute(`{{invalid javascript`, frozen, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
	se, ok := err.(*sandbox.SandboxError)
	if !ok {
		t.Fatalf("expected SandboxError got %T", err)
	}
	if se.Code != sandbox.ErrSyntax {
		t.Fatalf("want ErrSyntax, got %s", se.Code)
	}
}

func TestSandbox_RuntimeError_ReturnsErrRuntime(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	_, err := sb.Execute(`throw new Error("oops");`, frozen, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
	se, ok := err.(*sandbox.SandboxError)
	if !ok {
		t.Fatalf("expected SandboxError got %T", err)
	}
	if se.Code != sandbox.ErrRuntime {
		t.Fatalf("want ErrRuntime, got %s", se.Code)
	}
}

func TestSandbox_Timeout_ReturnsErrTimeout(t *testing.T) {
	sb := sandbox.NewSandbox(100 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	_, err := sb.Execute(`while(true){} `, frozen, 1)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	se, ok := err.(*sandbox.SandboxError)
	if !ok {
		t.Fatalf("expected SandboxError got %T", err)
	}
	if se.Code != sandbox.ErrTimeout {
		t.Fatalf("want ErrTimeout, got %s", se.Code)
	}
}

func TestSandbox_DeterministicRandom_SameSeed_SameSequence(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	src := `return { variables: { "r": String(tractl.random()) } };`
	frozen := runtime.FrozenContext{}
	m1, err := sb.Execute(src, frozen, 42)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	m2, err := sb.Execute(src, frozen, 42)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m1.Variables["r"] != m2.Variables["r"] {
		t.Fatalf("expected same randoms: %v vs %v", m1.Variables["r"], m2.Variables["r"])
	}
}

func TestSandbox_DifferentSeeds_DifferentRandom(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	src := `return { variables: { "r": String(tractl.random()) } };`
	frozen := runtime.FrozenContext{}
	m1, err := sb.Execute(src, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	m2, err := sb.Execute(src, frozen, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m1.Variables["r"] == m2.Variables["r"] {
		t.Fatalf("expected different randoms: %v vs %v", m1.Variables["r"], m2.Variables["r"])
	}
}

func TestSandbox_AssertionAddition(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { assertions: [{ kind: "status", operator: "equals", expected: "200" }] };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ms.Assertions) != 1 {
		t.Fatalf("expected 1 assertion got %d", len(ms.Assertions))
	}
	if ms.Assertions[0].Kind != "status" || ms.Assertions[0].Expected != "200" {
		t.Fatalf("assertion content mismatch: %+v", ms.Assertions[0])
	}
}

func TestSandbox_AssertionSeverityDefaultsToError(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { assertions: [{ kind: "status", operator: "equals", expected: "200" }] };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ms.Assertions[0].Severity != "" && ms.Assertions[0].Severity != "error" {
		t.Fatalf("unexpected severity: %q", ms.Assertions[0].Severity)
	}
}

func TestSandbox_NumericVariableValueCoercedToString(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { variables: { "n": 42 } };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ms.Variables["n"] != "42" {
		t.Fatalf("want 42 got %v", ms.Variables["n"])
	}
}

func TestSandbox_UnknownReturnFieldsIgnored(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	ms, err := sb.Execute(`return { variables: {}, unknownField: "ignored" };`, frozen, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ms.Variables) != 0 {
		t.Fatalf("expected empty variables")
	}
}

func TestSandbox_ConcurrentExecutions_NoDataRace(t *testing.T) {
	sb := sandbox.NewSandbox(200 * time.Millisecond)
	frozen := runtime.FrozenContext{}
	const n = 5
	ch := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			_, err := sb.Execute("return {};", frozen, 1)
			ch <- err
		}()
	}
	for i := 0; i < n; i++ {
		err := <-ch
		if err != nil {
			t.Fatalf("unexpected err in goroutine: %v", err)
		}
	}
}
