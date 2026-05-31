package assertion

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/spec"
)

// credentialHeaders is the set of header names whose values must be masked in observability events.
var credentialHeaders = map[string]bool{
	"authorization":       true,
	"x-api-key":           true,
	"cookie":              true,
	"set-cookie":          true,
	"proxy-authorization": true,
}

// Evaluator evaluates a single spec.Assertion against a StepResult.
type Evaluator interface {
	Evaluate(
		ctx context.Context,
		a spec.Assertion,
		result *runtime.StepResult,
		execCtx *runtime.ExecutionContext,
		traceID, spanID string,
		workflowStart time.Time,
	) (AssertionResult, EvalEvent, error)
}

// AssertionEvaluator implements Evaluator.
type AssertionEvaluator struct{} //nolint:revive

// Compile-time interface check.
var _ Evaluator = (*AssertionEvaluator)(nil)

// New returns an initialised *AssertionEvaluator.
func New() *AssertionEvaluator {
	return &AssertionEvaluator{}
}

// Evaluate dispatches on a.Kind and returns the assessment result plus an observability event.
func (ae *AssertionEvaluator) Evaluate(
	evalCtx context.Context,
	a spec.Assertion,
	result *runtime.StepResult,
	ctx *runtime.ExecutionContext,
	traceID, spanID string,
	workflowStart time.Time,
) (AssertionResult, EvalEvent, error) {
	// evalCtx is reserved for future cancellation support in assertion evaluation.
	_ = evalCtx
	return ae.evaluate(a.ID, a.ULID, a.Kind, a.Op, a.Target, a.Expected, a.Severity, nil, result, ctx, traceID, spanID, workflowStart)
}

// EvaluateCompiled is like Evaluate but accepts a pre-compiled *regexp.Regexp
// for Op=="matches" assertions, avoiding recompilation on every call.
func (ae *AssertionEvaluator) EvaluateCompiled(
	a compiler.CompiledAssertion,
	result *runtime.StepResult,
	ctx *runtime.ExecutionContext,
	traceID, spanID string,
	workflowStart time.Time,
) (AssertionResult, EvalEvent, error) {
	return ae.evaluate(a.ID, "", a.Kind, a.Op, a.Target, a.Expected, spec.AssertionSeverity(a.Severity), a.CompiledPattern, result, ctx, traceID, spanID, workflowStart)
}

// evaluate is the shared implementation for both Evaluate and EvaluateCompiled.
func (ae *AssertionEvaluator) evaluate(
	id, ulid, kind, op, target string,
	expected interface{},
	severity spec.AssertionSeverity,
	compiledPattern *regexp.Regexp,
	result *runtime.StepResult,
	ctx *runtime.ExecutionContext,
	traceID, spanID string,
	workflowStart time.Time,
) (AssertionResult, EvalEvent, error) {
	start := time.Now()

	if severity == "" {
		severity = spec.SeverityError
	}

	expectedStr, resolveErr := resolveExpected(expected, ctx)

	var (
		outcome     AssertionOutcome
		message     string
		placeholder string
		evalErr     error
	)

	// Reconstruct a minimal spec.Assertion for the kind-specific helpers.
	a := spec.Assertion{ID: id, ULID: ulid, Kind: kind, Op: op, Target: target, Severity: severity}

	switch kind {
	case "status":
		if resolveErr != nil {
			evalErr = resolveErr
			outcome = OutcomeFail
			message = fmt.Sprintf("expression resolution failed: %v", resolveErr)
		} else {
			outcome, message, placeholder, evalErr = evalStatus(a, result, expectedStr)
		}

	case "header":
		if resolveErr != nil {
			evalErr = resolveErr
			outcome = OutcomeFail
			message = fmt.Sprintf("expression resolution failed: %v", resolveErr)
			placeholder = headerPlaceholder(a.Target)
		} else {
			outcome, message, placeholder, evalErr = evalHeader(a, result, expectedStr, compiledPattern)
		}

	case "body":
		if resolveErr != nil {
			evalErr = resolveErr
			outcome = OutcomeFail
			message = fmt.Sprintf("expression resolution failed: %v", resolveErr)
			placeholder = "response.body." + a.Target
		} else {
			outcome, message, placeholder, evalErr = evalBody(a, result, expectedStr, compiledPattern)
		}

	case "schema":
		outcome = OutcomeSkipped
		message = "schema validation is not yet implemented"
		placeholder = "response.body"
		evalErr = &AssertionError{Code: ErrSchemaStub, Message: message}

	case "script":
		outcome = OutcomeSkipped
		message = "script evaluation is not yet implemented"
		placeholder = "script"
		evalErr = &AssertionError{Code: ErrScriptStub, Message: message}

	case "extension":
		outcome = OutcomeSkipped
		message = "extension evaluation is not yet implemented"
		placeholder = "extension"
		evalErr = &AssertionError{Code: ErrExtensionStub, Message: message}

	default:
		outcome = OutcomeFail
		message = fmt.Sprintf("unknown assertion kind %q", kind)
		placeholder = "unknown"
		evalErr = &AssertionError{Code: ErrUnknownKind, Message: message}
	}

	dur := time.Since(start)
	causesFailure := outcome == OutcomeFail && severity == spec.SeverityError

	ar := AssertionResult{
		AssertionID:   id,
		AssertionULID: ulid,
		Kind:          kind,
		Outcome:       outcome,
		Severity:      severity,
		Message:       message,
		CausesFailure: causesFailure,
		Duration:      dur,
	}

	ev := EvalEvent{
		TraceID:           traceID,
		SpanID:            spanID,
		AssertionID:       id,
		Kind:              kind,
		Outcome:           outcome,
		TargetPlaceholder: placeholder,
		MonotonicOffset:   time.Since(workflowStart),
	}

	return ar, ev, evalErr
}

// resolveExpected converts a.Expected to string and resolves any ${…} expressions.
func resolveExpected(expected interface{}, ctx *runtime.ExecutionContext) (string, error) {
	if expected == nil {
		return "", nil
	}
	var s string
	switch v := expected.(type) {
	case string:
		s = v
	case int:
		s = strconv.Itoa(v)
	case float64:
		s = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		s = strconv.FormatBool(v)
	default:
		s = fmt.Sprintf("%v", v)
	}
	if ctx != nil {
		return ctx.ResolveAll(s)
	}
	return s, nil
}

// ── status ────────────────────────────────────────────────────────────────────

func evalStatus(a spec.Assertion, result *runtime.StepResult, expected string) (AssertionOutcome, string, string, error) {
	placeholder := "response.status"

	switch a.Op {
	case "equals":
		exp, err := strconv.Atoi(expected)
		if err != nil {
			return OutcomeFail, fmt.Sprintf("expected value %q is not a valid integer", expected), placeholder,
				&AssertionError{Code: ErrOperatorMismatch, Message: "status equals requires numeric expected value"}
		}
		if result.Status == exp {
			return OutcomePass, fmt.Sprintf("status %d equals %d", result.Status, exp), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("expected status %d, got %d", exp, result.Status), placeholder, nil

	case "inRange":
		low, high, err := parseRange(expected)
		if err != nil {
			return OutcomeFail, err.Error(), placeholder,
				&AssertionError{Code: ErrOperatorMismatch, Message: err.Error()}
		}
		if result.Status >= low && result.Status <= high {
			return OutcomePass, fmt.Sprintf("status %d is in range %d-%d", result.Status, low, high), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("status %d is not in range %d-%d", result.Status, low, high), placeholder, nil

	case "exists":
		if result.Status != 0 {
			return OutcomePass, fmt.Sprintf("status %d exists", result.Status), placeholder, nil
		}
		return OutcomeFail, "status does not exist (0)", placeholder, nil

	default:
		msg := fmt.Sprintf("unknown operator %q for kind %q", a.Op, a.Kind)
		return OutcomeFail, msg, placeholder, &AssertionError{Code: ErrUnknownOperator, Message: msg}
	}
}

// ── header ────────────────────────────────────────────────────────────────────

func evalHeader(a spec.Assertion, result *runtime.StepResult, expected string, compiledPattern *regexp.Regexp) (AssertionOutcome, string, string, error) {
	name := strings.ToLower(a.Target)
	placeholder := headerPlaceholder(name)

	headerVal, present := result.Headers[name]

	switch a.Op {
	case "equals":
		if !present {
			return OutcomeFail, fmt.Sprintf("header %q not present", a.Target), placeholder, nil
		}
		if headerVal == expected {
			return OutcomePass, fmt.Sprintf("header %q equals %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("header %q: expected %q, got %q", a.Target, expected, headerVal), placeholder, nil

	case "matches":
		if !present {
			return OutcomeFail, fmt.Sprintf("header %q not present", a.Target), placeholder, nil
		}
		rx := compiledPattern
		if rx == nil {
			var err error
			rx, err = regexp.Compile(expected)
			if err != nil {
				msg := fmt.Sprintf("invalid regex %q: %v", expected, err)
				return OutcomeFail, msg, placeholder, &AssertionError{Code: ErrRegexInvalid, Message: msg}
			}
		}
		if rx.MatchString(headerVal) {
			return OutcomePass, fmt.Sprintf("header %q matches %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("header %q value %q does not match %q", a.Target, headerVal, expected), placeholder, nil

	case "contains":
		if !present {
			return OutcomeFail, fmt.Sprintf("header %q not present", a.Target), placeholder, nil
		}
		if strings.Contains(headerVal, expected) {
			return OutcomePass, fmt.Sprintf("header %q contains %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("header %q value %q does not contain %q", a.Target, headerVal, expected), placeholder, nil

	case "exists":
		if present {
			return OutcomePass, fmt.Sprintf("header %q is present", a.Target), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("header %q is not present", a.Target), placeholder, nil

	default:
		msg := fmt.Sprintf("unknown operator %q for kind %q", a.Op, a.Kind)
		return OutcomeFail, msg, placeholder, &AssertionError{Code: ErrUnknownOperator, Message: msg}
	}
}

// headerPlaceholder returns the TargetPlaceholder string, masking credential headers.
func headerPlaceholder(lowerName string) string {
	if credentialHeaders[lowerName] {
		return "response.headers." + lowerName + "[masked]"
	}
	return "response.headers." + lowerName
}

// ── body ──────────────────────────────────────────────────────────────────────

func evalBody(a spec.Assertion, result *runtime.StepResult, expected string, compiledPattern *regexp.Regexp) (AssertionOutcome, string, string, error) {
	placeholder := "response.body." + a.Target

	if len(result.Body) == 0 {
		return OutcomeFail, "response body is empty", placeholder, nil
	}

	gResult := gjson.GetBytes(result.Body, a.Target)

	switch a.Op {
	case "equals", "jsonpath":
		if gResult.String() == expected {
			return OutcomePass, fmt.Sprintf("body.%s equals %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("body.%s: expected %q, got %q", a.Target, expected, gResult.String()), placeholder, nil

	case "matches":
		rx := compiledPattern
		if rx == nil {
			var err error
			rx, err = regexp.Compile(expected)
			if err != nil {
				msg := fmt.Sprintf("invalid regex %q: %v", expected, err)
				return OutcomeFail, msg, placeholder, &AssertionError{Code: ErrRegexInvalid, Message: msg}
			}
		}
		if rx.MatchString(gResult.String()) {
			return OutcomePass, fmt.Sprintf("body.%s matches %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("body.%s value %q does not match %q", a.Target, gResult.String(), expected), placeholder, nil

	case "contains":
		if strings.Contains(gResult.String(), expected) {
			return OutcomePass, fmt.Sprintf("body.%s contains %q", a.Target, expected), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("body.%s value %q does not contain %q", a.Target, gResult.String(), expected), placeholder, nil

	case "exists":
		if gResult.Exists() && gResult.Type != gjson.Null {
			return OutcomePass, fmt.Sprintf("body.%s exists", a.Target), placeholder, nil
		}
		return OutcomeFail, fmt.Sprintf("body.%s does not exist or is null", a.Target), placeholder, nil

	default:
		msg := fmt.Sprintf("unknown operator %q for kind %q", a.Op, a.Kind)
		return OutcomeFail, msg, placeholder, &AssertionError{Code: ErrUnknownOperator, Message: msg}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// rangeRE matches exactly "NNN-NNN" (three or more digits each side).
var rangeRE = regexp.MustCompile(`^(\d+)-(\d+)$`)

// parseRange parses an "low-high" range string (e.g. "200-299") into two ints.
// The format must be "<integer>-<integer>" with no spaces or other characters.
func parseRange(s string) (int, int, error) {
	m := rangeRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, 0, fmt.Errorf("inRange expected format \"<low>-<high>\" (e.g. \"200-299\"), got %q", s)
	}
	low, _ := strconv.Atoi(m[1])
	high, _ := strconv.Atoi(m[2])
	if low > high {
		return 0, 0, fmt.Errorf("inRange: low bound %d must not exceed high bound %d", low, high)
	}
	return low, high, nil
}
