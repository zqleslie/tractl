package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/overlay"
)

// ── OV-004 ────────────────────────────────────────────────────────────────────

func TestValidateMergeAction(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: empty action (schema default)",
			patches: []overlay.Patch{{Target: overlay.Target{Path: "x"}, Data: map[string]any{"k": "v"}}},
			wantErr: false,
		},
		{
			name:    "valid: replace",
			patches: []overlay.Patch{{Target: overlay.Target{Path: "x"}, Action: overlay.ActionReplace, Data: map[string]any{"k": "v"}}},
			wantErr: false,
		},
		{
			name:    "valid: remove",
			patches: []overlay.Patch{{Target: overlay.Target{Path: "x"}, Action: overlay.ActionRemove}},
			wantErr: false,
		},
		{
			name: "invalid: unknown action",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "x"},
				Action: "upsert",
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayInvalidAction,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateMergeAction(tt.patches)
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

// ── OV-005 ────────────────────────────────────────────────────────────────────

func TestValidateRemovePatchShape(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: remove with no data",
			patches: []overlay.Patch{{Target: overlay.Target{Path: "x"}, Action: overlay.ActionRemove}},
			wantErr: false,
		},
		{
			name:    "valid: non-remove with data",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "invalid: remove with data present",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "x"},
				Action: overlay.ActionRemove,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayRemovePatchForbidden,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRemovePatchShape(tt.patches)
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

// ── OV-006 ────────────────────────────────────────────────────────────────────

func TestValidateNonRemovePatchShape(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: non-remove with data",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name:    "valid: remove without data",
			patches: []overlay.Patch{{Target: overlay.Target{Path: "x"}, Action: overlay.ActionRemove}},
			wantErr: false,
		},
		{
			name: "invalid: replace with no data",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "x"},
				Action: overlay.ActionReplace,
			}},
			wantErr:  true,
			wantCode: errOverlayPatchRequired,
		},
		{
			name: "invalid: empty action with no data",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "x"},
			}},
			wantErr:  true,
			wantCode: errOverlayPatchRequired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateNonRemovePatchShape(tt.patches)
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

// ── OV-010 ────────────────────────────────────────────────────────────────────

func TestValidateAppendUniqueContract(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name: "valid: appendUnique on assertions",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "workflows.main.steps.login.assertions"},
				Action: overlay.ActionAppendUnique,
				Data:   map[string]any{"id": "check-status"},
			}},
			wantErr: false,
		},
		{
			name: "valid: appendUnique on workflows (top-level)",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "workflows"},
				Action: overlay.ActionAppendUnique,
				Data:   map[string]any{"id": "wf2"},
			}},
			wantErr: false,
		},
		{
			name:    "valid: replace on non-contract field",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: appendUnique with match targeting (path check skipped)",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "x"}},
				Action: overlay.ActionAppendUnique,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: appendUnique on non-contract field",
			patches: []overlay.Patch{{
				Target: overlay.Target{Path: "workflows.main.steps.login.request.headers"},
				Action: overlay.ActionAppendUnique,
				Data:   map[string]any{"X-Extra": "value"},
			}},
			wantErr:  true,
			wantCode: errOverlayAppendUniqueContract,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateAppendUniqueContract(tt.patches)
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
