package overlay

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

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

// ─── Source must not be mutated ───────────────────────────────────────────────

func TestSourceNotMutated(t *testing.T) {
	src := specDoc()
	original := mustMarshal(t, src)

	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "schemaVersion"},
				Action: ActionReplace,
				Data:   map[string]any{"_value": float64(99)},
			},
		},
	}
	// Use a simple scalar replace.
	doc2 := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "capabilities"},
				Action: ActionReplace,
				Data:   map[string]any{},
			},
		},
	}
	_, _ = eng().Apply(src, []*OverlayDocument{doc, doc2})

	after := mustMarshal(t, src)
	if string(original) != string(after) {
		t.Errorf("source was mutated: before=%s after=%s", original, after)
	}
}

// ─── Clone failure ────────────────────────────────────────────────────────────

func TestCloneFailure(t *testing.T) {
	// A channel cannot be marshaled to JSON, so clone will fail.
	badSrc := make(chan int)
	_, err := eng().Apply(badSrc, nil)
	if err == nil {
		t.Fatal("expected error for unmarshalable source")
	}
	if code := engineErrCode(err); code != ErrCloneFailed {
		t.Errorf("expected %s, got %s", ErrCloneFailed, code)
	}
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

func TestPathTargeting_NotFound(t *testing.T) {
	src := specDoc()
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "nonexistent.field"},
				Action: ActionReplace,
				Data:   map[string]any{"x": "y"},
			},
		},
	}
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
	if code := engineErrCode(err); code != ErrPathNotFound {
		t.Errorf("expected %s, got %s", ErrPathNotFound, code)
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

// ─── Patch order preservation ─────────────────────────────────────────────────

func TestPatchOrderPreservation(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "first"}

	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"x": "second"},
			},
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"x": "third"},
			},
		},
	}
	result := applyOne(t, src, doc)
	vars, _ := result["variables"].(map[string]any)
	if vars["x"] != "third" {
		t.Errorf("patch order not preserved: want 'third', got %v", vars["x"])
	}
}

// ─── Overlay stacking / last-wins ────────────────────────────────────────────

func TestOverlayStackingLastWins(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"env": "dev"}

	o1 := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "overlay1"},
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"env": "staging"}},
		},
	}
	o2 := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "overlay2"},
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"env": "production"}},
		},
	}

	result, err := eng().Apply(src, []*OverlayDocument{o1, o2})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	m := result.(map[string]any)
	vars, _ := m["variables"].(map[string]any)
	if vars["env"] != "production" {
		t.Errorf("last-wins violated: want 'production', got %v", vars["env"])
	}
}

func TestThreeOverlayStackLastWins(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "a"}

	overlays := []*OverlayDocument{
		{Patches: []Patch{{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"x": "b"}}}},
		{Patches: []Patch{{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"x": "c"}}}},
		{Patches: []Patch{{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"x": "d"}}}},
	}
	result, err := eng().Apply(src, overlays)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	vars, _ := result.(map[string]any)["variables"].(map[string]any)
	if vars["x"] != "d" {
		t.Errorf("three-overlay last-wins violated: want 'd', got %v", vars["x"])
	}
}

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

func TestMergeAppendUniqueNoIdentityContract(t *testing.T) {
	_, err := applyMerge([]any{}, []any{}, ActionAppendUnique, "retryOn", "", 0, "path:retryOn")
	if err == nil {
		t.Fatal("expected error for unknown identity contract")
	}
	if code := engineErrCode(err); code != ErrNoIdentityContract {
		t.Errorf("want %s, got %s", ErrNoIdentityContract, code)
	}
}

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

// ─── Removal ─────────────────────────────────────────────────────────────────

func TestRemovalScalar(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"toRemove": "bye", "keep": "hi"}

	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables.toRemove"}, Action: ActionRemove},
		},
	}
	result := applyOne(t, src, doc)
	vars, _ := result["variables"].(map[string]any)
	if _, ok := vars["toRemove"]; ok {
		t.Errorf("field 'toRemove' still present after remove")
	}
	if vars["keep"] != "hi" {
		t.Errorf("unrelated field 'keep' was also removed")
	}
}

func TestRemovalObjectField(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"a": "1"}

	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionRemove},
		},
	}
	result := applyOne(t, src, doc)
	if _, ok := result["variables"]; ok {
		t.Errorf("'variables' still present after remove")
	}
}

func TestRemovalWithPatchBodyErrors(t *testing.T) {
	src := specDoc()
	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "capabilities"}, Action: ActionRemove, Data: map[string]any{"bad": "field"}},
		},
	}
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err == nil {
		t.Fatal("expected error: remove with patch body")
	}
	if code := engineErrCode(err); code != ErrRemoveWithPatch {
		t.Errorf("want %s, got %s", ErrRemoveWithPatch, code)
	}
}

// ─── Selector cardinality ─────────────────────────────────────────────────────

func TestModeOneZeroMatchesErrors(t *testing.T) {
	src := specDoc()
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{
					Match: MatchSelector{"protocol": "grpc"},
					Mode:  TargetModeOne,
				},
				Action: ActionDeepMerge,
				Data:   map[string]any{"x": "y"},
			},
		},
	}
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err == nil {
		t.Fatal("expected ErrModeOneNoMatch")
	}
	if code := engineErrCode(err); code != ErrModeOneNoMatch {
		t.Errorf("want %s, got %s", ErrModeOneNoMatch, code)
	}
}

func TestModeOneMultipleMatchesErrors(t *testing.T) {
	// Build a doc with two nodes both matching protocol=http.
	src := map[string]any{
		"schemaVersion": float64(1),
		"capabilities":  []any{"protocol.http"},
		"workflows": []any{
			map[string]any{
				"id": "wf1",
				"steps": []any{
					map[string]any{
						"id":      "s1",
						"kind":    "request",
						"request": map[string]any{"protocol": "http", "target": "https://a.com"},
					},
					map[string]any{
						"id":      "s2",
						"kind":    "request",
						"request": map[string]any{"protocol": "http", "target": "https://b.com"},
					},
				},
			},
		},
	}
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Match: MatchSelector{"protocol": "http"}, Mode: TargetModeOne},
				Action: ActionDeepMerge,
				Data:   map[string]any{"timeout": "PT30S"},
			},
		},
	}
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err == nil {
		t.Fatal("expected ErrModeOneMultipleMatches")
	}
	if code := engineErrCode(err); code != ErrModeOneMultipleMatches {
		t.Errorf("want %s, got %s", ErrModeOneMultipleMatches, code)
	}
}

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

// ─── Explicit action legality ─────────────────────────────────────────────────

func TestUnrecognisedActionReturnsError(t *testing.T) {
	src := specDoc()
	doc := &OverlayDocument{
		Patches: []Patch{
			{
				Target: Target{Path: "capabilities"},
				Action: MergeAction("invalidAction"),
				Data:   map[string]any{"x": "y"},
			},
		},
	}
	_, err := eng().Apply(src, []*OverlayDocument{doc})
	if err == nil {
		t.Fatal("expected ErrInvalidAction for unrecognised action")
	}
	if code := engineErrCode(err); code != ErrInvalidAction {
		t.Errorf("want %s, got %s", ErrInvalidAction, code)
	}
}

// ─── Provenance ───────────────────────────────────────────────────────────────

func TestProvenancePopulated(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "1"}

	doc := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "my-overlay"},
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"y": "2"}},
		},
	}
	result, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	m := result.(map[string]any)
	meta, _ := m["metadata"].(map[string]any)
	if meta == nil {
		t.Fatal("metadata missing from candidate")
	}

	overlayRefs, _ := meta["overlayRefs"].([]any)
	if len(overlayRefs) == 0 {
		t.Error("overlayRefs not populated")
	}

	provenance, _ := meta["provenance"].(map[string]any)
	if provenance == nil {
		t.Fatal("provenance map missing")
	}
	entry, _ := provenance["variables"].(map[string]any)
	if entry == nil {
		t.Fatalf("provenance entry for 'variables' missing; provenance=%v", provenance)
	}
	if entry["patchIndex"] != 0 {
		t.Errorf("provenance patchIndex: want 0, got %v", entry["patchIndex"])
	}
}

func TestProvenanceNotInExecutionFields(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "1"}

	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"y": "2"}},
		},
	}
	result, err := eng().Apply(src, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	m := result.(map[string]any)

	// Provenance must only appear inside metadata — not in workflows, steps, or any
	// execution field. "workflows" is a []any slice; check it has no provenance field.
	wfs, _ := m["workflows"].([]any)
	for i, wf := range wfs {
		wfMap, _ := wf.(map[string]any)
		if _, ok := wfMap["provenance"]; ok {
			t.Errorf("provenance leaked into workflows[%d]", i)
		}
	}
	vars, _ := m["variables"].(map[string]any)
	if _, ok := vars["provenance"]; ok {
		t.Error("provenance leaked into variables")
	}
	// "capabilities" is a []any; no provenance should be there either.
	if _, ok := m["capabilities"].(map[string]any); ok {
		t.Error("capabilities unexpectedly became a map")
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

// ─── Parse ────────────────────────────────────────────────────────────────────

func TestParseYAML(t *testing.T) {
	src := []byte(`
metadata:
  name: test-overlay
patches:
  - target:
      path: variables
    action: deepMerge
    data:
      env: staging
`)
	doc, err := Parse(src, "yaml")
	if err != nil {
		t.Fatalf("Parse yaml: %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Name != "test-overlay" {
		t.Errorf("metadata.name: got %v", doc.Metadata)
	}
	if len(doc.Patches) != 1 {
		t.Fatalf("patches count: want 1, got %d", len(doc.Patches))
	}
	if doc.Patches[0].Target.Path != "variables" {
		t.Errorf("patch target.path: got %q", doc.Patches[0].Target.Path)
	}
}

func TestParseJSON(t *testing.T) {
	src := []byte(`{
  "metadata": {"name": "json-overlay"},
  "patches": [
    {
      "target": {"path": "variables"},
      "action": "deepMerge",
      "data": {"env": "production"}
    }
  ]
}`)
	doc, err := Parse(src, "json")
	if err != nil {
		t.Fatalf("Parse json: %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Name != "json-overlay" {
		t.Errorf("metadata.name: got %v", doc.Metadata)
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	_, err := Parse([]byte("x"), "toml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

// ─── Cross-format parity ──────────────────────────────────────────────────────

// TestCrossFormatParity verifies that semantically equivalent overlays in different
// formats produce identical results when applied to the same source.
// Reference: overlay spec §6, §20, ADR-002 §3a
func TestCrossFormatParity(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"env": "dev"}

	yamlOverlay := []byte(`
metadata:
  name: parity-overlay
patches:
  - target:
      path: variables
    action: deepMerge
    data:
      region: us-east-1
`)
	jsonOverlay := []byte(`{
  "metadata": {"name": "parity-overlay"},
  "patches": [
    {
      "target": {"path": "variables"},
      "action": "deepMerge",
      "data": {"region": "us-east-1"}
    }
  ]
}`)

	docYAML, err := Parse(yamlOverlay, "yaml")
	if err != nil {
		t.Fatalf("Parse yaml: %v", err)
	}
	docJSON, err := Parse(jsonOverlay, "json")
	if err != nil {
		t.Fatalf("Parse json: %v", err)
	}

	resYAML, err := eng().Apply(src, []*OverlayDocument{docYAML})
	if err != nil {
		t.Fatalf("Apply yaml overlay: %v", err)
	}
	resJSON, err := eng().Apply(src, []*OverlayDocument{docJSON})
	if err != nil {
		t.Fatalf("Apply json overlay: %v", err)
	}

	// Compare variables (execution fields only; provenance may differ in timestamps).
	mY, _ := resYAML.(map[string]any)
	mJ, _ := resJSON.(map[string]any)

	varsY, _ := mY["variables"].(map[string]any)
	varsJ, _ := mJ["variables"].(map[string]any)

	if !reflect.DeepEqual(varsY, varsJ) {
		t.Errorf("cross-format parity violated:\nyaml result vars: %v\njson result vars: %v", varsY, varsJ)
	}
}

// ─── ApplyToJSON ─────────────────────────────────────────────────────────────

func TestApplyToJSON_RoundTrip(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "1"}
	srcJSON := mustMarshal(t, src)

	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"y": "2"}},
		},
	}

	out, err := eng().ApplyToJSON(srcJSON, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("ApplyToJSON: %v", err)
	}

	result := mustUnmarshal(t, out)
	vars, _ := result["variables"].(map[string]any)
	if vars["x"] != "1" || vars["y"] != "2" {
		t.Errorf("ApplyToJSON round-trip vars: %v", vars)
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

func TestResolvePath_NotFound(t *testing.T) {
	doc := map[string]any{"a": map[string]any{"b": "v"}}
	_, err := resolvePath(doc, "a.x.y")
	if err == nil {
		t.Fatal("expected error")
	}
	if engineErrCode(err) != ErrPathNotFound {
		t.Errorf("want ErrPathNotFound, got %v", err)
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

func TestDefaultActionUnknownArrayErrors(t *testing.T) {
	_, err := defaultAction([]any{}, "unknownArray", "", 0, "path:unknownArray")
	if err == nil {
		t.Fatal("expected error for unknown array with no schema default")
	}
	if engineErrCode(err) != ErrInvalidAction {
		t.Errorf("want ErrInvalidAction, got %v", err)
	}
}
