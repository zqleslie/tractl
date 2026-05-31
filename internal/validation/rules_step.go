// rules_step.go defines validation rules for individual steps: step identifier
// format and uniqueness (Rules 7–8), step kind legality (Rule 9), step body
// and kind consistency (Rules 10–12), and assertion/extract field validation.
package validation

import (
	"fmt"
	"strings"

	"github.com/tractl/tractl/internal/spec"
)

// validateSteps enforces Rules 7–12 for all steps in one workflow.
func validateSteps(wf spec.Workflow, wi int) []ValidationError {
	var errs []ValidationError
	wField := fmt.Sprintf("workflows[%d]", wi)
	seenIDs := make(map[string]int) // id → first index; Rule 8

	for si, st := range wf.Steps {
		sField := fmt.Sprintf("%s.steps[%d]", wField, si)

		// Rule 7 — step id format.
		errs = append(errs, validateIdentifier(st.ID, sField+".id")...)

		// Rule 8 — step id uniqueness (workflow-local).
		if st.ID != "" {
			if firstIdx, seen := seenIDs[st.ID]; seen {
				errs = append(errs, ValidationError{
					Field:   sField + ".id",
					Code:    DuplicateStepID,
					Message: fmt.Sprintf("duplicate step id %q in workflow %q (first seen at steps[%d])", st.ID, wf.ID, firstIdx),
				})
			} else {
				seenIDs[st.ID] = si
			}
		}

		// Rule 9 — step kind.
		errs = append(errs, validateStepKind(st, sField)...)

		// Rule 10 — step body/kind consistency.
		errs = append(errs, validateStepBody(st, sField)...)
	}

	// Rules 11 & 12 — dependsOn resolution and cycle detection.
	stepIDs := buildStepIDSet(wf.Steps)
	errs = append(errs, validateDependsOn(wf, wi, stepIDs)...)
	errs = append(errs, validateCycles(wf, wi, stepIDs)...)

	return errs
}

// buildStepIDSet returns a set of valid (non-empty) step IDs in the workflow.
func buildStepIDSet(steps []spec.Step) map[string]bool {
	ids := make(map[string]bool, len(steps))
	for _, st := range steps {
		if st.ID != "" {
			ids[st.ID] = true
		}
	}
	return ids
}

// validateIdentifier validates a user-defined identifier field.
// Reference: tractl_spec.md §4.1
func validateIdentifier(id, field string) []ValidationError {
	if id == "" {
		return []ValidationError{{
			Field:   field,
			Code:    MissingRequiredField,
			Message: field + " is required and must not be empty",
		}}
	}
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(id, prefix) {
			return []ValidationError{{
				Field:   field,
				Code:    ReservedIdentifier,
				Message: fmt.Sprintf("%s uses reserved prefix %q", field, prefix),
			}}
		}
	}
	if !identifierRe.MatchString(id) {
		return []ValidationError{{
			Field:   field,
			Code:    InvalidIdentifier,
			Message: fmt.Sprintf("%s %q does not match required pattern ^[a-zA-Z][a-zA-Z0-9_.-]*$", field, id),
		}}
	}
	return nil
}

// validateStepKind enforces Rule 9. Reference: tractl_spec.md §7.2
func validateStepKind(st spec.Step, sField string) []ValidationError {
	if st.Kind == "" {
		return []ValidationError{{
			Field:   sField + ".kind",
			Code:    MissingRequiredField,
			Message: sField + ".kind is required",
		}}
	}
	if !allowedStepKinds[st.Kind] {
		return []ValidationError{{
			Field:   sField + ".kind",
			Code:    InvalidStepKind,
			Message: fmt.Sprintf("%s.kind %q is not a valid step kind; allowed: request, script, extensionCall, composite", sField, st.Kind),
		}}
	}
	return nil
}

// stepBodyPresent maps each allowed kind to a function that checks its body pointer.
var stepBodyPresent = map[string]func(spec.Step) bool{
	"request":       func(st spec.Step) bool { return st.Request != nil },
	"script":        func(st spec.Step) bool { return st.Script != nil },
	"extensionCall": func(st spec.Step) bool { return st.ExtensionCall != nil },
	"composite":     func(st spec.Step) bool { return st.Composite != nil },
}

// bodyCount returns the number of non-nil body fields set on the step.
func bodyCount(st spec.Step) int {
	n := 0
	for _, present := range stepBodyPresent {
		if present(st) {
			n++
		}
	}
	return n
}

// validateStepBody enforces Rule 10. Reference: tractl_spec.md §7.2
func validateStepBody(st spec.Step, sField string) []ValidationError {
	// Only validate when kind is known; Rule 9 already emits for missing/invalid kind.
	if st.Kind == "" || !allowedStepKinds[st.Kind] {
		return nil
	}
	if bodyCount(st) > 1 {
		return []ValidationError{{
			Field:   sField,
			Code:    StepBodyKindMismatch,
			Message: fmt.Sprintf("%s has multiple body fields set; exactly one must match kind %q", sField, st.Kind),
		}}
	}
	if present := stepBodyPresent[st.Kind]; !present(st) {
		return []ValidationError{{
			Field:   sField,
			Code:    StepBodyKindMismatch,
			Message: fmt.Sprintf("%s kind is %q but %s body is nil", sField, st.Kind, st.Kind),
		}}
	}
	return nil
}
