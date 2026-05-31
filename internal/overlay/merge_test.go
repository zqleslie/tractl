// merge_test.go tests field-level merge strategies: scalar replace, deepMerge,
// array append/appendUnique, identity contracts, and default action selection.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import "testing"

// ─── Merge semantics ─────────────────────────────────────────────────────────

func TestMergeScalarReplace(t *testing.T) {
	existing := "old"
	patched, err := applyMerge(existing, "new", ActionReplace, "timeout", "", 0, "path:timeout")
	if err != nil {
		t.Fatal(err)
	}
	if patched != "new" {
		t.Errorf("scalar replace: want 'new', got %v", patched)
	}
}

func TestMergeDeepMergeObject(t *testing.T) {
	existing := map[string]any{"a": "1", "b": "2"}
	patch := map[string]any{"b": "99", "c": "3"}
	result, err := applyMerge(existing, patch, ActionDeepMerge, "headers", "", 0, "path:headers")
	if err != nil {
		t.Fatal(err)
	}
	m, _ := result.(map[string]any)
	if m["a"] != "1" {
		t.Errorf("existing key 'a' lost: %v", m)
	}
	if m["b"] != "99" {
		t.Errorf("patch key 'b' not applied: %v", m)
	}
	if m["c"] != "3" {
		t.Errorf("new key 'c' not set: %v", m)
	}
}

func TestMergeArrayAppend(t *testing.T) {
	existing := []any{"tag1", "tag2"}
	patch := []any{"tag3"}
	result, err := applyMerge(existing, patch, ActionAppend, "tags", "", 0, "path:tags")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.([]any)
	if len(s) != 3 || s[2] != "tag3" {
		t.Errorf("append failed: %v", s)
	}
}

func TestMergeArrayReplace(t *testing.T) {
	existing := []any{"dep1"}
	patch := []any{"dep2", "dep3"}
	result, err := applyMerge(existing, patch, ActionReplace, "dependsOn", "", 0, "path:dependsOn")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.([]any)
	if len(s) != 2 || s[0] != "dep2" {
		t.Errorf("replace failed: %v", s)
	}
}

func TestMergeAppendUniqueNoCollision(t *testing.T) {
	existing := []any{
		map[string]any{"id": "assert1", "kind": "status"},
	}
	patch := []any{
		map[string]any{"id": "assert2", "kind": "schema"},
	}
	result, err := applyMerge(existing, patch, ActionAppendUnique, "assertions", "", 0, "path:assertions")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.([]any)
	if len(s) != 2 {
		t.Errorf("appendUnique no-collision: want 2 entries, got %d", len(s))
	}
}

func TestMergeAppendUniqueCollisionLastWins(t *testing.T) {
	existing := []any{
		map[string]any{"id": "assert1", "kind": "status", "op": "equals"},
		map[string]any{"id": "assert2", "kind": "schema"},
	}
	patch := []any{
		map[string]any{"id": "assert1", "kind": "status", "op": "notEquals"},
	}
	result, err := applyMerge(existing, patch, ActionAppendUnique, "assertions", "", 0, "path:assertions")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.([]any)
	if len(s) != 2 {
		t.Errorf("appendUnique collision: want 2 entries, got %d", len(s))
	}
	m0, _ := s[0].(map[string]any)
	if m0["op"] != "notEquals" {
		t.Errorf("appendUnique collision: existing not replaced by incoming: %v", m0)
	}
	m1, _ := s[1].(map[string]any)
	if m1["id"] != "assert2" {
		t.Errorf("appendUnique collision: untouched entry lost: %v", m1)
	}
}

// ─── Identity contracts ───────────────────────────────────────────────────────

func TestIdentityContractAssertions(t *testing.T) {
	_, err := applyMerge([]any{}, []any{}, ActionAppendUnique, "assertions", "", 0, "")
	if err != nil {
		t.Errorf("assertions should have identity contract: %v", err)
	}
}

func TestIdentityContractExtracts(t *testing.T) {
	_, err := applyMerge([]any{}, []any{}, ActionAppendUnique, "extracts", "", 0, "")
	if err != nil {
		t.Errorf("extracts should have identity contract: %v", err)
	}
}

// ─── deepMerge unit tests ─────────────────────────────────────────────────────

func TestDeepMergePreservesExistingKeys(t *testing.T) {
	existing := map[string]any{"a": "1", "b": map[string]any{"c": "2"}}
	patch := map[string]any{"b": map[string]any{"d": "3"}, "e": "4"}
	result, _ := deepMerge(existing, patch).(map[string]any)
	if result["a"] != "1" {
		t.Errorf("'a' lost: %v", result)
	}
	b, _ := result["b"].(map[string]any)
	if b["c"] != "2" {
		t.Errorf("'b.c' lost: %v", b)
	}
	if b["d"] != "3" {
		t.Errorf("'b.d' not added: %v", b)
	}
	if result["e"] != "4" {
		t.Errorf("'e' not added: %v", result)
	}
}

// ─── defaultAction unit tests ─────────────────────────────────────────────────

func TestDefaultActionScalar(t *testing.T) {
	action, err := defaultAction("scalar", "timeout", "", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if action != ActionReplace {
		t.Errorf("scalar default: want replace, got %s", action)
	}
}

func TestDefaultActionObject(t *testing.T) {
	action, err := defaultAction(map[string]any{}, "headers", "", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if action != ActionDeepMerge {
		t.Errorf("object default: want deepMerge, got %s", action)
	}
}

func TestDefaultActionAppendArray(t *testing.T) {
	for _, name := range []string{"assertions", "extracts", "tags"} {
		action, err := defaultAction([]any{}, name, "", 0, "")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if action != ActionAppend {
			t.Errorf("%s: want append, got %s", name, action)
		}
	}
}

func TestDefaultActionReplaceArray(t *testing.T) {
	for _, name := range []string{"dependsOn", "capabilities", "invariants"} {
		action, err := defaultAction([]any{}, name, "", 0, "")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if action != ActionReplace {
			t.Errorf("%s: want replace, got %s", name, action)
		}
	}
}
