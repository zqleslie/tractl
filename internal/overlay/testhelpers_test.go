// testhelpers_test.go provides shared test helper functions for the
// internal/overlay package test suite.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import (
	"encoding/json"
	"errors"
	"testing"
)

// specDoc returns a minimal JSON spec document as map[string]any. It mirrors
// the shape of *spec.TraCtlSpec after JSON marshaling.
func specDoc(extra ...map[string]any) map[string]any {
	base := map[string]any{
		"schemaVersion": float64(1),
		"capabilities":  []any{"protocol.http"},
		"workflows": []any{
			map[string]any{
				"id":   "main",
				"name": "Main Workflow",
				"steps": []any{
					map[string]any{
						"id":   "login",
						"kind": "request",
						"request": map[string]any{
							"protocol": "http",
							"target":   "https://example.com/login",
							"headers": map[string]any{
								"Accept": "application/json",
							},
						},
						"assertions": []any{},
						"extracts":   []any{},
					},
				},
			},
		},
	}
	for _, m := range extra {
		for k, v := range m {
			base[k] = v
		}
	}
	return base
}

func eng() *Engine { return NewEngine() }

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustMarshal: %v", err)
	}
	return b
}

func mustUnmarshal(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("mustUnmarshal: %v", err)
	}
	return m
}

// applyOne applies a single overlay and returns the result as map[string]any.
func applyOne(t *testing.T, src map[string]any, doc *OverlayDocument) map[string]any {
	t.Helper()
	result, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Apply returned unexpected type %T", result)
	}
	return m
}

func engineErrCode(err error) ErrorCode {
	var ee *EngineError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return ""
}
