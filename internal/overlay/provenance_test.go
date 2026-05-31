// provenance_test.go tests overlay provenance tracking — that the engine
// records which overlay applied which change and confines provenance data
// to the metadata section only.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import "testing"

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
