package validation

import (
	"fmt"
	"strings"

	"github.com/tractl/tractl/internal/overlay"
)

var validMergeActions = map[overlay.MergeAction]bool{
	"":                         true,
	overlay.ActionReplace:      true,
	overlay.ActionDeepMerge:    true,
	overlay.ActionAppend:       true,
	overlay.ActionAppendUnique: true,
	overlay.ActionRemove:       true,
}

// validateMergeAction enforces OV-004: Action must be one of the five valid values or empty.
func validateMergeAction(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if !validMergeActions[p.Action] {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "action"),
				Code:    errOverlayInvalidAction,
				Message: fmt.Sprintf("invalid action %q: must be one of replace, deepMerge, append, appendUnique, remove", p.Action),
			})
		}
	}
	return errs
}

// validateRemovePatchShape enforces OV-005: action=remove must not carry a Data field.
func validateRemovePatchShape(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action == overlay.ActionRemove && len(p.Data) > 0 {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "data"),
				Code:    errOverlayRemovePatchForbidden,
				Message: "action=remove must not include a data field",
			})
		}
	}
	return errs
}

// validateNonRemovePatchShape enforces OV-006: non-remove patches must carry a non-empty Data field.
func validateNonRemovePatchShape(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action != overlay.ActionRemove && len(p.Data) == 0 {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "data"),
				Code:    errOverlayPatchRequired,
				Message: "non-remove patches must include a non-empty data field",
			})
		}
	}
	return errs
}

// validateAppendUniqueContract enforces OV-010: appendUnique is only legal on
// identity-contract arrays. Only checked for path targeting; match/source targets
// are not subject to terminal-segment resolution here (Phase 3+ concern).
func validateAppendUniqueContract(patches []overlay.Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action != overlay.ActionAppendUnique || p.Target.Path == "" {
			continue
		}
		segment := p.Target.Path
		if idx := strings.LastIndex(p.Target.Path, "."); idx >= 0 {
			segment = p.Target.Path[idx+1:]
		}
		if !overlay.IdentityContractArrays[segment] {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.path"),
				Code:    errOverlayAppendUniqueContract,
				Message: fmt.Sprintf("appendUnique is only permitted on identity-contract arrays; %q is not in the allowlist", segment),
			})
		}
	}
	return errs
}
