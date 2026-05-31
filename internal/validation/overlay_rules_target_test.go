package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/overlay"
)

// ── OV-002 ────────────────────────────────────────────────────────────────────

func TestValidateTargetMutualExclusion(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: path only",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match only",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "login"}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "valid: source only",
			patches: []overlay.Patch{{
				Target: overlay.Target{Source: &overlay.SourceSelector{Type: "openapi"}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: path and match both set",
			patches: []overlay.Patch{{
				Target: overlay.Target{
					Path:  "workflows.main",
					Match: overlay.MatchSelector{"id": "login"},
				},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayTargetModeConflict,
		},
		{
			name: "invalid: all three set",
			patches: []overlay.Patch{{
				Target: overlay.Target{
					Path:   "workflows.main",
					Match:  overlay.MatchSelector{"id": "x"},
					Source: &overlay.SourceSelector{Type: "openapi"},
				},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayTargetModeConflict,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateTargetMutualExclusion(tt.patches)
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

// ── OV-003 ────────────────────────────────────────────────────────────────────

func TestValidateTargetMode(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: empty mode",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: mode=one",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "x"}, Mode: overlay.TargetModeOne},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "valid: mode=all",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "x"}, Mode: overlay.TargetModeAll},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: unrecognised mode value",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "x"}, Mode: "first"},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayInvalidTargetMode,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateTargetMode(tt.patches)
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

// ── OV-007 ────────────────────────────────────────────────────────────────────

func TestValidatePathNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: non-empty path",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match targeting (no path field involved)",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "x"}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: all selectors absent (path intended but empty)",
			patches: []overlay.Patch{{
				Target: overlay.Target{},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayPathEmpty,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validatePathNotEmpty(tt.patches)
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

// ── OV-008 ────────────────────────────────────────────────────────────────────

func TestValidateSourceType(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name: "valid: source with type set",
			patches: []overlay.Patch{{
				Target: overlay.Target{Source: &overlay.SourceSelector{Type: "openapi"}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name:    "valid: no source targeting",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "invalid: source present but type empty",
			patches: []overlay.Patch{{
				Target: overlay.Target{Source: &overlay.SourceSelector{}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlaySourceTypeRequired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateSourceType(tt.patches)
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

// ── OV-009 ────────────────────────────────────────────────────────────────────

func TestValidateMatchNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name: "valid: match with entries",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{"id": "login"}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name:    "valid: nil match (path targeting)",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "invalid: match set to empty map",
			patches: []overlay.Patch{{
				Target: overlay.Target{Match: overlay.MatchSelector{}},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayMatchEmpty,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validateMatchNotEmpty(tt.patches)
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

// ── OV-011 ────────────────────────────────────────────────────────────────────

func TestValidatePathModeAbsent(t *testing.T) {
	tests := []struct {
		name     string
		patches  []overlay.Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: path targeting with no mode",
			patches: []overlay.Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match targeting with mode=all",
			patches: []overlay.Patch{{
				Target: overlay.Target{
					Match: overlay.MatchSelector{"id": "x"},
					Mode:  overlay.TargetModeAll,
				},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: path targeting with mode set",
			patches: []overlay.Patch{{
				Target: overlay.Target{
					Path: "workflows.main",
					Mode: overlay.TargetModeAll,
				},
				Action: overlay.ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayPathModeInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := validatePathModeAbsent(tt.patches)
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
