// validator_rules_target_test.go tests the target validation rules defined in
// validator_rules_target.go: OV-002 (mutual exclusion), OV-003 (target mode),
// OV-007 (empty path), OV-008 (source type), OV-009 (empty match), OV-011
// (mode with path targeting).
package overlay

import "testing"

// ── OV-002 ────────────────────────────────────────────────────────────────────

func TestValidateTargetMutualExclusion(t *testing.T) {
	tests := []struct {
		name     string
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: path only",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match only",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "login"}},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "valid: source only",
			patches: []Patch{{
				Target: Target{Source: &SourceSelector{Type: "openapi"}},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: path and match both set",
			patches: []Patch{{
				Target: Target{
					Path:  "workflows.main",
					Match: MatchSelector{"id": "login"},
				},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr:  true,
			wantCode: errOverlayTargetModeConflict,
		},
		{
			name: "invalid: all three set",
			patches: []Patch{{
				Target: Target{
					Path:   "workflows.main",
					Match:  MatchSelector{"id": "x"},
					Source: &SourceSelector{Type: "openapi"},
				},
				Action: ActionReplace,
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
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: empty mode",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: mode=one",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "x"}, Mode: TargetModeOne},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "valid: mode=all",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "x"}, Mode: TargetModeAll},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: unrecognised mode value",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "x"}, Mode: "first"},
				Action: ActionReplace,
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
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: non-empty path",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match targeting (no path field involved)",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "x"}},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: all selectors absent (path intended but empty)",
			patches: []Patch{{
				Target: Target{},
				Action: ActionReplace,
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
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name: "valid: source with type set",
			patches: []Patch{{
				Target: Target{Source: &SourceSelector{Type: "openapi"}},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name:    "valid: no source targeting",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "invalid: source present but type empty",
			patches: []Patch{{
				Target: Target{Source: &SourceSelector{}},
				Action: ActionReplace,
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
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name: "valid: match with entries",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{"id": "login"}},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name:    "valid: nil match (path targeting)",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "invalid: match set to empty map",
			patches: []Patch{{
				Target: Target{Match: MatchSelector{}},
				Action: ActionReplace,
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
		patches  []Patch
		wantErr  bool
		wantCode ErrorCode
	}{
		{
			name:    "valid: path targeting with no mode",
			patches: []Patch{validPathPatch()},
			wantErr: false,
		},
		{
			name: "valid: match targeting with mode=all",
			patches: []Patch{{
				Target: Target{
					Match: MatchSelector{"id": "x"},
					Mode:  TargetModeAll,
				},
				Action: ActionReplace,
				Data:   map[string]any{"k": "v"},
			}},
			wantErr: false,
		},
		{
			name: "invalid: path targeting with mode set",
			patches: []Patch{{
				Target: Target{
					Path: "workflows.main",
					Mode: TargetModeAll,
				},
				Action: ActionReplace,
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
