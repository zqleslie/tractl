package validation

import (
	"fmt"

	"github.com/tractl/tractl/internal/spec"
)

// Diagnostics validation error code constants.
const (
	DiagnosticsUnknownKind      ErrorCode = "diagnostics.unknown_kind"
	DiagnosticsUnknownRetention ErrorCode = "diagnostics.unknown_retention"
)

var validDiagnosticsKinds = map[spec.DiagnosticsKind]struct{}{
	spec.DiagnosticsKindTLS:        {},
	spec.DiagnosticsKindTCP:        {},
	spec.DiagnosticsKindTransport:  {},
	spec.DiagnosticsKindLifecycle:  {},
	spec.DiagnosticsKindDNS:        {},
	spec.DiagnosticsKindConnection: {},
}

var validDiagnosticsRetentions = map[spec.DiagnosticsRetention]struct{}{
	spec.RetentionStep:     {},
	spec.RetentionWorkflow: {},
	spec.RetentionSpec:     {},
}

const validKindsMsg = "valid kinds: tls, tcp, transport, lifecycle, dns, connection"
const validRetentionsMsg = "valid values: step, workflow, spec"

// validateDiagnosticsKinds rejects any diagnostics block (step-level or
// workflow-level) that contains a kind not in the six permitted values from §15.2.
// The block is validated regardless of whether Enabled is true or false.
func validateDiagnosticsKinds(s *spec.TraCtlSpec) []ValidationError {
	var errs []ValidationError
	for _, wf := range s.Workflows {
		if wf.Diagnostics != nil {
			for _, kind := range wf.Diagnostics.Kinds {
				if _, ok := validDiagnosticsKinds[kind]; !ok {
					errs = append(errs, ValidationError{
						Field: fmt.Sprintf("workflows[%s].diagnostics.kinds", wf.ID),
						Code:  DiagnosticsUnknownKind,
						Message: fmt.Sprintf(
							"workflow %q: diagnostics kind %q is not permitted; %s",
							wf.ID, kind, validKindsMsg,
						),
					})
				}
			}
		}
		for _, step := range wf.Steps {
			if step.Diagnostics != nil {
				for _, kind := range step.Diagnostics.Kinds {
					if _, ok := validDiagnosticsKinds[kind]; !ok {
						errs = append(errs, ValidationError{
							Field: fmt.Sprintf("workflows[%s].steps[%s].diagnostics.kinds", wf.ID, step.ID),
							Code:  DiagnosticsUnknownKind,
							Message: fmt.Sprintf(
								"workflow %q step %q: diagnostics kind %q is not permitted; %s",
								wf.ID, step.ID, kind, validKindsMsg,
							),
						})
					}
				}
			}
		}
	}
	return errs
}

// validateDiagnosticsRetention rejects any diagnostics block whose retention
// value is non-empty and not one of the three permitted values from §15.1.
// An empty retention string is valid (treated as the default "workflow").
func validateDiagnosticsRetention(s *spec.TraCtlSpec) []ValidationError {
	var errs []ValidationError
	for _, wf := range s.Workflows {
		if wf.Diagnostics != nil && wf.Diagnostics.Retention != "" {
			if _, ok := validDiagnosticsRetentions[wf.Diagnostics.Retention]; !ok {
				errs = append(errs, ValidationError{
					Field: fmt.Sprintf("workflows[%s].diagnostics.retention", wf.ID),
					Code:  DiagnosticsUnknownRetention,
					Message: fmt.Sprintf(
						"workflow %q: diagnostics retention %q is not valid; %s",
						wf.ID, wf.Diagnostics.Retention, validRetentionsMsg,
					),
				})
			}
		}
		for _, step := range wf.Steps {
			if step.Diagnostics != nil && step.Diagnostics.Retention != "" {
				if _, ok := validDiagnosticsRetentions[step.Diagnostics.Retention]; !ok {
					errs = append(errs, ValidationError{
						Field: fmt.Sprintf("workflows[%s].steps[%s].diagnostics.retention", wf.ID, step.ID),
						Code:  DiagnosticsUnknownRetention,
						Message: fmt.Sprintf(
							"workflow %q step %q: diagnostics retention %q is not valid; %s",
							wf.ID, step.ID, step.Diagnostics.Retention, validRetentionsMsg,
						),
					})
				}
			}
		}
	}
	return errs
}
