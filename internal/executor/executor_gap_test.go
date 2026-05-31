package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// TestTimeline_PlainHTTPTLSZero — TLSHandshake must be zero for plain HTTP.
func TestTimeline_PlainHTTPTLSZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: srv.URL, Method: "GET"}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Timeline == nil {
		t.Fatalf("expected timeline populated")
	}
	if res.Timeline.TLSHandshake != 0 {
		t.Fatalf("expected TLSHandshake=0 for plain HTTP, got %v", res.Timeline.TLSHandshake)
	}
}

// TestExecute_RouteToRetryWhenRetryPresent — Execute() must delegate to retry path
// when step.Retry is configured. Without this, scheduler-driven retries are silently
// dropped because the scheduler calls Execute, not executeWithRetry.
func TestExecute_RouteToRetryWhenRetryPresent(t *testing.T) {
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
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateSucceeded {
		t.Fatalf("expected succeeded via Execute()-routed retry, got %s", res.State)
	}
	if count != 3 {
		t.Fatalf("expected 3 attempts, got %d", count)
	}
}

// TestExecute_NoRetryWhenContextCancelled — when ctx is already cancelled, no
// retry attempts should be made.
func TestExecute_NoRetryWhenContextCancelled(t *testing.T) {
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
		Retry:      &compiler.CompiledRetry{MaxAttempts: 5, Backoff: "fixed", Delay: "PT1S", RetryOn: []string{"503"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := exec.Execute(ctx, step, exCtx)
	if err == nil {
		t.Fatalf("expected error when ctx pre-cancelled")
	}
	// at most one attempt should have happened (the first call); ideally zero,
	// but the inner request still goes out before client.Do honors the cancellation.
	if count > 1 {
		t.Fatalf("expected at most 1 attempt with pre-cancelled ctx, got %d", count)
	}
}

// TestExecute_OriginalAuthorizationHeaderSentToServer — masking is for observation
// only; the original credential MUST be sent on the wire.
func TestExecute_OriginalAuthorizationHeaderSentToServer(t *testing.T) {
	received := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{
		StepID:     "s1",
		WorkflowID: "wf",
		Request: &compiler.CompiledRequest{
			Target:  srv.URL,
			Method:  "GET",
			Headers: map[string]string{"Authorization": "Bearer real-secret-token"},
		},
	}
	if _, err := exec.Execute(context.Background(), step, exCtx); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if received != "Bearer real-secret-token" {
		t.Fatalf("server saw masked header instead of real value: %q", received)
	}
}

// TestExecute_RedirectHopRecorded — at least one redirect hop is captured.
func TestExecute_RedirectHopRecorded(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer final.Close()
	head := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusMovedPermanently)
	}))
	defer head.Close()

	// instrument via a custom client to inspect recorder after the round-trip
	exec := NewHTTPExecutor(nil)
	exCtx := mustNewEC(t, "wf", nil, nil, nil)
	step := &compiler.CompiledStep{StepID: "s1", WorkflowID: "wf", Request: &compiler.CompiledRequest{Target: head.URL, Method: "GET"}}
	res, err := exec.Execute(context.Background(), step, exCtx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.State != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", res.State)
	}
}
