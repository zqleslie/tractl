package assertion_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/spec"
)

var workflowStart = time.Now()

func makeResult(status int, headers map[string]string, body []byte) *runtime.StepResult {
	return &runtime.StepResult{
		StepID:  "step1",
		Status:  status,
		Headers: headers,
		Body:    body,
	}
}

func makeAssertion(kind, op, target string, expected interface{}, severity spec.AssertionSeverity) spec.Assertion {
	return spec.Assertion{
		ID:       "a1",
		Kind:     kind,
		Op:       op,
		Target:   target,
		Expected: expected,
		Severity: severity,
	}
}

func evaluate(t *testing.T, a spec.Assertion, result *runtime.StepResult, ctx *runtime.ExecutionContext) (assertion.AssertionResult, assertion.EvalEvent) {
	t.Helper()
	ev := assertion.New()
	ar, event, _ := ev.Evaluate(context.Background(), a, result, ctx, "trace-1", "span-1", workflowStart)
	return ar, event
}

// ── Status assertions ─────────────────────────────────────────────────────────

func TestStatus_Equals_Pass(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
	if ar.CausesFailure {
		t.Error("pass should not cause failure")
	}
}

func TestStatus_Equals_Fail(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityError)
	r := makeResult(404, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail, got %s", ar.Outcome)
	}
	if !ar.CausesFailure {
		t.Error("error-severity fail should cause failure")
	}
}

func TestStatus_InRange_Pass(t *testing.T) {
	a := makeAssertion("status", "inRange", "", "200-299", spec.SeverityError)
	r := makeResult(201, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestStatus_InRange_Fail(t *testing.T) {
	a := makeAssertion("status", "inRange", "", "200-299", spec.SeverityError)
	r := makeResult(404, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail, got %s", ar.Outcome)
	}
}

func TestStatus_Exists_Pass(t *testing.T) {
	a := makeAssertion("status", "exists", "", nil, spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s", ar.Outcome)
	}
}

func TestStatus_Exists_Fail_ZeroStatus(t *testing.T) {
	a := makeAssertion("status", "exists", "", nil, spec.SeverityError)
	r := makeResult(0, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail for status=0, got %s", ar.Outcome)
	}
}

// ── Header assertions ─────────────────────────────────────────────────────────

func TestHeader_Equals_Pass(t *testing.T) {
	a := makeAssertion("header", "equals", "content-type", "application/json", spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "application/json"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestHeader_Equals_Fail(t *testing.T) {
	a := makeAssertion("header", "equals", "content-type", "application/json", spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "text/plain"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail, got %s", ar.Outcome)
	}
}

func TestHeader_Contains_Pass(t *testing.T) {
	a := makeAssertion("header", "contains", "content-type", "json", spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "application/json"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestHeader_Exists_Pass(t *testing.T) {
	a := makeAssertion("header", "exists", "x-request-id", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"x-request-id": "abc123"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s", ar.Outcome)
	}
}

func TestHeader_Exists_Fail_Absent(t *testing.T) {
	a := makeAssertion("header", "exists", "x-request-id", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail when header absent, got %s", ar.Outcome)
	}
}

func TestHeader_CredentialHeader_Masked(t *testing.T) {
	a := makeAssertion("header", "exists", "authorization", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"authorization": "Bearer token"}, nil)
	_, ev := evaluate(t, a, r, nil)
	if !strings.Contains(ev.TargetPlaceholder, "[masked]") {
		t.Errorf("expected [masked] in placeholder for authorization header, got %q", ev.TargetPlaceholder)
	}
}

func TestHeader_Matches_Pass(t *testing.T) {
	a := makeAssertion("header", "matches", "content-type", `^application/`, spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "application/json"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestHeader_Matches_Fail(t *testing.T) {
	a := makeAssertion("header", "matches", "content-type", `^application/`, spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "text/html"}, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail, got %s", ar.Outcome)
	}
}

func TestHeader_Matches_InvalidRegex(t *testing.T) {
	a := makeAssertion("header", "matches", "content-type", `[invalid`, spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "application/json"}, nil)
	ev := assertion.New()
	_, _, err := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if err == nil {
		t.Error("expected error for invalid regex")
	}
	ae, ok := err.(*assertion.AssertionError)
	if !ok || ae.Code != assertion.ErrRegexInvalid {
		t.Errorf("expected ErrRegexInvalid, got %v", err)
	}
}

// ── Body assertions ───────────────────────────────────────────────────────────

func TestBody_Equals_Pass(t *testing.T) {
	body := []byte(`{"status":"ok"}`)
	a := makeAssertion("body", "equals", "status", "ok", spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestBody_Equals_NestedPath_Pass(t *testing.T) {
	body := []byte(`{"items":[{"id":"abc"}]}`)
	a := makeAssertion("body", "equals", "items.0.id", "abc", spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestBody_Exists_Pass(t *testing.T) {
	body := []byte(`{"user":{"name":"alice"}}`)
	a := makeAssertion("body", "exists", "user.name", nil, spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestBody_Exists_Fail_AbsentPath(t *testing.T) {
	body := []byte(`{"user":{}}`)
	a := makeAssertion("body", "exists", "user.name", nil, spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail for absent path, got %s", ar.Outcome)
	}
}

func TestBody_EmptyBody_Fail(t *testing.T) {
	a := makeAssertion("body", "equals", "status", "ok", spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail for empty body, got %s", ar.Outcome)
	}
	if !strings.Contains(ar.Message, "response body is empty") {
		t.Errorf("expected 'response body is empty' in message, got %q", ar.Message)
	}
}

func TestBody_Contains_Pass(t *testing.T) {
	body := []byte(`{"message":"hello world"}`)
	a := makeAssertion("body", "contains", "message", "hello", spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

func TestBody_Jsonpath_Pass(t *testing.T) {
	body := []byte(`{"code":42}`)
	a := makeAssertion("body", "jsonpath", "code", "42", spec.SeverityError)
	r := makeResult(200, nil, body)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass, got %s: %s", ar.Outcome, ar.Message)
	}
}

// ── Severity ──────────────────────────────────────────────────────────────────

func TestSeverity_Error_Fail_CausesFailure(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityError)
	r := makeResult(500, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if !ar.CausesFailure {
		t.Error("error severity fail should cause failure")
	}
}

func TestSeverity_Warning_Fail_NoCausesFailure(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityWarning)
	r := makeResult(500, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.CausesFailure {
		t.Error("warning severity fail should NOT cause failure")
	}
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail outcome, got %s", ar.Outcome)
	}
}

func TestSeverity_DefaultIsError(t *testing.T) {
	a := spec.Assertion{ID: "a1", Kind: "status", Op: "equals", Expected: "200"}
	r := makeResult(500, nil, nil)
	ev := assertion.New()
	ar, _, _ := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if !ar.CausesFailure {
		t.Error("default severity should be error; fail should cause failure")
	}
}

func TestStepFailed_WithErrorFail(t *testing.T) {
	results := []assertion.AssertionResult{
		{Outcome: assertion.OutcomeFail, CausesFailure: true},
		{Outcome: assertion.OutcomeFail, CausesFailure: false},
	}
	if !assertion.StepFailed(results) {
		t.Error("expected StepFailed=true when any CausesFailure=true")
	}
}

func TestStepFailed_OnlyWarningFails(t *testing.T) {
	results := []assertion.AssertionResult{
		{Outcome: assertion.OutcomeFail, CausesFailure: false},
	}
	if assertion.StepFailed(results) {
		t.Error("expected StepFailed=false when only warning failures")
	}
}

func TestStepFailed_Empty(t *testing.T) {
	if assertion.StepFailed(nil) {
		t.Error("expected StepFailed=false for empty slice")
	}
}

// ── Stub kinds ────────────────────────────────────────────────────────────────

func TestStub_Schema(t *testing.T) {
	a := makeAssertion("schema", "", "", nil, spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeSkipped {
		t.Errorf("expected skipped for schema kind, got %s", ar.Outcome)
	}
}

func TestStub_Script(t *testing.T) {
	a := makeAssertion("script", "", "", nil, spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeSkipped {
		t.Errorf("expected skipped for script kind, got %s", ar.Outcome)
	}
}

func TestStub_Extension(t *testing.T) {
	a := makeAssertion("extension", "", "", nil, spec.SeverityError)
	r := makeResult(200, nil, nil)
	ar, _ := evaluate(t, a, r, nil)
	if ar.Outcome != assertion.OutcomeSkipped {
		t.Errorf("expected skipped for extension kind, got %s", ar.Outcome)
	}
}

// ── Expression resolution ─────────────────────────────────────────────────────

func TestExpression_Resolved_Expected(t *testing.T) {
	ctx, err := runtime.NewExecutionContext("wf1", map[string]string{"expectedStatus": "200"}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error creating ExecutionContext: %v", err)
	}
	a := spec.Assertion{
		ID:       "a1",
		Kind:     "status",
		Op:       "equals",
		Expected: "${vars.expectedStatus}",
		Severity: spec.SeverityError,
	}
	r := makeResult(200, nil, nil)
	ev := assertion.New()
	ar, _, _ := ev.Evaluate(context.Background(), a, r, ctx, "t", "s", workflowStart)
	if ar.Outcome != assertion.OutcomePass {
		t.Errorf("expected pass after expression resolution, got %s: %s", ar.Outcome, ar.Message)
	}
}

// ── EvalEvent ─────────────────────────────────────────────────────────────────

func TestEvalEvent_FieldsPassedThrough(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityError)
	r := makeResult(200, nil, nil)
	ev := assertion.New()
	_, event, _ := ev.Evaluate(context.Background(), a, r, nil, "trace-xyz", "span-abc", workflowStart)
	if event.TraceID != "trace-xyz" {
		t.Errorf("expected TraceID=trace-xyz, got %q", event.TraceID)
	}
	if event.SpanID != "span-abc" {
		t.Errorf("expected SpanID=span-abc, got %q", event.SpanID)
	}
	if event.MonotonicOffset < 0 {
		t.Error("MonotonicOffset must be non-negative")
	}
}

func TestEvalEvent_MaskedHeader(t *testing.T) {
	a := makeAssertion("header", "exists", "authorization", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"authorization": "Bearer secret"}, nil)
	ev := assertion.New()
	_, event, _ := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if !strings.Contains(event.TargetPlaceholder, "[masked]") {
		t.Errorf("expected [masked] in TargetPlaceholder, got %q", event.TargetPlaceholder)
	}
}

func TestEvalEvent_StatusPlaceholder(t *testing.T) {
	a := makeAssertion("status", "exists", "", nil, spec.SeverityError)
	r := makeResult(200, nil, nil)
	_, ev := evaluate(t, a, r, nil)
	if ev.TargetPlaceholder != "response.status" {
		t.Errorf("expected response.status placeholder, got %q", ev.TargetPlaceholder)
	}
}

// ── Additional credential header variants ────────────────────────────────────

func TestHeader_XApiKey_Masked(t *testing.T) {
	a := makeAssertion("header", "exists", "x-api-key", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"x-api-key": "mykey"}, nil)
	_, ev := evaluate(t, a, r, nil)
	if !strings.Contains(ev.TargetPlaceholder, "[masked]") {
		t.Errorf("expected [masked] for x-api-key, got %q", ev.TargetPlaceholder)
	}
}

func TestHeader_Cookie_Masked(t *testing.T) {
	a := makeAssertion("header", "exists", "cookie", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"cookie": "session=abc"}, nil)
	_, ev := evaluate(t, a, r, nil)
	if !strings.Contains(ev.TargetPlaceholder, "[masked]") {
		t.Errorf("expected [masked] for cookie, got %q", ev.TargetPlaceholder)
	}
}

func TestHeader_NonCredential_NotMasked(t *testing.T) {
	a := makeAssertion("header", "exists", "content-type", nil, spec.SeverityError)
	r := makeResult(200, map[string]string{"content-type": "application/json"}, nil)
	_, ev := evaluate(t, a, r, nil)
	if strings.Contains(ev.TargetPlaceholder, "[masked]") {
		t.Errorf("content-type should NOT be masked, got %q", ev.TargetPlaceholder)
	}
}

// ── Unknown kind / operator ───────────────────────────────────────────────────

func TestUnknownKind(t *testing.T) {
	a := makeAssertion("grpc", "equals", "", "0", spec.SeverityError)
	r := makeResult(0, nil, nil)
	ev := assertion.New()
	ar, _, err := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if err == nil {
		t.Error("expected error for unknown kind")
	}
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail outcome for unknown kind, got %s", ar.Outcome)
	}
}

func TestUnknownOperator_Status(t *testing.T) {
	a := makeAssertion("status", "notAnOp", "", "200", spec.SeverityError)
	r := makeResult(200, nil, nil)
	ev := assertion.New()
	ar, _, err := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if err == nil {
		t.Error("expected error for unknown operator")
	}
	if ar.Outcome != assertion.OutcomeFail {
		t.Errorf("expected fail for unknown operator, got %s", ar.Outcome)
	}
}

// ── Duration is populated ─────────────────────────────────────────────────────

func TestDuration_IsPopulated(t *testing.T) {
	a := makeAssertion("status", "equals", "", "200", spec.SeverityError)
	r := makeResult(200, nil, nil)
	ev := assertion.New()
	ar, _, _ := ev.Evaluate(context.Background(), a, r, nil, "t", "s", workflowStart)
	if ar.Duration < 0 {
		t.Error("Duration must be non-negative")
	}
}
