package validation

import (
	"fmt"

	"github.com/tractl/tractl/internal/overlay"
)

var validTargetModes = map[overlay.TargetMode]bool{
	"":                    true,
	overlay.TargetModeOne: true,
	overlay.TargetModeAll: true,
}

// validateTargetMutualExclusion enforces OV-002: exactly one of Path/Match/Source must be set.
func validateTargetMutualExclusion(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		set := 0
		if p.Target.Path != "" {
			set++
		}
		if len(p.Target.Match) > 0 {
			set++
		}
		if p.Target.Source != nil {
			set++
		}
		if set != 1 {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target"),
				Code:    errOverlayTargetModeConflict,
				Message: "exactly one of target.path, target.match, or target.source must be set",
			})
		}
	}
	return errs
}

// validateTargetMode enforces OV-003: Mode must be empty, "one", or "all".
func validateTargetMode(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if !validTargetModes[p.Target.Mode] {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.mode"),
				Code:    errOverlayInvalidTargetMode,
				Message: fmt.Sprintf("invalid target mode %q: must be empty, %q, or %q", p.Target.Mode, overlay.TargetModeOne, overlay.TargetModeAll),
			})
		}
	}
	return errs
}

// validatePathNotEmpty enforces OV-007: path string must be non-empty when path targeting is used.
func validatePathNotEmpty(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		// Fire when all selectors are absent: OV-002 treats an empty path as "not set",
		// so this rule catches the degenerate all-empty case independently.
		if p.Target.Path == "" && p.Target.Match == nil && p.Target.Source == nil {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.path"),
				Code:    errOverlayPathEmpty,
				Message: "target.path must not be empty when path targeting is used",
			})
		}
	}
	return errs
}

// validateSourceType enforces OV-008: Source.Type must be non-empty when source targeting is used.
func validateSourceType(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Target.Source != nil && p.Target.Source.Type == "" {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.source.type"),
				Code:    errOverlaySourceTypeRequired,
				Message: "target.source.type must be non-empty when source targeting is used",
			})
		}
	}
	return errs
}

// validateMatchNotEmpty enforces OV-009: MatchSelector must have at least one entry.
func validateMatchNotEmpty(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		// A non-nil but zero-length map means the user explicitly set match to an
		// empty object — the selector is meaningless and must be rejected.
		if p.Target.Match != nil && len(p.Target.Match) == 0 {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.match"),
				Code:    errOverlayMatchEmpty,
				Message: "target.match must contain at least one selector property",
			})
		}
	}
	return errs
}

// validatePathModeAbsent enforces OV-011: Mode must be empty when path targeting is used.
func validatePathModeAbsent(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Target.Path != "" && p.Target.Mode != "" {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.mode"),
				Code:    errOverlayPathModeInvalid,
				Message: "target.mode must be empty when path targeting is used; path always resolves to exactly one node",
			})
		}
	}
	return errs
}
