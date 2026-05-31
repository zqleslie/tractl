// validator_test.go tests OverlayValidator: OV-001 (patches required) and
// integration tests verifying that all errors are collected independently
// and that field paths use the correct patch index notation.
package overlay

import "testing"

// ── test helpers ──────────────────────────────────────────────────────────────

func makeDoc(patches ...Patch) *OverlayDocument {
	return &OverlayDocument{Patches: patches}
}

func validPathPatch() Patch {
	return Patch{
		Target: Target{Path: "workflows.main.steps"},
		Action: ActionReplace,
		Data:   map[string]any{"id": "login"},
	}
}

func hasError(errs []ValidationError, code ErrorCode) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

func hasErrorOnField(errs []ValidationError, code ErrorCode, field string) bool {
	for _, e := range errs {
		if e.Code == code && e.Field == field {
			return true
		}
	}
	return false
}

// ── OV-001 ────────────────────────────────────────────────────────────────────

func TestValidatePatchesPresent(t *testing.T) {
	v := NewOverlayValidator()
	tests := []struct {
		name     string
		doc      *OverlayDocument
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: one patch present",
			doc:     makeDoc(validPathPatch()),
			wantErr: false,
		},
		{
			name:     "invalid: nil patches",
			doc:      &OverlayDocument{},
			wantErr:  true,
			wantCode: errOverlayPatchesRequired,
		},
		{
			name:     "invalid: empty patches slice",
			doc:      makeDoc(),
			wantErr:  true,
			wantCode: errOverlayPatchesRequired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(tt.doc)
			if !tt.wantErr {
				if len(errs) != 0 {
					t.Errorf("expected no errors, got %v", errs)
				}
				return
			}
			if !hasError(errs, tt.wantCode) {
				t.Errorf("expected error %q, got %v", tt.wantCode, errs)
			}
		})
	}
}

// ── Integration ───────────────────────────────────────────────────────────────

// TestOverlayValidate_AllErrorsCollected confirms Validate never stops early:
// a document with three independent violations must report all three error codes.
func TestOverlayValidate_AllErrorsCollected(t *testing.T) {
	v := NewOverlayValidator()
	doc := makeDoc(
		// OV-002: two selectors set simultaneously
		Patch{
			Target: Target{
				Path:  "workflows.main",
				Match: MatchSelector{"id": "x"},
			},
			Action: ActionReplace,
			Data:   map[string]any{"k": "v"},
		},
		// OV-004: unknown action
		Patch{
			Target: Target{Path: "workflows.other"},
			Action: "upsert",
			Data:   map[string]any{"k": "v"},
		},
		// OV-006: non-remove patch with no data
		Patch{
			Target: Target{Path: "workflows.third"},
			Action: ActionReplace,
		},
	)

	errs := v.Validate(doc)
	for _, code := range []ErrorCode{errOverlayTargetModeConflict, errOverlayInvalidAction, errOverlayPatchRequired} {
		if !hasError(errs, code) {
			t.Errorf("expected error %q to be present; got: %v", code, errs)
		}
	}
}

// TestOverlayValidate_PatchFieldPaths verifies the error Field uses the correct
// patch index when the violation is on the second patch.
func TestOverlayValidate_PatchFieldPaths(t *testing.T) {
	v := NewOverlayValidator()
	doc := makeDoc(
		validPathPatch(), // patches[0] — valid
		Patch{ // patches[1] — OV-007: all selectors absent
			Target: Target{},
			Action: ActionReplace,
			Data:   map[string]any{"k": "v"},
		},
	)

	errs := v.Validate(doc)
	want := "patches[1].target.path"
	if !hasErrorOnField(errs, errOverlayPathEmpty, want) {
		t.Errorf("expected error %q on field %q; got: %v", errOverlayPathEmpty, want, errs)
	}
}
