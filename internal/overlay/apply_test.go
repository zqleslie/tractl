// apply_test.go tests overlay application behaviour: source immutability,
// patch ordering, overlay stacking (last-wins), and field removal.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import "testing"

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
