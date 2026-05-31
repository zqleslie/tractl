// path_test.go tests overlay path resolution, selector matching, and
// mode:all traversal behaviour.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import (
	"encoding/json"
	"reflect"
	"testing"
)

// navPath walks a dot-separated path through a map[string]any, panicking if any segment
// is absent. Used in tests to reach nested values.
func navPath(t *testing.T, m map[string]any, segments ...string) any {
	t.Helper()
	var cur any = m
	for _, s := range segments {
		mm, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("navPath: expected map at segment %q, got %T", s, cur)
		}
		val, ok := mm[s]
		if !ok {
			t.Fatalf("navPath: key %q not found", s)
		}
		cur = val
	}
	return cur
}

// ─── Path targeting ──────────────────────────────────────────────────────────

func TestPathTargeting_ScalarReplace(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"timeout": "PT10S"}

	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"timeout": "PT30S"},
			},
		},
	}
	result := applyOne(t, src, doc)
	vars, _ := result["variables"].(map[string]any)
	if vars["timeout"] != "PT30S" {
		t.Errorf("scalar replace via deepMerge: want 'PT30S', got %v", vars["timeout"])
	}
}

func TestPathTargeting_NestedObjectDeepMerge(t *testing.T) {
	src := specDoc()
	// Add a variables map we can merge into.
	src["variables"] = map[string]any{
		"baseURL": "https://example.com",
	}
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"env": "production"},
			},
		},
	}
	result := applyOne(t, src, doc)
	vars, _ := result["variables"].(map[string]any)
	if vars["baseURL"] != "https://example.com" {
		t.Errorf("existing key lost after deepMerge: vars=%v", vars)
	}
	if vars["env"] != "production" {
		t.Errorf("patch key not applied: vars=%v", vars)
	}
}

// ─── Selector cardinality ─────────────────────────────────────────────────────

func TestModeAllZeroMatchesSucceeds(t *testing.T) {
	src := specDoc()
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Match: MatchSelector{"protocol": "grpc"}, Mode: TargetModeAll},
				Action: ActionDeepMerge,
				Data:   map[string]any{"timeout": "PT30S"},
			},
		},
	}
	// Should succeed with no-op.
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Errorf("mode:all with zero matches should succeed, got: %v", err)
	}
}

func TestModeAllMultipleMatchesAppliesAll(t *testing.T) {
	src := map[string]any{
		"schemaVersion": float64(1),
		"capabilities":  []any{"protocol.http"},
		"workflows": []any{
			map[string]any{
				"id": "wf1",
				"steps": []any{
					map[string]any{
						"id":       "s1",
						"kind":     "request",
						"protocol": "http",
					},
					map[string]any{
						"id":       "s2",
						"kind":     "request",
						"protocol": "http",
					},
				},
			},
		},
	}
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Match: MatchSelector{"protocol": "http"}, Mode: TargetModeAll},
				Action: ActionDeepMerge,
				Data:   map[string]any{"timeout": "PT30S"},
			},
		},
	}
	result, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	// Both matching nodes should have timeout set.
	m := result.(map[string]any)
	wfs, _ := m["workflows"].([]any)
	steps, _ := wfs[0].(map[string]any)["steps"].([]any)
	for i, step := range steps {
		sm, _ := step.(map[string]any)
		if sm["timeout"] != "PT30S" {
			t.Errorf("step[%d] timeout not set; mode:all should have applied to all matching nodes", i)
		}
	}
}

// ─── mode:all determinism ─────────────────────────────────────────────────────

func TestModeAllDeterministic(t *testing.T) {
	// Build a doc with multiple matching nodes.
	src := map[string]any{
		"schemaVersion": float64(1),
		"capabilities":  []any{"protocol.http"},
		"a":             map[string]any{"kind": "x", "val": "1"},
		"b":             map[string]any{"kind": "x", "val": "2"},
		"c":             map[string]any{"kind": "x", "val": "3"},
	}
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Match: MatchSelector{"kind": "x"}, Mode: TargetModeAll},
				Action: ActionDeepMerge,
				Data:   map[string]any{"patched": true},
			},
		},
	}

	// Run Apply twice and compare marshaled output.
	run1, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatal(err)
	}
	run2, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatal(err)
	}

	b1 := mustMarshal(t, run1)
	b2 := mustMarshal(t, run2)

	// JSON marshal of map[string]any is non-deterministic by key order, so compare
	// the decoded structs instead.
	var m1, m2 map[string]any
	_ = json.Unmarshal(b1, &m1)
	_ = json.Unmarshal(b2, &m2)
	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("mode:all produced non-deterministic output:\nrun1=%s\nrun2=%s", b1, b2)
	}
}

// ─── navPath smoke test ───────────────────────────────────────────────────────

func TestNavPath(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"region": "eu-west-1"}
	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"env": "test"}},
		},
	}
	result := applyOne(t, src, doc)
	val := navPath(t, result, "variables", "env")
	if val != "test" {
		t.Errorf("navPath: want 'test', got %v", val)
	}
}

// ─── resolvePath unit tests ───────────────────────────────────────────────────

func TestResolvePath_HappyPath(t *testing.T) {
	doc := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "leaf",
			},
		},
	}
	result, err := resolvePath(doc, "a.b.c")
	if err != nil {
		t.Fatal(err)
	}
	if result.value != "leaf" {
		t.Errorf("want 'leaf', got %v", result.value)
	}
	if result.key != "c" {
		t.Errorf("key: want 'c', got %q", result.key)
	}
}
