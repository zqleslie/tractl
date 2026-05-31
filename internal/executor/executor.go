package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/extract"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/spec"
)

// Executor interface allows dependency injection (for testing).
type Executor interface {
	Execute(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error)
}

// context key for storing TimingRecorder
type ctxKeyTimingRecorder struct{}

// NewHTTPExecutor constructs an HTTPExecutor. If client==nil a default is built.
func NewHTTPExecutor(client *http.Client) *HTTPExecutor {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &HTTPExecutor{client: client, masker: NewCredentialMasker()}
}

// HTTPExecutor is responsible for executing compiled http steps.
type HTTPExecutor struct {
	client *http.Client
	masker *CredentialMasker
}

func (h *HTTPExecutor) doRequest(req *http.Request) (resp *http.Response, err error) {
	if goruntime.GOOS == "js" {
		defer func() {
			if recovered := recover(); recovered != nil {
				resp = nil
				err = fmt.Errorf("http request panic: %v", recovered)
			}
		}()
	}
	return h.client.Do(req)
}

// Execute executes a single step. If step.Retry is non-nil with MaxAttempts > 1,
// retry policy is applied transparently. Callers should use Execute exclusively;
// executeOnce is the inner single-attempt method.
func (h *HTTPExecutor) Execute(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error) {
	if step != nil && step.Retry != nil && step.Retry.MaxAttempts > 1 {
		return h.executeWithRetry(ctx, step, execCtx)
	}
	return h.executeOnce(ctx, step, execCtx)
}

// executeOnce performs a single attempt and populates execCtx with extracts and result.
// This is the inner method that retry logic wraps; it must NOT itself apply retry.
func (h *HTTPExecutor) executeOnce(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error) {
	if step == nil || step.Request == nil {
		return nil, &ExecutorError{Code: ErrUnsupportedProtocol, StepID: "", WorkflowID: "", Message: "unsupported or missing request"}
	}
	// handle timeout
	if step.Timeout != "" {
		dur, err := parseISODuration(step.Timeout)
		if err != nil {
			return failedResult(step), &ExecutorError{
				Code:       ErrRequestBuildFailed,
				StepID:     step.StepID,
				WorkflowID: step.WorkflowID,
				Message:    fmt.Sprintf("invalid timeout %q: %v (supported format: PTnS, e.g. PT30S)", step.Timeout, err),
			}
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dur)
		defer cancel()
	}

	// build evaluator
	resolver := runtime.NewVariableResolver(execCtx)
	eval := runtime.NewExpressionEvaluator(resolver, execCtx)
	builder := NewRequestBuilder(eval)

	// build timing recorder
	recorder := NewTimingRecorder()
	// put recorder into context for CheckRedirect to pick up
	ctx = context.WithValue(ctx, ctxKeyTimingRecorder{}, recorder)

	ctx = attachHTTPTrace(ctx, recorder)

	// build request
	req, err := builder.Build(ctx, step)
	if err != nil {
		return failedResult(step), &ExecutorError{Code: ErrRequestBuildFailed, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: err.Error()}
	}

	// mask headers for logging/observation (do not mutate req.Header)
	_ = h.masker.MaskHeaders(req.Header)

	// perform request
	resp, doErr := h.doRequest(req)
	if doErr != nil {
		if errors.Is(doErr, context.DeadlineExceeded) || strings.Contains(doErr.Error(), "context deadline") {
			return failedResult(step), &ExecutorError{Code: ErrTimeout, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: doErr.Error()}
		}
		return failedResult(step), &ExecutorError{Code: ErrRequestFailed, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: doErr.Error()}
	}
	// read body
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	// record body done timestamp
	recorder.bodyDone = time.Now()

	// build timeline
	tl := recorder.ToRequestTimeline(resp.StatusCode, int64(len(bodyBytes)))

	// canonicalise response headers to lowercase keys
	respHeaders := make(map[string]string, len(resp.Header))
	for k, vals := range resp.Header {
		if len(vals) > 0 {
			respHeaders[strings.ToLower(k)] = vals[0]
		}
	}

	// build the partial result with HTTP fields so assertion/extract engines can read them
	workflowStart := execCtx.StartedAt()
	res := &runtime.StepResult{
		StepID:     step.StepID,
		WorkflowID: step.WorkflowID,
		State:      runtime.StateSucceeded,
		Extracts:   make(map[string]string),
		StartedAt:  recorder.wallClockStart,
		FinishedAt: time.Now(),
		Timeline:   tl,
		Status:     resp.StatusCode,
		Headers:    respHeaders,
		Body:       bodyBytes,
		Metadata: map[string]string{
			"url":     req.URL.String(),
			"method":  req.Method,
			"attempt": "1",
		},
	}

	// run extracts via the engine (gjson body extraction, scope binding, observability).
	// The engine writes raw values into execCtx via SetStepExtract / SetVar.
	// We then read the real value back and store it in res.Extracts.
	extractEng := extract.New()
	for _, ex := range step.Extracts {
		specEx := spec.Extract{
			ID:     ex.ID,
			Source: ex.Source,
			Path:   ex.Path,
			As:     ex.As,
			Scope:  spec.ExtractScope(ex.Scope),
		}
		extractResult, _, _ := extractEng.Extract(specEx, res, execCtx, "", "", workflowStart)
		if extractResult.Resolved {
			// Read the real value back from execCtx (the engine wrote it there).
			if v, ok := execCtx.GetExtract(step.StepID, ex.As); ok {
				res.Extracts[ex.As] = v
			}
		}
	}

	// run all assertions via the evaluator (all kinds, all operators, severity-aware)
	assertEval := assertion.New()
	for _, a := range step.Assertions {
		ar, _, _ := assertEval.EvaluateCompiled(a, res, execCtx, "", "", workflowStart)
		if ar.CausesFailure {
			res.State = runtime.StateFailed
		}
	}

	// persist final state into execCtx (mirrors state and stores result pointer)
	execCtx.SetStepResult(res)

	return res, nil
}

// failedResult returns a minimal failed StepResult for error return paths.
func failedResult(step *compiler.CompiledStep) *runtime.StepResult {
	now := time.Now()
	return &runtime.StepResult{
		StepID:     step.StepID,
		WorkflowID: step.WorkflowID,
		State:      runtime.StateFailed,
		Extracts:   map[string]string{},
		StartedAt:  now,
		FinishedAt: now,
	}
}

// parseISODuration supports a subset like PT0.05S
func parseISODuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "PT") || !strings.HasSuffix(s, "S") {
		return 0, fmt.Errorf("unsupported duration format: %s", s)
	}
	num := strings.TrimSuffix(strings.TrimPrefix(s, "PT"), "S")
	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, err
	}
	secs := time.Duration(math.Round(f*1e9)) * time.Nanosecond
	return secs, nil
}

// executeWithRetry wraps executeOnce with retry logic as per step.Retry.
// It must NOT call Execute (which would re-enter retry and recurse infinitely).
func (h *HTTPExecutor) executeWithRetry(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error) {
	if step.Retry == nil || step.Retry.MaxAttempts <= 1 {
		return h.executeOnce(ctx, step, execCtx)
	}
	maxAttempts := step.Retry.MaxAttempts
	delayDur, _ := parseISODuration(step.Retry.Delay)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res, err := h.executeOnce(ctx, step, execCtx)

		// context cancelled?
		if ctx.Err() != nil {
			return res, &ExecutorError{Code: ErrTimeout, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: ctx.Err().Error()}
		}

		// determine if we should retry
		shouldRetry := false
		// network error retry
		if err != nil {
			for _, ro := range step.Retry.RetryOn {
				if ro == "network_error" {
					shouldRetry = true
					break
				}
			}
		} else {
			// check status code
			statusCode := 0
			if res != nil && res.Timeline != nil {
				statusCode = res.Timeline.StatusCode
			}
			for _, ro := range step.Retry.RetryOn {
				if ro == strconv.Itoa(statusCode) {
					shouldRetry = true
					break
				}
			}
		}

		// if not retrying, return last result or error
		if !shouldRetry {
			if err != nil {
				return res, err
			}
			return res, nil
		}

		// if this was the last attempt, break and return max retries error
		if attempt == maxAttempts {
			break
		}

		// compute backoff
		var sleep time.Duration
		switch step.Retry.Backoff {
		case "linear":
			sleep = time.Duration(attempt) * delayDur
		case "exponential":
			sleep = time.Duration(math.Pow(2, float64(attempt-1))) * delayDur
		default: // fixed
			sleep = delayDur
		}
		if sleep > 60*time.Second {
			sleep = 60 * time.Second
		}

		// ensure context has enough budget
		select {
		case <-time.After(sleep):
			// continue
		case <-ctx.Done():
			return res, &ExecutorError{Code: ErrTimeout, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: ctx.Err().Error()}
		}
	}
	return nil, &ExecutorError{Code: ErrMaxRetriesExceeded, StepID: step.StepID, WorkflowID: step.WorkflowID, Message: "max retries exhausted"}
}
