// validator_rules_patch.go defines OverlayValidator patch validation rules:
// OV-004 (merge action legality), OV-005 (remove must have no data body),
// OV-006 (non-remove must have non-empty data), and OV-010 (appendUnique
// identity contract enforcement).
package overlay

import (
	"fmt"
	"strings"
)

var validMergeActions = map[MergeAction]bool{
	"":                 true,
	ActionReplace:      true,
	ActionDeepMerge:    true,
	ActionAppend:       true,
	ActionAppendUnique: true,
	ActionRemove:       true,
}

// validateMergeAction enforces OV-004: Action must be one of the five valid values or empty.
func validateMergeAction(patches []Patch) []ValidationError {
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
func validateRemovePatchShape(patches []Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action == ActionRemove && len(p.Data) > 0 {
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
func validateNonRemovePatchShape(patches []Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action != ActionRemove && len(p.Data) == 0 {
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
func validateAppendUniqueContract(patches []Patch) []ValidationError {
	var errs []ValidationError
	for i, p := range patches {
		if p.Action != ActionAppendUnique || p.Target.Path == "" {
			continue
		}
		segment := p.Target.Path
		if idx := strings.LastIndex(p.Target.Path, "."); idx >= 0 {
			segment = p.Target.Path[idx+1:]
		}
		if !IdentityContractArrays[segment] {
			errs = append(errs, ValidationError{
				Field:   patchField(i, "target.path"),
				Code:    errOverlayAppendUniqueContract,
				Message: fmt.Sprintf("appendUnique is only permitted on identity-contract arrays; %q is not in the allowlist", segment),
			})
		}
	}
	return errs
}
