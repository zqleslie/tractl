package extract_test

import (
	"testing"
	"time"

	"github.com/tractl/tractl/internal/extract"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/spec"
)

// makeResult returns a StepResult populated with sensible test defaults.
func makeResult() *runtime.StepResult {
	return &runtime.StepResult{
		StepID: "step-1",
		Status: 200,
		Headers: map[string]string{
			"content-type": "application/json",
			"x-request-id": "abc-123",
		},
		Body: []byte(`{"token":"abc123","user":{"id":"42"}}`),
		Metadata: map[string]string{
			"url":     "https://example.com/api",
			"method":  "POST",
			"attempt": "1",
		},
		Timeline: &runtime.RequestTimeline{
			DNSResolution:    5 * time.Millisecond,
			TCPConnect:       10 * time.Millisecond,
			TLSHandshake:     20 * time.Millisecond,
			RequestSent:      1 * time.Millisecond,
			TimeToFirstByte:  100 * time.Millisecond,
			ResponseTransfer: 6 * time.Millisecond,
			TotalDuration:    142 * time.Millisecond,
		},
	}
}

func makeExtract(id, source, path, as string, scope spec.ExtractScope) spec.Extract {
	return spec.Extract{ID: id, Source: source, Path: path, As: as, Scope: scope}
}

func newCtx(t *testing.T) *runtime.ExecutionContext {
	t.Helper()
	ec, err := runtime.NewExecutionContext("wf-test", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error creating ExecutionContext: %v", err)
	}
	return ec
}

func TestExtractEngine(t *testing.T) {
	eng := extract.New()
	workflowStart := time.Now()

	tests := []struct {
		name         string
		extract      spec.Extract
		result       *runtime.StepResult
		wantValue    string
		wantKey      string
		wantScope    string // "spec" | "workflow" | "env" | "step"
		wantErrCode  extract.ErrorCode
		wantResolved bool
	}{
		// ── Status ─────────────────────────────────────────────────────────────
		{
			name:         "status/201",
			extract:      makeExtract("e1", "status", "", "statusCode", spec.ExtractScopeWorkflow),
			result:       func() *runtime.StepResult { r := makeResult(); r.Status = 201; return r }(),
			wantValue:    "201",
			wantKey:      "statusCode",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "status/zero",
			extract:      makeExtract("e2", "status", "", "s", spec.ExtractScopeWorkflow),
			result:       func() *runtime.StepResult { r := makeResult(); r.Status = 0; return r }(),
			wantValue:    "0",
			wantKey:      "s",
			wantScope:    "workflow",
			wantResolved: true,
		},

		// ── Header ─────────────────────────────────────────────────────────────
		{
			name:         "header/present-lowercase",
			extract:      makeExtract("e3", "header", "content-type", "ct", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "application/json",
			wantKey:      "ct",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "header/case-insensitive",
			extract:      makeExtract("e4", "header", "Content-Type", "ct2", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "application/json",
			wantKey:      "ct2",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:        "header/missing",
			extract:     makeExtract("e5", "header", "x-missing", "miss", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrPathNotFound,
		},

		// ── Body ───────────────────────────────────────────────────────────────
		{
			name:         "body/top-level-key",
			extract:      makeExtract("e6", "body", "token", "tok", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "abc123",
			wantKey:      "tok",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "body/nested-path",
			extract:      makeExtract("e7", "body", "user.id", "uid", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "42",
			wantKey:      "uid",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:        "body/missing-path",
			extract:     makeExtract("e8", "body", "missing.key", "mk", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrPathNotFound,
		},
		{
			name:        "body/nil-body",
			extract:     makeExtract("e9", "body", "token", "tok", spec.ExtractScopeWorkflow),
			result:      func() *runtime.StepResult { r := makeResult(); r.Body = nil; return r }(),
			wantErrCode: extract.ErrPathNotFound,
		},

		// ── Metadata ───────────────────────────────────────────────────────────
		{
			name:         "metadata/url",
			extract:      makeExtract("e10", "metadata", "url", "reqURL", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "https://example.com/api",
			wantKey:      "reqURL",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "metadata/method",
			extract:      makeExtract("e11", "metadata", "method", "meth", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "POST",
			wantKey:      "meth",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:        "metadata/unknown-field",
			extract:     makeExtract("e12", "metadata", "unknown", "unk", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrMetadataField,
		},

		// ── Timing ─────────────────────────────────────────────────────────────
		{
			name:         "timing/total",
			extract:      makeExtract("e13", "timing", "total", "tot", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "142ms",
			wantKey:      "tot",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "timing/dns",
			extract:      makeExtract("e14", "timing", "dns", "dnsTime", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "5ms",
			wantKey:      "dnsTime",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:        "timing/bad-field",
			extract:     makeExtract("e15", "timing", "bad", "b", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrTimingField,
		},

		// ── Scope ──────────────────────────────────────────────────────────────
		{
			name:         "scope/step",
			extract:      makeExtract("e16", "status", "", "sc", spec.ExtractScopeStep),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "sc",
			wantScope:    "step",
			wantResolved: true,
		},
		{
			name:         "scope/workflow-explicit",
			extract:      makeExtract("e17", "status", "", "sc", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "sc",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "scope/spec",
			extract:      makeExtract("e18", "status", "", "sc", spec.ExtractScopeSpec),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "sc",
			wantScope:    "spec",
			wantResolved: true,
		},
		{
			name:         "scope/empty-defaults-to-workflow",
			extract:      makeExtract("e19", "status", "", "sc", ""),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "sc",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:        "scope/invalid",
			extract:     makeExtract("e20", "status", "", "sc", "invalid"),
			result:      makeResult(),
			wantErrCode: extract.ErrInvalidScope,
		},

		// ── As binding ─────────────────────────────────────────────────────────
		{
			name:         "as/explicit-name",
			extract:      makeExtract("e21", "status", "", "myToken", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "myToken",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "as/defaults-to-id",
			extract:      makeExtract("extractID-22", "status", "", "", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "200",
			wantKey:      "extractID-22",
			wantScope:    "workflow",
			wantResolved: true,
		},

		// ── Extension stub ─────────────────────────────────────────────────────
		{
			name:        "extension/stub",
			extract:     makeExtract("e23", "extension", "", "x", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrExtensionStub,
		},

		// ── Unknown source ─────────────────────────────────────────────────────
		{
			name:        "source/unknown",
			extract:     makeExtract("e24", "grpc", "", "x", spec.ExtractScopeWorkflow),
			result:      makeResult(),
			wantErrCode: extract.ErrUnknownSource,
		},

		// ── Timing sub-fields ──────────────────────────────────────────────────
		{
			name:         "timing/tcp",
			extract:      makeExtract("e25", "timing", "tcp", "tcpTime", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "10ms",
			wantKey:      "tcpTime",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "timing/tls",
			extract:      makeExtract("e26", "timing", "tls", "tlsTime", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "20ms",
			wantKey:      "tlsTime",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "timing/ttfb",
			extract:      makeExtract("e27", "timing", "ttfb", "ttfbTime", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "100ms",
			wantKey:      "ttfbTime",
			wantScope:    "workflow",
			wantResolved: true,
		},
		{
			name:         "timing/transfer",
			extract:      makeExtract("e28", "timing", "transfer", "transferTime", spec.ExtractScopeWorkflow),
			result:       makeResult(),
			wantValue:    "6ms",
			wantKey:      "transferTime",
			wantScope:    "workflow",
			wantResolved: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newCtx(t)
			res, ev, err := eng.Extract(tc.extract, tc.result, ctx, "trace-1", "span-1", workflowStart)

			if tc.wantErrCode != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErrCode)
				}
				ee, ok := err.(*extract.ExtractError)
				if !ok {
					t.Fatalf("expected *extract.ExtractError, got %T: %v", err, err)
				}
				if ee.Code != tc.wantErrCode {
					t.Errorf("error code: got %q, want %q", ee.Code, tc.wantErrCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !res.Resolved {
				t.Error("Resolved should be true on success")
			}
			if res.ExtractID != tc.extract.ID {
				t.Errorf("ExtractID: got %q, want %q", res.ExtractID, tc.extract.ID)
			}
			if res.Source != tc.extract.Source {
				t.Errorf("Source: got %q, want %q", res.Source, tc.extract.Source)
			}
			expectedPlaceholder := "[extract:" + tc.extract.ID + "]"
			if res.ValuePlaceholder != expectedPlaceholder {
				t.Errorf("ValuePlaceholder: got %q, want %q", res.ValuePlaceholder, expectedPlaceholder)
			}
			if res.Duration < 0 {
				t.Error("Duration must be non-negative")
			}

			// Verify value was written to the correct scope
			var got string
			var ok bool
			if tc.wantScope == "step" {
				got, ok = ctx.GetExtract(tc.result.StepID, tc.wantKey)
			} else {
				got, ok = ctx.GetVar(tc.wantScope, tc.wantKey)
			}
			if !ok {
				t.Errorf("key %q not found in scope %q", tc.wantKey, tc.wantScope)
				return
			}
			if got != tc.wantValue {
				t.Errorf("context value: got %q, want %q", got, tc.wantValue)
			}

			if ev.Variable != tc.wantKey {
				t.Errorf("EvalEvent.Variable: got %q, want %q", ev.Variable, tc.wantKey)
			}
			if ev.ValuePlaceholder != expectedPlaceholder {
				t.Errorf("EvalEvent.ValuePlaceholder: got %q, want %q", ev.ValuePlaceholder, expectedPlaceholder)
			}
			if ev.MonotonicOffset < 0 {
				t.Error("EvalEvent.MonotonicOffset must be non-negative")
			}
			if ev.ValuePlaceholder == tc.wantValue {
				t.Error("EvalEvent.ValuePlaceholder must not equal the actual extracted value")
			}
		})
	}
}

// TestExtractContextIntegration verifies that a subsequent GetVar call returns
// the bound value after a successful Extract.
func TestExtractContextIntegration(t *testing.T) {
	eng := extract.New()
	ctx := newCtx(t)
	result := makeResult()
	result.Body = []byte(`{"auth":{"token":"secret-jwt"}}`)

	e := spec.Extract{
		ID:     "authToken",
		Source: "body",
		Path:   "auth.token",
		As:     "myToken",
		Scope:  spec.ExtractScopeWorkflow,
	}

	_, _, err := eng.Extract(e, result, ctx, "t1", "s1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := ctx.GetVar("workflow", "myToken")
	if !ok {
		t.Fatal("expected 'myToken' in workflow scope")
	}
	if v != "secret-jwt" {
		t.Errorf("got %q, want %q", v, "secret-jwt")
	}
}

// TestEvalEventCredentialSafety verifies the actual extracted value never
// appears in the EvalEvent.
func TestEvalEventCredentialSafety(t *testing.T) {
	eng := extract.New()
	ctx := newCtx(t)
	result := makeResult()
	result.Headers["authorization"] = "Bearer super-secret-token"

	e := spec.Extract{
		ID:     "authHeader",
		Source: "header",
		Path:   "authorization",
		As:     "authVal",
		Scope:  spec.ExtractScopeWorkflow,
	}

	_, ev, err := eng.Extract(e, result, ctx, "t1", "s1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.ValuePlaceholder == "Bearer super-secret-token" {
		t.Error("EvalEvent.ValuePlaceholder must not contain the actual credential")
	}
	if ev.ValuePlaceholder != "[extract:authHeader]" {
		t.Errorf("unexpected placeholder: %q", ev.ValuePlaceholder)
	}
}
