package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

func TestBuild_ResolvesExpressionInURL(t *testing.T) {
	ec := mustNewEC(t, "wf", map[string]string{"baseUrl": "http://example.local"}, nil, nil)
	eval := runtime.NewExpressionEvaluator(runtime.NewVariableResolver(ec), ec)
	b := NewRequestBuilder(eval)
	step := &compiler.CompiledStep{Request: &compiler.CompiledRequest{Target: "${vars.baseUrl}/ping", Method: "GET"}}
	req, err := b.Build(context.Background(), step)
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "http://example.local/ping" {
		t.Fatalf("unexpected url: %s", req.URL.String())
	}
}

func TestBuild_ResolvesExpressionInHeader(t *testing.T) {
	ec := mustNewEC(t, "wf", nil, nil, nil)
	ec.SetStepExtract("login", "token", "abc123")
	eval := runtime.NewExpressionEvaluator(runtime.NewVariableResolver(ec), ec)
	b := NewRequestBuilder(eval)
	step := &compiler.CompiledStep{Request: &compiler.CompiledRequest{Target: "http://x", Method: "GET", Headers: map[string]string{"Authorization": "Bearer ${steps.login.extracts.token}"}}}
	req, err := b.Build(context.Background(), step)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer abc123" {
		t.Fatalf("unexpected header: %s", req.Header.Get("Authorization"))
	}
}

func TestBuild_JSONBody(t *testing.T) {
	ec := mustNewEC(t, "wf", nil, nil, nil)
	eval := runtime.NewExpressionEvaluator(runtime.NewVariableResolver(ec), ec)
	b := NewRequestBuilder(eval)
	step := &compiler.CompiledStep{Request: &compiler.CompiledRequest{Target: "http://x", Method: "POST", Body: &compiler.CompiledBody{Encoding: "json", Content: map[string]interface{}{"key": "value"}}}}
	req, err := b.Build(context.Background(), step)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content-type: %s", req.Header.Get("Content-Type"))
	}
}

func TestMaskHeaders_Authorization(t *testing.T) {
	m := NewCredentialMasker()
	h := http.Header{}
	h.Set("Authorization", "Bearer secret")
	out := m.MaskHeaders(h)
	if out.Get("Authorization") == "Bearer secret" {
		t.Fatalf("expected masked header")
	}
}

func TestMaskURL_TokenParam(t *testing.T) {
	m := NewCredentialMasker()
	u := "http://x?a=1&token=secret"
	out := m.MaskURL(u)
	if out == u {
		t.Fatalf("expected masked url")
	}
}

func TestExecute_200Succeeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: srv.URL, Method: "GET"}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", res.State)
	}
}

func TestExecute_AssertionStatusEquals_Fail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("not"))
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: srv.URL, Method: "GET"}, Assertions: []compiler.CompiledAssertion{{ID: "a1", Kind: "status", Op: "equals", Expected: "200", Severity: "error"}}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateFailed {
		t.Fatalf("expected failed, got %s", res.State)
	}
}

func TestExecute_ExtractStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: srv.URL, Method: "GET"}, Extracts: []compiler.CompiledExtract{{ID: "e1", Source: "status", As: "statusCode"}}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	v, ok := exCtx.GetExtract("s1", "statusCode")
	if !ok || v != "200" {
		t.Fatalf("unexpected extract: %v %v", v, ok)
	}
	_ = res
}

func TestExecute_TimeoutExceeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: srv.URL, Method: "GET"}, Timeout: "PT0.05S"}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	_ = res
}

func TestExecute_RetryOnStatus(t *testing.T) {
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count++
		if count < 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{
		StepID:     "s1",
		WorkflowID: "wf",
		Request:    &compiler.CompiledRequest{Target: srv.URL, Method: "GET"},
		Retry:      &compiler.CompiledRetry{MaxAttempts: 3, Backoff: "fixed", Delay: "PT0S", RetryOn: []string{"503"}},
	}
	res, err := exec.executeWithRetry(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", res.State)
	}
}

func TestExecute_MaxRetriesExceeded(t *testing.T) {
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count++
		w.WriteHeader(503)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{
		StepID:     "s1",
		WorkflowID: "wf",
		Request:    &compiler.CompiledRequest{Target: srv.URL, Method: "GET"},
		Retry:      &compiler.CompiledRetry{MaxAttempts: 2, Backoff: "fixed", Delay: "PT0S", RetryOn: []string{"503"}},
	}
	res, err := exec.executeWithRetry(context.Background(), step, exCtx)
	if err == nil {
		t.Fatalf("expected max retries error")
	}
	_ = res
}

func TestExecute_NilRequest(t *testing.T) {
	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	_, err := exec.Execute(context.Background(), &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf"}, exCtx)
	if err == nil {
		t.Fatalf("expected unsupported protocol error")
	}
}

func TestExecute_RedirectFollowed(t *testing.T) {
	// final server
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("done"))
	}))
	defer final.Close()
	// redirect server
	head := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusMovedPermanently)
	}))
	defer head.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: head.URL, Method: "GET"}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateSucceeded {
		t.Fatalf("expected succeeded")
	}
}
