// errors_test.go tests error conditions in the overlay engine: invalid inputs,
// missing targets, unrecognised actions, selector cardinality violations, and
// clone failures.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import "testing"

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

// ─── Path targeting errors ────────────────────────────────────────────────────

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

// ─── Merge errors ─────────────────────────────────────────────────────────────

func TestMergeAppendUniqueNoIdentityContract(t *testing.T) {
	_, err := applyMerge([]any{}, []any{}, ActionAppendUnique, "retryOn", "", 0, "path:retryOn")
	if err == nil {
		t.Fatal("expected error for unknown identity contract")
	}
	if code := engineErrCode(err); code != ErrNoIdentityContract {
		t.Errorf("want %s, got %s", ErrNoIdentityContract, code)
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

// ─── Selector cardinality errors ──────────────────────────────────────────────

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

// ─── Explicit action errors ───────────────────────────────────────────────────

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

// ─── resolvePath errors ───────────────────────────────────────────────────────

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

// ─── defaultAction errors ─────────────────────────────────────────────────────

func TestDefaultActionUnknownArrayErrors(t *testing.T) {
	_, err := defaultAction([]any{}, "unknownArray", "", 0, "path:unknownArray")
	if err == nil {
		t.Fatal("expected error for unknown array with no schema default")
	}
	if engineErrCode(err) != ErrInvalidAction {
		t.Errorf("want ErrInvalidAction, got %v", err)
	}
}
