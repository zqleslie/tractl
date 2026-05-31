// expr.go defines code for the runtime package.

package runtime

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var exprRE = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpressionEvaluator replaces ${...} patterns using the provided resolver and context.
type ExpressionEvaluator struct {
	resolver *VariableResolver
	ctx      *ExecutionContext
}

// NewExpressionEvaluator creates a new evaluator.
func NewExpressionEvaluator(resolver *VariableResolver, ctx *ExecutionContext) *ExpressionEvaluator {
	return &ExpressionEvaluator{resolver: resolver, ctx: ctx}
}

// Evaluate finds all ${...} and resolves them. Returns error on unresolvable patterns.
func (e *ExpressionEvaluator) Evaluate(ctx context.Context, expression string) (string, error) {
	// ctx is reserved for future use (e.g. cancellation of long-running evaluations).
	// It is not currently used internally but is part of the public API per standard 6.2.
	_ = ctx
	if expression == "" {
		return "", nil
	}
	out := expression
	matches := exprRE.FindAllStringSubmatch(expression, -1)
	for _, m := range matches {
		raw := m[0]
		inner := strings.TrimSpace(m[1])

		// support default fallback: <expr> ?? "default"
		if strings.Contains(inner, "??") {
			parts := strings.SplitN(inner, "??", 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// strip quotes from right if present
			right = strings.Trim(right, " \t\n\r\"")
			val, err := e.resolveInner(left)
			if err == nil && val != "" {
				out = strings.ReplaceAll(out, raw, val)
			} else {
				out = strings.ReplaceAll(out, raw, right)
			}
			continue
		}

		val, err := e.resolveInner(inner)
		if err != nil {
			return "", err
		}
		out = strings.ReplaceAll(out, raw, val)
	}
	return out, nil
}

// EvaluateMap evaluates all values in a map and returns a new map.
func (e *ExpressionEvaluator) EvaluateMap(ctx context.Context, m map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(m))
	for k, v := range m {
		val, err := e.Evaluate(ctx, v)
		if err != nil {
			return nil, err
		}
		out[k] = val
	}
	return out, nil
}

// resolveInner resolves a single inner expression (without ${}).
func (e *ExpressionEvaluator) resolveInner(inner string) (string, error) {
	// patterns:
	// vars.<key>
	// steps.<stepId>.extracts.<extractId>
	// steps.<stepId>.response.status
	// steps.<stepId>.response.body.<path> (stub)

	if strings.HasPrefix(inner, "vars.") {
		key := strings.TrimPrefix(inner, "vars.")
		// Resolution order: workflow > spec > env (most-specific wins).
		for _, scope := range []string{"workflow", "spec", "env"} {
			if v, ok := e.resolver.Resolve(scope, key); ok {
				return v, nil
			}
		}
		return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("unresolvable vars.%s", key)}
	}

	if strings.HasPrefix(inner, "steps.") {
		parts := strings.Split(inner, ".")
		if len(parts) < 3 {
			return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("malformed steps expression: %s", inner)}
		}
		stepID := parts[1]
		if parts[2] == "extracts" {
			if len(parts) != 4 {
				return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("malformed extracts expression: %s", inner)}
			}
			extractID := parts[3]
			if v, ok := e.ctx.GetExtract(stepID, extractID); ok {
				return v, nil
			}
			return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("extract not found: %s.%s", stepID, extractID)}
		}
		if len(parts) >= 4 && parts[2] == "response" {
			if parts[3] == "status" {
				r, ok := e.ctx.GetStepResult(stepID)
				if !ok || r == nil || r.Timeline == nil {
					return "", &RuntimeError{Code: ErrStepNotFound, Message: fmt.Sprintf("no response available for step %s", stepID)}
				}
				return strconv.Itoa(r.Timeline.StatusCode), nil
			}
			if parts[3] == "body" {
				// Phase 5.2: body JSONPath lookup is a stub for MVP — return empty string.
				// TODO: implement JSONPath extraction in Phase 5.2.
				return "", nil
			}
		}
		return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("unknown steps expression: %s", inner)}
	}

	// unknown pattern
	return "", &RuntimeError{Code: ErrUnresolvableExpression, Message: fmt.Sprintf("unrecognized expression: %s", inner)}
}
