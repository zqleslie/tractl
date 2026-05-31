package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/overlay"
)

// makeDoc returns a minimal valid OverlayDocument with the given patches.
func makeDoc(patches ...overlay.Patch) *overlay.OverlayDocument {
	return &overlay.OverlayDocument{Patches: patches}
}

// validPathPatch returns a patch that passes all rules on its own.
func validPathPatch() overlay.Patch {
	return overlay.Patch{
		Target: overlay.Target{Path: "workflows.main.steps"},
		Action: overlay.ActionReplace,
		Data:   map[string]any{"id": "login"},
	}
}

// ── OV-001 ────────────────────────────────────────────────────────────────────

func TestValidatePatchesPresent(t *testing.T) {
	v := NewOverlayValidator()
	tests := []struct {
		name     string
		doc      *overlay.OverlayDocument
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
			doc:      &overlay.OverlayDocument{},
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
		overlay.Patch{
			Target: overlay.Target{
				Path:  "workflows.main",
				Match: overlay.MatchSelector{"id": "x"},
			},
			Action: overlay.ActionReplace,
			Data:   map[string]any{"k": "v"},
		},
		// OV-004: unknown action
		overlay.Patch{
			Target: overlay.Target{Path: "workflows.other"},
			Action: "upsert",
			Data:   map[string]any{"k": "v"},
		},
		// OV-006: non-remove patch with no data
		overlay.Patch{
			Target: overlay.Target{Path: "workflows.third"},
			Action: overlay.ActionReplace,
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
		overlay.Patch{ // patches[1] — OV-007: all selectors absent
			Target: overlay.Target{},
			Action: overlay.ActionReplace,
			Data:   map[string]any{"k": "v"},
		},
	)

	errs := v.Validate(doc)
	want := "patches[1].target.path"
	if !hasErrorOnField(errs, errOverlayPathEmpty, want) {
		t.Errorf("expected error %q on field %q; got: %v", errOverlayPathEmpty, want, errs)
	}
}
