// validator.go defines OverlayValidator, which validates an overlay document against the overlay schema rules.
//
// ValidationError and ErrorCode are defined here (not imported from internal/validation) so that
// internal/overlay remains import-cycle–free: internal/validation must not import internal/overlay,
// and internal/overlay must not import internal/validation.
//
// COORDINATION(Stream C): if engine.go or any caller currently imports validation.OverlayValidator,
// update that import to overlay.OverlayValidator after this PR lands.
package overlay

import "fmt"

// ValidationError describes a single overlay validation failure.
// Field uses dot-bracket notation matching the overlay document structure.
// ErrorCode is shared with the engine layer (defined in errors.go).
type ValidationError struct { //nolint:revive
	Field   string
	Code    ErrorCode
	Message string
}

const (
	errOverlayPatchesRequired      ErrorCode = "OV-001"
	errOverlayTargetModeConflict   ErrorCode = "OV-002"
	errOverlayInvalidTargetMode    ErrorCode = "OV-003"
	errOverlayInvalidAction        ErrorCode = "OV-004"
	errOverlayRemovePatchForbidden ErrorCode = "OV-005"
	errOverlayPatchRequired        ErrorCode = "OV-006"
	errOverlayPathEmpty            ErrorCode = "OV-007"
	errOverlaySourceTypeRequired   ErrorCode = "OV-008"
	errOverlayMatchEmpty           ErrorCode = "OV-009"
	errOverlayAppendUniqueContract ErrorCode = "OV-010"
	errOverlayPathModeInvalid      ErrorCode = "OV-011"
)

// OverlayValidator validates OverlayDocument structure against overlay spec §8–§11, §15.1, §20.
type OverlayValidator struct{}

// NewOverlayValidator returns a ready-to-use OverlayValidator.
func NewOverlayValidator() *OverlayValidator { return &OverlayValidator{} }

// Validate runs all overlay validation rules and returns every error found.
// Returns a single OV-001 error immediately when patches is nil or empty.
func (v *OverlayValidator) Validate(doc *OverlayDocument) []ValidationError {
	errs := validatePatchesPresent(doc.Patches)
	if len(errs) > 0 {
		return errs
	}

	errs = append(errs, validateTargetMutualExclusion(doc.Patches)...)
	errs = append(errs, validateTargetMode(doc.Patches)...)
	errs = append(errs, validateMergeAction(doc.Patches)...)
	errs = append(errs, validateRemovePatchShape(doc.Patches)...)
	errs = append(errs, validateNonRemovePatchShape(doc.Patches)...)
	errs = append(errs, validatePathNotEmpty(doc.Patches)...)
	errs = append(errs, validateSourceType(doc.Patches)...)
	errs = append(errs, validateMatchNotEmpty(doc.Patches)...)
	errs = append(errs, validateAppendUniqueContract(doc.Patches)...)
	errs = append(errs, validatePathModeAbsent(doc.Patches)...)
	return errs
}

// patchField returns the canonical dot-bracket field path for a patch sub-field.
func patchField(i int, field string) string {
	return fmt.Sprintf("patches[%d].%s", i, field)
}

func validatePatchesPresent(patches []Patch) []ValidationError {
	if len(patches) == 0 {
		return []ValidationError{{
			Field:   "patches",
			Code:    errOverlayPatchesRequired,
			Message: "overlay document must contain at least one patch",
		}}
	}
	return nil
}
